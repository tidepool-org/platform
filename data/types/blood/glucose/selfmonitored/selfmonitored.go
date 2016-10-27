package selfmonitored

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
)

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

func (s *SelfMonitored) Validate(validator data.Validator) error {
	if err := s.Glucose.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("type", &s.Type).EqualTo(Type())

	validator.ValidateString("subType", s.SubType).OneOf([]string{"linked", "manual"})

	return nil
}
