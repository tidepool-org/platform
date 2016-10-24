package selfmonitored

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base"
)

type SelfMonitored struct {
	base.Base `bson:",inline"`

	Value   *float64 `json:"value,omitempty" bson:"value,omitempty"`
	Units   *string  `json:"units,omitempty" bson:"units,omitempty"`
	SubType *string  `json:"subType,omitempty" bson:"subType,omitempty"`
}

func Type() string {
	return "smbg"
}

func NewDatum() data.Datum {
	return New()
}

func New() *SelfMonitored {
	return &SelfMonitored{}
}

func Init() *SelfMonitored {
	selfMonitored := New()
	selfMonitored.Init()
	return selfMonitored
}

func (s *SelfMonitored) Init() {
	s.Base.Init()
	s.Base.Type = Type()

	s.Value = nil
	s.Units = nil
	s.SubType = nil
}

func (s *SelfMonitored) Parse(parser data.ObjectParser) error {
	parser.SetMeta(s.Meta())

	if err := s.Base.Parse(parser); err != nil {
		return err
	}

	s.Value = parser.ParseFloat("value")
	s.Units = parser.ParseString("units")
	s.SubType = parser.ParseString("subType")

	return nil
}

func (s *SelfMonitored) Validate(validator data.Validator) error {
	validator.SetMeta(s.Meta())

	if err := s.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateStringAsBloodGlucoseUnits("units", s.Units).Exists()
	validator.ValidateFloatAsBloodGlucoseValue("value", s.Value).Exists().InRangeForUnits(s.Units)
	validator.ValidateString("subType", s.SubType).OneOf([]string{"manual", "linked"})

	return nil
}

func (s *SelfMonitored) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(s.Meta())

	if err := s.Base.Normalize(normalizer); err != nil {
		return err
	}

	s.Units, s.Value = normalizer.NormalizeBloodGlucose(s.Units).UnitsAndValue(s.Value)

	return nil
}
