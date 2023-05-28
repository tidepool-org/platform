package dexcom

import (
	"fmt"

	"github.com/tidepool-org/platform/pointer"
)

// NOTE: currently we are finding that g7 data occasionally doesn't have the required units
// data. This is a workaround until that issue is resolved
func StringOrDefault(givenValue *string, defaultValue interface{}) *string {
	if givenValue == nil && defaultValue != nil {
		strDefault := fmt.Sprintf("%v", defaultValue)
		if strDefault != "" {
			givenValue = pointer.FromString(strDefault)
		}
	}
	return givenValue
}
