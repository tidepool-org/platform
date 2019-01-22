package test

import (
	"fmt"
	"strings"

	"github.com/tidepool-org/platform/test"
)

const (
	CharsetMediaType = test.CharsetLowercase + test.CharsetNumeric + "._-"
	CharsetSubDomain = test.CharsetAlphaNumeric + "-"
)

func RandomEmail() string {
	return fmt.Sprintf("%s+%s@%s", test.RandomStringFromRangeAndCharset(1, 8, test.CharsetAlpha), test.RandomStringFromRangeAndCharset(1, 4, test.CharsetNumeric), RandomFQDN())
}

func RandomFQDN() string {
	return RandomSubDomains(2, 4)
}

func RandomMediaType() string {
	return fmt.Sprintf("%s/%s", test.RandomStringFromRangeAndCharset(1, 32, CharsetMediaType), test.RandomStringFromRangeAndCharset(1, 32, CharsetMediaType))
}

func RandomMediaTypes(minimumLength int, maximumLength int) []string {
	return test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(minimumLength, maximumLength, RandomMediaType)
}

func RandomReverseDomain() string {
	return RandomSubDomains(2, 4)
}

func RandomSemanticVersion() string {
	return fmt.Sprintf("%d.%d.%d", test.RandomIntFromRange(0, 20), test.RandomIntFromRange(0, 20), test.RandomIntFromRange(0, 20))
}

func RandomSubDomain() string {
	return test.RandomStringFromRangeAndCharset(1, 1, test.CharsetAlphaNumeric) + test.RandomStringFromRangeAndCharset(0, 6, CharsetSubDomain) + test.RandomStringFromRangeAndCharset(1, 1, test.CharsetAlphaNumeric)
}

func RandomSubDomains(minimumLength int, maximumLength int) string {
	return strings.Join(test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(minimumLength, maximumLength, RandomSubDomain), ".")
}
