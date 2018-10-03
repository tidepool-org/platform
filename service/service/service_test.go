package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	applicationTest "github.com/tidepool-org/platform/application/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/service/service"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Service", func() {
	Context("New", func() {
		It("returns successfully", func() {
			Expect(service.New()).ToNot(BeNil())
		})
	})

	Context("with started server, config reporter, and new service", func() {
		var provider *applicationTest.Provider
		var svc *service.Service
		var serverConfig map[string]interface{}
		var serviceConfig map[string]interface{}

		BeforeEach(func() {
			provider = applicationTest.NewProviderWithDefaults()

			serverConfig = map[string]interface{}{
				"address": testHttp.NewAddress(),
				"tls":     "false",
			}
			serviceConfig = map[string]interface{}{
				"secret": authTest.NewServiceSecret(),
				"server": serverConfig,
			}
			(*provider.ConfigReporterOutput).(*configTest.Reporter).Config = serviceConfig

			svc = service.New()
			Expect(svc).ToNot(BeNil())
		})

		AfterEach(func() {
			provider.AssertOutputsEmpty()
		})

		Context("with Terminate after", func() {
			AfterEach(func() {
				svc.Terminate()
			})

			Context("Initialize", func() {
				It("returns an error if the provider is missing", func() {
					Expect(svc.Initialize(nil)).To(MatchError("provider is missing"))
				})

				It("returns an error if the secret is missing", func() {
					delete(serviceConfig, "secret")
					Expect(svc.Initialize(provider)).To(MatchError("secret is missing"))
				})

				It("returns an error if the timeout is invalid during Load", func() {
					serverConfig["timeout"] = "invalid"
					Expect(svc.Initialize(provider)).To(MatchError("unable to load server config; timeout is invalid"))
				})

				It("returns an error if the timeout is invalid during Validate", func() {
					serverConfig["timeout"] = "0"
					Expect(svc.Initialize(provider)).To(MatchError("unable to create server; config is invalid; timeout is invalid"))
				})

				It("returns successfully", func() {
					Expect(svc.Initialize(provider)).To(Succeed())
				})
			})

			Context("with Initialize before", func() {
				BeforeEach(func() {
					Expect(svc.Initialize(provider)).To(Succeed())
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
						Expect(svc.Secret()).To(Equal(serviceConfig["secret"]))
					})
				})

				Context("AuthClient", func() {
					It("returns nil if not set", func() {
						Expect(svc.AuthClient()).To(BeNil())
					})

					It("returns successfully if set", func() {
						authClient := authTest.NewClient()
						svc.SetAuthClient(authClient)
						Expect(svc.AuthClient()).To(Equal(authClient))
					})
				})

				Context("SetAuthClient", func() {
					It("returns successfully", func() {
						svc.SetAuthClient(authTest.NewClient())
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
})
