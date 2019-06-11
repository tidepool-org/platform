package pointer

import "time"

func EqualBool(a *bool, b *bool) bool {
	if (a == nil) != (b == nil) {
		return false
	} else if a == nil && b == nil {
		return true
	}
	return *a == *b
}

func EqualDuration(a *time.Duration, b *time.Duration) bool {
	if (a == nil) != (b == nil) {
		return false
	} else if a == nil && b == nil {
		return true
	}
	return *a == *b
}

func EqualFloat64(a *float64, b *float64) bool {
	if (a == nil) != (b == nil) {
		return false
	} else if a == nil && b == nil {
		return true
	}
	return *a == *b
}

func EqualInt(a *int, b *int) bool {
	if (a == nil) != (b == nil) {
		return false
	} else if a == nil && b == nil {
		return true
	}
	return *a == *b
}

func EqualString(a *string, b *string) bool {
	if (a == nil) != (b == nil) {
		return false
	} else if a == nil && b == nil {
		return true
	}
	return *a == *b
}

func EqualStringArray(a *[]string, b *[]string) bool {
	if (a == nil) != (b == nil) {
		return false
	} else if a == nil && b == nil {
		return true
	}
	if len(*a) != len(*b) {
		return false
	}
	for index := range *a {
		if (*a)[index] != (*b)[index] {
			return false
		}
	}
	return true
}

func EqualTime(a *time.Time, b *time.Time) bool {
	if (a == nil) != (b == nil) {
		return false
	} else if a == nil && b == nil {
		return true
	}
	return a.Equal(*b)
}
