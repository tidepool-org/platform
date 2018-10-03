package pointer

import "time"

func ToBool(ptr *bool) bool {
	if ptr == nil {
		return false
	}
	return *ptr
}

func ToDuration(ptr *time.Duration) time.Duration {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func ToFloat64(ptr *float64) float64 {
	if ptr == nil {
		return 0.
	}
	return *ptr
}

func ToInt(ptr *int) int {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func ToString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func ToStringArray(ptr *[]string) []string {
	if ptr == nil {
		return nil
	}
	return *ptr
}

func ToTime(ptr *time.Time) time.Time {
	if ptr == nil {
		return time.Time{}
	}
	return *ptr
}
