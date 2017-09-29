package client_test

// TODO
// import (
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// 	. "github.com/onsi/gomega/ghttp"

// 	"net/http"
// 	"time"

// 	"github.com/tidepool-org/platform/auth/client"
// 	testAuth "github.com/tidepool-org/platform/auth/test"
// 	"github.com/tidepool-org/platform/log"
// 	nullLog "github.com/tidepool-org/platform/log/null"
// 	"github.com/tidepool-org/platform/test"
// 	testHTTP "github.com/tidepool-org/platform/test/http"
// )

// var _ = Describe("Client", func() {
// 	var serverTokenSecret string
// 	var serverTokenTimeout int
// 	var name string
// 	var logger log.Logger
// 	var ctx *testAuth.Context
// 	var serverToken string
// 	var token string

// 	BeforeEach(func() {
// 		serverTokenSecret = test.NewText(32, 128)
// 		serverTokenTimeout = testHTTP.NewTimeout()
// 		name = test.NewText(4, 16)
// 		logger = nullLog.NewLogger()
// 		Expect(logger).ToNot(BeNil())
// 		ctx = testAuth.NewContext()
// 		Expect(ctx).ToNot(BeNil())
// 		serverToken = testAuth.NewSessionToken()
// 		token = testAuth.NewSessionToken()
// 	})

// 	Context("NewClient", func() {
// 		var config *client.Config

// 		BeforeEach(func() {
// 			config = client.NewConfig()
// 			Expect(config).ToNot(BeNil())
// 			Expect(config.Config).ToNot(BeNil())
// 			config.Config.Address = testHTTP.NewAddress()
// 			config.Config.Timeout = time.Duration(testHTTP.NewTimeout()) * time.Second
// 			config.ServerSessionTokenSecret = serverTokenSecret
// 			config.ServerSessionTokenTimeout = time.Duration(serverTokenTimeout) * time.Second
// 		})

// 		It("returns an error if config is missing", func() {
// 			clnt, err := client.NewClient(nil, name, logger)
// 			Expect(err).To(MatchError("config is missing"))
// 			Expect(clnt).To(BeNil())
// 		})

// 		It("returns an error if name is missing", func() {
// 			clnt, err := client.NewClient(config, "", logger)
// 			Expect(err).To(MatchError("name is missing"))
// 			Expect(clnt).To(BeNil())
// 		})

// 		It("returns an error if logger is missing", func() {
// 			clnt, err := client.NewClient(config, name, nil)
// 			Expect(err).To(MatchError("logger is missing"))
// 			Expect(clnt).To(BeNil())
// 		})

// 		It("returns an error if config address is missing", func() {
// 			config.Address = ""
// 			clnt, err := client.NewClient(config, name, logger)
// 			Expect(err).To(MatchError("config is invalid; address is missing"))
// 			Expect(clnt).To(BeNil())
// 		})

// 		It("returns an error if config server token secret is missing", func() {
// 			config.ServerSessionTokenSecret = ""
// 			clnt, err := client.NewClient(config, name, logger)
// 			Expect(err).To(MatchError("config is invalid; server token secret is missing"))
// 			Expect(clnt).To(BeNil())
// 		})

// 		It("returns success", func() {
// 			clnt, err := client.NewClient(config, name, logger)
// 			Expect(err).ToNot(HaveOccurred())
// 			Expect(clnt).ToNot(BeNil())
// 			clnt.Close()
// 		})
// 	})

// 	Context("with started server and new client", func() {
// 		var svr *Server
// 		var config *client.Config
// 		var clnt *client.Client

// 		BeforeEach(func() {
// 			svr = NewServer()
// 			config = client.NewConfig()
// 			Expect(config).ToNot(BeNil())
// 			Expect(config.Config).ToNot(BeNil())
// 			config.Config.Address = svr.URL()
// 			config.ServerSessionTokenSecret = serverTokenSecret
// 		})

// 		JustBeforeEach(func() {
// 			var err error
// 			clnt, err = client.NewClient(config, name, logger)
// 			Expect(err).ToNot(HaveOccurred())
// 			Expect(clnt).ToNot(BeNil())
// 			ctx.AuthClientMock = clnt
// 		})

// 		AfterEach(func() {
// 			clnt.Close()
// 			if svr != nil {
// 				svr.Close()
// 			}
// 		})

// 		Context("Start", func() {
// 			Context("with immediate success of server login", func() {
// 				BeforeEach(func() {
// 					svr.AppendHandlers(
// 						CombineHandlers(
// 							VerifyRequest("POST", "/auth/serverlogin"),
// 							VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 							VerifyBody([]byte{}),
// 							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
// 					)
// 				})

// 				It("returns nil and only invokes server login once", func() {
// 					Expect(clnt.Start()).To(Succeed())
// 					Eventually(func() []*http.Request {
// 						return svr.ReceivedRequests()
// 					}, 10, 1).Should(HaveLen(1))
// 				})
// 			})

// 			Context("with one failure and then success of server login (delay 1 second)", func() {
// 				BeforeEach(func() {
// 					svr.AppendHandlers(
// 						CombineHandlers(
// 							VerifyRequest("POST", "/auth/serverlogin"),
// 							VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 							VerifyBody([]byte{}),
// 							RespondWith(http.StatusBadRequest, nil, nil)),
// 						CombineHandlers(
// 							VerifyRequest("POST", "/auth/serverlogin"),
// 							VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 							VerifyBody([]byte{}),
// 							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
// 					)
// 				})

// 				It("returns nil and only invokes server login twice", func() {
// 					Expect(clnt.Start()).To(Succeed())
// 					Eventually(func() []*http.Request {
// 						return svr.ReceivedRequests()
// 					}, 10, 1).Should(HaveLen(2))

// 				})
// 			})

// 			Context("with two failures and then success of server login (delay 1 second, then 2 seconds)", func() {
// 				BeforeEach(func() {
// 					svr.AppendHandlers(
// 						CombineHandlers(
// 							VerifyRequest("POST", "/auth/serverlogin"),
// 							VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 							VerifyBody([]byte{}),
// 							RespondWith(http.StatusBadRequest, nil, nil)),
// 						CombineHandlers(
// 							VerifyRequest("POST", "/auth/serverlogin"),
// 							VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 							VerifyBody([]byte{}),
// 							RespondWith(http.StatusBadRequest, nil, nil)),
// 						CombineHandlers(
// 							VerifyRequest("POST", "/auth/serverlogin"),
// 							VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 							VerifyBody([]byte{}),
// 							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
// 					)
// 				})

// 				It("returns nil and only invokes server login thrice", func() {
// 					Expect(clnt.Start()).To(Succeed())
// 					Eventually(func() []*http.Request {
// 						return svr.ReceivedRequests()
// 					}, 10, 1).Should(HaveLen(3))
// 				})
// 			})

// 			Context("with one missing session header and then success of server login (delay 1 second)", func() {
// 				BeforeEach(func() {
// 					svr.AppendHandlers(
// 						CombineHandlers(
// 							VerifyRequest("POST", "/auth/serverlogin"),
// 							VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 							VerifyBody([]byte{}),
// 							RespondWith(http.StatusOK, nil, nil)),
// 						CombineHandlers(
// 							VerifyRequest("POST", "/auth/serverlogin"),
// 							VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 							VerifyBody([]byte{}),
// 							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
// 					)
// 				})

// 				It("returns nil and only invokes server login twice", func() {
// 					Expect(clnt.Start()).To(Succeed())
// 					Eventually(func() []*http.Request {
// 						return svr.ReceivedRequests()
// 					}, 10, 1).Should(HaveLen(2))
// 				})
// 			})

// 			Context("with 1 second token timeout", func() {
// 				BeforeEach(func() {
// 					config.ServerSessionTokenTimeout = 1 * time.Second
// 					svr.AppendHandlers(
// 						CombineHandlers(
// 							VerifyRequest("POST", "/auth/serverlogin"),
// 							VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 							VerifyBody([]byte{}),
// 							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
// 						CombineHandlers(
// 							VerifyRequest("POST", "/auth/serverlogin"),
// 							VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 							VerifyBody([]byte{}),
// 							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
// 						CombineHandlers(
// 							VerifyRequest("POST", "/auth/serverlogin"),
// 							VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 							VerifyBody([]byte{}),
// 							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
// 						CombineHandlers(
// 							VerifyRequest("POST", "/auth/serverlogin"),
// 							VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 							VerifyBody([]byte{}),
// 							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
// 					)
// 				})

// 				It("returns nil and only invokes server login thrice", func() {
// 					Expect(clnt.Start()).To(Succeed())
// 					Eventually(func() []*http.Request {
// 						return svr.ReceivedRequests()
// 					}, 10, 1).Should(HaveLen(3))
// 				})
// 			})

// 			It("returns nil and even if server is unreachable", func() {
// 				svr.Close()
// 				svr = nil
// 				Expect(clnt.Start()).To(Succeed())
// 			})
// 		})

// 		Context("with client started and obtained a server token", func() {
// 			BeforeEach(func() {
// 				svr.AppendHandlers(
// 					CombineHandlers(
// 						VerifyRequest("POST", "/auth/serverlogin"),
// 						VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 						VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 						VerifyBody([]byte{}),
// 						RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
// 				)
// 			})

// 			JustBeforeEach(func() {
// 				Expect(clnt.Start()).To(Succeed())
// 			})

// 			Context("ServerSessionToken", func() {
// 				It("returns a server token", func() {
// 					returnedServerSessionToken, err := clnt.ServerSessionToken()
// 					Expect(err).ToNot(HaveOccurred())
// 					Expect(returnedServerSessionToken).To(Equal(serverToken))
// 				})

// 				It("returns error if client is closed", func() {
// 					clnt.Close()
// 					returnedServerSessionToken, err := clnt.ServerSessionToken()
// 					Expect(err).To(MatchError("client is closed"))
// 					Expect(returnedServerSessionToken).To(BeEmpty())
// 				})
// 			})

// 			Context("ValidateSessionToken", func() {
// 				It("returns error if context is missing", func() {
// 					details, err := clnt.ValidateSessionToken(nil, token)
// 					Expect(err).To(MatchError("context is missing"))
// 					Expect(details).To(BeNil())
// 					Expect(svr.ReceivedRequests()).To(HaveLen(1))
// 				})

// 				It("returns error if session token is missing", func() {
// 					details, err := clnt.ValidateSessionToken(ctx, "")
// 					Expect(err).To(MatchError("token is missing"))
// 					Expect(details).To(BeNil())
// 					Expect(svr.ReceivedRequests()).To(HaveLen(1))
// 				})

// 				It("returns error if client is closed", func() {
// 					clnt.Close()
// 					details, err := clnt.ValidateSessionToken(ctx, token)
// 					Expect(err).To(MatchError("client is closed"))
// 					Expect(details).To(BeNil())
// 					Expect(svr.ReceivedRequests()).To(HaveLen(1))
// 				})

// 				It("returns error if the server is not reachable", func() {
// 					svr.Close()
// 					svr = nil
// 					details, err := clnt.ValidateSessionToken(ctx, token)
// 					Expect(err).To(HaveOccurred())
// 					Expect(details).To(BeNil())
// 					Expect(err.Error()).To(HavePrefix("unable to perform request GET "))
// 				})

// 				Context("with an unexpected response", func() {
// 					BeforeEach(func() {
// 						svr.AppendHandlers(
// 							CombineHandlers(
// 								VerifyRequest("GET", "/auth/token/"+token),
// 								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
// 								VerifyBody([]byte{}),
// 								RespondWith(http.StatusBadRequest, nil, nil)),
// 						)
// 					})

// 					It("returns an error", func() {
// 						details, err := clnt.ValidateSessionToken(ctx, token)
// 						Expect(err).To(HaveOccurred())
// 						Expect(details).To(BeNil())
// 						Expect(err.Error()).To(HavePrefix("unexpected response status code 400 from GET "))
// 						Expect(svr.ReceivedRequests()).To(HaveLen(2))
// 					})
// 				})

// 				Context("with an unauthorized response", func() {
// 					BeforeEach(func() {
// 						svr.AppendHandlers(
// 							CombineHandlers(
// 								VerifyRequest("GET", "/auth/token/"+token),
// 								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
// 								VerifyBody([]byte{}),
// 								RespondWith(http.StatusUnauthorized, nil, nil)),
// 						)
// 					})

// 					It("returns an error", func() {
// 						details, err := clnt.ValidateSessionToken(ctx, token)
// 						Expect(err).To(MatchError("unauthorized"))
// 						Expect(details).To(BeNil())
// 						Expect(svr.ReceivedRequests()).To(HaveLen(2))
// 					})
// 				})

// 				Context("with a successful response, but not parseable", func() {
// 					BeforeEach(func() {
// 						svr.AppendHandlers(
// 							CombineHandlers(
// 								VerifyRequest("GET", "/auth/token/"+token),
// 								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
// 								VerifyBody([]byte{}),
// 								RespondWith(http.StatusOK, "}{", nil)),
// 						)
// 					})

// 					It("returns an error", func() {
// 						details, err := clnt.ValidateSessionToken(ctx, token)
// 						Expect(err).To(HaveOccurred())
// 						Expect(details).To(BeNil())
// 						Expect(err.Error()).To(HavePrefix("error decoding JSON response from GET "))
// 						Expect(svr.ReceivedRequests()).To(HaveLen(2))
// 					})
// 				})

// 				Context("with a successful response, but is not a server and missing the user id", func() {
// 					BeforeEach(func() {
// 						svr.AppendHandlers(
// 							CombineHandlers(
// 								VerifyRequest("GET", "/auth/token/"+token),
// 								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
// 								VerifyBody([]byte{}),
// 								RespondWith(http.StatusOK, "{}", nil)),
// 						)
// 					})

// 					It("returns an error", func() {
// 						details, err := clnt.ValidateSessionToken(ctx, token)
// 						Expect(err).To(MatchError("user id is missing"))
// 						Expect(details).To(BeNil())
// 						Expect(svr.ReceivedRequests()).To(HaveLen(2))
// 					})
// 				})

// 				Context("with a successful response and a user id", func() {
// 					BeforeEach(func() {
// 						svr.AppendHandlers(
// 							CombineHandlers(
// 								VerifyRequest("GET", "/auth/token/"+token),
// 								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
// 								VerifyBody([]byte{}),
// 								RespondWith(http.StatusOK, `{"userid": "session-user-id"}`, nil)),
// 						)
// 					})

// 					It("returns the user id", func() {
// 						details, err := clnt.ValidateSessionToken(ctx, token)
// 						Expect(details).ToNot(BeNil())
// 						Expect(err).ToNot(HaveOccurred())
// 						Expect(details.Token()).To(Equal(token))
// 						Expect(details.IsService()).To(BeFalse())
// 						Expect(details.UserID()).To(Equal("session-user-id"))
// 					})
// 				})

// 				Context("with a successful response and is server", func() {
// 					BeforeEach(func() {
// 						svr.AppendHandlers(
// 							CombineHandlers(
// 								VerifyRequest("GET", "/auth/token/"+token),
// 								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
// 								VerifyBody([]byte{}),
// 								RespondWith(http.StatusOK, "{\"isserver\": true}", nil)),
// 						)
// 					})

// 					It("returns is server", func() {
// 						details, err := clnt.ValidateSessionToken(ctx, token)
// 						Expect(details).ToNot(BeNil())
// 						Expect(err).ToNot(HaveOccurred())
// 						Expect(details.Token()).To(Equal(token))
// 						Expect(details.IsService()).To(BeTrue())
// 						Expect(details.UserID()).To(Equal(""))
// 					})
// 				})
// 			})
// 		})

// 		Context("with client started and did NOT obtain a server token", func() {
// 			BeforeEach(func() {
// 				svr.AppendHandlers(
// 					CombineHandlers(
// 						VerifyRequest("POST", "/auth/serverlogin"),
// 						VerifyHeaderKV("X-Tidepool-Server-Name", name),
// 						VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
// 						VerifyBody([]byte{}),
// 						RespondWith(http.StatusBadRequest, nil, nil)),
// 				)
// 			})

// 			JustBeforeEach(func() {
// 				Expect(clnt.Start()).To(Succeed())
// 			})

// 			Context("ServerSessionToken", func() {
// 				It("returns an error", func() {
// 					returnedServiceToken, err := clnt.ServerSessionToken()
// 					Expect(err).To(MatchError("unable to obtain server token"))
// 					Expect(returnedServiceToken).To(BeEmpty())
// 				})
// 			})

// 			Context("ValidateSessionToken", func() {
// 				It("returns an error", func() {
// 					details, err := clnt.ValidateSessionToken(ctx, token)
// 					Expect(err).To(MatchError("unable to obtain server token"))
// 					Expect(details).To(BeNil())
// 					Expect(svr.ReceivedRequests()).To(HaveLen(1))
// 				})
// 			})
// 		})
// 	})
// })
