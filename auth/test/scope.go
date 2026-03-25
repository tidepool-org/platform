package test

import (
	"sort"

	auth "github.com/tidepool-org/platform/auth"
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

func CloneScope(scope []string) []string {
	return test.CloneStringArray(scope)
}

func NewObjectFromScope(scope []string, objectFormat test.ObjectFormat) any {
	return auth.JoinScope(scope)
}
