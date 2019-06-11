package blob

import (
	"github.com/tidepool-org/platform/errors"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func ErrorValueStringAsIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as blob id", value)
}
