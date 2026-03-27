package provider_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/oura"
	ouraProvider "github.com/tidepool-org/platform/oura/provider"
	ouraProviderTest "github.com/tidepool-org/platform/oura/provider/test"
	ouraWork "github.com/tidepool-org/platform/oura/work"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("provider", func() {
	var lgr *logTest.Logger
	var ctx context.Context
	var mockController *gomock.Controller
	var mockProviderSessionClient *providerSessionTest.MockClient
	var mockDataSourceClient *dataSourceTest.MockClient
	var mockWorkClient *workTest.MockClient
	var dependencies ouraProvider.Dependencies

	BeforeEach(func() {
		lgr = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), lgr)
		mockController = gomock.NewController(GinkgoT())
		mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
		mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
		mockWorkClient = workTest.NewMockClient(mockController)
		dependencies = ouraProvider.Dependencies{
			Config:                *ouraProviderTest.RandomConfig(),
			ProviderSessionClient: mockProviderSessionClient,
			DataSourceClient:      mockDataSourceClient,
			WorkClient:            mockWorkClient,
		}
	})

	Context("Dependencies", func() {
		Context("Validate", func() {
			It("returns an error if provider session client is missing", func() {
				dependencies.ProviderSessionClient = nil
				Expect(dependencies.Validate()).To(MatchError("provider session client is missing"))
			})

			It("returns an error if data source client is missing", func() {
				dependencies.DataSourceClient = nil
				Expect(dependencies.Validate()).To(MatchError("data source client is missing"))
			})

			It("returns an error if work client is missing", func() {
				dependencies.WorkClient = nil
				Expect(dependencies.Validate()).To(MatchError("work client is missing"))
			})

			It("returns successfully", func() {
				Expect(dependencies.Validate()).To(Succeed())
			})
		})
	})

	Context("Provider", func() {
		Context("New", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.ProviderSessionClient = nil
				prvdr, err := ouraProvider.New(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; provider session client is missing"))
				Expect(prvdr).To(BeNil())
			})

			It("returns successfully", func() {
				prvdr, err := ouraProvider.New(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(prvdr).ToNot(BeNil())
			})
		})

		Context("with new provider", func() {
			var prvdr *ouraProvider.Provider

			BeforeEach(func() {
				var err error
				prvdr, err = ouraProvider.New(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(prvdr).ToNot(BeNil())
			})

			Context("OnDelete", func() {
				var providerSession *auth.ProviderSession

				BeforeEach(func() {
					providerSession = authTest.RandomProviderSession()
				})

				It("returns an error if provider session is nil", func() {
					Expect(prvdr.OnDelete(ctx, nil)).To(MatchError("unable to disconnect data source; provider session is missing"))
				})

				It("returns an error if GetFromProviderSession fails", func() {
					testErr := errorsTest.RandomError()
					mockDataSourceClient.EXPECT().GetFromProviderSession(gomock.Any(), providerSession.ID).Return(nil, testErr)
					Expect(prvdr.OnDelete(ctx, providerSession)).To(MatchError("unable to disconnect data source; unable to get data source from provider session; " + testErr.Error()))
				})

				Context("with data source", func() {
					var dataSrc *dataSource.Source
					var expectedDataSrcUpdate *dataSource.Update

					BeforeEach(func() {
						dataSrc = dataSourceTest.RandomSource()
						dataSrc.ProviderSessionID = pointer.FromString(providerSession.ID)
						expectedDataSrcUpdate = &dataSource.Update{State: pointer.FromString(dataSource.StateDisconnected)}
						mockDataSourceClient.EXPECT().GetFromProviderSession(gomock.Any(), providerSession.ID).Return(dataSrc, nil)
					})

					It("returns an error if data source Update fails", func() {
						testErr := errorsTest.RandomError()
						mockDataSourceClient.EXPECT().Update(gomock.Any(), dataSrc.ID, gomock.Nil(), expectedDataSrcUpdate).Return(nil, testErr)
						Expect(prvdr.OnDelete(ctx, providerSession)).To(MatchError("unable to disconnect data source; unable to update data source; " + testErr.Error()))
					})

					Context("with disconnected data source", func() {
						var expectedGroupID string

						BeforeEach(func() {
							expectedGroupID = ouraWork.GroupIDFromProviderSessionID(providerSession.ID)
							mockDataSourceClient.EXPECT().Update(gomock.Any(), dataSrc.ID, gomock.Nil(), expectedDataSrcUpdate).Return(dataSrc, nil)
						})

						AfterEach(func() {
							lgr.AssertDebug("disconnected data source from provider session")
						})

						It("returns an error if DeleteAllByGroupID fails", func() {
							testErr := errorsTest.RandomError()
							mockWorkClient.EXPECT().DeleteAllByGroupID(gomock.Any(), expectedGroupID).Return(0, testErr)
							Expect(prvdr.OnDelete(ctx, providerSession)).To(MatchError("unable to delete work for provider session; unable to delete all work by group id; " + testErr.Error()))
						})

						Context("with deleted work", func() {
							var expectedCount int

							BeforeEach(func() {
								expectedCount = test.RandomIntFromRange(0, 3)
								mockWorkClient.EXPECT().DeleteAllByGroupID(gomock.Any(), expectedGroupID).Return(expectedCount, nil)
							})

							AfterEach(func() {
								lgr.AssertDebug("deleted work for provider session", log.Fields{"count": expectedCount})
							})

							It("returns an error if Create revoke work fails", func() {
								testErr := errorsTest.RandomError()
								mockWorkClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, testErr)
								Expect(prvdr.OnDelete(ctx, providerSession)).To(MatchError("unable to create users revoke work; " + testErr.Error()))
							})

							It("succeeds", func() {
								mockWorkClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(workTest.RandomWork(), nil)
								Expect(prvdr.OnDelete(ctx, providerSession)).To(Succeed())
							})
						})
					})
				})
			})

			Context("AllowUserInitiatedAction", func() {
				var userID string
				var expectedFilter *dataSource.Filter

				BeforeEach(func() {
					userID = userTest.RandomUserID()
					expectedFilter = &dataSource.Filter{
						ProviderType: pointer.FromString(oauth.ProviderType),
						ProviderName: pointer.FromString(oura.ProviderName),
					}
				})

				Context("with authorize action", func() {
					It("returns an error if data source list fails", func() {
						testErr := errorsTest.RandomError()
						mockDataSourceClient.EXPECT().List(gomock.Any(), userID, expectedFilter, page.NewPaginationMinimum()).Return(nil, testErr)
						allowed, err := prvdr.AllowUserInitiatedAction(ctx, userID, oauth.ActionAuthorize)
						Expect(err).To(MatchError("unable to get data sources; " + testErr.Error()))
						Expect(allowed).To(BeFalse())
					})

					It("returns false if user has nil data sources", func() {
						mockDataSourceClient.EXPECT().List(gomock.Any(), userID, expectedFilter, page.NewPaginationMinimum()).Return(nil, nil)
						Expect(prvdr.AllowUserInitiatedAction(ctx, userID, oauth.ActionAuthorize)).To(BeFalse())
					})

					It("returns false if user has empty data sources", func() {
						mockDataSourceClient.EXPECT().List(gomock.Any(), userID, expectedFilter, page.NewPaginationMinimum()).Return(dataSource.SourceArray{}, nil)
						Expect(prvdr.AllowUserInitiatedAction(ctx, userID, oauth.ActionAuthorize)).To(BeFalse())
					})

					It("returns true if user has existing data sources", func() {
						mockDataSourceClient.EXPECT().List(gomock.Any(), userID, expectedFilter, page.NewPaginationMinimum()).Return(dataSourceTest.RandomSourceArray(1, 3, test.AllowOptional()), nil)
						Expect(prvdr.AllowUserInitiatedAction(ctx, userID, oauth.ActionAuthorize)).To(BeTrue())
					})
				})

				Context("with revoke action", func() {
					It("returns true", func() {
						Expect(prvdr.AllowUserInitiatedAction(ctx, userID, oauth.ActionRevoke)).To(BeTrue())
					})
				})
			})

			Context("UserActionAcceptURL", func() {
				It("returns the user action accept URL on authorize", func() {
					Expect(prvdr.UserActionAcceptURL(ctx, userTest.RandomUserID(), oauth.ActionAuthorize)).To(Equal(dependencies.Config.AcceptURL))
				})

				It("returns the user action accept URL on revoke", func() {
					Expect(prvdr.UserActionAcceptURL(ctx, userTest.RandomUserID(), oauth.ActionRevoke)).To(BeNil())
				})
			})

			Context("PartnerURL", func() {
				It("returns the partner URL", func() {
					Expect(prvdr.PartnerURL()).To(Equal(dependencies.Config.PartnerURL))
				})
			})

			Context("PartnerSecret", func() {
				It("returns the partner secret", func() {
					Expect(prvdr.PartnerSecret()).To(Equal(dependencies.Config.PartnerSecret))
				})
			})

			Context("Client", func() {
				It("returns the oura client", func() {
					Expect(prvdr.Client()).ToNot(BeNil())
				})
			})
		})
	})
})
