package provider_test

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/page"

	"github.com/lestrrat-go/jwx/v2/jwk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/data"
	dataClientTest "github.com/tidepool-org/platform/data/client/test"
	"github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	logInternal "github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/twiist/provider"
	providerTest "github.com/tidepool-org/platform/twiist/provider/test"
)

var _ = Describe("Provider", func() {
	var prvdr *provider.Provider
	var jwks jwk.Set
	var userID string
	var dataClient *dataClientTest.MockClient
	var dataClientCtrl *gomock.Controller
	var dataSourceClient *dataSourceTest.MockClient
	var dataSourceCtrl *gomock.Controller

	BeforeEach(func() {
		userID = "1234567890"
		twiistConfig := map[string]interface{}{
			"client_id":     "tidepool",
			"client_secret": "super_secret",
			"authorize_url": "https://twiist.com/authorize",
			"token_url":     "https://twiist.com/token",
			"redirect_url":  "https://test.tidepool.org/twiist/callback",
			"state_salt":    "salt",
		}

		configReporter := configTest.NewReporter()
		configReporter.Config = map[string]interface{}{
			"twiist": twiistConfig,
		}

		dataClientCtrl = gomock.NewController(GinkgoT())
		dataClient = dataClientTest.NewMockClient(dataClientCtrl)

		dataSourceCtrl = gomock.NewController(GinkgoT())
		dataSourceClient = dataSourceTest.NewMockClient(dataSourceCtrl)

		var err error
		jwks, err = jwk.ParseString(providerTest.JWKSRaw)
		Expect(err).ToNot(HaveOccurred())

		prvdr, err = provider.New(configReporter, dataClient, dataSourceClient, jwks)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		dataClientCtrl.Finish()
		dataSourceCtrl.Finish()
	})

	Describe("BeforeCreate", func() {
		It("returns an error when create is missing", func() {
			err := prvdr.BeforeCreate(context.Background(), userID, nil)
			Expect(err).To(MatchError("create is missing"))
		})

		It("returns an error when the id token is missing", func() {
			create := &auth.ProviderSessionCreate{}
			err := prvdr.BeforeCreate(context.Background(), userID, create)
			Expect(err).To(MatchError("unable to get claims from id token; oauth token is missing"))
		})

		It("returns an error when the oauth token is missing", func() {
			create := &auth.ProviderSessionCreate{OAuthToken: auth.NewOAuthToken()}
			err := prvdr.BeforeCreate(context.Background(), userID, create)
			Expect(err).To(MatchError("unable to get claims from id token; id token is missing"))
		})

		It("returns an error when the id token is invalid", func() {
			token := auth.NewOAuthToken()
			token.IDToken = pointer.FromString("invalid")

			create := &auth.ProviderSessionCreate{OAuthToken: token}
			err := prvdr.BeforeCreate(context.Background(), userID, create)
			Expect(err).To(MatchError("unable to get claims from id token; unable to parse id_token; unable to verify id token; failed to parse jws: invalid compact serialization format: invalid number of segments"))
		})

		Context("with id token", func() {
			var tidepoolLinkID string
			var idToken string

			BeforeEach(func() {
				var err error
				tidepoolLinkID = providerTest.RandomTidepoolLinkID()
				idToken, err = providerTest.GenerateIDToken(providerTest.RandomSubjectID(), tidepoolLinkID, jwks)
				Expect(err).ToNot(HaveOccurred())
			})

			It("sets the external id of the session create to the tidepool link id from the id token", func() {
				token := auth.NewOAuthToken()
				token.IDToken = pointer.FromString(idToken)

				create := &auth.ProviderSessionCreate{OAuthToken: token}
				Expect(prvdr.BeforeCreate(context.Background(), "1234567890", create)).To(Succeed())
				Expect(create.ExternalID).To(PointTo(Equal(tidepoolLinkID)))
			})

			It("returns an error if it cannot verify the signature of the idToken", func() {
				// Remove one character from the token so the signature is invalid
				invalidToken := idToken[:len(idToken)-1]
				token := auth.NewOAuthToken()
				token.IDToken = pointer.FromString(invalidToken)

				create := &auth.ProviderSessionCreate{OAuthToken: token}
				err := prvdr.BeforeCreate(context.Background(), userID, create)
				Expect(err).To(MatchError("unable to get claims from id token; unable to parse id_token; unable to verify id token; could not verify message using any of the signatures or keys"))
			})
		})

	})

	Describe("OnCreate", func() {
		var session *auth.ProviderSession
		var subjectID string

		BeforeEach(func() {
			subjectID = providerTest.RandomSubjectID()
			externalID := providerTest.RandomTidepoolLinkID()
			idToken, err := providerTest.GenerateIDToken(subjectID, externalID, jwks)
			Expect(err).ToNot(HaveOccurred())

			token := auth.NewOAuthToken()
			token.IDToken = pointer.FromString(idToken)

			session = &auth.ProviderSession{
				ID:          "session-id",
				UserID:      userID,
				OAuthToken:  token,
				Type:        "oauth",
				Name:        "twiist",
				ExternalID:  pointer.FromString(externalID),
				CreatedTime: time.Now(),
			}
		})

		It("creates new data source and new data set for new connections", func() {
			ctx := logInternal.NewContextWithLogger(context.Background(), null.NewLogger())
			dataSourceClient.EXPECT().
				List(ctx, gomock.Eq(session.UserID), gomock.Any(), gomock.Any()).
				Return(nil, nil)

			dataSource := dataSourceTest.RandomSource()
			dataSource.UserID = &userID
			dataSource.DataSetIDs = nil
			dataSourceClient.EXPECT().
				Create(ctx, gomock.Eq(session.UserID), gomock.Any()).
				Return(dataSource, nil)

			dataSetID := dataTest.RandomID()
			dataSet := data.DataSet{ID: pointer.FromString(dataSetID)}
			dataClient.EXPECT().
				CreateUserDataSet(ctx, gomock.Eq(session.UserID), gomock.Any()).
				Return(&dataSet, nil)

			dataSourceClient.EXPECT().
				Update(ctx, gomock.Eq(*dataSource.ID), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, id string, condition *request.Condition, update *source.Update) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"DataSetIDs":         PointTo(ConsistOf(dataSetID)),
						"ProviderExternalID": Equal(session.ExternalID),
						"ProviderSessionID":  PointTo(Equal(session.ID)),
						"State":              PointTo(Equal("connected")),
						"Metadata":           HaveKeyWithValue("externalSubjectID", subjectID),
					})))
				}).
				Return(dataSource, nil)

			Expect(prvdr.OnCreate(ctx, userID, session)).To(Succeed())
		})

		It("reuses data source and data set for existing connections", func() {
			ctx := logInternal.NewContextWithLogger(context.Background(), null.NewLogger())
			dataSetID := dataTest.RandomID()
			dataSet := data.DataSet{ID: pointer.FromString(dataSetID)}
			dataSource := dataSourceTest.RandomSource()
			dataSource.UserID = &userID
			dataSource.DataSetIDs = pointer.FromStringArray([]string{dataSetID})
			dataSource.State = pointer.FromString(source.StateConnected)

			dataSourceClient.EXPECT().
				List(ctx, gomock.Eq(session.UserID), gomock.Any(), gomock.Any()).
				Return(source.SourceArray{dataSource}, nil)
			dataSourceClient.EXPECT().
				Update(ctx, gomock.Eq(*dataSource.ID), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, id string, condition *request.Condition, update *source.Update) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"State": PointTo(Equal("disconnected")),
					})))
				}).
				Return(dataSource, nil)
			dataClient.EXPECT().
				GetDataSet(ctx, gomock.Eq(dataSetID)).
				Return(&dataSet, nil)

			dataSourceClient.EXPECT().
				Update(ctx, gomock.Eq(*dataSource.ID), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, id string, condition *request.Condition, update *source.Update) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"DataSetIDs":         PointTo(ConsistOf(dataSetID)),
						"ProviderExternalID": Equal(session.ExternalID),
						"ProviderSessionID":  PointTo(Equal(session.ID)),
						"State":              PointTo(Equal("connected")),
						"Metadata":           HaveKeyWithValue("externalSubjectID", subjectID),
					})))
				}).
				Return(dataSource, nil)

			Expect(prvdr.OnCreate(ctx, userID, session)).To(Succeed())
		})

		It("reuses data source and creates data set if it doesn't exist", func() {
			ctx := logInternal.NewContextWithLogger(context.Background(), null.NewLogger())
			missingDataSetID := dataTest.RandomID()
			dataSetID := dataTest.RandomID()
			dataSet := data.DataSet{ID: pointer.FromString(dataSetID)}
			dataSource := dataSourceTest.RandomSource()
			dataSource.UserID = &userID
			dataSource.DataSetIDs = pointer.FromStringArray([]string{missingDataSetID})
			dataSource.State = pointer.FromString(source.StateConnected)

			dataSourceClient.EXPECT().
				List(ctx, gomock.Eq(session.UserID), gomock.Any(), gomock.Any()).
				Return(source.SourceArray{dataSource}, nil)
			dataSourceClient.EXPECT().
				Update(ctx, gomock.Eq(*dataSource.ID), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, id string, condition *request.Condition, update *source.Update) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"State": PointTo(Equal("disconnected")),
					})))
				}).
				Return(dataSource, nil)
			dataClient.EXPECT().
				GetDataSet(ctx, gomock.Eq(missingDataSetID)).
				Return(nil, nil)
			dataClient.EXPECT().
				CreateUserDataSet(ctx, gomock.Eq(session.UserID), gomock.Any()).
				Return(&dataSet, nil)

			dataSourceClient.EXPECT().
				Update(ctx, gomock.Eq(*dataSource.ID), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, id string, condition *request.Condition, update *source.Update) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"DataSetIDs":         PointTo(ConsistOf(missingDataSetID, dataSetID)),
						"ProviderExternalID": Equal(session.ExternalID),
						"ProviderSessionID":  PointTo(Equal(session.ID)),
						"State":              PointTo(Equal("connected")),
						"Metadata":           HaveKeyWithValue("externalSubjectID", subjectID),
					})))
				}).
				Return(dataSource, nil)

			Expect(prvdr.OnCreate(ctx, userID, session)).To(Succeed())
		})
	})

	Describe("OnDelete", func() {
		var session *auth.ProviderSession
		var subjectID string

		BeforeEach(func() {
			subjectID = providerTest.RandomSubjectID()
			externalID := providerTest.RandomTidepoolLinkID()
			idToken, err := providerTest.GenerateIDToken(subjectID, externalID, jwks)
			Expect(err).ToNot(HaveOccurred())

			token := auth.NewOAuthToken()
			token.IDToken = pointer.FromString(idToken)

			session = &auth.ProviderSession{
				ID:          "session-id",
				UserID:      userID,
				OAuthToken:  token,
				Type:        "oauth",
				Name:        "twiist",
				ExternalID:  pointer.FromString(externalID),
				CreatedTime: time.Now(),
			}
		})

		It("disconnects all data source for the provider", func() {
			ctx := logInternal.NewContextWithLogger(context.Background(), null.NewLogger())
			dataSource := dataSourceTest.RandomSource()
			dataSource.UserID = &userID
			dataSources := source.SourceArray{dataSourceTest.RandomSource(), dataSourceTest.RandomSource()}
			for _, s := range dataSources {
				s.ProviderType = pointer.FromString("oauth")
				s.ProviderName = pointer.FromString("twiist")
				s.State = pointer.FromString(source.StateConnected)
			}

			dataSourceClient.EXPECT().
				List(ctx, gomock.Eq(session.UserID), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, _ string, filter *source.Filter, _ *page.Pagination) {
					Expect(filter).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"ProviderType": PointTo(ConsistOf("oauth")),
						"ProviderName": PointTo(ConsistOf("twiist")),
						"State":        PointTo(ConsistOf("connected")),
					})))
				}).
				Return(dataSources, nil)

			for _, s := range dataSources {
				dataSourceClient.EXPECT().
					Update(ctx, gomock.Eq(*s.ID), gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, id string, condition *request.Condition, update *source.Update) {
						Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
							"State": PointTo(Equal("disconnected")),
						})))
					}).
					Return(nil, nil)
			}

			Expect(prvdr.OnDelete(ctx, userID, session)).To(Succeed())
		})
	})
})
