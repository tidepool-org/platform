package cgm

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

func OutOfRangeAlertDEPRECATEDThresholds() []int {
	return []int{
		0, 60000, 1200000, 1500000, 1800000, 2100000, 2400000, 2700000, 3000000, 3300000,
		3600000, 3900000, 4200000, 4500000, 4800000, 5100000, 5400000, 5700000,
		6000000, 6300000, 6600000, 6900000, 7200000, 7500000, 7800000, 8100000,
		8400000, 8700000, 9000000, 9300000, 9600000, 9900000, 10200000,
		10500000, 10800000, 11100000, 11400000, 11700000, 12000000, 12300000,
		12600000, 12900000, 13200000, 13500000, 13800000, 14100000, 14400000,
	}
}

type OutOfRangeAlertDEPRECATED struct {
	Enabled   *bool `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Threshold *int  `json:"snooze,omitempty" bson:"snooze,omitempty"`
}

func ParseOutOfRangeAlertDEPRECATED(parser structure.ObjectParser) *OutOfRangeAlertDEPRECATED {
	if !parser.Exists() {
		return nil
	}
	datum := NewOutOfRangeAlertDEPRECATED()
	parser.Parse(datum)
	return datum
}

func NewOutOfRangeAlertDEPRECATED() *OutOfRangeAlertDEPRECATED {
	return &OutOfRangeAlertDEPRECATED{}
}

func (o *OutOfRangeAlertDEPRECATED) Parse(parser structure.ObjectParser) {
	o.Enabled = parser.Bool("enabled")
	o.Threshold = parser.Int("snooze")
}

func (o *OutOfRangeAlertDEPRECATED) Validate(validator structure.Validator) {
	validator.Bool("enabled", o.Enabled).Exists()
	validator.Int("snooze", o.Threshold).Exists().OneOf(OutOfRangeAlertDEPRECATEDThresholds()...)
}

func (o *OutOfRangeAlertDEPRECATED) Normalize(normalizer data.Normalizer) {}
