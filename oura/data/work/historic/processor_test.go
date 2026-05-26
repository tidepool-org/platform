package historic_test

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
	ouraDataWorkHistoric "github.com/tidepool-org/platform/oura/data/work/historic"
	ouraDataWorkHistoricTest "github.com/tidepool-org/platform/oura/data/work/historic/test"
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
	It("FailingRetryDuration is expected", func() {
		Expect(ouraDataWorkHistoric.FailingRetryDuration).To(Equal(1 * time.Minute))
	})

	It("FailingRetryDurationJitter is expected", func() {
		Expect(ouraDataWorkHistoric.FailingRetryDurationJitter).To(Equal(5 * time.Second))
	})

	Context("DataTypes", func() {
		It("returns expected data types", func() {
			Expect(ouraDataWorkHistoric.DataTypes()).To(Equal(oura.EventDataTypes()))
		})
	})

	It("MetadataKeyDataTypeNextTokens is expected", func() {
		Expect(ouraDataWorkHistoric.MetadataKeyDataTypeNextTokens).To(Equal("dataTypeNextTokens"))
	})

	Context("Metadata", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *ouraDataWorkHistoric.Metadata)) {
				datum := ouraDataWorkHistoricTest.RandomMetadata(test.AllowOptionals())
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
					datum.ProviderSessionMetadata = *providerSessionWorkTest.RandomMetadata()
					datum.TimeRangeMetadata = *timesTest.RandomTimeRangeMetadata()
					datum.DataTypeNextTokens = ouraDataWorkHistoricTest.RandomDataTypeNextTokens(test.AllowOptionals())
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *ouraDataWorkHistoric.Metadata), expectedErrors ...error) {
					expectedDatum := ouraDataWorkHistoricTest.RandomMetadata(test.AllowOptionals())
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
						object["dataTypeNextTokens"] = true
						expectedDatum.ProviderSessionID = nil
						expectedDatum.TimeRange = nil
						expectedDatum.DataTypeNextTokens = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerSessionId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/timeRange"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/dataTypeNextTokens"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *ouraDataWorkHistoric.Metadata), expectedErrors ...error) {
					datum := ouraDataWorkHistoricTest.RandomMetadata(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *ouraDataWorkHistoric.Metadata) {},
				),
				Entry("provider session id missing",
					func(datum *ouraDataWorkHistoric.Metadata) {
						datum.ProviderSessionID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerSessionId"),
				),
				Entry("time range missing",
					func(datum *ouraDataWorkHistoric.Metadata) {
						datum.TimeRange = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/timeRange"),
				),
				Entry("data type next tokens missing",
					func(datum *ouraDataWorkHistoric.Metadata) {
						datum.DataTypeNextTokens = nil
					},
				),
				Entry("data type next tokens empty",
					func(datum *ouraDataWorkHistoric.Metadata) {
						datum.DataTypeNextTokens = &dataWork.StringStringMap{}
					},
				),
				Entry("data type next tokens invalid",
					func(datum *ouraDataWorkHistoric.Metadata) {
						datum.DataTypeNextTokens = &dataWork.StringStringMap{"invalid": pointer.From("")}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", ouraDataWorkHistoric.DataTypes()), "/dataTypeNextTokens/invalid/#"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataTypeNextTokens/invalid"),
				),
				Entry("multiple errors",
					func(datum *ouraDataWorkHistoric.Metadata) {
						datum.ProviderSessionID = nil
						datum.TimeRange = nil
						datum.DataTypeNextTokens = &dataWork.StringStringMap{"invalid": pointer.From("")}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerSessionId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/timeRange"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", ouraDataWorkHistoric.DataTypes()), "/dataTypeNextTokens/invalid/#"),
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
				var expectedTimeRange *times.TimeRange
				var dataTypeNextTokens *dataWork.StringStringMap
				var wrk *work.Work
				var processor *ouraDataWorkHistoric.Processor
				var mockProcessingUpdater *workTest.MockProcessingUpdater

				BeforeEach(func() {
					now = time.Now().UTC()
					userID = userTest.RandomUserID()
					providerSessionID = authTest.RandomProviderSessionID()
					ouraUserID = ouraTest.RandomUserID()
					expectedTimeRange = timesTest.RandomTimeRange(test.AllowOptionals())
					dataTypeNextTokens = test.RandomOptionalPointerWithOptions(ouraDataWorkHistoricTest.RandomDataTypeNextTokens, test.AllowOptionals())
					create, err := ouraDataWorkHistoric.NewWorkCreate(providerSessionID, expectedTimeRange)
					Expect(err).ToNot(HaveOccurred())
					Expect(create).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(create, work.StateProcessing)
					if dataTypeNextTokens != nil {
						wrk.Metadata[ouraDataWorkHistoric.MetadataKeyDataTypeNextTokens] = ouraDataWorkHistoricTest.NewObjectFromDataTypeNextTokens(dataTypeNextTokens, test.ObjectFormatJSON)
					}
					processor, err = ouraDataWorkHistoric.NewProcessor(dependencies)
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
							providerSession.OAuthToken.Scope = pointer.From(test.RandomStringArrayFromArrayWithoutDuplicates(oura.ScopesForDataTypes(ouraDataWorkHistoric.DataTypes())))
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
								var expectedToken *string

								paginate := func(inner func()) {
									Context("it paginates through the data types", func() {
										BeforeEach(func() {
											if dataTypeNextTokens != nil && expectedDataType != nil {
												if nextToken, ok := (*dataTypeNextTokens)[*expectedDataType]; ok && nextToken == nil {
													if test.RandomBool() {
														(*dataTypeNextTokens)[*expectedDataType] = pointer.From(ouraTest.RandomNextToken())
													} else {
														delete(*dataTypeNextTokens, *expectedDataType)
													}
													wrk.Metadata[ouraDataWorkHistoric.MetadataKeyDataTypeNextTokens] = ouraDataWorkHistoricTest.NewObjectFromDataTypeNextTokens(dataTypeNextTokens, test.ObjectFormatJSON)
												}
											}

											for _, dataType := range ouraDataWorkHistoric.DataTypes() {
												if !oura.DataTypeInScopes(dataType, providerSession.OAuthToken.Scope) {
													continue
												}

												expectedPagination := &oura.Pagination{}
												if dataTypeNextTokens != nil {
													if nextToken, ok := (*dataTypeNextTokens)[dataType]; ok {
														expectedPagination.NextToken = nextToken
														if !expectedPagination.HasNext() {
															continue
														}
													}
												}

												nextTokens := test.RandomArrayWithLength(test.RandomIntFromRange(0, 2), func() *string { return pointer.From(ouraTest.RandomNextToken()) })
												if expectedDataType != nil && *expectedDataType == dataType {
													nextTokens = append(nextTokens, expectedToken)
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
														Create(gomock.Not(gomock.Nil()), userID, *dataSrc.DataSetID, gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil())).
														DoAndReturn(func(_ context.Context, _ string, _ string, dataRawCreate *dataRaw.Create, reader io.Reader) (*dataRaw.Raw, error) {
															Expect(dataRawCreate).To(PointTo(MatchAllFields(Fields{
																"Metadata": Equal(map[string]any{
																	"dataType":  dataType,
																	"timeRange": timesTest.NewObjectFromTimeRange(expectedTimeRange, test.ObjectFormatJSON),
																}),
																"DigestMD5":      BeNil(),
																"DigestSHA256":   BeNil(),
																"MediaType":      PointTo(Equal(net.MediaTypeJSON)),
																"ArchivableTime": PointTo(Equal(now)),
															})))
															bites, err := io.ReadAll(reader)
															Expect(err).ToNot(HaveOccurred())
															Expect(bites).To(MatchJSON(test.Must(json.Marshal(oura.DataMap{dataType: expectedData}))))
															return dataRawResult, nil
														})

													mockDataSourceClient.EXPECT().
														Update(gomock.Not(gomock.Nil()), dataSourceID, nil, gomock.Not(gomock.Nil())).
														DoAndReturn(func(_ context.Context, _ string, _ *request.Condition, dataSrcUpdate *dataSource.Update) (*dataSource.Source, error) {
															Expect(dataSrcUpdate).To(PointTo(MatchFields(IgnoreExtras, Fields{
																"LastImportTime": PointTo(Equal(dataRawResult.CreatedTime)),
																"Metadata":       PointTo(Equal(dataSrc.Metadata)),
															})))
															dataSrcUpdated := dataSourceTest.CloneSource(dataSrc)
															dataSrcUpdated.LastImportTime = dataSrcUpdate.LastImportTime
															if dataSrcUpdate.Metadata != nil {
																dataSrcUpdated.Metadata = *dataSrcUpdate.Metadata
															}
															return dataSrcUpdated, nil
														})

													expectedPagination = &dataResult.Pagination

													if dataTypeNextTokens == nil {
														dataTypeNextTokens = &dataWork.StringStringMap{}
													}
													(*dataTypeNextTokens)[dataType] = nextToken
													expectedProcessingUpdate := work.ProcessingUpdate{
														Metadata: map[string]any{
															"providerSessionId":  providerSessionID,
															"timeRange":          timesTest.NewObjectFromTimeRange(expectedTimeRange, test.ObjectFormatJSON),
															"dataTypeNextTokens": ouraDataWorkHistoricTest.NewObjectFromDataTypeNextTokens(dataTypeNextTokens, test.ObjectFormatJSON),
														},
													}
													mockProcessingUpdater.EXPECT().ProcessingUpdate(gomock.Not(gomock.Nil()), expectedProcessingUpdate).Return(wrk, nil)
												}

												if expectedDataType != nil && *expectedDataType == dataType {
													break
												}
											}
										})

										inner()
									})
								}

								Context("it errors as expected after paging through the data types", func() {
									var expectedData oura.Data

									BeforeEach(func() {
										dataTypes := slices.DeleteFunc(ouraDataWorkHistoric.DataTypes(), func(dataType string) bool {
											return !oura.DataTypeInScopes(dataType, providerSession.OAuthToken.Scope)
										})
										expectedDataType = pointer.From(test.RandomStringFromArray(dataTypes))
										expectedToken = pointer.From(ouraTest.RandomNextToken())
										expectedData = ouraTest.RandomData()
									})

									paginate(func() {
										It("returns failing process result if get data fails", func() {
											testErr := errorsTest.RandomError()
											mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), *expectedDataType, expectedTimeRange, &oura.Pagination{NextToken: expectedToken}, gomock.Not(gomock.Nil())).Return(nil, testErr)
											Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
										})

										It("returns failing process result if get data response is missing", func() {
											mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), *expectedDataType, expectedTimeRange, &oura.Pagination{NextToken: expectedToken}, gomock.Not(gomock.Nil())).Return(nil, nil)
											Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(fmt.Sprintf("data response for data type %q is missing", *expectedDataType))))
										})

										It("returns failing process result if create data raw fails", func() {
											testErr := errorsTest.RandomError()
											mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), *expectedDataType, expectedTimeRange, &oura.Pagination{NextToken: expectedToken}, gomock.Not(gomock.Nil())).Return(&oura.DataResponse{Data: expectedData}, nil)
											mockDataRawClient.EXPECT().
												Create(gomock.Not(gomock.Nil()), userID, *dataSrc.DataSetID, gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil())).
												DoAndReturn(func(_ context.Context, _ string, _ string, dataRawCreate *dataRaw.Create, reader io.Reader) (*dataRaw.Raw, error) {
													Expect(dataRawCreate).To(PointTo(MatchAllFields(Fields{
														"Metadata": Equal(map[string]any{
															"dataType":  *expectedDataType,
															"timeRange": timesTest.NewObjectFromTimeRange(expectedTimeRange, test.ObjectFormatJSON),
														}),
														"DigestMD5":      BeNil(),
														"DigestSHA256":   BeNil(),
														"MediaType":      PointTo(Equal(net.MediaTypeJSON)),
														"ArchivableTime": PointTo(Equal(now)),
													})))
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
											mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), *expectedDataType, expectedTimeRange, &oura.Pagination{NextToken: expectedToken}, gomock.Not(gomock.Nil())).Return(&oura.DataResponse{Data: expectedData}, nil)
											mockDataRawClient.EXPECT().
												Create(gomock.Not(gomock.Nil()), userID, *dataSrc.DataSetID, gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil())).
												DoAndReturn(func(_ context.Context, _ string, _ string, dataRawCreate *dataRaw.Create, reader io.Reader) (*dataRaw.Raw, error) {
													Expect(dataRawCreate).To(PointTo(MatchAllFields(Fields{
														"Metadata": Equal(map[string]any{
															"dataType":  *expectedDataType,
															"timeRange": timesTest.NewObjectFromTimeRange(expectedTimeRange, test.ObjectFormatJSON),
														}),
														"DigestMD5":      BeNil(),
														"DigestSHA256":   BeNil(),
														"MediaType":      PointTo(Equal(net.MediaTypeJSON)),
														"ArchivableTime": PointTo(Equal(now)),
													})))
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
											mockOuraClient.EXPECT().GetData(gomock.Not(gomock.Nil()), *expectedDataType, expectedTimeRange, &oura.Pagination{NextToken: expectedToken}, gomock.Not(gomock.Nil())).Return(&oura.DataResponse{Data: expectedData}, nil)
											mockDataRawClient.EXPECT().
												Create(gomock.Not(gomock.Nil()), userID, *dataSrc.DataSetID, gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil())).
												DoAndReturn(func(_ context.Context, _ string, _ string, dataRawCreate *dataRaw.Create, reader io.Reader) (*dataRaw.Raw, error) {
													Expect(dataRawCreate).To(PointTo(MatchAllFields(Fields{
														"Metadata": Equal(map[string]any{
															"dataType":  *expectedDataType,
															"timeRange": timesTest.NewObjectFromTimeRange(expectedTimeRange, test.ObjectFormatJSON),
														}),
														"DigestMD5":      BeNil(),
														"DigestSHA256":   BeNil(),
														"MediaType":      PointTo(Equal(net.MediaTypeJSON)),
														"ArchivableTime": PointTo(Equal(now)),
													})))
													bites, err := io.ReadAll(reader)
													Expect(err).ToNot(HaveOccurred())
													Expect(bites).To(MatchJSON(test.Must(json.Marshal(oura.DataMap{*expectedDataType: expectedData}))))
													return dataRw, nil
												})
											mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataSourceID, nil, &dataSource.Update{LastImportTime: pointer.From(dataRw.CreatedTime), Metadata: &dataSrc.Metadata}).Return(dataSrc, nil)
											(*dataTypeNextTokens)[*expectedDataType] = nil
											expectedProcessingUpdate := work.ProcessingUpdate{
												Metadata: map[string]any{
													"providerSessionId":  providerSessionID,
													"timeRange":          timesTest.NewObjectFromTimeRange(expectedTimeRange, test.ObjectFormatJSON),
													"dataTypeNextTokens": ouraDataWorkHistoricTest.NewObjectFromDataTypeNextTokens(dataTypeNextTokens, test.ObjectFormatJSON),
												},
											}
											mockProcessingUpdater.EXPECT().ProcessingUpdate(gomock.Not(gomock.Nil()), expectedProcessingUpdate).Return(nil, testErr)
											Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
										})
									})
								})

								Context("it completes successfully after paging through the data types", func() {
									BeforeEach(func() {
										expectedDataType = nil
										expectedToken = nil
									})

									paginate(func() {
										It("returns delete process result if successful", func() {
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
	})
})
