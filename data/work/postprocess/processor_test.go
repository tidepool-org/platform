package postprocess_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	clinic "github.com/tidepool-org/clinic/client"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	"github.com/tidepool-org/platform/request"
	summaryTest "github.com/tidepool-org/platform/summary/test"
	summaryTypes "github.com/tidepool-org/platform/summary/types"
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

	newSummariesUpdate := func(updatedTypes ...string) dataWorkPostprocess.SummariesUpdate {
		update := dataWorkPostprocess.SummariesUpdate{
			CGM:          summaryTest.RandomCGMSummary(userID),
			BGM:          summaryTest.RandomBGMSummary(userID),
			UpdatedTypes: updatedTypes,
		}
		update.CGM.ID = primitive.NewObjectID()
		update.BGM.ID = primitive.NewObjectID()
		return update
	}

	// The processing update persists the metadata of the request, so it is echoed back the way the
	// store does
	expectProcessingUpdate := func() {
		processingUpdater.EXPECT().ProcessingUpdate(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, update work.ProcessingUpdate) (*work.Work, error) {
				updated := *wrk
				updated.Metadata = update.Metadata
				updated.Revision = wrk.Revision + 1
				return &updated, nil
			})
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
		summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(dataWorkPostprocess.SummariesUpdate{}, nil)

		Expect(process().Result).To(Equal(work.ResultDelete))
	})

	DescribeTable("requests a synchronization only for the reasons that report a change is complete",
		func(reasons []string, expectSync bool) {
			wrk = newWork(work.StateProcessing, reasons, time.Now().Add(-time.Minute))
			expectListNone()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(dataWorkPostprocess.SummariesUpdate{}, nil)
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
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(dataWorkPostprocess.SummariesUpdate{}, nil),
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

	It("fails without retrying when the work has no metadata", func() {
		wrk.Metadata = nil

		result := process()
		Expect(result.Result).To(Equal(work.ResultFailed))
		Expect(result.FailedUpdate.FailedError.Error).To(MatchError(ContainSubstring("user id is missing")))
	})

	// Work created outside Enqueue could otherwise mutate the user outside the serialization of
	// the user
	DescribeTable("fails without retrying when the work is not scoped to the user its metadata reports",
		func(mutate func(wrk *work.Work)) {
			mutate(wrk)

			result := process()
			Expect(result.Result).To(Equal(work.ResultFailed))
			Expect(result.FailedUpdate.FailedError.Error).To(MatchError(ContainSubstring("does not match metadata user id")))
		},
		Entry("the group id names another user", func(wrk *work.Work) {
			wrk.GroupID = pointer.FromString(dataWorkPostprocess.IDFromUserID(userTest.RandomUserID()))
		}),
		Entry("the group id is missing", func(wrk *work.Work) {
			wrk.GroupID = nil
		}),
		Entry("the serial id names another user", func(wrk *work.Work) {
			wrk.SerialID = pointer.FromString(dataWorkPostprocess.IDFromUserID(userTest.RandomUserID()))
		}),
		Entry("the serial id is missing", func(wrk *work.Work) {
			wrk.SerialID = nil
		}),
	)

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
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(dataWorkPostprocess.SummariesUpdate{}, errorsTest.RandomError())
		}),
		Entry("requesting a synchronization", func() {
			expectListNone()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(dataWorkPostprocess.SummariesUpdate{}, nil)
			clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(errorsTest.RandomError())
		}),
		Entry("recording the changed summaries", func() {
			expectListNone()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(newSummariesUpdate(summaryTypes.SummaryTypeCGM), nil)
			processingUpdater.EXPECT().ProcessingUpdate(gomock.Any(), gomock.Any()).Return(nil, errorsTest.RandomError())
		}),
		Entry("reporting a summary update to the clinic service", func() {
			expectListNone()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(newSummariesUpdate(summaryTypes.SummaryTypeCGM), nil)
			expectProcessingUpdate()
			clinicsClient.EXPECT().UpdatePatientSummary(gomock.Any(), userID, gomock.Any()).Return(errorsTest.RandomError())
		}),
		Entry("reporting a summary deletion to the clinic service", func() {
			expectListNone()
			update := dataWorkPostprocess.SummariesUpdate{Deleted: map[string]string{summaryTypes.SummaryTypeCGM: primitive.NewObjectID().Hex()}}
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(update, nil)
			expectProcessingUpdate()
			clinicsClient.EXPECT().DeletePatientSummary(gomock.Any(), gomock.Any()).Return(errorsTest.RandomError())
		}),
	)

	// The clinic service stores a copy of the summaries of its patients, and the synchronization
	// reports from that copy, so a change must reach it first, and a summary that did not change must
	// not be reported at all
	Context("when the calculation changes the summaries", func() {
		var summariesUpdate dataWorkPostprocess.SummariesUpdate

		BeforeEach(func() {
			summariesUpdate = newSummariesUpdate(summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeBGM)
			wrk = newWork(work.StateProcessing, []string{dataWorkPostprocess.ReasonUploadCompleted}, time.Now().Add(-time.Minute))
		})

		// The changes are recorded in the metadata before they are reported, so that a failure
		// between the two reports them again rather than not at all
		It("records them, reports them to the clinic service, and requests a synchronization, in that order", func() {
			expectListNone()
			gomock.InOrder(
				summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(summariesUpdate, nil),
				processingUpdater.EXPECT().ProcessingUpdate(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, update work.ProcessingUpdate) (*work.Work, error) {
						Expect(update.Metadata).To(HaveKeyWithValue("pendingSummaryUpdates",
							ConsistOf(summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeBGM)))
						updated := *wrk
						updated.Metadata = update.Metadata
						updated.Revision = wrk.Revision + 1
						return &updated, nil
					}),
				clinicsClient.EXPECT().UpdatePatientSummary(gomock.Any(), userID, gomock.Any()).DoAndReturn(
					func(_ context.Context, _ string, patientSummary *clinic.PatientSummaryV1) error {
						Expect(patientSummary.CgmStats).ToNot(BeNil())
						Expect(patientSummary.CgmStats.Id).To(PointTo(Equal(summariesUpdate.CGM.ID.Hex())))
						Expect(patientSummary.BgmStats).ToNot(BeNil())
						Expect(patientSummary.BgmStats.Id).To(PointTo(Equal(summariesUpdate.BGM.ID.Hex())))
						return nil
					}),
				clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil),
			)

			Expect(process().Result).To(Equal(work.ResultDelete))
		})

		It("reports only the summaries that changed", func() {
			summariesUpdate = newSummariesUpdate(summaryTypes.SummaryTypeCGM)
			expectListNone()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(summariesUpdate, nil)
			expectProcessingUpdate()
			clinicsClient.EXPECT().UpdatePatientSummary(gomock.Any(), userID, gomock.Any()).DoAndReturn(
				func(_ context.Context, _ string, patientSummary *clinic.PatientSummaryV1) error {
					Expect(patientSummary.CgmStats).ToNot(BeNil())
					Expect(patientSummary.BgmStats).To(BeNil())
					return nil
				})
			clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil)

			Expect(process().Result).To(Equal(work.ResultDelete))
		})

		// The summaries returned without a change recorded were recalculated without change, or not
		// at all, and reporting them would store them again for every no-op work of the user
		It("reports nothing when the calculation reports the summaries did not change", func() {
			summariesUpdate = newSummariesUpdate()
			expectListNone()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(summariesUpdate, nil)
			clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil)

			Expect(process().Result).To(Equal(work.ResultDelete))
		})

		It("reports a deletion, before any update", func() {
			deletedSummaryID := primitive.NewObjectID().Hex()
			summariesUpdate = newSummariesUpdate(summaryTypes.SummaryTypeBGM)
			summariesUpdate.CGM = nil
			summariesUpdate.Deleted = map[string]string{summaryTypes.SummaryTypeCGM: deletedSummaryID}
			expectListNone()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(summariesUpdate, nil)
			processingUpdater.EXPECT().ProcessingUpdate(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, update work.ProcessingUpdate) (*work.Work, error) {
					Expect(update.Metadata).To(HaveKeyWithValue("pendingSummaryDeletes", ConsistOf(deletedSummaryID)))
					updated := *wrk
					updated.Metadata = update.Metadata
					updated.Revision = wrk.Revision + 1
					return &updated, nil
				})
			gomock.InOrder(
				clinicsClient.EXPECT().DeletePatientSummary(gomock.Any(), deletedSummaryID).Return(nil),
				clinicsClient.EXPECT().UpdatePatientSummary(gomock.Any(), userID, gomock.Any()).DoAndReturn(
					func(_ context.Context, _ string, patientSummary *clinic.PatientSummaryV1) error {
						Expect(patientSummary.CgmStats).To(BeNil())
						Expect(patientSummary.BgmStats).ToNot(BeNil())
						return nil
					}),
			)
			clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil)

			Expect(process().Result).To(Equal(work.ResultDelete))
		})

		// A change recorded by an earlier attempt was calculated but not yet reported; the retried
		// calculation reports no further change, so the record alone drives the report
		It("reports the changes recorded by an earlier attempt even though this calculation reports none", func() {
			deletedSummaryID := primitive.NewObjectID().Hex()
			encoded, err := metadata.Encode(&dataWorkPostprocess.Metadata{
				Metadata:              userWork.Metadata{UserID: pointer.FromString(userID)},
				Reasons:               []string{dataWorkPostprocess.ReasonUploadCompleted},
				PendingSummaryUpdates: []string{summaryTypes.SummaryTypeCGM},
				PendingSummaryDeletes: []string{deletedSummaryID},
			})
			Expect(err).ToNot(HaveOccurred())
			wrk.Metadata = encoded

			summariesUpdate = newSummariesUpdate()
			expectListNone()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(summariesUpdate, nil)
			clinicsClient.EXPECT().DeletePatientSummary(gomock.Any(), deletedSummaryID).Return(nil)
			clinicsClient.EXPECT().UpdatePatientSummary(gomock.Any(), userID, gomock.Any()).DoAndReturn(
				func(_ context.Context, _ string, patientSummary *clinic.PatientSummaryV1) error {
					Expect(patientSummary.CgmStats).ToNot(BeNil())
					Expect(patientSummary.CgmStats.Id).To(PointTo(Equal(summariesUpdate.CGM.ID.Hex())))
					Expect(patientSummary.BgmStats).To(BeNil())
					return nil
				})
			clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil)

			Expect(process().Result).To(Equal(work.ResultDelete))
		})

		// The changes ride along on the failing update, so the retry still knows a report is owed
		It("keeps them recorded when reporting them fails", func() {
			expectListNone()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(summariesUpdate, nil)
			expectProcessingUpdate()
			clinicsClient.EXPECT().UpdatePatientSummary(gomock.Any(), userID, gomock.Any()).Return(errorsTest.RandomError())

			result := process()
			Expect(result.Result).To(Equal(work.ResultFailing))
			Expect(result.FailingUpdate.Metadata).To(HaveKeyWithValue("pendingSummaryUpdates",
				ConsistOf(summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeBGM)))
		})

		// A change calculated before the failure is recorded on the failing update, so the retry
		// reports it even though its own calculation reports no further change
		It("records a change calculated before the calculation of another summary fails", func() {
			summariesUpdate = newSummariesUpdate(summaryTypes.SummaryTypeCGM)
			expectListNone()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(summariesUpdate, errorsTest.RandomError())

			result := process()
			Expect(result.Result).To(Equal(work.ResultFailing))
			Expect(result.FailingUpdate.Metadata).To(HaveKeyWithValue("pendingSummaryUpdates",
				ConsistOf(summaryTypes.SummaryTypeCGM)))
		})
	})

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
				workClient.EXPECT().Delete(gomock.Any(), sibling.ID, &request.Condition{Revision: pointer.FromInt(sibling.Revision)}).Return(sibling, nil),
			)
		}

		BeforeEach(func() {
			sibling = newWork(work.StatePending, []string{dataWorkPostprocess.ReasonUploadCompleted}, time.Now().Add(-time.Second))
		})

		It("reports its reasons, deletes it, and processes once", func() {
			expectListWithSibling()
			expectProcessingUpdateThenDelete()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(dataWorkPostprocess.SummariesUpdate{}, nil)
			clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil)

			Expect(process().Result).To(Equal(work.ResultDelete))
		})

		// A sibling left pending fails on its own pickup; failing this valid work with it would
		// lose the reasons this work reports
		DescribeTable("skips it and processes without failing when",
			func(mutate func()) {
				mutate()
				expectListWithSibling()
				summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(dataWorkPostprocess.SummariesUpdate{}, nil)

				Expect(process().Result).To(Equal(work.ResultDelete))
			},
			Entry("its metadata is invalid", func() {
				sibling.Metadata["reasons"] = []any{"INVALID"}
			}),
			Entry("its metadata is missing", func() {
				sibling.Metadata = nil
			}),
			// A mislabeled row listed by group belongs to the user its own metadata names; absorbing
			// it would merge and destroy another user's change
			Entry("it is not scoped to the user its metadata reports", func() {
				sibling.Metadata["userId"] = userTest.RandomUserID()
			}),
		)

		It("absorbs the others when one of them is invalid", func() {
			invalid := newWork(work.StatePending, []string{dataWorkPostprocess.ReasonUploadCompleted}, time.Now().Add(-time.Second))
			invalid.Metadata = nil
			workClient.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*work.Work{invalid, sibling}, nil)
			expectProcessingUpdateThenDelete()
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(dataWorkPostprocess.SummariesUpdate{}, nil)
			clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil)

			Expect(process().Result).To(Equal(work.ResultDelete))
		})

		// An upload reporting only full batches is still in progress, so the summaries are not
		// calculated for every batch so far. Jellyfish defers the same way.
		Context("that is deferred and reports only a full batch", func() {
			var deferredUntil time.Time

			BeforeEach(func() {
				deferredUntil = time.Now().Add(90 * time.Second)
				wrk = newWork(work.StateProcessing, []string{dataWorkPostprocess.ReasonLegacyDataAdded}, time.Now().Add(-time.Minute))
				sibling = newWork(work.StatePending, []string{dataWorkPostprocess.ReasonLegacyDataAdded}, deferredUntil)
			})

			It("defers rather than calculating the summaries", func() {
				expectListWithSibling()
				processingUpdater.EXPECT().ProcessingUpdate(gomock.Any(), gomock.Any()).Return(wrk, nil)
				workClient.EXPECT().Delete(gomock.Any(), sibling.ID, &request.Condition{Revision: pointer.FromInt(sibling.Revision)}).Return(sibling, nil)

				result := process()
				Expect(result.Result).To(Equal(work.ResultPending))
				Expect(result.PendingUpdate.ProcessingAvailableTime).To(BeTemporally("==", deferredUntil))
			})

			// Deferral requires every reason to defer, which is not the complement of requesting a
			// synchronization: data added requests none, yet still cancels the deferral
			It("calculates the summaries when any reason does not defer, even one reporting no completion", func() {
				sibling = newWork(work.StatePending, []string{dataWorkPostprocess.ReasonDataAdded}, deferredUntil)
				expectListWithSibling()
				processingUpdater.EXPECT().ProcessingUpdate(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, update work.ProcessingUpdate) (*work.Work, error) {
						updated := *wrk
						updated.Metadata = update.Metadata
						return &updated, nil
					})
				workClient.EXPECT().Delete(gomock.Any(), sibling.ID, &request.Condition{Revision: pointer.FromInt(sibling.Revision)}).Return(sibling, nil)
				summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(dataWorkPostprocess.SummariesUpdate{}, nil)
				clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil)

				Expect(process().Result).To(Equal(work.ResultDelete))
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
				workClient.EXPECT().Delete(gomock.Any(), sibling.ID, &request.Condition{Revision: pointer.FromInt(sibling.Revision)}).Return(sibling, nil)
				summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(dataWorkPostprocess.SummariesUpdate{}, nil)
				clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil)

				Expect(process().Result).To(Equal(work.ResultDelete))
			})
		})

		It("retries when it cannot be deleted", func() {
			expectListWithSibling()
			processingUpdater.EXPECT().ProcessingUpdate(gomock.Any(), gomock.Any()).Return(wrk, nil)
			workClient.EXPECT().Delete(gomock.Any(), sibling.ID, &request.Condition{Revision: pointer.FromInt(sibling.Revision)}).Return(nil, errorsTest.RandomError())

			Expect(process().Result).To(Equal(work.ResultFailing))
		})

		// The sibling changed after it was listed — a producer may have merged another change into
		// it. Deleting it would destroy that change; leaving it reprocesses the reasons already
		// absorbed redundantly, which is the at-least-once contract.
		It("continues without deleting it when it changed since it was listed", func() {
			expectListWithSibling()
			processingUpdater.EXPECT().ProcessingUpdate(gomock.Any(), gomock.Any()).Return(wrk, nil)
			workClient.EXPECT().Delete(gomock.Any(), sibling.ID, &request.Condition{Revision: pointer.FromInt(sibling.Revision)}).Return(nil, nil)
			summarizers.EXPECT().UpdateSummaries(gomock.Any(), userID).Return(dataWorkPostprocess.SummariesUpdate{}, nil)
			clinicsClient.EXPECT().SyncEHRDataForPatient(gomock.Any(), userID).Return(nil)

			Expect(process().Result).To(Equal(work.ResultDelete))
		})
	})
})
