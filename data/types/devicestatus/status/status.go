package status

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	StatusArrayMaxLength = 1000
)

type TypeStatusArray []*Status

type Status struct {
	Battery            *Battery            `json:"battery,omitempty" bson:"battery,omitempty"`
	SignalStrength     *SignalStrength     `json:"signalStrength,omitempty" bson:"signalStrength,omitempty"`
	ReservoirRemaining *ReservoirRemaining `json:"reservoirRemaining,omitempty" bson:"reservoirRemaining,omitempty"`
	Forecast           *data.Forecast      `json:"forecast,omitempty" bson:"forecast,omitempty"`
	Alerts             *[]string           `json:"alerts,omitempty" bson:"alerts,omitempty"`
}

func ParseStatus(parser structure.ObjectParser) *Status {
	if !parser.Exists() {
		return nil
	}
	datum := NewStatus()
	parser.Parse(datum)
	return datum
}

func NewStatus() *Status {
	return &Status{}
}

func (c *Status) Parse(parser structure.ObjectParser) {
	c.Battery = ParseBattery(parser.WithReferenceObjectParser("battery"))
	c.Alerts = parser.StringArray("alerts")
	c.ReservoirRemaining = ParseReservoirRemaining(parser.WithReferenceObjectParser("reservoirRemaining"))
	c.SignalStrength = ParseSignalStrength(parser.WithReferenceObjectParser("signalStrength"))
	c.Forecast = data.ParseForecast(parser.WithReferenceObjectParser("forecast"))
}

func (c *Status) Validate(validator structure.Validator) {
	if c.Battery != nil {
		c.Battery.Validate(validator.WithReference("battery"))
	}
	if c.ReservoirRemaining != nil {
		c.ReservoirRemaining.Validate(validator.WithReference("reservoirRemaining"))
	}
	if c.SignalStrength != nil {
		c.SignalStrength.Validate(validator.WithReference("signalStrength"))
	}
	if c.Forecast != nil {
		c.Forecast.Validate(validator.WithReference("forecast"))
	}
}

func (c *Status) Normalize(normalizer data.Normalizer) {
	if c.Battery != nil {
		c.Battery.Normalize(normalizer.WithReference("battery"))
	}
	if c.SignalStrength != nil {
		c.SignalStrength.Normalize(normalizer.WithReference("signalStrength"))
	}
	if c.ReservoirRemaining != nil {
		c.ReservoirRemaining.Normalize(normalizer.WithReference("reservoirRemaining"))
	}
	if c.Forecast != nil {
		c.Forecast.Normalize(normalizer.WithReference("forecast"))
	}
}

func ParseStatusArray(parser structure.ArrayParser) *TypeStatusArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewParseStatusArray()
	parser.Parse(datum)
	return datum
}

func NewParseStatusArray() *TypeStatusArray {
	return &TypeStatusArray{}
}

func (t *TypeStatusArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*t = append(*t, ParseStatus(parser.WithReferenceObjectParser(reference)))
	}
}

func (t *TypeStatusArray) Validate(validator structure.Validator) {
	if length := len(*t); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	} else if length > StatusArrayMaxLength {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, StatusArrayMaxLength))
	}
	for index, datum := range *t {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (t *TypeStatusArray) Normalize(normalizer data.Normalizer) {
	for index, datum := range *t {
		if datum != nil {
			datum.Normalize(normalizer.WithReference(strconv.Itoa(index)))
		}
	}
}
