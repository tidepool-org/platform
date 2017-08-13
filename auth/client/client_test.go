package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"net/http"
	"time"

	"github.com/tidepool-org/platform/auth/client"
	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	nullLog "github.com/tidepool-org/platform/log/null"
)

var _ = Describe("Client", func() {
	var name string
	var logger log.Logger
	var ctx *testAuth.Context
	var serverTokenSecret string
	var serverToken string
	var token string

	BeforeEach(func() {
		name = id.New()
		logger = nullLog.NewLogger()
		Expect(logger).ToNot(BeNil())
		ctx = testAuth.NewContext()
		Expect(ctx).ToNot(BeNil())
		serverTokenSecret = id.New()
		serverToken = id.New()
		token = id.New()
	})

	Context("NewClient", func() {
		var config *client.Config

		BeforeEach(func() {
			config = client.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Config).ToNot(BeNil())
			config.Config.Address = "http://localhost:1234"
			config.Config.Timeout = 30 * time.Second
			config.ServerTokenSecret = serverTokenSecret
			config.ServerTokenTimeout = 1800 * time.Second
		})

		It("returns an error if config is missing", func() {
			clnt, err := client.NewClient(nil, name, logger)
			Expect(err).To(MatchError("client: config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if name is missing", func() {
			clnt, err := client.NewClient(config, "", logger)
			Expect(err).To(MatchError("client: name is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if logger is missing", func() {
			clnt, err := client.NewClient(config, name, nil)
			Expect(err).To(MatchError("client: logger is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config address is missing", func() {
			config.Address = ""
			clnt, err := client.NewClient(config, name, logger)
			Expect(err).To(MatchError("client: config is invalid; client: address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config server token secret is missing", func() {
			config.ServerTokenSecret = ""
			clnt, err := client.NewClient(config, name, logger)
			Expect(err).To(MatchError("client: config is invalid; client: server token secret is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			clnt, err := client.NewClient(config, name, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
			clnt.Close()
		})
	})

	Context("with started server and new client", func() {
		var svr *Server
		var config *client.Config
		var clnt *client.Client

		BeforeEach(func() {
			svr = NewServer()
			config = client.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Config).ToNot(BeNil())
			config.Config.Address = svr.URL()
			config.Config.Timeout = 30 * time.Second
			config.ServerTokenSecret = serverTokenSecret
			config.ServerTokenTimeout = 1800 * time.Second
		})

		JustBeforeEach(func() {
			var err error
			clnt, err = client.NewClient(config, name, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
			ctx.AuthClientMock = clnt
		})

		AfterEach(func() {
			clnt.Close()
			if svr != nil {
				svr.Close()
			}
		})

		Context("Start", func() {
			Context("with immediate success of server login", func() {
				BeforeEach(func() {
					svr.AppendHandlers(
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody([]byte{}),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
					)
				})

				It("returns nil and only invokes server login once", func() {
					Expect(clnt.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return svr.ReceivedRequests()
					}, 10, 1).Should(HaveLen(1))
				})
			})

			Context("with one failure and then success of server login (delay 1 second)", func() {
				BeforeEach(func() {
					svr.AppendHandlers(
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody([]byte{}),
							RespondWith(http.StatusBadRequest, nil, nil)),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody([]byte{}),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
					)
				})

				It("returns nil and only invokes server login twice", func() {
					Expect(clnt.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return svr.ReceivedRequests()
					}, 10, 1).Should(HaveLen(2))

				})
			})

			Context("with two failures and then success of server login (delay 1 second, then 2 seconds)", func() {
				BeforeEach(func() {
					svr.AppendHandlers(
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody([]byte{}),
							RespondWith(http.StatusBadRequest, nil, nil)),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody([]byte{}),
							RespondWith(http.StatusBadRequest, nil, nil)),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody([]byte{}),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
					)
				})

				It("returns nil and only invokes server login thrice", func() {
					Expect(clnt.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return svr.ReceivedRequests()
					}, 10, 1).Should(HaveLen(3))
				})
			})

			Context("with one missing session header and then success of server login (delay 1 second)", func() {
				BeforeEach(func() {
					svr.AppendHandlers(
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody([]byte{}),
							RespondWith(http.StatusOK, nil, nil)),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody([]byte{}),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
					)
				})

				It("returns nil and only invokes server login twice", func() {
					Expect(clnt.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return svr.ReceivedRequests()
					}, 10, 1).Should(HaveLen(2))
				})
			})

			Context("with 1 second token timeout", func() {
				BeforeEach(func() {
					config.ServerTokenTimeout = 1 * time.Second
					svr.AppendHandlers(
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody([]byte{}),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody([]byte{}),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody([]byte{}),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
						CombineHandlers(
							VerifyRequest("POST", "/auth/serverlogin"),
							VerifyHeaderKV("X-Tidepool-Server-Name", name),
							VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
							VerifyBody([]byte{}),
							RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
					)
				})

				It("returns nil and only invokes server login thrice", func() {
					Expect(clnt.Start()).To(Succeed())
					Eventually(func() []*http.Request {
						return svr.ReceivedRequests()
					}, 10, 1).Should(HaveLen(3))
				})
			})

			It("returns nil and even if server is unreachable", func() {
				svr.Close()
				svr = nil
				Expect(clnt.Start()).To(Succeed())
			})
		})

		Context("with client started and obtained a server token", func() {
			BeforeEach(func() {
				svr.AppendHandlers(
					CombineHandlers(
						VerifyRequest("POST", "/auth/serverlogin"),
						VerifyHeaderKV("X-Tidepool-Server-Name", name),
						VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
						VerifyBody([]byte{}),
						RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{serverToken}})),
				)
			})

			JustBeforeEach(func() {
				Expect(clnt.Start()).To(Succeed())
			})

			Context("ServerToken", func() {
				It("returns a server token", func() {
					returnedServerToken, err := clnt.ServerToken()
					Expect(err).ToNot(HaveOccurred())
					Expect(returnedServerToken).To(Equal(serverToken))
				})

				It("returns error if client is closed", func() {
					clnt.Close()
					returnedServerToken, err := clnt.ServerToken()
					Expect(err).To(MatchError("client: client is closed"))
					Expect(returnedServerToken).To(BeEmpty())
				})
			})

			Context("ValidateToken", func() {
				It("returns error if context is missing", func() {
					details, err := clnt.ValidateToken(nil, token)
					Expect(err).To(MatchError("client: context is missing"))
					Expect(details).To(BeNil())
					Expect(svr.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if session token is missing", func() {
					details, err := clnt.ValidateToken(ctx, "")
					Expect(err).To(MatchError("client: token is missing"))
					Expect(details).To(BeNil())
					Expect(svr.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if client is closed", func() {
					clnt.Close()
					details, err := clnt.ValidateToken(ctx, token)
					Expect(err).To(MatchError("client: client is closed"))
					Expect(details).To(BeNil())
					Expect(svr.ReceivedRequests()).To(HaveLen(1))
				})

				It("returns error if the server is not reachable", func() {
					svr.Close()
					svr = nil
					details, err := clnt.ValidateToken(ctx, token)
					Expect(err).To(HaveOccurred())
					Expect(details).To(BeNil())
					Expect(err.Error()).To(HavePrefix("client: unable to perform request GET "))
				})

				It("returns error if the context request is missing", func() {
					ctx.RequestImpl = nil
					details, err := clnt.ValidateToken(ctx, token)
					Expect(err).To(MatchError("client: unable to copy request trace; service: source request is missing"))
					Expect(details).To(BeNil())
					Expect(svr.ReceivedRequests()).To(HaveLen(1))
				})

				Context("with an unexpected response", func() {
					BeforeEach(func() {
						svr.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+token),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusBadRequest, nil, nil)),
						)
					})

					It("returns an error", func() {
						details, err := clnt.ValidateToken(ctx, token)
						Expect(err).To(HaveOccurred())
						Expect(details).To(BeNil())
						Expect(err.Error()).To(HavePrefix("client: unexpected response status code 400 from GET "))
						Expect(svr.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with an unauthorized response", func() {
					BeforeEach(func() {
						svr.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+token),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusUnauthorized, nil, nil)),
						)
					})

					It("returns an error", func() {
						details, err := clnt.ValidateToken(ctx, token)
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(details).To(BeNil())
						Expect(svr.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response, but not parseable", func() {
					BeforeEach(func() {
						svr.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+token),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusOK, "}{", nil)),
						)
					})

					It("returns an error", func() {
						details, err := clnt.ValidateToken(ctx, token)
						Expect(err).To(HaveOccurred())
						Expect(details).To(BeNil())
						Expect(err.Error()).To(HavePrefix("client: error decoding JSON response from GET "))
						Expect(svr.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response, but is not a server and missing the user id", func() {
					BeforeEach(func() {
						svr.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+token),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusOK, "{}", nil)),
						)
					})

					It("returns an error", func() {
						details, err := clnt.ValidateToken(ctx, token)
						Expect(err).To(MatchError("client: user id is missing"))
						Expect(details).To(BeNil())
						Expect(svr.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response and a user id", func() {
					BeforeEach(func() {
						svr.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+token),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusOK, `{"userid": "session-user-id"}`, nil)),
						)
					})

					It("returns the user id", func() {
						details, err := clnt.ValidateToken(ctx, token)
						Expect(details).ToNot(BeNil())
						Expect(err).ToNot(HaveOccurred())
						Expect(details.Token()).To(Equal(token))
						Expect(details.IsServer()).To(BeFalse())
						Expect(details.UserID()).To(Equal("session-user-id"))
					})
				})

				Context("with a successful response and is server", func() {
					BeforeEach(func() {
						svr.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/auth/token/"+token),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusOK, "{\"isserver\": true}", nil)),
						)
					})

					It("returns is server", func() {
						details, err := clnt.ValidateToken(ctx, token)
						Expect(details).ToNot(BeNil())
						Expect(err).ToNot(HaveOccurred())
						Expect(details.Token()).To(Equal(token))
						Expect(details.IsServer()).To(BeTrue())
						Expect(details.UserID()).To(Equal(""))
					})
				})
			})

			Context("GetStatus", func() {
				It("returns an error if unsuccessful", func() {
					sts, err := clnt.GetStatus(nil)
					Expect(err).To(HaveOccurred())
					Expect(sts).To(BeNil())
					Expect(svr.ReceivedRequests()).To(HaveLen(1))
				})

				Context("with an empty body", func() {
					BeforeEach(func() {
						svr.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/status"),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusOK, nil, nil)),
						)
					})

					It("returns an error", func() {
						sts, err := clnt.GetStatus(ctx)
						Expect(err.Error()).To(HavePrefix("client: error decoding JSON response from GET "))
						Expect(sts).To(BeNil())
						Expect(svr.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful, but empty response", func() {
					BeforeEach(func() {
						svr.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/status"),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusOK, `{}`, nil)),
						)
					})

					It("returns successfully", func() {
						sts, err := clnt.GetStatus(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(sts).ToNot(BeNil())
						Expect(sts.Version).To(BeEmpty())
						Expect(svr.ReceivedRequests()).To(HaveLen(2))
					})
				})

				Context("with a successful response", func() {
					BeforeEach(func() {
						svr.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/status"),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusOK, `{"version": "1.2.3"}`, nil)),
						)
					})

					It("returns successfully", func() {
						sts, err := clnt.GetStatus(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(sts).ToNot(BeNil())
						Expect(sts.Version).To(Equal("1.2.3"))
						Expect(svr.ReceivedRequests()).To(HaveLen(2))
					})
				})
			})
		})

		Context("with client started and did NOT obtain a server token", func() {
			BeforeEach(func() {
				svr.AppendHandlers(
					CombineHandlers(
						VerifyRequest("POST", "/auth/serverlogin"),
						VerifyHeaderKV("X-Tidepool-Server-Name", name),
						VerifyHeaderKV("X-Tidepool-Server-Secret", serverTokenSecret),
						VerifyBody([]byte{}),
						RespondWith(http.StatusBadRequest, nil, nil)),
				)
			})

			JustBeforeEach(func() {
				Expect(clnt.Start()).To(Succeed())
			})

			Context("ServerToken", func() {
				It("returns an error", func() {
					returnedServiceToken, err := clnt.ServerToken()
					Expect(err).To(MatchError("client: unable to obtain server token"))
					Expect(returnedServiceToken).To(BeEmpty())
				})
			})

			Context("ValidateToken", func() {
				It("returns an error", func() {
					details, err := clnt.ValidateToken(ctx, token)
					Expect(err).To(MatchError("client: unable to obtain server token"))
					Expect(details).To(BeNil())
					Expect(svr.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
