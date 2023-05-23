package dexcom

import (
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

// NOTE: currently we are finding that g7 data occasionally doesn't have the required units
// data. This is a workaround until that issue is resolved
func UnitFromParser(parser structure.ObjectParser, defaultUnit string) *string {
	unitVal := parser.String("unit")
	if unitVal == nil {
		unitVal = pointer.FromString(defaultUnit)
	}
	return unitVal
}
