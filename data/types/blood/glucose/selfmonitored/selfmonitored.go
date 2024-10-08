package selfmonitored

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
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

	s.Glucose.Normalize(normalizer)
}

func (s *SelfMonitored) LegacyIdentityFields() ([]string, error) {
	identityFields, err := s.Blood.LegacyIdentityFields()
	if err != nil {
		return nil, err
	}
	value, units, err := s.GetRawValueAndUnits()
	if err != nil {
		return nil, err
	}
	fullPrecisionValue := dataBloodGlucose.NormalizeValueForUnitsWithFullPrecision(value, units)
	identityFields = append(identityFields, strconv.FormatFloat(*fullPrecisionValue, 'f', -1, 64))
	return identityFields, nil
}
