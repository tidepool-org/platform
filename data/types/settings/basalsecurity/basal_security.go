package basalsecurity

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "basalSecurity"
)

type BasalSecurity struct {
	types.Base `bson:",inline"`

	BasalRateSchedule *pump.BasalRateStartArray `json:"basalSchedule,omitempty" bson:"basalSchedule,omitempty"`
}

func New() *BasalSecurity {
	return &BasalSecurity{
		Base: types.New(Type),
	}
}

func (bs *BasalSecurity) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(bs.Meta())
	}

	bs.Base.Parse(parser)

	bs.BasalRateSchedule = pump.ParseBasalRateStartArray(parser.WithReferenceArrayParser("basalSchedule"))
}

func (bs *BasalSecurity) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(bs.Meta())
	}

	bs.Base.Validate(validator)

	if bs.Type != "" {
		validator.String("type", &bs.Type).EqualTo(Type)
	}

	if bs.BasalRateSchedule != nil {
		bs.BasalRateSchedule.Validate(validator.WithReference("basalSchedule"))
	}
}

// IsValid returns true if there is no error in the validator
func (bs *BasalSecurity) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (bs *BasalSecurity) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(bs.Meta())
	}

	bs.Base.Normalize(normalizer)

	if bs.BasalRateSchedule != nil {
		bs.BasalRateSchedule.Normalize(normalizer.WithReference("basalSchedule"))
	}

}
