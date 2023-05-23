package dexcom

import (
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

// NOTE: currently we are finding that g7 data occasionally doesn't have the required units
// data. This is a workaround until that issue is resolved
func BGUnitFromParser(parser structure.ObjectParser) *string {
	unitVal := parser.String("unit")
	if unitVal == nil {
		unitVal = pointer.FromString(dataBloodGlucose.MgdL)
	}
	return unitVal
}
