package blood

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
)

type Blood struct {
	types.Base `bson:",inline"`

	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func (b *Blood) Init() {
	b.Base.Init()

	b.Units = nil
	b.Value = nil
}

func (b *Blood) Parse(parser data.ObjectParser) error {
	parser.SetMeta(b.Meta())

	if err := b.Base.Parse(parser); err != nil {
		return err
	}

	b.Units = parser.ParseString("units")
	b.Value = parser.ParseFloat("value")

	return nil
}

func (b *Blood) Validate(validator data.Validator) error {
	validator.SetMeta(b.Meta())

	if err := b.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("units", b.Units).Exists()
	validator.ValidateFloat("value", b.Value).Exists()

	return nil
}

func (b *Blood) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(b.Meta())

	return b.Base.Normalize(normalizer)
}

func (b *Blood) IdentityFields() ([]string, error) {
	identityFields, err := b.Base.IdentityFields()
	if err != nil {
		return nil, err
	}

	if b.Units == nil {
		return nil, errors.New("units is missing")
	}
	if b.Value == nil {
		return nil, errors.New("value is missing")
	}

	return append(identityFields, *b.Units, strconv.FormatFloat(*b.Value, 'f', -1, 64)), nil
}
