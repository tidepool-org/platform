package selfmonitored

import (
	"errors"
	"strconv"

	"github.com/tidepool-org/platform/data"
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

	if s.Value == nil {
		return nil, errors.New("value is missing")
	}

	return append(identityFields, strconv.FormatFloat(*s.Value, 'f', -1, 64)), nil
}
