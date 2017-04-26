package prime

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
)

type Prime struct {
	device.Device `bson:",inline"`

	Target *string  `json:"primeTarget,omitempty" bson:"primeTarget,omitempty"`
	Volume *float64 `json:"volume,omitempty" bson:"volume,omitempty"`
}

func SubType() string {
	return "prime"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Prime {
	return &Prime{}
}

func Init() *Prime {
	prime := New()
	prime.Init()
	return prime
}

func (p *Prime) Init() {
	p.Device.Init()
	p.SubType = SubType()

	p.Target = nil
	p.Volume = nil
}

func (p *Prime) Parse(parser data.ObjectParser) error {
	if err := p.Device.Parse(parser); err != nil {
		return err
	}

	p.Target = parser.ParseString("primeTarget")
	p.Volume = parser.ParseFloat("volume")

	return nil
}

func (p *Prime) Validate(validator data.Validator) error {
	if err := p.Device.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("subType", &p.SubType).EqualTo(SubType())

	validator.ValidateString("primeTarget", p.Target).Exists().OneOf([]string{"cannula", "tubing"})

	if p.Target != nil {
		if *p.Target == "cannula" {
			validator.ValidateFloat("volume", p.Volume).InRange(0.0, 3.0)
		} else if *p.Target == "tubing" {
			validator.ValidateFloat("volume", p.Volume).InRange(0.0, 100.0)
		}
	}

	return nil
}
