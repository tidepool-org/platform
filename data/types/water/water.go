package water

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "water"
)

type Water struct {
	types.Base `bson:",inline"`

	Amount *Amount `json:"amount,omitempty" bson:"amount,omitempty"`
}

func New() *Water {
	return &Water{
		Base: types.New(Type),
	}
}

func (w *Water) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(w.Meta())
	}

	w.Base.Parse(parser)

	w.Amount = ParseAmount(parser.WithReferenceObjectParser("amount"))
}

func (w *Water) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(w.Meta())
	}

	w.Base.Validate(validator)

	if w.Type != "" {
		validator.String("type", &w.Type).EqualTo(Type)
	}

	if w.Amount != nil {
		w.Amount.Validate(validator.WithReference("amount"))
	} else {
		validator.WithReference("amount").ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (w *Water) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(w.Meta())
	}

	w.Base.Normalize(normalizer)

	if w.Amount != nil {
		w.Amount.Normalize(normalizer.WithReference("amount"))
	}
}
