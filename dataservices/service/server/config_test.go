package server_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dataservices/service/server"
)

var _ = Describe("Config", func() {
	Context("Validate", func() {
		var config *server.Config

		BeforeEach(func() {
			config = &server.Config{
				Address: "127.0.0.1",
				TLS: &server.TLS{
					CertificateFile: "config_test.go",
					KeyFile:         "config_test.go",
				},
				Timeout: 120,
			}
		})

		It("return success if all are valid", func() {
			Expect(config.Validate()).To(Succeed())
		})

		It("returns an error if the address is missing", func() {
			config.Address = ""
			Expect(config.Validate()).To(MatchError("server: address is missing"))
		})

		It("returns success if TLS is not specified", func() {
			config.TLS = nil
			Expect(config.Validate()).To(Succeed())
		})

		It("returns an error if TLS is specified, but the certificate file is missing", func() {
			config.TLS.CertificateFile = ""
			Expect(config.Validate()).To(MatchError("server: tls certificate file is missing"))
		})

		It("returns an error if TLS is specified, but the certificate file does not exist", func() {
			config.TLS.CertificateFile = "does_not_exist"
			Expect(config.Validate()).To(MatchError("server: tls certificate file does not exist"))
		})

		It("returns an error if TLS is specified, but the certificate file is a directory", func() {
			config.TLS.CertificateFile = "."
			Expect(config.Validate()).To(MatchError("server: tls certificate file is a directory"))
		})

		It("returns an error if TLS is specified, but the key file is missing", func() {
			config.TLS.KeyFile = ""
			Expect(config.Validate()).To(MatchError("server: tls key file is missing"))
		})

		It("returns an error if TLS is specified, but the key file does not exist", func() {
			config.TLS.KeyFile = "does_not_exist"
			Expect(config.Validate()).To(MatchError("server: tls key file does not exist"))
		})

		It("returns an error if TLS is specified, but the key file is a directory", func() {
			config.TLS.KeyFile = "."
			Expect(config.Validate()).To(MatchError("server: tls key file is a directory"))
		})

		It("returns success if the timeout is zero", func() {
			config.Timeout = 0
			Expect(config.Validate()).To(Succeed())
		})

		It("returns an error if the timeout is less than zero", func() {
			config.Timeout = -1
			Expect(config.Validate()).To(MatchError("server: timeout is invalid"))
		})
	})
})
