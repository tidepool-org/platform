package platform_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"context"
	"net/http"
	"time"

	"github.com/tidepool-org/platform/auth"
	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHTTP "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Client", func() {
	Context("NewClient", func() {
		var config *platform.Config

		BeforeEach(func() {
			config = platform.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Config).ToNot(BeNil())
			config.Config.Address = testHTTP.NewAddress()
			config.Config.UserAgent = testHTTP.NewUserAgent()
			config.Timeout = time.Duration(testHTTP.NewTimeout()) * time.Second
		})

		It("returns an error if config is missing", func() {
			clnt, err := platform.NewClient(nil)
			Expect(err).To(MatchError("config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			clnt, err := platform.NewClient(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})
	})

	Context("with new client", func() {
		var timeout time.Duration
		var clnt *platform.Client

		BeforeEach(func() {
			timeout = time.Duration(testHTTP.NewTimeout()) * time.Second
			config := platform.NewConfig()
			Expect(config).ToNot(BeNil())
			Expect(config.Config).ToNot(BeNil())
			config.Config.Address = testHTTP.NewAddress()
			config.Config.UserAgent = testHTTP.NewUserAgent()
			config.Timeout = timeout
			var err error
			clnt, err = platform.NewClient(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})

		Context("HTTPClient", func() {
			It("returns successfully", func() {
				Expect(clnt.HTTPClient()).ToNot(BeNil())
			})

			It("uses the specified timeout", func() {
				httpClient := clnt.HTTPClient()
				Expect(httpClient).ToNot(BeNil())
				Expect(httpClient.Timeout).To(Equal(timeout))
			})
		})
	})

	Context("with started server and new client", func() {
		var server *Server
		var userAgent string
		var clnt *platform.Client
		var ctx context.Context
		var method string
		var path string
		var url string

		BeforeEach(func() {
			server = NewServer()
			userAgent = testHTTP.NewUserAgent()
			config := platform.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = server.URL()
			config.UserAgent = userAgent
			var err error
			clnt, err = platform.NewClient(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
			ctx = context.Background()
			method = testHTTP.NewMethod()
			path = testHTTP.NewPath()
			url = server.URL() + path
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("SendRequestAsUser", func() {
			It("returns error if context is missing", func() {
				Expect(clnt.SendRequestAsUser(nil, method, url, nil, nil, nil)).To(MatchError("context is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			Context("with session token", func() {
				var userID string
				var sessionToken string

				BeforeEach(func() {
					userID = test.NewString(8, test.CharsetAlphaNumeric)
					sessionToken = testAuth.NewSessionToken()
					ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, userID, sessionToken))
				})

				It("returns error if session token is missing", func() {
					ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, userID, ""))
					Expect(clnt.SendRequestAsUser(ctx, method, url, nil, nil, nil)).To(MatchError("unable to mutate request; session token is missing"))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				Context("with a successful response 200", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(method, path),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyHeaderKV(auth.TidepoolSessionTokenHeaderKey, sessionToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusOK, nil)),
						)
					})

					It("returns success", func() {
						Expect(clnt.SendRequestAsUser(ctx, method, url, nil, nil, nil)).To(Succeed())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})
			})
		})

		Context("SendRequestAsServer", func() {
			It("returns error if context is missing", func() {
				Expect(clnt.SendRequestAsServer(nil, method, url, nil, nil, nil)).To(MatchError("context is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			Context("with server token", func() {
				var serverSessionToken string

				BeforeEach(func() {
					serverSessionToken = testAuth.NewSessionToken()
					ctx = auth.NewContextWithServerSessionToken(ctx, serverSessionToken)
				})

				It("returns error if server token is missing", func() {
					ctx = auth.NewContextWithServerSessionToken(ctx, "")
					Expect(clnt.SendRequestAsServer(ctx, method, url, nil, nil, nil)).To(MatchError("server session token is missing"))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				Context("with a successful response 200", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(method, path),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyHeaderKV(auth.TidepoolSessionTokenHeaderKey, serverSessionToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusOK, nil)),
						)
					})

					It("returns success", func() {
						Expect(clnt.SendRequestAsServer(ctx, method, url, nil, nil, nil)).To(Succeed())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})
			})
		})
	})
})
