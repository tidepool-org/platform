package log_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
)

var _ = Describe("Default", func() {
	Context("DebugLevel", func() {
		It("returns the expected string", func() {
			Expect(string(log.DebugLevel)).To(Equal("debug"))
		})
	})

	Context("InfoLevel", func() {
		It("returns the expected string", func() {
			Expect(string(log.InfoLevel)).To(Equal("info"))
		})
	})

	Context("WarnLevel", func() {
		It("returns the expected string", func() {
			Expect(string(log.WarnLevel)).To(Equal("warn"))
		})
	})

	Context("ErrorLevel", func() {
		It("returns the expected string", func() {
			Expect(string(log.ErrorLevel)).To(Equal("error"))
		})
	})

	Context("DefaultLevels", func() {
		It("returns the expected map", func() {
			Expect(log.DefaultLevels()).To(Equal(log.Levels{log.DebugLevel: 10, log.InfoLevel: 20, log.WarnLevel: 40, log.ErrorLevel: 80}))
		})
	})

	Context("DefaultLevel", func() {
		It("returns the expected level", func() {
			Expect(log.DefaultLevel()).To(Equal(log.WarnLevel))
		})
	})
})
