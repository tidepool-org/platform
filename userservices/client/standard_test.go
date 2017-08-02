package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
)

type TestContext struct {
	TestLogger  log.Logger
	TestRequest *rest.Request
}

func (t *TestContext) Logger() log.Logger                                                      { return t.TestLogger }
func (t *TestContext) Request() *rest.Request                                                  { return t.TestRequest }
func (t *TestContext) Response() rest.ResponseWriter                                           { return nil }
func (t *TestContext) RespondWithError(err *service.Error)                                     {}
func (t *TestContext) RespondWithInternalServerFailure(message string, failure ...interface{}) {}
func (t *TestContext) RespondWithStatusAndErrors(statusCode int, errors []*service.Error)      {}
func (t *TestContext) RespondWithStatusAndData(statusCode int, data interface{})               {}

var _ = Describe("Standard", func() {
	var logger log.Logger
	var context *TestContext

	BeforeEach(func() {
		logger = null.NewLogger()
		context = &TestContext{
			TestLogger:  logger,
			TestRequest: &rest.Request{},
		}
	})

	Context("NewStandard", func() {
		var config *client.Config

		BeforeEach(func() {
			config = &client.Config{
				Address:            "http://localhost:1234",
				Timeout:            30 * time.Second,
				ServerTokenSecret:  "I Have A Good Secret!",
				ServerTokenTimeout: 1800 * time.Second,
			}
		})

		It("returns an error if logger is missing", func() {
			standard, err := client.NewStandard(nil, "testservices", config)
			Expect(err).To(MatchError("client: logger is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if name is missing", func() {
			standard, err := client.NewStandard(logger, "", config)
			Expect(err).To(MatchError("client: name is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config is missing", func() {
			standard, err := client.NewStandard(logger, "testservices", nil)
			Expect(err).To(MatchError("client: config is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config address is missing", func() {
			config.Address = ""
			standard, err := client.NewStandard(logger, "testservices", config)
			Expect(err).To(MatchError("client: config is invalid; client: address is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config timeout is invalid", func() {
			config.Timeout = 0
			standard, err := client.NewStandard(logger, "testservices", config)
			Expect(err).To(MatchError("client: config is invalid; client: timeout is invalid"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config server token secret is invalid", func() {
			config.ServerTokenSecret = ""
			standard, err := client.NewStandard(logger, "testservices", config)
			Expect(err).To(MatchError("client: config is invalid; client: server token secret is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config server token timeout is invalid", func() {
			config.ServerTokenTimeout = 0
			standard, err := client.NewStandard(logger, "testservices", config)
			Expect(err).To(MatchError("client: config is invalid; client: server token timeout is invalid"))
			Expect(standard).To(BeNil())
		})

		It("returns success", func() {
			standard, err := client.NewStandard(logger, "testservices", config)
			Expect(err).ToNot(HaveOccurred())
			Expect(standard).ToNot(BeNil())
			standard.Close()
		})
	})

	Context("with server", func() {
		var server *ghttp.Server
		var config *client.Config
		var standard *client.Standard

		BeforeEach(func() {
			server = ghttp.NewServer()
			config = &client.Config{
				Address:            server.URL(),
				Timeout:            30 * time.Second,
				ServerTokenSecret:  " I Have A Good Secret! ",
				ServerTokenTimeout: 1800 * time.Second,
			}
		})

		JustBeforeEach(func() {
			var err error
			standard, err = client.NewStandard(logger, "testservices", config)
			Expect(err).ToNot(HaveOccurred())
			Expect(standard).ToNot(BeNil())
		})

		AfterEach(func() {
			standard.Close()
			if server != nil {
				server.Close()
			}
		})

		Context("Start", func() {
			Context("with immediate success of server login", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{"test-authentication-token"}})),
					)
				})

				It("returns nil and only invokes server login once", func() {
					Expect(standard.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return server.ReceivedRequests()
					}, 10, 1).Should(HaveLen(1))
				})
			})

			Context("with one failure and then success of server login (delay 1 second)", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{"test-authentication-token"}})),
					)
				})

				It("returns nil and only invokes server login twice", func() {
					Expect(standard.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return server.ReceivedRequests()
					}, 10, 1).Should(HaveLen(2))

				})
			})

			Context("with two failures and then success of server login (delay 1 second, then 2 seconds)", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{"test-authentication-token"}})),
					)
				})

				It("returns nil and only invokes server login thrice", func() {
					Expect(standard.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return server.ReceivedRequests()
					}, 10, 1).Should(HaveLen(3))
				})
			})

			Context("with one missing session header and then success of server login (delay 1 second)", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, nil)),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{"test-authentication-token"}})),
					)
				})

				It("returns nil and only invokes server login twice", func() {
					Expect(standard.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return server.ReceivedRequests()
					}, 10, 1).Should(HaveLen(2))
				})
			})

			Context("with 1 second token timeout", func() {
				BeforeEach(func() {
					config.ServerTokenTimeout = 1 * time.Second
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{"test-authentication-token"}})),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{"test-authentication-token"}})),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{"test-authentication-token"}})),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{"test-authentication-token"}})),
					)
				})

				It("returns nil and only invokes server login thrice", func() {
					Expect(standard.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return server.ReceivedRequests()
					}, 10, 1).Should(HaveLen(3))
				})
			})

			It("returns nil and even if server is unreachable", func() {
				server.Close()
				server = nil
				Expect(standard.Start()).To(Succeed())
			})
		})

		Context("with client started and obtained a server token", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/auth/serverlogin"),
						ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
						ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
						ghttp.VerifyBody([]byte{}),
						ghttp.RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{"test-authentication-token"}})),
				)
			})

			JustBeforeEach(func() {
				Expect(standard.Start()).To(Succeed())
			})

			Context("ValidateAuthenticationToken", func() {
				It("returns error if context is missing", func() {
					authenticationDetails, err := standard.ValidateAuthenticationToken(nil, "test-authentication-token")
					Expect(err).To(MatchError("client: context is missing"))
					Expect(authenticationDetails).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if session token is missing", func() {
					authenticationDetails, err := standard.ValidateAuthenticationToken(context, "")
					Expect(err).To(MatchError("client: authentication token is missing"))
					Expect(authenticationDetails).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if client is closed", func() {
					standard.Close()
					authenticationDetails, err := standard.ValidateAuthenticationToken(context, "test-authentication-token")
					Expect(err).To(MatchError("client: client is closed"))
					Expect(authenticationDetails).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if the server is not reachable", func() {
					server.Close()
					server = nil
					authenticationDetails, err := standard.ValidateAuthenticationToken(context, "test-authentication-token")
					Expect(err).To(HaveOccurred())
					Expect(authenticationDetails).To(BeNil())
					Expect(err.Error()).To(HavePrefix("client: unable to perform request GET "))
				})

				It("returns error if the context request is missing", func() {
					context.TestRequest = nil
					authenticationDetails, err := standard.ValidateAuthenticationToken(context, "test-authentication-token")
					Expect(err).To(MatchError("client: unable to copy request trace; service: source request is missing"))
					Expect(authenticationDetails).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				Context("with an unexpected response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/auth/token/test-authentication-token"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						)
					})

					It("returns an error", func() {
						authenticationDetails, err := standard.ValidateAuthenticationToken(context, "test-authentication-token")
						Expect(err).To(HaveOccurred())
						Expect(authenticationDetails).To(BeNil())
						Expect(err.Error()).To(HavePrefix("client: unexpected response status code 400 from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an unauthorized response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/auth/token/test-authentication-token"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusUnauthorized, nil, nil)),
						)
					})

					It("returns an error", func() {
						authenticationDetails, err := standard.ValidateAuthenticationToken(context, "test-authentication-token")
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(authenticationDetails).To(BeNil())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response, but not parseable", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/auth/token/test-authentication-token"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "}{", nil)),
						)
					})

					It("returns an error", func() {
						authenticationDetails, err := standard.ValidateAuthenticationToken(context, "test-authentication-token")
						Expect(err).To(HaveOccurred())
						Expect(authenticationDetails).To(BeNil())
						Expect(err.Error()).To(HavePrefix("client: error decoding JSON response from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response, but is not a server and missing the user id", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/auth/token/test-authentication-token"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{}", nil)),
						)
					})

					It("returns an error", func() {
						authenticationDetails, err := standard.ValidateAuthenticationToken(context, "test-authentication-token")
						Expect(err).To(MatchError("client: user id is missing"))
						Expect(authenticationDetails).To(BeNil())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response and a user id", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/auth/token/test-authentication-token"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, `{"userid": "session-user-id"}`, nil)),
						)
					})

					It("returns the user id", func() {
						authenticationDetails, err := standard.ValidateAuthenticationToken(context, "test-authentication-token")
						Expect(authenticationDetails).ToNot(BeNil())
						Expect(err).ToNot(HaveOccurred())
						Expect(authenticationDetails.Token()).To(Equal("test-authentication-token"))
						Expect(authenticationDetails.IsServer()).To(BeFalse())
						Expect(authenticationDetails.UserID()).To(Equal("session-user-id"))
					})
				})

				Context("with a successful response and is server", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/auth/token/test-authentication-token"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{\"isserver\": true}", nil)),
						)
					})

					It("returns is server", func() {
						authenticationDetails, err := standard.ValidateAuthenticationToken(context, "test-authentication-token")
						Expect(authenticationDetails).ToNot(BeNil())
						Expect(err).ToNot(HaveOccurred())
						Expect(authenticationDetails.Token()).To(Equal("test-authentication-token"))
						Expect(authenticationDetails.IsServer()).To(BeTrue())
						Expect(authenticationDetails.UserID()).To(Equal(""))
					})
				})
			})

			Context("GetUserPermissions", func() {
				It("returns error if context is missing", func() {
					permissions, err := standard.GetUserPermissions(nil, "request-user-id", "target-user-id")
					Expect(err).To(MatchError("client: context is missing"))
					Expect(permissions).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if request user id is missing", func() {
					permissions, err := standard.GetUserPermissions(context, "", "target-user-id")
					Expect(err).To(MatchError("client: request user id is missing"))
					Expect(permissions).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if target user id is missing", func() {
					permissions, err := standard.GetUserPermissions(context, "request-user-id", "")
					Expect(err).To(MatchError("client: target user id is missing"))
					Expect(permissions).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if client is closed", func() {
					standard.Close()
					permissions, err := standard.GetUserPermissions(context, "request-user-id", "target-user-id")
					Expect(err).To(MatchError("client: client is closed"))
					Expect(permissions).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if server is unreachable", func() {
					server.Close()
					server = nil
					permissions, err := standard.GetUserPermissions(context, "request-user-id", "target-user-id")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("client: unable to perform request GET "))
					Expect(permissions).To(BeNil())
				})

				It("returns error if the context request is missing", func() {
					context.TestRequest = nil
					permissions, err := standard.GetUserPermissions(context, "request-user-id", "target-user-id")
					Expect(err).To(MatchError("client: unable to copy request trace; service: source request is missing"))
					Expect(permissions).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				Context("with an unexpected response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						)
					})

					It("returns an error", func() {
						permissions, err := standard.GetUserPermissions(context, "request-user-id", "target-user-id")
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(HavePrefix("client: unexpected response status code 400 from GET "))
						Expect(permissions).To(BeNil())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an unauthorized response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusUnauthorized, nil, nil)),
						)
					})

					It("returns an error", func() {
						permissions, err := standard.GetUserPermissions(context, "request-user-id", "target-user-id")
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(permissions).To(BeNil())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a not found response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusNotFound, nil, nil)),
						)
					})

					It("returns an error", func() {
						permissions, err := standard.GetUserPermissions(context, "request-user-id", "target-user-id")
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(permissions).To(BeNil())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response, but not parseable", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "}{", nil)),
						)
					})

					It("returns an error", func() {
						permissions, err := standard.GetUserPermissions(context, "request-user-id", "target-user-id")
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(HavePrefix("client: error decoding JSON response from GET "))
						Expect(permissions).To(BeNil())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response, but with no permissions", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{}", nil)),
						)
					})

					It("returns an error", func() {
						Expect(standard.GetUserPermissions(context, "request-user-id", "target-user-id")).To(BeEmpty())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response with upload and view permissions", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, `{"upload": {}, "view": {}}`, nil)),
						)
					})

					It("returns an error", func() {
						Expect(standard.GetUserPermissions(context, "request-user-id", "target-user-id")).To(Equal(client.Permissions{
							client.UploadPermission: client.Permission{},
							client.ViewPermission:   client.Permission{},
						}))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response with owner permissions that already includes upload permissions", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, `{"root": {"root-inner": "unused"}, "upload": {}}`, nil)),
						)
					})

					It("returns an error", func() {
						Expect(standard.GetUserPermissions(context, "request-user-id", "target-user-id")).To(Equal(client.Permissions{
							client.OwnerPermission:  client.Permission{"root-inner": "unused"},
							client.UploadPermission: client.Permission{},
							client.ViewPermission:   client.Permission{"root-inner": "unused"},
						}))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response with owner permissions that already includes view permissions", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, `{"root": {"root-inner": "unused"}, "view": {}}`, nil)),
						)
					})

					It("returns an error", func() {
						Expect(standard.GetUserPermissions(context, "request-user-id", "target-user-id")).To(Equal(client.Permissions{
							client.OwnerPermission:  client.Permission{"root-inner": "unused"},
							client.UploadPermission: client.Permission{"root-inner": "unused"},
							client.ViewPermission:   client.Permission{},
						}))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response with owner permissions that already includes upload and view permissions", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, `{"root": {"root-inner": "unused"}, "upload": {}, "view": {}}`, nil)),
						)
					})

					It("returns an error", func() {
						Expect(standard.GetUserPermissions(context, "request-user-id", "target-user-id")).To(Equal(client.Permissions{
							client.OwnerPermission:  client.Permission{"root-inner": "unused"},
							client.UploadPermission: client.Permission{},
							client.ViewPermission:   client.Permission{},
						}))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})
			})

			Context("ServerToken", func() {
				It("returns a server token", func() {
					serverToken, err := standard.ServerToken()
					Expect(err).ToNot(HaveOccurred())
					Expect(serverToken).ToNot(BeEmpty())
				})

				It("returns error if client is closed", func() {
					standard.Close()
					serverToken, err := standard.ServerToken()
					Expect(err).To(MatchError("client: client is closed"))
					Expect(serverToken).To(BeEmpty())
				})
			})
		})

		Context("with client started and did NOT obtain a server token", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/auth/serverlogin"),
						ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "testservices"),
						ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
						ghttp.VerifyBody([]byte{}),
						ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
				)
			})

			JustBeforeEach(func() {
				Expect(standard.Start()).To(Succeed())
			})

			Context("ValidateAuthenticationToken", func() {
				It("returns an error", func() {
					authenticationDetails, err := standard.ValidateAuthenticationToken(context, "test-authentication-token")
					Expect(err).To(HaveOccurred())
					Expect(authenticationDetails).To(BeNil())
					Expect(err.Error()).To(HavePrefix("client: unable to obtain server token for GET "))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("GetUserPermissions", func() {
				It("returns an error", func() {
					permissions, err := standard.GetUserPermissions(context, "request-user-id", "target-user-id")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("client: unable to obtain server token for GET "))
					Expect(permissions).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("ServerToken", func() {
				It("returns an error", func() {
					serverToken, err := standard.ServerToken()
					Expect(err).To(MatchError("client: unable to obtain server token"))
					Expect(serverToken).To(BeEmpty())
				})
			})
		})
	})
})
