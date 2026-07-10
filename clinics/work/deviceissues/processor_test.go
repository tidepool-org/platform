package deviceissues_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	clinicsWorkDeviceIssues "github.com/tidepool-org/platform/clinics/work/deviceissues"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("processor", func() {
	It("PendingRetryDuration is expected", func() {
		Expect(clinicsWorkDeviceIssues.PendingRetryDuration).To(Equal(15 * time.Minute))
	})

	It("FailingRetryDuration is expected", func() {
		Expect(clinicsWorkDeviceIssues.FailingRetryDuration).To(Equal(time.Minute))
	})

	It("FailingRetryDurationJitter is expected", func() {
		Expect(clinicsWorkDeviceIssues.FailingRetryDurationJitter).To(Equal(10 * time.Second))
	})

	Context("with dependencies", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockClinicClient *clinicsTest.MockClient
		var dependencies clinicsWorkDeviceIssues.Dependencies

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockClinicClient = clinicsTest.NewMockClient(mockController)
			dependencies = clinicsWorkDeviceIssues.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				ClinicClient: mockClinicClient,
			}
		})

		Context("NewProcessor", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.WorkClient = nil
				processor, err := clinicsWorkDeviceIssues.NewProcessor(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns successfully", func() {
				processor, err := clinicsWorkDeviceIssues.NewProcessor(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})

			Context("with processor", func() {
				var wrk *work.Work
				var mockProcessingUpdater *workTest.MockProcessingUpdater
				var processor *clinicsWorkDeviceIssues.Processor

				BeforeEach(func() {
					wrkCreate, err := clinicsWorkDeviceIssues.NewWorkCreate()
					Expect(err).ToNot(HaveOccurred())
					Expect(wrkCreate).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(wrkCreate, work.StateProcessing)
					mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)
					processor, err = clinicsWorkDeviceIssues.NewProcessor(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processor).ToNot(BeNil())
				})

				Context("Process", func() {
					It("returns failing if unable to update device issues", func() {
						testErr := errorsTest.RandomError()
						mockClinicClient.EXPECT().UpdateDeviceIssues(gomock.Not(gomock.Nil())).Return(testErr)
						Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
					})

					It("returns pending on success, rescheduled after the pending retry duration", func() {
						mockClinicClient.EXPECT().UpdateDeviceIssues(gomock.Not(gomock.Nil())).Return(nil)
						Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchPendingProcessResult(
							MatchAllFields(Fields{
								"ProcessingAvailableTime": BeTemporally("~", time.Now().Add(clinicsWorkDeviceIssues.PendingRetryDuration), time.Second),
								"ProcessingPriority":      Equal(0),
								"ProcessingTimeout":       Equal(int(clinicsWorkDeviceIssues.ProcessingTimeout.Seconds())),
								"Metadata":                Equal(map[string]any{}),
							}),
						))
					})
				})
			})
		})
	})
})
