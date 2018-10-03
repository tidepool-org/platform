package test

import (
	"strings"

	"github.com/tidepool-org/platform/test"
)

const CharsetKeyInitial = test.CharsetAlphaNumeric
const CharsetKeyRemaining = CharsetKeyInitial + "._-"

func RandomKey() string {
	segments := make([]string, test.RandomIntFromRange(1, 3))
	for index := range segments {
		segments[index] = RandomKeySegment()
	}
	return strings.Join(segments, "/")
}

func RandomKeySegment() string {
	return test.RandomStringFromRangeAndCharset(1, 1, CharsetKeyInitial) +
		test.RandomStringFromRangeAndCharset(0, 63, CharsetKeyRemaining)
}
