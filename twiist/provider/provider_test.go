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
	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/data"
	dataSetTest "github.com/tidepool-org/platform/data/set/test"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	twiistProvider "github.com/tidepool-org/platform/twiist/provider"
	twiistProviderTest "github.com/tidepool-org/platform/twiist/provider/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Provider", func() {
	var userID string
	var mockController *gomock.Controller
	var providerSessionClient *providerSessionTest.MockClient
	var dataSetClient *dataSetTest.MockClient
	var dataSourceClient *dataSourceTest.MockClient
	var jwks jwk.Set
	var prvdr *twiistProvider.Provider

	BeforeEach(func() {
		userID = userTest.RandomUserID()
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

		mockController = gomock.NewController(GinkgoT())

		providerSessionClient = providerSessionTest.NewMockClient(mockController)
		dataSetClient = dataSetTest.NewMockClient(mockController)
		dataSourceClient = dataSourceTest.NewMockClient(mockController)

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
				ID:                 dataSourceTest.RandomDataSourceID(),
				UserID:             userID,
				ProviderType:       "oauth",
				ProviderName:       "twiist",
				ProviderExternalID: pointer.FromString(dataSourceExternalID),
				State:              dataSource.StateDisconnected,
			}
			dataSourceClient.EXPECT().
				Create(ctx, gomock.Eq(session.UserID), gomock.Any()).
				Return(dataSrc, nil)

			dataSetID := dataTest.RandomDatumID()
			dataSet := data.DataSet{ID: pointer.FromString(dataSetID)}
			dataSetClient.EXPECT().
				CreateUserDataSet(ctx, gomock.Eq(session.UserID), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, create *data.DataSetCreate) (*data.DataSet, error) {
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
					return &dataSet, nil
				})

			providerSessionClient.EXPECT().
				UpdateProviderSession(ctx, gomock.Eq(session.ID), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"OAuthToken": PointTo(Equal(*session.OAuthToken)),
						"ExternalID": PointTo(Equal(providerSessionExternalID)),
					})))
					return session, nil
				})

			dataSourceClient.EXPECT().
				Update(ctx, gomock.Eq(dataSrc.ID), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"DataSetID":          PointTo(Equal(dataSetID)),
						"ProviderExternalID": PointTo(Equal(dataSourceExternalID)),
						"ProviderSessionID":  PointTo(Equal(session.ID)),
						"State":              PointTo(Equal("connected")),
					})))
					return dataSrc, nil
				})

			Expect(prvdr.OnCreate(ctx, session)).To(Succeed())
		})

		It("reuses data source and data set for existing connections", func() {
			ctx := log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			dataSetID := dataTest.RandomDatumID()

			dataSrc := &dataSource.Source{
				ID:                 dataSourceTest.RandomDataSourceID(),
				UserID:             userID,
				ProviderType:       "oauth",
				ProviderName:       "twiist",
				ProviderSessionID:  pointer.FromString(session.ID),
				ProviderExternalID: pointer.FromString(dataSourceExternalID),
				State:              dataSource.StateConnected,
				DataSetID:          pointer.FromString(dataSetID),
			}

			dataSourceClient.EXPECT().
				List(ctx, session.UserID, gomock.Any(), gomock.Any()).
				Return(dataSource.SourceArray{dataSrc}, nil)
			providerSessionClient.EXPECT().
				DeleteProviderSession(ctx, gomock.Eq(session.ID)).
				Return(nil)
			dataSourceClient.EXPECT().
				Update(ctx, gomock.Eq(dataSrc.ID), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"State": PointTo(Equal("disconnected")),
					})))
					return dataSrc, nil
				})

			providerSessionClient.EXPECT().
				UpdateProviderSession(ctx, gomock.Eq(session.ID), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"OAuthToken": PointTo(Equal(*session.OAuthToken)),
						"ExternalID": PointTo(Equal(providerSessionExternalID)),
					})))
					return session, nil
				})

			dataSourceClient.EXPECT().
				Update(ctx, gomock.Eq(dataSrc.ID), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"DataSetID":          PointTo(Equal(dataSetID)),
						"ProviderExternalID": PointTo(Equal(dataSourceExternalID)),
						"ProviderSessionID":  PointTo(Equal(session.ID)),
						"State":              PointTo(Equal("connected")),
					})))
					return dataSrc, nil
				})

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

		It("disconnects the data source for the provider", func() {
			ctx := log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			dataSrc := dataSourceTest.RandomSource()
			dataSrc.UserID = userID

			dataSourceClient.EXPECT().
				GetFromProviderSession(ctx, session.ID).
				Return(dataSrc, nil)

			dataSourceClient.EXPECT().
				Update(gomock.Any(), gomock.Eq(dataSrc.ID), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
					Expect(update).To(PointTo(MatchFields(IgnoreExtras, Fields{
						"State": PointTo(Equal("disconnected")),
					})))
					return nil, nil
				})

			Expect(prvdr.OnDelete(ctx, session)).To(Succeed())
		})
	})

	Describe("AllowUserInitiatedAction", func() {
		var ctx context.Context

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
		})

		It("allows authorize action for any user", func() {
			Expect(prvdr.AllowUserInitiatedAction(ctx, userID, oauth.ActionAuthorize)).To(BeTrue())
		})

		It("disallows authorize action for any user", func() {
			Expect(prvdr.AllowUserInitiatedAction(ctx, userID, oauth.ActionRevoke)).To(BeFalse())
		})
	})
})
