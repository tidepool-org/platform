package test

import (
	"regexp"
	"runtime"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

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

func init() {
	if matches := initPackageRegexp.FindStringSubmatch(getFrameName(1)); matches != nil {
		callerPackageRegexp = regexp.MustCompile("^" + matches[1] + "/(.+?)(?:_test)[^/]+$")
	}
}

var callerPackageRegexp = regexp.MustCompile("^(.+?)(?:_test)[^/]+$")
var initPackageRegexp = regexp.MustCompile("^(.+)/[^/]+$")
