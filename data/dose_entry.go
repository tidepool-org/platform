package data

import (
	"time"

	"github.com/tidepool-org/platform/structure"
)

const (
	MinValue = 0
	MaxValue = 1000

	MinDeliveredUnits = 0
	MaxDeliveredUnits = 1000

	MinBasalRate = 0
	MaxBasalRate = 1000
)

func DoseTypes() []string {
	return []string{
		"basal",
		"bolus",
		"resume",
		"suspend",
		"tempBasal",
	}
}

func DoseUnits() []string {
	return []string{
		"units",
		"unitsPerHour",
	}
}

type DoseEntry struct {
	DoseType           *string  `json:"doseType,omitempty" bson:"doseType,omitempty"`
	StartDate          *string  `json:"startDate,omitempty" bson:"startDate,omitempty"`
	EndDate            *string  `json:"endDate,omitempty" bson:"endDate,omitempty"`
	Value              *float64 `json:"value,omitempty" bson:"value,omitempty"`
	Unit               *string  `json:"unit,omitempty" bson:"unit,omitempty"`
	DeliveredUnits     *float64 `json:"deliveredUnits,omitempty" bson:"deliveredUnits,omitempty"`
	Description        *string  `json:"description,omitempty" bson:"description,omitempty"`
	SyncIdentifier     *string  `json:"syncIdentifier,omitempty" bson:"syncIdentifier,omitempty"`
	ScheduledBasalRate *float64 `json:"scheduledBasalRate,omitempty" bson:"scheduledBasalRate,omitempty"`
}

func ParseDoseEntry(parser structure.ObjectParser) *DoseEntry {
	if !parser.Exists() {
		return nil
	}
	datum := NewDoseEntry()
	parser.Parse(datum)
	return datum
}

func NewDoseEntry() *DoseEntry {
	return &DoseEntry{}
}

func (f *DoseEntry) Parse(parser structure.ObjectParser) {
	f.DoseType = parser.String("doseType")
	f.StartDate = parser.String("startDate")
	f.EndDate = parser.String("endDate")
	f.Value = parser.Float64("value")
	f.Unit = parser.String("unit")
	f.DeliveredUnits = parser.Float64("deliveredUnits")
	f.Description = parser.String("description")
	f.SyncIdentifier = parser.String("syncIdentifier")
	f.ScheduledBasalRate = parser.Float64("scheduledBasalRate")
}

func (f *DoseEntry) Validate(validator structure.Validator) {
	var startDate time.Time

	validator.String("doseType", f.DoseType).Exists().OneOf(DoseTypes()...)
	validator.String("unit", f.Unit).Exists().OneOf(DoseUnits()...)
	if f.StartDate != nil {
		validator.String("startDate", f.StartDate).AsTime(time.RFC3339Nano)
		startDate, _ = time.Parse(time.RFC3339Nano, *f.StartDate)
	}
	if f.EndDate != nil {
		validator.String("endDate", f.EndDate).AsTime(time.RFC3339Nano).After(startDate)
	}
	if f.Value != nil {
		validator.Float64("value", f.Value).InRange(MinValue, MaxValue)
	}
	if f.DeliveredUnits != nil {
		validator.Float64("deliveredUnits", f.DeliveredUnits).InRange(MinDeliveredUnits, MaxDeliveredUnits)
	}
	if f.ScheduledBasalRate != nil {
		validator.Float64("scheduledBasalRate", f.ScheduledBasalRate).InRange(MinBasalRate, MaxBasalRate)
	}
}

func (f *DoseEntry) Normalize(normalizer Normalizer) {}
