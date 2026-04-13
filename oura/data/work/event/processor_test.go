package event_test

import (
	"context"
	"io"
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
	ouraDataWorkEvent "github.com/tidepool-org/platform/oura/data/work/event"
	ouraDataWorkEventTest "github.com/tidepool-org/platform/oura/data/work/event/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	ouraWebhookTest "github.com/tidepool-org/platform/oura/webhook/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("processor", func() {
	It("FailingRetryDuration is expected", func() {
		Expect(ouraDataWorkEvent.FailingRetryDuration).To(Equal(time.Minute))
	})

	It("FailingRetryDurationJitter is expected", func() {
		Expect(ouraDataWorkEvent.FailingRetryDurationJitter).To(Equal(5 * time.Second))
	})

	Context("Metadata", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *ouraDataWorkEvent.Metadata)) {
				datum := ouraDataWorkEventTest.RandomMetadata(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraDataWorkEventTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraDataWorkEventTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *ouraDataWorkEvent.Metadata) {},
			),
			Entry("empty",
				func(datum *ouraDataWorkEvent.Metadata) {
					*datum = ouraDataWorkEvent.Metadata{}
				},
			),
			Entry("all",
				func(datum *ouraDataWorkEvent.Metadata) {
					datum.ProviderSessionID = pointer.From(authTest.RandomProviderSessionID())
					datum.Event = ouraWebhookTest.RandomEvent()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *ouraDataWorkEvent.Metadata), expectedErrors ...error) {
					expectedDatum := ouraDataWorkEventTest.RandomMetadata(test.AllowOptional())
					object := ouraDataWorkEventTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &ouraDataWorkEvent.Metadata{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *ouraDataWorkEvent.Metadata) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *ouraDataWorkEvent.Metadata) {
						clear(object)
						*expectedDatum = ouraDataWorkEvent.Metadata{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *ouraDataWorkEvent.Metadata) {
						object["providerSessionId"] = true
						object["event"] = true
						expectedDatum.ProviderSessionID = nil
						expectedDatum.Event = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerSessionId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/event"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *ouraDataWorkEvent.Metadata), expectedErrors ...error) {
					datum := ouraDataWorkEventTest.RandomMetadata(test.AllowOptional())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *ouraDataWorkEvent.Metadata) {},
				),
				Entry("multiple errors",
					func(datum *ouraDataWorkEvent.Metadata) {
						datum.ProviderSessionID = pointer.From("")
						datum.Event = ouraWebhookTest.RandomEvent()
						datum.Event.EventTime = pointer.From(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/event/event_time"),
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
		var dependencies ouraDataWorkEvent.Dependencies

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockDataRawClient = dataRawTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraDataWorkEvent.Dependencies{
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
				processor, err := ouraDataWorkEvent.NewProcessor(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns successfully", func() {
				processor, err := ouraDataWorkEvent.NewProcessor(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})

			Context("with processor", func() {
				var now time.Time
				var userID string
				var providerSessionID string
				var ouraUserID string
				var event *ouraWebhook.Event
				var wrk *work.Work
				var processor *ouraDataWorkEvent.Processor
				var mockProcessingUpdater *workTest.MockProcessingUpdater

				BeforeEach(func() {
					now = time.Now()
					userID = userTest.RandomUserID()
					providerSessionID = authTest.RandomProviderSessionID()
					ouraUserID = ouraTest.RandomUserID()
					event = ouraWebhookTest.RandomEvent(test.AllowOptional())
					event.UserID = pointer.From(ouraUserID)
				})

				JustBeforeEach(func() {
					create, err := ouraDataWorkEvent.NewWorkCreate(providerSessionID, event)
					Expect(err).ToNot(HaveOccurred())
					Expect(create).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(create, work.StateProcessing)
					processor, err = ouraDataWorkEvent.NewProcessor(dependencies)
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
							providerSession = authTest.RandomProviderSession(test.AllowOptional())
							providerSession.ID = providerSessionID
							providerSession.UserID = userID
							providerSession.Type = oauth.ProviderType
							providerSession.Name = oura.ProviderName
							providerSession.OAuthToken.Scope = pointer.From(slices.DeleteFunc(oura.Scopes(), func(s string) bool {
								return !slices.Contains(oura.DataTypesForScope(s), *event.DataType) && test.RandomBool()
							}))
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

							It("returns failing process result if data source data set id is missing", func() {
								dataSrc.DataSetID = nil
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailedProcessResultError(MatchError("data set id is missing")))
							})

							It("returns successfully if event data type is not in scope", func() {
								providerSession.OAuthToken.Scope = pointer.From(slices.DeleteFunc(oura.Scopes(), func(s string) bool {
									return slices.Contains(oura.DataTypesForScope(s), *event.DataType)
								}))
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchDeleteProcessResult())
							})

							withCreateDataRaw := func() {
								Context("with create data raw", func() {
									It("returns failing process result if create data raw fails", func() {
										testErr := errorsTest.RandomError()
										mockDataRawClient.EXPECT().
											Create(gomock.Not(gomock.Nil()), userID, *dataSrc.DataSetID, gomock.Not(gomock.Nil()), gomock.Not(gomock.Nil())).
											DoAndReturn(func(_ context.Context, _ string, _ string, dataRawCreate *dataRaw.Create, reader io.Reader) (*dataRaw.Raw, error) {
												Expect(dataRawCreate).To(PointTo(MatchAllFields(Fields{
													"Metadata": Equal(map[string]any{
														"scope": test.AsAnyArray(*providerSession.OAuthToken.Scope),
														"event": ouraWebhookTest.NewObjectFromEvent(event, test.ObjectFormatJSON),
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
											var err error
											createdDataRaw, err = metadata.WithMetadata(
												dataRawTest.RandomRaw(test.AllowOptional()),
												&ouraDataWork.Metadata{
													Scope: providerSession.OAuthToken.Scope,
													EventMetadata: ouraDataWorkEvent.EventMetadata{
														Event: event,
													},
												},
											)
											Expect(err).ToNot(HaveOccurred())
											Expect(createdDataRaw).ToNot(BeNil())
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
															"event": ouraWebhookTest.NewObjectFromEvent(event, test.ObjectFormatJSON),
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
							}

							withGetDatum := func() {
								Context("with get datum", func() {
									It("returns failing process result if get datum fails", func() {
										testErr := errorsTest.RandomError()
										mockOuraClient.EXPECT().GetDatum(gomock.Not(gomock.Nil()), *event.DataType, *event.ObjectID, gomock.Not(gomock.Nil())).Return(nil, testErr)
										Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
									})

									Context("with successful get datum", func() {
										var datum map[string]any

										BeforeEach(func() {
											datum = metadataTest.RandomMetadataMap()
											mockOuraClient.EXPECT().GetDatum(gomock.Not(gomock.Nil()), *event.DataType, *event.ObjectID, gomock.Not(gomock.Nil())).Return(datum, nil)
										})

										withCreateDataRaw()
									})
								})
							}

							Context("with event type create", func() {
								BeforeEach(func() {
									event.EventType = pointer.From(oura.EventTypeCreate)

								})

								withGetDatum()
							})

							Context("with event type update", func() {
								BeforeEach(func() {
									event.EventType = pointer.From(oura.EventTypeUpdate)
								})

								withGetDatum()
							})

							Context("with event type delete", func() {
								BeforeEach(func() {
									event.EventType = pointer.From(oura.EventTypeDelete)
								})

								withCreateDataRaw()
							})
						})
					})
				})
			})
		})
	})
})
