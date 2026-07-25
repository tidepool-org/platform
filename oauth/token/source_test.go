package context_test

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"
	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/auth"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	oauthTest "github.com/tidepool-org/platform/oauth/test"
	oauthToken "github.com/tidepool-org/platform/oauth/token"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Source", func() {
	var token *auth.OAuthToken
	var mockController *gomock.Controller
	var mockTokenSourceSource *oauthTest.MockTokenSourceSource

	BeforeEach(func() {
		token = &auth.OAuthToken{
			AccessToken:    test.RandomStringFromRangeAndCharset(1, 32, test.CharsetAlphaNumeric),
			TokenType:      "Bearer",
			RefreshToken:   test.RandomStringFromRangeAndCharset(1, 32, test.CharsetAlphaNumeric),
			ExpirationTime: test.FutureFarTime(),
		}
		mockController = gomock.NewController(GinkgoT())
		mockTokenSourceSource = oauthTest.NewMockTokenSourceSource(mockController)
	})

	Context("NewSourceWithToken", func() {
		It("returns an error when the token is missing", func() {
			source, err := oauthToken.NewSourceWithToken(nil)
			Expect(err).To(MatchError("token is missing"))
			Expect(source).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(oauthToken.NewSourceWithToken(token)).ToNot(BeNil())
		})
	})

	Context("with new source", func() {
		var source *oauthToken.Source

		BeforeEach(func() {
			source = test.Must(oauthToken.NewSourceWithToken(token))
		})

		Context("HTTPClient", func() {
			var ctx context.Context

			BeforeEach(func() {
				ctx = context.Background()
			})

			It("returns an error when the context is missing", func() {
				httpClient, err := source.HTTPClient(context.Context(nil), mockTokenSourceSource)
				Expect(err).To(MatchError("context is missing"))
				Expect(httpClient).To(BeNil())
			})

			It("returns an error when the token source source is missing", func() {
				httpClient, err := source.HTTPClient(ctx, nil)
				Expect(err).To(MatchError("token source source is missing"))
				Expect(httpClient).To(BeNil())
			})

			It("returns an error when the token source source returns an error", func() {
				tokenSourceErr := errorsTest.RandomError()
				mockTokenSourceSource.EXPECT().TokenSource(gomock.Any(), gomock.Eq(token)).Return(nil, tokenSourceErr)
				httpClient, err := source.HTTPClient(ctx, mockTokenSourceSource)
				Expect(err).To(Equal(tokenSourceErr))
				Expect(httpClient).To(BeNil())
			})

			When("the token source source returns successfully", func() {
				BeforeEach(func() {
					mockTokenSourceSource.EXPECT().TokenSource(gomock.Any(), gomock.Eq(token)).Return(oauth2.StaticTokenSource(token.RawToken()), nil)
				})

				It("returns an http client without a timeout when the context has no http client", func() {
					httpClient := test.Must(source.HTTPClient(ctx, mockTokenSourceSource))
					Expect(httpClient).ToNot(BeNil())
					Expect(httpClient.Timeout).To(BeZero())
				})

				It("returns an http client with the timeout of the http client in the context", func() {
					timeout := test.RandomDurationFromRange(time.Second, time.Hour)
					ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Timeout: timeout})
					httpClient := test.Must(source.HTTPClient(ctx, mockTokenSourceSource))
					Expect(httpClient).ToNot(BeNil())
					Expect(httpClient.Timeout).To(Equal(timeout))
				})

				It("returns the same http client on subsequent calls", func() {
					httpClient := test.Must(source.HTTPClient(ctx, mockTokenSourceSource))
					Expect(test.Must(source.HTTPClient(ctx, mockTokenSourceSource))).To(BeIdenticalTo(httpClient))
				})
			})
		})
	})
})
