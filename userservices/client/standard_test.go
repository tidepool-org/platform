package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
)

type TestLogger struct{}

func (t *TestLogger) Debug(message string)                               {}
func (t *TestLogger) Info(message string)                                {}
func (t *TestLogger) Warn(message string)                                {}
func (t *TestLogger) Error(message string)                               {}
func (t *TestLogger) WithError(err error) log.Logger                     { return t }
func (t *TestLogger) WithField(key string, value interface{}) log.Logger { return t }
func (t *TestLogger) WithFields(fields log.Fields) log.Logger            { return t }

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

var _ = Describe("Standard", func() {
	var logger *TestLogger
	var context *TestContext

	BeforeEach(func() {
		logger = &TestLogger{}
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
				RequestTimeout:     30,
				ServerTokenSecret:  "I Have A Good Secret!",
				ServerTokenTimeout: 1800,
			}
		})

		It("returns an error if logger is missing", func() {
			standard, err := client.NewStandard(nil, config)
			Expect(err).To(MatchError("client: logger is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config is missing", func() {
			standard, err := client.NewStandard(logger, nil)
			Expect(err).To(MatchError("client: config is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config address is invalid", func() {
			config.Address = ""
			standard, err := client.NewStandard(logger, config)
			Expect(err).To(MatchError("client: config is invalid; client: address is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config request timeout is invalid", func() {
			config.RequestTimeout = -1
			standard, err := client.NewStandard(logger, config)
			Expect(err).To(MatchError("client: config is invalid; client: request timeout is invalid"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config server token secret is invalid", func() {
			config.ServerTokenSecret = ""
			standard, err := client.NewStandard(logger, config)
			Expect(err).To(MatchError("client: config is invalid; client: server token secret is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config server token timeout is invalid", func() {
			config.ServerTokenTimeout = -1
			standard, err := client.NewStandard(logger, config)
			Expect(err).To(MatchError("client: config is invalid; client: server token timeout is invalid"))
			Expect(standard).To(BeNil())
		})

		It("returns success", func() {
			standard, err := client.NewStandard(logger, config)
			Expect(err).ToNot(HaveOccurred())
			Expect(standard).ToNot(BeNil())
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
				RequestTimeout:     30,
				ServerTokenSecret:  " I Have A Good Secret! ",
				ServerTokenTimeout: 1800,
			}
		})

		JustBeforeEach(func() {
			var err error
			standard, err = client.NewStandard(logger, config)
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
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "dataservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"x-tidepool-session-token": []string{"test-session-token"}})),
					)
				})

				It("returns nil and only invokes server login once", func() {
					Expect(standard.Start()).To(BeNil())
					time.Sleep(2 * time.Second)
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with one failure and then success of server login (delay 1 second)", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "dataservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "dataservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"x-tidepool-session-token": []string{"test-session-token"}})),
					)
				})

				It("returns nil and only invokes server login twice", func() {
					Expect(standard.Start()).To(BeNil())
					time.Sleep(4 * time.Second)
					Expect(server.ReceivedRequests()).To(HaveLen(2))
				})
			})

			Context("with two failures and then success of server login (delay 1 second, then 2 seconds)", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "dataservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "dataservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "dataservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"x-tidepool-session-token": []string{"test-session-token"}})),
					)
				})

				It("returns nil and only invokes server login thrice", func() {
					Expect(standard.Start()).To(BeNil())
					time.Sleep(8 * time.Second)
					Expect(server.ReceivedRequests()).To(HaveLen(3))
				})
			})

			Context("with one missing session header and then success of server login (delay 1 second)", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "dataservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, nil)),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "dataservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"x-tidepool-session-token": []string{"test-session-token"}})),
					)
				})

				It("returns nil and only invokes server login twice", func() {
					Expect(standard.Start()).To(BeNil())
					time.Sleep(4 * time.Second)
					Expect(server.ReceivedRequests()).To(HaveLen(2))
				})
			})

			Context("with 1 second token timeout", func() {
				BeforeEach(func() {
					config.ServerTokenTimeout = 1
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "dataservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"x-tidepool-session-token": []string{"test-session-token"}})),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "dataservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"x-tidepool-session-token": []string{"test-session-token"}})),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/auth/serverlogin"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "dataservices"),
							ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, http.Header{"x-tidepool-session-token": []string{"test-session-token"}})),
					)
				})

				It("returns nil and only invokes server login thrice", func() {
					Expect(standard.Start()).To(BeNil())
					time.Sleep(2500 * time.Millisecond)
					Expect(server.ReceivedRequests()).To(HaveLen(3))
				})
			})

			It("returns nil and even if server is unreachable", func() {
				server.Close()
				server = nil
				Expect(standard.Start()).To(BeNil())
			})
		})

		Context("with client started and obtained a server token", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/auth/serverlogin"),
						ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "dataservices"),
						ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
						ghttp.VerifyBody([]byte{}),
						ghttp.RespondWith(http.StatusOK, nil, http.Header{"x-tidepool-session-token": []string{"test-session-token"}})),
				)
			})

			JustBeforeEach(func() {
				Expect(standard.Start()).To(BeNil())
			})

			Context("ValidateUserSession", func() {
				It("returns error if context is missing", func() {
					sessionToken, err := standard.ValidateUserSession(nil, "session-token")
					Expect(err).To(MatchError("client: context is missing"))
					Expect(sessionToken).To(Equal(""))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if session token is missing", func() {
					sessionToken, err := standard.ValidateUserSession(context, "")
					Expect(err).To(MatchError("client: session token is missing"))
					Expect(sessionToken).To(Equal(""))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if client is closed", func() {
					standard.Close()
					sessionToken, err := standard.ValidateUserSession(context, "session-token")
					Expect(err).To(MatchError("client: client is closed"))
					Expect(sessionToken).To(Equal(""))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if the server is not reachable", func() {
					server.Close()
					server = nil
					sessionToken, err := standard.ValidateUserSession(context, "session-token")
					Expect(err).To(HaveOccurred())
					Expect(sessionToken).To(Equal(""))
					Expect(err.Error()).To(HavePrefix("client: unable to perform request GET "))
				})

				It("returns error if the context request is missing", func() {
					context.TestRequest = nil
					sessionToken, err := standard.ValidateUserSession(context, "session-token")
					Expect(err).To(MatchError("client: unable to copy request trace; service: source request is missing"))
					Expect(sessionToken).To(Equal(""))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				Context("with an unexpected response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/auth/token/session-token"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						)
					})

					It("returns an error", func() {
						sessionToken, err := standard.ValidateUserSession(context, "session-token")
						Expect(err).To(HaveOccurred())
						Expect(sessionToken).To(Equal(""))
						Expect(err.Error()).To(HavePrefix("client: unexpected response status code 400 from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an unauthorized response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/auth/token/session-token"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusUnauthorized, nil, nil)),
						)
					})

					It("returns an error", func() {
						sessionToken, err := standard.ValidateUserSession(context, "session-token")
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(sessionToken).To(Equal(""))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an successful response, but not parseable", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/auth/token/session-token"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "}{", nil)),
						)
					})

					It("returns an error", func() {
						sessionToken, err := standard.ValidateUserSession(context, "session-token")
						Expect(err).To(HaveOccurred())
						Expect(sessionToken).To(Equal(""))
						Expect(err.Error()).To(HavePrefix("client: error decoding JSON response from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an successful response, but missing the user id", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/auth/token/session-token"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{}", nil)),
						)
					})

					It("returns an error", func() {
						sessionToken, err := standard.ValidateUserSession(context, "session-token")
						Expect(err).To(MatchError("client: user id is missing"))
						Expect(sessionToken).To(Equal(""))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an successful response and a user id", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/auth/token/session-token"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{\"userid\": \"session-user-id\"}", nil)),
						)
					})

					It("returns the user id", func() {
						sessionToken, err := standard.ValidateUserSession(context, "session-token")
						Expect(err).ToNot(HaveOccurred())
						Expect(sessionToken).To(Equal("session-user-id"))
					})
				})
			})

			Context("ValidateTargetUserPermissions", func() {
				It("returns error if context is missing", func() {
					err := standard.ValidateTargetUserPermissions(nil, "request-user-id", "target-user-id", client.ViewPermissions)
					Expect(err).To(MatchError("client: context is missing"))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if request user id is missing", func() {
					err := standard.ValidateTargetUserPermissions(context, "", "target-user-id", client.ViewPermissions)
					Expect(err).To(MatchError("client: request user id is missing"))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if target user id is missing", func() {
					err := standard.ValidateTargetUserPermissions(context, "request-user-id", "", client.ViewPermissions)
					Expect(err).To(MatchError("client: target user id is missing"))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if target permissions is missing", func() {
					err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", nil)
					Expect(err).To(MatchError("client: target permissions is missing"))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if client is closed", func() {
					standard.Close()
					err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.ViewPermissions)
					Expect(err).To(MatchError("client: client is closed"))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if server is unreachable", func() {
					server.Close()
					server = nil
					err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.ViewPermissions)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("client: unable to perform request GET "))
				})

				It("returns error if the context request is missing", func() {
					context.TestRequest = nil
					err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.ViewPermissions)
					Expect(err).To(MatchError("client: unable to copy request trace; service: source request is missing"))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				Context("with an unexpected response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.ViewPermissions)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(HavePrefix("client: unexpected response status code 400 from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an unauthorized response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusUnauthorized, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.ViewPermissions)
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a not found response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusNotFound, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.ViewPermissions)
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an successful response, but not parseable", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "}{", nil)),
						)
					})

					It("returns an error", func() {
						err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.ViewPermissions)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(HavePrefix("client: error decoding JSON response from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an successful response, but with no permissions", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{}", nil)),
						)
					})

					It("returns an error", func() {
						err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.ViewPermissions)
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an successful response, but without the requested permission", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{\"upload\": {}}", nil)),
						)
					})

					It("returns an error", func() {
						err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.ViewPermissions)
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an successful response, but without the requested permissions", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{\"upload\": {}}", nil)),
						)
					})

					It("returns an error", func() {
						err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.Permissions{"upload": {}, "view": {}})
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an successful response and the requested permission", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{\"view\": {}}", nil)),
						)
					})

					It("returns no error", func() {
						err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.ViewPermissions)
						Expect(err).ToNot(HaveOccurred())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an successful response and the requested permissions", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{\"view\": {}, \"upload\": {}}", nil)),
						)
					})

					It("returns no error", func() {
						err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.Permissions{"upload": {}, "view": {}})
						Expect(err).ToNot(HaveOccurred())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an successful response and the root permission", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/target-user-id/request-user-id"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{\"root\": {}}", nil)),
						)
					})

					It("returns no error", func() {
						err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.ViewPermissions)
						Expect(err).ToNot(HaveOccurred())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})
			})

			Context("GetUserGroupID", func() {
				It("returns error if context is missing", func() {
					groupID, err := standard.GetUserGroupID(nil, "user-id")
					Expect(err).To(MatchError("client: context is missing"))
					Expect(groupID).To(Equal(""))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if user id is missing", func() {
					groupID, err := standard.GetUserGroupID(context, "")
					Expect(err).To(MatchError("client: user id is missing"))
					Expect(groupID).To(Equal(""))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if client is closed", func() {
					standard.Close()
					groupID, err := standard.GetUserGroupID(context, "user-id")
					Expect(err).To(MatchError("client: client is closed"))
					Expect(groupID).To(Equal(""))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if server is unreachable", func() {
					server.Close()
					server = nil
					groupID, err := standard.GetUserGroupID(context, "user-id")
					Expect(err).To(HaveOccurred())
					Expect(groupID).To(Equal(""))
					Expect(err.Error()).To(HavePrefix("client: unable to perform request GET "))
				})

				It("returns error if the context request is missing", func() {
					context.TestRequest = nil
					groupID, err := standard.GetUserGroupID(context, "user-id")
					Expect(err).To(MatchError("client: unable to copy request trace; service: source request is missing"))
					Expect(groupID).To(Equal(""))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				Context("with an unexpected response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/metadata/user-id/private/uploads"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						)
					})

					It("returns an error", func() {
						groupID, err := standard.GetUserGroupID(context, "user-id")
						Expect(err).To(HaveOccurred())
						Expect(groupID).To(Equal(""))
						Expect(err.Error()).To(HavePrefix("client: unexpected response status code 400 from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an unauthorized response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/metadata/user-id/private/uploads"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusUnauthorized, nil, nil)),
						)
					})

					It("returns an error", func() {
						groupID, err := standard.GetUserGroupID(context, "user-id")
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(groupID).To(Equal(""))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an successful response, but not parseable", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/metadata/user-id/private/uploads"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "}{", nil)),
						)
					})

					It("returns an error", func() {
						groupID, err := standard.GetUserGroupID(context, "user-id")
						Expect(err).To(HaveOccurred())
						Expect(groupID).To(Equal(""))
						Expect(err.Error()).To(HavePrefix("client: error decoding JSON response from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an successful response, but missing the group id", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/metadata/user-id/private/uploads"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{}", nil)),
						)
					})

					It("returns an error", func() {
						groupID, err := standard.GetUserGroupID(context, "user-id")
						Expect(err).To(MatchError("client: group id is missing"))
						Expect(groupID).To(Equal(""))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an successful response and a group id", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/metadata/user-id/private/uploads"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-session-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{\"id\": \"session-group-id\"}", nil)),
						)
					})

					It("returns the group id", func() {
						groupID, err := standard.GetUserGroupID(context, "user-id")
						Expect(err).ToNot(HaveOccurred())
						Expect(groupID).To(Equal("session-group-id"))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})
			})
		})

		Context("with client started and did NOT obtain a server token", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/auth/serverlogin"),
						ghttp.VerifyHeaderKV("X-Tidepool-Server-Name", "dataservices"),
						ghttp.VerifyHeaderKV("X-Tidepool-Server-Secret", "I Have A Good Secret!"),
						ghttp.VerifyBody([]byte{}),
						ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
				)
			})

			JustBeforeEach(func() {
				Expect(standard.Start()).To(BeNil())
			})

			Context("ValidateUserSession", func() {
				It("returns an error", func() {
					sessionToken, err := standard.ValidateUserSession(context, "session-token")
					Expect(err).To(HaveOccurred())
					Expect(sessionToken).To(Equal(""))
					Expect(err.Error()).To(HavePrefix("client: unable to obtain server token for GET "))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("ValidateTargetUserPermissions", func() {
				It("returns an error", func() {
					err := standard.ValidateTargetUserPermissions(context, "request-user-id", "target-user-id", client.ViewPermissions)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("client: unable to obtain server token for GET "))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("GetUserGroupID", func() {
				It("returns an error", func() {
					groupID, err := standard.GetUserGroupID(context, "user-id")
					Expect(err).To(HaveOccurred())
					Expect(groupID).To(Equal(""))
					Expect(err.Error()).To(HavePrefix("client: unable to obtain server token for GET "))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
