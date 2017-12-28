package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/platform"
	testHTTP "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Client", func() {
	Context("New", func() {
		var config *platform.Config

		BeforeEach(func() {
			config = platform.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = testHTTP.NewAddress()
			config.UserAgent = testHTTP.NewUserAgent()
			config.Timeout = 30 * time.Second
		})

		It("returns an error if config is missing", func() {
			clnt, err := dataClient.New(nil)
			Expect(err).To(MatchError("config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config address is missing", func() {
			config.Address = ""
			clnt, err := dataClient.New(config)
			Expect(err).To(MatchError("config is invalid; address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config user agent is missing", func() {
			config.UserAgent = ""
			clnt, err := dataClient.New(config)
			Expect(err).To(MatchError("config is invalid; user agent is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			clnt, err := dataClient.New(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})
	})

	Context("with started server and new client", func() {
		var server *Server
		var userAgent string
		var clnt dataClient.Client
		var ctx context.Context

		BeforeEach(func() {
			server = NewServer()
			userAgent = testHTTP.NewUserAgent()
			config := platform.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = server.URL()
			config.UserAgent = userAgent
			config.Timeout = 30 * time.Second
			var err error
			clnt, err = dataClient.New(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
			ctx = context.Background()
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("DestroyDataForUserByID", func() {
			var userID string

			BeforeEach(func() {
				userID = id.New()
			})

			It("returns error if context is missing", func() {
				Expect(clnt.DestroyDataForUserByID(nil, userID)).To(MatchError("context is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if user id is missing", func() {
				Expect(clnt.DestroyDataForUserByID(ctx, "")).To(MatchError("user id is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			Context("with server token", func() {
				var token string

				BeforeEach(func() {
					token = id.New()
					ctx = auth.NewContextWithServerSessionToken(ctx, token)
				})

				Context("with an unauthorized response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("DELETE", fmt.Sprintf("/v1/users/%s/data", userID)),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyHeaderKV("X-Tidepool-Session-Token", token),
								VerifyBody([]byte{}),
								RespondWith(http.StatusUnauthorized, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := clnt.DestroyDataForUserByID(ctx, userID)
						Expect(err).To(MatchError("authentication token is invalid"))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a forbidden response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("DELETE", fmt.Sprintf("/v1/users/%s/data", userID)),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyHeaderKV("X-Tidepool-Session-Token", token),
								VerifyBody([]byte{}),
								RespondWith(http.StatusForbidden, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := clnt.DestroyDataForUserByID(ctx, userID)
						Expect(err).To(MatchError("authentication token is not authorized for requested action"))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a successful response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("DELETE", fmt.Sprintf("/v1/users/%s/data", userID)),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyHeaderKV("X-Tidepool-Session-Token", token),
								VerifyBody([]byte{}),
								RespondWith(http.StatusOK, nil, nil)),
						)
					})

					It("returns success", func() {
						Expect(clnt.DestroyDataForUserByID(ctx, userID)).To(Succeed())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})
			})
		})
	})
})
