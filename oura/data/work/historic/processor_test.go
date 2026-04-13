package historic_test

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"slices"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataRawTest "github.com/tidepool-org/platform/data/raw/test"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metadata"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/oura"
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
	ouraDataWorkHistoric "github.com/tidepool-org/platform/oura/data/work/historic"
	ouraDataWorkHistoricTest "github.com/tidepool-org/platform/oura/data/work/historic/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/times"
	timesTest "github.com/tidepool-org/platform/times/test"
	userTest "github.com/tidepool-org/platform/user/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("processor", func() {
	It("FailingRetryDuration is expected", func() {
		Expect(ouraDataWorkHistoric.FailingRetryDuration).To(Equal(time.Minute))
	})

	It("FailingRetryDurationJitter is expected", func() {
		Expect(ouraDataWorkHistoric.FailingRetryDurationJitter).To(Equal(5 * time.Second))
	})

	Context("Metadata", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *ouraDataWorkHistoric.Metadata)) {
				datum := ouraDataWorkHistoricTest.RandomMetadata(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraDataWorkHistoricTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraDataWorkHistoricTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *ouraDataWorkHistoric.Metadata) {},
			),
			Entry("empty",
				func(datum *ouraDataWorkHistoric.Metadata) {
					*datum = ouraDataWorkHistoric.Metadata{}
				},
			),
			Entry("all",
				func(datum *ouraDataWorkHistoric.Metadata) {
					datum.ProviderSessionID = pointer.From(authTest.RandomProviderSessionID())
					datum.TimeRange = timesTest.RandomTimeRange()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *ouraDataWorkHistoric.Metadata), expectedErrors ...error) {
					expectedDatum := ouraDataWorkHistoricTest.RandomMetadata(test.AllowOptional())
					object := ouraDataWorkHistoricTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &ouraDataWorkHistoric.Metadata{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *ouraDataWorkHistoric.Metadata) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *ouraDataWorkHistoric.Metadata) {
						clear(object)
						*expectedDatum = ouraDataWorkHistoric.Metadata{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *ouraDataWorkHistoric.Metadata) {
						object["providerSessionId"] = true
						object["timeRange"] = true
						expectedDatum.ProviderSessionID = nil
						expectedDatum.TimeRange = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerSessionId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/timeRange"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *ouraDataWorkHistoric.Metadata), expectedErrors ...error) {
					datum := ouraDataWorkHistoricTest.RandomMetadata(test.AllowOptional())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *ouraDataWorkHistoric.Metadata) {},
				),
				Entry("multiple errors",
					func(datum *ouraDataWorkHistoric.Metadata) {
						datum.ProviderSessionID = pointer.From("")
						datum.TimeRange = timesTest.RandomTimeRange()
						datum.TimeRange.From = pointer.From(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timeRange/from"),
				),
			)
		})
	})

	Context("with dependencies", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockProviderSessionClient *providerSessionTest.MockClient
		var mockDataSourceClient *dataSourceTest.MockClient
		var mockDataRawClient *dataRawTest.MockClient
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraDataWorkHistoric.Dependencies

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockDataRawClient = dataRawTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraDataWorkHistoric.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				ProviderSessionClient: mockProviderSessionClient,
				DataSourceClient:      mockDataSourceClient,
				DataRawClient:         mockDataRawClient,
				OuraClient:            mockOuraClient,
			}
		})

		Context("NewProcessor", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.WorkClient = nil
				processor, err := ouraDataWorkHistoric.NewProcessor(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns successfully", func() {
				processor, err := ouraDataWorkHistoric.NewProcessor(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})

			Context("with processor", func() {
				var now time.Time
				var userID string
				var providerSessionID string
				var ouraUserID string
				var timeRange *times.TimeRange
				var expectedTimeRange times.TimeRange
				var wrk *work.Work
				var processor *ouraDataWorkHistoric.Processor
				var mockProcessingUpdater *workTest.MockProcessingUpdater

				BeforeEach(func() {
					now = time.Now()
					userID = userTest.RandomUserID()
					providerSessionID = authTest.RandomProviderSessionID()
					ouraUserID = ouraTest.RandomUserID()
					timeRange = timesTest.RandomTimeRange(test.AllowOptional())
					expectedTimeRange = ouraDataWorkHistoric.NormalizeTimeRange(timeRange, ouraDataWorkHistoric.LaunchDate, now)
				})

				JustBeforeEach(func() {
					create, err := ouraDataWorkHistoric.NewWorkCreate(providerSessionID, *timeRange)
					Expect(err).ToNot(HaveOccurred())
					Expect(create).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(create, work.StateProcessing)
					processor, err = ouraDataWorkHistoric.NewProcessor(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processor).ToNot(BeNil())
					processor.Now = func() time.Time { return now }
					mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)
				})

				Context("Process", func() {
					It("returns failing process result if unable to fetch provider session from work", func() {
						testErr := errorsTest.RandomError()
						mockProviderSessionClient.EXPECT().GetProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(nil, testErr)
						Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
					})

					Context("with provider session", func() {
						var providerSession *auth.ProviderSession

						BeforeEach(func() {
							scopesWithoutDataTypes := []string{oura.ScopeEmail, oura.ScopePersonal}
							scope := slices.DeleteFunc(oura.Scopes(), func(s string) bool { return slices.Contains(scopesWithoutDataTypes, s) })
							rand.Shuffle(len(scope), func(i, j int) { scope[i], scope[j] = scope[j], scope[i] })
							scope = scope[:test.RandomIntFromRange(1, len(scope))]
							providerSession = authTest.RandomProviderSession(test.AllowOptional())
							providerSession.ID = providerSessionID
							providerSession.UserID = userID
							providerSession.Type = oauth.ProviderType
							providerSession.Name = oura.ProviderName
							providerSession.OAuthToken.Scope = pointer.From(scope)
							providerSession.ExternalID = pointer.From(ouraUserID)
							mockProviderSessionClient.EXPECT().GetProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(providerSession, nil)
						})

						It("returns failing process result if unable to fetch data source from provider session", func() {
							testErr := errorsTest.RandomError()
							mockDataSourceClient.EXPECT().GetFromProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(nil, testErr)
							Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
						})

						Context("with data source", func() {
							var dataSourceID string
							var dataSrc *dataSource.Source

							BeforeEach(func() {
								dataSourceID = dataSourceTest.RandomDataSourceID()
								dataSrc = dataSourceTest.RandomSource(test.AllowOptional())
								dataSrc.ID = dataSourceID
								dataSrc.UserID = userID
								dataSrc.ProviderType = oauth.ProviderType
								dataSrc.ProviderName = oura.ProviderName
								dataSrc.ProviderExternalID = pointer.From(ouraUserID)
								dataSrc.ProviderSessionID = pointer.From(providerSessionID)
								dataSrc.State = dataSource.StateConnected
								dataSrc.DataSetID = pointer.From(dataTest.RandomDataSetID())
								mockDataSourceClient.EXPECT().GetFromProviderSession(gomock.Not(gomock.Nil()), providerSessionID).Return(dataSrc, nil)
							})

							It("returns failed process result if data set id is missing", func() {
								dataSrc.DataSetID = nil
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailedProcessResultError(MatchError("data set id is missing")))
							})

							It("returns successfully if there are scope is missing", func() {
								providerSession.OAuthToken.Scope = nil
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchDeleteProcessResult())
							})

							It("returns successfully if there are scope is empty", func() {
								providerSession.OAuthToken.Scope = pointer.From([]string{})
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchDeleteProcessResult())
							})

							It("returns failing process result if get data fails", func() {
								testErr := errorsTest.RandomError()
								dataType := oura.DataTypesForScopes(*providerSession.OAuthToken.Scope)[0]
								mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), dataType, &expectedTimeRange, &oura.Pagination{}, gomock.Not(gomock.Nil())).Return(nil, testErr)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
							})

							It("returns failing process result if get data response is missing", func() {
								dataType := oura.DataTypesForScopes(*providerSession.OAuthToken.Scope)[0]
								mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), dataType, &expectedTimeRange, &oura.Pagination{}, gomock.Not(gomock.Nil())).Return(nil, nil)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(fmt.Sprintf("data response for data type %q is missing", dataType))))
							})

							It("correctly paginates through data and then errors", func() {
								testErr := errorsTest.RandomError()
								dataType := oura.DataTypesForScopes(*providerSession.OAuthToken.Scope)[0]
								var token *string
								for range test.RandomIntFromRange(3, 10) {
									nextToken := pointer.From(ouraTest.RandomNextToken())
									mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), dataType, &expectedTimeRange, &oura.Pagination{NextToken: token}, gomock.Not(gomock.Nil())).Return(&oura.DataResponse{Data: ouraTest.RandomData(), Pagination: oura.Pagination{NextToken: nextToken}}, nil)
									token = nextToken
								}
								mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), dataType, &expectedTimeRange, &oura.Pagination{NextToken: token}, gomock.Not(gomock.Nil())).Return(nil, testErr)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
							})

							Context("with data", func() {
								BeforeEach(func() {
									for _, dataType := range oura.DataTypesForScopes(*providerSession.OAuthToken.Scope) {
										mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), dataType, &expectedTimeRange, &oura.Pagination{}, gomock.Not(gomock.Nil())).Return(&oura.DataResponse{Data: ouraTest.RandomData()}, nil)
									}
								})

								It("returns failing process result if create data raw fails", func() {
									testErr := errorsTest.RandomError()
									mockDataRawClient.EXPECT().
										Create(gomock.Not(gomock.Nil()), userID, *dataSrc.DataSetID, gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil())).
										DoAndReturn(func(_ context.Context, _ string, _ string, dataRawCreate *dataRaw.Create, reader io.Reader) (*dataRaw.Raw, error) {
											Expect(dataRawCreate).To(PointTo(MatchAllFields(Fields{
												"Metadata": Equal(map[string]any{
													"scope": test.AsAnyArray(*providerSession.OAuthToken.Scope),
													"timeRange": map[string]any{
														"from": expectedTimeRange.From.Format(time.RFC3339Nano),
														"to":   expectedTimeRange.To.Format(time.RFC3339Nano),
													},
												}),
												"DigestMD5":      BeNil(),
												"DigestSHA256":   BeNil(),
												"MediaType":      PointTo(Equal(net.MediaTypeJSON)),
												"ArchivableTime": PointTo(BeTemporally("~", now)),
											})))
											bites, err := io.ReadAll(reader)
											Expect(err).ToNot(HaveOccurred())
											Expect(bites).ToNot(BeEmpty())
											return nil, testErr
										})
									Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
								})

								Context("with successful create data raw", func() {
									var createdDataRaw *dataRaw.Raw
									var expectedDataSourceUpdate *dataSource.Update

									BeforeEach(func() {
										createdDataRaw = test.Must(metadata.WithMetadata(
											dataRawTest.RandomRaw(test.AllowOptional()),
											&ouraDataWork.Metadata{
												Scope: providerSession.OAuthToken.Scope,
												TimeRangeMetadata: times.TimeRangeMetadata{
													TimeRange: pointer.From(expectedTimeRange),
												},
											},
										))
										expectedDataSourceUpdate = &dataSource.Update{
											Metadata:       metadataTest.PointerFromMetadataMap(dataSrc.Metadata),
											LastImportTime: pointer.From(createdDataRaw.CreatedTime),
										}
										mockDataRawClient.EXPECT().
											Create(gomock.Not(gomock.Nil()), userID, *dataSrc.DataSetID, gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil())).
											DoAndReturn(func(_ context.Context, _ string, _ string, dataRawCreate *dataRaw.Create, reader io.Reader) (*dataRaw.Raw, error) {
												Expect(dataRawCreate).To(PointTo(MatchAllFields(Fields{
													"Metadata": Equal(map[string]any{
														"scope": test.AsAnyArray(*providerSession.OAuthToken.Scope),
														"timeRange": map[string]any{
															"from": expectedTimeRange.From.Format(time.RFC3339Nano),
															"to":   expectedTimeRange.To.Format(time.RFC3339Nano),
														},
													}),
													"DigestMD5":      BeNil(),
													"DigestSHA256":   BeNil(),
													"MediaType":      PointTo(Equal(net.MediaTypeJSON)),
													"ArchivableTime": PointTo(BeTemporally("~", now)),
												})))
												bites, err := io.ReadAll(reader)
												Expect(err).ToNot(HaveOccurred())
												Expect(bites).ToNot(BeEmpty())
												return createdDataRaw, nil
											})
									})

									It("returns failing process result if update data source fails", func() {
										testErr := errorsTest.RandomError()
										mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataSourceID, nil, expectedDataSourceUpdate).Return(nil, testErr)
										Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
									})

									It("returns delete process result if successful", func() {
										updatedDataSource := dataSourceTest.RandomSource(test.AllowOptional())
										mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataSourceID, nil, expectedDataSourceUpdate).Return(updatedDataSource, nil)
										Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchDeleteProcessResult())
									})
								})
							})
						})
					})
				})
			})
		})
	})

	Context("NormalizeTimeRange", func() {
		now := time.Now()
		nowNormalized := time.Now().UTC().Truncate(24 * time.Hour)
		location := time.FixedZone("Test", -5*60*60) // UTC-5

		DescribeTable("normalizes the time range as expected",
			func(timeRange *times.TimeRange, expectedTimeRange times.TimeRange) {
				Expect(ouraDataWorkHistoric.NormalizeTimeRange(timeRange, ouraDataWorkHistoric.LaunchDate, now)).To(Equal(expectedTimeRange))
			},
			Entry("with nil input",
				nil,
				times.TimeRange{
					From: pointer.From(ouraDataWorkHistoric.LaunchDate),
					To:   pointer.From(nowNormalized),
				},
			),
			Entry("with from only",
				&times.TimeRange{
					From: pointer.From(time.Date(2020, 1, 1, 19, 34, 56, 789000000, location)),
				},
				times.TimeRange{
					From: pointer.From(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)),
					To:   pointer.From(nowNormalized),
				},
			),
			Entry("with to only",
				&times.TimeRange{
					To: pointer.From(time.Date(2020, 1, 1, 19, 34, 56, 789000000, location)),
				},
				times.TimeRange{
					From: pointer.From(ouraDataWorkHistoric.LaunchDate),
					To:   pointer.From(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)),
				},
			),
			Entry("with from and to",
				&times.TimeRange{
					From: pointer.From(time.Date(2016, 5, 10, 12, 34, 56, 789000000, location)),
					To:   pointer.From(time.Date(2019, 8, 20, 23, 45, 1, 123000000, location)),
				},
				times.TimeRange{
					From: pointer.From(time.Date(2016, 5, 10, 0, 0, 0, 0, time.UTC)),
					To:   pointer.From(time.Date(2019, 8, 21, 0, 0, 0, 0, time.UTC)),
				},
			),
		)
	})
})
