package test

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/test"
)

func RandomRestrictedTokenID() string {
	return auth.NewRestrictedTokenID()
}

func RandomRestrictedTokenIDs() []string {
	return test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(1, 3, RandomRestrictedTokenID)
}
