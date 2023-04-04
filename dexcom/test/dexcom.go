package test

import "github.com/tidepool-org/platform/test"

const CharsetTransmitterID = test.CharsetNumeric + test.CharsetLowercase

func RandomTransmitterID() string {
	return test.RandomStringFromRangeAndCharset(64, 64, CharsetTransmitterID)
}
