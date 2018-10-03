package test

import "github.com/tidepool-org/platform/test"

func RandomBucket() string {
	return test.NewVariableString(1, 32, test.CharsetAlphaNumeric)
}
