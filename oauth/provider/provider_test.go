package provider_test

import (
	"context"
	"net/http"
	"net/url"

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
			prvdr, err := oauthProvider.New(name, cfg, nil)
			Expect(err).To(MatchError("config is invalid; state salt is missing"))
			Expect(prvdr).To(BeNil())
		})
	})

	Context("with pkce enabled", func() {
		var prvdr *oauthProvider.Provider

		BeforeEach(func() {
			cfg.PKCEEnabled = true
			var err error
			prvdr, err = oauthProvider.New(name, cfg, nil)
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
				otherPrvdr, err := oauthProvider.New(name, otherCfg, nil)
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
			prvdr, err = oauthProvider.New(name, cfg, nil)
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
