package service_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// 	. "github.com/onsi/gomega/ghttp"

// 	"net/http"

// 	"github.com/tidepool-org/platform/config"
// 	"github.com/tidepool-org/platform/config/env"
// 	"github.com/tidepool-org/platform/id"
// 	"github.com/tidepool-org/platform/service/service"

// 	_ "github.com/tidepool-org/platform/application/version/test"
// )

// var _ = Describe("DEPRECATEDService", func() {
// 	Context("NewDEPRECATEDService", func() {
// 		It("returns an error if the prefix is missing", func() {
// 			svc, err := service.NewDEPRECATEDService("")
// 			Expect(err).To(MatchError("prefix is missing"))
// 			Expect(svc).To(BeNil())
// 		})

// 		It("returns successfully", func() {
// 			Expect(service.NewDEPRECATEDService("TIDEPOOL")).ToNot(BeNil())
// 		})
// 	})

// 	Context("with started server, config reporter, and new service", func() {
// 		var serverTokenSecret string
// 		var serverToken string
// 		var server *Server
// 		var configReporter config.Reporter
// 		var svc *service.DEPRECATEDService

// 		BeforeEach(func() {
// 			serverTokenSecret = id.New()
// 			serverToken = id.New()
// 			server = NewServer()
// 			Expect(server).ToNot(BeNil())
// 			server.AppendHandlers(
// 				CombineHandlers(
// 					VerifyRequest("POST", "/auth/serverlogin"),
// 					VerifyHeaderKV("X-Tidepool-Server-Name", "service.test"),
// 					VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 					VerifyBody([]byte{}),
// 					RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
// 			)
// 			var err error
// 			configReporter, err = env.NewReporter("TIDEPOOL")
// 			Expect(err).ToNot(HaveOccurred())
// 			Expect(configReporter).ToNot(BeNil())
// 			configReporter = configReporter.WithScopes("service.test", "service", "auth", "client")
// 			configReporter.Set("address", server.URL())
// 			configReporter.Set("timeout", "60")
// 			configReporter.Set("server_token_secret", serverTokenSecret)
// 			svc, err = service.NewDEPRECATEDService("TIDEPOOL")
// 			Expect(err).ToNot(HaveOccurred())
// 			Expect(svc).ToNot(BeNil())
// 		})

// 		AfterEach(func() {
// 			if server != nil {
// 				server.Close()
// 			}
// 		})

// 		Context("Initialize", func() {
// 			It("returns an error if the timeout is invalid during Load", func() {
// 				configReporter.Set("timeout", "abc")
// 				Expect(svc.Initialize()).To(MatchError("unable to load auth client config; timeout is invalid"))
// 			})

// 			It("returns an error if the timeout is invalid during Validate", func() {
// 				configReporter.Set("timeout", "0")
// 				Expect(svc.Initialize()).To(MatchError("unable to create auth client; config is invalid; timeout is invalid"))
// 			})

// 			It("returns successfully", func() {
// 				Expect(svc.Initialize()).To(Succeed())
// 				svc.Terminate()
// 			})
// 		})

// 		Context("with being initialized", func() {
// 			BeforeEach(func() {
// 				Expect(svc.Initialize()).To(Succeed())
// 			})

// 			AfterEach(func() {
// 				svc.Terminate()
// 			})

// 			Context("Terminate", func() {
// 				It("returns successfully", func() {
// 					svc.Terminate()
// 				})
// 			})

// 			Context("AuthClient", func() {
// 				It("returns successfully with server token", func() {
// 					authClient := svc.AuthClient()
// 					Expect(authClient).ToNot(BeNil())
// 					Eventually(authClient.ServerSessionToken).Should(Equal(serverToken))
// 				})
// 			})
// 		})
// 	})
// })
