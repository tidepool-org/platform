package store_test

import (
	"context"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	dataStoreSummary "github.com/tidepool-org/platform/data/summary/store"
	"github.com/tidepool-org/platform/data/summary/test"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Continuous", Label("mongodb", "slow", "integration"), func() {
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

			bgmStore := dataStoreSummary.NewSummaries[*types.ContinuousPeriods, *types.ContinuousBucket](summaryRepository)
			Expect(bgmStore).ToNot(BeNil())
		})
	})

	Context("Store", func() {
		var summaryCollection *mongo.Collection
		var userId string
		var userIdOther string
		var userContinuousSummary *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]
		var continuousStore *dataStoreSummary.Summaries[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]

		BeforeEach(func() {
			config := storeStructuredMongoTest.NewConfig()
			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			summaryCollection = store.GetCollection("summary")
			summaryRepository = store.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			continuousStore = dataStoreSummary.NewSummaries[*types.ContinuousPeriods, *types.ContinuousBucket](summaryRepository)

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
				userContinuousSummary = test.RandomContinuousSummary(userId)
				userContinuousSummary.Type = ""

				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type '', expected 'con'"))
			})

			It("Insert Summary with invalid Type", func() {
				userContinuousSummary = test.RandomContinuousSummary(userId)
				userContinuousSummary.Type = "asdf"

				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type 'asdf', expected 'con'"))
			})

			It("Insert Summary", func() {
				userContinuousSummary = test.RandomContinuousSummary(userId)
				Expect(userContinuousSummary.Type).To(Equal("con"))

				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).ToNot(HaveOccurred())

				userContinuousSummaryWritten, err := continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// copy id, as that was mongo generated
				userContinuousSummary.ID = userContinuousSummaryWritten.ID
				Expect(userContinuousSummaryWritten).To(Equal(userContinuousSummary))
			})

			It("Update Summary", func() {
				var userContinuousSummaryTwo *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]
				var userContinuousSummaryWritten *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]
				var userContinuousSummaryWrittenTwo *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]

				// generate and insert first summary
				userContinuousSummary = test.RandomContinuousSummary(userId)
				Expect(userContinuousSummary.Type).To(Equal("con"))

				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).ToNot(HaveOccurred())

				// confirm first summary was written, get ID
				userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// copy id, as that was mongo generated
				userContinuousSummary.ID = userContinuousSummaryWritten.ID
				Expect(userContinuousSummaryWritten).To(Equal(userContinuousSummary))

				// generate a new summary with same type and user, and upsert
				userContinuousSummaryTwo = test.RandomContinuousSummary(userId)
				err = continuousStore.ReplaceSummary(ctx, userContinuousSummaryTwo)
				Expect(err).ToNot(HaveOccurred())

				userContinuousSummaryWrittenTwo, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// confirm the ID was unchanged
				Expect(userContinuousSummaryWrittenTwo.ID).To(Equal(userContinuousSummaryWritten.ID))

				// confirm the written summary matches the new summary
				userContinuousSummaryWrittenTwo.ID = userContinuousSummaryTwo.ID
				opts := cmpopts.IgnoreUnexported(types.ContinuousPeriod{})
				Expect(userContinuousSummaryWrittenTwo).To(BeComparableTo(userContinuousSummaryTwo, opts))
			})
		})

		Context("DeleteSummary", func() {
			It("Delete Summary with empty context", func() {
				err = continuousStore.DeleteSummary(nil, userId)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
			})

			It("Delete Summary with empty userId", func() {
				err = continuousStore.DeleteSummary(ctx, "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing"))
			})

			It("Delete Summary", func() {
				var userContinuousSummaryWritten *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]

				userContinuousSummary = test.RandomContinuousSummary(userId)
				Expect(userContinuousSummary.Type).To(Equal("con"))

				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).ToNot(HaveOccurred())

				// confirm writes
				userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummaryWritten).ToNot(BeNil())

				// delete
				err = continuousStore.DeleteSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// confirm delete
				userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummaryWritten).To(BeNil())
			})
		})

		Context("CreateSummaries", func() {
			It("Create summaries with missing context", func() {
				var summaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
				}

				_, err = continuousStore.CreateSummaries(nil, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
			})

			It("Create summaries with missing summaries", func() {
				_, err = continuousStore.CreateSummaries(ctx, nil)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("summaries for create missing"))
			})

			It("Create summaries with an invalid type", func() {
				var summaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
				}

				summaries[0].Type = "bgm"

				_, err = continuousStore.CreateSummaries(ctx, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type 'bgm', expected 'con' at index 0"))
			})

			It("Create summaries with an empty userId", func() {
				var summaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
				}

				summaries[0].UserID = ""

				_, err = continuousStore.CreateSummaries(ctx, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing at index 0"))
			})

			It("Create summaries", func() {
				var count int
				var summaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
				}

				count, err = continuousStore.CreateSummaries(ctx, summaries)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(2))

				for i := 0; i < 2; i++ {
					userContinuousSummary, err = continuousStore.GetSummary(ctx, summaries[0].UserID)
					Expect(err).ToNot(HaveOccurred())
					Expect(userContinuousSummary).ToNot(BeNil())
					summaries[i].ID = userContinuousSummary.ID
					Expect(userContinuousSummary).To(Equal(summaries[0]))
				}
			})
		})

		Context("SetOutdated", func() {
			var outdatedSince *time.Time
			var userContinuousSummaryWritten *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]

			It("With missing context", func() {
				outdatedSince, err = continuousStore.SetOutdated(nil, userId, types.OutdatedReasonDataAdded)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
				Expect(outdatedSince).To(BeNil())
			})

			It("With missing userId", func() {
				outdatedSince, err = continuousStore.SetOutdated(ctx, "", types.OutdatedReasonDataAdded)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing"))
				Expect(outdatedSince).To(BeNil())
			})

			It("With multiple reasons", func() {
				outdatedSinceOriginal, err := continuousStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSinceOriginal).ToNot(BeNil())

				userContinuousSummary, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummary.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userContinuousSummary.Dates.OutdatedSince).To(Equal(outdatedSinceOriginal))
				Expect(userContinuousSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded}))

				outdatedSince, err = continuousStore.SetOutdated(ctx, userId, types.OutdatedReasonSchemaMigration)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userContinuousSummary, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummary.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userContinuousSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
				Expect(userContinuousSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonSchemaMigration}))

				outdatedSince, err = continuousStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userContinuousSummary, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummary.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userContinuousSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
				Expect(userContinuousSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonSchemaMigration}))
			})

			It("With no existing summary", func() {
				outdatedSince, err = continuousStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userContinuousSummary, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummary.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userContinuousSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
			})

			It("With an existing non-outdated summary", func() {
				userContinuousSummary = test.RandomContinuousSummary(userId)
				userContinuousSummary.Dates.OutdatedSince = nil
				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).ToNot(HaveOccurred())

				outdatedSince, err = continuousStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userContinuousSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))

			})

			It("With an existing outdated summary", func() {
				var fiveMinutesAgo = time.Now().Add(time.Duration(-5) * time.Minute).UTC().Truncate(time.Millisecond)

				userContinuousSummary = test.RandomContinuousSummary(userId)
				userContinuousSummary.Dates.OutdatedSince = &fiveMinutesAgo
				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).ToNot(HaveOccurred())

				outdatedSince, err = continuousStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userContinuousSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))
			})

			It("With an existing outdated summary beyond the outdatedSinceLimit", func() {
				now := time.Now().UTC().Truncate(time.Millisecond)

				userContinuousSummary = test.RandomContinuousSummary(userId)
				userContinuousSummary.Dates.OutdatedSince = &now
				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).ToNot(HaveOccurred())

				outdatedSince, err = continuousStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
			})

			It("With an existing outdated summary with schema migration reason", func() {
				now := time.Now().UTC().Truncate(time.Millisecond)
				fiveMinutesAgo := now.Add(time.Duration(-5) * time.Minute)

				userContinuousSummary = test.RandomContinuousSummary(userId)
				userContinuousSummary.Dates.OutdatedSince = &fiveMinutesAgo
				userContinuousSummary.Dates.OutdatedReason = []string{types.OutdatedReasonUploadCompleted}
				Expect(*userContinuousSummary.Periods).ToNot(HaveLen(0))

				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).ToNot(HaveOccurred())

				outdatedSince, err = continuousStore.SetOutdated(ctx, userId, types.OutdatedReasonSchemaMigration)
				Expect(err).ToNot(HaveOccurred())
				Expect(outdatedSince).ToNot(BeNil())

				userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
				Expect(userContinuousSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))
				Expect(*userContinuousSummaryWritten.Periods).To(HaveLen(0))
				Expect(userContinuousSummaryWritten.Dates.LastData).To(BeZero())
				Expect(userContinuousSummaryWritten.Dates.FirstData).To(BeZero())
				Expect(userContinuousSummaryWritten.Dates.LastUpdatedDate).To(BeZero())
				Expect(userContinuousSummaryWritten.Dates.LastUploadDate).To(BeZero())
				Expect(userContinuousSummaryWritten.Dates.OutdatedReason).To(ConsistOf(types.OutdatedReasonSchemaMigration, types.OutdatedReasonUploadCompleted))
			})
		})

		Context("GetSummary", func() {
			It("With missing context", func() {
				userContinuousSummary, err = continuousStore.GetSummary(nil, userId)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
				Expect(userContinuousSummary).To(BeNil())
			})

			It("With missing userId", func() {
				userContinuousSummary, err = continuousStore.GetSummary(ctx, "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing"))
				Expect(userContinuousSummary).To(BeNil())
			})

			It("With no summary", func() {
				userContinuousSummary, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummary).To(BeNil())
			})

			It("With multiple summaries", func() {
				var summaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
				}

				_, err = continuousStore.CreateSummaries(ctx, summaries)
				Expect(err).ToNot(HaveOccurred())

				userContinuousSummary, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummary).ToNot(BeNil())

				summaries[0].ID = userContinuousSummary.ID
				Expect(userContinuousSummary).To(Equal(summaries[0]))
			})

			It("Get with multiple summaries of different type", func() {
				cgmStore := dataStoreSummary.NewSummaries[*types.CGMPeriods, *types.GlucoseBucket](summaryRepository)
				bgmStore := dataStoreSummary.NewSummaries[*types.BGMPeriods, *types.GlucoseBucket](summaryRepository)

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

				userContinuousSummary, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummary).ToNot(BeNil())

				continuousSummaries[0].ID = userContinuousSummary.ID
				opts := cmpopts.IgnoreUnexported(types.ContinuousPeriod{})
				Expect(userContinuousSummary).To(BeComparableTo(continuousSummaries[0], opts))
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
				userIds, err = continuousStore.GetOutdatedUserIDs(nil, page.NewPagination())
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
				Expect(userIds).To(BeNil())
			})

			It("With missing pagination", func() {
				userIds, err = continuousStore.GetOutdatedUserIDs(ctx, nil)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("pagination is missing"))
				Expect(userIds).To(BeNil())
			})

			It("With no outdated summaries", func() {
				var pagination = page.NewPagination()

				userIds, err = continuousStore.GetOutdatedUserIDs(ctx, pagination)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userIds.UserIds)).To(Equal(0))
			})

			It("With outdated CGM summaries", func() {
				var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
				var continuousSummaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
					test.RandomContinuousSummary(userIdTwo),
				}

				// mark 2/3 summaries outdated
				continuousSummaries[0].Dates.OutdatedSince = &outdatedTime
				continuousSummaries[1].Dates.OutdatedSince = nil
				continuousSummaries[2].Dates.OutdatedSince = &outdatedTime
				_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = continuousStore.GetOutdatedUserIDs(ctx, page.NewPagination())
				Expect(err).ToNot(HaveOccurred())
				Expect(userIds.UserIds).To(ConsistOf([]string{userId, userIdTwo}))
			})

			It("With a specific pagination size", func() {
				var pagination = page.NewPagination()
				var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
				var continuousSummaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
					test.RandomContinuousSummary(userIdTwo),
					test.RandomContinuousSummary(userIdThree),
				}

				pagination.Size = 3

				for i := len(continuousSummaries) - 1; i >= 0; i-- {
					continuousSummaries[i].Dates.OutdatedSince = pointer.FromAny(outdatedTime.Add(-time.Duration(i) * time.Second))
				}
				_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = continuousStore.GetOutdatedUserIDs(ctx, pagination)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userIds.UserIds)).To(Equal(3))
				Expect(userIds.UserIds).To(ConsistOf([]string{userIdThree, userIdTwo, userIdOther}))
			})

			It("Check sort order", func() {
				var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
				var continuousSummaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
					test.RandomContinuousSummary(userIdTwo),
				}

				for i := 0; i < len(continuousSummaries); i++ {
					continuousSummaries[i].Dates.OutdatedSince = pointer.FromAny(outdatedTime.Add(time.Duration(-i) * time.Minute))
				}
				_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
				Expect(err).ToNot(HaveOccurred())

				userIds, err = continuousStore.GetOutdatedUserIDs(ctx, page.NewPagination())
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userIds.UserIds)).To(Equal(3))

				// we expect these to come back in reverse order than inserted
				for i := 0; i < len(userIds.UserIds); i++ {
					Expect(userIds.UserIds[i]).To(Equal(continuousSummaries[len(continuousSummaries)-i-1].UserID))
				}
			})

			It("Get outdated summaries with all types present", func() {
				userIdFour := userTest.RandomID()
				userIdFive := userTest.RandomID()
				cgmStore := dataStoreSummary.NewSummaries[*types.CGMPeriods, *types.GlucoseBucket](summaryRepository)
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

				userIds, err = continuousStore.GetOutdatedUserIDs(ctx, page.NewPagination())
				Expect(err).ToNot(HaveOccurred())
				Expect(userIds.UserIds).To(ConsistOf([]string{userIdFive}))
			})
		})
	})
})
