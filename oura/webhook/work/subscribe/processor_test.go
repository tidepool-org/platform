package subscribe_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	ouraWebhookWorkSubscribe "github.com/tidepool-org/platform/oura/webhook/work/subscribe"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("processor", func() {
	It("PendingAvailableDuration is expected", func() {
		Expect(ouraWebhookWorkSubscribe.PendingAvailableDuration).To(Equal(24 * time.Hour))
	})

	It("FailingRetryDuration is expected", func() {
		Expect(ouraWebhookWorkSubscribe.FailingRetryDuration).To(Equal(10 * time.Minute))
	})

	It("FailingRetryDurationJitter is expected", func() {
		Expect(ouraWebhookWorkSubscribe.FailingRetryDurationJitter).To(Equal(time.Minute))
	})

	Context("with dependencies", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraWebhookWorkSubscribe.Dependencies

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraWebhookWorkSubscribe.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				OuraClient: mockOuraClient,
			}
		})

		Context("NewProcessor", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.WorkClient = nil
				processor, err := ouraWebhookWorkSubscribe.NewProcessor(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns successfully", func() {
				processor, err := ouraWebhookWorkSubscribe.NewProcessor(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})

			Context("with processor", func() {
				var wrk *work.Work
				var mockProcessingUpdater *workTest.MockProcessingUpdater
				var processor *ouraWebhookWorkSubscribe.Processor

				BeforeEach(func() {
					wrkCreate, err := ouraWebhookWorkSubscribe.NewWorkCreate()
					Expect(err).ToNot(HaveOccurred())
					Expect(wrkCreate).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(wrkCreate, work.StateProcessing)
					mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)
					processor, err = ouraWebhookWorkSubscribe.NewProcessor(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processor).ToNot(BeNil())
				})

				Context("Process", func() {
					It("returns successful process result if able to revoke oauth token", func() {
						Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchPendingProcessResult(
							MatchAllFields(Fields{
								"ProcessingAvailableTime": BeTemporally("~", time.Now().Add(ouraWebhookWorkSubscribe.PendingAvailableDuration), time.Second),
								"ProcessingPriority":      Equal(0),
								"ProcessingTimeout":       Equal(int(ouraWebhookWorkSubscribe.ProcessingTimeout.Seconds())),
								"Metadata":                BeNil(),
							}),
						))
					})
				})
			})
		})
	})
})
