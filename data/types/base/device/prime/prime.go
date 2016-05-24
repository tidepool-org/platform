package prime

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
	"github.com/tidepool-org/platform/data/types/base/device"
)

type Prime struct {
	device.Device `bson:",inline"`

	Target *string  `json:"primeTarget,omitempty" bson:"primeTarget,omitempty"`
	Volume *float64 `json:"volume,omitempty" bson:"volume,omitempty"`
}

func SubType() string {
	return "prime"
}

func New() (*Prime, error) {
	primtDevice, err := device.New(SubType())
	if err != nil {
		return nil, err
	}

	return &Prime{
		Device: *primtDevice,
	}, nil
}

func (p *Prime) Parse(parser data.ObjectParser) {
	p.Device.Parse(parser)

	p.Target = parser.ParseString("primeTarget")
	p.Volume = parser.ParseFloat("volume")
}

func (p *Prime) Validate(validator data.Validator) {
	p.Device.Validate(validator)

	validator.ValidateString("primeTarget", p.Target).Exists().OneOf([]string{"cannula", "tubing"})

	if p.Target != nil {
		if *p.Target == "cannula" {
			validator.ValidateFloat("volume", p.Volume).InRange(0.0, 3.0)
		} else if *p.Target == "tubing" {
			validator.ValidateFloat("volume", p.Volume).InRange(0.0, 100.0)
		}
	}
}
