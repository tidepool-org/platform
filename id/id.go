package id

import (
	"regexp"
	"strings"

	uuid "github.com/satori/go.uuid"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var expression = regexp.MustCompile("^[0-9a-z]{32}$")

func New() string {
	return strings.Replace(uuid.Must(uuid.NewV4()).String(), "-", "", -1)
}

func Validate(value string, errorReporter structure.ErrorReporter) {
	if value == "" {
		errorReporter.ReportError(structureValidator.ErrorValueEmpty())
	} else if !expression.MatchString(value) {
		errorReporter.ReportError(ErrorValueStringAsIDNotValid(value))
	}
}

func ErrorValueStringAsIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as id", value)
}
