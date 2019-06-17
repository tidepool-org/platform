package test

import "github.com/tidepool-org/platform/test"

func NewServiceSecret() string {
	return test.RandomStringFromRangeAndCharset(128, 128, test.CharsetAlphaNumeric)
}

func NewAccessToken() string {
	return test.RandomStringFromRangeAndCharset(256, 256, test.CharsetAlphaNumeric)
}

func NewSessionToken() string {
	return test.RandomStringFromRangeAndCharset(256, 256, test.CharsetAlphaNumeric)
}

func NewRestrictedToken() string {
	return test.RandomStringFromRangeAndCharset(40, 40, test.CharsetAlphaNumeric)
}
