package glucose

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood"
)

type Glucose struct {
	blood.Blood `bson:",inline"`
}

func (g *Glucose) Validate(validator data.Validator) error {
	if err := g.Blood.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("units", g.Units).OneOf(glucose.Units())
	validator.ValidateFloat("value", g.Value).InRange(glucose.ValueRangeForUnits(g.Units))

	return nil
}

func (g *Glucose) Normalize(normalizer data.Normalizer) error {
	if err := g.Blood.Normalize(normalizer); err != nil {
		return err
	}

	g.Value = glucose.NormalizeValueForUnits(g.Value, g.Units)
	g.Units = glucose.NormalizeUnits(g.Units)

	return nil
}
