package cgm

import (
	"math"

	"github.com/tidepool-org/platform/structure"
)

const (
	DurationAlertUnitsHours   = "hours"
	DurationAlertUnitsMinutes = "minutes"
	DurationAlertUnitsSeconds = "seconds"

	NoDataAlertDurationHoursMaximum       = 6.0
	NoDataAlertDurationHoursMinimum       = 0.0
	NoDataAlertDurationMinutesMaximum     = NoDataAlertDurationHoursMaximum * 60.0
	NoDataAlertDurationMinutesMinimum     = NoDataAlertDurationHoursMinimum * 60.0
	NoDataAlertDurationSecondsMaximum     = NoDataAlertDurationMinutesMaximum * 60.0
	NoDataAlertDurationSecondsMinimum     = NoDataAlertDurationMinutesMinimum * 60.0
	OutOfRangeAlertDurationHoursMaximum   = 6.0
	OutOfRangeAlertDurationHoursMinimum   = 0.0
	OutOfRangeAlertDurationMinutesMaximum = OutOfRangeAlertDurationHoursMaximum * 60.0
	OutOfRangeAlertDurationMinutesMinimum = OutOfRangeAlertDurationHoursMinimum * 60.0
	OutOfRangeAlertDurationSecondsMaximum = OutOfRangeAlertDurationMinutesMaximum * 60.0
	OutOfRangeAlertDurationSecondsMinimum = OutOfRangeAlertDurationMinutesMinimum * 60.0
)

func DurationAlertUnits() []string {
	return []string{
		DurationAlertUnitsHours,
		DurationAlertUnitsMinutes,
		DurationAlertUnitsSeconds,
	}
}

type DurationAlert struct {
	Alert    `bson:",inline"`
	Duration *float64 `json:"duration,omitempty" bson:"duration,omitempty"`
	Units    *string  `json:"units,omitempty" bson:"units,omitempty"`
}

func (d *DurationAlert) Parse(parser structure.ObjectParser) {
	d.Alert.Parse(parser)
	d.Duration = parser.Float64("duration")
	d.Units = parser.String("units")
}

func (d *DurationAlert) Validate(validator structure.Validator) {
	d.Alert.Validate(validator)
	if unitsValidator := validator.String("units", d.Units); d.Duration != nil {
		unitsValidator.Exists().OneOf(DurationAlertUnits()...)
	} else {
		unitsValidator.NotExists()
	}
}

type NoDataAlert struct {
	DurationAlert `bson:",inline"`
}

func ParseNoDataAlert(parser structure.ObjectParser) *NoDataAlert {
	if !parser.Exists() {
		return nil
	}
	datum := NewNoDataAlert()
	parser.Parse(datum)
	return datum
}

func NewNoDataAlert() *NoDataAlert {
	return &NoDataAlert{}
}

func (n *NoDataAlert) Validate(validator structure.Validator) {
	n.DurationAlert.Validate(validator)
	validator.Float64("duration", n.Duration).InRange(NoDataAlertDurationRangeForUnits(n.Units))
}

func NoDataAlertDurationRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case DurationAlertUnitsHours:
			return NoDataAlertDurationHoursMinimum, NoDataAlertDurationHoursMaximum
		case DurationAlertUnitsMinutes:
			return NoDataAlertDurationMinutesMinimum, NoDataAlertDurationMinutesMaximum
		case DurationAlertUnitsSeconds:
			return NoDataAlertDurationSecondsMinimum, NoDataAlertDurationSecondsMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}

type OutOfRangeAlert struct {
	DurationAlert `bson:",inline"`
}

func ParseOutOfRangeAlert(parser structure.ObjectParser) *OutOfRangeAlert {
	if !parser.Exists() {
		return nil
	}
	datum := NewOutOfRangeAlert()
	parser.Parse(datum)
	return datum
}

func NewOutOfRangeAlert() *OutOfRangeAlert {
	return &OutOfRangeAlert{}
}

func (o *OutOfRangeAlert) Validate(validator structure.Validator) {
	o.DurationAlert.Validate(validator)
	validator.Float64("duration", o.Duration).InRange(OutOfRangeAlertDurationRangeForUnits(o.Units))
}

func OutOfRangeAlertDurationRangeForUnits(units *string) (float64, float64) {
	if units != nil {
		switch *units {
		case DurationAlertUnitsHours:
			return OutOfRangeAlertDurationHoursMinimum, OutOfRangeAlertDurationHoursMaximum
		case DurationAlertUnitsMinutes:
			return OutOfRangeAlertDurationMinutesMinimum, OutOfRangeAlertDurationMinutesMaximum
		case DurationAlertUnitsSeconds:
			return OutOfRangeAlertDurationSecondsMinimum, OutOfRangeAlertDurationSecondsMaximum
		}
	}
	return -math.MaxFloat64, math.MaxFloat64
}
