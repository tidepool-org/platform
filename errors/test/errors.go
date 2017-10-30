package test

import (
	"errors"
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

func NewError() error {
	return errors.New(test.NewText(1, 64))
}

func NewMeta() interface{} {
	meta := map[string]interface{}{}
	for index := 0; index <= rand.Intn(2); index++ {
		meta[test.NewVariableString(1, 8, test.CharsetAlphaNumeric)] = test.NewText(1, 32)
	}
	return meta
}
