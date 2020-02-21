package test

import "github.com/tidepool-org/platform/test"

func RandomDevicePushToken() string {
	return test.RandomStringFromRangeAndCharset(64, 64, test.CharsetAlphaNumeric)
}
