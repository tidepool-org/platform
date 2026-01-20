package test

import (
	"sort"

	"github.com/tidepool-org/platform/test"
)

// See https://datatracker.ietf.org/doc/html/rfc6749#section-3.3

const CharsetScopeToken = "!#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmnopqrstuvwxyz{|}~"

func RandomScopeToken() string {
	return test.RandomStringFromCharset(CharsetScopeToken)
}

func RandomScope() []string {
	scope := test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(1, 3, RandomScopeToken)
	sort.Strings(scope)
	return scope
}
