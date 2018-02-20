package test

func CloneFloat64(datum *float64) *float64 {
	if datum == nil {
		return nil
	}
	clone := *datum
	return &clone
}

func CloneInt(datum *int) *int {
	if datum == nil {
		return nil
	}
	clone := *datum
	return &clone
}

func CloneString(datum *string) *string {
	if datum == nil {
		return nil
	}
	clone := *datum
	return &clone
}

func CloneStringArray(datum *[]string) *[]string {
	if datum == nil {
		return nil
	}
	clone := make([]string, len(*datum))
	for index, value := range *datum {
		clone[index] = value
	}
	return &clone
}
