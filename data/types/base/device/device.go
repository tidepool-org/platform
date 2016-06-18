package device

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
	"github.com/tidepool-org/platform/data/types/base"
)

type Device struct {
	base.Base `bson:",inline"`

	SubType string `json:"subType,omitempty" bson:"subType,omitempty"`
}

type Meta struct {
	Type    string `json:"type,omitempty"`
	SubType string `json:"subType,omitempty"`
}

func Type() string {
	return "deviceEvent"
}

func (d *Device) Init() {
	d.Base.Init()
	d.Base.Type = Type()

	d.SubType = ""
}

func (d *Device) Meta() interface{} {
	return &Meta{
		Type:    d.Type,
		SubType: d.SubType,
	}
}

func (d *Device) Parse(parser data.ObjectParser) error {
	parser.SetMeta(d.Meta())

	return d.Base.Parse(parser)
}

func (d *Device) Validate(validator data.Validator) error {
	validator.SetMeta(d.Meta())

	return d.Base.Validate(validator)
}

func (d *Device) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(d.Meta())

	return d.Base.Normalize(normalizer)
}
