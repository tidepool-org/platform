package test

import "github.com/tidepool-org/platform/test"

func RandomClientID() string {
	return test.RandomStringFromCharset(test.CharsetAlphaNumeric)
}

func RandomClientSecret() string {
	return test.RandomStringFromCharset(test.CharsetAlphaNumeric)
}
