package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/golang/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	dataClient "github.com/tidepool-org/platform/data/client"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Client", func() {
	Context("New", func() {
		var config *platform.Config

		BeforeEach(func() {
			config = platform.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = testHttp.NewAddress()
			config.UserAgent = testHttp.NewUserAgent()
		})

		It("returns an error if config is missing", func() {
			clnt, err := dataClient.New(nil, platform.AuthorizeAsService)
			Expect(err).To(MatchError("config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config address is missing", func() {
			config.Address = ""
			clnt, err := dataClient.New(config, platform.AuthorizeAsService)
			Expect(err).To(MatchError("config is invalid; address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			clnt, err := dataClient.New(config, platform.AuthorizeAsService)
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
			userAgent = testHttp.NewUserAgent()
			config := platform.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = server.URL()
			config.UserAgent = userAgent
			var err error
			clnt, err = dataClient.New(config, platform.AuthorizeAsService)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
			ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("DestroyDataForUserByID", func() {
			var userID string

			BeforeEach(func() {
				userID = userTest.RandomID()
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
				var serverSessionTokenProviderController *gomock.Controller
				var serverSessionTokenProvider *authTest.MockServerSessionTokenProvider
				var sessionToken string

				BeforeEach(func() {
					serverSessionTokenProviderController = gomock.NewController(GinkgoT())
					serverSessionTokenProvider = authTest.NewMockServerSessionTokenProvider(serverSessionTokenProviderController)
					sessionToken = dataTest.NewSessionToken()
					serverSessionTokenProvider.EXPECT().ServerSessionToken().Return(sessionToken, nil).AnyTimes()
					ctx = auth.NewContextWithServerSessionTokenProvider(ctx, serverSessionTokenProvider)
				})

				Context("with an unauthorized response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("DELETE", fmt.Sprintf("/v1/users/%s/data", userID)),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyHeaderKV("X-Tidepool-Session-Token", sessionToken),
								VerifyBody(nil),
								RespondWith(http.StatusUnauthorized, nil)),
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
								VerifyHeaderKV("X-Tidepool-Session-Token", sessionToken),
								VerifyBody(nil),
								RespondWith(http.StatusForbidden, nil)),
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
								VerifyHeaderKV("X-Tidepool-Session-Token", sessionToken),
								VerifyBody(nil),
								RespondWith(http.StatusOK, nil)),
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

	Context("NewSerializableDataErrorResponseParser", func() {
		It("returns success", func() {
			Expect(dataClient.NewSerializableDataErrorResponseParser()).ToNot(BeNil())
		})
	})

	Context("SerializableDataErrorResponseParser", func() {
		It("returns nil if response body is not parseable", func() {
			serializableErrorResponseParser := dataClient.NewSerializableDataErrorResponseParser()
			err := serializableErrorResponseParser.ParseErrorResponse(context.Background(), &http.Response{Body: io.NopCloser(bytes.NewReader([]byte("NOT JSON")))}, testHttp.NewRequest())
			Expect(err).To(BeNil())
		})

		It("returns deserialized error if response body is parseable", func() {
			responseErr := request.ErrorResourceNotFoundWithID(test.RandomStringFromRangeAndCharset(1, 16, test.CharsetHexidecimalLowercase))
			body, err := json.Marshal(map[string]any{"errors": errors.Serializable{Error: responseErr}})
			Expect(err).ToNot(HaveOccurred())
			Expect(body).ToNot(BeNil())
			serializableErrorResponseParser := dataClient.NewSerializableDataErrorResponseParser()
			err = serializableErrorResponseParser.ParseErrorResponse(context.Background(), &http.Response{Body: io.NopCloser(bytes.NewReader(body))}, testHttp.NewRequest())
			Expect(err).To(Equal(responseErr))
		})
	})
})
