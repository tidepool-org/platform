package cgm

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

func OutOfRangeAlertThresholds() []int {
	return []int{
		1200000, 1500000, 1800000, 2100000, 2400000, 2700000, 3000000, 3300000,
		3600000, 3900000, 4200000, 4500000, 4800000, 5100000, 5400000, 5700000,
		6000000, 6300000, 6600000, 6900000, 7200000, 7500000, 7800000, 8100000,
		8400000, 8700000, 9000000, 9300000, 9600000, 9900000, 10200000,
		10500000, 10800000, 11100000, 11400000, 11700000, 12000000, 12300000,
		12600000, 12900000, 13200000, 13500000, 13800000, 14100000, 14400000,
	}
}

type OutOfRangeAlert struct {
	Enabled   *bool `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Threshold *int  `json:"snooze,omitempty" bson:"snooze,omitempty"` // TODO: Rename threshold
}

func ParseOutOfRangeAlert(parser data.ObjectParser) *OutOfRangeAlert {
	if parser.Object() == nil {
		return nil
	}
	outOfRangeAlert := NewOutOfRangeAlert()
	outOfRangeAlert.Parse(parser)
	parser.ProcessNotParsed()
	return outOfRangeAlert
}

func NewOutOfRangeAlert() *OutOfRangeAlert {
	return &OutOfRangeAlert{}
}

func (o *OutOfRangeAlert) Parse(parser data.ObjectParser) {
	o.Enabled = parser.ParseBoolean("enabled")
	o.Threshold = parser.ParseInteger("snooze")
}

func (o *OutOfRangeAlert) Validate(validator structure.Validator) {
	validator.Bool("enabled", o.Enabled).Exists()
	validator.Int("snooze", o.Threshold).Exists().OneOf(OutOfRangeAlertThresholds()...)
}

func (o *OutOfRangeAlert) Normalize(normalizer data.Normalizer) {}
