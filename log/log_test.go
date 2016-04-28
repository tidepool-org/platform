package log_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
)

var _ = Describe("Log", func() {

	Describe("RootLogger", func() {
		var rootLogger log.Logger

		BeforeEach(func() {
			rootLogger = log.RootLogger()
		})

		It("exists", func() {
			Expect(rootLogger).Should(Not(BeNil()))
		})

		It("receives Debug", func() {
			Expect(rootLogger.Debug).Should(Not(BeNil()))
		})

		It("receives Info", func() {
			Expect(rootLogger.Info).Should(Not(BeNil()))
		})

		It("receives Warn", func() {
			Expect(rootLogger.Warn).Should(Not(BeNil()))
		})

		It("receives Error", func() {
			Expect(rootLogger.Error).Should(Not(BeNil()))
		})

		It("receives Fatal", func() {
			Expect(rootLogger.Fatal).Should(Not(BeNil()))
		})

		It("returns a new Logger from WithError", func() {
			Expect(rootLogger.WithError(fmt.Errorf("test"))).Should(Not(BeNil()))
		})

		It("returns a new Logger from WithField", func() {
			Expect(rootLogger.WithField("key", "value")).Should(Not(BeNil()))
		})

		It("returns a new Logger from WithFields", func() {
			Expect(rootLogger.WithFields(map[string]interface{}{"key": "value"})).Should(Not(BeNil()))
		})
	})
})
