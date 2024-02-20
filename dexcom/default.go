package dexcom

import (
	"fmt"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

// NOTE: Currently we are finding that g7 data occasionally doesn't have the required units
// data. This is a workaround until that issue is resolved.
func StringOrDefault(parser structure.ObjectParser, reference string, defaultValue string) *string {
	value := parser.String(reference)
	if value == nil || *value == "" {
		value = pointer.FromString(defaultValue)

		// TODO: DO NOT COMMIT!!!
		fmt.Printf("DARIN: COUNT[StringOrDefault] %s\n", reference)
	}
	return value
}
