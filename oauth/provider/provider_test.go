package provider_test

import (
	"context"
	"net/http"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	"golang.org/x/oauth2"

	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	oauthProviderTest "github.com/tidepool-org/platform/oauth/provider/test"
	"github.com/tidepool-org/platform/pointer"
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
		var state string
		var provider *oauthProvider.Provider

		BeforeEach(func() {
			server = NewServer()
			ctx = context.Background()
			authorizationCode = test.RandomStringFromRangeAndCharset(1, 32, test.CharsetAlphaNumeric)
			state = test.RandomStringFromRangeAndCharset(1, 32, test.CharsetAlphaNumeric)
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
			token := test.Must(provider.ExchangeAuthorizationCodeForToken(ctx, authorizationCode, state))
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
			token, err := provider.ExchangeAuthorizationCodeForToken(ctx, authorizationCode, state)
			Expect(err).To(MatchError(ContainSubstring("unable to exchange authorization code for token")))
			Expect(token).To(BeNil())
		})

		It("returns an error when the token endpoint stalls beyond the client timeout", func() {
			config.ClientTimeout = 100 * time.Millisecond
			provider = test.Must(oauthProvider.New(name, config, nil, nil))

			server.AppendHandlers(func(res http.ResponseWriter, req *http.Request) {
				time.Sleep(500 * time.Millisecond)
			})
			token, err := provider.ExchangeAuthorizationCodeForToken(ctx, authorizationCode, state)
			Expect(err).To(MatchError(ContainSubstring("Client.Timeout exceeded")))
			Expect(token).To(BeNil())
		})
	})

	Context("PKCE", func() {
		var name string
		var cfg *oauthProvider.Config
		var server *Server
		var state string
		var code string

		BeforeEach(func() {
			name = test.RandomStringFromCharset(test.CharsetAlphaNumeric)
			server = NewServer()
			cfg = oauthProviderTest.RandomConfig()
			cfg.StateSalt = pointer.FromString(oauthProviderTest.RandomStateSalt())
			cfg.TokenURL = server.URL() + "/token"
			state = test.RandomStringFromCharset(test.CharsetAlphaNumeric)
			code = test.RandomStringFromCharset(test.CharsetAlphaNumeric)
		})

		AfterEach(func() {
			server.Close()
		})

		tokenResponse := func() map[string]any {
			return map[string]any{
				"access_token":  test.RandomStringFromCharset(test.CharsetAlphaNumeric),
				"token_type":    "Bearer",
				"refresh_token": test.RandomStringFromCharset(test.CharsetAlphaNumeric),
				"expires_in":    3600,
			}
		}

		Context("New", func() {
			It("returns an error when pkce is enabled without a state salt", func() {
				cfg.CookieDisabled = true
				cfg.StateSalt = nil
				cfg.PKCEEnabled = true
				prvdr, err := oauthProvider.New(name, cfg, nil, nil)
				Expect(err).To(MatchError("config is invalid; state salt is missing"))
				Expect(prvdr).To(BeNil())
			})
		})

		Context("with pkce enabled", func() {
			var prvdr *oauthProvider.Provider

			BeforeEach(func() {
				cfg.PKCEEnabled = true
				var err error
				prvdr, err = oauthProvider.New(name, cfg, nil, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(prvdr).ToNot(BeNil())
			})

			It("reports pkce enabled", func() {
				Expect(prvdr.PKCEEnabled()).To(BeTrue())
			})

			Context("CodeVerifierForState", func() {
				It("returns a verifier with valid length and characters", func() {
					verifier := prvdr.CodeVerifierForState(state)
					Expect(verifier).To(HaveLen(43))
					Expect(verifier).To(MatchRegexp(`^[A-Za-z0-9._~-]+$`))
				})

				It("returns the same verifier for the same state", func() {
					Expect(prvdr.CodeVerifierForState(state)).To(Equal(prvdr.CodeVerifierForState(state)))
				})

				It("returns a different verifier for a different state", func() {
					Expect(prvdr.CodeVerifierForState(state)).ToNot(Equal(prvdr.CodeVerifierForState(state + "x")))
				})

				It("returns a different verifier for a different state salt", func() {
					otherCfg := oauthProviderTest.CloneConfig(cfg)
					otherCfg.StateSalt = pointer.FromString(*cfg.StateSalt + "x")
					otherPrvdr, err := oauthProvider.New(name, otherCfg, nil, nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(prvdr.CodeVerifierForState(state)).ToNot(Equal(otherPrvdr.CodeVerifierForState(state)))
				})
			})

			Context("GetAuthorizationCodeURLWithState", func() {
				It("includes the state and the S256 code challenge", func() {
					authorizeURL, err := url.Parse(prvdr.GetAuthorizationCodeURLWithState(state))
					Expect(err).ToNot(HaveOccurred())
					query := authorizeURL.Query()
					Expect(query.Get("state")).To(Equal(state))
					Expect(query.Get("code_challenge_method")).To(Equal("S256"))
					Expect(query.Get("code_challenge")).To(Equal(oauth2.S256ChallengeFromVerifier(prvdr.CodeVerifierForState(state))))
				})
			})

			Context("ExchangeAuthorizationCodeForToken", func() {
				It("sends the code verifier derived from the state", func() {
					response := tokenResponse()
					server.AppendHandlers(CombineHandlers(
						VerifyRequest(http.MethodPost, "/token"),
						VerifyFormKV("grant_type", "authorization_code"),
						VerifyFormKV("code", code),
						VerifyFormKV("code_verifier", prvdr.CodeVerifierForState(state)),
						RespondWithJSONEncoded(http.StatusOK, response),
					))
					token, err := prvdr.ExchangeAuthorizationCodeForToken(context.Background(), code, state)
					Expect(err).ToNot(HaveOccurred())
					Expect(token).ToNot(BeNil())
					Expect(token.AccessToken).To(Equal(response["access_token"]))
					Expect(token.RefreshToken).To(Equal(response["refresh_token"]))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})

		Context("with pkce disabled", func() {
			var prvdr *oauthProvider.Provider

			BeforeEach(func() {
				var err error
				prvdr, err = oauthProvider.New(name, cfg, nil, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(prvdr).ToNot(BeNil())
			})

			It("reports pkce disabled", func() {
				Expect(prvdr.PKCEEnabled()).To(BeFalse())
			})

			Context("GetAuthorizationCodeURLWithState", func() {
				It("includes the state without a code challenge", func() {
					authorizeURL, err := url.Parse(prvdr.GetAuthorizationCodeURLWithState(state))
					Expect(err).ToNot(HaveOccurred())
					query := authorizeURL.Query()
					Expect(query.Get("state")).To(Equal(state))
					Expect(query).ToNot(HaveKey("code_challenge"))
					Expect(query).ToNot(HaveKey("code_challenge_method"))
				})
			})

			Context("ExchangeAuthorizationCodeForToken", func() {
				It("does not send a code verifier", func() {
					server.AppendHandlers(CombineHandlers(
						VerifyRequest(http.MethodPost, "/token"),
						VerifyFormKV("grant_type", "authorization_code"),
						VerifyFormKV("code", code),
						func(res http.ResponseWriter, req *http.Request) {
							Expect(req.ParseForm()).To(Succeed())
							Expect(req.PostForm).ToNot(HaveKey("code_verifier"))
						},
						RespondWithJSONEncoded(http.StatusOK, tokenResponse()),
					))
					token, err := prvdr.ExchangeAuthorizationCodeForToken(context.Background(), code, state)
					Expect(err).ToNot(HaveOccurred())
					Expect(token).ToNot(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
