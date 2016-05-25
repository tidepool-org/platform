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

	Normal *float64 `json:"normal,omitempty" bson:"normal,omitempty"`
}

func SubType() string {
	return "normal"
}

func New() (*Normal, error) {
	normalBolus, err := bolus.New(SubType())
	if err != nil {
		return nil, err
	}

	return &Normal{
		Bolus: *normalBolus,
	}, nil
}

func (n *Normal) Parse(parser data.ObjectParser) {
	n.Bolus.Parse(parser)

	n.Normal = parser.ParseFloat("normal")
}

func (n *Normal) Validate(validator data.Validator) {
	n.Bolus.Validate(validator)

	validator.ValidateFloat("normal", n.Normal).Exists().InRange(0.0, 100.0)
}
