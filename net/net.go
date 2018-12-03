package net

import (
	"mime"
	"net/url"
	"regexp"

	"github.com/blang/semver"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func IsValidMediaType(value string) bool {
	return ValidateMediaType(value) == nil
}

func MediaTypeValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateMediaType(value))
}

func ValidateMediaType(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if _, _, err := mime.ParseMediaType(value); err != nil {
		return ErrorValueStringAsMediaTypeNotValid(value)
	} else if length := len(value); length > mediaTypeLengthMaximum {
		return structureValidator.ErrorLengthNotLessThanOrEqualTo(length, mediaTypeLengthMaximum)
	}
	return nil
}

func NormalizeMediaType(value string) (string, bool) {
	mediaType, parameters, err := mime.ParseMediaType(value)
	if err != nil {
		return "", false
	}
	result := mime.FormatMediaType(mediaType, parameters)
	if result == "" {
		return "", false
	}
	return result, true
}

func IsValidReverseDomain(value string) bool {
	return ValidateReverseDomain(value) == nil
}

func ReverseDomainValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateReverseDomain(value))
}

func ValidateReverseDomain(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !reverseDomainExpression.MatchString(value) {
		return ErrorValueStringAsReverseDomainNotValid(value)
	} else if length := len(value); length > reverseDomainLengthMaximum {
		return structureValidator.ErrorLengthNotLessThanOrEqualTo(length, reverseDomainLengthMaximum)
	}
	return nil
}

func IsValidSemanticVersion(value string) bool {
	return ValidateSemanticVersion(value) == nil
}

func SemanticVersionValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateSemanticVersion(value))
}

func ValidateSemanticVersion(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if _, err := semver.Parse(value); err != nil {
		return ErrorValueStringAsSemanticVersionNotValid(value)
	} else if length := len(value); length > semanticVersionLengthMaximum {
		return structureValidator.ErrorLengthNotLessThanOrEqualTo(length, semanticVersionLengthMaximum)
	}
	return nil
}

func IsValidURL(value string) bool {
	return ValidateURL(value) == nil
}

func URLValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateURL(value))
}

func ValidateURL(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if earl, err := url.Parse(value); err != nil || earl == nil || !earl.IsAbs() || earl.Host == "" {
		return ErrorValueStringAsURLNotValid(value)
	} else if length := len(value); length > urlLengthMaximum {
		return structureValidator.ErrorLengthNotLessThanOrEqualTo(length, urlLengthMaximum)
	}
	return nil
}

func ErrorValueStringAsMediaTypeNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as media type", value)
}

func ErrorValueStringAsReverseDomainNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as reverse domain", value)
}

func ErrorValueStringAsSemanticVersionNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as semantic version", value)
}

func ErrorValueStringAsURLNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as url", value)
}

const (
	mediaTypeLengthMaximum       = 256
	reverseDomainLengthMaximum   = 253
	semanticVersionLengthMaximum = 256
	urlLengthMaximum             = 2047
)

var reverseDomainExpression = regexp.MustCompile(`^[a-zA-Z0-9](|[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])(\.[a-zA-Z0-9](|[a-zA-Z0-9-]{0,61}[a-zA-Z0-9]))+$`)
