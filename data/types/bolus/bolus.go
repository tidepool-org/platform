package bolus

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/bolus/iob"
	"github.com/tidepool-org/platform/data/types/bolus/prescriptor"
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "bolus"
)

type Bolus struct {
	types.Base `bson:",inline"`

	SubType string `json:"subType,omitempty" bson:"subType,omitempty"`

	InsulinFormulation *insulin.Formulation     `json:"insulinFormulation,omitempty" bson:"insulinFormulation,omitempty"`
	Prescriptor        *prescriptor.Prescriptor `bson:",inline"`
	InsulinOnBoard     *iob.Iob                 `bson:",inline"`
}

type Meta struct {
	Type    string `json:"type,omitempty"`
	SubType string `json:"subType,omitempty"`
}

func New(subType string) Bolus {
	return Bolus{
		Base:           types.New(Type),
		SubType:        subType,
		InsulinOnBoard: iob.NewIob(),
		Prescriptor:    prescriptor.NewPrescriptor(),
	}
}

func (b *Bolus) Meta() interface{} {
	return &Meta{
		Type:    b.Type,
		SubType: b.SubType,
	}
}

func (b *Bolus) Parse(parser structure.ObjectParser) {
	b.Base.Parse(parser)

	b.InsulinFormulation = insulin.ParseFormulation(parser.WithReferenceObjectParser("insulinFormulation"))
	b.InsulinOnBoard.Parse(parser)
	b.Prescriptor.Parse(parser)
}

func (b *Bolus) Validate(validator structure.Validator) {
	b.Base.Validate(validator)

	if b.Type != "" {
		validator.String("type", &b.Type).EqualTo(Type)
	}

	validator.String("subType", &b.SubType).Exists().NotEmpty()

	if b.InsulinFormulation != nil {
		b.InsulinFormulation.Validate(validator.WithReference("insulinFormulation"))
	}
	if b.InsulinOnBoard != nil {
		b.InsulinOnBoard.Validate(validator)
	}
	if b.Prescriptor != nil {
		b.Prescriptor.Validate(validator)
	}
}

func (b *Bolus) Normalize(normalizer data.Normalizer) {
	b.Base.Normalize(normalizer)

	if b.InsulinFormulation != nil {
		b.InsulinFormulation.Normalize(normalizer.WithReference("insulinFormulation"))
	}
	if b.Prescriptor != nil {
		b.Prescriptor.Normalize(normalizer)
	}
	if b.InsulinOnBoard != nil {
		if b.Prescriptor != nil && *b.Prescriptor.Prescriptor == prescriptor.ManualPrescriptor {
			b.InsulinOnBoard = nil
		} else {
			b.InsulinOnBoard.Normalize(normalizer)
		}
	}
}

func (b *Bolus) IdentityFields() ([]string, error) {
	identityFields, err := b.Base.IdentityFields()
	if err != nil {
		return nil, err
	}

	if b.SubType == "" {
		return nil, errors.New("sub type is empty")
	}

	return append(identityFields, b.SubType), nil
}
