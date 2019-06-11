package test

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/test"
)

func RandomFields() log.Fields {
	datum := log.Fields{}
	for count := test.RandomIntFromRange(1, 3); count > 0; count-- {
		datum[RandomKey()] = RandomValue()
	}
	return datum
}

func RandomKey() string {
	return test.RandomStringFromRangeAndCharset(4, 16, test.CharsetAlphaNumeric)
}

func RandomValue() interface{} {
	return test.RandomString()
}
