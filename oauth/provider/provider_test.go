package provider_test

import (
	"context"
	"net/http"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	oauthProviderTest "github.com/tidepool-org/platform/oauth/provider/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Provider", func() {
	var name string
	var config *oauthProvider.Config

	BeforeEach(func() {
		name = test.RandomStringFromRangeAndCharset(1, 16, test.CharsetAlphaNumeric)
		config = oauthProviderTest.RandomConfig()
	})

	Context("New", func() {
		It("returns an error when the name is missing", func() {
			provider, err := oauthProvider.New("", config, nil, nil)
			Expect(err).To(MatchError("name is missing"))
			Expect(provider).To(BeNil())
		})

		It("returns an error when the config is missing", func() {
			provider, err := oauthProvider.New(name, nil, nil, nil)
			Expect(err).To(MatchError("config is missing"))
			Expect(provider).To(BeNil())
		})

		It("returns an error when the config is invalid", func() {
			config.ClientID = ""
			provider, err := oauthProvider.New(name, config, nil, nil)
			Expect(err).To(MatchError("config is invalid; client id is empty"))
			Expect(provider).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(oauthProvider.New(name, config, nil, nil)).ToNot(BeNil())
		})

		It("returns a provider whose client timeout matches the config", func() {
			provider := test.Must(oauthProvider.New(name, config, nil, nil))
			Expect(provider.ClientTimeout()).To(Equal(config.ClientTimeout))
		})

		It("uses a copy of the default HTTP client with the config client timeout when the HTTP client is missing", func() {
			provider := test.Must(oauthProvider.New(name, config, nil, nil))
			Expect(provider.HTTPClient()).ToNot(BeNil())
			Expect(provider.HTTPClient()).ToNot(BeIdenticalTo(http.DefaultClient))
			Expect(provider.HTTPClient().Timeout).To(Equal(config.ClientTimeout))
		})

		It("uses a copy of the provided HTTP client without mutating it", func() {
			httpClient := &http.Client{}
			provider := test.Must(oauthProvider.New(name, config, httpClient, nil))
			Expect(provider.HTTPClient()).ToNot(BeNil())
			Expect(provider.HTTPClient()).ToNot(BeIdenticalTo(httpClient))
			Expect(httpClient.Timeout).To(BeZero())
		})

		It("sets the config client timeout on the copy when the provided HTTP client has no timeout", func() {
			httpClient := &http.Client{}
			provider := test.Must(oauthProvider.New(name, config, httpClient, nil))
			Expect(provider.HTTPClient().Timeout).To(Equal(config.ClientTimeout))
		})

		It("keeps the provided HTTP client timeout when it is not zero", func() {
			httpClientTimeout := config.ClientTimeout + time.Second
			httpClient := &http.Client{Timeout: httpClientTimeout}
			provider := test.Must(oauthProvider.New(name, config, httpClient, nil))
			Expect(provider.HTTPClient().Timeout).To(Equal(httpClientTimeout))
		})
	})

	Context("ExchangeAuthorizationCodeForToken with started server", func() {
		var server *Server
		var ctx context.Context
		var authorizationCode string
		var provider *oauthProvider.Provider

		BeforeEach(func() {
			server = NewServer()
			ctx = context.Background()
			authorizationCode = test.RandomStringFromRangeAndCharset(1, 32, test.CharsetAlphaNumeric)
			config.AuthStyleInParams = true
			config.TokenURL = server.URL() + "/token"
			provider = test.Must(oauthProvider.New(name, config, nil, nil))
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		It("returns the token from the token endpoint", func() {
			accessToken := test.RandomStringFromRangeAndCharset(1, 32, test.CharsetAlphaNumeric)
			refreshToken := test.RandomStringFromRangeAndCharset(1, 32, test.CharsetAlphaNumeric)
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/token"),
					VerifyForm(url.Values{"grant_type": []string{"authorization_code"}, "code": []string{authorizationCode}}),
					RespondWithJSONEncoded(http.StatusOK, map[string]any{
						"access_token":  accessToken,
						"token_type":    "Bearer",
						"refresh_token": refreshToken,
						"expires_in":    3600,
					}),
				),
			)
			token := test.Must(provider.ExchangeAuthorizationCodeForToken(ctx, authorizationCode))
			Expect(token).ToNot(BeNil())
			Expect(token.AccessToken).To(Equal(accessToken))
			Expect(token.TokenType).To(Equal("Bearer"))
			Expect(token.RefreshToken).To(Equal(refreshToken))
			Expect(token.ExpirationTime).To(BeTemporally("~", time.Now().Add(time.Hour), 10*time.Second))
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns an error when the token endpoint responds with an error", func() {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/token"),
					RespondWithJSONEncoded(http.StatusBadRequest, map[string]any{"error": "invalid_grant"}),
				),
			)
			token, err := provider.ExchangeAuthorizationCodeForToken(ctx, authorizationCode)
			Expect(err).To(MatchError(ContainSubstring("unable to exchange authorization code for token")))
			Expect(token).To(BeNil())
		})

		It("returns an error when the token endpoint stalls beyond the client timeout", func() {
			config.ClientTimeout = 100 * time.Millisecond
			provider = test.Must(oauthProvider.New(name, config, nil, nil))

			server.AppendHandlers(func(res http.ResponseWriter, req *http.Request) {
				time.Sleep(500 * time.Millisecond)
			})
			token, err := provider.ExchangeAuthorizationCodeForToken(ctx, authorizationCode)
			Expect(err).To(MatchError(ContainSubstring("Client.Timeout exceeded")))
			Expect(token).To(BeNil())
		})
	})
})
