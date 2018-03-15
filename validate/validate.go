package validate

import (
	"net/url"
	"regexp"

	"github.com/blang/semver"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	ReverseDomainLengthMaximum   = 253
	SemanticVersionLengthMaximum = 100
	URLLengthMaximum             = 2000
)

var reverseDomainExpression = regexp.MustCompile(`^[a-z]{2,63}(\.([a-z0-9]|[a-z0-9][a-z0-9-]{0,61}[a-z0-9]))+$`)

func ReverseDomain(value string, errorReporter structure.ErrorReporter) {
	if value == "" {
		errorReporter.ReportError(structureValidator.ErrorValueEmpty())
	} else if !reverseDomainExpression.MatchString(value) {
		errorReporter.ReportError(ErrorValueStringAsReverseDomainNotValid(value))
	} else if length := len(value); length > ReverseDomainLengthMaximum {
		errorReporter.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, ReverseDomainLengthMaximum))
	}
}

func SemanticVersion(value string, errorReporter structure.ErrorReporter) {
	if value == "" {
		errorReporter.ReportError(structureValidator.ErrorValueEmpty())
	} else if _, err := semver.Parse(value); err != nil {
		errorReporter.ReportError(ErrorValueStringAsSemanticVersionNotValid(value))
	} else if length := len(value); length > SemanticVersionLengthMaximum {
		errorReporter.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, SemanticVersionLengthMaximum))
	}
}

func URL(value string, errorReporter structure.ErrorReporter) {
	if value == "" {
		errorReporter.ReportError(structureValidator.ErrorValueEmpty())
	} else if urlValue, err := url.Parse(value); err != nil || urlValue == nil || !urlValue.IsAbs() || urlValue.Host == "" {
		errorReporter.ReportError(ErrorValueStringAsURLNotValid(value))
	} else if length := len(value); length > URLLengthMaximum {
		errorReporter.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, URLLengthMaximum))
	}
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
