package summary_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clinic "github.com/tidepool-org/clinic/client"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/mock/gomock"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	dataWorkSummary "github.com/tidepool-org/platform/data/work/summary"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	summaryTest "github.com/tidepool-org/platform/summary/test"
	summaryTypes "github.com/tidepool-org/platform/summary/types"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("summary", func() {
	var controller *gomock.Controller
	var ctx context.Context
	var userID string

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		userID = userTest.RandomUserID()
	})

	Context("CalculateSummary", func() {
		type cgmSummarizer = summaryTest.MockSummarizer[
			*summaryTypes.CGMPeriods, *summaryTypes.GlucoseBucket,
			summaryTypes.CGMPeriods, summaryTypes.GlucoseBucket]

		var summarizer *cgmSummarizer
		var update dataWorkSummary.SummariesUpdate
		var before *summaryTypes.CGMSummary
		var after *summaryTypes.CGMSummary
		var calculateErr error

		calculate := func() (*summaryTypes.CGMSummary, error) {
			return dataWorkSummary.CalculateSummary(ctx, summarizer,
				func(_ context.Context, calculateUserID string) (
					*summaryTypes.CGMSummary, error) {

					Expect(calculateUserID).To(Equal(userID))
					return after, calculateErr
				}, userID, &update)
		}

		BeforeEach(func() {
			summarizer = summaryTest.NewMockSummarizer[*summaryTypes.CGMPeriods,
				*summaryTypes.GlucoseBucket](controller)
			update = dataWorkSummary.SummariesUpdate{}
			before = summaryTest.RandomCGMSummary(userID)
			before.ID = primitive.NewObjectID()
			after = summaryTest.RandomCGMSummary(userID)
			after.ID = before.ID
			after.Dates.LastUpdatedDate = before.Dates.LastUpdatedDate.Add(time.Minute)
			calculateErr = nil
		})

		It("returns an error if the summary can't be read", func() {
			summarizer.EXPECT().GetSummary(gomock.Any(), userID).
				Return(nil, errorsTest.RandomError())

			result, err := calculate()
			Expect(err).To(MatchError(ContainSubstring("unable to get cgm summary")))
			Expect(result).To(BeNil())
			Expect(update).To(Equal(dataWorkSummary.SummariesUpdate{}))
		})

		It("returns an error if the summary can't be calculated", func() {
			summarizer.EXPECT().GetSummary(gomock.Any(), userID).Return(before, nil)
			calculateErr = errorsTest.RandomError()

			result, err := calculate()
			Expect(err).To(MatchError(ContainSubstring("unable to update cgm summary")))
			Expect(result).To(BeNil())
			Expect(update).To(Equal(dataWorkSummary.SummariesUpdate{}))
		})

		It("records a summary that was created as updated", func() {
			summarizer.EXPECT().GetSummary(gomock.Any(), userID).Return(nil, nil)

			Expect(calculate()).To(Equal(after))
			Expect(update.UpdatedTypes).To(ConsistOf(summaryTypes.SummaryTypeCGM))
			Expect(update.Deleted).To(BeEmpty())
		})

		It("records a summary that was calculated again as updated", func() {
			summarizer.EXPECT().GetSummary(gomock.Any(), userID).Return(before, nil)

			Expect(calculate()).To(Equal(after))
			Expect(update.UpdatedTypes).To(ConsistOf(summaryTypes.SummaryTypeCGM))
			Expect(update.Deleted).To(BeEmpty())
		})

		It("records nothing for a summary that was left as it was", func() {
			summarizer.EXPECT().GetSummary(gomock.Any(), userID).Return(before, nil)
			after.Dates.LastUpdatedDate = before.Dates.LastUpdatedDate

			Expect(calculate()).To(Equal(after))
			Expect(update).To(Equal(dataWorkSummary.SummariesUpdate{}))
		})

		It("records a summary that was deleted by its id", func() {
			summarizer.EXPECT().GetSummary(gomock.Any(), userID).Return(before, nil)
			after = nil

			Expect(calculate()).To(BeNil())
			Expect(update.UpdatedTypes).To(BeEmpty())
			Expect(update.Deleted).To(Equal(map[string]string{
				summaryTypes.SummaryTypeCGM: before.ID.Hex(),
			}))
		})

		It("records nothing for a user who never had a summary", func() {
			summarizer.EXPECT().GetSummary(gomock.Any(), userID).Return(nil, nil)
			after = nil

			Expect(calculate()).To(BeNil())
			Expect(update).To(Equal(dataWorkSummary.SummariesUpdate{}))
		})
	})

	Context("RecordSummariesUpdate", func() {
		var pendingUpdates []string
		var pendingDeletes []string

		BeforeEach(func() {
			pendingUpdates = nil
			pendingDeletes = nil
		})

		It("reports no change for an update that changed nothing", func() {
			Expect(dataWorkSummary.RecordSummariesUpdate(&pendingUpdates, &pendingDeletes,
				dataWorkSummary.SummariesUpdate{})).To(BeFalse())
			Expect(pendingUpdates).To(BeEmpty())
			Expect(pendingDeletes).To(BeEmpty())
		})

		It("adds the updated types and the deleted ids, in order", func() {
			pendingDeletes = []string{"b"}
			update := dataWorkSummary.SummariesUpdate{
				UpdatedTypes: []string{
					summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeBGM,
				},
				Deleted: map[string]string{summaryTypes.SummaryTypeCGM: "a"},
			}

			Expect(dataWorkSummary.RecordSummariesUpdate(&pendingUpdates, &pendingDeletes,
				update)).To(BeTrue())
			Expect(pendingUpdates).To(Equal(
				[]string{summaryTypes.SummaryTypeBGM, summaryTypes.SummaryTypeCGM}))
			Expect(pendingDeletes).To(Equal([]string{"a", "b"}))
		})

		It("reports no change for what is already pending", func() {
			pendingUpdates = []string{summaryTypes.SummaryTypeCGM}
			pendingDeletes = []string{"a"}
			update := dataWorkSummary.SummariesUpdate{
				UpdatedTypes: []string{summaryTypes.SummaryTypeCGM},
				Deleted:      map[string]string{summaryTypes.SummaryTypeBGM: "a"},
			}

			Expect(dataWorkSummary.RecordSummariesUpdate(&pendingUpdates, &pendingDeletes,
				update)).To(BeFalse())
			Expect(pendingUpdates).To(Equal([]string{summaryTypes.SummaryTypeCGM}))
			Expect(pendingDeletes).To(Equal([]string{"a"}))
		})
	})

	Context("SyncClinicSummaries", func() {
		var clinicsClient *clinicsTest.MockClient
		var update dataWorkSummary.SummariesUpdate

		BeforeEach(func() {
			clinicsClient = clinicsTest.NewMockClient(controller)
			update = dataWorkSummary.SummariesUpdate{
				CGM: summaryTest.RandomCGMSummary(userID),
				BGM: summaryTest.RandomBGMSummary(userID),
			}
		})

		It("syncs nothing when nothing is pending", func() {
			Expect(dataWorkSummary.SyncClinicSummaries(ctx, clinicsClient, userID, nil, nil,
				update)).To(BeFalse())
		})

		It("deletes the summaries pending deletion, then updates only the types pending",
			func() {
				gomock.InOrder(
					clinicsClient.EXPECT().DeletePatientSummary(gomock.Any(), "a").
						Return(nil),
					clinicsClient.EXPECT().
						UpdatePatientSummary(gomock.Any(), userID, gomock.Any()).
						DoAndReturn(func(_ context.Context, _ string,
							patientSummary *clinic.PatientSummaryV1) error {

							Expect(patientSummary.CgmStats).To(BeNil())
							Expect(patientSummary.BgmStats).ToNot(BeNil())
							return nil
						}),
				)

				Expect(dataWorkSummary.SyncClinicSummaries(ctx, clinicsClient, userID,
					[]string{summaryTypes.SummaryTypeBGM}, []string{"a"}, update)).
					To(BeTrue())
			})

		It("doesn't update the patient with a summary that wasn't calculated", func() {
			update.CGM = nil

			Expect(dataWorkSummary.SyncClinicSummaries(ctx, clinicsClient, userID,
				[]string{summaryTypes.SummaryTypeCGM}, nil, update)).To(BeTrue())
		})

		It("returns an error if a summary can't be deleted", func() {
			clinicsClient.EXPECT().DeletePatientSummary(gomock.Any(), "a").
				Return(errorsTest.RandomError())

			synced, err := dataWorkSummary.SyncClinicSummaries(ctx, clinicsClient, userID,
				[]string{summaryTypes.SummaryTypeCGM}, []string{"a"}, update)
			Expect(err).To(MatchError(ContainSubstring("unable to delete patient summary")))
			Expect(synced).To(BeFalse())
		})

		It("returns an error if the patient can't be updated", func() {
			clinicsClient.EXPECT().UpdatePatientSummary(gomock.Any(), userID, gomock.Any()).
				Return(errorsTest.RandomError())

			synced, err := dataWorkSummary.SyncClinicSummaries(ctx, clinicsClient, userID,
				[]string{summaryTypes.SummaryTypeCGM}, nil, update)
			Expect(err).To(MatchError(ContainSubstring("unable to update patient summary")))
			Expect(synced).To(BeFalse())
		})
	})
})
