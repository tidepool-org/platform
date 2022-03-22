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
	BiphasicID  *string      `json:"biphasicId,omitempty" bson:"biphasicId,omitempty"`
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

	b.BiphasicID = parser.String("biphasicId")
	// In old data model eventId field was used in order to link the two parts of the bolus
	// We now use biphasicId field for this purpose
	if b.BiphasicID == nil || *b.BiphasicID == "" {
		// injecting eventId field from the payload into BiphasicID
		b.BiphasicID = parser.String("eventId")
		// re-injecting guid field from the payload into GUID
		// because of base type parsing (see Parse method in data/types/base.go)
		// that injects eventId when guid is not defined
		b.GUID = parser.String("guid")
	}
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
	validator.String("biphasicId", b.BiphasicID).Exists().NotEmpty()
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
