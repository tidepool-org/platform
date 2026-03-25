package work_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"
	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/auth"
	providerSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oauth"
	oauthTest "github.com/tidepool-org/platform/oauth/test"
	oauthWork "github.com/tidepool-org/platform/oauth/work"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("mixin", func() {
	Context("with context and mocks", func() {
		var mockLogger *logTest.Logger
		var ctx context.Context
		var mockController *gomock.Controller
		var mockWorkProvider *workTest.Provider
		var mockProviderSessionMixin *providerSessionWorkTest.MockMixin

		BeforeEach(func() {
			mockLogger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), mockLogger)
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkProvider = workTest.NewProvider(ctx)
			mockProviderSessionMixin = providerSessionWorkTest.NewMockMixin(mockController)
		})

		Context("NewMixin", func() {
			It("returns an error when provider is missing", func() {
				mixin, err := oauthWork.NewMixin(nil, mockProviderSessionMixin)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when provider session mixin is missing", func() {
				mixin, err := oauthWork.NewMixin(mockWorkProvider, nil)
				Expect(err).To(MatchError("provider session mixin is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := oauthWork.NewMixin(mockWorkProvider, mockProviderSessionMixin)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("mixin", func() {
			var mixin oauthWork.Mixin

			BeforeEach(func() {
				var err error
				mixin, err = oauthWork.NewMixin(mockWorkProvider, mockProviderSessionMixin)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})

			Context("TokenSource", func() {
				It("returns itself", func() {
					Expect(mixin.TokenSource()).To(Equal(mixin))
				})
			})

			Context("FetchTokenSource", func() {
				It("returns failed process result when provider session mixin does not have a provider session", func() {
					mockProviderSessionMixin.EXPECT().HasProviderSession().Return(false)
					Expect(mixin.FetchTokenSource()).To(workTest.MatchFailedProcessResultError(MatchError("provider session is missing")))
				})

				It("returns failed process result when provider session mixin fails to create a new source with the token", func() {
					providerSession := authTest.RandomProviderSession(test.AllowOptional())
					providerSession.OAuthToken = nil
					mockProviderSessionMixin.EXPECT().HasProviderSession().Return(true)
					mockProviderSessionMixin.EXPECT().ProviderSession().Return(providerSession)
					Expect(mixin.FetchTokenSource()).To(workTest.MatchFailedProcessResultError(MatchError("unable to create token source; token is missing")))
				})

				It("sets the token source and returns nil on success", func() {
					providerSession := authTest.RandomProviderSession(test.AllowOptional())
					mockProviderSessionMixin.EXPECT().HasProviderSession().Return(true)
					mockProviderSessionMixin.EXPECT().ProviderSession().Return(providerSession)
					Expect(mixin.FetchTokenSource()).To(BeNil())
				})
			})

			Context("without a token source", func() {
				Context("HTTPClient", func() {
					It("returns an error", func() {
						tokenSourceSource := oauthTest.NewMockTokenSourceSource(mockController)
						httpClient, err := mixin.HTTPClient(ctx, tokenSourceSource)
						Expect(err).To(MatchError("token source is missing"))
						Expect(httpClient).To(BeNil())
					})
				})

				Context("UpdateToken", func() {
					It("returns an error", func() {
						updated, err := mixin.UpdateToken(ctx)
						Expect(err).To(MatchError("token source is missing"))
						Expect(updated).To(BeFalse())
					})
				})

				Context("ExpireToken", func() {
					It("returns an error", func() {
						expired, err := mixin.ExpireToken(ctx)
						Expect(err).To(MatchError("token source is missing"))
						Expect(expired).To(BeFalse())
					})
				})
			})

			Context("with a token source", func() {
				var providerSession *auth.ProviderSession
				var tokenSourceSource *oauthTest.MockTokenSourceSource

				BeforeEach(func() {
					providerSession = authTest.RandomProviderSession(test.AllowOptional())
					mockProviderSessionMixin.EXPECT().HasProviderSession().Return(true)
					mockProviderSessionMixin.EXPECT().ProviderSession().Return(providerSession)
					Expect(mixin.FetchTokenSource()).To(BeNil())
					tokenSourceSource = oauthTest.NewMockTokenSourceSource(mockController)
				})

				Context("HTTPClient", func() {

					BeforeEach(func() {
					})

					It("returns an error when context is nil", func() {
						httpClient, err := mixin.HTTPClient(context.Context(nil), tokenSourceSource)
						Expect(err).To(MatchError("context is missing"))
						Expect(httpClient).To(BeNil())
					})

					It("returns an error when token source source is nil", func() {
						httpClient, err := mixin.HTTPClient(ctx, oauth.TokenSourceSource(nil))
						Expect(err).To(MatchError("token source source is missing"))
						Expect(httpClient).To(BeNil())
					})

					It("returns successfully", func() {
						tokenSourceSource.EXPECT().TokenSource(ctx, providerSession.OAuthToken).Return(oauth2.StaticTokenSource(providerSession.OAuthToken.RawToken()), nil)
						httpClient, err := mixin.HTTPClient(ctx, tokenSourceSource)
						Expect(err).To(Not(HaveOccurred()))
						Expect(httpClient).ToNot(BeNil())
					})
				})

				Context("with a token source", func() {
					var mockTokenSource *MockTokenSource

					BeforeEach(func() {
						mockTokenSource = &MockTokenSource{
							token: providerSession.OAuthToken.RawToken(),
							err:   nil,
						}
					})

					JustBeforeEach(func() {
						tokenSourceSource.EXPECT().TokenSource(ctx, providerSession.OAuthToken).Return(mockTokenSource, nil)
						httpClient, err := mixin.HTTPClient(ctx, tokenSourceSource)
						Expect(err).To(Not(HaveOccurred()))
						Expect(httpClient).ToNot(BeNil())
					})

					Context("UpdateToken", func() {
						It("returns an error when the token source returns an error", func() {
							mockTokenSource.err = errorsTest.RandomError()
							updated, err := mixin.UpdateToken(ctx)
							Expect(err).To(MatchError(fmt.Sprintf("unable to get token; %s", mockTokenSource.err.Error())))
							Expect(updated).To(BeFalse())
						})

						It("returns successfully if not updated", func() {
							updated, err := mixin.UpdateToken(ctx)
							Expect(err).ToNot(HaveOccurred())
							Expect(updated).To(BeFalse())
						})

						Context("with updated token source", func() {
							var updatedOAuthToken *auth.OAuthToken

							BeforeEach(func() {
								updatedOAuthToken = authTest.RandomToken()
								mockTokenSource.token = updatedOAuthToken.RawToken()
							})

							It("returns error if update provider session returns an error", func() {
								failingProcessResult := workTest.RandomFailingProcessResult()
								mockProviderSessionMixin.EXPECT().ProviderSession().Return(providerSession)
								mockProviderSessionMixin.EXPECT().UpdateProviderSession(&auth.ProviderSessionUpdate{OAuthToken: updatedOAuthToken, ExternalID: providerSession.ExternalID}).Return(failingProcessResult)
								updated, err := mixin.UpdateToken(ctx)
								errorsTest.ExpectEqual(err, failingProcessResult.Error())
								Expect(updated).To(BeTrue())
							})

							It("returns successfully", func() {
								mockProviderSessionMixin.EXPECT().ProviderSession().Return(providerSession)
								mockProviderSessionMixin.EXPECT().UpdateProviderSession(&auth.ProviderSessionUpdate{OAuthToken: updatedOAuthToken, ExternalID: providerSession.ExternalID}).Return(nil)
								updated, err := mixin.UpdateToken(ctx)
								Expect(err).ToNot(HaveOccurred())
								Expect(updated).To(BeTrue())
							})
						})
					})

					Context("ExpireToken", func() {
						It("returns successfully if not expired", func() {
							providerSession.OAuthToken.ExpirationTime = test.RandomTimeBeforeNow()
							expired, err := mixin.ExpireToken(ctx)
							Expect(err).ToNot(HaveOccurred())
							Expect(expired).To(BeFalse())
						})

						Context("with expired token source", func() {
							BeforeEach(func() {
								providerSession.OAuthToken.ExpirationTime = test.RandomTimeAfterNow()
							})

							It("returns error if update provider session returns an error", func() {
								failingProcessResult := workTest.RandomFailingProcessResult()
								mockProviderSessionMixin.EXPECT().ProviderSession().Return(providerSession)
								mockProviderSessionMixin.EXPECT().UpdateProviderSession(MatchProviderSessionIgnoringExpirationTime(providerSession)).Return(failingProcessResult)
								expired, err := mixin.ExpireToken(ctx)
								errorsTest.ExpectEqual(err, failingProcessResult.Error())
								Expect(expired).To(BeTrue())
							})

							It("returns successfully", func() {
								mockProviderSessionMixin.EXPECT().ProviderSession().Return(providerSession)
								mockProviderSessionMixin.EXPECT().UpdateProviderSession(MatchProviderSessionIgnoringExpirationTime(providerSession)).Return(nil)
								expired, err := mixin.ExpireToken(ctx)
								Expect(err).ToNot(HaveOccurred())
								Expect(expired).To(BeTrue())
							})
						})
					})
				})
			})
		})
	})
})

type MockTokenSource struct {
	token *oauth2.Token
	err   error
}

func (m MockTokenSource) Token() (*oauth2.Token, error) {
	return m.token, m.err
}

func MatchProviderSessionIgnoringExpirationTime(providerSession *auth.ProviderSession) gomock.Matcher {
	return &providerSessionMatcherIgnoringExpirationTime{
		providerSession: providerSession,
	}
}

type providerSessionMatcherIgnoringExpirationTime struct {
	providerSession *auth.ProviderSession
}

func (p *providerSessionMatcherIgnoringExpirationTime) Matches(match any) bool {
	if providerSessionUpdate, ok := match.(*auth.ProviderSessionUpdate); ok {
		return providerSessionUpdate != nil &&
			providerSessionUpdate.OAuthToken.AccessToken == p.providerSession.OAuthToken.AccessToken &&
			providerSessionUpdate.OAuthToken.TokenType == p.providerSession.OAuthToken.TokenType &&
			providerSessionUpdate.OAuthToken.RefreshToken == p.providerSession.OAuthToken.RefreshToken &&
			pointer.EqualStringArray(providerSessionUpdate.OAuthToken.Scope, p.providerSession.OAuthToken.Scope) &&
			pointer.EqualString(providerSessionUpdate.OAuthToken.IDToken, p.providerSession.OAuthToken.IDToken) &&
			pointer.EqualString(providerSessionUpdate.ExternalID, p.providerSession.ExternalID)
	}
	return false
}

func (p *providerSessionMatcherIgnoringExpirationTime) String() string {
	return fmt.Sprintf("matches provider session update with provider session %v", p.providerSession)
}
