package test

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/test"
)

func RandomProviderSessionID() string {
	return auth.NewProviderSessionID()
}

func RandomProviderSessionIDs() []string {
	return test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(1, 3, RandomProviderSessionID)
}

func RandomProviderType() string {
	return test.RandomStringFromArray(auth.ProviderTypes())
}

func RandomProviderTypes() []string {
	return test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(auth.ProviderTypes()), auth.ProviderTypes())
}

func RandomProviderName() string {
	return test.RandomStringFromRangeAndCharset(1, auth.ProviderNameLengthMaximum, test.CharsetAlphaNumeric)
}

func RandomProviderNames() []string {
	return test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(1, 2, RandomProviderName)
}

func RandomProviderExternalID() string {
	return test.RandomStringFromRangeAndCharset(1, auth.ProviderExternalIDLengthMaximum, test.CharsetAlphaNumeric)
}

func RandomProviderExternalIDs() []string {
	return test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(1, 2, RandomProviderExternalID)
}
