package postprocess_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"go.uber.org/mock/gomock"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	dataWorkPostprocessTest "github.com/tidepool-org/platform/data/work/postprocess/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/user"
	userTest "github.com/tidepool-org/platform/user/test"
	userWork "github.com/tidepool-org/platform/user/work"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("Processor", func() {
	var controller *gomock.Controller
	var workClient *workTest.MockClient
	var summarizers *dataWorkPostprocessTest.MockSummarizers
	var clinicsClient *clinicsTest.MockClient
	var processingUpdater *workTest.MockProcessingUpdater
	var userClient *userTest.MockClient
	var fetchUser func() (*user.User, error)
	var processor *dataWorkPostprocess.Processor
	var ctx context.Context
	var userID string
	var wrk *work.Work

	newWork := func(state string, reasons []string, availableTime time.Time) *work.Work {
		encoded, err := metadata.Encode(&dataWorkPostprocess.Metadata{
			Metadata: userWork.Metadata{UserID: pointer.FromString(userID)},
			Reasons:  reasons,
		})
		Expect(err).ToNot(HaveOccurred())
		return &work.Work{
			ID:                      workTest.RandomID(),
			Type:                    dataWorkPostprocess.Type,
			GroupID:                 pointer.FromString(dataWorkPostprocess.IDFromUserID(userID)),
			SerialID:                pointer.FromString(dataWorkPostprocess.IDFromUserID(userID)),
			ProcessingAvailableTime: availableTime,
			ProcessingTimeout:       int(dataWorkPostprocess.ProcessingTimeout.Seconds()),
			Metadata:                encoded,
			State:                   state,
			Revision:                2,
		}
	}

	// Nothing else is pending for the user, which is the ordinary case. The work being processed is
	// not reported, as only work that is pending is requested.
	expectListNone := func() {
		workClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, filter *work.Filter, _ *page.Pagination) ([]*work.Work, error) {
				Expect(filter.Types).To(PointTo(ConsistOf(dataWorkPostprocess.Type)))
				Expect(filter.State).To(PointTo(Equal(work.StatePending)))
				Expect(filter.GroupID).To(PointTo(Equal(dataWorkPostprocess.IDFromUserID(userID))))
				return nil, nil
			})
	}

	process := func() *work.ProcessResult {
		return processor.Process(ctx, wrk, processingUpdater)
	}

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		workClient = workTest.NewMockClient(controller)
		summarizers = dataWorkPostprocessTest.NewMockSummarizers(controller)
		clinicsClient = clinicsTest.NewMockClient(controller)
		processingUpdater = workTest.NewMockProcessingUpdater(controller)
		userClient = userTest.NewMockClient(controller)
		ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		userID = userTest.RandomUserID()
		wrk = newWork(work.StateProcessing, []string{dataWorkPostprocess.ReasonDataAdded}, time.Now().Add(-time.Minute))

		var err error
		processor, err = dataWorkPostprocess.NewProcessor(dataWorkPostprocess.Dependencies{
			Dependencies:  workBase.Dependencies{WorkClient: workClient},
			Summarizers:   summarizers,
			ClinicsClient: clinicsClient,
			UserClient:    userClient,
		})
		Expect(err).ToNot(HaveOccurred())

		// The user mixin fetches the user reported by the metadata before any step runs
		fetchUser = func() (*user.User, error) { return &user.User{UserID: pointer.FromString(userID)}, nil }
		userClient.EXPECT().Get(gomock.Any(), userID).
			DoAndReturn(func(_ context.Context, _ string) (*user.User, error) { return fetchUser() }).AnyTimes()
	})

	AfterEach(func() {
		controller.Finish()
	})

	It("calculates the summaries and deletes the work", func() {
		expectListNone()
		summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(nil)

		Expect(process().Result).To(Equal(work.ResultDelete))
	})

	DescribeTable("requests a synchronization only for the reasons that report a change is complete",
		func(reasons []string, expectSync bool) {
			wrk = newWork(work.StateProcessing, reasons, time.Now().Add(-time.Minute))
			expectListNone()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(nil)
			if expectSync {
				clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil)
			}

			Expect(process().Result).To(Equal(work.ResultDelete))
		},
		Entry("data added", []string{dataWorkPostprocess.ReasonDataAdded}, false),
		Entry("schema migration", []string{dataWorkPostprocess.ReasonSchemaMigration}, false),
		Entry("upload completed", []string{dataWorkPostprocess.ReasonUploadCompleted}, true),
		Entry("legacy data added", []string{dataWorkPostprocess.ReasonLegacyDataAdded}, true),
		Entry("data added and upload completed",
			[]string{dataWorkPostprocess.ReasonDataAdded, dataWorkPostprocess.ReasonUploadCompleted}, true),
	)

	// A synchronization reports the summaries, so it must not be requested before they are calculated
	It("requests a synchronization only after the summaries are calculated", func() {
		wrk = newWork(work.StateProcessing, []string{dataWorkPostprocess.ReasonUploadCompleted}, time.Now().Add(-time.Minute))
		expectListNone()
		gomock.InOrder(
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(nil),
			clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil),
		)

		Expect(process().Result).To(Equal(work.ResultDelete))
	})

	It("fails without retrying when the work reports no user", func() {
		wrk.Metadata = map[string]any{"reasons": []any{dataWorkPostprocess.ReasonDataAdded}}

		result := process()
		Expect(result.Result).To(Equal(work.ResultFailed))
		Expect(result.FailedUpdate.FailedError.Error).To(MatchError(ContainSubstring("user id is missing")))
	})

	It("fails without retrying when the user no longer exists", func() {
		fetchUser = func() (*user.User, error) { return nil, nil }

		result := process()
		Expect(result.Result).To(Equal(work.ResultFailed))
		Expect(result.FailedUpdate.FailedError.Error).To(MatchError(ContainSubstring("user is missing")))
	})

	It("retries when the user cannot be fetched", func() {
		fetchUser = func() (*user.User, error) { return nil, errorsTest.RandomError() }

		Expect(process().Result).To(Equal(work.ResultFailing))
	})

	DescribeTable("retries when a step fails",
		func(expect func()) {
			wrk = newWork(work.StateProcessing, []string{dataWorkPostprocess.ReasonUploadCompleted}, time.Now().Add(-time.Minute))
			expect()

			result := process()
			Expect(result.Result).To(Equal(work.ResultFailing))
			Expect(result.FailingUpdate.FailingRetryTime).To(BeTemporally(">", time.Now()))
		},
		Entry("listing the work also pending", func() {
			workClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorsTest.RandomError())
		}),
		Entry("calculating the summaries", func() {
			expectListNone()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(errorsTest.RandomError())
		}),
		Entry("requesting a synchronization", func() {
			expectListNone()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(nil)
			clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(errorsTest.RandomError())
		}),
	)

	Context("with work also pending for the user", func() {
		var sibling *work.Work

		expectListWithSibling := func() {
			workClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*work.Work{sibling}, nil)
		}

		// The reasons must be persisted before the work reporting them is deleted, so that a failure
		// between the two leaves them reported twice rather than not at all
		expectProcessingUpdateThenDelete := func() {
			gomock.InOrder(
				processingUpdater.EXPECT().ProcessingUpdate(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, update work.ProcessingUpdate) (*work.Work, error) {
						Expect(update.Metadata).To(HaveKeyWithValue("reasons",
							ConsistOf(dataWorkPostprocess.ReasonDataAdded, dataWorkPostprocess.ReasonUploadCompleted)))
						updated := *wrk
						updated.Metadata = update.Metadata
						updated.Revision = wrk.Revision + 1
						return &updated, nil
					}),
				workClient.EXPECT().Delete(gomock.Any(), sibling.ID, gomock.Nil()).Return(sibling, nil),
			)
		}

		BeforeEach(func() {
			sibling = newWork(work.StatePending, []string{dataWorkPostprocess.ReasonUploadCompleted}, time.Now().Add(-time.Second))
		})

		It("reports its reasons, deletes it, and processes once", func() {
			expectListWithSibling()
			expectProcessingUpdateThenDelete()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(nil)
			clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil)

			Expect(process().Result).To(Equal(work.ResultDelete))
		})

		// An upload reporting only full batches is still in progress, so the summaries are not
		// calculated for every batch so far. Jellyfish defers the same way.
		Context("that is shouldDeffer and reports only a full batch", func() {
			var deferredUntil time.Time

			BeforeEach(func() {
				deferredUntil = time.Now().Add(dataWorkPostprocess.JellyfishQuietDelay)
				wrk = newWork(work.StateProcessing, []string{dataWorkPostprocess.ReasonLegacyDataAdded}, time.Now().Add(-time.Minute))
				sibling = newWork(work.StatePending, []string{dataWorkPostprocess.ReasonLegacyDataAdded}, deferredUntil)
			})

			It("defers rather than calculating the summaries", func() {
				expectListWithSibling()
				processingUpdater.EXPECT().ProcessingUpdate(gomock.Any(), gomock.Any()).Return(wrk, nil)
				workClient.EXPECT().Delete(gomock.Any(), sibling.ID, gomock.Nil()).Return(sibling, nil)

				result := process()
				Expect(result.Result).To(Equal(work.ResultPending))
				Expect(result.PendingUpdate.ProcessingAvailableTime).To(BeTemporally("==", deferredUntil))
			})

			It("calculates the summaries once a reason reports a change is complete", func() {
				sibling = newWork(work.StatePending, []string{dataWorkPostprocess.ReasonUploadCompleted}, deferredUntil)
				expectListWithSibling()
				processingUpdater.EXPECT().ProcessingUpdate(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, update work.ProcessingUpdate) (*work.Work, error) {
						updated := *wrk
						updated.Metadata = update.Metadata
						return &updated, nil
					})
				workClient.EXPECT().Delete(gomock.Any(), sibling.ID, gomock.Nil()).Return(sibling, nil)
				summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(nil)
				clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil)

				Expect(process().Result).To(Equal(work.ResultDelete))
			})
		})

		It("retries when it cannot be deleted", func() {
			expectListWithSibling()
			processingUpdater.EXPECT().ProcessingUpdate(gomock.Any(), gomock.Any()).Return(wrk, nil)
			workClient.EXPECT().Delete(gomock.Any(), sibling.ID, gomock.Nil()).Return(nil, errorsTest.RandomError())

			Expect(process().Result).To(Equal(work.ResultFailing))
		})
	})
})
