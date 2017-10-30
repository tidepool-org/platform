package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/structure"
)

type Source interface {
	Parameter() string
	Pointer() string
}

type source struct {
	Parameter string `json:"parameter,omitempty" bson:"parameter,omitempty"`
	Pointer   string `json:"pointer,omitempty" bson:"pointer,omitempty"`
}

func (s *source) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("parameter"); ptr != nil {
		s.Parameter = *ptr
	}
	if ptr := parser.String("pointer"); ptr != nil {
		s.Pointer = *ptr
	}
}

func (s *source) Validate(validator structure.Validator) {
	if s.Parameter == "" {
		validator.String("pointer", &s.Pointer).NotEmpty()
	}
}

func (s *source) Normalize(normalizer structure.Normalizer) {}

func New(detail string) error {
	return &object{
		Detail: detail,
		Caller: GetCaller(1),
	}
}

func Newf(format string, a ...interface{}) error {
	return &object{
		Detail: fmt.Sprintf(format, a...),
		Caller: GetCaller(1),
	}
}

func Wrap(err error, detail string) error {
	var cause *Serializable
	if err != nil {
		cause = &Serializable{
			Error: err,
		}
	}
	return &object{
		Detail: detail,
		Caller: GetCaller(1),
		Cause:  cause,
	}
}

func Wrapf(err error, format string, a ...interface{}) error {
	var cause *Serializable
	if err != nil {
		cause = &Serializable{
			Error: err,
		}
	}
	return &object{
		Detail: fmt.Sprintf(format, a...),
		Caller: GetCaller(1),
		Cause:  cause,
	}
}

func Prepared(code string, title string, detail string) error {
	return &object{
		Code:   code,
		Title:  title,
		Detail: detail,
		Caller: GetCaller(2),
	}
}

func Preparedf(code string, title string, format string, a ...interface{}) error {
	return &object{
		Code:   code,
		Title:  title,
		Detail: fmt.Sprintf(format, a...),
		Caller: GetCaller(2),
	}
}

func WithSource(err error, src Source) error {
	var s *source

	if src != nil {
		parameter := src.Parameter()
		pointer := src.Pointer()
		if parameter != "" || pointer != "" {
			s = &source{
				Parameter: parameter,
				Pointer:   pointer,
			}
		}
	}

	if _, arrayOK := err.(*array); arrayOK {
		return err
	} else if objectErr, objectOK := err.(*object); objectOK {
		return &object{
			Code:   objectErr.Code,
			Title:  objectErr.Title,
			Detail: objectErr.Detail,
			Source: s,
			Meta:   objectErr.Meta,
			Caller: objectErr.Caller,
			Cause:  objectErr.Cause,
		}
	} else if err != nil {
		return &object{
			Detail: err.Error(),
			Source: s,
		}
	}
	return nil
}

func WithMeta(err error, meta interface{}) error {
	if _, arrayOK := err.(*array); arrayOK {
		return err
	} else if objectErr, objectOK := err.(*object); objectOK {
		return &object{
			Code:   objectErr.Code,
			Title:  objectErr.Title,
			Detail: objectErr.Detail,
			Source: objectErr.Source,
			Meta:   meta,
			Caller: objectErr.Caller,
			Cause:  objectErr.Cause,
		}
	} else if err != nil {
		return &object{
			Detail: err.Error(),
			Meta:   meta,
		}
	}
	return nil
}

func Code(err error) string {
	if objectErr, objectOK := err.(*object); objectOK {
		return objectErr.Code
	}
	return "internal-error"
}

func Cause(err error) error {
	if objectErr, objectOK := err.(*object); objectOK && objectErr.Cause != nil && objectErr.Cause.Error != nil {
		return Cause(objectErr.Cause.Error)
	}
	return err
}

func Append(errs ...error) error {
	var errors []error
	for _, err := range errs {
		if err != nil {
			errors = appendErrors(errors, err)
		}
	}
	switch len(errors) {
	case 0:
		return nil
	case 1:
		return errors[0]
	default:
		return &array{
			Errors: errors,
		}
	}
}

func appendErrors(errors []error, err error) []error {
	if arrayErr, arrayOK := err.(*array); arrayOK {
		return append(errors, arrayErr.Errors...)
	} else if objectErr, objectOK := err.(*object); objectOK {
		return append(errors, objectErr)
	} else if err != nil {
		return append(errors, Normalize(err))
	}
	return errors
}

func Normalize(err error) error {
	if _, arrayOK := err.(*array); arrayOK {
		return err
	} else if _, objectOK := err.(*object); objectOK {
		return err
	} else if err != nil {
		return &object{
			Detail: err.Error(),
		}
	}
	return nil
}

type Sanitizable interface {
	Sanitize() error
}

func Sanitize(err error) error {
	if sanitizable, ok := err.(Sanitizable); ok {
		return sanitizable.Sanitize()
	}
	return err
}

type Serializable struct {
	Error error
}

func (s *Serializable) Parse(reference string, parser structure.ObjectParser) {
	if iface := parser.Interface(reference); iface != nil {
		if _, arrayOK := (*iface).([]interface{}); arrayOK {
			arrayErr := &array{}
			arrayParser := parser.WithReferenceArrayParser(reference)
			arrayErr.Parse(arrayParser)
			arrayParser.NotParsed()
			s.Error = arrayErr
		} else if _, objectOK := (*iface).(map[string]interface{}); objectOK {
			objectErr := &object{}
			objectParser := parser.WithReferenceObjectParser(reference)
			objectErr.Parse(objectParser)
			objectParser.NotParsed()
			s.Error = objectErr
		} else if ptr := parser.String(reference); ptr != nil {
			s.Error = errors.New(*ptr)
		}
	}
}

func (s *Serializable) Validate(validator structure.Validator) {
	if arrayErr, arrayOK := s.Error.(*array); arrayOK {
		arrayErr.Validate(validator)
	} else if objectErr, objectOK := s.Error.(*object); objectOK {
		objectErr.Validate(validator)
	}
}

func (s *Serializable) Normalize(normalizer structure.Normalizer) {
	if arrayErr, arrayOK := s.Error.(*array); arrayOK {
		arrayErr.Normalize(normalizer)
	} else if objectErr, objectOK := s.Error.(*object); objectOK {
		objectErr.Normalize(normalizer)
	}
}

func (s Serializable) MarshalJSON() ([]byte, error) {
	if arrayErr, arrayOK := s.Error.(*array); arrayOK {
		return json.Marshal(arrayErr.Errors)
	} else if objectErr, objectOK := s.Error.(*object); objectOK {
		return json.Marshal(objectErr)
	} else if s.Error != nil {
		return []byte(strconv.Quote(s.Error.Error())), nil
	}
	return nil, nil
}

func (s *Serializable) UnmarshalJSON(bytes []byte) error {
	errObject := &object{}
	if err := json.Unmarshal(bytes, &errObject); err != nil {
		errObjects := []*object{}
		if err = json.Unmarshal(bytes, &errObjects); err != nil {
			var errString string
			if err = json.Unmarshal(bytes, &errString); err != nil {
				return err
			}
			s.Error = errors.New(errString)
		} else {
			var errors []error
			for _, errObject := range errObjects {
				errors = append(errors, errObject)
			}
			s.Error = &array{Errors: errors}
		}
	} else {
		s.Error = errObject
	}
	return nil
}

func (s Serializable) GetBSON() (interface{}, error) {
	if arrayErr, arrayOK := s.Error.(*array); arrayOK {
		return arrayErr.Errors, nil
	} else if objectErr, objectOK := s.Error.(*object); objectOK {
		return objectErr, nil
	} else if s.Error != nil {
		return s.Error.Error(), nil
	}
	return nil, nil
}

func (s *Serializable) SetBSON(raw bson.Raw) error {
	errObject := &object{}
	if err := raw.Unmarshal(&errObject); err != nil {
		errObjects := []*object{}
		if err = raw.Unmarshal(&errObjects); err != nil {
			var errString string
			if err = raw.Unmarshal(&errString); err != nil {
				return err
			}
			s.Error = errors.New(errString)
		} else {
			var errors []error
			for _, errObject := range errObjects {
				errors = append(errors, errObject)
			}
			s.Error = &array{Errors: errors}
		}
	} else {
		s.Error = errObject
	}
	return nil
}

func ErrorInternal(err error) error {
	return &object{
		Code:   "internal-error",
		Title:  "internal error",
		Detail: "internal error",
		Caller: GetCaller(1),
		Cause:  &Serializable{Error: err},
	}
}

type array struct {
	Errors []error `json:"errors,omitempty" bson:"errors,omitempty"`
}

func (a *array) Error() string {
	var strs []string
	for _, err := range a.Errors {
		strs = append(strs, err.Error())
	}
	return strings.Join(strs, ", ")
}

func (a *array) Format(state fmt.State, verb rune) {
	for index, err := range a.Errors {
		if index > 0 {
			fmt.Fprintf(state, ", ")
		}
		if formatter, ok := err.(fmt.Formatter); ok {
			formatter.Format(state, verb)
		} else {
			switch verb {
			case 's':
				fmt.Fprintf(state, "%s", err)
			case 'q':
				fmt.Fprintf(state, "%q", err)
			case 'v':
				if state.Flag('#') {
					fmt.Fprintf(state, "%#v", err)
				} else if state.Flag('+') {
					fmt.Fprintf(state, "%+v", err)
				} else {
					fmt.Fprintf(state, "%v", err)
				}
			}
		}
	}
}

func (a *array) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		objectErr := &object{}
		objectParser := parser.WithReferenceObjectParser(reference)
		objectErr.Parse(objectParser)
		objectParser.NotParsed()
		a.Errors = append(a.Errors, objectErr)
	}
}

func (a *array) Validate(validator structure.Validator) {
	for index, err := range a.Errors {
		if objectErr, objectOK := err.(*object); objectOK {
			objectErr.Validate(validator.WithReference(strconv.Itoa(index)))
		}
	}
}

func (a *array) Normalize(normalizer structure.Normalizer) {
	for index, err := range a.Errors {
		if objectErr, objectOK := err.(*object); objectOK {
			objectErr.Normalize(normalizer.WithReference(strconv.Itoa(index)))
		}
	}
}

func (a *array) Sanitize() error {
	var errors []error
	for _, err := range a.Errors {
		if sanitizedable, ok := err.(Sanitizable); ok {
			err = sanitizedable.Sanitize()
		}
		if err != nil {
			errors = append(errors, err)
		}
	}
	return &array{
		Errors: errors,
	}
}

type object struct {
	Code   string        `json:"code,omitempty" bson:"code,omitempty"`
	Title  string        `json:"title,omitempty" bson:"title,omitempty"`
	Detail string        `json:"detail" bson:"detail"`
	Source *source       `json:"source,omitempty" bson:"source,omitempty"`
	Meta   interface{}   `json:"meta,omitempty" bson:"meta,omitempty"`
	Caller *Caller       `json:"caller,omitempty" bson:"caller,omitempty"`
	Cause  *Serializable `json:"cause,omitempty" bson:"cause,omitempty"`
}

func (o *object) Error() string {
	str := o.Detail
	if o.Cause != nil && o.Cause.Error != nil {
		str += "; " + o.Cause.Error.Error()
	}
	return str
}

func (o *object) Format(state fmt.State, verb rune) {
	switch verb {
	case 's':
		fmt.Fprintf(state, "%s", o.Error())
	case 'q':
		fmt.Fprintf(state, "%q", o.Error())
	case 'v':
		if state.Flag('#') {
			var parts []string
			if o.Code != "" {
				parts = append(parts, fmt.Sprintf("Code:%q", o.Code))
			}
			if o.Title != "" {
				parts = append(parts, fmt.Sprintf("Title:%q", o.Title))
			}
			parts = append(parts, fmt.Sprintf("Detail:%q", o.Detail))
			if o.Source != nil {
				parts = append(parts, fmt.Sprintf("Source:%#v", o.Source))
			}
			if o.Meta != nil {
				parts = append(parts, fmt.Sprintf("Meta:%#v", o.Meta))
			}
			if o.Caller != nil {
				parts = append(parts, fmt.Sprintf("Caller:%#v", o.Caller))
			}
			if o.Cause != nil {
				parts = append(parts, fmt.Sprintf("Cause:%#+v", o.Cause.Error))
			}
			fmt.Fprintf(state, "&errors.object{%s}", strings.Join(parts, ", "))
		} else {
			fmt.Fprintf(state, "%s", o.Error())
		}
	}
}

func (o *object) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("code"); ptr != nil {
		o.Code = *ptr
	}
	if ptr := parser.String("title"); ptr != nil {
		o.Title = *ptr
	}
	if ptr := parser.String("detail"); ptr != nil {
		o.Detail = *ptr
	}
	if sourceParser := parser.WithReferenceObjectParser("source"); sourceParser.Exists() {
		o.Source = &source{}
		o.Source.Parse(sourceParser)
		sourceParser.NotParsed()
	}
	if ptr := parser.Interface("meta"); ptr != nil {
		o.Meta = *ptr
	}
	if callerParser := parser.WithReferenceObjectParser("caller"); callerParser.Exists() {
		o.Caller = &Caller{}
		o.Caller.Parse(callerParser)
		callerParser.NotParsed()
	}
	if parser.ReferenceExists("cause") {
		o.Cause = &Serializable{}
		o.Cause.Parse("cause", parser)
	}
}

func (o *object) Validate(validator structure.Validator) {
	if o.Source != nil {
		o.Source.Validate(validator.WithReference("source"))
	}
	if o.Caller != nil {
		o.Caller.Validate(validator.WithReference("caller"))
	}
	if o.Cause != nil {
		o.Cause.Validate(validator.WithReference("cause"))
	}
}

func (o *object) Normalize(normalizer structure.Normalizer) {
	if o.Source != nil {
		o.Source.Normalize(normalizer.WithReference("source"))
	}
	if o.Caller != nil {
		o.Caller.Normalize(normalizer.WithReference("caller"))
	}
	if o.Cause != nil {
		o.Cause.Normalize(normalizer.WithReference("cause"))
	}
}

func (o *object) Sanitize() error {
	return &object{
		Code:   o.Code,
		Title:  o.Title,
		Detail: o.Detail,
		Source: o.Source,
		Meta:   o.Meta,
	}
}

type contextKey string

const errorContextKey contextKey = "error"

func NewContextWithError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, errorContextKey, err)
}

func ErrorFromContext(ctx context.Context) error {
	if ctx != nil {
		if err, ok := ctx.Value(errorContextKey).(error); ok {
			return err
		}
	}
	return nil
}
