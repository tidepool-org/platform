package provider_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/data"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	twiistProvider "github.com/tidepool-org/platform/twiist/provider"
	twiistProviderTest "github.com/tidepool-org/platform/twiist/provider/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Provider", func() {
	var userID string
	var providerSessionClientController *gomock.Controller
	var providerSessionClient *twiistProviderTest.MockProviderSessionClient
	var dataSetClientController *gomock.Controller
	var dataSetClient *twiistProviderTest.MockDataSetClient
	var dataSourceController *gomock.Controller
	var dataSourceClient *twiistProviderTest.MockDataSourceClient
	var jwks jwk.Set
	var prvdr *twiistProvider.Provider

	BeforeEach(func() {
		userID = userTest.RandomID()
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

		providerSessionClientController = gomock.NewController(GinkgoT())
		providerSessionClient = twiistProviderTest.NewMockProviderSessionClient(providerSessionClientController)

		dataSetClientController = gomock.NewController(GinkgoT())
		dataSetClient = twiistProviderTest.NewMockDataSetClient(dataSetClientController)

		dataSourceController = gomock.NewController(GinkgoT())
		dataSourceClient = twiistProviderTest.NewMockDataSourceClient(dataSourceController)

		var err error
		jwks, err = jwk.ParseString(twiistProviderTest.JWKSRaw)
		Expect(err).ToNot(HaveOccurred())

		providerDependencies := twiistProvider.ProviderDependencies{
			ConfigReporter:        configReporter,
			ProviderSessionClient: providerSessionClient,
			DataSetClient:         dataSetClient,
			DataSourceClient:      dataSourceClient,
			JWKS:                  jwks,
		}

		prvdr, err = twiistProvider.New(providerDependencies)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		dataSetClientController.Finish()
		dataSourceController.Finish()
	})

	Describe("OnCreate", func() {
		var providerSessionExternalID string
		var dataSourceExternalID string
		var session *auth.ProviderSession

		BeforeEach(func() {
			providerSessionExternalID = twiistProviderTest.RandomTidepoolLinkID()
			dataSourceExternalID = twiistProviderTest.RandomSubjectID()

			idToken, err := twiistProviderTest.GenerateIDToken("tidepool", dataSourceExternalID, providerSessionExternalID, jwks)
			Expect(err).ToNot(HaveOccurred())

			token := auth.NewOAuthToken()
			token.IDToken = pointer.FromString(idToken)

			session = &auth.ProviderSession{
				ID:          "session-id",
				UserID:      userID,
				OAuthToken:  token,
				Type:        "oauth",
				Name:        "twiist",
				CreatedTime: time.Now(),
			}
		})

		It("creates new data source and new data set for new connections", func() {
			ctx := log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			dataSourceClient.EXPECT().
				List(ctx, session.UserID, gomock.Any(), gomock.Any()).
				Return(nil, nil)

			dataSrc := &dataSource.Source{
				ID:                 pointer.FromString(dataSourceTest.RandomID()),
				UserID:             &userID,
				ProviderType:       pointer.FromString("oauth"),
				ProviderName:       pointer.FromString("twiist"),
				ProviderExternalID: pointer.FromString(dataSourceExternalID),
				State:              pointer.FromString(dataSource.StateDisconnected),
			}
			dataSourceClient.EXPECT().
				Create(ctx, gomock.Eq(session.UserID), gomock.Any()).
				Return(dataSrc, nil)

			dataSetID := dataTest.RandomID()
			dataSet := data.DataSet{ID: pointer.FromString(dataSetID)}
			dataSetClient.EXPECT().
				CreateUserDataSet(ctx, gomock.Eq(session.UserID), gomock.Any()).
				Do(func(_ context.Context, _ string, create *data.DataSetCreate) {
					Expect(create).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"Client": PointTo(MatchFields(IgnoreExtras, Fields{
							"Name":    PointTo(Equal("com.sequelmedtech.tidepool-service")),
							"Version": PointTo(Equal("3.0.0")),
						})),
						"DataSetType": PointTo(Equal("continuous")),
						"Deduplicator": PointTo(MatchFields(IgnoreExtras, Fields{
							"Name": PointTo(Equal("org.tidepool.deduplicator.dataset.delete.origin.older")),
						})),
						"DeviceManufacturers": PointTo(ConsistOf(Equal("Sequel"))),
						"DeviceTags":          PointTo(ConsistOf("cgm", "bgm", "insulin-pump")),
						"Time":                PointTo(BeTemporally("<=", time.Now())),
						"TimeProcessing":      PointTo(Equal("none")),
					})))
				}).
				Return(&dataSet, nil)

			providerSessionClient.EXPECT().
				UpdateProviderSession(ctx, gomock.Eq(session.ID), gomock.Any()).
				Do(func(_ context.Context, _ string, update *auth.ProviderSessionUpdate) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"OAuthToken": PointTo(Equal(*session.OAuthToken)),
						"ExternalID": PointTo(Equal(providerSessionExternalID)),
					})))
				}).
				Return(session, nil)

			dataSourceClient.EXPECT().
				Update(ctx, gomock.Eq(*dataSrc.ID), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, id string, condition *request.Condition, update *dataSource.Update) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"DataSetIDs":         PointTo(ConsistOf(dataSetID)),
						"ProviderExternalID": PointTo(Equal(dataSourceExternalID)),
						"ProviderSessionID":  PointTo(Equal(session.ID)),
						"State":              PointTo(Equal("connected")),
					})))
				}).
				Return(dataSrc, nil)

			Expect(prvdr.OnCreate(ctx, session)).To(Succeed())
		})

		It("reuses data source and data set for existing connections", func() {
			ctx := log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			dataSetID := dataTest.RandomID()
			dataSet := data.DataSet{ID: pointer.FromString(dataSetID)}

			dataSrc := &dataSource.Source{
				ID:                 pointer.FromString(dataSourceTest.RandomID()),
				UserID:             &userID,
				ProviderType:       pointer.FromString("oauth"),
				ProviderName:       pointer.FromString("twiist"),
				ProviderSessionID:  pointer.FromString(session.ID),
				ProviderExternalID: pointer.FromString(dataSourceExternalID),
				State:              pointer.FromString(dataSource.StateConnected),
				DataSetIDs:         pointer.FromStringArray([]string{dataSetID}),
			}

			dataSourceClient.EXPECT().
				List(ctx, session.UserID, gomock.Any(), gomock.Any()).
				Return(dataSource.SourceArray{dataSrc}, nil)
			providerSessionClient.EXPECT().
				DeleteProviderSession(ctx, gomock.Eq(session.ID)).
				Return(nil)
			dataSourceClient.EXPECT().
				Update(ctx, gomock.Eq(*dataSrc.ID), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, id string, condition *request.Condition, update *dataSource.Update) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"State": PointTo(Equal("disconnected")),
					})))
				}).
				Return(dataSrc, nil)
			dataSetClient.EXPECT().
				GetDataSet(ctx, gomock.Eq(dataSetID)).
				Return(&dataSet, nil)

			providerSessionClient.EXPECT().
				UpdateProviderSession(ctx, gomock.Eq(session.ID), gomock.Any()).
				Do(func(_ context.Context, _ string, update *auth.ProviderSessionUpdate) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"OAuthToken": PointTo(Equal(*session.OAuthToken)),
						"ExternalID": PointTo(Equal(providerSessionExternalID)),
					})))
				}).
				Return(session, nil)

			dataSourceClient.EXPECT().
				Update(ctx, gomock.Eq(*dataSrc.ID), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, id string, condition *request.Condition, update *dataSource.Update) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"DataSetIDs":         PointTo(ConsistOf(dataSetID)),
						"ProviderExternalID": PointTo(Equal(dataSourceExternalID)),
						"ProviderSessionID":  PointTo(Equal(session.ID)),
						"State":              PointTo(Equal("connected")),
					})))
				}).
				Return(dataSrc, nil)

			Expect(prvdr.OnCreate(ctx, session)).To(Succeed())
		})

		It("reuses data source and creates data set if it doesn't exist", func() {
			ctx := log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			missingDataSetID := dataTest.RandomID()
			dataSetID := dataTest.RandomID()
			dataSet := data.DataSet{ID: pointer.FromString(dataSetID)}

			dataSrc := &dataSource.Source{
				ID:                 pointer.FromString(dataSourceTest.RandomID()),
				UserID:             &userID,
				ProviderType:       pointer.FromString("oauth"),
				ProviderName:       pointer.FromString("twiist"),
				ProviderSessionID:  pointer.FromString(session.ID),
				ProviderExternalID: pointer.FromString(dataSourceExternalID),
				State:              pointer.FromString(dataSource.StateConnected),
				DataSetIDs:         pointer.FromStringArray([]string{missingDataSetID}),
			}

			dataSourceClient.EXPECT().
				List(ctx, session.UserID, gomock.Any(), gomock.Any()).
				Return(dataSource.SourceArray{dataSrc}, nil)
			providerSessionClient.EXPECT().
				DeleteProviderSession(ctx, gomock.Eq(session.ID)).
				Return(nil)
			dataSourceClient.EXPECT().
				Update(ctx, gomock.Eq(*dataSrc.ID), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, id string, condition *request.Condition, update *dataSource.Update) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"State": PointTo(Equal("disconnected")),
					})))
				}).
				Return(dataSrc, nil)
			dataSetClient.EXPECT().
				GetDataSet(ctx, gomock.Eq(missingDataSetID)).
				Return(nil, nil)
			dataSetClient.EXPECT().
				CreateUserDataSet(ctx, gomock.Eq(session.UserID), gomock.Any()).
				Return(&dataSet, nil)

			providerSessionClient.EXPECT().
				UpdateProviderSession(ctx, gomock.Eq(session.ID), gomock.Any()).
				Do(func(_ context.Context, _ string, update *auth.ProviderSessionUpdate) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"OAuthToken": PointTo(Equal(*session.OAuthToken)),
						"ExternalID": PointTo(Equal(providerSessionExternalID)),
					})))
				}).
				Return(session, nil)

			dataSourceClient.EXPECT().
				Update(ctx, gomock.Eq(*dataSrc.ID), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, id string, condition *request.Condition, update *dataSource.Update) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"DataSetIDs":         PointTo(ConsistOf(missingDataSetID, dataSetID)),
						"ProviderExternalID": PointTo(Equal(dataSourceExternalID)),
						"ProviderSessionID":  PointTo(Equal(session.ID)),
						"State":              PointTo(Equal("connected")),
					})))
				}).
				Return(dataSrc, nil)

			Expect(prvdr.OnCreate(ctx, session)).To(Succeed())
		})
	})

	Describe("OnDelete", func() {
		var session *auth.ProviderSession

		BeforeEach(func() {
			providerSessionExternalID := twiistProviderTest.RandomTidepoolLinkID()
			dataSourceExternalID := twiistProviderTest.RandomSubjectID()

			idToken, err := twiistProviderTest.GenerateIDToken("tidepool", dataSourceExternalID, providerSessionExternalID, jwks)
			Expect(err).ToNot(HaveOccurred())

			token := auth.NewOAuthToken()
			token.IDToken = pointer.FromString(idToken)

			session = &auth.ProviderSession{
				ID:          "session-id",
				UserID:      userID,
				OAuthToken:  token,
				Type:        "oauth",
				Name:        "twiist",
				ExternalID:  pointer.FromString(providerSessionExternalID),
				CreatedTime: time.Now(),
			}
		})

		It("disconnects all data sources for the provider", func() {
			ctx := log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			dataSrc := dataSourceTest.RandomSource()
			dataSrc.UserID = &userID
			dataSrcs := dataSource.SourceArray{dataSourceTest.RandomSource(), dataSourceTest.RandomSource()}
			for _, s := range dataSrcs {
				s.ProviderType = pointer.FromString("oauth")
				s.ProviderName = pointer.FromString("twiist")
				s.State = pointer.FromString(dataSource.StateConnected)
			}

			dataSourceClient.EXPECT().
				List(ctx, session.UserID, gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, userID string, filter *dataSource.Filter, _ *page.Pagination) {
					Expect(filter).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"ProviderType":      PointTo(ConsistOf("oauth")),
						"ProviderName":      PointTo(ConsistOf("twiist")),
						"ProviderSessionID": PointTo(ConsistOf(session.ID)),
					})))
				}).
				Return(dataSrcs, nil)

			for _, s := range dataSrcs {
				dataSourceClient.EXPECT().
					Update(gomock.Any(), gomock.Eq(*s.ID), gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, id string, condition *request.Condition, update *dataSource.Update) {
						Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
							"State": PointTo(Equal("disconnected")),
						})))
					}).
					Return(nil, nil)
			}

			Expect(prvdr.OnDelete(ctx, session)).To(Succeed())
		})
	})
})
