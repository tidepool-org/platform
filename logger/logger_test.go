package logger_test

import (
	. "github.com/tidepool-org/platform/logger"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Logger", func() {

	It("should be initialise on creation", func() {
		Expect(Logging).Should(Not(BeZero()))
	})

	It("should be assignable to the interface", func() {
		var testLogger Logger
		testLogger = NewPlatformLogger()
		Expect(testLogger).To(Not(BeNil()))
	})

})
