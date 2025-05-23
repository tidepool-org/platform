package auth_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("OAuthToken", func() {
	Context("NewOAuthToken", func() {
		It("returns a new OAuthToken with zero values", func() {
			token := auth.NewOAuthToken()
			Expect(token).ToNot(BeNil())
			Expect(token.AccessToken).To(BeEmpty())
			Expect(token.TokenType).To(BeEmpty())
			Expect(token.RefreshToken).To(BeEmpty())
			Expect(token.ExpirationTime).To(BeZero())
			Expect(token.IDToken).To(BeNil())
		})
	})

	Context("NewOAuthTokenFromRawToken", func() {
		It("returns error if rawToken is missing", func() {
			token, err := auth.NewOAuthTokenFromRawToken(nil)
			Expect(token).To(BeNil())
			Expect(err).To(MatchError("raw token is missing"))
		})

		It("returns OAuthToken with values from rawToken", func() {
			rawToken := authTest.RandomToken().RawToken()
			token, err := auth.NewOAuthTokenFromRawToken(rawToken)
			Expect(err).ToNot(HaveOccurred())
			Expect(token.AccessToken).To(Equal(rawToken.AccessToken))
			Expect(token.TokenType).To(Equal(rawToken.TokenType))
			Expect(token.RefreshToken).To(Equal(rawToken.RefreshToken))
			Expect(token.ExpirationTime).To(Equal(rawToken.Expiry))
			Expect(token.IDToken).To(Equal(auth.GetIDToken(rawToken)))
		})
	})

	Context("OAuthToken", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *auth.OAuthToken), expectedErrors ...error) {
					expectedDatum := authTest.RandomToken()
					object := authTest.NewObjectFromToken(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &auth.OAuthToken{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *auth.OAuthToken) {},
				),
				Entry("accessToken invalid type",
					func(object map[string]any, expectedDatum *auth.OAuthToken) {
						object["accessToken"] = true
						expectedDatum.AccessToken = ""
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/accessToken"),
				),
				Entry("tokenType invalid type",
					func(object map[string]any, expectedDatum *auth.OAuthToken) {
						object["tokenType"] = true
						expectedDatum.TokenType = ""
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/tokenType"),
				),
				Entry("refreshToken invalid type",
					func(object map[string]any, expectedDatum *auth.OAuthToken) {
						object["refreshToken"] = true
						expectedDatum.RefreshToken = ""
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/refreshToken"),
				),
				Entry("expirationTime invalid type",
					func(object map[string]any, expectedDatum *auth.OAuthToken) {
						object["expirationTime"] = true
						expectedDatum.ExpirationTime = time.Time{}
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/expirationTime"),
				),
				Entry("idToken invalid type",
					func(object map[string]any, expectedDatum *auth.OAuthToken) {
						object["idToken"] = true
						expectedDatum.IDToken = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/idToken"),
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *auth.OAuthToken) {
						object["accessToken"] = true
						object["tokenType"] = true
						object["refreshToken"] = true
						object["expirationTime"] = true
						object["idToken"] = true
						expectedDatum.AccessToken = ""
						expectedDatum.TokenType = ""
						expectedDatum.RefreshToken = ""
						expectedDatum.ExpirationTime = time.Time{}
						expectedDatum.IDToken = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/accessToken"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/tokenType"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/refreshToken"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/expirationTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/idToken"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *auth.OAuthToken), expectedErrors ...error) {
					datum := authTest.RandomToken()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *auth.OAuthToken) {},
				),
				Entry("accessToken empty",
					func(datum *auth.OAuthToken) { datum.AccessToken = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/accessToken"),
				),
				Entry("idToken empty",
					func(datum *auth.OAuthToken) { datum.IDToken = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/idToken"),
				),
				Entry("multiple errors",
					func(datum *auth.OAuthToken) {
						datum.AccessToken = ""
						datum.IDToken = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/accessToken"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/idToken"),
				),
			)
		})

		Context("Refreshed", func() {
			It("returns error if rawToken is nil", func() {
				token := auth.NewOAuthToken()
				refreshed, err := token.Refreshed(nil)
				Expect(refreshed).To(BeNil())
				Expect(err).To(MatchError("raw token is missing"))
			})

			It("returns a refreshed token with updated values", func() {
				token := authTest.RandomToken()
				rawToken := authTest.RandomToken().RawToken()
				refreshed, err := token.Refreshed(rawToken)
				Expect(err).ToNot(HaveOccurred())
				Expect(refreshed.AccessToken).To(Equal(rawToken.AccessToken))
				Expect(refreshed.TokenType).To(Equal(rawToken.TokenType))
				Expect(refreshed.RefreshToken).To(Equal(rawToken.RefreshToken))
				Expect(refreshed.ExpirationTime).To(Equal(rawToken.Expiry))
				Expect(refreshed.IDToken).To(Equal(auth.GetIDToken(rawToken)))
			})
		})

		Context("Expire", func() {
			It("sets ExpirationTime to the past", func() {
				token := authTest.RandomToken()
				token.Expire()
				Expect(token.ExpirationTime).To(BeTemporally("<", time.Now()))
			})
		})

		Context("RawToken", func() {
			It("returns an oauth2.Token with matching values", func() {
				token := authTest.RandomToken()
				rawToken := token.RawToken()
				Expect(rawToken.AccessToken).To(Equal(token.AccessToken))
				Expect(rawToken.TokenType).To(Equal(token.TokenType))
				Expect(rawToken.RefreshToken).To(Equal(token.RefreshToken))
				Expect(rawToken.Expiry).To(Equal(token.ExpirationTime))
				Expect(auth.GetIDToken(rawToken)).To(Equal(token.IDToken))
			})
		})

		Context("MatchesRawToken", func() {
			DescribeTable("returns false if any field does not match",
				func(mutator func(token *auth.OAuthToken, rawToken *oauth2.Token)) {
					token := authTest.RandomToken()
					rawToken := token.RawToken()
					mutator(token, rawToken)
					Expect(token.MatchesRawToken(rawToken)).To(BeFalse())
				},
				Entry("AccessToken does not match", func(token *auth.OAuthToken, rawToken *oauth2.Token) {
					rawToken.AccessToken = test.RandomString()
				}),
				Entry("TokenType does not match", func(token *auth.OAuthToken, rawToken *oauth2.Token) {
					rawToken.TokenType = test.RandomString()
				}),
				Entry("RefreshToken does not match", func(token *auth.OAuthToken, rawToken *oauth2.Token) {
					rawToken.RefreshToken = test.RandomString()
				}),
				Entry("Expiry does not match", func(token *auth.OAuthToken, rawToken *oauth2.Token) {
					rawToken.Expiry = test.RandomTime()
				}),
				Entry("IDToken does not match", func(token *auth.OAuthToken, rawToken *oauth2.Token) {
					*rawToken = *rawToken.WithExtra(map[string]any{auth.KeyIDToken: test.RandomString()})
				}),
			)

			It("returns true if all fields match", func() {
				token := authTest.RandomToken()
				Expect(token.MatchesRawToken(token.RawToken())).To(BeTrue())
			})
		})
	})

	Context("GetIDToken", func() {
		It("returns nil if id token is not present", func() {
			rawToken := &oauth2.Token{}
			Expect(auth.GetIDToken(rawToken)).To(BeNil())
		})

		It("returns id token if present and not empty", func() {
			idToken := test.RandomString()
			rawToken := &oauth2.Token{}
			rawToken = rawToken.WithExtra(map[string]any{auth.KeyIDToken: idToken})
			result := auth.GetIDToken(rawToken)
			Expect(result).ToNot(BeNil())
			Expect(*result).To(Equal(idToken))
		})
	})

	Context("SetIDToken", func() {
		It("does not set id token if nil", func() {
			rawToken := &oauth2.Token{}
			result := auth.SetIDToken(rawToken, nil)
			Expect(result.Extra(auth.KeyIDToken)).To(BeNil())
		})

		It("sets id token if not nil", func() {
			idToken := test.RandomString()
			rawToken := &oauth2.Token{}
			result := auth.SetIDToken(rawToken, &idToken)
			Expect(result.Extra(auth.KeyIDToken)).To(Equal(idToken))
		})
	})
})
