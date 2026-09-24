package test

import "github.com/tidepool-org/platform/test"

func RandomServiceSecret() string {
	return test.RandomStringFromRangeAndCharset(128, 128, test.CharsetAlphaNumeric)
}

func RandomAccessToken() string {
	return test.RandomStringFromRangeAndCharset(256, 256, test.CharsetAlphaNumeric)
}

func RandomSessionToken() string {
	return test.RandomStringFromRangeAndCharset(256, 256, test.CharsetAlphaNumeric)
}

func RandomRestrictedToken() string {
	return test.RandomStringFromRangeAndCharset(40, 40, test.CharsetAlphaNumeric)
}

// DEPRECATED: The functions below are deprecated. Please use the functions above instead.

func NewServiceSecret() string {
	return RandomServiceSecret()
}

func NewAccessToken() string {
	return RandomAccessToken()
}

func NewSessionToken() string {
	return RandomSessionToken()
}

func NewRestrictedToken() string {
	return RandomRestrictedToken()
}
