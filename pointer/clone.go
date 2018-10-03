package pointer

import "time"

func CloneBool(source *bool) *bool {
	if source == nil {
		return nil
	}
	clone := *source
	return &clone
}

func CloneDuration(source *time.Duration) *time.Duration {
	if source == nil {
		return nil
	}
	clone := *source
	return &clone
}

func CloneFloat64(source *float64) *float64 {
	if source == nil {
		return nil
	}
	clone := *source
	return &clone
}

func CloneInt(source *int) *int {
	if source == nil {
		return nil
	}
	clone := *source
	return &clone
}

func CloneString(source *string) *string {
	if source == nil {
		return nil
	}
	clone := *source
	return &clone
}

func CloneStringArray(source *[]string) *[]string {
	if source == nil {
		return nil
	}
	var clone []string
	if *source != nil {
		clone = make([]string, len(*source))
		copy(clone, *source)
	}
	return &clone
}

func CloneTime(source *time.Time) *time.Time {
	if source == nil {
		return nil
	}
	clone := *source
	return &clone
}
