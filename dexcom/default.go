package dexcom

import (
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

// NOTE: Currently we are finding that g7 data occasionally doesn't have the required units
// data. This is a workaround until that issue is resolved.
func ParseStringOrDefault(parser structure.ObjectParser, reference string, defaultValue string) *string {
	value := parser.String(reference)
	if value == nil || *value == "" {
		parser.Logger().WithField("meta", parser.Meta()).Warnf("Missing value for field '%s'; using default value '%s'", reference, defaultValue)
		value = pointer.FromString(defaultValue)
	}
	return value
}
