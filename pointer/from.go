package pointer

import "time"

func FromBool(value bool) *bool {
	return &value
}

func FromDuration(value time.Duration) *time.Duration {
	return &value
}

func FromFloat64(value float64) *float64 {
	return &value
}

func FromInt(value int) *int {
	return &value
}

func FromInt64(value int64) *int64 {
	return &value
}

func FromString(value string) *string {
	return &value
}

func FromStringArray(value []string) *[]string {
	return &value
}

func FromTime(value time.Time) *time.Time {
	return &value
}
