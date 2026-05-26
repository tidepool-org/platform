package personal_test

import (
	"context"
	"encoding/json"
	"io"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	providerSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/crypto"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataRawTest "github.com/tidepool-org/platform/data/raw/test"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/oura"
	ouraData "github.com/tidepool-org/platform/oura/data"
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
	ouraDataWorkPersonal "github.com/tidepool-org/platform/oura/data/work/personal"
	ouraDataWorkPersonalTest "github.com/tidepool-org/platform/oura/data/work/personal/test"
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
	It("PendingAvailableDuration is expected", func() {
		Expect(ouraDataWorkPersonal.PendingAvailableDuration).To(Equal(12 * time.Hour))
	})

	It("FailingRetryDuration is expected", func() {
		Expect(ouraDataWorkPersonal.FailingRetryDuration).To(Equal(1 * time.Minute))
	})

	It("FailingRetryDurationJitter is expected", func() {
		Expect(ouraDataWorkPersonal.FailingRetryDurationJitter).To(Equal(5 * time.Second))
	})

	It("MetadataKeyPreviousHash is expected", func() {
		Expect(ouraDataWorkPersonal.MetadataKeyPreviousHash).To(Equal("previousHash"))
	})

	Context("Metadata", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *ouraDataWorkPersonal.Metadata)) {
				datum := ouraDataWorkPersonalTest.RandomMetadata(test.AllowOptionals())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraDataWorkPersonalTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraDataWorkPersonalTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *ouraDataWorkPersonal.Metadata) {},
			),
			Entry("empty",
				func(datum *ouraDataWorkPersonal.Metadata) {
					*datum = ouraDataWorkPersonal.Metadata{}
				},
			),
			Entry("all",
				func(datum *ouraDataWorkPersonal.Metadata) {
					datum.ProviderSessionMetadata = *providerSessionWorkTest.RandomMetadata()
					datum.PreviousHash = pointer.From(cryptoTest.RandomBase64EncodedSHA256Hash())
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *ouraDataWorkPersonal.Metadata), expectedErrors ...error) {
					expectedDatum := ouraDataWorkPersonalTest.RandomMetadata(test.AllowOptionals())
					object := ouraDataWorkPersonalTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &ouraDataWorkPersonal.Metadata{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *ouraDataWorkPersonal.Metadata) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *ouraDataWorkPersonal.Metadata) {
						clear(object)
						*expectedDatum = ouraDataWorkPersonal.Metadata{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *ouraDataWorkPersonal.Metadata) {
						object["providerSessionId"] = true
						object["previousHash"] = true
						expectedDatum.ProviderSessionID = nil
						expectedDatum.PreviousHash = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerSessionId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/previousHash"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *ouraDataWorkPersonal.Metadata), expectedErrors ...error) {
					datum := ouraDataWorkPersonalTest.RandomMetadata(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *ouraDataWorkPersonal.Metadata) {},
				),
				Entry("provider session id missing",
					func(datum *ouraDataWorkPersonal.Metadata) {
						datum.ProviderSessionID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerSessionId"),
				),
				Entry("previous hash missing",
					func(datum *ouraDataWorkPersonal.Metadata) {
						datum.PreviousHash = nil
					},
				),
				Entry("previous hash invalid",
					func(datum *ouraDataWorkPersonal.Metadata) {
						datum.PreviousHash = pointer.From("invalid")
					},
					errorsTest.WithPointerSource(crypto.ErrorValueStringAsBase64EncodedSHA256HashNotValid("invalid"), "/previousHash"),
				),
				Entry("multiple errors",
					func(datum *ouraDataWorkPersonal.Metadata) {
						datum.ProviderSessionID = nil
						datum.PreviousHash = pointer.From("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerSessionId"),
					errorsTest.WithPointerSource(crypto.ErrorValueStringAsBase64EncodedSHA256HashNotValid("invalid"), "/previousHash"),
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
				processor, err := ouraDataWorkPersonal.NewProcessor(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns successfully", func() {
				processor, err := ouraDataWorkPersonal.NewProcessor(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})

			Context("with processor", func() {
				var now time.Time
				var userID string
				var providerSessionID string
				var ouraUserID string
				var wrk *work.Work
				var processor *ouraDataWorkPersonal.Processor
				var mockProcessingUpdater *workTest.MockProcessingUpdater

				BeforeEach(func() {
					now = time.Now().UTC()
					userID = userTest.RandomUserID()
					providerSessionID = authTest.RandomProviderSessionID()
					ouraUserID = ouraTest.RandomUserID()
					create, err := ouraDataWorkPersonal.NewWorkCreate(providerSessionID)
					Expect(err).ToNot(HaveOccurred())
					Expect(create).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(create, work.StateProcessing)
					processor, err = ouraDataWorkPersonal.NewProcessor(dependencies)
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
							providerSession.OAuthToken.Scope = pointer.From(test.RandomStringArrayFromArrayWithoutDuplicates(oura.ScopesForDataType(oura.DataTypePersonalInfo)))
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

							It("returns pending process result if data type scope is not authorized", func() {
								providerSession.OAuthToken.Scope = pointer.From(test.RandomStringArrayFromArrayWithoutDuplicates(oura.ScopesForDataType(oura.DataTypeSleep)))
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchPendingProcessResult(
									MatchAllFields(Fields{
										"ProcessingAvailableTime": BeTemporally("~", time.Now().Add(ouraDataWorkPersonal.PendingAvailableDuration), time.Second),
										"ProcessingPriority":      Equal(0),
										"ProcessingTimeout":       Equal(int(ouraDataWorkPersonal.ProcessingTimeout.Seconds())),
										"Metadata": Equal(map[string]any{
											"providerSessionId": providerSessionID,
										}),
									}),
								))
							})

							It("returns failing process result if get personal info fails", func() {
								testErr := errorsTest.RandomError()
								mockOuraClient.EXPECT().GetPersonalInfo(gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil())).Return(nil, testErr)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
							})

							Context("with datum", func() {
								var expectedDatum *oura.PersonalInfo
								var expectedHash string

								BeforeEach(func() {
									expectedDatum = ouraTest.RandomPersonalInfo()
									expectedHash = test.Must(expectedDatum.Hash())
									mockOuraClient.EXPECT().GetPersonalInfo(gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil())).Return(expectedDatum, nil)
								})

								It("returns pending process result if hash matches previous hash", func() {
									wrk.Metadata[ouraDataWorkPersonal.MetadataKeyPreviousHash] = expectedHash
									Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchPendingProcessResult(
										MatchAllFields(Fields{
											"ProcessingAvailableTime": BeTemporally("~", time.Now().Add(ouraDataWorkPersonal.PendingAvailableDuration), time.Second),
											"ProcessingPriority":      Equal(0),
											"ProcessingTimeout":       Equal(int(ouraDataWorkPersonal.ProcessingTimeout.Seconds())),
											"Metadata": Equal(map[string]any{
												"providerSessionId": providerSessionID,
												"previousHash":      expectedHash,
											}),
										}),
									))
								})

								Context("with data raw create", func() {
									var expectedTimeRange *times.TimeRange
									var expectedDataRawCreate *dataRaw.Create

									BeforeEach(func() {
										expectedTimeRange = &times.TimeRange{
											To: pointer.From(now),
										}
										expectedDataRawCreate = &dataRaw.Create{
											Metadata: map[string]any{
												"dataType":  oura.DataTypePersonalInfo,
												"timeRange": timesTest.NewObjectFromTimeRange(expectedTimeRange, test.ObjectFormatJSON),
											},
											MediaType:      pointer.From(net.MediaTypeJSON),
											ArchivableTime: pointer.From(now),
										}
									})

									It("returns failing process result if create data raw fails", func() {
										testErr := errorsTest.RandomError()
										mockDataRawClient.EXPECT().
											Create(gomock.Not(gomock.Nil()), userID, *dataSrc.DataSetID, expectedDataRawCreate, gomock.Not(gomock.Nil())).
											DoAndReturn(func(_ context.Context, _ string, _ string, dataRawCreate *dataRaw.Create, reader io.Reader) (*dataRaw.Raw, error) {
												bites, err := io.ReadAll(reader)
												Expect(err).ToNot(HaveOccurred())
												Expect(bites).To(MatchJSON(test.Must(json.Marshal(map[string][]any{oura.DataTypePersonalInfo: {expectedDatum}}))))
												return nil, testErr
											})
										Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
									})

									Context("with data raw", func() {
										var expectedDataRaw *dataRaw.Raw

										BeforeEach(func() {
											expectedDataRaw = test.Must(metadata.WithMetadata(
												dataRawTest.RandomRaw(test.AllowOptionals()),
												&ouraData.Metadata{
													DataType: oura.DataTypePersonalInfo,
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
													Expect(bites).To(MatchJSON(test.Must(json.Marshal(map[string][]any{oura.DataTypePersonalInfo: {expectedDatum}}))))
													return expectedDataRaw, nil
												})
										})

										Context("with data source update", func() {
											var expectedDataSourceUpdate *dataSource.Update

											BeforeEach(func() {
												expectedDataSourceUpdate = &dataSource.Update{
													LastImportTime: pointer.From(expectedDataRaw.CreatedTime),
													Metadata:       &dataSrc.Metadata,
												}
											})

											It("returns failing process result if update data source fails", func() {
												testErr := errorsTest.RandomError()
												mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataSourceID, nil, expectedDataSourceUpdate).Return(nil, testErr)
												Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
											})

											Context("with data source", func() {
												BeforeEach(func() {
													mockDataSourceClient.EXPECT().Update(gomock.Not(gomock.Nil()), dataSourceID, nil, expectedDataSourceUpdate).Return(dataSrc, nil)
												})

												It("returns delete process result if successful", func() {
													Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchPendingProcessResult(
														MatchAllFields(Fields{
															"ProcessingAvailableTime": BeTemporally("~", time.Now().Add(ouraDataWorkPersonal.PendingAvailableDuration), time.Second),
															"ProcessingPriority":      Equal(0),
															"ProcessingTimeout":       Equal(int(ouraDataWorkPersonal.ProcessingTimeout.Seconds())),
															"Metadata": Equal(map[string]any{
																"providerSessionId": providerSessionID,
																"previousHash":      expectedHash,
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
	})
})
