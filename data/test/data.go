package test

import "github.com/tidepool-org/platform/test"

func NewSessionToken() string {
	return test.NewString(256, test.CharsetAlphaNumeric)
}

func NewDeviceID() string {
	return test.NewString(32, test.CharsetText)
}
