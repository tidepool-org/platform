package server_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/service/server"
)

var _ = Describe("Config", func() {
	Context("NewConfig", func() {
		It("returns a new config with default values", func() {
			config := server.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Address).To(Equal(""))
			Expect(config.TLS).To(BeTrue())
			Expect(config.TLSCertificateFile).To(Equal(""))
			Expect(config.TLSKeyFile).To(Equal(""))
			Expect(config.Timeout).To(Equal(60 * time.Second))
		})
	})

	Context("with new config", func() {
		var config *server.Config

		BeforeEach(func() {
			config = server.NewConfig()
			Expect(config).ToNot(BeNil())
		})

		Context("Load", func() {
			var configReporter *test.Reporter

			BeforeEach(func() {
				configReporter = test.NewReporter()
				configReporter.Config["address"] = "https://1.2.3.4:5678"
				configReporter.Config["tls"] = "false"
				configReporter.Config["tls_certificate_file"] = "my-certificate-file"
				configReporter.Config["tls_key_file"] = "my-key-file"
				configReporter.Config["timeout"] = "120"
			})

			It("returns an error if config reporter is missing", func() {
				Expect(config.Load(nil)).To(MatchError("config reporter is missing"))
			})

			It("uses default address if not set", func() {
				delete(configReporter.Config, "address")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Address).To(Equal(""))
			})

			It("uses default tls if not set", func() {
				delete(configReporter.Config, "tls")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.TLS).To(BeTrue())
			})

			It("returns an error if the tls cannot be parsed to a boolean", func() {
				configReporter.Config["tls"] = "abc"
				Expect(config.Load(configReporter)).To(MatchError("tls is invalid"))
				Expect(config.TLS).To(BeTrue())
			})

			It("uses default tls certificate file if not set", func() {
				delete(configReporter.Config, "tls_certificate_file")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.TLSCertificateFile).To(Equal(""))
			})

			It("uses default tls key file if not set", func() {
				delete(configReporter.Config, "tls_key_file")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.TLSKeyFile).To(Equal(""))
			})

			It("uses default timeout if not set", func() {
				delete(configReporter.Config, "timeout")
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Timeout).To(Equal(60 * time.Second))
			})

			It("returns an error if the timeout cannot be parsed to an integer", func() {
				configReporter.Config["timeout"] = "abc"
				Expect(config.Load(configReporter)).To(MatchError("timeout is invalid"))
				Expect(config.Timeout).To(Equal(60 * time.Second))
			})

			It("returns successfully and uses values from config reporter", func() {
				Expect(config.Load(configReporter)).To(Succeed())
				Expect(config.Address).To(Equal("https://1.2.3.4:5678"))
				Expect(config.TLS).To(BeFalse())
				Expect(config.TLSCertificateFile).To(Equal("my-certificate-file"))
				Expect(config.TLSKeyFile).To(Equal("my-key-file"))
				Expect(config.Timeout).To(Equal(120 * time.Second))
			})
		})

		Context("with valid values", func() {
			BeforeEach(func() {
				config.Address = "127.0.0.1"
				config.TLS = true
				config.TLSCertificateFile = "config_test.go"
				config.TLSKeyFile = "config_test.go"
				config.Timeout = 120 * time.Second
			})

			Context("Validate", func() {
				It("returns success if all are valid", func() {
					Expect(config.Validate()).To(Succeed())
				})

				It("returns an error if the address is missing", func() {
					config.Address = ""
					Expect(config.Validate()).To(MatchError("address is missing"))
				})

				It("returns an error if TLS is specified, but the certificate file is missing", func() {
					config.TLSCertificateFile = ""
					Expect(config.Validate()).To(MatchError("tls certificate file is missing"))
				})

				It("returns an error if TLS is specified, but the certificate file does not exist", func() {
					config.TLSCertificateFile = "does_not_exist"
					Expect(config.Validate()).To(MatchError("tls certificate file does not exist"))
				})

				It("returns an error if TLS is specified, but the certificate file is a directory", func() {
					config.TLSCertificateFile = "."
					Expect(config.Validate()).To(MatchError("tls certificate file is a directory"))
				})

				It("returns an error if TLS is specified, but the key file is missing", func() {
					config.TLSKeyFile = ""
					Expect(config.Validate()).To(MatchError("tls key file is missing"))
				})

				It("returns an error if TLS is specified, but the key file does not exist", func() {
					config.TLSKeyFile = "does_not_exist"
					Expect(config.Validate()).To(MatchError("tls key file does not exist"))
				})

				It("returns an error if TLS is specified, but the key file is a directory", func() {
					config.TLSKeyFile = "."
					Expect(config.Validate()).To(MatchError("tls key file is a directory"))
				})

				It("returns an error if the timeout is invalid", func() {
					config.Timeout = 0
					Expect(config.Validate()).To(MatchError("timeout is invalid"))
				})
			})
		})
	})
})
