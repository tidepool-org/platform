package store_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"

	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	dataStoreSummary "github.com/tidepool-org/platform/summary/store"
	"github.com/tidepool-org/platform/summary/test"
	"github.com/tidepool-org/platform/summary/types"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("CGM", Label("mongodb", "slow", "integration"), func() {
	var logger *logTest.Logger
	var err error
	var ctx context.Context
	var store *dataStoreMongo.Store
	var summaryRepository *storeStructuredMongo.Repository

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
	})

	Context("Create repo and store", func() {
		var createStore *dataStoreMongo.Store

		It("Repo", func() {
			createStore = GetSuiteStore()

			summaryRepository = createStore.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			cgmStore := dataStoreSummary.NewSummaries[*types.CGMPeriods, *types.GlucoseBucket](summaryRepository)
			Expect(cgmStore).ToNot(BeNil())
		})
	})

	Context("Store", func() {
		var userId string
		var userIdOther string
		var userCGMSummary *types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]
		var cgmStore *dataStoreSummary.Summaries[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]

		BeforeEach(func() {
			store = GetSuiteStore()
			summaryRepository = store.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			cgmStore = dataStoreSummary.NewSummaries[*types.CGMPeriods, *types.GlucoseBucket](summaryRepository)

			userId = userTest.RandomUserID()
			userIdOther = userTest.RandomUserID()
		})

		AfterEach(func() {
			if summaryRepository != nil {
				_, err = summaryRepository.DeleteMany(ctx, bson.D{})
				Expect(err).To(Succeed())
			}
		})

		Context("ReplaceSummary", func() {
			It("Insert Summary with missing context", func() {
				userCGMSummary = test.RandomCGMSummary(userId)
				err = cgmStore.ReplaceSummary(nil, userCGMSummary)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
			})

			It("Insert Summary with missing Summary", func() {
				err = cgmStore.ReplaceSummary(ctx, nil)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("summary object is missing"))
			})

			It("Insert Summary with missing UserId", func() {
				userCGMSummary = test.RandomCGMSummary(userId)
				Expect(userCGMSummary.Type).To(Equal("cgm"))

				userCGMSummary.UserID = ""

				err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("summary is missing UserID"))
			})

			It("Insert Summary with missing Type", func() {
				userCGMSummary = test.RandomCGMSummary(userId)
				userCGMSummary.Type = ""

				err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type '', expected 'cgm'"))
			})

			It("Insert Summary with invalid Type", func() {
				userCGMSummary = test.RandomCGMSummary(userId)
				userCGMSummary.Type = "bgm"

				err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type 'bgm', expected 'cgm'"))
			})

			It("Insert Summary", func() {
				userCGMSummary = test.RandomCGMSummary(userId)
				Expect(userCGMSummary.Type).To(Equal("cgm"))

				err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
				Expect(err).ToNot(HaveOccurred())

				userCGMSummaryWritten, err := cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// copy id, as that was mongo generated
				userCGMSummary.ID = userCGMSummaryWritten.ID
				Expect(userCGMSummaryWritten).To(Equal(userCGMSummary))
			})

			It("Update Summary", func() {
				var userCGMSummaryTwo *types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]
				var userCGMSummaryWritten *types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]
				var userCGMSummaryWrittenTwo *types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]

				// generate and insert first summary
				userCGMSummary = test.RandomCGMSummary(userId)
				Expect(userCGMSummary.Type).To(Equal("cgm"))

				err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
				Expect(err).ToNot(HaveOccurred())

				// confirm first summary was written, get ID
				userCGMSummaryWritten, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// copy id, as that was mongo generated
				userCGMSummary.ID = userCGMSummaryWritten.ID
				Expect(userCGMSummaryWritten).To(Equal(userCGMSummary))

				// generate a new summary with same type and user, and upsert
				userCGMSummaryTwo = test.RandomCGMSummary(userId)
				err = cgmStore.ReplaceSummary(ctx, userCGMSummaryTwo)
				Expect(err).ToNot(HaveOccurred())

				userCGMSummaryWrittenTwo, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// confirm the ID was unchanged
				Expect(userCGMSummaryWrittenTwo.ID).To(Equal(userCGMSummaryWritten.ID))

				// confirm the written summary matches the new summary
				userCGMSummaryTwo.ID = userCGMSummaryWritten.ID
				Expect(userCGMSummaryWrittenTwo).To(Equal(userCGMSummaryTwo))
			})
		})

		Context("DeleteSummary", func() {
			It("Delete Summary with empty context", func() {
				err = cgmStore.DeleteSummary(nil, userId)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
			})

			It("Delete Summary with empty userId", func() {
				err = cgmStore.DeleteSummary(ctx, "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing"))
			})

			It("Delete Summary", func() {
				var userCGMSummaryWritten *types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]

				userCGMSummary = test.RandomCGMSummary(userId)
				Expect(userCGMSummary.Type).To(Equal("cgm"))

				err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
				Expect(err).ToNot(HaveOccurred())

				// confirm writes
				userCGMSummaryWritten, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummaryWritten).ToNot(BeNil())

				// delete
				err = cgmStore.DeleteSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// confirm delete
				userCGMSummaryWritten, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummaryWritten).To(BeNil())
			})
		})

		Context("CreateSummaries", func() {
			It("Create summaries with missing context", func() {
				var summaries = []*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
					test.RandomCGMSummary(userId),
					test.RandomCGMSummary(userIdOther),
				}

				_, err = cgmStore.CreateSummaries(nil, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
			})

			It("Create summaries with missing summaries", func() {
				_, err = cgmStore.CreateSummaries(ctx, nil)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("summaries for create missing"))
			})

			It("Create summaries with an invalid type", func() {
				var summaries = []*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
					test.RandomCGMSummary(userId),
					test.RandomCGMSummary(userIdOther),
				}

				summaries[0].Type = "bgm"

				_, err = cgmStore.CreateSummaries(ctx, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type 'bgm', expected 'cgm' at index 0"))
			})

			It("Create summaries with an empty userId", func() {
				var summaries = []*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
					test.RandomCGMSummary(userId),
					test.RandomCGMSummary(userIdOther),
				}

				summaries[0].UserID = ""

				_, err = cgmStore.CreateSummaries(ctx, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing at index 0"))
			})

			It("Create summaries", func() {
				var count int
				var summaries = []*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
					test.RandomCGMSummary(userId),
					test.RandomCGMSummary(userIdOther),
				}

				count, err = cgmStore.CreateSummaries(ctx, summaries)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(2))

				for i := 0; i < 2; i++ {
					userCGMSummary, err = cgmStore.GetSummary(ctx, summaries[0].UserID)
					Expect(err).ToNot(HaveOccurred())
					Expect(userCGMSummary).ToNot(BeNil())
					summaries[i].ID = userCGMSummary.ID
					Expect(userCGMSummary).To(Equal(summaries[0]))
				}
			})
		})

		Context("SetOutdated", func() {
			var outdatedSince *time.Time
			var userCGMSummaryWritten *types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]

			It("With missing context", func() {
				outdatedSince, err = cgmStore.SetOutdated(nil, userId, types.OutdatedReasonDataAdded)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
				Expect(outdatedSince).To(BeNil())
			})

			It("With missing userId", func() {
				outdatedSince, err = cgmStore.SetOutdated(ctx, "", types.OutdatedReasonDataAdded)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing"))
				Expect(outdatedSince).To(BeNil())
			})

			It("With multiple reasons", func() {
				outdatedSinceOriginal, err := cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSinceOriginal).ToNot(BeNil())

				userCGMSummary, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummary.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userCGMSummary.Dates.OutdatedSince).To(Equal(outdatedSinceOriginal))
				Expect(userCGMSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded}))

				outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonSchemaMigration)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userCGMSummary, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummary.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userCGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
				Expect(userCGMSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonSchemaMigration}))

				outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userCGMSummary, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummary.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userCGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
				Expect(userCGMSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonSchemaMigration}))
			})

			It("With no existing summary", func() {
				outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userCGMSummary, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummary.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userCGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
			})

			It("With an existing non-outdated summary", func() {
				userCGMSummary = test.RandomCGMSummary(userId)
				userCGMSummary.Dates.OutdatedSince = nil
				err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
				Expect(err).ToNot(HaveOccurred())

				outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userCGMSummaryWritten, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userCGMSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))

			})

			It("With an existing outdated summary", func() {
				var fiveMinutesAgo = time.Now().Add(time.Duration(-5) * time.Minute).UTC().Truncate(time.Millisecond)

				userCGMSummary = test.RandomCGMSummary(userId)
				userCGMSummary.Dates.OutdatedSince = &fiveMinutesAgo
				err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
				Expect(err).ToNot(HaveOccurred())

				outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userCGMSummaryWritten, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userCGMSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))
			})

			It("With an existing outdated summary beyond the outdatedSinceLimit", func() {
				now := time.Now().UTC().Truncate(time.Millisecond)

				userCGMSummary = test.RandomCGMSummary(userId)
				userCGMSummary.Dates.OutdatedSince = &now
				err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
				Expect(err).ToNot(HaveOccurred())

				outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userCGMSummaryWritten, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
			})

			It("With an existing outdated summary with schema migration reason", func() {
				now := time.Now().UTC().Truncate(time.Millisecond)
				fiveMinutesAgo := now.Add(time.Duration(-5) * time.Minute)

				userCGMSummary = test.RandomCGMSummary(userId)
				userCGMSummary.Dates.OutdatedSince = &fiveMinutesAgo
				userCGMSummary.Dates.OutdatedReason = []string{types.OutdatedReasonUploadCompleted}
				Expect(userCGMSummary.Periods.GlucosePeriods).ToNot(HaveLen(0))

				err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
				Expect(err).ToNot(HaveOccurred())

				outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonSchemaMigration)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userCGMSummaryWritten, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userCGMSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))
				Expect(userCGMSummaryWritten.Periods.GlucosePeriods).ToNot(HaveLen(0))
				Expect(userCGMSummaryWritten.Dates.LastData).To(Equal(userCGMSummary.Dates.LastData))
				Expect(userCGMSummaryWritten.Dates.FirstData).To(Equal(userCGMSummary.Dates.FirstData))
				Expect(userCGMSummaryWritten.Dates.LastUpdatedDate).To(Equal(userCGMSummary.Dates.LastUpdatedDate))
				Expect(userCGMSummaryWritten.Dates.LastUploadDate).To(Equal(userCGMSummary.Dates.LastUploadDate))
				Expect(userCGMSummaryWritten.Dates.OutdatedReason).To(ConsistOf(types.OutdatedReasonSchemaMigration, types.OutdatedReasonUploadCompleted))
			})
		})

		Context("GetSummary", func() {
			It("With missing context", func() {
				userCGMSummary, err = cgmStore.GetSummary(nil, userId)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
				Expect(userCGMSummary).To(BeNil())
			})

			It("With missing userId", func() {
				userCGMSummary, err = cgmStore.GetSummary(ctx, "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing"))
				Expect(userCGMSummary).To(BeNil())
			})

			It("With no summary", func() {
				userCGMSummary, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummary).To(BeNil())
			})

			It("With multiple summaries", func() {
				var summaries = []*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
					test.RandomCGMSummary(userId),
					test.RandomCGMSummary(userIdOther),
				}

				_, err = cgmStore.CreateSummaries(ctx, summaries)
				Expect(err).ToNot(HaveOccurred())

				userCGMSummary, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummary).ToNot(BeNil())

				summaries[0].ID = userCGMSummary.ID
				Expect(userCGMSummary).To(Equal(summaries[0]))
			})

			It("Get with multiple summaries of different type", func() {
				bgmStore := dataStoreSummary.NewSummaries[*types.BGMPeriods, *types.GlucoseBucket](summaryRepository)
				continuousStore := dataStoreSummary.NewSummaries[*types.ContinuousPeriods, *types.ContinuousBucket](summaryRepository)

				var cgmSummaries = []*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
					test.RandomCGMSummary(userId),
					test.RandomCGMSummary(userIdOther),
				}

				var bgmSummaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
				}

				var continuousSummaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
				}

				_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
				Expect(err).ToNot(HaveOccurred())

				userCGMSummary, err = cgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userCGMSummary).ToNot(BeNil())

				cgmSummaries[0].ID = userCGMSummary.ID
				Expect(userCGMSummary).To(Equal(cgmSummaries[0]))
			})
		})

		Context("GetOutdatedUserIDs", func() {
			var userIds *types.OutdatedSummariesResponse
			var userIdTwo string
			var userIdThree string

			BeforeEach(func() {
				userIdTwo = userTest.RandomUserID()
				userIdThree = userTest.RandomUserID()
			})

			It("With missing context", func() {
				userIds, err = cgmStore.GetOutdatedUserIDs(nil, page.NewPagination())
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
				Expect(userIds).To(BeNil())
			})

			It("With missing pagination", func() {
				userIds, err = cgmStore.GetOutdatedUserIDs(ctx, nil)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("pagination is missing"))
				Expect(userIds).To(BeNil())
			})

			It("With no outdated summaries", func() {
				var pagination = page.NewPagination()

				userIds, err = cgmStore.GetOutdatedUserIDs(ctx, pagination)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userIds.UserIds)).To(Equal(0))
			})

			It("With outdated CGM summaries", func() {
				var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
				var cgmSummaries = []*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
					test.RandomCGMSummary(userId),
					test.RandomCGMSummary(userIdOther),
					test.RandomCGMSummary(userIdTwo),
				}

				// mark 2/3 summaries outdated
				cgmSummaries[0].Dates.OutdatedSince = &outdatedTime
				cgmSummaries[1].Dates.OutdatedSince = nil
				cgmSummaries[2].Dates.OutdatedSince = &outdatedTime
				_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = cgmStore.GetOutdatedUserIDs(ctx, page.NewPagination())
				Expect(err).ToNot(HaveOccurred())
				Expect(userIds.UserIds).To(ConsistOf([]string{userId, userIdTwo}))
			})

			It("With a specific pagination size", func() {
				var pagination = page.NewPagination()
				var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
				var cgmSummaries = []*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
					test.RandomCGMSummary(userId),
					test.RandomCGMSummary(userIdOther),
					test.RandomCGMSummary(userIdTwo),
					test.RandomCGMSummary(userIdThree),
				}

				pagination.Size = 3

				for i := len(cgmSummaries) - 1; i >= 0; i-- {
					cgmSummaries[i].Dates.OutdatedSince = pointer.FromAny(outdatedTime.Add(-time.Duration(i) * time.Second))
				}
				_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = cgmStore.GetOutdatedUserIDs(ctx, pagination)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userIds.UserIds)).To(Equal(3))
				Expect(userIds.UserIds).To(ConsistOf([]string{userIdThree, userIdTwo, userIdOther}))
			})

			It("Check sort order", func() {
				var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
				var cgmSummaries = []*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
					test.RandomCGMSummary(userId),
					test.RandomCGMSummary(userIdOther),
					test.RandomCGMSummary(userIdTwo),
				}

				for i := 0; i < len(cgmSummaries); i++ {
					cgmSummaries[i].Dates.OutdatedSince = pointer.FromAny(outdatedTime.Add(time.Duration(-i) * time.Minute))
				}
				_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = cgmStore.GetOutdatedUserIDs(ctx, page.NewPagination())
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userIds.UserIds)).To(Equal(3))

				// we expect these to come back in reverse order than inserted
				for i := 0; i < len(userIds.UserIds); i++ {
					Expect(userIds.UserIds[i]).To(Equal(cgmSummaries[len(cgmSummaries)-i-1].UserID))
				}
			})

			It("Get outdated summaries with all types present", func() {
				userIdFour := userTest.RandomUserID()
				userIdFive := userTest.RandomUserID()
				continuousStore := dataStoreSummary.NewSummaries[*types.ContinuousPeriods, *types.ContinuousBucket](summaryRepository)
				bgmStore := dataStoreSummary.NewSummaries[*types.BGMPeriods, *types.GlucoseBucket](summaryRepository)

				var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)

				var cgmSummaries = []*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
					test.RandomCGMSummary(userId),
					test.RandomCGMSummary(userIdOther),
				}

				var bgmSummaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userIdTwo),
					test.RandomBGMSummary(userIdThree),
				}

				var continuousSummaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userIdFour),
					test.RandomContinuousSummary(userIdFive),
				}

				// mark 1 outdated per type
				cgmSummaries[0].Dates.OutdatedSince = &outdatedTime
				cgmSummaries[1].Dates.OutdatedSince = nil
				_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				bgmSummaries[0].Dates.OutdatedSince = nil
				bgmSummaries[1].Dates.OutdatedSince = &outdatedTime
				_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				continuousSummaries[0].Dates.OutdatedSince = nil
				continuousSummaries[1].Dates.OutdatedSince = &outdatedTime
				_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = cgmStore.GetOutdatedUserIDs(ctx, page.NewPagination())
				Expect(err).ToNot(HaveOccurred())
				Expect(userIds.UserIds).To(ConsistOf([]string{userId}))
			})
		})
	})
})
