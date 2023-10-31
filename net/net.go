package net

import (
	"net/url"
	"regexp"

	"github.com/blang/semver"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

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
	reverseDomainLengthMaximum   = 253
	semanticVersionLengthMaximum = 256
	urlLengthMaximum             = 2047
)

var reverseDomainExpression = regexp.MustCompile(`^[a-zA-Z0-9](|[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])(\.[a-zA-Z0-9](|[a-zA-Z0-9-]{0,61}[a-zA-Z0-9]))+$`)
