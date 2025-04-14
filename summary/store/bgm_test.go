package store_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	dataStoreSummary "github.com/tidepool-org/platform/summary/store"
	"github.com/tidepool-org/platform/summary/test"
	"github.com/tidepool-org/platform/summary/types"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("BGM", Label("mongodb", "slow", "integration"), func() {
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
		var config *storeStructuredMongo.Config
		var createStore *dataStoreMongo.Store

		BeforeEach(func() {
			config = storeStructuredMongoTest.NewConfig()
		})

		AfterEach(func() {
			if createStore != nil {
				_ = createStore.Terminate(ctx)
			}
		})

		It("Repo", func() {
			createStore, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(createStore).ToNot(BeNil())

			summaryRepository = createStore.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			bgmStore := dataStoreSummary.NewSummaries[*types.BGMPeriods, *types.GlucoseBucket](summaryRepository)
			Expect(bgmStore).ToNot(BeNil())
		})
	})

	Context("Store", func() {
		var summaryCollection *mongo.Collection
		var userId string
		var userIdOther string
		var userBGMSummary *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]
		var bgmStore *dataStoreSummary.Summaries[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]

		BeforeEach(func() {
			config := storeStructuredMongoTest.NewConfig()
			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			summaryCollection = store.GetCollection("summary")
			summaryRepository = store.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			bgmStore = dataStoreSummary.NewSummaries[*types.BGMPeriods, *types.GlucoseBucket](summaryRepository)

			userId = userTest.RandomID()
			userIdOther = userTest.RandomID()
		})

		AfterEach(func() {
			if summaryCollection != nil {
				_, err = summaryCollection.DeleteMany(ctx, bson.D{})
				Expect(err).To(Succeed())
			}
			if store != nil {
				_ = store.Terminate(ctx)
			}
		})

		Context("ReplaceSummary", func() {
			It("Insert Summary with missing Type", func() {
				userBGMSummary = test.RandomBGMSummary(userId)
				userBGMSummary.Type = ""

				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type '', expected 'bgm'"))
			})

			It("Insert Summary with invalid Type", func() {
				userBGMSummary = test.RandomBGMSummary(userId)
				userBGMSummary.Type = "asdf"

				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type 'asdf', expected 'bgm'"))
			})

			It("Insert Summary", func() {
				userBGMSummary = test.RandomBGMSummary(userId)
				Expect(userBGMSummary.Type).To(Equal("bgm"))

				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).ToNot(HaveOccurred())

				userBGMSummaryWritten, err := bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// copy id, as that was mongo generated
				userBGMSummary.ID = userBGMSummaryWritten.ID
				Expect(userBGMSummaryWritten).To(Equal(userBGMSummary))
			})

			It("Update Summary", func() {
				var userBGMSummaryTwo *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]
				var userBGMSummaryWritten *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]
				var userBGMSummaryWrittenTwo *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]

				// generate and insert first summary
				userBGMSummary = test.RandomBGMSummary(userId)
				Expect(userBGMSummary.Type).To(Equal("bgm"))

				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).ToNot(HaveOccurred())

				// confirm first summary was written, get ID
				userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// copy id, as that was mongo generated
				userBGMSummary.ID = userBGMSummaryWritten.ID
				Expect(userBGMSummaryWritten).To(Equal(userBGMSummary))

				// generate a new summary with same type and user, and upsert
				userBGMSummaryTwo = test.RandomBGMSummary(userId)
				err = bgmStore.ReplaceSummary(ctx, userBGMSummaryTwo)
				Expect(err).ToNot(HaveOccurred())

				userBGMSummaryWrittenTwo, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// confirm the ID was unchanged
				Expect(userBGMSummaryWrittenTwo.ID).To(Equal(userBGMSummaryWritten.ID))

				// confirm the written summary matches the new summary
				userBGMSummaryTwo.ID = userBGMSummaryWritten.ID
				Expect(userBGMSummaryWrittenTwo).To(Equal(userBGMSummaryTwo))
			})
		})

		Context("DeleteSummary", func() {
			It("Delete Summary with empty context", func() {
				err = bgmStore.DeleteSummary(nil, userId)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
			})

			It("Delete Summary with empty userId", func() {
				err = bgmStore.DeleteSummary(ctx, "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing"))
			})

			It("Delete Summary", func() {
				var userBGMSummaryWritten *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]

				userBGMSummary = test.RandomBGMSummary(userId)
				Expect(userBGMSummary.Type).To(Equal("bgm"))

				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).ToNot(HaveOccurred())

				// confirm writes
				userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummaryWritten).ToNot(BeNil())

				// delete
				err = bgmStore.DeleteSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// confirm delete
				userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummaryWritten).To(BeNil())
			})
		})

		Context("CreateSummaries", func() {
			It("Create summaries with missing context", func() {
				var summaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
				}

				_, err = bgmStore.CreateSummaries(nil, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
			})

			It("Create summaries with missing summaries", func() {
				_, err = bgmStore.CreateSummaries(ctx, nil)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("summaries for create missing"))
			})

			It("Create summaries with an invalid type", func() {
				var summaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
				}

				summaries[0].Type = "cgm"

				_, err = bgmStore.CreateSummaries(ctx, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type 'cgm', expected 'bgm' at index 0"))
			})

			It("Create summaries with an invalid type", func() {
				var summaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
				}

				summaries[0].Type = "cgm"

				_, err = bgmStore.CreateSummaries(ctx, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type 'cgm', expected 'bgm' at index 0"))
			})

			It("Create summaries with an empty userId", func() {
				var summaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
				}

				summaries[0].UserID = ""

				_, err = bgmStore.CreateSummaries(ctx, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing at index 0"))
			})

			It("Create summaries", func() {
				var count int
				var summaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
				}

				count, err = bgmStore.CreateSummaries(ctx, summaries)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(2))

				for i := 0; i < 2; i++ {
					userBGMSummary, err = bgmStore.GetSummary(ctx, summaries[0].UserID)
					Expect(err).ToNot(HaveOccurred())
					Expect(userBGMSummary).ToNot(BeNil())
					summaries[i].ID = userBGMSummary.ID
					Expect(userBGMSummary).To(Equal(summaries[0]))
				}
			})
		})

		Context("SetOutdated", func() {
			var outdatedSince *time.Time
			var userBGMSummaryWritten *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]

			It("With missing context", func() {
				outdatedSince, err = bgmStore.SetOutdated(nil, userId, types.OutdatedReasonDataAdded)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
				Expect(outdatedSince).To(BeNil())
			})

			It("With missing userId", func() {
				outdatedSince, err = bgmStore.SetOutdated(ctx, "", types.OutdatedReasonDataAdded)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing"))
				Expect(outdatedSince).To(BeNil())
			})

			It("With multiple reasons", func() {
				outdatedSinceOriginal, err := bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSinceOriginal).ToNot(BeNil())

				userBGMSummary, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummary.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userBGMSummary.Dates.OutdatedSince).To(Equal(outdatedSinceOriginal))
				Expect(userBGMSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded}))

				outdatedSince, err = bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonSchemaMigration)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userBGMSummary, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummary.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userBGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
				Expect(userBGMSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonSchemaMigration}))

				outdatedSince, err = bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userBGMSummary, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummary.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userBGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
				Expect(userBGMSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonSchemaMigration}))
			})

			It("With no existing summary", func() {
				outdatedSince, err = bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userBGMSummary, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummary.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userBGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
			})

			It("With an existing non-outdated summary", func() {
				userBGMSummary = test.RandomBGMSummary(userId)
				userBGMSummary.Dates.OutdatedSince = nil
				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).ToNot(HaveOccurred())

				outdatedSince, err = bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userBGMSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))

			})

			It("With an existing outdated summary", func() {
				var fiveMinutesAgo = time.Now().Add(time.Duration(-5) * time.Minute).UTC().Truncate(time.Millisecond)

				userBGMSummary = test.RandomBGMSummary(userId)
				userBGMSummary.Dates.OutdatedSince = &fiveMinutesAgo
				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).ToNot(HaveOccurred())

				outdatedSince, err = bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userBGMSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))
			})

			It("With an existing outdated summary beyond the outdatedSinceLimit", func() {
				now := time.Now().UTC().Truncate(time.Millisecond)

				userBGMSummary = test.RandomBGMSummary(userId)
				userBGMSummary.Dates.OutdatedSince = &now
				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).ToNot(HaveOccurred())

				outdatedSince, err = bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
			})

			It("With an existing outdated summary with schema migration reason", func() {
				now := time.Now().UTC().Truncate(time.Millisecond)
				fiveMinutesAgo := now.Add(time.Duration(-5) * time.Minute)

				userBGMSummary = test.RandomBGMSummary(userId)
				userBGMSummary.Dates.OutdatedSince = &fiveMinutesAgo
				userBGMSummary.Dates.OutdatedReason = []string{types.OutdatedReasonUploadCompleted}
				Expect(userBGMSummary.Periods.GlucosePeriods).ToNot(HaveLen(0))

				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).ToNot(HaveOccurred())

				outdatedSince, err = bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonSchemaMigration)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userBGMSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))
				Expect(userBGMSummaryWritten.Periods.GlucosePeriods).ToNot(HaveLen(0))
				Expect(userBGMSummaryWritten.Dates.LastData).To(Equal(userBGMSummary.Dates.LastData))
				Expect(userBGMSummaryWritten.Dates.FirstData).To(Equal(userBGMSummary.Dates.FirstData))
				Expect(userBGMSummaryWritten.Dates.LastUpdatedDate).To(Equal(userBGMSummary.Dates.LastUpdatedDate))
				Expect(userBGMSummaryWritten.Dates.LastUploadDate).To(Equal(userBGMSummary.Dates.LastUploadDate))
				Expect(userBGMSummaryWritten.Dates.OutdatedReason).To(ConsistOf(types.OutdatedReasonSchemaMigration, types.OutdatedReasonUploadCompleted))
			})
		})

		Context("GetSummary", func() {
			It("With missing context", func() {
				userBGMSummary, err = bgmStore.GetSummary(nil, userId)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
				Expect(userBGMSummary).To(BeNil())
			})

			It("With missing userId", func() {
				userBGMSummary, err = bgmStore.GetSummary(ctx, "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing"))
				Expect(userBGMSummary).To(BeNil())
			})

			It("With no summary", func() {
				userBGMSummary, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummary).To(BeNil())
			})

			It("With multiple summaries", func() {
				var summaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
				}

				_, err = bgmStore.CreateSummaries(ctx, summaries)
				Expect(err).ToNot(HaveOccurred())

				userBGMSummary, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummary).ToNot(BeNil())

				summaries[0].ID = userBGMSummary.ID
				Expect(userBGMSummary).To(Equal(summaries[0]))
			})

			It("Get with multiple summaries of different type a", func() {
				cgmStore := dataStoreSummary.NewSummaries[*types.CGMPeriods, *types.GlucoseBucket](summaryRepository)
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

				userBGMSummary, err = bgmStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummary).ToNot(BeNil())

				bgmSummaries[0].ID = userBGMSummary.ID
				Expect(userBGMSummary).To(Equal(bgmSummaries[0]))
			})
		})

		Context("GetOutdatedUserIDs", func() {
			var userIds *types.OutdatedSummariesResponse
			var userIdTwo string
			var userIdThree string

			BeforeEach(func() {
				userIdTwo = userTest.RandomID()
				userIdThree = userTest.RandomID()
			})

			It("With missing context", func() {
				userIds, err = bgmStore.GetOutdatedUserIDs(nil, page.NewPagination())
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
				Expect(userIds).To(BeNil())
			})

			It("With missing pagination", func() {
				userIds, err = bgmStore.GetOutdatedUserIDs(ctx, nil)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("pagination is missing"))
				Expect(userIds).To(BeNil())
			})

			It("With no outdated summaries", func() {
				var pagination = page.NewPagination()

				userIds, err = bgmStore.GetOutdatedUserIDs(ctx, pagination)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userIds.UserIds)).To(Equal(0))
			})

			It("With outdated CGM summaries", func() {
				var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
				var bgmSummaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
					test.RandomBGMSummary(userIdTwo),
				}

				// mark 2/3 summaries outdated
				bgmSummaries[0].Dates.OutdatedSince = &outdatedTime
				bgmSummaries[1].Dates.OutdatedSince = nil
				bgmSummaries[2].Dates.OutdatedSince = &outdatedTime
				_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = bgmStore.GetOutdatedUserIDs(ctx, page.NewPagination())
				Expect(err).ToNot(HaveOccurred())
				Expect(userIds.UserIds).To(ConsistOf([]string{userId, userIdTwo}))
			})

			It("With a specific pagination size", func() {
				var pagination = page.NewPagination()
				var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
				var bgmSummaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
					test.RandomBGMSummary(userIdTwo),
					test.RandomBGMSummary(userIdThree),
				}

				pagination.Size = 3

				for i := len(bgmSummaries) - 1; i >= 0; i-- {
					bgmSummaries[i].Dates.OutdatedSince = pointer.FromAny(outdatedTime.Add(-time.Duration(i) * time.Second))
				}
				_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = bgmStore.GetOutdatedUserIDs(ctx, pagination)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userIds.UserIds)).To(Equal(3))
				Expect(userIds.UserIds).To(ConsistOf([]string{userIdThree, userIdTwo, userIdOther}))
			})

			It("Check sort order", func() {
				var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
				var bgmSummaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
					test.RandomBGMSummary(userIdTwo),
				}

				for i := 0; i < len(bgmSummaries); i++ {
					bgmSummaries[i].Dates.OutdatedSince = pointer.FromAny(outdatedTime.Add(time.Duration(-i) * time.Minute))
				}
				_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = bgmStore.GetOutdatedUserIDs(ctx, page.NewPagination())
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userIds.UserIds)).To(Equal(3))

				// we expect these to come back in reverse order than inserted
				for i := 0; i < len(userIds.UserIds); i++ {
					Expect(userIds.UserIds[i]).To(Equal(bgmSummaries[len(bgmSummaries)-i-1].UserID))
				}
			})

			It("Get outdated summaries with all types present", func() {
				userIdFour := userTest.RandomID()
				userIdFive := userTest.RandomID()
				continuousStore := dataStoreSummary.NewSummaries[*types.ContinuousPeriods, *types.ContinuousBucket](summaryRepository)
				cgmStore := dataStoreSummary.NewSummaries[*types.CGMPeriods, *types.GlucoseBucket](summaryRepository)

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

				userIds, err = bgmStore.GetOutdatedUserIDs(ctx, page.NewPagination())
				Expect(err).ToNot(HaveOccurred())
				Expect(userIds.UserIds).To(ConsistOf([]string{userIdThree}))
			})
		})

		Context("GetMigratableUserIDs", func() {
			var userIds []string
			var userIdTwo string
			var userIdThree string

			BeforeEach(func() {
				userIdTwo = userTest.RandomID()
				userIdThree = userTest.RandomID()
			})

			It("With missing context", func() {
				userIds, err = bgmStore.GetMigratableUserIDs(nil, page.NewPagination())
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
				Expect(userIds).To(BeNil())
			})

			It("With missing pagination", func() {
				userIds, err = bgmStore.GetMigratableUserIDs(ctx, nil)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("pagination is missing"))
				Expect(userIds).To(BeNil())
			})

			It("With no migratable summaries", func() {
				var pagination = page.NewPagination()

				userIds, err = bgmStore.GetMigratableUserIDs(ctx, pagination)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userIds)).To(Equal(0))
			})

			It("With migratable CGM summaries", func() {
				var bgmSummaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
					test.RandomBGMSummary(userIdTwo),
				}

				// mark 2/3 summaries for migration
				bgmSummaries[0].Config.SchemaVersion = types.SchemaVersion - 1
				bgmSummaries[0].Dates.OutdatedSince = nil
				bgmSummaries[1].Config.SchemaVersion = types.SchemaVersion
				bgmSummaries[1].Dates.OutdatedSince = nil
				bgmSummaries[2].Config.SchemaVersion = types.SchemaVersion - 1
				bgmSummaries[2].Dates.OutdatedSince = nil
				_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = bgmStore.GetMigratableUserIDs(ctx, page.NewPagination())
				Expect(err).ToNot(HaveOccurred())
				Expect(userIds).To(ConsistOf([]string{userId, userIdTwo}))
			})

			It("With migratable and outdated CGM summaries", func() {
				var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
				var bgmSummaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
					test.RandomBGMSummary(userIdTwo),
				}

				// mark 2/3 summaries for migration, and 1/3 as outdated
				bgmSummaries[0].Config.SchemaVersion = types.SchemaVersion - 1
				bgmSummaries[0].Dates.OutdatedSince = nil
				bgmSummaries[1].Config.SchemaVersion = types.SchemaVersion - 1
				bgmSummaries[1].Dates.OutdatedSince = &outdatedTime
				bgmSummaries[2].Config.SchemaVersion = types.SchemaVersion - 1
				bgmSummaries[2].Dates.OutdatedSince = nil
				_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = bgmStore.GetMigratableUserIDs(ctx, page.NewPagination())
				Expect(err).ToNot(HaveOccurred())
				Expect(userIds).To(ConsistOf([]string{userId, userIdTwo}))
			})

			It("With a specific pagination size", func() {
				var lastUpdatedTime = time.Now().UTC().Truncate(time.Millisecond)
				var pagination = page.NewPagination()
				var bgmSummaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
					test.RandomBGMSummary(userIdTwo),
					test.RandomBGMSummary(userIdThree),
				}

				pagination.Size = 3

				for i := len(bgmSummaries) - 1; i >= 0; i-- {
					bgmSummaries[i].Config.SchemaVersion = types.SchemaVersion - 1
					bgmSummaries[i].Dates.OutdatedSince = nil
					bgmSummaries[i].Dates.LastUpdatedDate = lastUpdatedTime.Add(time.Duration(-i) * time.Minute)
				}
				_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = bgmStore.GetMigratableUserIDs(ctx, pagination)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userIds)).To(Equal(3))
				Expect(userIds).To(ConsistOf([]string{userIdThree, userIdTwo, userIdOther}))
			})

			It("Check sort order", func() {
				var lastUpdatedTime = time.Now().UTC().Truncate(time.Millisecond)
				var bgmSummaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
					test.RandomBGMSummary(userIdTwo),
				}

				for i := 0; i < len(bgmSummaries); i++ {
					bgmSummaries[i].Config.SchemaVersion = types.SchemaVersion - 1
					bgmSummaries[i].Dates.OutdatedSince = nil
					bgmSummaries[i].Dates.LastUpdatedDate = lastUpdatedTime.Add(time.Duration(-i) * time.Minute)
				}
				_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = bgmStore.GetMigratableUserIDs(ctx, page.NewPagination())
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userIds)).To(Equal(3))

				// we expect these to come back in reverse order than inserted
				for i := 0; i < len(userIds); i++ {
					Expect(userIds[i]).To(Equal(bgmSummaries[len(bgmSummaries)-i-1].UserID))
				}
			})

			It("Get migratable summaries with all types present", func() {
				userIdFour := userTest.RandomID()
				userIdFive := userTest.RandomID()
				continuousStore := dataStoreSummary.NewSummaries[*types.ContinuousPeriods, *types.ContinuousBucket](summaryRepository)
				cgmStore := dataStoreSummary.NewSummaries[*types.CGMPeriods, *types.GlucoseBucket](summaryRepository)

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
				cgmSummaries[0].Config.SchemaVersion = types.SchemaVersion - 1
				cgmSummaries[0].Dates.OutdatedSince = nil
				cgmSummaries[1].Config.SchemaVersion = types.SchemaVersion
				cgmSummaries[1].Dates.OutdatedSince = nil
				_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				bgmSummaries[0].Config.SchemaVersion = types.SchemaVersion
				bgmSummaries[0].Dates.OutdatedSince = nil
				bgmSummaries[1].Config.SchemaVersion = types.SchemaVersion - 1
				bgmSummaries[1].Dates.OutdatedSince = nil
				_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				continuousSummaries[0].Config.SchemaVersion = types.SchemaVersion
				continuousSummaries[0].Dates.OutdatedSince = nil
				continuousSummaries[1].Config.SchemaVersion = types.SchemaVersion - 1
				continuousSummaries[1].Dates.OutdatedSince = nil
				_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = bgmStore.GetMigratableUserIDs(ctx, page.NewPagination())
				Expect(err).ToNot(HaveOccurred())
				Expect(userIds).To(ConsistOf([]string{userIdThree}))
			})
		})
	})
})
