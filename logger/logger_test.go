package logger_test

import (
	. "github.com/tidepool-org/platform/logger"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("Log", func() {

	It("should GetNamed", func() {
		Expect(Log.GetNamed("test")).Should(Not(BeNil()))
	})

	It("should have Debug", func() {
		var debug = Log.Debug
		Expect(debug).Should(Not(BeNil()))
	})

	It("should have Info", func() {
		var info = Log.Info
		Expect(info).Should(Not(BeNil()))
	})

	It("should have Warn", func() {
		var warn = Log.Warn
		Expect(warn).Should(Not(BeNil()))
	})

	It("should have Error", func() {
		var err = Log.Error
		Expect(err).Should(Not(BeNil()))
	})

	It("should have Fatal", func() {
		var fatal = Log.Fatal
		Expect(fatal).Should(Not(BeNil()))
	})

	It("should have WithField", func() {
		var withField = Log.WithField
		Expect(withField).Should(Not(BeNil()))
	})

	It("should have AddTrace", func() {
		var addTrace = Log.AddTrace
		Expect(addTrace).Should(Not(BeNil()))
	})

	It("should have AddTraceUUID", func() {
		var addTraceUUID = Log.AddTraceUUID
		Expect(addTraceUUID).Should(Not(BeNil()))
	})
})
