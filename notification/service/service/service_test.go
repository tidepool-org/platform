package service_test

// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// 	. "github.com/onsi/gomega/ghttp"

// 	"net/http"

// 	"github.com/tidepool-org/platform/config"
// 	"github.com/tidepool-org/platform/config/env"
// 	"github.com/tidepool-org/platform/id"
// 	"github.com/tidepool-org/platform/notification/service/service"

// )

// var _ = Describe("Service", func() {
// 	Context("New", func() {
// 		It("returns an error if unsuccessful", func() {
// 			svc, err := service.New("")
// 			Expect(err).To(HaveOccurred())
// 			Expect(svc).To(BeNil())
// 		})

// 		It("returns successfully", func() {
// 			Expect(service.New("TIDEPOOL")).ToNot(BeNil())
// 		})
// 	})

// 	Context("with started server, config reporter, and new service", func() {
// 		var serverTokenSecret string
// 		var serverToken string
// 		var server *Server
// 		var clientConfigReporter config.Reporter
// 		var storeConfigReporter config.Reporter
// 		var svc *service.Service

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
// 					VerifyBody(nil),
// 					RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
// 			)

// 			configReporter, err := env.NewReporter("TIDEPOOL")
// 			Expect(err).ToNot(HaveOccurred())
// 			Expect(configReporter).ToNot(BeNil())
// 			configReporter = configReporter.WithScopes("service.test", "service")

// 			configReporter.Set("secret", "This is a secret")

// 			clientConfigReporter = configReporter.WithScopes("auth", "client")
// 			clientConfigReporter.Set("address", server.URL())
// 			clientConfigReporter.Set("timeout", "60")
// 			clientConfigReporter.Set("server_token_secret", serverTokenSecret)

// 			storeConfigReporter = configReporter.WithScopes("notification", "store")
// 			storeConfigReporter.Set("address", "http://localhost:1234")
// 			storeConfigReporter.Set("timeout", "60")

// 			configReporter.WithScopes("server").Set("address", "http://localhost:5678")

// 			svc, err = service.New("TIDEPOOL")
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
// 				clientConfigReporter.Set("timeout", "abc")
// 				Expect(svc.Initialize()).To(MatchError("unable to load auth client config; timeout is invalid"))
// 			})

// 			It("returns an error if the timeout is invalid during Load", func() {
// 				storeConfigReporter.Set("timeout", "abc")
// 				Expect(svc.Initialize()).To(MatchError("unable to load notification store config; timeout is invalid"))
// 			})

// 			It("returns an error if the timeout is invalid during Validate", func() {
// 				storeConfigReporter.Set("timeout", "0")
// 				Expect(svc.Initialize()).To(MatchError("unable to create notification store; config is invalid; timeout is invalid"))
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

// 			Context("NotificationStore", func() {
// 				It("returns successfully", func() {
// 					Expect(svc.NotificationStore()).ToNot(BeNil())
// 				})
// 			})

// 			Context("Status", func() {
// 				It("returns successfully", func() {
// 					Expect(svc.Status()).ToNot(BeNil())
// 				})
// 			})
// 		})
// 	})
// })
