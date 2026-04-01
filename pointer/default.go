package pointer

import "time"

func Default[T any, S *T](value S, defaultValue T) T {
	if value == nil {
		return defaultValue
	}
	return *value
}

// DefaultBool
// Deprecated: use the generic version instead
func DefaultBool(value *bool, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	return *value
}

// DefaultDuration
// Deprecated: use the generic version instead
func DefaultDuration(value *time.Duration, defaultValue time.Duration) time.Duration {
	if value == nil {
		return defaultValue
	}
	return *value
}

// DefaultFloat64
// Deprecated: use the generic version instead
func DefaultFloat64(value *float64, defaultValue float64) float64 {
	if value == nil {
		return defaultValue
	}
	return *value
}

// DefaultInt
// Deprecated: use the generic version instead
func DefaultInt(value *int, defaultValue int) int {
	if value == nil {
		return defaultValue
	}
	return *value
}

// DefaultString
// Deprecated: use the generic version instead
func DefaultString(value *string, defaultValue string) string {
	if value == nil {
		return defaultValue
	}
	return *value
}

// DefaultStringArray
// Deprecated: use the generic version instead
func DefaultStringArray(value *[]string, defaultValue []string) []string {
	if value == nil {
		return defaultValue
	}
	return *value
}

// DefaultTime
// Deprecated: use the generic version instead
func DefaultTime(value *time.Time, defaultValue time.Time) time.Time {
	if value == nil {
		return defaultValue
	}
	return *value
}
