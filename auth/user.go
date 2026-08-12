package auth

import (
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

func NewUserID() string {
	return id.Must(id.New(5))
}

func IsValidUserID(value string) bool {
	return ValidateUserID(value) == nil
}

func UserIDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateUserID(value))
}

func ValidateUserID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !user.IdExpression.MatchString(value) {
		return ErrorValueStringAsUserIDNotValid(value)
	}
	return nil
}

func ErrorValueStringAsUserIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as user id", value)
}
