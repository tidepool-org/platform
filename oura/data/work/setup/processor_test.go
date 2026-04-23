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
	customerioWork "github.com/tidepool-org/platform/customerio/work/event"
	"github.com/tidepool-org/platform/data"
	dataSetTest "github.com/tidepool-org/platform/data/set/test"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/oura"
	ouraDataWorkHistoric "github.com/tidepool-org/platform/oura/data/work/historic"
	ouraDataWorkSetup "github.com/tidepool-org/platform/oura/data/work/setup"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/page"
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
		Expect(ouraDataWorkSetup.FailingRetryDuration).To(Equal(1 * time.Minute))
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
		var mockDataSetClient *dataSetTest.MockClient
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraDataWorkSetup.Dependencies

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockDataSetClient = dataSetTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraDataWorkSetup.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				ProviderSessionClient: mockProviderSessionClient,
				DataSourceClient:      mockDataSourceClient,
				DataSetClient:         mockDataSetClient,
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
				var providerSessionID string
				var wrk *work.Work
				var processor *ouraDataWorkSetup.Processor
				var mockProcessingUpdater *workTest.MockProcessingUpdater

				BeforeEach(func() {
					userID = userTest.RandomUserID()
					providerSessionID = authTest.RandomProviderSessionID()
				})

				JustBeforeEach(func() {
					create, err := ouraDataWorkSetup.NewWorkCreate(providerSessionID)
					Expect(err).ToNot(HaveOccurred())
					Expect(create).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(create, work.StateProcessing)
					processor, err = ouraDataWorkSetup.NewProcessor(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processor).ToNot(BeNil())
					mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)
				})

				Context("Process", func() {
					var providerSessionStack *Stack[*auth.ProviderSession]

					BeforeEach(func() {
						providerSession := authTest.RandomProviderSession(test.AllowOptionals())
						providerSession.ID = providerSessionID
						providerSession.UserID = userID
						providerSession.Type = oauth.ProviderType
						providerSession.Name = oura.ProviderName
						providerSession.ExternalID = nil
						providerSessionStack = NewStack(providerSession)
					})

					It("returns failing process result if unable to fetch provider session from work", func() {
						testErr := errorsTest.RandomError()
						mockProviderSessionClient.EXPECT().GetProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(nil, testErr)
						Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
					})

					Context("with provider session", func() {
						BeforeEach(func() {
							mockProviderSessionClient.EXPECT().GetProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(providerSessionStack.Latest(), nil)
						})

						It("returns failing process result if unable to fetch data source from provider session", func() {
							testErr := errorsTest.RandomError()
							mockDataSourceClient.EXPECT().GetFromProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(nil, testErr)
							Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
						})

						Context("with personal info and data source", func() {
							var ouraUserID string
							var personalInfo *oura.PersonalInfo
							var dataSourceStack *Stack[*dataSource.Source]

							BeforeEach(func() {
								ouraUserID = ouraTest.RandomUserID()
								personalInfo = ouraTest.RandomPersonalInfo()
								personalInfo.ID = pointer.From(ouraUserID)
								initialDataSource := dataSourceTest.RandomSource(test.AllowOptionals())
								initialDataSource.UserID = userID
								initialDataSource.ProviderType = oauth.ProviderType
								initialDataSource.ProviderName = oura.ProviderName
								initialDataSource.ProviderSessionID = pointer.From(providerSessionID)
								initialDataSource.ProviderExternalID = nil
								initialDataSource.State = dataSource.StateConnected
								initialDataSource.DataSetID = nil
								initialDataSource.EarliestDataTime = nil
								initialDataSource.LatestDataTime = nil
								initialDataSource.LastImportTime = nil
								dataSourceStack = NewStack(initialDataSource)
								mockDataSourceClient.EXPECT().GetFromProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(initialDataSource, nil)
							})

							assertWorkCreate := func() {
								Context("with data source state change event work create", func() {
									var expectedDataSourceStateChangeEventWorkCreate *work.Create

									BeforeEach(func() {
										expectedDataSourceStateChangeEventWorkCreate = test.Must(customerioWork.NewDataSourceStateChangedEventWorkCreate(dataSourceStack.Latest()))
									})

									It("returned failing process result if unable to create data source state change event work", func() {
										testErr := errorsTest.RandomError()
										mockWorkClient.EXPECT().Create(gomock.Not(gomock.Nil()), expectedDataSourceStateChangeEventWorkCreate).Return(nil, testErr)
										Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
									})

									Context("with data source state change event work create and data historic work create", func() {
										var expectedDataHistoricWorkCreate *work.Create

										BeforeEach(func() {
											dataSourceStateChangeEventWork := workTest.NewWorkFromCreateWithState(expectedDataSourceStateChangeEventWorkCreate, work.StatePending)
											mockWorkClient.EXPECT().Create(gomock.Not(gomock.Nil()), expectedDataSourceStateChangeEventWorkCreate).Return(dataSourceStateChangeEventWork, nil)
											expectedDataHistoricWorkCreate = test.Must(ouraDataWorkHistoric.NewWorkCreate(providerSessionID, times.TimeRange{From: dataSourceStack.Latest().LatestDataTime}))
										})

										It("returned failing process result if unable to create data historic work", func() {
											testErr := errorsTest.RandomError()
											mockWorkClient.EXPECT().Create(gomock.Not(gomock.Nil()), expectedDataHistoricWorkCreate).Return(nil, testErr)
											Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
										})

										Context("with successful data historic work create", func() {
											BeforeEach(func() {
												dataHistoricWork := workTest.NewWorkFromCreateWithState(expectedDataHistoricWorkCreate, work.StatePending)
												mockWorkClient.EXPECT().Create(gomock.Not(gomock.Nil()), expectedDataHistoricWorkCreate).Return(dataHistoricWork, nil)
											})

											It("returns delete process result when successful", func() {
												Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchDeleteProcessResult())
											})
										})
									})
								})
							}

							assertEnsureDataSetAndWorkCreate := func() {
								Context("with ensuring data set for data source", func() {
									var initialDataSet *data.DataSet

									BeforeEach(func() {
										initialDataSet = dataTest.RandomDataSet(test.AllowOptionals())
										initialDataSet.UserID = pointer.From(userID)
									})

									Context("with no existing data set", func() {
										It("returns failing process result if unable to create data set for data source", func() {
											testErr := errorsTest.RandomError()
											mockDataSetClient.EXPECT().CreateUserDataSet(gomock.Not(gomock.Nil()), userID, ouraDataWorkSetup.NewDataSetCreate()).Return(nil, testErr)
											Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
										})

										Context("with successful create user data set", func() {
											var expectedDataSourceID string
											var expectedDataSourceUpdate *dataSource.Update

											BeforeEach(func() {
												mockDataSetClient.EXPECT().CreateUserDataSet(gomock.Not(gomock.Nil()), userID, ouraDataWorkSetup.NewDataSetCreate()).Return(initialDataSet, nil)
												expectedDataSourceID = dataSourceStack.Latest().ID
												expectedDataSourceUpdate = &dataSource.Update{
													Metadata:  metadataTest.PointerFromMetadataMap(dataSourceStack.Latest().Metadata),
													DataSetID: initialDataSet.ID,
												}
											})

											It("returns failing process result if unable to update data source with data set id", func() {
												testErr := errorsTest.RandomError()
												mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), expectedDataSourceID, nil, expectedDataSourceUpdate).Return(nil, testErr)
												Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
											})

											Context("with successful update data source with data set id", func() {
												BeforeEach(func() {
													updatedDataSource := dataSourceTest.CloneSource(dataSourceStack.Latest())
													updatedDataSource.DataSetID = initialDataSet.ID
													dataSourceStack.Push(updatedDataSource)
													mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), expectedDataSourceID, nil, expectedDataSourceUpdate).Return(updatedDataSource, nil)
												})

												assertWorkCreate()
											})
										})
									})

									Context("with existing data set", func() {
										BeforeEach(func() {
											dataSourceStack.Each(func(s *dataSource.Source) { s.DataSetID = initialDataSet.ID })
										})

										It("returns failing process result if unable to get data set", func() {
											testErr := errorsTest.RandomError()
											mockDataSetClient.EXPECT().GetDataSet(gomock.Not(gomock.Nil()), *initialDataSet.ID).Return(nil, testErr)
											Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
										})

										Context("with successful get data set", func() {
											BeforeEach(func() {
												mockDataSetClient.EXPECT().GetDataSet(gomock.Not(gomock.Nil()), *initialDataSet.ID).Return(initialDataSet, nil)
											})

											assertWorkCreate()
										})
									})
								})
							}

							Context("with provider session external id and data source provider external id", func() {
								BeforeEach(func() {
									providerSessionStack.Each(func(p *auth.ProviderSession) {
										p.ExternalID = pointer.From(ouraUserID)
									})
									dataSourceStack.Each(func(s *dataSource.Source) {
										s.ProviderExternalID = pointer.From(ouraUserID)
										s.LastImportTime = pointer.From(test.RandomTimeBeforeNow())
										s.LatestDataTime = pointer.From(test.RandomTimeBefore(*s.LastImportTime))
										s.EarliestDataTime = pointer.From(test.RandomTimeBefore(*s.LatestDataTime))
									})
								})

								assertEnsureDataSetAndWorkCreate()
							})

							It("returns failing process result if oura client returns an error when getting personal info", func() {
								testErr := errorsTest.RandomError()
								mockOuraClient.EXPECT().GetPersonalInfo(gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil())).Return(nil, testErr)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
							})

							Context("with personal info", func() {
								var expectedDataSourceFilter *dataSource.Filter

								BeforeEach(func() {
									mockOuraClient.EXPECT().GetPersonalInfo(gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil())).Return(personalInfo, nil)
									expectedDataSourceFilter = &dataSource.Filter{
										ProviderType:       pointer.From(oauth.ProviderType),
										ProviderName:       pointer.From(oura.ProviderName),
										ProviderExternalID: pointer.From(ouraUserID),
									}
								})

								It("returns failing process result if data source list returns an error", func() {
									testErr := errorsTest.RandomError()
									mockDataSourceClient.EXPECT().List(gomock.Not(gomock.Nil()), userID, expectedDataSourceFilter, page.NewPagination()).Return(nil, testErr)
									Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
								})

								assertProviderSessionUpdateEnsureDataSetAndWorkCreate := func() {
									Context("with provider session update", func() {
										var expectedProviderSessionUpdate *auth.ProviderSessionUpdate

										BeforeEach(func() {
											expectedProviderSessionUpdate = &auth.ProviderSessionUpdate{
												OAuthToken: providerSessionStack.Latest().OAuthToken,
												ExternalID: pointer.From(ouraUserID),
											}
										})

										It("returns failing process result if unable to update provider session", func() {
											testErr := errorsTest.RandomError()
											mockProviderSessionClient.EXPECT().UpdateProviderSession(gomock.Not(gomock.Nil()), providerSessionID, expectedProviderSessionUpdate).Return(nil, testErr)
											Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
										})

										Context("with successful provider session update", func() {
											BeforeEach(func() {
												providerSession := authTest.CloneProviderSession(providerSessionStack.Latest())
												providerSession.ExternalID = pointer.From(ouraUserID)
												providerSessionStack.Push(providerSession)
												mockProviderSessionClient.EXPECT().UpdateProviderSession(gomock.Not(gomock.Nil()), providerSessionID, expectedProviderSessionUpdate).Return(providerSession, nil)
											})

											assertEnsureDataSetAndWorkCreate()
										})
									})
								}

								Context("without existing data source", func() {
									var expectedDataSourceID string
									var expectedDataSourceUpdate *dataSource.Update

									BeforeEach(func() {
										mockDataSourceClient.EXPECT().List(gomock.Not(gomock.Nil()), userID, expectedDataSourceFilter, page.NewPagination()).Return(dataSource.SourceArray{}, nil)
										expectedDataSourceID = dataSourceStack.Latest().ID
										expectedDataSourceUpdate = &dataSource.Update{
											ProviderExternalID: pointer.From(ouraUserID),
											Metadata:           metadataTest.PointerFromMetadataMap(dataSourceStack.Latest().Metadata),
										}
									})

									It("returns failing process result if unable to update data source", func() {
										testErr := errorsTest.RandomError()
										mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), expectedDataSourceID, nil, expectedDataSourceUpdate).Return(nil, testErr)
										Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
									})

									Context("with successful data source update", func() {
										BeforeEach(func() {
											updatedDataSource := dataSourceTest.CloneSource(dataSourceStack.Latest())
											updatedDataSource.ProviderExternalID = pointer.From(ouraUserID)
											dataSourceStack.Push(updatedDataSource)
											mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), expectedDataSourceID, nil, expectedDataSourceUpdate).Return(updatedDataSource, nil)
										})

										assertProviderSessionUpdateEnsureDataSetAndWorkCreate()
									})
								})

								Context("with existing data source", func() {
									var existingDataSource *dataSource.Source
									var expectedDataSourceUpdate *dataSource.Update

									BeforeEach(func() {
										existingDataSource = dataSourceTest.RandomSource(test.AllowOptionals())
										existingDataSource.UserID = userID
										existingDataSource.ProviderType = oauth.ProviderType
										existingDataSource.ProviderName = oura.ProviderName
										existingDataSource.ProviderExternalID = pointer.From(ouraUserID)
										existingDataSource.State = dataSource.StateDisconnected
										existingDataSource.DataSetID = nil
										expectedDataSourceUpdate = &dataSource.Update{
											ProviderSessionID: pointer.From(providerSessionID),
											State:             pointer.From(dataSource.StateConnected),
										}
										if existingDataSource.Metadata != nil {
											expectedDataSourceUpdate.Metadata = pointer.From(existingDataSource.Metadata)
										}
										mockDataSourceClient.EXPECT().List(gomock.Not(gomock.Nil()), userID, expectedDataSourceFilter, page.NewPagination()).Return(dataSource.SourceArray{existingDataSource, existingDataSource}, nil)
									})

									It("returns failing process result if unable to update replacement data source", func() {
										testErr := errorsTest.RandomError()
										mockDataSourceClient.EXPECT().Delete(gomock.Not(gomock.Nil()), dataSourceStack.Latest().ID, nil).Return(true, nil)
										mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), existingDataSource.ID, nil, expectedDataSourceUpdate).Return(nil, testErr)
										Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
									})

									Context("with successful existing data source update", func() {
										BeforeEach(func() {
											mockDataSourceClient.EXPECT().Delete(gomock.Not(gomock.Nil()), dataSourceStack.Latest().ID, nil).Return(true, nil)
											dataSourceStack.Push(existingDataSource)
											mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), existingDataSource.ID, nil, expectedDataSourceUpdate).Return(existingDataSource, nil)
										})

										assertProviderSessionUpdateEnsureDataSetAndWorkCreate()
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

type Stack[T any] []T

func NewStack[T any](element T) *Stack[T] {
	return &Stack[T]{element}
}

func (s *Stack[T]) Push(element T) {
	*s = append(*s, element)
}

func (s *Stack[T]) Initial() T {
	if len(*s) == 0 {
		panic("stack is empty")
	}
	return (*s)[0]
}

func (s *Stack[T]) Latest() T {
	if len(*s) == 0 {
		panic("stack is empty")
	}
	return (*s)[len(*s)-1]
}

func (s *Stack[T]) Each(fn func(T)) {
	for _, element := range *s {
		fn(element)
	}
}
