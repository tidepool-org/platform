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
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/device"
)

type Prime struct {
	device.Device

	Target *string  `json:"primeTarget" bson:"primeTarget"`
	Volume *float64 `json:"volume,omitempty" bson:"volume,omitempty"`
}

func Type() string {
	return device.Type()
}

func SubType() string {
	return "prime"
}

func New() *Prime {
	primeType := Type()
	primeSubType := SubType()

	prime := &Prime{}
	prime.Type = &primeType
	prime.SubType = &primeSubType
	return prime
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

func (p *Prime) Normalize(normalizer data.Normalizer) {
	p.Device.Normalize(normalizer)
}
