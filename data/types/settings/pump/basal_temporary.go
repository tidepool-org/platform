package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	BasalTemporaryTypeOff          = "off"
	BasalTemporaryTypePercent      = "percent"
	BasalTemporaryTypeUnitsPerHour = "Units/hour"
)

func BasalTemporaryTypes() []string {
	return []string{
		BasalTemporaryTypeOff,
		BasalTemporaryTypePercent,
		BasalTemporaryTypeUnitsPerHour,
	}
}

type BasalTemporary struct {
	Type *string `json:"type,omitempty" bson:"type,omitempty"`
}

func ParseBasalTemporary(parser structure.ObjectParser) *BasalTemporary {
	if !parser.Exists() {
		return nil
	}
	datum := NewBasalTemporary()
	parser.Parse(datum)
	return datum
}

func NewBasalTemporary() *BasalTemporary {
	return &BasalTemporary{}
}

func (b *BasalTemporary) Parse(parser structure.ObjectParser) {
	b.Type = parser.String("type")
}

func (b *BasalTemporary) Validate(validator structure.Validator) {
	validator.String("type", b.Type).Exists().OneOf(BasalTemporaryTypes()...)
}

func (b *BasalTemporary) Normalize(normalizer data.Normalizer) {}
