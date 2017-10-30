package test

import "github.com/tidepool-org/platform/test"

func NewTraceRequest() string {
	return test.NewString(64, test.CharsetAlphaNumeric)
}

func NewTraceSession() string {
	return test.NewString(64, test.CharsetAlphaNumeric)
}
