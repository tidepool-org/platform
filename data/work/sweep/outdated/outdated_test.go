package outdated_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	dataWorkSweepOutdated "github.com/tidepool-org/platform/data/work/sweep/outdated"
	dataWorkSweepOutdatedTest "github.com/tidepool-org/platform/data/work/sweep/outdated/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	summaryStore "github.com/tidepool-org/platform/summary/store"
	summaryTypes "github.com/tidepool-org/platform/summary/types"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

func TestSuite(t *testing.T) {
	test.Test(t)
}

var _ = Describe("Outdated", func() {
	var controller *gomock.Controller
	var workClient *workTest.MockClient
	var summaries *dataWorkSweepOutdatedTest.MockSummaries
	var processor *dataWorkSweepOutdated.Processor
	var ctx context.Context
	var wrk *work.Work
	var outdatedSince time.Time

	process := func() *work.ProcessResult {
		return processor.Process(ctx, wrk, workTest.NewMockProcessingUpdater(controller))
	}

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		workClient = workTest.NewMockClient(controller)
		summaries = dataWorkSweepOutdatedTest.NewMockSummaries(controller)
		ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		outdatedSince = time.Now().UTC().Add(-time.Hour)
		wrk = &work.Work{
			ID:                workTest.RandomID(),
			Type:              dataWorkSweepOutdated.Type,
			ProcessingTimeout: int(dataWorkSweepOutdated.ProcessingTimeout.Seconds()),
			State:             work.StateProcessing,
		}

		var err error
		processor, err = dataWorkSweepOutdated.NewProcessor(dataWorkSweepOutdated.Dependencies{
			Dependencies: workBase.Dependencies{WorkClient: workClient},
			Summaries:    summaries,
		})
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		controller.Finish()
	})

	It("waits to run again when nothing is outdated", func() {
		summaries.EXPECT().ListOutdated(gomock.Any(), gomock.Any()).Return(nil, nil)

		result := process()
		Expect(result.Result).To(Equal(work.ResultPending))
		Expect(result.PendingUpdate.ProcessingAvailableTime).To(BeTemporally("~", time.Now().Add(dataWorkSweepOutdated.PendingAvailableDuration), time.Second))
	})

	Context("with outdated summaries", func() {
		var userID string

		BeforeEach(func() {
			userID = userTest.RandomUserID()
		})

		// Reported on the first request, then nothing, so the run stops rather than repeating
		expectListOnce := func(outdated ...summaryStore.OutdatedSummary) {
			reported := false
			summaries.EXPECT().ListOutdated(gomock.Any(), dataWorkSweepOutdated.BatchSize).DoAndReturn(
				func(_ context.Context, _ int) ([]summaryStore.OutdatedSummary, error) {
					if reported {
						return nil, nil
					}
					reported = true
					return outdated, nil
				}).Times(2)
		}

		// The work is created before the mark is cleared, so that a failure between the two reports the
		// change twice rather than not at all
		It("reports the change as work before clearing the mark", func() {
			expectListOnce(summaryStore.OutdatedSummary{UserID: userID, Type: summaryTypes.SummaryTypeCGM, OutdatedSince: outdatedSince})
			gomock.InOrder(
				workClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&work.Work{}, nil),
				summaries.EXPECT().ClearOutdated(gomock.Any(), userID, summaryTypes.SummaryTypeCGM, outdatedSince).Return(nil),
			)

			Expect(process().Result).To(Equal(work.ResultPending))
		})

		// A user with more than one summary marked is reported once per mark, and every mark is
		// cleared with its own report — the work created collapses into one calculation on pickup, so
		// nothing groups them here
		It("reports each mark of a user, and clears each", func() {
			expectListOnce(
				summaryStore.OutdatedSummary{UserID: userID, Type: summaryTypes.SummaryTypeCGM, OutdatedSince: outdatedSince},
				summaryStore.OutdatedSummary{UserID: userID, Type: summaryTypes.SummaryTypeBGM, OutdatedSince: outdatedSince},
				summaryStore.OutdatedSummary{UserID: userID, Type: summaryTypes.SummaryTypeContinuous, OutdatedSince: outdatedSince},
			)
			workClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&work.Work{}, nil).Times(3)
			summaries.EXPECT().ClearOutdated(gomock.Any(), userID, gomock.Any(), outdatedSince).Return(nil).Times(3)

			Expect(process().Result).To(Equal(work.ResultPending))
		})

		It("reports the change as data added, which requests no synchronization", func() {
			expectListOnce(summaryStore.OutdatedSummary{UserID: userID, Type: summaryTypes.SummaryTypeCGM, OutdatedSince: outdatedSince})
			workClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, create *work.Create) (*work.Work, error) {
					Expect(create.Metadata).To(HaveKeyWithValue("reasons", ConsistOf(dataWorkPostprocess.ReasonDataAdded)))
					Expect(create.Metadata).To(HaveKeyWithValue("userId", userID))
					return &work.Work{}, nil
				})
			summaries.EXPECT().ClearOutdated(gomock.Any(), userID, gomock.Any(), gomock.Any()).Return(nil)

			Expect(process().Result).To(Equal(work.ResultPending))
		})

		It("retries when the work cannot be created, without clearing the mark", func() {
			summaries.EXPECT().ListOutdated(gomock.Any(), gomock.Any()).Return(
				[]summaryStore.OutdatedSummary{{UserID: userID, Type: summaryTypes.SummaryTypeCGM, OutdatedSince: outdatedSince}}, nil)
			workClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errorsTest.RandomError())

			Expect(process().Result).To(Equal(work.ResultFailing))
		})

		It("retries when the mark cannot be cleared", func() {
			summaries.EXPECT().ListOutdated(gomock.Any(), gomock.Any()).Return(
				[]summaryStore.OutdatedSummary{{UserID: userID, Type: summaryTypes.SummaryTypeCGM, OutdatedSince: outdatedSince}}, nil)
			workClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&work.Work{}, nil)
			summaries.EXPECT().ClearOutdated(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errorsTest.RandomError())

			Expect(process().Result).To(Equal(work.ResultFailing))
		})

		It("retries when the summaries cannot be listed", func() {
			summaries.EXPECT().ListOutdated(gomock.Any(), gomock.Any()).Return(nil, errorsTest.RandomError())

			Expect(process().Result).To(Equal(work.ResultFailing))
		})

		// A backlog larger than one run is left to the next, so that a run cannot hold the processor
		// against it indefinitely
		It("reports no more than the page limit in one run", func() {
			summaries.EXPECT().ListOutdated(gomock.Any(), gomock.Any()).Return(
				[]summaryStore.OutdatedSummary{{UserID: userID, Type: summaryTypes.SummaryTypeCGM, OutdatedSince: outdatedSince}}, nil).
				Times(dataWorkSweepOutdated.PageLimit)
			workClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&work.Work{}, nil).Times(dataWorkSweepOutdated.PageLimit)
			summaries.EXPECT().ClearOutdated(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(dataWorkSweepOutdated.PageLimit)

			Expect(process().Result).To(Equal(work.ResultPending))
		})
	})

	Context("Dependencies", func() {
		It("reports the work client is missing", func() {
			_, err := dataWorkSweepOutdated.NewProcessorFactory(dataWorkSweepOutdated.Dependencies{
				Summaries: summaries,
			})
			Expect(err).To(MatchError(ContainSubstring("work client is missing")))
		})

		It("reports the summaries are missing", func() {
			_, err := dataWorkSweepOutdated.NewProcessorFactory(dataWorkSweepOutdated.Dependencies{
				Dependencies: workBase.Dependencies{WorkClient: workClient},
			})
			Expect(err).To(MatchError(ContainSubstring("summaries is missing")))
		})

		It("reports the type, quantity and frequency of the work", func() {
			processorFactory, err := dataWorkSweepOutdated.NewProcessorFactory(dataWorkSweepOutdated.Dependencies{
				Dependencies: workBase.Dependencies{WorkClient: workClient},
				Summaries:    summaries,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(processorFactory.Type()).To(Equal("org.tidepool.data.sweep.outdated"))
			Expect(processorFactory.Quantity()).To(Equal(1))
			Expect(processorFactory.Frequency()).To(Equal(time.Minute))
		})

		It("creates the work as a singleton, so that one sweep runs however many services register it", func() {
			workCreate, err := dataWorkSweepOutdated.NewWorkCreate()
			Expect(err).ToNot(HaveOccurred())
			Expect(workCreate.DeduplicationID).ToNot(BeNil())
			Expect(*workCreate.DeduplicationID).To(Equal(work.DeduplicationIDSingleton))
		})
	})
})
