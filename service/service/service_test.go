package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"strconv"

	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/config/env"
	"github.com/tidepool-org/platform/service/service"
	testHTTP "github.com/tidepool-org/platform/test/http"

	_ "github.com/tidepool-org/platform/application/version/test"
)

var _ = Describe("Service", func() {
	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			svc, err := service.New("")
			Expect(err).To(MatchError("prefix is missing"))
			Expect(svc).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(service.New("TIDEPOOL")).ToNot(BeNil())
		})
	})

	Context("with started server, config reporter, and new service", func() {
		var svc *service.Service
		var serviceSecret string
		var serviceConfigReporter config.Reporter
		var serverConfigReporter config.Reporter

		BeforeEach(func() {
			var err error
			svc, err = service.New("TIDEPOOL")
			Expect(err).ToNot(HaveOccurred())
			Expect(svc).ToNot(BeNil())

			serviceSecret = testAuth.NewServiceSecret()

			configReporter, err := env.NewReporter("TIDEPOOL")
			Expect(err).ToNot(HaveOccurred())
			Expect(configReporter).ToNot(BeNil())

			serviceConfigReporter = configReporter.WithScopes("service.test", "service")
			serviceConfigReporter.Set("secret", serviceSecret)

			serverConfigReporter = configReporter.WithScopes("server")
			serverConfigReporter.Set("address", testHTTP.NewAddress())
			serverConfigReporter.Set("timeout", strconv.Itoa(testHTTP.NewTimeout()))
		})

		Context("Initialize", func() {
			It("returns an error if the secret is missing", func() {
				serviceConfigReporter.Delete("secret")
				Expect(svc.Initialize()).To(MatchError("secret is missing"))
			})

			It("returns an error if the timeout is invalid during Load", func() {
				serverConfigReporter.Set("timeout", "invalid")
				Expect(svc.Initialize()).To(MatchError("unable to load server config; timeout is invalid"))
			})

			It("returns an error if the timeout is invalid during Validate", func() {
				serverConfigReporter.Set("timeout", "0")
				Expect(svc.Initialize()).To(MatchError("unable to create server; config is invalid; timeout is invalid"))
			})

			It("returns successfully", func() {
				Expect(svc.Initialize()).To(Succeed())
				svc.Terminate()
			})
		})

		Context("Terminate", func() {
			It("returns successfully", func() {
				svc.Terminate()
			})
		})

		Context("Run", func() {
			It("returns an error if it is not initialized", func() {
				Expect(svc.Run()).To(MatchError("service not initialized"))
			})
		})

		Context("with being initialized", func() {
			BeforeEach(func() {
				Expect(svc.Initialize()).To(Succeed())
			})

			AfterEach(func() {
				svc.Terminate()
			})

			Context("Terminate", func() {
				It("returns successfully", func() {
					svc.Terminate()
				})
			})

			Context("Run", func() {
				// Cannot invoke Run since it starts a server that requires user intervention
			})

			Context("Secret", func() {
				It("returns the secret", func() {
					Expect(svc.Secret()).To(Equal(serviceSecret))
				})
			})

			Context("AuthClient", func() {
				It("returns nil if not set", func() {
					Expect(svc.AuthClient()).To(BeNil())
				})

				It("returns successfully if set", func() {
					authClient := testAuth.NewClient()
					svc.SetAuthClient(authClient)
					Expect(svc.AuthClient()).To(Equal(authClient))
				})
			})

			Context("SetAuthClient", func() {
				It("returns successfully", func() {
					svc.SetAuthClient(testAuth.NewClient())
				})
			})

			Context("API", func() {
				It("returns successfully", func() {
					Expect(svc.API()).ToNot(BeNil())
				})
			})
		})
	})
})
