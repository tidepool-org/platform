package test

import "github.com/tidepool-org/platform/test"

func NewTraceRequest() string {
	return test.RandomStringFromRangeAndCharset(64, 64, test.CharsetAlphaNumeric)
}

func NewTraceSession() string {
	return test.RandomStringFromRangeAndCharset(64, 64, test.CharsetAlphaNumeric)
}
