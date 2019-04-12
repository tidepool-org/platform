package test

import (
	"strings"

	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
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

func RandomOptions() *storeUnstructured.Options {
	datum := storeUnstructured.NewOptions()
	datum.MediaType = pointer.FromString(netTest.RandomMediaType())
	return datum
}
