package blood

import (
	"strconv"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
)

type Blood struct {
	types.Base `bson:",inline"`

	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
}

func New(typ string) Blood {
	return Blood{
		Base: types.New(typ),
	}
}

func (b *Blood) Parse(parser structure.ObjectParser) {
	b.Base.Parse(parser)

	b.Units = parser.String("units")
	b.Value = parser.Float64("value")
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

func (b *Blood) LegacyIdentityFields() ([]string, error) {
	return types.NewLegacyIdentityBuilder(&b.Base, types.TypeDeviceIDTimeFormat).Build()
}
