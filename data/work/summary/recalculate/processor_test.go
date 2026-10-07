package recalculate_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	dataWorkSummaryRecalculate "github.com/tidepool-org/platform/data/work/summary/recalculate"
	dataWorkSummaryRecalculateTest "github.com/tidepool-org/platform/data/work/summary/recalculate/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	userTest "github.com/tidepool-org/platform/user/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("Processor", func() {
	var controller *gomock.Controller
	var workClient *workTest.MockClient
	var summaryLister *dataWorkSummaryRecalculateTest.MockSummaryLister
	var processingUpdater *workTest.MockProcessingUpdater
	var processor *dataWorkSummaryRecalculate.Processor
	var ctx context.Context
	var wrk *work.Work

	newWork := func(lastUserID *string) *work.Work {
		encoded, err := metadata.Encode(&dataWorkSummaryRecalculate.Metadata{
			LastUserID: lastUserID,
		})
		Expect(err).ToNot(HaveOccurred())
		return &work.Work{
			ID:                      workTest.RandomID(),
			Type:                    dataWorkSummaryRecalculate.Type,
			DeduplicationID:         pointer.From(dataWorkSummaryRecalculate.ID),
			ProcessingAvailableTime: time.Now().Add(-time.Minute),
			ProcessingTimeout:       int(dataWorkSummaryRecalculate.ProcessingTimeout.Seconds()),
			Metadata:                encoded,
			State:                   work.StateProcessing,
			Revision:                2,
		}
	}

	// Lists the given number of pending postprocess work items
	expectListPending := func(count int) {
		workClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, filter *work.Filter, pagination *page.Pagination) ([]*work.Work, error) {
				Expect(filter.Types).To(PointTo(ConsistOf(dataWorkPostprocess.Type)))
				Expect(filter.State).To(PointTo(Equal(work.StatePending)))
				Expect(filter.GroupID).To(BeNil())
				Expect(pagination.Size).To(Equal(dataWorkSummaryRecalculate.PendingMaximum))
				return make([]*work.Work, count), nil
			})
	}

	// Captures the postprocess work created
	expectCreate := func(count int) *[]*work.Create {
		var created []*work.Create
		workClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, create *work.Create) (*work.Work, error) {
				created = append(created, create)
				return &work.Work{ID: workTest.RandomID()}, nil
			}).Times(count)
		return &created
	}

	decodeLastUserID := func(result *work.ProcessResult) *string {
		Expect(result.PendingUpdate).ToNot(BeNil())
		decoded, err := metadata.Decode[dataWorkSummaryRecalculate.Metadata](ctx,
			result.PendingUpdate.Metadata)
		Expect(err).ToNot(HaveOccurred())
		return decoded.LastUserID
	}

	process := func() *work.ProcessResult {
		return processor.Process(ctx, wrk, processingUpdater)
	}

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		workClient = workTest.NewMockClient(controller)
		summaryLister = dataWorkSummaryRecalculateTest.NewMockSummaryLister(controller)
		processingUpdater = workTest.NewMockProcessingUpdater(controller)
		ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		wrk = newWork(nil)

		var err error
		processor, err = dataWorkSummaryRecalculate.NewProcessor(dataWorkSummaryRecalculate.Dependencies{
			Dependencies:  workBase.Dependencies{WorkClient: workClient},
			SummaryLister: summaryLister,
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error if dependencies is invalid", func() {
		processor, err := dataWorkSummaryRecalculate.NewProcessor(dataWorkSummaryRecalculate.Dependencies{
			Dependencies: workBase.Dependencies{WorkClient: workClient},
		})
		Expect(err).To(MatchError("dependencies is invalid; summary lister is missing"))
		Expect(processor).To(BeNil())
	})

	It("requests the recalculation of the first batch of users", func() {
		userIDs := []string{userTest.RandomUserID(), userTest.RandomUserID()}
		expectListPending(0)
		summaryLister.EXPECT().
			ListUserIDs(gomock.Any(), nil, dataWorkSummaryRecalculate.BatchSize).
			Return(userIDs, nil)
		created := expectCreate(len(userIDs))

		result := process()
		Expect(result.Result).To(Equal(work.ResultPending))
		Expect(result.PendingUpdate.ProcessingAvailableTime).To(BeTemporally("~",
			time.Now().Add(dataWorkSummaryRecalculate.PendingAvailableDuration), time.Second))
		Expect(decodeLastUserID(result)).To(PointTo(Equal(userIDs[1])))

		Expect(*created).To(HaveLen(len(userIDs)))
		for index, create := range *created {
			Expect(create.Type).To(Equal(dataWorkPostprocess.Type))
			Expect(create.ProcessingPriority).To(Equal(dataWorkPostprocess.ProcessingPriorityLow))
			Expect(create.Metadata).To(HaveKeyWithValue("userId", userIDs[index]))
			Expect(create.Metadata).To(HaveKeyWithValue("reasons",
				ConsistOf(dataWorkPostprocess.ReasonSummaryRecalculation)))
		}
	})

	It("requests the recalculation of the users after the last user requested", func() {
		lastUserID := userTest.RandomUserID()
		userIDs := []string{userTest.RandomUserID()}
		wrk = newWork(pointer.From(lastUserID))
		expectListPending(dataWorkSummaryRecalculate.PendingMaximum - 1)
		summaryLister.EXPECT().
			ListUserIDs(gomock.Any(), pointer.From(lastUserID), dataWorkSummaryRecalculate.BatchSize).
			Return(userIDs, nil)
		expectCreate(len(userIDs))

		result := process()
		Expect(result.Result).To(Equal(work.ResultPending))
		Expect(decodeLastUserID(result)).To(PointTo(Equal(userIDs[0])))
	})

	// The work is kept, rather than deleted, so that it deduplicates the work created by any
	// later start of the service
	It("succeeds once there are no more users", func() {
		lastUserID := userTest.RandomUserID()
		wrk = newWork(pointer.From(lastUserID))
		expectListPending(0)
		summaryLister.EXPECT().
			ListUserIDs(gomock.Any(), pointer.From(lastUserID), dataWorkSummaryRecalculate.BatchSize).
			Return([]string{}, nil)

		Expect(process().Result).To(Equal(work.ResultSuccess))
	})

	It("defers without requesting recalculations while postprocess work is backed up", func() {
		lastUserID := userTest.RandomUserID()
		wrk = newWork(pointer.From(lastUserID))
		expectListPending(dataWorkSummaryRecalculate.PendingMaximum)

		result := process()
		Expect(result.Result).To(Equal(work.ResultPending))
		Expect(decodeLastUserID(result)).To(PointTo(Equal(lastUserID)))
	})

	It("is failing if pending postprocess work can't be listed", func() {
		workClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errorsTest.RandomError())

		Expect(process().Result).To(Equal(work.ResultFailing))
	})

	It("is failing if the users can't be listed", func() {
		expectListPending(0)
		summaryLister.EXPECT().
			ListUserIDs(gomock.Any(), nil, dataWorkSummaryRecalculate.BatchSize).
			Return(nil, errorsTest.RandomError())

		Expect(process().Result).To(Equal(work.ResultFailing))
	})

	// The retry requests the whole batch again, which postprocess coalesces
	It("is failing without advancing if a recalculation can't be requested", func() {
		lastUserID := userTest.RandomUserID()
		wrk = newWork(pointer.From(lastUserID))
		expectListPending(0)
		summaryLister.EXPECT().
			ListUserIDs(gomock.Any(), pointer.From(lastUserID), dataWorkSummaryRecalculate.BatchSize).
			Return([]string{userTest.RandomUserID(), userTest.RandomUserID()}, nil)
		gomock.InOrder(
			workClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&work.Work{}, nil),
			workClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errorsTest.RandomError()),
		)

		result := process()
		Expect(result.Result).To(Equal(work.ResultFailing))
		Expect(result.FailingUpdate).ToNot(BeNil())
		decoded, err := metadata.Decode[dataWorkSummaryRecalculate.Metadata](ctx,
			result.FailingUpdate.Metadata)
		Expect(err).ToNot(HaveOccurred())
		Expect(decoded.LastUserID).To(PointTo(Equal(lastUserID)))
	})
})
