package test

import "github.com/tidepool-org/platform/test"

const CharsetTransmitterID = test.CharsetNumeric + test.CharsetUppercase

func RandomTransmitterID() string {
	return test.RandomStringFromRangeAndCharset(5, 6, CharsetTransmitterID)
}
