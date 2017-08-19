package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"net/http"

	_ "github.com/tidepool-org/platform/application/version/test"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/config/env"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/service/service"
)

var _ = Describe("Service", func() {
	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			svc, err := service.New("")
			Expect(err).To(HaveOccurred())
			Expect(svc).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(service.New("TIDEPOOL")).ToNot(BeNil())
		})
	})

	Context("with started server, config reporter, and new service", func() {
		var serverTokenSecret string
		var serverToken string
		var server *Server
		var clientConfigReporter config.Reporter
		var serverConfigReporter config.Reporter
		var svc *service.Service

		BeforeEach(func() {
			serverTokenSecret = id.New()
			serverToken = id.New()
			server = NewServer()
			Expect(server).ToNot(BeNil())
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest("POST", "/auth/serverlogin"),
					VerifyHeaderKV("X-Tidepool-Server-Name", "service.test"),
					VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
					VerifyBody([]byte{}),
					RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
			)

			configReporter, err := env.NewReporter("TIDEPOOL")
			Expect(err).ToNot(HaveOccurred())
			Expect(configReporter).ToNot(BeNil())
			configReporter = configReporter.WithScopes("service.test", "service")

			clientConfigReporter = configReporter.WithScopes("auth", "client")
			clientConfigReporter.Set("address", server.URL())
			clientConfigReporter.Set("timeout", "60")
			clientConfigReporter.Set("server_token_secret", serverTokenSecret)

			serverConfigReporter = configReporter.WithScopes("server")
			serverConfigReporter.Set("address", "http://localhost:5678")

			svc, err = service.New("TIDEPOOL")
			Expect(err).ToNot(HaveOccurred())
			Expect(svc).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("Initialize", func() {
			It("returns an error if the timeout is invalid during Load", func() {
				clientConfigReporter.Set("timeout", "abc")
				Expect(svc.Initialize()).To(MatchError("service: unable to load auth client config; client: timeout is invalid"))
			})

			It("returns an error if the timeout is invalid during Validate", func() {
				clientConfigReporter.Set("timeout", "0")
				Expect(svc.Initialize()).To(MatchError("service: unable to create auth client; client: config is invalid; client: timeout is invalid"))
			})

			It("returns an error if the address is invalid during Validate", func() {
				serverConfigReporter.Delete("address")
				Expect(svc.Initialize()).To(MatchError("service: unable to create server; server: config is invalid; server: address is missing"))
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
			It("returns an error since it is not initialized", func() {
				Expect(svc.Run()).To(MatchError("service: service not initialized"))
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

			Context("AuthClient", func() {
				It("returns successfully with server token", func() {
					authClient := svc.AuthClient()
					Expect(authClient).ToNot(BeNil())
					Eventually(authClient.ServerToken).Should(Equal(serverToken))
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
