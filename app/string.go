package app

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"strconv"
	"strings"
)

func FirstStringNotEmpty(sourceStrings ...string) string {
	for _, sourceString := range sourceStrings {
		if sourceString != "" {
			return sourceString
		}
	}
	return ""
}

func SplitStringAndRemoveWhitespace(sourceString string, separator string) []string {
	splitStrings := []string{}
	for _, splitString := range strings.Split(sourceString, separator) {
		splitString = strings.TrimSpace(splitString)
		if splitString != "" {
			splitStrings = append(splitStrings, splitString)
		}
	}
	return splitStrings
}

func QuoteIfString(interfaceValue interface{}) interface{} {
	if stringValue, ok := interfaceValue.(string); ok {
		return strconv.Quote(stringValue)
	}
	return interfaceValue
}

func StringsContainsString(sourceStrings []string, searchString string) bool {
	for _, sourceString := range sourceStrings {
		if sourceString == searchString {
			return true
		}
	}
	return false
}

func StringsContainsAnyStrings(sourceStrings []string, searchStrings []string) bool {
	for _, sourceString := range sourceStrings {
		for _, searchString := range searchStrings {
			if sourceString == searchString {
				return true
			}
		}
	}
	return false
}
