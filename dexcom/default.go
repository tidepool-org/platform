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

	if givenValue != nil && *givenValue != "" {
		return givenValue
	}
	strDefault := fmt.Sprintf("%v", defaultValue)
	if defaultValue != nil && strDefault != "" {
		fmt.Printf("No value found for [%s] so falling back to default [%s]",
			reference,
			strDefault,
		)
		return pointer.FromString(strDefault)
	}

	fmt.Printf("Neither value found for [%s] or the default has been set",
		reference,
	)
	return nil
}
