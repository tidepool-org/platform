package selfmonitored

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubTypeLinked = "linked"
	SubTypeManual = "manual"
)

func SubTypes() []string {
	return []string{
		SubTypeLinked,
		SubTypeManual,
	}
}

type SelfMonitored struct {
	glucose.Glucose `bson:",inline"`

	SubType *string `json:"subType,omitempty" bson:"subType,omitempty"`
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
	s.Glucose.Init()
	s.Type = Type()

	s.SubType = nil
}

func (s *SelfMonitored) Parse(parser data.ObjectParser) error {
	if err := s.Glucose.Parse(parser); err != nil {
		return err
	}

	s.SubType = parser.ParseString("subType")

	return nil
}

func (s *SelfMonitored) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(s.Meta())
	}

	s.Glucose.Validate(validator)

	if s.Type != "" {
		validator.String("type", &s.Type).EqualTo(Type())
	}

	validator.String("subType", s.SubType).OneOf(SubTypes()...)
}

func (s *SelfMonitored) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(s.Meta())
	}

	s.Glucose.Normalize(normalizer)
}
