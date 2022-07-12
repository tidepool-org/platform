package dexcom

import (
	"regexp"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	TimeFormat             = "2006-01-02T15:04:05"
	SystemTimeNowThreshold = 24 * time.Hour
)

func IsValidTransmitterID(value string) bool {
	return ValidateTransmitterID(value) == nil
}

func TransmitterIDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateTransmitterID(value))
}

func ValidateTransmitterID(value string) error {
	if value == "" {
		// dexcom started sending empty transmitter id on a very small portion of users which recently were created
		//return structureValidator.ErrorValueEmpty()
		return nil
	} else if !transmitterIDExpression.MatchString(value) {
		return ErrorValueStringAsTransmitterIDNotValid(value)
	}
	return nil
}

func ErrorValueStringAsTransmitterIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as transmitter id", value)
}

var transmitterIDExpression = regexp.MustCompile("^[0-9A-Z]{5,6}$")
