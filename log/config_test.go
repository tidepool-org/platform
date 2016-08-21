package log_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
)

var _ = Describe("Config", func() {
	Context("Validate", func() {
		It("returns success for the debug level", func() {
			config := &log.Config{Level: "debug"}
			Expect(config.Validate()).To(Succeed())
		})

		It("returns success for the info level", func() {
			config := &log.Config{Level: "info"}
			Expect(config.Validate()).To(Succeed())
		})

		It("returns success for the warn level", func() {
			config := &log.Config{Level: "warn"}
			Expect(config.Validate()).To(Succeed())
		})

		It("returns success for the error level", func() {
			config := &log.Config{Level: "error"}
			Expect(config.Validate()).To(Succeed())
		})

		It("returns success for the fatal level", func() {
			config := &log.Config{Level: "fatal"}
			Expect(config.Validate()).To(Succeed())
		})

		It("returns an error for the panic level", func() {
			config := &log.Config{Level: "panic"}
			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("log: level is invalid"))
		})

		It("returns an error for any other invalid level", func() {
			config := &log.Config{Level: "invalid"}
			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("log: level is invalid"))
		})

		It("returns an error for missing level", func() {
			config := &log.Config{}
			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("log: level is invalid"))
		})
	})

	Context("Clone", func() {
		It("returns successfully", func() {
			config := &log.Config{Level: "debug"}
			clone := config.Clone()
			Expect(clone).ToNot(BeIdenticalTo(config))
			Expect(clone.Level).To(Equal(config.Level))
		})
	})
})
