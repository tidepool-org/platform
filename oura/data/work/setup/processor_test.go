package setup_test

import (
	"context"
	"time"

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
	ouraDataWorkHistoric "github.com/tidepool-org/platform/oura/data/work/historic"
	ouraDataWorkSetup "github.com/tidepool-org/platform/oura/data/work/setup"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/times"
	userTest "github.com/tidepool-org/platform/user/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("processor", func() {
	It("FailingRetryDuration is expected", func() {
		Expect(ouraDataWorkSetup.FailingRetryDuration).To(Equal(time.Minute))
	})

	It("FailingRetryDurationJitter is expected", func() {
		Expect(ouraDataWorkSetup.FailingRetryDurationJitter).To(Equal(5 * time.Second))
	})

	Context("with dependencies", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockProviderSessionClient *providerSessionTest.MockClient
		var mockDataSourceClient *dataSourceTest.MockClient
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraDataWorkSetup.Dependencies

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraDataWorkSetup.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				ProviderSessionClient: mockProviderSessionClient,
				DataSourceClient:      mockDataSourceClient,
				OuraClient:            mockOuraClient,
			}
		})

		Context("NewProcessor", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.WorkClient = nil
				processor, err := ouraDataWorkSetup.NewProcessor(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns successfully", func() {
				processor, err := ouraDataWorkSetup.NewProcessor(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})

			Context("with processor", func() {
				var userID string
				var ouraUserID string
				var providerSessionID string
				var wrk *work.Work
				var mockProcessingUpdater *workTest.MockProcessingUpdater
				var processor *ouraDataWorkSetup.Processor

				BeforeEach(func() {
					userID = userTest.RandomUserID()
					ouraUserID = ouraTest.RandomUserID()
					providerSessionID = authTest.RandomProviderSessionID()
					wrkCreate, err := ouraDataWorkSetup.NewWorkCreate(providerSessionID)
					Expect(err).ToNot(HaveOccurred())
					Expect(wrkCreate).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(wrkCreate, work.StateProcessing)
					mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)
					processor, err = ouraDataWorkSetup.NewProcessor(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processor).ToNot(BeNil())
				})

				Context("Process", func() {
					It("returns failing process result if unable to fetch provider session from work", func() {
						testErr := errorsTest.RandomError()
						mockProviderSessionClient.EXPECT().GetProviderSession(gomock.Any(), providerSessionID).Return(nil, testErr)
						Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
					})

					Context("with provider session", func() {
						var providerSession *auth.ProviderSession

						BeforeEach(func() {
							providerSession = &auth.ProviderSession{
								ID:          providerSessionID,
								UserID:      userID,
								Type:        oauth.ProviderType,
								Name:        oura.ProviderName,
								OAuthToken:  authTest.RandomToken(),
								CreatedTime: time.Now(),
							}
							mockProviderSessionClient.EXPECT().GetProviderSession(gomock.Any(), providerSessionID).Return(providerSession, nil)
						})

						It("returns failing process result if unable to fetch data source from provider session", func() {
							testErr := errorsTest.RandomError()
							mockDataSourceClient.EXPECT().GetFromProviderSession(gomock.Any(), providerSessionID).Return(nil, testErr)
							Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
						})

						Context("with data source", func() {
							var dataSrcID string
							var dataSrc *dataSource.Source

							BeforeEach(func() {
								dataSrcID = dataSourceTest.RandomDataSourceID()
								dataSrc = &dataSource.Source{
									ID:                dataSrcID,
									UserID:            userID,
									ProviderType:      oauth.ProviderType,
									ProviderName:      oura.ProviderName,
									ProviderSessionID: pointer.FromString(providerSessionID),
									State:             dataSource.StateConnected,
									CreatedTime:       time.Now(),
									Revision:          test.RandomInt(),
								}
								mockDataSourceClient.EXPECT().GetFromProviderSession(gomock.Any(), providerSessionID).Return(dataSrc, nil)
							})

							assertWorkCreate := func() {
								Context("with data historic work create", func() {
									var dataHistoricWorkCreate *work.Create
									var dataHistoricWork *work.Work

									BeforeEach(func() {
										var err error
										dataHistoricWorkCreate, err = ouraDataWorkHistoric.NewWorkCreate(providerSessionID, times.TimeRange{From: dataSrc.LatestDataTime})
										Expect(err).ToNot(HaveOccurred())
										Expect(dataHistoricWorkCreate).ToNot(BeNil())
										dataHistoricWork = workTest.NewWorkFromCreateWithState(dataHistoricWorkCreate, work.StatePending)
									})

									It("returned failing process result if unable to create data historic work", func() {
										testErr := errorsTest.RandomError()
										mockWorkClient.EXPECT().
											Create(gomock.Any(), gomock.Any()).
											Do(func(ctx context.Context, create *work.Create) {
												Expect(create).To(Equal(dataHistoricWorkCreate))
											}).
											Return(nil, testErr)
										Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
									})

									It("returns delete process result when successful", func() {
										mockWorkClient.EXPECT().
											Create(gomock.Any(), gomock.Any()).
											Do(func(ctx context.Context, create *work.Create) {
												Expect(create).To(Equal(dataHistoricWorkCreate))
											}).
											Return(dataHistoricWork, nil)
										Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchDeleteProcessResult())
									})
								})
							}

							Context("with provider session external id and data source provider external id", func() {
								BeforeEach(func() {
									providerSession.ExternalID = pointer.FromString(ouraUserID)
									dataSrc.ProviderExternalID = pointer.FromString(ouraUserID)
								})

								assertWorkCreate()
							})

							It("returns failing process result if oura client returns an error when getting personal info", func() {
								testErr := errorsTest.RandomError()
								mockOuraClient.EXPECT().GetPersonalInfo(gomock.Any(), gomock.Any()).Return(nil, testErr)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
							})

							Context("with personal info", func() {
								var personalInfo *oura.PersonalInfo
								var expectedDataSourceFilter *dataSource.Filter

								BeforeEach(func() {
									personalInfo = ouraTest.RandomPersonalInfo()
									personalInfo.ID = pointer.FromString(ouraUserID)
									mockOuraClient.EXPECT().GetPersonalInfo(gomock.Any(), gomock.Any()).Return(personalInfo, nil)
									expectedDataSourceFilter = &dataSource.Filter{
										ProviderType:       pointer.FromString(oauth.ProviderType),
										ProviderName:       pointer.FromString(oura.ProviderName),
										ProviderExternalID: pointer.FromString(ouraUserID),
									}
								})

								It("returns failing process result if data source list returns an error", func() {
									testErr := errorsTest.RandomError()
									mockDataSourceClient.EXPECT().List(gomock.Any(), userID, expectedDataSourceFilter, gomock.Any()).Return(nil, testErr)
									Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
								})

								assertProviderSessionUpdateAndWorkCreate := func() {
									Context("with provider session update", func() {
										var expectedProviderSessionUpdate *auth.ProviderSessionUpdate

										BeforeEach(func() {
											expectedProviderSessionUpdate = &auth.ProviderSessionUpdate{
												OAuthToken: providerSession.OAuthToken,
												ExternalID: pointer.FromString(ouraUserID),
											}
										})

										It("returns failing process result if unable to update provider session", func() {
											testErr := errorsTest.RandomError()
											mockProviderSessionClient.EXPECT().UpdateProviderSession(gomock.Any(), providerSessionID, expectedProviderSessionUpdate).Return(nil, testErr)
											Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
										})

										Context("with successful provider session update", func() {
											BeforeEach(func() {
												updatedProviderSession := authTest.CloneProviderSession(providerSession)
												updatedProviderSession.ExternalID = pointer.FromString(ouraUserID)
												mockProviderSessionClient.EXPECT().UpdateProviderSession(gomock.Any(), providerSessionID, expectedProviderSessionUpdate).Return(updatedProviderSession, nil)
											})

											assertWorkCreate()
										})
									})
								}

								Context("without existing data source", func() {
									var expectedDataSrcUpdate *dataSource.Update

									BeforeEach(func() {
										mockDataSourceClient.EXPECT().List(gomock.Any(), userID, expectedDataSourceFilter, gomock.Any()).Return(dataSource.SourceArray{}, nil)
										expectedDataSrcUpdate = &dataSource.Update{
											ProviderExternalID: pointer.FromString(ouraUserID),
										}
									})

									It("returns failing process result if unable to update data source", func() {
										testErr := errorsTest.RandomError()
										mockDataSourceClient.EXPECT().Update(gomock.Any(), dataSrcID, nil, expectedDataSrcUpdate).Return(nil, testErr)
										Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
									})

									Context("with successful data source update", func() {
										BeforeEach(func() {
											updatedDataSrc := dataSourceTest.CloneSource(dataSrc)
											updatedDataSrc.ProviderExternalID = pointer.FromString(ouraUserID)
											mockDataSourceClient.EXPECT().Update(gomock.Any(), dataSrcID, nil, expectedDataSrcUpdate).Return(updatedDataSrc, nil)
										})

										assertProviderSessionUpdateAndWorkCreate()
									})
								})

								Context("with existing data source", func() {
									var existingDataSrc *dataSource.Source
									var expectedDataSrcUpdate *dataSource.Update

									BeforeEach(func() {
										existingDataSrc = &dataSource.Source{
											ID:                 dataSourceTest.RandomDataSourceID(),
											UserID:             userID,
											ProviderType:       oauth.ProviderType,
											ProviderName:       oura.ProviderName,
											ProviderExternalID: pointer.FromString(ouraUserID),
											State:              dataSource.StateDisconnected,
											CreatedTime:        test.RandomTimeBeforeNow(),
											Revision:           test.RandomInt(),
										}
										mockDataSourceClient.EXPECT().List(gomock.Any(), userID, expectedDataSourceFilter, gomock.Any()).Return(dataSource.SourceArray{existingDataSrc, existingDataSrc}, nil)
										expectedDataSrcUpdate = &dataSource.Update{
											ProviderSessionID: pointer.FromString(providerSessionID),
											State:             pointer.FromString(dataSource.StateConnected),
										}
									})

									It("returns failing process result if unable to update replacement data source", func() {
										testErr := errorsTest.RandomError()
										mockDataSourceClient.EXPECT().Update(gomock.Any(), existingDataSrc.ID, nil, expectedDataSrcUpdate).Return(nil, testErr)
										Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
									})

									Context("with successful existing data source update", func() {
										BeforeEach(func() {
											updatedDataSrc := dataSourceTest.CloneSource(existingDataSrc)
											updatedDataSrc.ProviderSessionID = pointer.FromString(providerSessionID)
											updatedDataSrc.State = dataSource.StateConnected
											mockDataSourceClient.EXPECT().Update(gomock.Any(), existingDataSrc.ID, nil, expectedDataSrcUpdate).Return(updatedDataSrc, nil)
											mockDataSourceClient.EXPECT().Delete(gomock.Any(), dataSrcID, nil).Return(true, nil)
										})

										assertProviderSessionUpdateAndWorkCreate()
									})
								})
							})
						})
					})
				})
			})
		})
	})
})
