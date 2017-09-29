package test

import "github.com/tidepool-org/platform/test"

func NewServiceSecret() string {
	return test.NewString(128, test.CharsetAlphaNumeric)
}

func NewAccessToken() string {
	return test.NewString(256, test.CharsetAlphaNumeric)
}

func NewSessionToken() string {
	return test.NewString(256, test.CharsetAlphaNumeric)
}

func NewRestrictedToken() string {
	return test.NewString(40, test.CharsetAlphaNumeric)
}
