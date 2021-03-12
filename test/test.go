package test

import (
	"os"
	"regexp"
	"runtime"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
)

func init() {
	if os.Getenv("TIDEPOOL_ENV") != "test" {
		panic("Test packages only supported in test environment!!!")
	}
	if matches := initPackageRegexp.FindStringSubmatch(getFrameName(1)); matches != nil {
		callerPackageRegexp = regexp.MustCompile("^" + matches[1] + "/(.+?)(?:_test)[^/]+$")
	}
}

func Test(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	if os.Getenv("JENKINS_TEST") == "on" {
		ginkgo.RunSpecsWithCustomReporters(t, "Platform test suite", []ginkgo.Reporter{junitReporter})
	} else {
		ginkgo.RunSpecs(t, getCallerPackage())
	}
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
