package log_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
)

var _ = Describe("Null", func() {
	Context("NewNull", func() {
		It("returns successfully", func() {
			Expect(log.NewNull()).ToNot(BeNil())
		})

		Context("with new null logger", func() {
			var logger *log.Null

			BeforeEach(func() {
				logger = log.NewNull()
				Expect(logger).ToNot(BeNil())
			})

			Context("Debug", func() {
				It("returns successfully", func() {
					logger.Debug("test-debug")
				})
			})

			Context("Info", func() {
				It("returns successfully", func() {
					logger.Info("test-info")
				})
			})

			Context("Warn", func() {
				It("returns successfully", func() {
					logger.Warn("test-warn")
				})
			})

			Context("Error", func() {
				It("returns successfully", func() {
					logger.Error("test-error")
				})
			})

			Context("WithError", func() {
				It("returns a logger", func() {
					Expect(logger.WithError(errors.New("test error"))).ToNot(BeNil())
				})
			})

			Context("WithField", func() {
				It("returns a logger", func() {
					Expect(logger.WithField("testKey", "test value")).ToNot(BeNil())
				})
			})

			Context("WithFields", func() {
				It("returns a logger", func() {
					Expect(logger.WithFields(log.Fields{"testKey": "test value"})).ToNot(BeNil())
				})
			})
		})
	})
})
