package test

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	gomegaTypes "github.com/onsi/gomega/types"
	"go.uber.org/mock/gomock"
)

func init() {
	if os.Getenv("TIDEPOOL_ENV") != "test" {
		//panic("Test packages only supported in test environment!!!")
	}
	if matches := initPackageRegexp.FindStringSubmatch(getFrameName(1)); matches != nil {
		callerPackageRegexp = regexp.MustCompile("^" + matches[1] + "/(.+?)(?:_test)[^/]+$")
	}
}

func Test(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, getCallerPackage())
}

func getFrameName(frame int) string {
	var frameName string
	if pc, _, _, ok := runtime.Caller(frame); ok {
		frameName = runtime.FuncForPC(pc).Name()
	}
	return frameName
}

func getCallerPackage() string {
	var callerPackage string
	if matches := callerPackageRegexp.FindStringSubmatch(getFrameName(3)); matches != nil {
		callerPackage = matches[1]
	}
	return callerPackage
}

var callerPackageRegexp = regexp.MustCompile("^(.+?)(?:_test)[^/]+$")
var initPackageRegexp = regexp.MustCompile("^(.+)/[^/]+$")

func MockMatch(matcher gomegaTypes.GomegaMatcher) gomock.Matcher {
	return &MockMatcher{matcher: matcher}
}

type MockMatcher struct {
	matcher gomegaTypes.GomegaMatcher
}

func (m *MockMatcher) Matches(x any) bool {
	if match, err := m.matcher.Match(x); err != nil {
		return false
	} else {
		return match
	}
}

func (m *MockMatcher) String() string {
	return fmt.Sprintf("%#v", m.matcher)
}
