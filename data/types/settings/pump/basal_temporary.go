package pump

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	BasalTemporaryTypePercent      = "percent"
	BasalTemporaryTypeUnitsPerHour = "Units/hour"
)

func BasalTemporaryTypes() []string {
	return []string{
		BasalTemporaryTypePercent,
		BasalTemporaryTypeUnitsPerHour,
	}
}

type BasalTemporary struct {
	Type *string `json:"type,omitempty" bson:"type,omitempty"`
}

func ParseBasalTemporary(parser data.ObjectParser) *BasalTemporary {
	if parser.Object() == nil {
		return nil
	}
	datum := NewBasalTemporary()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewBasalTemporary() *BasalTemporary {
	return &BasalTemporary{}
}

func (b *BasalTemporary) Parse(parser data.ObjectParser) {
	b.Type = parser.ParseString("type")
}

func (b *BasalTemporary) Validate(validator structure.Validator) {
	validator.String("type", b.Type).Exists().OneOf(BasalTemporaryTypes()...)
}

func (b *BasalTemporary) Normalize(normalizer data.Normalizer) {}
