package selfmonitored

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "smbg"

	SubTypeLinked  = "linked"
	SubTypeManual  = "manual"
	SubTypeScanned = "scanned"
)

func SubTypes() []string {
	return []string{
		SubTypeLinked,
		SubTypeManual,
		SubTypeScanned,
	}
}

type SelfMonitored struct {
	glucose.Glucose `bson:",inline"`

	SubType *string `json:"subType,omitempty" bson:"subType,omitempty"`
}

func New() *SelfMonitored {
	return &SelfMonitored{
		Glucose: glucose.New(Type),
	}
}

func (s *SelfMonitored) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(s.Meta())
	}

	s.Glucose.Parse(parser)

	s.SubType = parser.String("subType")
}

func (s *SelfMonitored) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(s.Meta())
	}

	s.Glucose.Validate(validator)

	if s.Type != "" {
		validator.String("type", &s.Type).EqualTo(Type)
	}

	validator.String("subType", s.SubType).OneOf(SubTypes()...)
}

func (s *SelfMonitored) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(s.Meta())
	}

	if normalizer.Origin() == structure.OriginExternal {

		rawUnits := pointer.CloneString(s.Units)
		rawValue := pointer.CloneFloat64(s.Value)
		s.SetRawUnitsAndValue(rawUnits, rawValue)

		s.Units = dataBloodGlucose.NormalizeUnits(rawUnits)
		s.Value = dataBloodGlucose.NormalizeValueForUnits(rawValue, rawUnits)
	}
}

func (s *SelfMonitored) LegacyIdentityFields() ([]string, error) {
	identityFields, err := s.Blood.LegacyIdentityFields()
	if err != nil {
		return nil, err
	}
	units, value, err := s.GetRawUnitsAndValue()
	if err != nil {
		return nil, err
	}
	fullPrecisionValue := dataBloodGlucose.NormalizeValueForUnitsWithFullPrecision(value, units)
	return append(identityFields, strconv.FormatFloat(*fullPrecisionValue, 'f', -1, 64)), nil
}
