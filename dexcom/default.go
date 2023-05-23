package dexcom

import (
	"github.com/tidepool-org/platform/pointer"
)

// NOTE: currently we are finding that g7 data occasionally doesn't have the required units
// data. This is a workaround until that issue is resolved
func StringOrDefault(givenValue *string, defaultValue string) *string {
	if givenValue == nil && defaultValue != "" {
		givenValue = pointer.FromString(defaultValue)
	}
	return givenValue
}
