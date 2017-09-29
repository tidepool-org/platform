package test

import (
	"math/rand"

	"github.com/tidepool-org/platform/test"
)

func NewSourceParameter() string {
	return test.NewVariableString(1, 8, test.CharsetAlphaNumeric)
}

func NewSourcePointer() string {
	sourcePointer := ""
	for index := 0; index <= rand.Intn(4); index++ {
		sourcePointer += "/" + test.NewVariableString(1, 8, test.CharsetAlphaNumeric)
	}
	return sourcePointer
}
