package request

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type DecodeOption struct {
	ignoreNotParsed *bool
}

func (d DecodeOption) IgnoreNotParsed() bool {
	return d.ignoreNotParsed != nil && *d.ignoreNotParsed
}

func DecodeOptions(options []DecodeOption) DecodeOption {
	option := DecodeOption{}
	for _, o := range options {
		if o.ignoreNotParsed != nil {
			option.ignoreNotParsed = o.ignoreNotParsed
		}
	}
	return option
}

func IgnoreNotParsed() DecodeOption {
	return DecodeOption{
		ignoreNotParsed: pointer.FromAny(true),
	}
}

func DecodeRequestBody(req *http.Request, decodable any, decodeOptions ...DecodeOption) error {
	if req == nil {
		return errors.New("request is missing")
	}
	if req.Body == nil {
		return ErrorJSONNotFound()
	}

	defer req.Body.Close()
	return DecodeStream(req.Context(), structure.NewPointerSource(), req.Body, decodable, decodeOptions...)
}

func DecodeResponseBody(ctx context.Context, res *http.Response, decodable any, decodeOptions ...DecodeOption) error {
	if res == nil {
		return errors.New("response is missing")
	}
	if res.Body == nil {
		return ErrorJSONNotFound()
	}

	defer res.Body.Close()
	return DecodeStream(ctx, structure.NewPointerSource(), res.Body, decodable, decodeOptions...)
}

func DecodeStream(ctx context.Context, source structure.Source, reader io.Reader, decodable any, decodeOptions ...DecodeOption) error {
	if err := ParseStream(ctx, source, reader, decodable, decodeOptions...); err != nil {
		return err
	}
	if err := ValidateObjects(ctx, source, decodable); err != nil {
		return err
	}
	return NormalizeObjects(ctx, source, decodable)
}

func DecodeObject(ctx context.Context, source structure.Source, object map[string]any, decodable any, decodeOptions ...DecodeOption) error {
	if err := ParseObject(ctx, source, object, decodable, decodeOptions...); err != nil {
		return err
	}
	if err := ValidateObjects(ctx, source, decodable); err != nil {
		return err
	}
	return NormalizeObjects(ctx, source, decodable)
}

func DecodeArray(ctx context.Context, source structure.Source, array []any, decodable any, decodeOptions ...DecodeOption) error {
	if err := ParseArray(ctx, source, array, decodable, decodeOptions...); err != nil {
		return err
	}
	if err := ValidateObjects(ctx, source, decodable); err != nil {
		return err
	}
	return NormalizeObjects(ctx, source, decodable)
}

func ParseStream(ctx context.Context, source structure.Source, reader io.Reader, parsable any, decodeOptions ...DecodeOption) error {
	if objectParsable, ok := parsable.(structure.ObjectParsable); ok {
		return ParseStreamObjectParsable(ctx, source, reader, objectParsable, decodeOptions...)
	}
	if arrayParsable, ok := parsable.(structure.ArrayParsable); ok {
		return ParseStreamArrayParsable(ctx, source, reader, arrayParsable, decodeOptions...)
	}
	return ParseStreamSimple(ctx, reader, parsable, decodeOptions...)
}

func ParseStreamObjectParsable(ctx context.Context, source structure.Source, reader io.Reader, objectParsable structure.ObjectParsable, decodeOptions ...DecodeOption) error {
	object := map[string]any{}
	if err := ParseStreamSimple(ctx, reader, &object); err != nil {
		return err
	}
	return ParseObjectParsable(ctx, source, object, objectParsable, decodeOptions...)
}

func ParseStreamArrayParsable(ctx context.Context, source structure.Source, reader io.Reader, arrayParsable structure.ArrayParsable, decodeOptions ...DecodeOption) error {
	array := []any{}
	if err := ParseStreamSimple(ctx, reader, &array); err != nil {
		return err
	}
	return ParseArrayParsable(ctx, source, array, arrayParsable, decodeOptions...)
}

func ParseStreamSimple(ctx context.Context, reader io.Reader, parsable any, decodeOptions ...DecodeOption) error {
	if reader == nil {
		return errors.New("reader is missing")
	}
	if parsable == nil {
		return errors.New("parsable is missing")
	}

	if err := json.NewDecoder(reader).Decode(parsable); err != nil {
		if err == io.EOF {
			return ErrorJSONNotFound()
		}
		return ErrorJSONMalformed()
	}

	return nil
}

func ParseObject(ctx context.Context, source structure.Source, object map[string]any, parsable any, decodeOptions ...DecodeOption) error {
	if objectParsable, ok := parsable.(structure.ObjectParsable); ok {
		return ParseObjectParsable(ctx, source, object, objectParsable, decodeOptions...)
	} else {
		return ParseSimple(ctx, object, parsable, decodeOptions...)
	}
}

func ParseObjectParsable(ctx context.Context, source structure.Source, object map[string]any, objectParsable structure.ObjectParsable, decodeOptions ...DecodeOption) error {
	if object == nil {
		return errors.New("object is missing")
	}
	if objectParsable == nil {
		return errors.New("object parsable is missing")
	}

	parser := structureParser.NewObject(log.LoggerFromContext(ctx), &object).WithSource(source)
	objectParsable.Parse(parser)
	if !DecodeOptions(decodeOptions).IgnoreNotParsed() {
		parser.NotParsed()
	}
	return parser.Error()
}

func ParseArray(ctx context.Context, source structure.Source, array []any, parsable any, decodeOptions ...DecodeOption) error {
	if arrayParsable, ok := parsable.(structure.ArrayParsable); ok {
		return ParseArrayParsable(ctx, source, array, arrayParsable, decodeOptions...)
	} else {
		return ParseSimple(ctx, array, parsable, decodeOptions...)
	}
}

func ParseArrayParsable(ctx context.Context, source structure.Source, array []any, arrayParsable structure.ArrayParsable, decodeOptions ...DecodeOption) error {
	if array == nil {
		return errors.New("array is missing")
	}
	if arrayParsable == nil {
		return errors.New("array parsable is missing")
	}

	parser := structureParser.NewArray(log.LoggerFromContext(ctx), &array).WithSource(source)
	arrayParsable.Parse(parser)
	if !DecodeOptions(decodeOptions).IgnoreNotParsed() {
		parser.NotParsed()
	}
	return parser.Error()
}

func ParseSimple(ctx context.Context, simple any, parsable any, decodeOptions ...DecodeOption) error {
	if simple == nil {
		return errors.New("simple is missing")
	}
	if parsable == nil {
		return errors.New("parsable is missing")
	}

	if bites, err := json.Marshal(simple); err != nil {
		return errors.New("unable to encode simple")
	} else if err := json.Unmarshal(bites, parsable); err != nil {
		return errors.New("unable to decode parsable")
	}
	return nil
}

func ValidateObjects(ctx context.Context, source structure.Source, anyObjects ...any) error {
	var validatableObjects []structure.Validatable
	for _, anyObject := range anyObjects {
		if validatable, ok := anyObject.(structure.Validatable); ok {
			validatableObjects = append(validatableObjects, validatable)
		}
	}

	if len(validatableObjects) == 0 {
		return nil
	}

	validator := structureValidator.New(log.LoggerFromContext(ctx)).WithSource(source)
	for _, validatable := range validatableObjects {
		validatable.Validate(validator)
	}
	return validator.Error()
}

func NormalizeObjects(ctx context.Context, source structure.Source, anyObjects ...any) error {
	var normalizableObjects []structure.Normalizable
	for _, anyObject := range anyObjects {
		if normalizable, ok := anyObject.(structure.Normalizable); ok {
			normalizableObjects = append(normalizableObjects, normalizable)
		}
	}

	if len(normalizableObjects) == 0 {
		return nil
	}

	normalizer := structureNormalizer.New(log.LoggerFromContext(ctx)).WithSource(source)
	for _, normalizable := range normalizableObjects {
		normalizable.Normalize(normalizer)
	}
	return normalizer.Error()
}

func DecodeRequestQuery(req *http.Request, objectParsableObjects ...structure.ObjectParsable) error {
	if req == nil {
		return errors.New("request is missing")
	}
	if req.URL == nil {
		return errors.New("request url is missing")
	}

	values, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return errors.New("unable to parse request query")
	}

	return DecodeValues(req.Context(), (map[string][]string)(values), objectParsableObjects...)
}

func DecodeValues(ctx context.Context, values map[string][]string, objectParsableObjects ...structure.ObjectParsable) error {
	objects := []any{}
	for _, object := range objectParsableObjects {
		objects = append(objects, object)
	}

	if err := ParseValuesObjects(ctx, values, objectParsableObjects...); err != nil {
		return err
	}
	if err := ValidateObjects(ctx, structure.NewParameterSource(), objects...); err != nil {
		return err
	}
	return NormalizeObjects(ctx, structure.NewParameterSource(), objects...)
}

func ParseValuesObjects(ctx context.Context, values map[string][]string, objectParsableObjects ...structure.ObjectParsable) error {
	parser := NewValues(log.LoggerFromContext(ctx), &values)
	for _, objectParsable := range objectParsableObjects {
		objectParsable.Parse(parser)
	}
	return parser.NotParsed()
}

func ParseSingletonHeader(header http.Header, key string) (*string, error) {
	if values, ok := header[key]; !ok {
		return nil, nil
	} else if length := len(values); length == 0 {
		return nil, nil
	} else if length == 1 {
		return &values[0], nil
	}
	return nil, ErrorHeaderInvalid(key)
}

func ParseDigestMD5Header(header http.Header, key string) (*string, error) {
	if stringValue, err := ParseSingletonHeader(header, key); err != nil || stringValue == nil {
		return nil, err
	} else if parts := strings.SplitN(*stringValue, "=", 2); len(parts) == 2 {
		if algorithm := strings.ToUpper(parts[0]); algorithm == "MD5" {
			if value := parts[1]; crypto.IsValidBase64EncodedMD5Hash(value) {
				return &value, nil
			}
		}
	}
	return nil, ErrorHeaderInvalid(key)
}

func ParseMediaTypeHeader(header http.Header, key string) (*string, error) {
	if stringValue, err := ParseSingletonHeader(header, key); err != nil || stringValue == nil {
		return nil, err
	} else if value, valid := net.NormalizeMediaType(*stringValue); valid {
		return &value, nil
	}
	return nil, ErrorHeaderInvalid(key)
}

func ParseTimeHeader(header http.Header, key string, layout string) (*time.Time, error) {
	if stringValue, err := ParseSingletonHeader(header, key); err != nil || stringValue == nil {
		return nil, err
	} else if value, valueErr := time.ParseInLocation(layout, *stringValue, time.UTC); valueErr == nil {
		return &value, nil
	}
	return nil, ErrorHeaderInvalid(key)
}

func ParseIntHeader(header http.Header, key string) (*int, error) {
	if stringValue, err := ParseSingletonHeader(header, key); err != nil || stringValue == nil {
		return nil, err
	} else if value, valueErr := strconv.Atoi(*stringValue); valueErr == nil {
		return &value, nil
	}
	return nil, ErrorHeaderInvalid(key)
}
