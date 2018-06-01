package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"net/http"

	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/config/env"
	"github.com/tidepool-org/platform/service/service"

	_ "github.com/tidepool-org/platform/application/version/test"
)

var _ = Describe("DEPRECATEDService", func() {
	Context("NewDEPRECATEDService", func() {
		It("returns an error if the prefix is missing", func() {
			svc, err := service.NewDEPRECATEDService("")
			Expect(err).To(MatchError("prefix is missing"))
			Expect(svc).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(service.NewDEPRECATEDService("TIDEPOOL")).ToNot(BeNil())
		})
	})

	Context("with started server, config reporter, and new service", func() {
		var svc *service.DEPRECATEDService
		var serverSessionTokenSecret string
		var serverSessionToken string
		var serviceSecret string
		var server *Server
		var serviceConfigReporter config.Reporter
		var authClientConfigReporter config.Reporter

		BeforeEach(func() {
			var err error
			svc, err = service.NewDEPRECATEDService("TIDEPOOL")
			Expect(err).ToNot(HaveOccurred())
			Expect(svc).ToNot(BeNil())

			serverSessionTokenSecret = testAuth.NewServiceSecret()
			serverSessionToken = testAuth.NewSessionToken()
			serviceSecret = testAuth.NewServiceSecret()

			server = NewServer()
			Expect(server).ToNot(BeNil())
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest("POST", "/auth/serverlogin"),
					VerifyHeaderKV("X-Tidepool-Server-Name", "service.test"),
					VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
					VerifyBody([]byte{}),
					RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverSessionToken}})),
			)

			configReporter, err := env.NewReporter("TIDEPOOL")
			Expect(err).ToNot(HaveOccurred())
			Expect(configReporter).ToNot(BeNil())

			serviceConfigReporter = configReporter.WithScopes("service.test", "service")
			serviceConfigReporter.Set("secret", serviceSecret)

			authClientConfigReporter = serviceConfigReporter.WithScopes("auth", "client")
			authClientConfigReporter.Set("address", server.URL())
			authClientConfigReporter.Set("timeout", "60")

			externalConfigReporter := authClientConfigReporter.WithScopes("external")
			externalConfigReporter.Set("address", server.URL())
			externalConfigReporter.Set("server_session_token_secret", serverSessionTokenSecret)
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("Initialize", func() {
			It("returns an error if the secret is missing", func() {
				serviceConfigReporter.Delete("secret")
				Expect(svc.Initialize()).To(MatchError("secret is missing"))
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

			Context("Secret", func() {
				It("returns the secret", func() {
					Expect(svc.Secret()).To(Equal(serviceSecret))
				})
			})

			Context("AuthClient", func() {
				It("returns successfully with server token", func() {
					authClient := svc.AuthClient()
					Expect(authClient).ToNot(BeNil())
					Eventually(authClient.ServerSessionToken).Should(Equal(serverSessionToken))
				})
			})
		})
	})
})
