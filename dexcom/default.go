package dexcom

import (
	"fmt"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

// NOTE: currently we are finding that g7 data occasionally doesn't have the required units
// data. This is a workaround until that issue is resolved
func StringOrDefault(parser structure.ObjectParser, reference string, defaultValue interface{}) *string {
	givenValue := parser.String(reference)
	if givenValue == nil && defaultValue != nil {
		strDefault := fmt.Sprintf("%v", defaultValue)
		if strDefault != "" {
			fmt.Printf("No value found for [%s] so falling back to default [%s]",
				reference,
				strDefault,
			)
			givenValue = pointer.FromString(strDefault)
		}
	}
	return givenValue
}
