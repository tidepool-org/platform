package config_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Init", func() {
	It("sets CONFIGOR_ENV", func() {
		Expect(os.Getenv("CONFIGOR_ENV")).To(Equal("test"))
	})

	It("sets CONFIGOR_ENV_PREFIX", func() {
		Expect(os.Getenv("CONFIGOR_ENV_PREFIX")).To(Equal("TIDEPOOL"))
	})
})
