package pointer

import "time"

func DefaultBool(value *bool, defaultValue bool) *bool {
	if value == nil {
		return &defaultValue
	}
	return value
}

func DefaultDuration(value *time.Duration, defaultValue time.Duration) *time.Duration {
	if value == nil {
		return &defaultValue
	}
	return value
}

func DefaultFloat64(value *float64, defaultValue float64) *float64 {
	if value == nil {
		return &defaultValue
	}
	return value
}

func DefaultInt(value *int, defaultValue int) *int {
	if value == nil {
		return &defaultValue
	}
	return value
}

func DefaultString(value *string, defaultValue string) *string {
	if value == nil {
		return &defaultValue
	}
	return value
}

func DefaultStringArray(value *[]string, defaultValue []string) *[]string {
	if value == nil {
		return &defaultValue
	}
	return value
}

func DefaultTime(value *time.Time, defaultValue time.Time) *time.Time {
	if value == nil {
		return &defaultValue
	}
	return value
}
