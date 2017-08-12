package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/onsi/gomega/ghttp"

	_ "github.com/tidepool-org/platform/application/version/test"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/config/env"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/service/service"
)

var _ = Describe("Service", func() {
	Context("New", func() {
		It("returns an error if the prefix is missing", func() {
			svc, err := service.New("")
			Expect(err).To(MatchError("application: prefix is missing"))
			Expect(svc).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(service.New("TIDEPOOL")).ToNot(BeNil())
		})
	})

	Context("with started server, config reporter, and new service", func() {
		var serverTokenSecret string
		var serverToken string
		var server *ghttp.Server
		var configReporter config.Reporter
		var oldAddress string
		var oldTimeout string
		var oldServerTokenSecret string
		var svc *service.Service

		BeforeEach(func() {
			serverTokenSecret = id.New()
			serverToken = id.New()
			server = ghttp.NewServer()
			Expect(server).ToNot(BeNil())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/auth/serverlogin"),
					ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "service.test"),
					ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
					ghttp.VerifyBody([]byte{}),
					ghttp.RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
			)
			var err error
			configReporter, err = env.NewReporter("TIDEPOOL")
			Expect(err).ToNot(HaveOccurred())
			Expect(configReporter).ToNot(BeNil())
			configReporter = configReporter.WithScopes("service.test", "auth", "client")
			oldAddress = configReporter.GetWithDefault("address", "")
			configReporter.Set("address", server.URL())
			oldTimeout = configReporter.GetWithDefault("timeout", "")
			configReporter.Set("timeout", "60")
			oldServerTokenSecret = configReporter.GetWithDefault("server_token_secret", "")
			configReporter.Set("server_token_secret", serverTokenSecret)
			svc, err = service.New("TIDEPOOL")
			Expect(err).ToNot(HaveOccurred())
			Expect(svc).ToNot(BeNil())
		})

		AfterEach(func() {
			configReporter.Set("server_token_secret", oldServerTokenSecret)
			configReporter.Set("timeout", oldTimeout)
			configReporter.Set("address", oldAddress)
			if server != nil {
				server.Close()
			}
		})

		Context("Initialize", func() {
			It("returns an error if the timeout is invalid during Load", func() {
				configReporter.Set("timeout", "abc")
				Expect(svc.Initialize()).To(MatchError("service: unable to load auth client config; client: timeout is invalid"))
			})

			It("returns an error if the timeout is invalid during Validate", func() {
				configReporter.Set("timeout", "0")
				Expect(svc.Initialize()).To(MatchError("service: unable to create auth client; client: config is invalid; client: timeout is invalid"))
			})

			It("returns successfully", func() {
				Expect(svc.Initialize()).To(Succeed())
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

			Context("AuthClient", func() {
				It("returns successfully with server token", func() {
					authClient := svc.AuthClient()
					Expect(authClient).ToNot(BeNil())
					Eventually(authClient.ServerToken).Should(Equal(serverToken))
				})
			})
		})
	})
})
