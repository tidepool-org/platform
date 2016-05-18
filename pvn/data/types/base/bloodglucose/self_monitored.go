package bloodglucose

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base"
)

type SelfMonitored struct {
	base.Base `bson:",inline"`

	Value *float64 `json:"value" bson:"value"`
	Units *string  `json:"units" bson:"units"`
}

func SelfMonitoredType() string {
	return "smbg"
}

func NewSelfMonitored() *SelfMonitored {
	bloodGlucoseType := SelfMonitoredType()

	selfMonitored := &SelfMonitored{}
	selfMonitored.Type = &bloodGlucoseType
	return selfMonitored
}

func (s *SelfMonitored) Parse(parser data.ObjectParser) {
	s.Base.Parse(parser)

	s.Value = parser.ParseFloat("value")
	s.Units = parser.ParseString("units")
}

func (s *SelfMonitored) Validate(validator data.Validator) {
	s.Base.Validate(validator)

	validator.ValidateFloat("value", s.Value).Exists().InRange(0.0, 1000.0)
	validator.ValidateString("units", s.Units).Exists().OneOf([]string{"mmol/l", "mmol/L", "mg/dl", "mg/dL"})

}

func (s *SelfMonitored) Normalize(normalizer data.Normalizer) {
	s.Base.Normalize(normalizer)
}
