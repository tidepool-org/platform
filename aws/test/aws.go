package test

import "github.com/tidepool-org/platform/test"

func RandomBucket() string {
	return test.RandomStringFromRangeAndCharset(1, 32, test.CharsetAlphaNumeric)
}
