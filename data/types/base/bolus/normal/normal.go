package normal

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base/bolus"
)

type Normal struct {
	bolus.Bolus `bson:",inline"`

	Normal         *float64 `json:"normal,omitempty" bson:"normal,omitempty"`
	ExpectedNormal *float64 `json:"expectedNormal,omitempty" bson:"expectedNormal,omitempty"`
}

func SubType() string {
	return "normal"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Normal {
	return &Normal{}
}

func Init() *Normal {
	normal := New()
	normal.Init()
	return normal
}

func (n *Normal) Init() {
	n.Bolus.Init()
	n.Bolus.SubType = SubType()

	n.Normal = nil
	n.ExpectedNormal = nil
}

func (n *Normal) Parse(parser data.ObjectParser) error {
	if err := n.Bolus.Parse(parser); err != nil {
		return err
	}

	n.Normal = parser.ParseFloat("normal")
	n.ExpectedNormal = parser.ParseFloat("expectedNormal")

	return nil
}

func (n *Normal) Validate(validator data.Validator) error {
	if err := n.Bolus.Validate(validator); err != nil {
		return err
	}

	validator.ValidateFloat("normal", n.Normal).Exists().InRange(0.0, 100.0)

	expectedNormalValidator := validator.ValidateFloat("expectedNormal", n.ExpectedNormal)
	if n.Normal != nil {
		if *n.Normal == 0.0 {
			expectedNormalValidator.Exists()
		}
		expectedNormalValidator.GreaterThan(*n.Normal).LessThanOrEqualTo(100.0)
	} else {
		expectedNormalValidator.InRange(0.0, 100.0)
	}

	return nil
}
