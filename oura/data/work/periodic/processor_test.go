package periodic_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	providerSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataRawTest "github.com/tidepool-org/platform/data/raw/test"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	dataWork "github.com/tidepool-org/platform/data/work"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/oura"
	ouraData "github.com/tidepool-org/platform/oura/data"
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
	ouraDataWorkPeriodic "github.com/tidepool-org/platform/oura/data/work/periodic"
	ouraDataWorkPeriodicTest "github.com/tidepool-org/platform/oura/data/work/periodic/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
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
	It("PendingAvailableDuration is expected", func() {
		Expect(ouraDataWorkPeriodic.PendingAvailableDuration).To(Equal(12 * time.Hour))
	})

	It("FailingRetryDuration is expected", func() {
		Expect(ouraDataWorkPeriodic.FailingRetryDuration).To(Equal(1 * time.Minute))
	})

	It("FailingRetryDurationJitter is expected", func() {
		Expect(ouraDataWorkPeriodic.FailingRetryDurationJitter).To(Equal(5 * time.Second))
	})

	It("FailingRetryDurationMaximum is expected", func() {
		Expect(ouraDataWorkPeriodic.FailingRetryDurationMaximum).To(Equal(12 * time.Hour))
	})

	Context("DataTypes", func() {
		It("returns expected data types", func() {
			Expect(ouraDataWorkPeriodic.DataTypes()).To(Equal([]string{
				oura.DataTypeHeartRate,
				oura.DataTypeRingBatteryLevel,
			}))
		})
	})

	It("MetadataKeyDataTypeStartTimes is expected", func() {
		Expect(ouraDataWorkPeriodic.MetadataKeyDataTypeStartTimes).To(Equal("dataTypeStartTimes"))
	})

	It("MetadataKeyDataTypeNextTokens is expected", func() {
		Expect(ouraDataWorkPeriodic.MetadataKeyDataTypeNextTokens).To(Equal("dataTypeNextTokens"))
	})

	Context("Metadata", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *ouraDataWorkPeriodic.Metadata)) {
				datum := ouraDataWorkPeriodicTest.RandomMetadata(test.AllowOptionals())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraDataWorkPeriodicTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraDataWorkPeriodicTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *ouraDataWorkPeriodic.Metadata) {},
			),
			Entry("empty",
				func(datum *ouraDataWorkPeriodic.Metadata) {
					*datum = ouraDataWorkPeriodic.Metadata{}
				},
			),
			Entry("all",
				func(datum *ouraDataWorkPeriodic.Metadata) {
					datum.ProviderSessionMetadata = *providerSessionWorkTest.RandomMetadata()
					datum.DataTypeStartTimes = ouraDataWorkPeriodicTest.RandomDataTypeStartTimes(test.AllowOptionals())
					datum.DataTypeNextTokens = ouraDataWorkPeriodicTest.RandomDataTypeNextTokens(test.AllowOptionals())
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *ouraDataWorkPeriodic.Metadata), expectedErrors ...error) {
					expectedDatum := ouraDataWorkPeriodicTest.RandomMetadata(test.AllowOptionals())
					object := ouraDataWorkPeriodicTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &ouraDataWorkPeriodic.Metadata{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *ouraDataWorkPeriodic.Metadata) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *ouraDataWorkPeriodic.Metadata) {
						clear(object)
						*expectedDatum = ouraDataWorkPeriodic.Metadata{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *ouraDataWorkPeriodic.Metadata) {
						object["providerSessionId"] = true
						object["dataTypeStartTimes"] = true
						object["dataTypeNextTokens"] = true
						expectedDatum.ProviderSessionID = nil
						expectedDatum.DataTypeStartTimes = nil
						expectedDatum.DataTypeNextTokens = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerSessionId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/dataTypeStartTimes"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/dataTypeNextTokens"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *ouraDataWorkPeriodic.Metadata), expectedErrors ...error) {
					datum := ouraDataWorkPeriodicTest.RandomMetadata(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *ouraDataWorkPeriodic.Metadata) {},
				),
				Entry("provider session id missing",
					func(datum *ouraDataWorkPeriodic.Metadata) {
						datum.ProviderSessionID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerSessionId"),
				),
				Entry("data type start times missing",
					func(datum *ouraDataWorkPeriodic.Metadata) {
						datum.DataTypeStartTimes = nil
					},
				),
				Entry("data type start times empty",
					func(datum *ouraDataWorkPeriodic.Metadata) {
						datum.DataTypeStartTimes = &dataWork.StringTimeMap{}
					},
				),
				Entry("data type start times invalid",
					func(datum *ouraDataWorkPeriodic.Metadata) {
						datum.DataTypeStartTimes = &dataWork.StringTimeMap{"invalid": pointer.From(time.Time{})}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", ouraDataWorkPeriodic.DataTypes()), "/dataTypeStartTimes/invalid/#"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataTypeStartTimes/invalid"),
				),
				Entry("data type next tokens missing",
					func(datum *ouraDataWorkPeriodic.Metadata) {
						datum.DataTypeNextTokens = nil
					},
				),
				Entry("data type next tokens empty",
					func(datum *ouraDataWorkPeriodic.Metadata) {
						datum.DataTypeNextTokens = &dataWork.StringStringMap{}
					},
				),
				Entry("data type next tokens invalid",
					func(datum *ouraDataWorkPeriodic.Metadata) {
						datum.DataTypeNextTokens = &dataWork.StringStringMap{"invalid": pointer.From("")}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", ouraDataWorkPeriodic.DataTypes()), "/dataTypeNextTokens/invalid/#"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataTypeNextTokens/invalid"),
				),
				Entry("multiple errors",
					func(datum *ouraDataWorkPeriodic.Metadata) {
						datum.ProviderSessionID = nil
						datum.DataTypeStartTimes = &dataWork.StringTimeMap{"invalid": pointer.From(time.Time{})}
						datum.DataTypeNextTokens = &dataWork.StringStringMap{"invalid": pointer.From("")}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerSessionId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", ouraDataWorkPeriodic.DataTypes()), "/dataTypeStartTimes/invalid/#"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataTypeStartTimes/invalid"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", ouraDataWorkPeriodic.DataTypes()), "/dataTypeNextTokens/invalid/#"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataTypeNextTokens/invalid"),
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
		var dependencies ouraDataWork.Dependencies

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockDataRawClient = dataRawTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraDataWork.Dependencies{
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
				processor, err := ouraDataWorkPeriodic.NewProcessor(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns successfully", func() {
				processor, err := ouraDataWorkPeriodic.NewProcessor(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})

			Context("with processor", func() {
				var now time.Time
				var userID string
				var providerSessionID string
				var ouraUserID string
				var dataTypeStartTimes *dataWork.StringTimeMap
				var dataTypeNextTokens *dataWork.StringStringMap
				var wrk *work.Work
				var processor *ouraDataWorkPeriodic.Processor
				var mockProcessingUpdater *workTest.MockProcessingUpdater

				BeforeEach(func() {
					now = time.Now().UTC()
					userID = userTest.RandomUserID()
					providerSessionID = authTest.RandomProviderSessionID()
					ouraUserID = ouraTest.RandomUserID()
					dataTypeStartTimes = test.RandomOptionalPointerWithOptions(ouraDataWorkPeriodicTest.RandomDataTypeStartTimes, test.AllowOptionals())
					dataTypeNextTokens = test.RandomOptionalPointerWithOptions(ouraDataWorkPeriodicTest.RandomDataTypeNextTokens, test.AllowOptionals())
					create, err := ouraDataWorkPeriodic.NewWorkCreate(providerSessionID)
					Expect(err).ToNot(HaveOccurred())
					Expect(create).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(create, work.StateProcessing)
					if dataTypeStartTimes != nil {
						wrk.Metadata[ouraDataWorkPeriodic.MetadataKeyDataTypeStartTimes] = ouraDataWorkPeriodicTest.NewObjectFromDataTypeStartTimes(dataTypeStartTimes, test.ObjectFormatJSON)
					}
					if dataTypeNextTokens != nil {
						wrk.Metadata[ouraDataWorkPeriodic.MetadataKeyDataTypeNextTokens] = ouraDataWorkPeriodicTest.NewObjectFromDataTypeNextTokens(dataTypeNextTokens, test.ObjectFormatJSON)
					}
					processor, err = ouraDataWorkPeriodic.NewProcessor(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processor).ToNot(BeNil())
					processor.NowFunc = func() time.Time { return now }
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
							providerSession = authTest.RandomProviderSession(test.AllowOptionals())
							providerSession.ID = providerSessionID
							providerSession.UserID = userID
							providerSession.Type = oauth.ProviderType
							providerSession.Name = oura.ProviderName
							providerSession.OAuthToken.Scope = pointer.From(test.RandomStringArrayFromArrayWithoutDuplicates(oura.ScopesForDataTypes(ouraDataWorkPeriodic.DataTypes())))
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
								dataSrc = dataSourceTest.RandomSource(test.AllowOptionals())
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
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailedProcessResultError(MatchError("data source data set id is missing")))
							})

							Context("with data types from scope", func() {
								var expectedDataType *string
								var expectedTimeRange *times.TimeRange
								var expectedStartTime *time.Time
								var expectedNextToken *string

								paginate := func(inner func()) {
									Context("it paginates through the data types", func() {
										BeforeEach(func() {
											for _, dataType := range ouraDataWorkPeriodic.DataTypes() {
												if !oura.DataTypeInScopes(dataType, providerSession.OAuthToken.Scope) {
													continue
												}

												if dataTypeStartTimes == nil {
													dataTypeStartTimes = &dataWork.StringTimeMap{}
												}
												if dataTypeNextTokens == nil {
													dataTypeNextTokens = &dataWork.StringStringMap{}
												}

												expectedStartTime = (*dataTypeStartTimes)[dataType]
												expectedTimeRange = &times.TimeRange{
													From: expectedStartTime,
												}
												expectedPagination := &oura.Pagination{
													NextToken: (*dataTypeNextTokens)[dataType],
												}

												nextTokens := test.RandomArrayWithLength(test.RandomIntFromRange(0, 2), func() *string { return pointer.From(ouraTest.RandomNextToken()) })
												if expectedDataType != nil && *expectedDataType == dataType {
													nextTokens = append(nextTokens, expectedNextToken)
												} else {
													nextTokens = append(nextTokens, nil)
												}

												for _, nextToken := range nextTokens {
													expectedData := ouraTest.RandomData()

													dataResult := &oura.DataResponse{
														Data:       expectedData,
														Pagination: oura.Pagination{NextToken: nextToken},
													}
													mockOuraClient.EXPECT().
														GetData(gomock.Not(gomock.Nil()), dataType, expectedTimeRange, expectedPagination, gomock.Not(gomock.Nil())).
														DoAndReturn(func(_ context.Context, _ string, _ *times.TimeRange, _ *oura.Pagination, _ oauth.TokenSource) (*oura.DataResponse, error) {
															return dataResult, nil
														})

													expectedDataRawCreate := &dataRaw.Create{
														Metadata: map[string]any{
															"dataType":  dataType,
															"timeRange": timesTest.NewObjectFromTimeRange(expectedTimeRange, test.ObjectFormatJSON),
														},
														MediaType:      pointer.From(net.MediaTypeJSON),
														ArchivableTime: pointer.From(now),
													}
													dataRawResult := test.Must(metadata.WithMetadata(
														dataRawTest.RandomRaw(test.AllowOptionals()),
														&ouraData.Metadata{
															DataType: dataType,
															TimeRangeMetadata: times.TimeRangeMetadata{
																TimeRange: expectedTimeRange,
															},
														},
													))
													mockDataRawClient.EXPECT().
														Create(gomock.Not(gomock.Nil()), userID, *dataSrc.DataSetID, expectedDataRawCreate, gomock.Not(gomock.Nil())).
														DoAndReturn(func(_ context.Context, _ string, _ string, dataRawCreate *dataRaw.Create, reader io.Reader) (*dataRaw.Raw, error) {
															bites, err := io.ReadAll(reader)
															Expect(err).ToNot(HaveOccurred())
															Expect(bites).To(MatchJSON(test.Must(json.Marshal(oura.DataMap{dataType: expectedData}))))
															return dataRawResult, nil
														})

													mockDataSourceClient.EXPECT().
														Update(gomock.Not(gomock.Nil()), dataSourceID, nil, &dataSource.Update{LastImportTime: pointer.From(dataRawResult.CreatedTime), Metadata: &dataSrc.Metadata}).
														DoAndReturn(func(_ context.Context, _ string, _ *request.Condition, dataSrcUpdate *dataSource.Update) (*dataSource.Source, error) {
															dataSrcUpdated := dataSourceTest.CloneSource(dataSrc)
															dataSrcUpdated.LastImportTime = dataSrcUpdate.LastImportTime
															if dataSrcUpdate.Metadata != nil {
																dataSrcUpdated.Metadata = *dataSrcUpdate.Metadata
															}
															return dataSrcUpdated, nil
														})

													if timeMaximum := dataResult.Data.TimeMaximum(); timeMaximum != nil && (expectedStartTime == nil || timeMaximum.After(*expectedStartTime)) {
														expectedStartTime = timeMaximum
													}

													expectedPagination = &dataResult.Pagination

													if nextToken == nil {
														break
													}

													(*dataTypeNextTokens)[dataType] = nextToken
													expectedProcessingUpdate := work.ProcessingUpdate{
														Metadata: map[string]any{
															"providerSessionId":  providerSessionID,
															"dataTypeStartTimes": ouraDataWorkPeriodicTest.NewObjectFromDataTypeStartTimes(dataTypeStartTimes, test.ObjectFormatJSON),
															"dataTypeNextTokens": ouraDataWorkPeriodicTest.NewObjectFromDataTypeNextTokens(dataTypeNextTokens, test.ObjectFormatJSON),
														},
													}
													mockProcessingUpdater.EXPECT().ProcessingUpdate(gomock.Not(gomock.Nil()), expectedProcessingUpdate).Return(wrk, nil)

												}

												if expectedDataType != nil && *expectedDataType == dataType {
													break
												}

												(*dataTypeStartTimes)[dataType] = expectedStartTime
												delete(*dataTypeNextTokens, dataType)
												expectedProcessingUpdate := work.ProcessingUpdate{
													Metadata: map[string]any{
														"providerSessionId":  providerSessionID,
														"dataTypeStartTimes": ouraDataWorkPeriodicTest.NewObjectFromDataTypeStartTimes(dataTypeStartTimes, test.ObjectFormatJSON),
														"dataTypeNextTokens": ouraDataWorkPeriodicTest.NewObjectFromDataTypeNextTokens(dataTypeNextTokens, test.ObjectFormatJSON),
													},
												}

												mockProcessingUpdater.EXPECT().ProcessingUpdate(gomock.Not(gomock.Nil()), expectedProcessingUpdate).Return(wrk, nil)
											}
										})

										inner()
									})
								}

								Context("it errors as expected after paging through the data types", func() {
									var expectedData oura.Data

									BeforeEach(func() {
										dataTypes := slices.DeleteFunc(ouraDataWorkPeriodic.DataTypes(), func(dataType string) bool {
											return !oura.DataTypeInScopes(dataType, providerSession.OAuthToken.Scope)
										})
										expectedDataType = pointer.From(test.RandomStringFromArray(dataTypes))
										expectedNextToken = pointer.From(ouraTest.RandomNextToken())
										expectedData = ouraTest.RandomData()
									})

									paginate(func() {
										Context("with expected start time", func() {
											BeforeEach(func() {
												if timeMaximum := expectedData.TimeMaximum(); timeMaximum != nil && (expectedStartTime == nil || timeMaximum.After(*expectedStartTime)) {
													expectedStartTime = timeMaximum
												}
											})

											It("returns failing process result if get data fails", func() {
												testErr := errorsTest.RandomError()
												mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), *expectedDataType, expectedTimeRange, &oura.Pagination{NextToken: expectedNextToken}, gomock.Not(gomock.Nil())).Return(nil, testErr)
												Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
											})

											It("returns failing process result if get data response is missing", func() {
												mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), *expectedDataType, expectedTimeRange, &oura.Pagination{NextToken: expectedNextToken}, gomock.Not(gomock.Nil())).Return(nil, nil)
												Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(fmt.Sprintf("data response for data type %q is missing", *expectedDataType))))
											})

											It("returns failing process result if create data raw fails", func() {
												testErr := errorsTest.RandomError()
												mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), *expectedDataType, expectedTimeRange, &oura.Pagination{NextToken: expectedNextToken}, gomock.Not(gomock.Nil())).Return(&oura.DataResponse{Data: expectedData}, nil)
												expectedDataRawCreate := &dataRaw.Create{
													Metadata: map[string]any{
														"dataType":  *expectedDataType,
														"timeRange": timesTest.NewObjectFromTimeRange(expectedTimeRange, test.ObjectFormatJSON),
													},
													MediaType:      pointer.From(net.MediaTypeJSON),
													ArchivableTime: pointer.From(now),
												}
												mockDataRawClient.EXPECT().
													Create(gomock.Not(gomock.Nil()), userID, *dataSrc.DataSetID, expectedDataRawCreate, gomock.Not(gomock.Nil())).
													DoAndReturn(func(_ context.Context, _ string, _ string, dataRawCreate *dataRaw.Create, reader io.Reader) (*dataRaw.Raw, error) {
														bites, err := io.ReadAll(reader)
														Expect(err).ToNot(HaveOccurred())
														Expect(bites).To(MatchJSON(test.Must(json.Marshal(oura.DataMap{*expectedDataType: expectedData}))))
														return nil, testErr
													})
												Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
											})

											It("returns failing process result if update data source fails", func() {
												testErr := errorsTest.RandomError()
												dataRw := test.Must(metadata.WithMetadata(
													dataRawTest.RandomRaw(test.AllowOptionals()),
													&ouraData.Metadata{
														DataType: *expectedDataType,
														TimeRangeMetadata: times.TimeRangeMetadata{
															TimeRange: expectedTimeRange,
														},
													},
												))
												mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), *expectedDataType, expectedTimeRange, &oura.Pagination{NextToken: expectedNextToken}, gomock.Not(gomock.Nil())).Return(&oura.DataResponse{Data: expectedData}, nil)
												expectedDataRawCreate := &dataRaw.Create{
													Metadata: map[string]any{
														"dataType":  *expectedDataType,
														"timeRange": timesTest.NewObjectFromTimeRange(expectedTimeRange, test.ObjectFormatJSON),
													},
													MediaType:      pointer.From(net.MediaTypeJSON),
													ArchivableTime: pointer.From(now),
												}
												mockDataRawClient.EXPECT().
													Create(gomock.Not(gomock.Nil()), userID, *dataSrc.DataSetID, expectedDataRawCreate, gomock.Not(gomock.Nil())).
													DoAndReturn(func(_ context.Context, _ string, _ string, dataRawCreate *dataRaw.Create, reader io.Reader) (*dataRaw.Raw, error) {
														bites, err := io.ReadAll(reader)
														Expect(err).ToNot(HaveOccurred())
														Expect(bites).To(MatchJSON(test.Must(json.Marshal(oura.DataMap{*expectedDataType: expectedData}))))
														return dataRw, nil
													})
												mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataSourceID, nil, &dataSource.Update{LastImportTime: pointer.From(dataRw.CreatedTime), Metadata: &dataSrc.Metadata}).Return(nil, testErr)
												Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
											})

											It("returns failing process result if processing update fails", func() {
												testErr := errorsTest.RandomError()
												dataRw := test.Must(metadata.WithMetadata(
													dataRawTest.RandomRaw(test.AllowOptionals()),
													&ouraData.Metadata{
														DataType: *expectedDataType,
														TimeRangeMetadata: times.TimeRangeMetadata{
															TimeRange: expectedTimeRange,
														},
													},
												))
												mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), *expectedDataType, expectedTimeRange, &oura.Pagination{NextToken: expectedNextToken}, gomock.Not(gomock.Nil())).Return(&oura.DataResponse{Data: expectedData}, nil)
												expectedDataRawCreate := &dataRaw.Create{
													Metadata: map[string]any{
														"dataType":  *expectedDataType,
														"timeRange": timesTest.NewObjectFromTimeRange(expectedTimeRange, test.ObjectFormatJSON),
													},
													MediaType:      pointer.From(net.MediaTypeJSON),
													ArchivableTime: pointer.From(now),
												}
												mockDataRawClient.EXPECT().
													Create(gomock.Not(gomock.Nil()), userID, *dataSrc.DataSetID, expectedDataRawCreate, gomock.Not(gomock.Nil())).
													DoAndReturn(func(_ context.Context, _ string, _ string, dataRawCreate *dataRaw.Create, reader io.Reader) (*dataRaw.Raw, error) {
														bites, err := io.ReadAll(reader)
														Expect(err).ToNot(HaveOccurred())
														Expect(bites).To(MatchJSON(test.Must(json.Marshal(oura.DataMap{*expectedDataType: expectedData}))))
														return dataRw, nil
													})
												mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataSourceID, nil, &dataSource.Update{LastImportTime: pointer.From(dataRw.CreatedTime), Metadata: &dataSrc.Metadata}).Return(dataSrc, nil)
												(*dataTypeStartTimes)[*expectedDataType] = expectedStartTime
												delete(*dataTypeNextTokens, *expectedDataType)
												expectedProcessingUpdate := work.ProcessingUpdate{
													Metadata: map[string]any{
														"providerSessionId":  providerSessionID,
														"dataTypeStartTimes": ouraDataWorkPeriodicTest.NewObjectFromDataTypeStartTimes(dataTypeStartTimes, test.ObjectFormatJSON),
														"dataTypeNextTokens": ouraDataWorkPeriodicTest.NewObjectFromDataTypeNextTokens(dataTypeNextTokens, test.ObjectFormatJSON),
													},
												}
												mockProcessingUpdater.EXPECT().ProcessingUpdate(gomock.Not(gomock.Nil()), expectedProcessingUpdate).Return(nil, testErr)
												Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
											})
										})
									})
								})

								Context("it completes successfully after paging through the data types", func() {
									BeforeEach(func() {
										expectedDataType = nil
										expectedNextToken = nil
									})

									paginate(func() {
										It("returns pending process result if successful", func() {
											Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchPendingProcessResult(
												MatchAllFields(Fields{
													"ProcessingAvailableTime": BeTemporally("~", time.Now().Add(ouraDataWorkPeriodic.PendingAvailableDuration), time.Second),
													"ProcessingPriority":      Equal(0),
													"ProcessingTimeout":       Equal(int(ouraDataWorkPeriodic.ProcessingTimeout.Seconds())),
													"Metadata": Equal(map[string]any{
														"providerSessionId":  providerSessionID,
														"dataTypeStartTimes": ouraDataWorkPeriodicTest.NewObjectFromDataTypeStartTimes(dataTypeStartTimes, test.ObjectFormatJSON),
														"dataTypeNextTokens": ouraDataWorkPeriodicTest.NewObjectFromDataTypeNextTokens(dataTypeNextTokens, test.ObjectFormatJSON),
													}),
												}),
											))
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
})
