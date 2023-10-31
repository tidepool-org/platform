package platform_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Client", func() {
	Context("AuthorizeAsService", func() {
		It("returns the expected value", func() {
			Expect(platform.AuthorizeAsService).To(Equal(platform.AuthorizeAs(0)))
		})
	})

	Context("AuthorizeAsUser", func() {
		It("returns the expected value", func() {
			Expect(platform.AuthorizeAsUser).To(Equal(platform.AuthorizeAs(1)))
		})
	})

	Context("with config", func() {
		var address string
		var userAgent string
		var serviceSecret string
		var ctx context.Context
		var config *platform.Config

		BeforeEach(func() {
			address = testHttp.NewAddress()
			userAgent = testHttp.NewUserAgent()
			serviceSecret = authTest.NewServiceSecret()
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		})

		JustBeforeEach(func() {
			config = platform.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Config).ToNot(BeNil())
			config.Address = address
			config.UserAgent = userAgent
			config.ServiceSecret = serviceSecret
		})

		Context("NewClient", func() {
			It("returns an error if config is missing", func() {
				clnt, err := platform.NewClient(nil, platform.AuthorizeAsUser)
				Expect(err).To(MatchError("config is missing"))
				Expect(clnt).To(BeNil())
			})

			It("returns an error if config is invalid", func() {
				config.Address = ""
				clnt, err := platform.NewClient(config, platform.AuthorizeAsUser)
				Expect(err).To(MatchError("config is invalid; address is missing"))
				Expect(clnt).To(BeNil())
			})

			It("returns an error if authorize as is invalid", func() {
				clnt, err := platform.NewClient(config, platform.AuthorizeAs(-1))
				Expect(err).To(MatchError("authorize as is invalid"))
				Expect(clnt).To(BeNil())
			})

			It("returns success", func() {
				clnt, err := platform.NewClient(config, platform.AuthorizeAsUser)
				Expect(err).ToNot(HaveOccurred())
				Expect(clnt).ToNot(BeNil())
			})
		})

		Context("with new client authorize as service", func() {
			var clnt *platform.Client

			JustBeforeEach(func() {
				var err error
				clnt, err = platform.NewClient(config, platform.AuthorizeAsService)
				Expect(err).ToNot(HaveOccurred())
				Expect(clnt).ToNot(BeNil())
			})

			Context("IsAuthorizeAsService", func() {
				It("returns true", func() {
					Expect(clnt.IsAuthorizeAsService()).To(BeTrue())
				})
			})

			Context("Mutators", func() {
				It("returns an error if context is missing", func() {
					mutators, err := clnt.Mutators(nil)
					Expect(err).To(MatchError("context is missing"))
					Expect(mutators).To(BeNil())
				})

				It("returns the expected mutators", func() {
					mutators, err := clnt.Mutators(ctx)
					Expect(err).ToNot(HaveOccurred())
					Expect(mutators).To(ConsistOf(
						platform.NewServiceSecretHeaderMutator(serviceSecret),
						platform.NewTraceMutator(ctx),
					))
				})

				Context("without service secret", func() {
					BeforeEach(func() {
						serviceSecret = ""
					})

					It("returns an error", func() {
						mutators, err := clnt.Mutators(ctx)
						Expect(err).To(MatchError("service secret is missing"))
						Expect(mutators).To(BeNil())
					})
				})
			})

			Context("HTTPClient", func() {
				It("returns successfully", func() {
					Expect(clnt.HTTPClient()).ToNot(BeNil())
				})
			})

			Context("with started server and new client", func() {
				var server *Server
				var method string
				var path string
				var url string

				BeforeEach(func() {
					server = NewServer()
					address = server.URL()
					method = testHttp.NewMethod()
					path = testHttp.NewPath()
					url = server.URL() + path
				})

				AfterEach(func() {
					if server != nil {
						server.Close()
					}
				})

				Context("RequestData", func() {
					It("returns error if context is missing", func() {
						Expect(clnt.RequestData(nil, method, url, nil, nil)).To(MatchError("context is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error if method is missing", func() {
						Expect(clnt.RequestData(ctx, "", url, nil, nil)).To(MatchError("method is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error if url is missing", func() {
						Expect(clnt.RequestData(ctx, method, "", nil, nil)).To(MatchError("url is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					Context("with a successful response 200", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV(auth.TidepoolServiceSecretHeaderKey, serviceSecret),
									VerifyBody(nil),
									RespondWith(http.StatusOK, nil)),
							)
						})

						It("returns success", func() {
							Expect(clnt.RequestData(ctx, method, url, nil, nil)).To(Succeed())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a successful response 200 with additional inspectors", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV(auth.TidepoolServiceSecretHeaderKey, serviceSecret),
									VerifyBody(nil),
									RespondWith(http.StatusOK, nil)),
							)
						})

						It("returns success", func() {
							inspector := request.NewHeadersInspector()
							Expect(clnt.RequestData(ctx, method, url, nil, nil, inspector)).To(Succeed())

							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})
				})
			})
		})

		Context("with new client authorize as user", func() {
			var sessionToken string
			var clnt *platform.Client

			BeforeEach(func() {
				serviceSecret = ""
				sessionToken = authTest.NewSessionToken()
				ctx = request.NewContextWithDetails(ctx, request.NewDetails(
					request.MethodSessionToken,
					test.RandomStringFromRangeAndCharset(10, 10, test.CharsetAlphaNumeric),
					sessionToken, "patient"),
				)
			})

			JustBeforeEach(func() {
				var err error
				clnt, err = platform.NewClient(config, platform.AuthorizeAsUser)
				Expect(err).ToNot(HaveOccurred())
				Expect(clnt).ToNot(BeNil())
			})

			Context("IsAuthorizeAsService", func() {
				It("returns false", func() {
					Expect(clnt.IsAuthorizeAsService()).To(BeFalse())
				})
			})

			Context("Mutators", func() {
				It("returns an error if context is missing", func() {
					mutators, err := clnt.Mutators(nil)
					Expect(err).To(MatchError("context is missing"))
					Expect(mutators).To(BeNil())
				})

				It("returns an error if details are not in context", func() {
					mutators, err := clnt.Mutators(request.NewContextWithDetails(ctx, nil))
					Expect(err).To(MatchError("details is missing"))
					Expect(mutators).To(BeNil())
				})

				It("returns the expected mutators", func() {
					mutators, err := clnt.Mutators(ctx)
					Expect(err).ToNot(HaveOccurred())
					Expect(mutators).To(ConsistOf(
						platform.NewSessionTokenHeaderMutator(sessionToken),
						platform.NewTraceMutator(ctx),
					))
				})
			})

			Context("HTTPClient", func() {
				It("returns successfully", func() {
					Expect(clnt.HTTPClient()).ToNot(BeNil())
				})
			})

			Context("with started server and new client", func() {
				var server *Server
				var method string
				var path string
				var url string

				BeforeEach(func() {
					server = NewServer()
					address = server.URL()
					method = testHttp.NewMethod()
					path = testHttp.NewPath()
					url = server.URL() + path
				})

				AfterEach(func() {
					if server != nil {
						server.Close()
					}
				})

				Context("RequestData", func() {
					It("returns error if context is missing", func() {
						Expect(clnt.RequestData(nil, method, url, nil, nil)).To(MatchError("context is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error if method is missing", func() {
						Expect(clnt.RequestData(ctx, "", url, nil, nil)).To(MatchError("method is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error if url is missing", func() {
						Expect(clnt.RequestData(ctx, method, "", nil, nil)).To(MatchError("url is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					Context("with a successful response 200", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV(auth.TidepoolSessionTokenHeaderKey, sessionToken),
									VerifyBody(nil),
									RespondWith(http.StatusOK, nil)),
							)
						})

						It("returns success", func() {
							Expect(clnt.RequestData(ctx, method, url, nil, nil)).To(Succeed())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a successful response 200 with additional inspector", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV(auth.TidepoolSessionTokenHeaderKey, sessionToken),
									VerifyBody(nil),
									RespondWith(http.StatusOK, nil)),
							)
						})

						It("returns success", func() {
							inspector := request.NewHeadersInspector()
							Expect(clnt.RequestData(ctx, method, url, nil, nil, inspector)).To(Succeed())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})
				})
			})
		})
	})
})
