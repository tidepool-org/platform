package event_test

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
					datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
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
						datum.ProviderSessionID = pointer.FromString("")
						datum.Event = ouraWebhookTest.RandomEvent()
						datum.Event.EventTime = pointer.FromTime(time.Time{})
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
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraDataWorkEvent.Dependencies

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraDataWorkEvent.Dependencies{
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
				var userID string
				var ouraUserID string
				var providerSessionID string
				var event *ouraWebhook.Event
				var wrk *work.Work
				var mockProcessingUpdater *workTest.MockProcessingUpdater
				var processor *ouraDataWorkEvent.Processor

				BeforeEach(func() {
					userID = userTest.RandomUserID()
					ouraUserID = ouraTest.RandomUserID()
					providerSessionID = authTest.RandomProviderSessionID()
					event = ouraWebhookTest.RandomEvent(test.AllowOptional())
					event.UserID = pointer.FromString(ouraUserID)
					wrkCreate, err := ouraDataWorkEvent.NewWorkCreate(providerSessionID, event)
					Expect(err).ToNot(HaveOccurred())
					Expect(wrkCreate).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(wrkCreate, work.StateProcessing)
					mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)
					processor, err = ouraDataWorkEvent.NewProcessor(dependencies)
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
								ExternalID:  pointer.FromString(ouraUserID),
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
							var dataSourceID string
							var dataSrc *dataSource.Source

							BeforeEach(func() {
								dataSourceID = dataSourceTest.RandomDataSourceID()
								dataSrc = &dataSource.Source{
									ID:                 dataSourceID,
									UserID:             userID,
									ProviderType:       oauth.ProviderType,
									ProviderName:       oura.ProviderName,
									ProviderExternalID: pointer.FromString(ouraUserID),
									ProviderSessionID:  pointer.FromString(providerSessionID),
									State:              dataSource.StateConnected,
									CreatedTime:        time.Now(),
									Revision:           test.RandomInt(),
								}
								mockDataSourceClient.EXPECT().GetFromProviderSession(gomock.Any(), providerSessionID).Return(dataSrc, nil)
							})

							It("returns delete process result when successful", func() {
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchDeleteProcessResult())
							})
						})
					})
				})
			})
		})
	})
})
