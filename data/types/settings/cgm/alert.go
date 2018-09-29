package cgm

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type Alerts struct {
	Enabled            *bool            `json:"enabled,omitempty" bson:"enabled,omitempty"`
	UrgentLow          *UrgentLowAlert  `json:"urgentLow,omitempty" bson:"urgentLow,omitempty"`
	UrgentLowPredicted *UrgentLowAlert  `json:"urgentLowPredicted,omitempty" bson:"urgentLowPredicted,omitempty"`
	Low                *LowAlert        `json:"low,omitempty" bson:"low,omitempty"`
	LowPredicted       *LowAlert        `json:"lowPredicted,omitempty" bson:"lowPredicted,omitempty"`
	High               *HighAlert       `json:"high,omitempty" bson:"high,omitempty"`
	HighPredicted      *HighAlert       `json:"highPredicted,omitempty" bson:"highPredicted,omitempty"`
	Fall               *FallAlert       `json:"fall,omitempty" bson:"fall,omitempty"`
	Rise               *RiseAlert       `json:"rise,omitempty" bson:"rise,omitempty"`
	NoData             *NoDataAlert     `json:"noData,omitempty" bson:"noData,omitempty"`
	OutOfRange         *OutOfRangeAlert `json:"outOfRange,omitempty" bson:"outOfRange,omitempty"`
}

func ParseAlerts(parser data.ObjectParser) *Alerts {
	if parser.Object() == nil {
		return nil
	}
	datum := NewAlerts()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewAlerts() *Alerts {
	return &Alerts{}
}

func (a *Alerts) Parse(parser data.ObjectParser) {
	a.Enabled = parser.ParseBoolean("enabled")
	a.UrgentLow = ParseUrgentLowAlert(parser.NewChildObjectParser("urgentLow"))
	a.UrgentLowPredicted = ParseUrgentLowAlert(parser.NewChildObjectParser("urgentLowPredicted"))
	a.Low = ParseLowAlert(parser.NewChildObjectParser("low"))
	a.LowPredicted = ParseLowAlert(parser.NewChildObjectParser("lowPredicted"))
	a.High = ParseHighAlert(parser.NewChildObjectParser("high"))
	a.HighPredicted = ParseHighAlert(parser.NewChildObjectParser("highPredicted"))
	a.Fall = ParseFallAlert(parser.NewChildObjectParser("fall"))
	a.Rise = ParseRiseAlert(parser.NewChildObjectParser("rise"))
	a.NoData = ParseNoDataAlert(parser.NewChildObjectParser("noData"))
	a.OutOfRange = ParseOutOfRangeAlert(parser.NewChildObjectParser("outOfRange"))
}

func (a *Alerts) Validate(validator structure.Validator) {
	validator.Bool("enabled", a.Enabled).Exists()
	if a.UrgentLow != nil {
		a.UrgentLow.Validate(validator.WithReference("urgentLow"))
	}
	if a.UrgentLowPredicted != nil {
		a.UrgentLowPredicted.Validate(validator.WithReference("urgentLowPredicted"))
	}
	if a.Low != nil {
		a.Low.Validate(validator.WithReference("low"))
	}
	if a.LowPredicted != nil {
		a.LowPredicted.Validate(validator.WithReference("lowPredicted"))
	}
	if a.High != nil {
		a.High.Validate(validator.WithReference("high"))
	}
	if a.HighPredicted != nil {
		a.HighPredicted.Validate(validator.WithReference("highPredicted"))
	}
	if a.Fall != nil {
		a.Fall.Validate(validator.WithReference("fall"))
	}
	if a.Rise != nil {
		a.Rise.Validate(validator.WithReference("rise"))
	}
	if a.NoData != nil {
		a.NoData.Validate(validator.WithReference("noData"))
	}
	if a.OutOfRange != nil {
		a.OutOfRange.Validate(validator.WithReference("outOfRange"))
	}
}

type Alert struct {
	Enabled *bool   `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Snooze  *Snooze `json:"snooze,omitempty" bson:"snooze,omitempty"`
}

func (a *Alert) Parse(parser data.ObjectParser) {
	a.Enabled = parser.ParseBoolean("enabled")
	a.Snooze = ParseSnooze(parser.NewChildObjectParser("snooze"))
}

func (a *Alert) Validate(validator structure.Validator) {
	validator.Bool("enabled", a.Enabled).Exists()
	if a.Snooze != nil {
		a.Snooze.Validate(validator.WithReference("snooze"))
	}
}
