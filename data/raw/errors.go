package raw

import (
	"net/http"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

const (
	ErrorCodeTypeNotBool          = "type-not-bool"
	ErrorCodeTypeNotFloat64       = "type-not-float64"
	ErrorCodeTypeNotInt           = "type-not-int"
	ErrorCodeTypeNotString        = "type-not-string"
	ErrorCodeTypeNotTime          = "type-not-time"
	ErrorCodeTypeNotObject        = "type-not-object"
	ErrorCodeTypeNotArray         = "type-not-array"
	ErrorCodeTypeNotJSON          = "type-not-json"
	ErrorCodeValueNotParsable     = "value-not-parsable"
	ErrorCodeNotParsed            = "not-parsed"
	ErrorCodeValueNotExists       = "value-not-exists"
	ErrorCodeValueExists          = "value-exists"
	ErrorCodeValueNotEmpty        = "value-not-empty"
	ErrorCodeValueEmpty           = "value-empty"
	ErrorCodeValueDuplicate       = "value-duplicate"
	ErrorCodeValueNotTrue         = "value-not-true"
	ErrorCodeValueNotFalse        = "value-not-false"
	ErrorCodeValueOutOfRange      = "value-out-of-range"
	ErrorCodeValueDisallowed      = "value-disallowed"
	ErrorCodeValueNotAllowed      = "value-not-allowed"
	ErrorCodeValueMatches         = "value-matches"
	ErrorCodeValueNotMatches      = "value-not-matches"
	ErrorCodeValueNotAfter        = "value-not-after"
	ErrorCodeValueNotBefore       = "value-not-before"
	ErrorCodeValueNotValid        = "value-not-valid"
	ErrorCodeValueNotSerializable = "value-not-serializable"
	ErrorCodeValuesNotExistForAny = "values-not-exist-for-any"
	ErrorCodeValuesNotExistForOne = "values-not-exist-for-one"
	ErrorCodeLengthOutOfRange     = "length-out-of-range"
	ErrorCodeSizeOutOfRange       = "size-out-of-range"
	ErrorCodeJSONMalformed        = "json-malformed"

	SourceTypeBlob = "blob"
	SourceTypeRaw  = "raw"
)

func SourceTypes() []string {
	return []string{
		SourceTypeBlob,
		SourceTypeRaw,
	}
}

type ErrorsFilter struct {
	StartTime *time.Time
	EndTime   *time.Time
}

func NewErrorsFilter() *ErrorsFilter {
	return &ErrorsFilter{}
}

func (e *ErrorsFilter) Parse(parser structure.ObjectParser) {
	e.StartTime = parser.Time("startTime", time.RFC3339Nano)
	e.EndTime = parser.Time("endTime", time.RFC3339Nano)
}

func (e *ErrorsFilter) Validate(validator structure.Validator) {
	validator.Time("startTime", e.StartTime)
	if e.StartTime != nil {
		validator.Time("endTime", e.EndTime).After(*e.StartTime)
	}
}

func (e *ErrorsFilter) MutateRequest(req *http.Request) error {
	parameters := map[string]string{}
	if e.StartTime != nil {
		parameters["startTime"] = e.StartTime.Format(time.RFC3339Nano)
	}
	if e.EndTime != nil {
		parameters["endTime"] = e.EndTime.Format(time.RFC3339Nano)
	}
	return request.NewParametersMutator(parameters).MutateRequest(req)
}

type Error struct {
	Code        string  `json:"code,omitempty" bson:"code,omitempty"`
	Description *string `json:"description,omitempty" bson:"description,omitempty"`
	Reference   *string `json:"reference,omitempty" bson:"reference,omitempty"`
}

func ParseError(parser structure.ObjectParser) *Error {
	if !parser.Exists() {
		return nil
	}
	datum := NewError()
	parser.Parse(datum)
	return datum
}

func NewError() *Error {
	return &Error{}
}

func (e *Error) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("code"); ptr != nil {
		e.Code = *ptr
	}
	e.Description = parser.String("description")
	e.Reference = parser.String("reference")
}

func (e *Error) Validate(validator structure.Validator) {
	validator.String("code", &e.Code).NotEmpty()
	validator.String("description", e.Description).NotEmpty()
	validator.String("reference", e.Reference).NotEmpty()
}

type ErrorArray []*Error

func ParseErrorArray(parser structure.ArrayParser) *ErrorArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewErrorArray()
	parser.Parse(datum)
	return datum
}

func NewErrorArray() *ErrorArray {
	return &ErrorArray{}
}

func (e *ErrorArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*e = append(*e, ParseError(parser.WithReferenceObjectParser(reference)))
	}
}

func (e *ErrorArray) Validate(validator structure.Validator) {
	if length := len(*e); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	}

	for index, datum := range *e {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type SourceErrors struct {
	SourceType      string      `json:"sourceType,omitempty" bson:"sourceType,omitempty"`
	SourceReference string      `json:"sourceReference,omitempty" bson:"sourceReference,omitempty"`
	Errors          *ErrorArray `json:"errors,omitempty" bson:"errors,omitempty"`
}

func ParseSourceErrors(parser structure.ObjectParser) *SourceErrors {
	if !parser.Exists() {
		return nil
	}
	datum := NewSourceErrors()
	parser.Parse(datum)
	return datum
}

func NewSourceErrors() *SourceErrors {
	return &SourceErrors{}
}

func (s *SourceErrors) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("sourceType"); ptr != nil {
		s.SourceType = *ptr
	}
	if ptr := parser.String("sourceReference"); ptr != nil {
		s.SourceReference = *ptr
	}
	s.Errors = ParseErrorArray(parser.WithReferenceArrayParser("errors"))
}

func (s *SourceErrors) Validate(validator structure.Validator) {
	validator.String("sourceType", &s.SourceType).OneOf(SourceTypes()...)
	validator.String("sourceReference", &s.SourceReference).NotEmpty()
	if errorsValidator := validator.WithReference("errors"); s.Errors != nil {
		s.Errors.Validate(errorsValidator)
	} else {
		errorsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

type SourceErrorsArray []*SourceErrors

func ParseSourceErrorsArray(parser structure.ArrayParser) *SourceErrorsArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewSourceErrorsArray()
	parser.Parse(datum)
	return datum
}

func NewSourceErrorsArray() *SourceErrorsArray {
	return &SourceErrorsArray{}
}

func (e *SourceErrorsArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*e = append(*e, ParseSourceErrors(parser.WithReferenceObjectParser(reference)))
	}
}

func (e *SourceErrorsArray) Validate(validator structure.Validator) {
	if length := len(*e); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	}

	for index, datum := range *e {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type DataSetErrors struct {
	UserID       string             `json:"userId,omitempty" bson:"userId,omitempty"`
	DataSetID    string             `json:"dataSetId,omitempty" bson:"dataSetId,omitempty"`
	SourceErrors *SourceErrorsArray `json:"sourceErrors,omitempty" bson:"sourceErrors,omitempty"`
}

func (d *DataSetErrors) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("userId"); ptr != nil {
		d.UserID = *ptr
	}
	if ptr := parser.String("dataSetId"); ptr != nil {
		d.DataSetID = *ptr
	}
	d.SourceErrors = ParseSourceErrorsArray(parser.WithReferenceArrayParser("sourceErrors"))
}

func (d *DataSetErrors) Validate(validator structure.Validator) {
	validator.String("userId", &d.UserID).Exists().Using(user.IDValidator)
	validator.String("dataSetId", &d.DataSetID).Exists().Using(data.SetIDValidator)
	if sourceErrorsValidator := validator.WithReference("sourceErrors"); d.SourceErrors != nil {
		d.SourceErrors.Validate(sourceErrorsValidator)
	} else {
		sourceErrorsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

type DataSetErrorsArray []*DataSetErrors
