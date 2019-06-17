package unstructured

import (
	"context"
	"io"
	"regexp"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Store interface {
	Exists(ctx context.Context, key string) (bool, error)
	Put(ctx context.Context, key string, reader io.Reader, options *Options) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) (bool, error)
	DeleteDirectory(ctx context.Context, key string) error
}

func NewOptions() *Options {
	return &Options{}
}

type Options struct {
	MediaType *string `json:"mediaType,omitempty"`
}

func (o *Options) Validate(validator structure.Validator) {
	validator.String("mediaType", o.MediaType).Using(net.MediaTypeValidator)
}

func IsValidKey(value string) bool {
	return ValidateKey(value) == nil
}

func KeyValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateKey(value))
}

func ValidateKey(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !keyExpression.MatchString(value) {
		return ErrorValueStringAsKeyNotValid(value)
	} else if length := len(value); length > keyLengthMaximum {
		return structureValidator.ErrorLengthNotLessThanOrEqualTo(length, keyLengthMaximum)
	}
	return nil
}

func ErrorValueStringAsKeyNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as unstructured key", value)
}

const keyLengthMaximum = 2047

var keyExpression = regexp.MustCompile("^[0-9A-Za-z][0-9A-Za-z._=-]*(/[0-9A-Za-z][0-9A-Za-z._=-]*)*$")
