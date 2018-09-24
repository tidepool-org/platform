package platform_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"context"
	"io"
	"net/http"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHTTP "github.com/tidepool-org/platform/test/http"
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
			address = testHTTP.NewAddress()
			userAgent = testHTTP.NewUserAgent()
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

					Context("with server session token", func() {
						var sessionToken string

						BeforeEach(func() {
							sessionToken = authTest.NewSessionToken()
							ctx = auth.NewContextWithServerSessionToken(ctx, sessionToken)
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
					method = testHTTP.NewMethod()
					path = testHTTP.NewPath()
					url = server.URL() + path
				})

				AfterEach(func() {
					if server != nil {
						server.Close()
					}
				})

				Context("RequestStream", func() {
					var reader io.ReadCloser
					var err error

					AfterEach(func() {
						if reader != nil {
							reader.Close()
						}
					})

					It("returns error if context is missing", func() {
						reader, err = clnt.RequestStream(nil, method, url, nil, nil)
						Expect(err).To(MatchError("context is missing"))
						Expect(reader).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error if method is missing", func() {
						reader, err = clnt.RequestStream(ctx, "", url, nil, nil)
						Expect(err).To(MatchError("method is missing"))
						Expect(reader).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error if url is missing", func() {
						reader, err = clnt.RequestStream(ctx, method, "", nil, nil)
						Expect(err).To(MatchError("url is missing"))
						Expect(reader).To(BeNil())
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
							reader, err = clnt.RequestStream(ctx, method, url, nil, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(reader).ToNot(BeNil())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a successful response 200 with additional mutators and inspectors", func() {
						var headerKey string
						var headerValue string

						BeforeEach(func() {
							headerKey = testHTTP.NewHeaderKey()
							headerValue = testHTTP.NewHeaderValue()
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV(auth.TidepoolServiceSecretHeaderKey, serviceSecret),
									VerifyHeaderKV(headerKey, headerValue),
									VerifyBody(nil),
									RespondWith(http.StatusOK, nil)),
							)
						})

						It("returns success", func() {
							mutators := []request.RequestMutator{request.NewHeaderMutator(headerKey, headerValue)}
							inspector := request.NewHeadersInspector()
							reader, err = clnt.RequestStream(ctx, method, url, mutators, nil, inspector)
							Expect(err).ToNot(HaveOccurred())
							Expect(reader).ToNot(BeNil())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})
				})

				Context("RequestData", func() {
					It("returns error if context is missing", func() {
						Expect(clnt.RequestData(nil, method, url, nil, nil, nil)).To(MatchError("context is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error if method is missing", func() {
						Expect(clnt.RequestData(ctx, "", url, nil, nil, nil)).To(MatchError("method is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error if url is missing", func() {
						Expect(clnt.RequestData(ctx, method, "", nil, nil, nil)).To(MatchError("url is missing"))
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
							Expect(clnt.RequestData(ctx, method, url, nil, nil, nil)).To(Succeed())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a successful response 200 with additional mutators and inspectors", func() {
						var headerKey string
						var headerValue string

						BeforeEach(func() {
							headerKey = testHTTP.NewHeaderKey()
							headerValue = testHTTP.NewHeaderValue()
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV(auth.TidepoolServiceSecretHeaderKey, serviceSecret),
									VerifyHeaderKV(headerKey, headerValue),
									VerifyBody(nil),
									RespondWith(http.StatusOK, nil)),
							)
						})

						It("returns success", func() {
							mutators := []request.RequestMutator{request.NewHeaderMutator(headerKey, headerValue)}
							inspector := request.NewHeadersInspector()
							Expect(clnt.RequestData(ctx, method, url, mutators, nil, nil, inspector)).To(Succeed())
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
				ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, test.NewString(10, test.CharsetAlphaNumeric), sessionToken))
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
					method = testHTTP.NewMethod()
					path = testHTTP.NewPath()
					url = server.URL() + path
				})

				AfterEach(func() {
					if server != nil {
						server.Close()
					}
				})

				Context("RequestStream", func() {
					var reader io.ReadCloser
					var err error

					AfterEach(func() {
						if reader != nil {
							reader.Close()
						}
					})

					It("returns error if context is missing", func() {
						reader, err = clnt.RequestStream(nil, method, url, nil, nil)
						Expect(err).To(MatchError("context is missing"))
						Expect(reader).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error if method is missing", func() {
						reader, err = clnt.RequestStream(ctx, "", url, nil, nil)
						Expect(err).To(MatchError("method is missing"))
						Expect(reader).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error if url is missing", func() {
						reader, err = clnt.RequestStream(ctx, method, "", nil, nil)
						Expect(err).To(MatchError("url is missing"))
						Expect(reader).To(BeNil())
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
							reader, err = clnt.RequestStream(ctx, method, url, nil, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(reader).ToNot(BeNil())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a successful response 200 with additional mutators and inspectors", func() {
						var headerKey string
						var headerValue string

						BeforeEach(func() {
							headerKey = testHTTP.NewHeaderKey()
							headerValue = testHTTP.NewHeaderValue()
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV(auth.TidepoolSessionTokenHeaderKey, sessionToken),
									VerifyHeaderKV(headerKey, headerValue),
									VerifyBody(nil),
									RespondWith(http.StatusOK, nil)),
							)
						})

						It("returns success", func() {
							mutators := []request.RequestMutator{request.NewHeaderMutator(headerKey, headerValue)}
							inspector := request.NewHeadersInspector()
							reader, err = clnt.RequestStream(ctx, method, url, mutators, nil, inspector)
							Expect(err).ToNot(HaveOccurred())
							Expect(reader).ToNot(BeNil())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})
				})

				Context("RequestData", func() {
					It("returns error if context is missing", func() {
						Expect(clnt.RequestData(nil, method, url, nil, nil, nil)).To(MatchError("context is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error if method is missing", func() {
						Expect(clnt.RequestData(ctx, "", url, nil, nil, nil)).To(MatchError("method is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error if url is missing", func() {
						Expect(clnt.RequestData(ctx, method, "", nil, nil, nil)).To(MatchError("url is missing"))
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
							Expect(clnt.RequestData(ctx, method, url, nil, nil, nil)).To(Succeed())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a successful response 200 with additional mutators and inspectors", func() {
						var headerKey string
						var headerValue string

						BeforeEach(func() {
							headerKey = testHTTP.NewHeaderKey()
							headerValue = testHTTP.NewHeaderValue()
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV(auth.TidepoolSessionTokenHeaderKey, sessionToken),
									VerifyHeaderKV(headerKey, headerValue),
									VerifyBody(nil),
									RespondWith(http.StatusOK, nil)),
							)
						})

						It("returns success", func() {
							mutators := []request.RequestMutator{request.NewHeaderMutator(headerKey, headerValue)}
							inspector := request.NewHeadersInspector()
							Expect(clnt.RequestData(ctx, method, url, mutators, nil, nil, inspector)).To(Succeed())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})
				})
			})
		})
	})
})
