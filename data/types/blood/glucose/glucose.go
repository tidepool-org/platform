package glucose

import (
	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood"
	"github.com/tidepool-org/platform/structure"
)

type Glucose struct {
	blood.Blood `bson:",inline"`
}

func New(typ string) Glucose {
	return Glucose{
		Blood: blood.New(typ),
	}
}

func (g *Glucose) Validate(validator structure.Validator) {
	g.Blood.Validate(validator)

	validator.String("units", g.Units).Exists().OneOf(dataBloodGlucose.Units()...)
	validator.Float64("value", g.Value).Exists().InRange(dataBloodGlucose.ValueRangeForUnits(g.Units))
}

func (g *Glucose) Normalize(normalizer data.Normalizer) {
	g.Blood.Normalize(normalizer)

	if normalizer.Origin() == structure.OriginExternal {
		g.SetRawValueAndUnits(g.Value, g.Units)
		g.Value = dataBloodGlucose.NormalizeValueForUnits(g.Value, g.Units)
		g.Units = dataBloodGlucose.NormalizeUnits(g.Units)
	}
}
