package test

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/test"
)

func RandomID() string {
	return data.NewID()
}

func NewSessionToken() string {
	return test.NewString(256, test.CharsetAlphaNumeric)
}

func NewDeviceID() string {
	return test.NewString(32, test.CharsetText)
}
