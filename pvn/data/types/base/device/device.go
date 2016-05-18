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

import "github.com/tidepool-org/platform/pvn/data"
import "github.com/tidepool-org/platform/pvn/data/types/base"

type Device struct {
	base.Base
	SubType *string `json:"subType" bson:"subType"`
}

func Type() string {
	return "deviceEvent"
}

func (d *Device) Parse(parser data.ObjectParser) {
	d.Base.Parse(parser)
}

func (d *Device) Validate(validator data.Validator) {
	d.Base.Validate(validator)
	validator.ValidateString("type", d.Type).Exists().EqualTo(Type())
	validator.ValidateString("subType", d.SubType).Exists()
}

func (d *Device) Normalize(normalizer data.Normalizer) {
	d.Base.Normalize(normalizer)
}
