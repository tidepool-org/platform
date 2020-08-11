package service_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	applicationTest "github.com/tidepool-org/platform/application/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	configTest "github.com/tidepool-org/platform/config/test"
	serviceService "github.com/tidepool-org/platform/service/service"
)

var _ = Describe("DEPRECATEDService", func() {
	Context("NewDEPRECATEDService", func() {
		It("returns successfully", func() {
			Expect(serviceService.NewDEPRECATEDService()).ToNot(BeNil())
		})
	})

	Context("with started server, config reporter, and new service", func() {
		var provider *applicationTest.Provider
		var svc *serviceService.DEPRECATEDService
		var serverSecret string
		var sessionToken string
		var server *Server
		var authClientConfig map[string]interface{}
		var serviceConfig map[string]interface{}

		BeforeEach(func() {
			provider = applicationTest.NewProviderWithDefaults()

			serverSecret = authTest.NewServiceSecret()
			sessionToken = authTest.NewSessionToken()

			server = NewServer()
			Expect(server).ToNot(BeNil())
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest("POST", "/serverlogin"),
					VerifyHeaderKV("X-Tidepool-Server-Name", *provider.NameOutput),
					VerifyHeaderKV("X-Tidepool-Server-Secret", serverSecret),
					VerifyBody(nil),
					RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{sessionToken}})),
			)

			authClientConfig = map[string]interface{}{
				"address":             server.URL(),
				"server_token_secret": authTest.NewServiceSecret(),
				"external": map[string]interface{}{
					"authentication_address":      server.URL(),
					"authorization_address":       server.URL(),
					"server_session_token_secret": serverSecret,
				},
			}
			serviceConfig = map[string]interface{}{
				"secret": authTest.NewServiceSecret(),
				"auth": map[string]interface{}{
					"client": authClientConfig,
				},
			}
			(*provider.ConfigReporterOutput).(*configTest.Reporter).Config = serviceConfig

			svc = serviceService.NewDEPRECATEDService()
			Expect(svc).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
			provider.AssertOutputsEmpty()
		})

		Context("with Terminate after", func() {
			AfterEach(func() {
				svc.Terminate()
			})

			Context("Initialize", func() {
				It("returns an error when the provider is missing", func() {
					Expect(svc.Initialize(nil)).To(MatchError("provider is missing"))
				})

				It("returns an error when the secret is missing", func() {
					delete(serviceConfig, "secret")
					Expect(svc.Initialize(provider)).To(MatchError("secret is missing"))
				})

				It("returns an error when the auth client cannot be initialized", func() {
					delete(authClientConfig, "address")
					Expect(svc.Initialize(provider)).To(MatchError("unable to create auth client; config is invalid; address is missing"))
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

				Context("Secret", func() {
					It("returns the secret", func() {
						Expect(svc.Secret()).To(Equal(serviceConfig["secret"]))
					})
				})

				Context("AuthClient", func() {
					It("returns successfully with server token", func() {
						authClient := svc.AuthClient()
						Expect(authClient).ToNot(BeNil())
						Eventually(authClient.ServerSessionToken).Should(Equal(sessionToken))
					})
				})
			})
		})
	})
})
