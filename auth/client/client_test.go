package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/golang/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	authClient "github.com/tidepool-org/platform/auth/client"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Client", func() {
	var serverSessionTokenProviderController *gomock.Controller
	var serverSessionTokenProvider *authTest.MockServerSessionTokenProvider
	var serverTokenSecret string
	var serverTokenTimeout int
	var name string
	var logger log.Logger
	var serverToken string
	var token string
	var ctx context.Context

	BeforeEach(func() {
		serverSessionTokenProviderController = gomock.NewController(GinkgoT())
		serverSessionTokenProvider = authTest.NewMockServerSessionTokenProvider(serverSessionTokenProviderController)
		serverTokenSecret = authTest.NewServiceSecret()
		serverTokenTimeout = testHttp.NewTimeout()
		name = test.RandomStringFromRangeAndCharset(4, 16, test.CharsetAlphaNumeric)
		logger = logTest.NewLogger()
		Expect(logger).ToNot(BeNil())
		serverToken = authTest.NewSessionToken()
		serverSessionTokenProvider.EXPECT().ServerSessionToken().Return(serverToken, nil).AnyTimes()
		token = authTest.NewSessionToken()
		ctx = auth.NewContextWithServerSessionTokenProvider(log.NewContextWithLogger(context.Background(), logTest.NewLogger()), serverSessionTokenProvider)
	})

	AfterEach(func() {
		serverSessionTokenProviderController.Finish()
	})

	Context("NewClient", func() {
		var config *authClient.Config
		var authorizeAs platform.AuthorizeAs

		BeforeEach(func() {
			config = authClient.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Config).ToNot(BeNil())
			config.Config.Address = testHttp.NewAddress()
			config.Config.UserAgent = testHttp.NewUserAgent()
			config.Config.ServiceSecret = authTest.NewServiceSecret()
			config.ExternalConfig.Address = testHttp.NewAddress()
			config.ExternalConfig.UserAgent = testHttp.NewUserAgent()
			config.ExternalConfig.ServerSessionTokenSecret = serverTokenSecret
			config.ExternalConfig.ServerSessionTokenTimeout = time.Duration(serverTokenTimeout) * time.Second
			authorizeAs = platform.AuthorizeAsService
		})

		It("returns an error if config is missing", func() {
			client, err := authClient.NewClient(nil, authorizeAs, name, logger)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error if name is missing", func() {
			client, err := authClient.NewClient(config, authorizeAs, "", logger)
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
			client, err := authClient.NewClient(config, authorizeAs, name, logger)
			errorsTest.ExpectEqual(err, errors.New("config is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns an error if config server session token secret is missing", func() {
			config.ExternalConfig.ServerSessionTokenSecret = ""
			client, err := authClient.NewClient(config, authorizeAs, name, logger)
			errorsTest.ExpectEqual(err, errors.New("config is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns success", func() {
			client, err := authClient.NewClient(config, authorizeAs, name, logger)
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
			config.ExternalConfig.ServerSessionTokenSecret = serverTokenSecret
			authorizeAs = platform.AuthorizeAsService
		})

		JustBeforeEach(func() {
			var err error
			client, err = authClient.NewClient(config, authorizeAs, name, logger)
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
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
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
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusBadRequest, nil)),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
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
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusBadRequest, nil)),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusBadRequest, nil)),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
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
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil)),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
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
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody(nil),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
					)
				})

				It("returns nil and only invokes server login thrice", func() {
					Expect(client.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return server.ReceivedRequests()
					}, 10, 1).Should(HaveLen(3))
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
						VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
						VerifyBody(nil),
						RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
				)
			})

			JustBeforeEach(func() {
				Expect(client.Start()).To(Succeed())
			})

			Context("ServerSessionToken", func() {
				It("returns a server token", func() {
					returnedServerSessionToken, err := client.ServerSessionToken()
					Expect(err).ToNot(HaveOccurred())
					Expect(returnedServerSessionToken).To(Equal(serverToken))
				})

				It("returns error if client is closed", func() {
					client.Close()
					returnedServerSessionToken, err := client.ServerSessionToken()
					errorsTest.ExpectEqual(err, errors.New("client is closed"))
					Expect(returnedServerSessionToken).To(BeEmpty())
				})
			})

			Context("ValidateSessionToken", func() {
				It("returns error if context is missing", func() {
					details, err := client.ValidateSessionToken(nil, token)
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
					details, err := client.ValidateSessionToken(ctx, token)
					Expect(err).To(HaveOccurred())
					Expect(details).To(BeNil())
					Expect(err.Error()).To(HavePrefix("unable to perform request to GET "))
				})

				Context("with an unexpected response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+token),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody(nil),
								RespondWith(http.StatusBadRequest, nil)),
						)
					})

					It("returns an error", func() {
						details, err := client.ValidateSessionToken(ctx, token)
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
								VerifyRequest("GET", "/auth/token/"+token),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody(nil),
								RespondWith(http.StatusUnauthorized, nil)),
						)
					})

					It("returns an error", func() {
						details, err := client.ValidateSessionToken(ctx, token)
						errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
						Expect(details).To(BeNil())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response, but not parseable", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+token),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody(nil),
								RespondWith(http.StatusOK, "}{")),
						)
					})

					It("returns an error", func() {
						details, err := client.ValidateSessionToken(ctx, token)
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
								VerifyRequest("GET", "/auth/token/"+token),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody(nil),
								RespondWith(http.StatusOK, "{}")),
						)
					})

					It("returns an error", func() {
						details, err := client.ValidateSessionToken(ctx, token)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(details).To(BeNil())
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response and a user id", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+token),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody(nil),
								RespondWith(http.StatusOK, `{"userid": "session-user-id"}`)),
						)
					})

					It("returns the user id", func() {
						details, err := client.ValidateSessionToken(ctx, token)
						Expect(details).ToNot(BeNil())
						Expect(err).ToNot(HaveOccurred())
						Expect(details.Token()).To(Equal(token))
						Expect(details.IsService()).To(BeFalse())
						Expect(details.UserID()).To(Equal("session-user-id"))
					})
				})

				Context("with a successful response and is server", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+token),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody(nil),
								RespondWith(http.StatusOK, "{\"isserver\": true}")),
						)
					})

					It("returns is server", func() {
						details, err := client.ValidateSessionToken(ctx, token)
						Expect(details).ToNot(BeNil())
						Expect(err).ToNot(HaveOccurred())
						Expect(details.Token()).To(Equal(token))
						Expect(details.IsService()).To(BeTrue())
						Expect(details.UserID()).To(BeEmpty())
					})
				})
			})

			Describe("GetDeviceTokens", func() {
				var testUserID = "test-user-id"
				var testUserIDBadResponse = "test-user-id-bad-response"
				var testTokens = map[string]any{
					testUserID: []*devicetokens.DeviceToken{{
						Apple: &devicetokens.AppleDeviceToken{
							Token:       []byte("blah"),
							Environment: "sandbox",
						},
					}},
					testUserIDBadResponse: []map[string]any{
						{
							"Apple": "",
						},
					},
				}

				It("returns a token", func() {
					body, err := json.Marshal(testTokens[testUserID])
					Expect(err).To(Succeed())
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("GET", "/v1/users/"+testUserID+"/device_tokens"),
							RespondWith(http.StatusOK, body)),
					)

					tokens, err := client.GetDeviceTokens(ctx, testUserID)
					Expect(err).To(Succeed())
					Expect(tokens).To(HaveLen(1))
					Expect([]byte(tokens[0].Apple.Token)).To(Equal([]byte("blah")))
					Expect(tokens[0].Apple.Environment).To(Equal("sandbox"))
				})

				It("returns an error when receiving malformed responses", func() {
					body, err := json.Marshal(testTokens[testUserIDBadResponse])
					Expect(err).To(Succeed())
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("GET", "/v1/users/"+testUserIDBadResponse+"/device_tokens"),
							RespondWith(http.StatusOK, body)),
					)

					_, err = client.GetDeviceTokens(ctx, testUserIDBadResponse)
					Expect(err).To(HaveOccurred())
				})

				It("returns an error on non-200 responses", func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest("GET", "/v1/users/"+testUserID+"/device_tokens"),
							RespondWith(http.StatusBadRequest, nil)),
					)
					_, err := client.GetDeviceTokens(ctx, testUserID)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(ContainSubstring("Unable to request device token data")))
				})
			})
		})
	})
})
