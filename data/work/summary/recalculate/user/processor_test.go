package user_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clinic "github.com/tidepool-org/clinic/client"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/mock/gomock"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	dataWorkSummary "github.com/tidepool-org/platform/data/work/summary"
	dataWorkSummaryRecalculateUser "github.com/tidepool-org/platform/data/work/summary/recalculate/user"
	dataWorkSummaryRecalculateUserTest "github.com/tidepool-org/platform/data/work/summary/recalculate/user/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/pointer"
	summaryTest "github.com/tidepool-org/platform/summary/test"
	summaryTypes "github.com/tidepool-org/platform/summary/types"
	userTest "github.com/tidepool-org/platform/user/test"
	userWork "github.com/tidepool-org/platform/user/work"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("Processor", func() {
	var controller *gomock.Controller
	var workClient *workTest.MockClient
	var summarizers *dataWorkSummaryRecalculateUserTest.MockSummarizers
	var clinicsClient *clinicsTest.MockClient
	var processingUpdater *workTest.MockProcessingUpdater
	var processor *dataWorkSummaryRecalculateUser.Processor
	var ctx context.Context
	var userID string
	var wrk *work.Work

	newWork := func(workMetadata *dataWorkSummaryRecalculateUser.Metadata) *work.Work {
		encoded, err := metadata.Encode(workMetadata)
		Expect(err).ToNot(HaveOccurred())
		return &work.Work{
			ID:                      workTest.RandomID(),
			Type:                    dataWorkSummaryRecalculateUser.Type,
			SerialID:                pointer.From(dataWorkPostprocess.IDFromUserID(userID)),
			ProcessingAvailableTime: time.Now().Add(-time.Minute),
			ProcessingTimeout:       300,
			Metadata:                encoded,
			State:                   work.StateProcessing,
			Revision:                2,
		}
	}

	newSummariesUpdate := func(updatedTypes ...string) dataWorkSummary.SummariesUpdate {
		update := dataWorkSummary.SummariesUpdate{
			CGM:          summaryTest.RandomCGMSummary(userID),
			BGM:          summaryTest.RandomBGMSummary(userID),
			UpdatedTypes: updatedTypes,
		}
		update.CGM.ID = primitive.NewObjectID()
		update.BGM.ID = primitive.NewObjectID()
		return update
	}

	// The processing update persists the metadata of the request, so it is echoed back the
	// way the store does
	expectProcessingUpdate := func(
		expectations ...func(update work.ProcessingUpdate)) any {

		return processingUpdater.EXPECT().ProcessingUpdate(gomock.Any(), gomock.Any()).
			DoAndReturn(
				func(_ context.Context, update work.ProcessingUpdate) (*work.Work, error) {
					for _, expectation := range expectations {
						expectation(update)
					}
					updated := *wrk
					updated.Metadata = update.Metadata
					updated.Revision = wrk.Revision + 1
					return &updated, nil
				})
	}

	process := func() *work.ProcessResult {
		return processor.Process(ctx, wrk, processingUpdater)
	}

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		workClient = workTest.NewMockClient(controller)
		summarizers = dataWorkSummaryRecalculateUserTest.NewMockSummarizers(controller)
		clinicsClient = clinicsTest.NewMockClient(controller)
		processingUpdater = workTest.NewMockProcessingUpdater(controller)
		ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		userID = userTest.RandomUserID()
		wrk = newWork(&dataWorkSummaryRecalculateUser.Metadata{
			Metadata: userWork.Metadata{UserID: pointer.From(userID)},
		})

		var err error
		processor, err = dataWorkSummaryRecalculateUser.NewProcessor(
			dataWorkSummaryRecalculateUser.Dependencies{
				Dependencies:  workBase.Dependencies{WorkClient: workClient},
				Summarizers:   summarizers,
				ClinicsClient: clinicsClient,
			})
		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error if dependencies is invalid", func() {
		processor, err := dataWorkSummaryRecalculateUser.NewProcessor(
			dataWorkSummaryRecalculateUser.Dependencies{
				Dependencies:  workBase.Dependencies{WorkClient: workClient},
				ClinicsClient: clinicsClient,
			})
		Expect(err).To(MatchError("dependencies is invalid; summarizers is missing"))
		Expect(processor).To(BeNil())
	})

	// The clinics client mock rejects any call that isn't expected, so each of these also
	// shows that an EHR sync is never triggered
	It("records the recalculated summaries, syncs them to the clinic service, and deletes "+
		"the work, in that order", func() {

		summariesUpdate := newSummariesUpdate(summaryTypes.SummaryTypeCGM,
			summaryTypes.SummaryTypeBGM)
		gomock.InOrder(
			summarizers.EXPECT().RecalculateSummaries(gomock.Any(), userID).
				Return(summariesUpdate, nil),
			expectProcessingUpdate(func(update work.ProcessingUpdate) {
				Expect(update.Metadata).To(HaveKeyWithValue("pendingSummaryUpdates",
					ConsistOf(summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeBGM)))
			}),
			clinicsClient.EXPECT().UpdatePatientSummary(gomock.Any(), userID, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string,
					patientSummary *clinic.PatientSummaryV1) error {

					Expect(patientSummary.CgmStats).ToNot(BeNil())
					Expect(patientSummary.CgmStats.Id).
						To(HaveValue(Equal(summariesUpdate.CGM.ID.Hex())))
					Expect(patientSummary.BgmStats).ToNot(BeNil())
					Expect(patientSummary.BgmStats.Id).
						To(HaveValue(Equal(summariesUpdate.BGM.ID.Hex())))
					return nil
				}),
		)

		Expect(process().Result).To(Equal(work.ResultDelete))
	})

	It("syncs nothing for a user without summaries", func() {
		summarizers.EXPECT().RecalculateSummaries(gomock.Any(), userID).
			Return(dataWorkSummary.SummariesUpdate{}, nil)

		Expect(process().Result).To(Equal(work.ResultDelete))
	})

	It("syncs a deletion, before any update", func() {
		deletedSummaryID := primitive.NewObjectID().Hex()
		summariesUpdate := newSummariesUpdate(summaryTypes.SummaryTypeBGM)
		summariesUpdate.CGM = nil
		summariesUpdate.Deleted = map[string]string{
			summaryTypes.SummaryTypeCGM: deletedSummaryID,
		}
		summarizers.EXPECT().RecalculateSummaries(gomock.Any(), userID).
			Return(summariesUpdate, nil)
		expectProcessingUpdate(func(update work.ProcessingUpdate) {
			Expect(update.Metadata).To(HaveKeyWithValue("pendingSummaryDeletes",
				ConsistOf(deletedSummaryID)))
		})
		gomock.InOrder(
			clinicsClient.EXPECT().DeletePatientSummary(gomock.Any(), deletedSummaryID).
				Return(nil),
			clinicsClient.EXPECT().UpdatePatientSummary(gomock.Any(), userID, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string,
					patientSummary *clinic.PatientSummaryV1) error {

					Expect(patientSummary.CgmStats).To(BeNil())
					Expect(patientSummary.BgmStats).ToNot(BeNil())
					return nil
				}),
		)

		Expect(process().Result).To(Equal(work.ResultDelete))
	})

	// A deletion recorded by an earlier attempt was made but not yet synced. The retried
	// recalculation finds no summary to delete, so the record alone drives the sync.
	It("syncs a deletion recorded by an earlier attempt", func() {
		deletedSummaryID := primitive.NewObjectID().Hex()
		wrk = newWork(&dataWorkSummaryRecalculateUser.Metadata{
			Metadata:              userWork.Metadata{UserID: pointer.From(userID)},
			PendingSummaryDeletes: []string{deletedSummaryID},
		})
		summarizers.EXPECT().RecalculateSummaries(gomock.Any(), userID).
			Return(dataWorkSummary.SummariesUpdate{}, nil)
		clinicsClient.EXPECT().DeletePatientSummary(gomock.Any(), deletedSummaryID).
			Return(nil)

		Expect(process().Result).To(Equal(work.ResultDelete))
	})

	// A partial update is returned alongside the error, and is recorded so that the retry
	// still syncs it
	It("is failing, recording what was recalculated, if the recalculation fails partway",
		func() {
			summariesUpdate := newSummariesUpdate(summaryTypes.SummaryTypeCGM)
			summariesUpdate.BGM = nil
			summarizers.EXPECT().RecalculateSummaries(gomock.Any(), userID).
				Return(summariesUpdate, errorsTest.RandomError())

			result := process()
			Expect(result.Result).To(Equal(work.ResultFailing))
			Expect(result.FailingUpdate).ToNot(BeNil())
			Expect(result.FailingUpdate.Metadata).To(HaveKeyWithValue(
				"pendingSummaryUpdates", ConsistOf(summaryTypes.SummaryTypeCGM)))
		})

	It("is failing if the summaries can't be synced to the clinic service", func() {
		summarizers.EXPECT().RecalculateSummaries(gomock.Any(), userID).
			Return(newSummariesUpdate(summaryTypes.SummaryTypeCGM), nil)
		expectProcessingUpdate()
		clinicsClient.EXPECT().UpdatePatientSummary(gomock.Any(), userID, gomock.Any()).
			Return(errorsTest.RandomError())

		Expect(process().Result).To(Equal(work.ResultFailing))
	})

	It("fails if the user id is missing", func() {
		wrk = newWork(&dataWorkSummaryRecalculateUser.Metadata{})

		Expect(process().Result).To(Equal(work.ResultFailed))
	})

	It("fails if the work isn't serialized with the postprocess work of the user", func() {
		otherUserID := userTest.RandomUserID()
		wrk.SerialID = pointer.From(dataWorkPostprocess.IDFromUserID(otherUserID))

		Expect(process().Result).To(Equal(work.ResultFailed))
	})
})
