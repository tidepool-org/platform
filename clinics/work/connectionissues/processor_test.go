package connectionissues_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	clinicsWorkConnectionIssues "github.com/tidepool-org/platform/clinics/work/connectionissues"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("processor", func() {
	It("PendingAvailableDuration is expected", func() {
		Expect(clinicsWorkConnectionIssues.PendingAvailableDuration).
			To(Equal(5 * time.Minute))
	})

	It("FailingRetryDuration is expected", func() {
		Expect(clinicsWorkConnectionIssues.FailingRetryDuration).To(Equal(1 * time.Minute))
	})

	It("FailingRetryDurationJitter is expected", func() {
		Expect(clinicsWorkConnectionIssues.FailingRetryDurationJitter).
			To(Equal(10 * time.Second))
	})

	It("FailingRetryDurationMaximum is expected", func() {
		Expect(clinicsWorkConnectionIssues.FailingRetryDurationMaximum).
			To(Equal(15 * time.Minute))
	})

	Context("with dependencies", func() {
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockClinicClient *clinicsTest.MockClient
		var dependencies clinicsWorkConnectionIssues.Dependencies

		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockClinicClient = clinicsTest.NewMockClient(mockController)
			dependencies = clinicsWorkConnectionIssues.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				ClinicClient: mockClinicClient,
			}
		})

		Context("NewProcessor", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.ClinicClient = nil
				processor, err := clinicsWorkConnectionIssues.NewProcessor(dependencies)
				Expect(err).
					To(MatchError("dependencies is invalid; clinic client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns successfully", func() {
				processor, err := clinicsWorkConnectionIssues.NewProcessor(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})
		})

		Context("Process", func() {
			var ctx context.Context
			var wrk *work.Work
			var mockProcessingUpdater *workTest.MockProcessingUpdater
			var processor *clinicsWorkConnectionIssues.Processor

			BeforeEach(func() {
				ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
				wrkCreate, err := clinicsWorkConnectionIssues.NewWorkCreate()
				Expect(err).ToNot(HaveOccurred())
				wrk = workTest.NewWorkFromCreateWithState(wrkCreate, work.StateProcessing)
				mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)
				processor, err = clinicsWorkConnectionIssues.NewProcessor(dependencies)
				Expect(err).ToNot(HaveOccurred())
			})

			It("is failing when the clinic call fails", func() {
				testErr := errorsTest.RandomError()
				mockClinicClient.EXPECT().UpdateConnectionIssues(gomock.Not(gomock.Nil())).
					Return(testErr)

				Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).
					To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
			})

			It("schedules the next run when the clinic call succeeds", func() {
				mockClinicClient.EXPECT().UpdateConnectionIssues(gomock.Not(gomock.Nil())).
					Return(nil)

				result := processor.Process(ctx, wrk, mockProcessingUpdater)

				next := time.Now().Add(clinicsWorkConnectionIssues.PendingAvailableDuration)
				Expect(result).To(workTest.MatchPendingProcessResult(MatchAllFields(Fields{
					"ProcessingAvailableTime": BeTemporally("~", next, time.Second),
					"ProcessingPriority":      Equal(0),
					"ProcessingTimeout": Equal(
						int(clinicsWorkConnectionIssues.ProcessingTimeout.Seconds())),
					"Metadata": BeNil(),
				})))
			})
		})
	})
})
