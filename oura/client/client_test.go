package client_test

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"
	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/client"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	oauthClient "github.com/tidepool-org/platform/oauth/client"
	oauthProviderTest "github.com/tidepool-org/platform/oauth/provider/test"
	oauthTest "github.com/tidepool-org/platform/oauth/test"
	"github.com/tidepool-org/platform/oura"
	ouraClient "github.com/tidepool-org/platform/oura/client"
	ouraClientTest "github.com/tidepool-org/platform/oura/client/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("client", func() {
	It("HeaderClientID is expected", func() {
		Expect(ouraClient.HeaderClientID).To(Equal("x-client-id"))
	})

	It("HeaderClientSecret is expected", func() {
		Expect(ouraClient.HeaderClientSecret).To(Equal("x-client-secret"))
	})

	Context("with server and base client", func() {
		var (
			logger                *logTest.Logger
			ctx                   context.Context
			mockController        *gomock.Controller
			mockTokenSourceSource *oauthTest.MockTokenSourceSource
			mockProvider          *ouraClientTest.MockProvider
			server                *ghttp.Server
			baseClient            *oauthClient.Client
		)

		BeforeEach(func() {
			var err error
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockTokenSourceSource = oauthTest.NewMockTokenSourceSource(mockController)
			mockProvider = ouraClientTest.NewMockProvider(mockController)
			server = ghttp.NewServer()
			baseClient, err = oauthClient.NewWithErrorParser(&client.Config{Address: server.URL()}, mockTokenSourceSource, &ouraClient.ErrorResponseParser{})
			Expect(err).ToNot(HaveOccurred())
			Expect(baseClient).ToNot(BeNil())
		})

		AfterEach(func() {
			server.Close()
		})

		Context("NewWithClient", func() {
			It("returns error if client is missing", func() {
				clnt, err := ouraClient.NewWithClient(nil, mockProvider)
				Expect(clnt).To(BeNil())
				Expect(err).To(MatchError("client is missing"))
			})

			It("returns error if provider is missing", func() {
				clnt, err := ouraClient.NewWithClient(baseClient, nil)
				Expect(clnt).To(BeNil())
				Expect(err).To(MatchError("provider is missing"))
			})

			It("returns successfully", func() {
				clnt, err := ouraClient.NewWithClient(baseClient, mockProvider)
				Expect(clnt).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with client", func() {
			var clnt *ouraClient.Client

			BeforeEach(func() {
				var err error
				clnt, err = ouraClient.NewWithClient(baseClient, mockProvider)
				Expect(err).ToNot(HaveOccurred())
				Expect(clnt).ToNot(BeNil())
			})

			Context("with client headers", func() {
				var clientID string
				var clientSecret string
				var clientHeaders http.Header

				BeforeEach(func() {
					clientID = oauthProviderTest.RandomClientID()
					clientSecret = oauthProviderTest.RandomClientSecret()
					clientHeaders = http.Header{
						ouraClient.HeaderClientID:     []string{clientID},
						ouraClient.HeaderClientSecret: []string{clientSecret},
					}
				})

				Context("ListSubscriptions", func() {
					It("returns error if server returns non-http.StatusOK status code", func() {
						mockProvider.EXPECT().ClientID().Return(clientID)
						mockProvider.EXPECT().ClientSecret().Return(clientSecret)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/v2/webhook/subscription"),
								ghttp.VerifyHeader(clientHeaders),
								ghttp.VerifyBody(nil),
								ghttp.RespondWith(http.StatusInternalServerError, nil),
							),
						)

						subscriptions, err := clnt.ListSubscriptions(ctx)
						Expect(err).To(MatchError(ContainSubstring("unable to list subscriptions; unexpected response status code 500 from GET")))
						Expect(subscriptions).To(BeEmpty())
					})

					It("returns successfully if server returns http.StatusOK status code", func() {
						expectedSubscriptions := ouraTest.RandomSubscriptions()
						mockProvider.EXPECT().ClientID().Return(clientID)
						mockProvider.EXPECT().ClientSecret().Return(clientSecret)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/v2/webhook/subscription"),
								ghttp.VerifyHeader(clientHeaders),
								ghttp.VerifyBody(nil),
								ghttp.RespondWithJSONEncoded(http.StatusOK, expectedSubscriptions),
							),
						)

						subscriptions, err := clnt.ListSubscriptions(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(subscriptions).To(Equal(expectedSubscriptions))
					})
				})

				Context("CreateSubscription", func() {
					var createSubscription *oura.CreateSubscription

					BeforeEach(func() {
						createSubscription = ouraTest.RandomCreateSubscription(test.AllowOptional())
					})

					It("returns error if create is missing", func() {
						subscription, err := clnt.CreateSubscription(ctx, nil)
						Expect(err).To(MatchError("create is missing"))
						Expect(subscription).To(BeNil())
					})

					It("returns error if create is invalid", func() {
						createSubscription.CallbackURL = nil
						subscription, err := clnt.CreateSubscription(ctx, createSubscription)
						Expect(err).To(MatchError(ContainSubstring("create is invalid")))
						Expect(subscription).To(BeNil())
					})

					It("returns error if server returns non-http.StatusOK status code", func() {
						mockProvider.EXPECT().ClientID().Return(clientID)
						mockProvider.EXPECT().ClientSecret().Return(clientSecret)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", "/v2/webhook/subscription"),
								ghttp.VerifyContentType("application/json; charset=utf-8"),
								ghttp.VerifyHeader(clientHeaders),
								ghttp.VerifyJSONRepresenting(createSubscription),
								ghttp.RespondWith(http.StatusInternalServerError, nil),
							),
						)

						subscription, err := clnt.CreateSubscription(ctx, createSubscription)
						Expect(err).To(MatchError(ContainSubstring("unable to create subscription; unexpected response status code 500 from POST")))
						Expect(subscription).To(BeNil())
					})

					It("returns successfully if server returns http.StatusOK status code", func() {
						expectedSubscription := ouraTest.RandomSubscription()
						mockProvider.EXPECT().ClientID().Return(clientID)
						mockProvider.EXPECT().ClientSecret().Return(clientSecret)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", "/v2/webhook/subscription"),
								ghttp.VerifyContentType("application/json; charset=utf-8"),
								ghttp.VerifyHeader(clientHeaders),
								ghttp.VerifyJSONRepresenting(createSubscription),
								ghttp.RespondWithJSONEncoded(http.StatusOK, expectedSubscription),
							),
						)

						subscription, err := clnt.CreateSubscription(ctx, createSubscription)
						Expect(err).ToNot(HaveOccurred())
						Expect(subscription).To(Equal(expectedSubscription))
					})
				})

				Context("UpdateSubscription", func() {
					var (
						id                 string
						updateSubscription *oura.UpdateSubscription
					)

					BeforeEach(func() {
						id = ouraTest.RandomID()
						updateSubscription = ouraTest.RandomUpdateSubscription(test.AllowOptional())
					})

					It("returns error if id is missing", func() {
						subscription, err := clnt.UpdateSubscription(ctx, "", updateSubscription)
						Expect(err).To(MatchError("id is missing"))
						Expect(subscription).To(BeNil())
					})

					It("returns error if update is missing", func() {
						subscription, err := clnt.UpdateSubscription(ctx, id, nil)
						Expect(err).To(MatchError("update is missing"))
						Expect(subscription).To(BeNil())
					})

					It("returns error if update is invalid", func() {
						updateSubscription.CallbackURL = nil
						subscription, err := clnt.UpdateSubscription(ctx, id, updateSubscription)
						Expect(err).To(MatchError(ContainSubstring("update is invalid")))
						Expect(subscription).To(BeNil())
					})

					It("returns error if server returns non-http.StatusOK status code", func() {
						mockProvider.EXPECT().ClientID().Return(clientID)
						mockProvider.EXPECT().ClientSecret().Return(clientSecret)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/webhook/subscription/%s", id)),
								ghttp.VerifyContentType("application/json; charset=utf-8"),
								ghttp.VerifyHeader(clientHeaders),
								ghttp.VerifyJSONRepresenting(updateSubscription),
								ghttp.RespondWith(http.StatusInternalServerError, nil),
							),
						)

						subscription, err := clnt.UpdateSubscription(ctx, id, updateSubscription)
						Expect(err).To(MatchError(ContainSubstring("unable to update subscription; unexpected response status code 500 from PUT")))
						Expect(subscription).To(BeNil())
					})

					It("returns http.StatusNotFound error if server returns http.StatusForbidden status code", func() {
						mockProvider.EXPECT().ClientID().Return(clientID)
						mockProvider.EXPECT().ClientSecret().Return(clientSecret)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/webhook/subscription/%s", id)),
								ghttp.VerifyContentType("application/json; charset=utf-8"),
								ghttp.VerifyHeader(clientHeaders),
								ghttp.VerifyJSONRepresenting(updateSubscription),
								ghttp.RespondWith(http.StatusForbidden, nil),
							),
						)

						subscription, err := clnt.UpdateSubscription(ctx, id, updateSubscription)
						errorsTest.ExpectEqual(err, request.ErrorResourceNotFound())
						Expect(subscription).To(BeNil())
					})

					It("returns successfully if server returns http.StatusOK status code", func() {
						expectedSubscription := ouraTest.RandomSubscription()
						mockProvider.EXPECT().ClientID().Return(clientID)
						mockProvider.EXPECT().ClientSecret().Return(clientSecret)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/webhook/subscription/%s", id)),
								ghttp.VerifyContentType("application/json; charset=utf-8"),
								ghttp.VerifyHeader(clientHeaders),
								ghttp.VerifyJSONRepresenting(updateSubscription),
								ghttp.RespondWithJSONEncoded(http.StatusOK, expectedSubscription),
							),
						)

						subscription, err := clnt.UpdateSubscription(ctx, id, updateSubscription)
						Expect(err).ToNot(HaveOccurred())
						Expect(subscription).To(Equal(expectedSubscription))
					})
				})

				Context("RenewSubscription", func() {
					var id string

					BeforeEach(func() {
						id = ouraTest.RandomID()
					})

					It("returns error if id is missing", func() {
						subscription, err := clnt.RenewSubscription(ctx, "")
						Expect(err).To(MatchError("id is missing"))
						Expect(subscription).To(BeNil())
					})

					It("returns error if server returns non-http.StatusOK status code", func() {
						mockProvider.EXPECT().ClientID().Return(clientID)
						mockProvider.EXPECT().ClientSecret().Return(clientSecret)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/webhook/subscription/renew/%s", id)),
								ghttp.VerifyHeader(clientHeaders),
								ghttp.VerifyBody(nil),
								ghttp.RespondWith(http.StatusInternalServerError, nil),
							),
						)

						subscription, err := clnt.RenewSubscription(ctx, id)
						Expect(err).To(MatchError(ContainSubstring("unable to renew subscription; unexpected response status code 500 from PUT")))
						Expect(subscription).To(BeNil())
					})

					It("returns http.StatusNotFound error if server returns http.StatusForbidden status code", func() {
						mockProvider.EXPECT().ClientID().Return(clientID)
						mockProvider.EXPECT().ClientSecret().Return(clientSecret)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/webhook/subscription/renew/%s", id)),
								ghttp.VerifyHeader(clientHeaders),
								ghttp.VerifyBody(nil),
								ghttp.RespondWith(http.StatusForbidden, nil),
							),
						)

						subscription, err := clnt.RenewSubscription(ctx, id)
						errorsTest.ExpectEqual(err, request.ErrorResourceNotFound())
						Expect(subscription).To(BeNil())
					})

					It("returns successfully if server returns http.StatusOK status code", func() {
						expectedSubscription := ouraTest.RandomSubscription()
						mockProvider.EXPECT().ClientID().Return(clientID)
						mockProvider.EXPECT().ClientSecret().Return(clientSecret)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/webhook/subscription/renew/%s", id)),
								ghttp.VerifyHeader(clientHeaders),
								ghttp.VerifyBody(nil),
								ghttp.RespondWithJSONEncoded(http.StatusOK, expectedSubscription),
							),
						)

						subscription, err := clnt.RenewSubscription(ctx, id)
						Expect(err).ToNot(HaveOccurred())
						Expect(subscription).To(Equal(expectedSubscription))
					})
				})

				Context("DeleteSubscription", func() {
					var id string

					BeforeEach(func() {
						id = ouraTest.RandomID()
					})

					It("returns error if id is missing", func() {
						Expect(clnt.DeleteSubscription(ctx, "")).To(MatchError("id is missing"))
					})

					It("returns error if server returns non-http.StatusOK status code", func() {
						mockProvider.EXPECT().ClientID().Return(clientID)
						mockProvider.EXPECT().ClientSecret().Return(clientSecret)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("DELETE", fmt.Sprintf("/v2/webhook/subscription/%s", id)),
								ghttp.VerifyHeader(clientHeaders),
								ghttp.VerifyBody(nil),
								ghttp.RespondWith(http.StatusInternalServerError, nil),
							),
						)

						Expect(clnt.DeleteSubscription(ctx, id)).To(MatchError(ContainSubstring("unable to delete subscription; unexpected response status code 500 from DELETE")))
					})

					It("returns http.StatusNotFound error if server returns http.StatusForbidden status code", func() {
						mockProvider.EXPECT().ClientID().Return(clientID)
						mockProvider.EXPECT().ClientSecret().Return(clientSecret)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("DELETE", fmt.Sprintf("/v2/webhook/subscription/%s", id)),
								ghttp.VerifyHeader(clientHeaders),
								ghttp.VerifyBody(nil),
								ghttp.RespondWith(http.StatusForbidden, nil),
							),
						)

						errorsTest.ExpectEqual(clnt.DeleteSubscription(ctx, id), request.ErrorResourceNotFound())
					})

					It("returns successfully if server returns http.StatusOK status code", func() {
						expectedSubscription := ouraTest.RandomSubscription()
						mockProvider.EXPECT().ClientID().Return(clientID)
						mockProvider.EXPECT().ClientSecret().Return(clientSecret)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("DELETE", fmt.Sprintf("/v2/webhook/subscription/%s", id)),
								ghttp.VerifyHeader(clientHeaders),
								ghttp.VerifyBody(nil),
								ghttp.RespondWithJSONEncoded(http.StatusOK, expectedSubscription),
							),
						)

						Expect(clnt.DeleteSubscription(ctx, id)).ToNot(HaveOccurred())
					})
				})
			})

			Context("with token source", func() {
				var mockTokenSource *oauthTest.MockTokenSource

				BeforeEach(func() {
					mockTokenSource = oauthTest.NewMockTokenSource(mockController)
				})

				Context("GetPersonalInfo", func() {
					It("returns error if token source is missing", func() {
						personalInfo, err := clnt.GetPersonalInfo(ctx, nil)
						Expect(err).To(MatchError("token source is missing"))
						Expect(personalInfo).To(BeNil())
					})

					It("returns error if server returns non-http.StatusOK status code", func() {
						expectedPersonalInfo := ouraTest.RandomPersonalInfo()
						mockTokenSource.EXPECT().HTTPClient(gomock.Any(), gomock.Any()).Return(http.DefaultClient, nil)
						mockTokenSource.EXPECT().UpdateToken(gomock.Any()).Return(false, nil)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/v2/usercollection/personal_info"),
								ghttp.VerifyHeader(http.Header{}),
								ghttp.VerifyBody(nil),
								ghttp.RespondWithJSONEncoded(http.StatusInternalServerError, expectedPersonalInfo),
							),
						)

						personalInfo, err := clnt.GetPersonalInfo(ctx, mockTokenSource)
						Expect(err).To(MatchError(ContainSubstring("unable to get personal info; unexpected response status code 500 from GET")))
						Expect(personalInfo).To(BeNil())
					})

					It("returns successfully if server returns http.StatusOK status code", func() {
						expectedPersonalInfo := ouraTest.RandomPersonalInfo()
						mockTokenSource.EXPECT().HTTPClient(gomock.Any(), gomock.Any()).Return(http.DefaultClient, nil)
						mockTokenSource.EXPECT().UpdateToken(gomock.Any()).Return(false, nil)
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/v2/usercollection/personal_info"),
								// ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Bearer %s", mockTokenSource.Token().AccessToken)),
								ghttp.VerifyBody(nil),
								ghttp.RespondWithJSONEncoded(http.StatusOK, expectedPersonalInfo),
							),
						)

						personalInfo, err := clnt.GetPersonalInfo(ctx, mockTokenSource)
						Expect(err).ToNot(HaveOccurred())
						Expect(personalInfo).To(Equal(expectedPersonalInfo))
					})
				})

				Context("GetData", func() {
					// TODO: https://tidepool.atlassian.net/browse/BACK-4035
				})

				Context("GetDatum", func() {
					// TODO: https://tidepool.atlassian.net/browse/BACK-4034
				})
			})

			Context("with oauth token", func() {
				var oauthToken *auth.OAuthToken

				BeforeEach(func() {
					oauthToken = authTest.RandomToken()
				})

				Context("RevokeOAuthToken", func() {

					It("returns error if oauth token is missing", func() {
						Expect(clnt.RevokeOAuthToken(ctx, nil)).To(MatchError("oauth token is missing"))
					})

					It("returns error if server returns non-http.StatusOK status code", func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", "/oauth/revoke"),
								ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("%s %s", oauthToken.TokenType, oauthToken.RefreshToken)),
								ghttp.VerifyBody(nil),
								ghttp.RespondWith(http.StatusInternalServerError, nil),
							),
						)

						Expect(clnt.RevokeOAuthToken(ctx, oauthToken)).To(MatchError(ContainSubstring("unable to revoke oauth token; unexpected response status code 500 from POST")))
					})

					It("returns error if server returns non-http.StatusOK status code", func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", "/oauth/revoke"),
								ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("%s %s", oauthToken.TokenType, oauthToken.RefreshToken)),
								ghttp.VerifyBody(nil),
								ghttp.RespondWith(http.StatusOK, nil),
							),
						)

						Expect(clnt.RevokeOAuthToken(ctx, oauthToken)).To(Succeed())
					})
				})
			})
		})
	})
})
