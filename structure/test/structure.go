package test

import "github.com/tidepool-org/platform/test"

func NewReference() string {
	return test.NewVariableString(1, 8, test.CharsetAlphaNumeric)
}
