package dexcom

import (
	"regexp"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	TimeFormatMilli    = "2006-01-02T15:04:05.999"
	TimeFormatMilliUTC = "2006-01-02T15:04:05.999Z"
	TimeFormatMilliZ   = "2006-01-02T15:04:05.999-07:00"

	TimeFormat    = "2006-01-02T15:04:05"
	TimeFormatZ   = "2006-01-02T15:04:05-07:00"
	TimeFormatUTC = "2006-01-02T15:04:05Z"

	SystemTimeNowThreshold = 24 * time.Hour
)

func timeFormats() map[int]string {
	return map[int]string{
		len(TimeFormat):         TimeFormat,
		len(TimeFormatMilli):    TimeFormatMilli,
		len(TimeFormatZ):        TimeFormatZ,
		len(TimeFormatMilliZ):   TimeFormatMilliZ,
		len(TimeFormatMilliUTC): TimeFormatMilliUTC,
		len(TimeFormatUTC):      TimeFormatUTC,
	}
}

func GetTimeFormat(timeStr string) string {
	return timeFormats()[len(timeStr)]
}

func IsValidTransmitterID(value string) bool {
	return ValidateTransmitterID(value) == nil
}

func TransmitterIDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateTransmitterID(value))
}

func ValidateTransmitterID(value string) error {
	if value == "" {
		// dexcom started sending empty transmitter id on a very small portion of users which recently were created
		// and form the v3 api is now in the format of a hash
		// "transmitterId": "cdb4f8eea4392295413c64d5bc7a9e0e0ee9b215fb43c5a6d71d4431e540046b",
		return nil
	} else if !transmitterIDExpression.MatchString(value) {
		return ErrorValueStringAsTransmitterIDNotValid(value)
	}
	return nil
}

func ErrorValueStringAsTransmitterIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as transmitter id", value)
}

var transmitterIDExpression = regexp.MustCompile("^[0-9a-z]{64}$")
