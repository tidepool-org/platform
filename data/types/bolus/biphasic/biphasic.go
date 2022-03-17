package biphasic

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/bolus/normal"
	"github.com/tidepool-org/platform/structure"
)

const (
	SubType = normal.BiphasicSubType
	Part1   = "1"
	Part2   = "2"
)

func Parts() []string {
	return []string{
		Part1,
		Part2,
	}
}

type Biphasic struct {
	normal.Normal `bson:",inline"`

	Part        *string      `json:"part,omitempty" bson:"part,omitempty"`
	LinkedBolus *LinkedBolus `json:"linkedBolus,omitempty" bson:"linkedBolus,omitempty"`
}

func New() *Biphasic {
	return &Biphasic{
		Normal: normal.NewWithSubType(SubType),
	}
}

func (b *Biphasic) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(b.Meta())
	}

	b.Normal.Parse(parser)
	b.LinkedBolus = ParseLinkedBolus(parser.WithReferenceObjectParser("linkedBolus"))
	b.Part = parser.String("part")
}

func (b *Biphasic) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(b.Meta())
	}

	b.Normal.Validate(validator)

	if b.SubType != "" {
		validator.String("subType", &b.SubType).EqualTo(SubType)
	}
	validator.String("part", b.Part).Exists().NotEmpty().OneOf(Parts()...)
	validator.String("guid", b.GUID).Exists().NotEmpty()
	if b.LinkedBolus != nil {
		b.LinkedBolus.Validate(validator)
	}
}

// IsValid returns true if there is no error in the validator
func (b *Biphasic) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (b *Biphasic) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(b.Meta())
	}

	b.Bolus.Normalize(normalizer)
	if b.LinkedBolus != nil {
		b.LinkedBolus.Normalize(normalizer)
	}
}
