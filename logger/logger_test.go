package logger_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/logger"
)

var _ = Describe("Log", func() {

	It("should GetNamed", func() {
		Expect(logger.Log.GetNamed("test")).Should(Not(BeNil()))
	})

	It("should have Debug", func() {
		var debug = logger.Log.Debug
		Expect(debug).Should(Not(BeNil()))
	})

	It("should have Info", func() {
		var info = logger.Log.Info
		Expect(info).Should(Not(BeNil()))
	})

	It("should have Warn", func() {
		var warn = logger.Log.Warn
		Expect(warn).Should(Not(BeNil()))
	})

	It("should have Error", func() {
		var err = logger.Log.Error
		Expect(err).Should(Not(BeNil()))
	})

	It("should have Fatal", func() {
		var fatal = logger.Log.Fatal
		Expect(fatal).Should(Not(BeNil()))
	})

	It("should have WithField", func() {
		var withField = logger.Log.WithField
		Expect(withField).Should(Not(BeNil()))
	})

	It("should have AddTrace", func() {
		var addTrace = logger.Log.AddTrace
		Expect(addTrace).Should(Not(BeNil()))
	})

	It("should have AddTraceUUID", func() {
		var addTraceUUID = logger.Log.AddTraceUUID
		Expect(addTraceUUID).Should(Not(BeNil()))
	})
})
