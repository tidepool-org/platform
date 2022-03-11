package flush

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = "flush"

	Succeeded           = "success"
	Failed              = "failure"
	VolumeTargetMaximum = 10.0
	VolumeTargetMinimum = 0.0
)

func Statuses() []string {
	return []string{
		Succeeded,
		Failed,
	}
}

func StatusCodes() []int {
	return []int{
		0,
		1,
		2,
		3,
		4,
		5,
		6,
		7,
	}
}

type Flush struct {
	device.Device `bson:",inline"`

	Status     *string  `json:"status,omitempty" bson:"status,omitempty"`
	StatusCode *int     `json:"statusCode,omitempty" bson:"statusCode,omitempty"`
	Volume     *float64 `json:"volume,omitempty" bson:"volume,omitempty"`
}

func New() *Flush {
	return &Flush{
		Device: device.New(SubType),
	}
}

func (p *Flush) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(p.Meta())
	}

	p.Device.Parse(parser)

	p.Status = parser.String("status")
	p.StatusCode = parser.Int("statusCode")
	p.Volume = parser.Float64("volume")
}

func (p *Flush) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(p.Meta())
	}

	p.Device.Validate(validator)

	if p.SubType != "" {
		validator.String("subType", &p.SubType).EqualTo(SubType)
	}

	validator.String("status", p.Status).Exists().OneOf(Statuses()...)
	validator.Float64("volume", p.Volume).Exists().InRange(VolumeTargetMinimum, VolumeTargetMaximum)
	validator.Int("statusCode", p.StatusCode).Exists()
}

// IsValid returns true if there is no error in the validator
func (p *Flush) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (p *Flush) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(p.Meta())
	}

	p.Device.Normalize(normalizer)
}
