package client_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	authClient "github.com/tidepool-org/platform/auth/client"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
	"github.com/tidepool-org/platform/times"
)

var _ = Describe("Client", func() {
	var serverSessionToken string
	var lgr log.Logger
	var ctx context.Context
	var mockController *gomock.Controller
	var mockServerSessionTokenProvider *authTest.MockServerSessionTokenProvider
	var serverSessionTokenSecret string
	var name string

	BeforeEach(func() {
		serverSessionToken = authTest.NewSessionToken()
		lgr = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), lgr)
		mockController, ctx = gomock.WithContext(ctx, GinkgoT())
		mockServerSessionTokenProvider = authTest.NewMockServerSessionTokenProvider(mockController)
		mockServerSessionTokenProvider.EXPECT().ServerSessionToken().Return(serverSessionToken, nil).AnyTimes()
		ctx = auth.NewContextWithServerSessionTokenProvider(ctx, mockServerSessionTokenProvider)
		serverSessionTokenSecret = authTest.NewServiceSecret()
		name = test.RandomStringFromRangeAndCharset(4, 16, test.CharsetAlphaNumeric)
	})

	Context("NewClient", func() {
		var serverSessionTokenTimeout int
		var config *authClient.Config
		var authorizeAs platform.AuthorizeAs

		BeforeEach(func() {
			serverSessionTokenTimeout = testHttp.NewTimeout()
			config = authClient.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Config).ToNot(BeNil())
			config.Config.Address = testHttp.NewAddress()
			config.Config.UserAgent = testHttp.NewUserAgent()
			config.Config.ServiceSecret = authTest.NewServiceSecret()
			config.ExternalConfig.Address = testHttp.NewAddress()
			config.ExternalConfig.UserAgent = testHttp.NewUserAgent()
			config.ExternalConfig.ServerSessionTokenSecret = serverSessionTokenSecret
			config.ExternalConfig.ServerSessionTokenTimeout = time.Duration(serverSessionTokenTimeout) * time.Second
			authorizeAs = platform.AuthorizeAsService
		})

		It("returns an error if config is missing", func() {
			client, err := authClient.NewClient(nil, authorizeAs, name, lgr)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error if name is missing", func() {
			client, err := authClient.NewClient(config, authorizeAs, "", lgr)
			errorsTest.ExpectEqual(err, errors.New("name is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error if logger is missing", func() {
			client, err := authClient.NewClient(config, authorizeAs, name, nil)
			errorsTest.ExpectEqual(err, errors.New("logger is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error if config address is missing", func() {
			config.Config.Address = ""
			client, err := authClient.NewClient(config, authorizeAs, name, lgr)
			errorsTest.ExpectEqual(err, errors.New("config is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns an error if config server session token secret is missing", func() {
			config.ExternalConfig.ServerSessionTokenSecret = ""
			client, err := authClient.NewClient(config, authorizeAs, name, lgr)
			errorsTest.ExpectEqual(err, errors.New("config is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns success", func() {
			client, err := authClient.NewClient(config, authorizeAs, name, lgr)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
			client.Close()
		})
	})

	Context("with started server and new client", func() {
		var server *Server
		var config *authClient.Config
		var authorizeAs platform.AuthorizeAs
		var client *authClient.Client

		BeforeEach(func() {
			server = NewServer()
			config = authClient.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Config).ToNot(BeNil())
			config.Config.Address = server.URL()
			config.Config.UserAgent = testHttp.NewUserAgent()
			config.Config.ServiceSecret = authTest.NewServiceSecret()
			config.ExternalConfig.Address = server.URL()
			config.ExternalConfig.UserAgent = testHttp.NewUserAgent()
			config.ExternalConfig.ServerSessionTokenSecret = serverSessionTokenSecret
			authorizeAs = platform.AuthorizeAsService
		})

		JustBeforeEach(func() {
			var err error
			client, err = authClient.NewClient(config, authorizeAs, name, lgr)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
		})

		AfterEach(func() {
			client.Close()
			if server != nil {
				server.Close()
			}
		})

		Context("Start", func() {
			Context("with immediate success of server login", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverSessionToken}})),
					)
				})

				It("returns nil and only invokes server login once", func() {
					Expect(client.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return server.ReceivedRequests()
					}, 10, 1).Should(HaveLen(1))
				})
			})

			Context("with one failure and then success of server login (delay 1 second)", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusBadRequest, nil)),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverSessionToken}})),
					)
				})

				It("returns nil and only invokes server login twice", func() {
					Expect(client.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return server.ReceivedRequests()
					}, 10, 1).Should(HaveLen(2))
				})
			})

			Context("with two failures and then success of server login (delay 1 second, then 2 seconds)", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusBadRequest, nil)),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusBadRequest, nil)),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverSessionToken}})),
					)
				})

				It("returns nil and only invokes server login thrice", func() {
					Expect(client.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return server.ReceivedRequests()
					}, 10, 1).Should(HaveLen(3))
				})
			})

			Context("with one missing session header and then success of server login (delay 1 second)", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil)),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverSessionToken}})),
					)
				})

				It("returns nil and only invokes server login twice", func() {
					Expect(client.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return server.ReceivedRequests()
					}, 10, 1).Should(HaveLen(2))
				})
			})

			Context("with 1 second token timeout", func() {
				BeforeEach(func() {
					config.ServerSessionTokenTimeout = 1 * time.Second
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverSessionToken}})),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverSessionToken}})),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverSessionToken}})),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverSessionToken}})),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverSessionToken}})),
					)
				})

				It("returns nil and invokes server login three or more times (depending upon exact timing)", func() {
					Expect(client.Start()).To(Succeed())
					Eventually(func() int {
						return len(server.ReceivedRequests())
					}, 10, 1).Should(BeNumerically(">=", 3))
				})
			})

			It("returns nil and even if server is unreachable", func() {
				server.Close()
				server = nil
				Expect(client.Start()).To(Succeed())
			})
		})

		Context("with client started and obtained a server token", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest("POST", "/auth/serverlogin"),
						VerifyHeaderKV("X-Tidepool-Server-Name", name),
						VerifyHeaderKV("X-Tidepool-Server-Secret", serverSessionTokenSecret),
						VerifyBody(nil),
						RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverSessionToken}})),
				)
			})

			JustBeforeEach(func() {
				Expect(client.Start()).To(Succeed())
			})

			Context("ServerSessionToken", func() {
				It("returns a server token", func() {
					returnedServerSessionToken, err := client.ServerSessionToken()
					Expect(err).ToNot(HaveOccurred())
					Expect(returnedServerSessionToken).To(Equal(serverSessionToken))
				})

				It("returns error if client is closed", func() {
					client.Close()
					returnedServerSessionToken, err := client.ServerSessionToken()
					errorsTest.ExpectEqual(err, errors.New("client is closed"))
					Expect(returnedServerSessionToken).To(BeEmpty())
				})
			})

			Context("ValidateSessionToken", func() {
				var sessionToken string

				BeforeEach(func() {
					sessionToken = authTest.NewSessionToken()
				})

				It("returns error if context is missing", func() {
					details, err := client.ValidateSessionToken(context.Context(nil), sessionToken)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(details).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if session token is missing", func() {
					details, err := client.ValidateSessionToken(ctx, "")
					errorsTest.ExpectEqual(err, errors.New("token is missing"))
					Expect(details).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if the server is not reachable", func() {
					server.Close()
					server = nil
					details, err := client.ValidateSessionToken(ctx, sessionToken)
					Expect(err).To(HaveOccurred())
					Expect(details).To(BeNil())
					Expect(err.Error()).To(HavePrefix("unable to perform request to GET "))
				})

				Context("with an unexpected response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+sessionToken),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverSessionToken),
								VerifyBody(nil),
								RespondWith(http.StatusBadRequest, nil)),
						)
					})

					It("returns an error", func() {
						details, err := client.ValidateSessionToken(ctx, sessionToken)
						Expect(err).To(HaveOccurred())
						Expect(details).To(BeNil())
						errorsTest.ExpectEqual(err, request.ErrorBadRequest())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an unauthorized response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+sessionToken),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverSessionToken),
								VerifyBody(nil),
								RespondWith(http.StatusUnauthorized, nil)),
						)
					})

					It("returns an error", func() {
						details, err := client.ValidateSessionToken(ctx, sessionToken)
						errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
						Expect(details).To(BeNil())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response, but not parsable", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+sessionToken),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverSessionToken),
								VerifyBody(nil),
								RespondWith(http.StatusOK, "}{")),
						)
					})

					It("returns an error", func() {
						details, err := client.ValidateSessionToken(ctx, sessionToken)
						Expect(err).To(HaveOccurred())
						Expect(details).To(BeNil())
						errorsTest.ExpectEqual(err, request.ErrorJSONMalformed())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response, but is not a server and missing the user id", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+sessionToken),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverSessionToken),
								VerifyBody(nil),
								RespondWith(http.StatusOK, "{}")),
						)
					})

					It("returns an error", func() {
						details, err := client.ValidateSessionToken(ctx, sessionToken)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(details).To(BeNil())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response and a user id", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+sessionToken),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverSessionToken),
								VerifyBody(nil),
								RespondWith(http.StatusOK, `{"userid": "session-user-id"}`)),
						)
					})

					It("returns the user id", func() {
						details, err := client.ValidateSessionToken(ctx, sessionToken)
						Expect(details).ToNot(BeNil())
						Expect(err).ToNot(HaveOccurred())
						Expect(details.Token()).To(Equal(sessionToken))
						Expect(details.IsService()).To(BeFalse())
						Expect(details.UserID()).To(Equal("session-user-id"))
					})
				})

				Context("with a successful response and is server", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+sessionToken),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverSessionToken),
								VerifyBody(nil),
								RespondWith(http.StatusOK, "{\"isserver\": true}")),
						)
					})

					It("returns is server", func() {
						details, err := client.ValidateSessionToken(ctx, sessionToken)
						Expect(details).ToNot(BeNil())
						Expect(err).ToNot(HaveOccurred())
						Expect(details.Token()).To(Equal(sessionToken))
						Expect(details.IsService()).To(BeTrue())
						Expect(details.UserID()).To(BeEmpty())
					})
				})
			})
		})

		Context("RefreshProviderSession", func() {
			var id string
			var refresh *auth.ProviderSessionRefresh

			BeforeEach(func() {
				id = authTest.RandomProviderSessionID()
				refresh = authTest.RandomProviderSessionRefresh(test.AllowOptionals())
			})

			Context("without server response", func() {
				AfterEach(func() {
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns an error if the context is missing", func() {
					prvdrSssn, err := client.RefreshProviderSession(context.Context(nil), id, refresh)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(prvdrSssn).To(BeNil())
				})

				It("returns an error if the id is missing", func() {
					prvdrSssn, err := client.RefreshProviderSession(ctx, "", refresh)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(prvdrSssn).To(BeNil())
				})

				It("returns an error if the refresh is invalid", func() {
					refresh.TimeRange = &times.TimeRange{
						From: pointer.From(time.Time{}),
					}
					prvdrSssn, err := client.RefreshProviderSession(ctx, id, refresh)
					errorsTest.ExpectEqual(err, errors.New("refresh is invalid"))
					Expect(prvdrSssn).To(BeNil())
				})
			})

			Context("with server response", func() {
				var requestHandlers []http.HandlerFunc
				var responseHeaders http.Header

				BeforeEach(func() {
					requestHandlers = nil
					responseHeaders = http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
				})

				JustBeforeEach(func() {
					requestHandlers = append(requestHandlers,
						VerifyRequest("POST", fmt.Sprintf("/v1/provider_sessions/%s/refresh", id)),
						VerifyContentType("application/json; charset=utf-8"),
						VerifyBody(test.MarshalRequestBody(pointer.Default(refresh, auth.ProviderSessionRefresh{}))),
					)
					server.AppendHandlers(CombineHandlers(requestHandlers...))
				})

				AfterEach(func() {
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				When("the server responds with an unauthenticated error", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
					})

					It("returns an error", func() {
						result, err := client.RefreshProviderSession(ctx, id, refresh)
						errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
						Expect(result).To(BeNil())
					})
				})

				When("the server responds with an unauthorized error", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
					})

					It("returns an error", func() {
						result, err := client.RefreshProviderSession(ctx, id, refresh)
						errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
						Expect(result).To(BeNil())
					})
				})

				When("the server responds with a not found error", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(id)), responseHeaders))
					})

					It("returns an error", func() {
						result, err := client.RefreshProviderSession(ctx, id, refresh)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(BeNil())
					})
				})

				When("the server responds with no result", func() {
					var prvdrSssn *auth.ProviderSession

					BeforeEach(func() {
						prvdrSssn = authTest.RandomProviderSession(test.AllowOptionals())
						prvdrSssn.ID = id
						requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, prvdrSssn, responseHeaders))
					})

					It("returns successfully", func() {
						result, err := client.RefreshProviderSession(ctx, id, refresh)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(Equal(prvdrSssn))
					})
				})

				When("the server responds with result", func() {
					var prvdrSssn *auth.ProviderSession

					BeforeEach(func() {
						prvdrSssn = authTest.RandomProviderSession(test.AllowOptionals())
						prvdrSssn.ID = id
						requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, prvdrSssn, responseHeaders))
					})

					It("returns successfully", func() {
						Expect(client.RefreshProviderSession(ctx, id, refresh)).To(Equal(prvdrSssn))
					})

					When("refresh is nil", func() {
						BeforeEach(func() {
							refresh = nil
						})

						It("returns successfully", func() {
							Expect(client.RefreshProviderSession(ctx, id, refresh)).To(Equal(prvdrSssn))
						})
					})
				})
			})
		})
	})
})
