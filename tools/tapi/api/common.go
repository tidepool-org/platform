package api

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

func firstStringNotNil(strs ...*string) *string {
	for _, str := range strs {
		if str != nil {
			return str
		}
	}
	return nil
}

func firstStringArrayNotNil(strs ...*[]string) *[]string {
	for _, str := range strs {
		if str != nil {
			return str
		}
	}
	return nil
}

func subtractStringArray(minuend []string, subtrahend []string) []string {
	difference := []string{}
	for _, m := range minuend {
		var found bool
		for _, s := range subtrahend {
			if m == s {
				found = true
				break
			}
		}
		if !found {
			difference = append(difference, m)
		}
	}
	return difference
}
