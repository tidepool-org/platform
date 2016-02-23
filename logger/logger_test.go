package logger_test

import (
	. "github.com/tidepool-org/platform/logger"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Log", func() {

	It("should have Debug", func() {
		var debug = Debug
		Expect(debug).Should(Not(BeNil()))
	})

	It("should have Info", func() {
		var info = Info
		Expect(info).Should(Not(BeNil()))
	})

	It("should have Warn", func() {
		var warn = Warn
		Expect(warn).Should(Not(BeNil()))
	})

	It("should have Error", func() {
		var err = Error
		Expect(err).Should(Not(BeNil()))
	})

	It("should have Fatal", func() {
		var fatal = Fatal
		Expect(fatal).Should(Not(BeNil()))
	})

	It("should have WithField", func() {
		var withField = WithField
		Expect(withField).Should(Not(BeNil()))
	})

	It("should have AddTrace", func() {
		var addTrace = AddTrace
		Expect(addTrace).Should(Not(BeNil()))
	})
})
