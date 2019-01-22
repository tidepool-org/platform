package test

import "github.com/tidepool-org/platform/test"

func NewReference() string {
	return test.RandomStringFromRangeAndCharset(1, 8, test.CharsetAlphaNumeric)
}
