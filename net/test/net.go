package test

import (
	"fmt"

	"github.com/tidepool-org/platform/test"
)

const CharsetMediaType = test.CharsetAlphaNumeric + "._-"

func RandomMediaType() string {
	return fmt.Sprintf("%s/%s",
		test.RandomStringFromRangeAndCharset(1, 32, CharsetMediaType),
		test.RandomStringFromRangeAndCharset(1, 32, CharsetMediaType))
}

func RandomMediaTypes(minimumLength int, maximumLength int) []string {
	return test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(minimumLength, maximumLength, RandomMediaType)
}

func RandomSemanticVersion() string {
	return fmt.Sprintf("%d.%d.%d", test.RandomIntFromRange(0, 10), test.RandomIntFromRange(0, 10), test.RandomIntFromRange(0, 10))
}
