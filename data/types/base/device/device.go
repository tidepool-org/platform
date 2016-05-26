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
	"github.com/tidepool-org/platform/app"
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

func New(subType string) (*Device, error) {
	if subType == "" {
		return nil, app.Error("basal", "sub type is missing")
	}

	deviceBase, err := base.New(Type())
	if err != nil {
		return nil, err
	}

	return &Device{
		Base:    *deviceBase,
		SubType: subType,
	}, nil
}

func (d *Device) Meta() interface{} {
	return &Meta{
		Type:    d.Type,
		SubType: d.SubType,
	}
}

func (d *Device) Parse(parser data.ObjectParser) {
	parser.SetMeta(d.Meta())

	d.Base.Parse(parser)
}

func (d *Device) Validate(validator data.Validator) {
	validator.SetMeta(d.Meta())

	d.Base.Validate(validator)
}

func (d *Device) Normalize(normalizer data.Normalizer) {
	normalizer.SetMeta(d.Meta())

	d.Base.Normalize(normalizer)
}
