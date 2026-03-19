package historic_test

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
					datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
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
						datum.ProviderSessionID = pointer.FromString("")
						datum.TimeRange = timesTest.RandomTimeRange()
						datum.TimeRange.From = pointer.FromTime(time.Time{})
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
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraDataWorkHistoric.Dependencies

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraDataWorkHistoric.Dependencies{
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
				var userID string
				var ouraUserID string
				var providerSessionID string
				var workTimeRange times.TimeRange
				var wrk *work.Work
				var mockProcessingUpdater *workTest.MockProcessingUpdater
				var processor *ouraDataWorkHistoric.Processor

				BeforeEach(func() {
					userID = userTest.RandomUserID()
					ouraUserID = ouraTest.RandomUserID()
					providerSessionID = authTest.RandomProviderSessionID()
					workTimeRange = *timesTest.RandomTimeRange(test.AllowOptional())
					wrkCreate, err := ouraDataWorkHistoric.NewWorkCreate(providerSessionID, workTimeRange)
					Expect(err).ToNot(HaveOccurred())
					Expect(wrkCreate).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(wrkCreate, work.StateProcessing)
					mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)
					processor, err = ouraDataWorkHistoric.NewProcessor(dependencies)
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
