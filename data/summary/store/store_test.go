package store_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	dataStoreSummary "github.com/tidepool-org/platform/data/summary/store"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/data/summary/types/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Summary Stats Mongo", Label("mongodb", "slow", "integration"), func() {
	var logger *logTest.Logger
	var err error
	var ctx context.Context

	var store *dataStoreMongo.Store
	var summaryRepository *storeStructuredMongo.Repository

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
	})

	Context("Create Stores", func() {
		var config *storeStructuredMongo.Config
		var createStore *dataStoreMongo.Store

		BeforeEach(func() {
			config = storeStructuredMongoTest.NewConfig()
		})

		AfterEach(func() {
			if createStore != nil {
				_ = createStore.Terminate(context.Background())
			}
		})

		It("CGM Repo", func() {
			createStore, err := dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(createStore).ToNot(BeNil())

			summaryRepository = createStore.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			cgmStore := dataStoreSummary.New[*types.CGMStats](summaryRepository)
			Expect(cgmStore).ToNot(BeNil())
		})

		It("BGM Repo", func() {
			createStore, err := dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(createStore).ToNot(BeNil())

			summaryRepository = createStore.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			bgmStore := dataStoreSummary.New[*types.BGMStats](summaryRepository)
			Expect(bgmStore).ToNot(BeNil())
		})

		It("Continuous Repo", func() {
			createStore, err := dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(createStore).ToNot(BeNil())

			summaryRepository = createStore.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			continuousStore := dataStoreSummary.New[*types.ContinuousStats](summaryRepository)
			Expect(continuousStore).ToNot(BeNil())
		})
	})

	Context("With a new store", func() {
		var summaryCollection *mongo.Collection

		BeforeEach(func() {
			store = GetSuiteStore()
			summaryCollection = store.GetCollection("summary")
		})

		AfterEach(func() {
			if summaryCollection != nil {
				_, err = summaryCollection.DeleteMany(context.Background(), bson.D{})
				Expect(err).To(Succeed())
			}
		})

		Context("With a repository", func() {
			var userId string
			var userIdOther string
			var typelessStore *dataStoreSummary.TypelessRepo

			BeforeEach(func() {
				summaryRepository = store.NewSummaryRepository().GetStore()
				Expect(summaryRepository).ToNot(BeNil())

				userId = userTest.RandomID()
				userIdOther = userTest.RandomID()
				typelessStore = dataStoreSummary.NewTypeless(summaryRepository)
			})

			AfterEach(func() {
				_, err = summaryCollection.DeleteMany(ctx, bson.D{})
				Expect(err).ToNot(HaveOccurred())
			})

			Context("Continuous", func() {
				var continuousStore *dataStoreSummary.Repo[*types.ContinuousStats, types.ContinuousStats]
				var userContinuousSummary *types.Summary[*types.ContinuousStats, types.ContinuousStats]

				BeforeEach(func() {
					continuousStore = dataStoreSummary.New[*types.ContinuousStats](summaryRepository)
				})

				Context("ReplaceSummary", func() {
					It("Insert Summary with missing Type", func() {
						userContinuousSummary = test.RandomContinousSummary(userId)
						userContinuousSummary.Type = ""

						err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("invalid summary type '', expected 'continuous'"))
					})

					It("Insert Summary with invalid Type", func() {
						userContinuousSummary = test.RandomContinousSummary(userId)
						userContinuousSummary.Type = "asdf"

						err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("invalid summary type 'asdf', expected 'continuous'"))
					})

					It("Insert Summary", func() {
						userContinuousSummary = test.RandomContinousSummary(userId)
						Expect(userContinuousSummary.Type).To(Equal("continuous"))

						err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
						Expect(err).ToNot(HaveOccurred())

						userContinuousSummaryWritten, err := continuousStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())

						// copy id, as that was mongo generated
						userContinuousSummary.ID = userContinuousSummaryWritten.ID
						Expect(userContinuousSummaryWritten).To(Equal(userContinuousSummary))
					})

					It("Update Summary", func() {
						var userContinuousSummaryTwo *types.Summary[*types.ContinuousStats, types.ContinuousStats]
						var userContinuousSummaryWritten *types.Summary[*types.ContinuousStats, types.ContinuousStats]
						var userContinuousSummaryWrittenTwo *types.Summary[*types.ContinuousStats, types.ContinuousStats]

						// generate and insert first summary
						userContinuousSummary = test.RandomContinousSummary(userId)
						Expect(userContinuousSummary.Type).To(Equal("continuous"))

						err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
						Expect(err).ToNot(HaveOccurred())

						// confirm first summary was written, get ID
						userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())

						// copy id, as that was mongo generated
						userContinuousSummary.ID = userContinuousSummaryWritten.ID
						Expect(userContinuousSummaryWritten).To(Equal(userContinuousSummary))

						// generate a new summary with same type and user, and upsert
						userContinuousSummaryTwo = test.RandomContinousSummary(userId)
						err = continuousStore.ReplaceSummary(ctx, userContinuousSummaryTwo)
						Expect(err).ToNot(HaveOccurred())

						userContinuousSummaryWrittenTwo, err = continuousStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())

						// confirm the ID was unchanged
						Expect(userContinuousSummaryWrittenTwo.ID).To(Equal(userContinuousSummaryWritten.ID))

						// confirm the written summary matches the new summary
						userContinuousSummaryWrittenTwo.ID = userContinuousSummaryTwo.ID
						Expect(userContinuousSummaryWrittenTwo).To(BeComparableTo(userContinuousSummaryTwo))
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
						var userContinuousSummaryWritten *types.Summary[*types.ContinuousStats, types.ContinuousStats]

						userContinuousSummary = test.RandomContinousSummary(userId)
						Expect(userContinuousSummary.Type).To(Equal("continuous"))

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
						var summaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userId),
							test.RandomContinousSummary(userIdOther),
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
						var summaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userId),
							test.RandomContinousSummary(userIdOther),
						}

						summaries[0].Type = "bgm"

						_, err = continuousStore.CreateSummaries(ctx, summaries)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("invalid summary type 'bgm', expected 'continuous' at index 0"))
					})

					It("Create summaries with an empty userId", func() {
						var summaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userId),
							test.RandomContinousSummary(userIdOther),
						}

						summaries[0].UserID = ""

						_, err = continuousStore.CreateSummaries(ctx, summaries)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("userId is missing at index 0"))
					})

					It("Create summaries", func() {
						var count int
						var summaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userId),
							test.RandomContinousSummary(userIdOther),
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
					var userContinuousSummaryWritten *types.Summary[*types.ContinuousStats, types.ContinuousStats]

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

						outdatedSince, err = continuousStore.SetOutdated(ctx, userId, types.OutdatedReasonBackfill)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userContinuousSummary, err = continuousStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userContinuousSummary.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userContinuousSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(userContinuousSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonBackfill}))

						outdatedSince, err = continuousStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userContinuousSummary, err = continuousStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userContinuousSummary.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userContinuousSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(userContinuousSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonBackfill}))
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
						userContinuousSummary = test.RandomContinousSummary(userId)
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

						userContinuousSummary = test.RandomContinousSummary(userId)
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

						userContinuousSummary = test.RandomContinousSummary(userId)
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

						userContinuousSummary = test.RandomContinousSummary(userId)
						userContinuousSummary.Dates.OutdatedSince = &fiveMinutesAgo
						userContinuousSummary.Dates.OutdatedReason = []string{types.OutdatedReasonUploadCompleted}
						Expect(userContinuousSummary.Stats.Buckets).ToNot(HaveLen(0))
						Expect(userContinuousSummary.Stats.Periods).ToNot(HaveLen(0))

						err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
						Expect(err).ToNot(HaveOccurred())

						outdatedSince, err = continuousStore.SetOutdated(ctx, userId, types.OutdatedReasonSchemaMigration)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userContinuousSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userContinuousSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(userContinuousSummaryWritten.Stats.Buckets).To(HaveLen(0))
						Expect(userContinuousSummaryWritten.Stats.Periods).To(HaveLen(0))
						Expect(userContinuousSummaryWritten.Dates.LastData).To(BeNil())
						Expect(userContinuousSummaryWritten.Dates.FirstData).To(BeNil())
						Expect(userContinuousSummaryWritten.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
						Expect(userContinuousSummaryWritten.Dates.LastUploadDate).To(BeNil())
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
						var summaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userId),
							test.RandomContinousSummary(userIdOther),
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
						cgmStore := dataStoreSummary.New[*types.CGMStats](summaryRepository)
						bgmStore := dataStoreSummary.New[*types.BGMStats](summaryRepository)

						var cgmSummaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
							test.RandomBGMSummary(userId),
							test.RandomBGMSummary(userIdOther),
						}

						var continuousSummaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userId),
							test.RandomContinousSummary(userIdOther),
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
						Expect(userContinuousSummary).To(BeComparableTo(continuousSummaries[0]))
					})
				})

				Context("DistinctSummaryIDs", func() {
					var userIds []string

					It("With missing context", func() {
						userIds, err = continuousStore.DistinctSummaryIDs(nil)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("context is missing"))
						Expect(len(userIds)).To(Equal(0))
					})

					It("With no summaries", func() {
						userIds, err = continuousStore.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(userIds)).To(Equal(0))
					})

					It("With summaries", func() {
						var continuousSummaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userId),
							test.RandomContinousSummary(userIdOther),
						}

						_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
						Expect(err).ToNot(HaveOccurred())

						userIds, err = continuousStore.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(userIds)).To(Equal(2))
						Expect(userIds).To(ConsistOf([]string{userId, userIdOther}))
					})

					It("With summaries of all types", func() {
						userIdTwo := userTest.RandomID()
						userIdThree := userTest.RandomID()
						userIdFour := userTest.RandomID()
						userIdFive := userTest.RandomID()
						cgmStore := dataStoreSummary.New[*types.CGMStats](summaryRepository)
						bgmStore := dataStoreSummary.New[*types.BGMStats](summaryRepository)

						var cgmSummaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
							test.RandomBGMSummary(userIdTwo),
							test.RandomBGMSummary(userIdThree),
						}

						var continuousSummaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userIdFour),
							test.RandomContinousSummary(userIdFive),
						}

						_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
						Expect(err).ToNot(HaveOccurred())
						_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
						Expect(err).ToNot(HaveOccurred())
						_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
						Expect(err).ToNot(HaveOccurred())

						userIds, err = continuousStore.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(userIds)).To(Equal(2))
						Expect(userIds).To(ConsistOf([]string{userIdFour, userIdFive}))
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
						var continuousSummaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userId),
							test.RandomContinousSummary(userIdOther),
							test.RandomContinousSummary(userIdTwo),
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
						var continuousSummaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userId),
							test.RandomContinousSummary(userIdOther),
							test.RandomContinousSummary(userIdTwo),
							test.RandomContinousSummary(userIdThree),
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
						var continuousSummaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userId),
							test.RandomContinousSummary(userIdOther),
							test.RandomContinousSummary(userIdTwo),
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
						cgmStore := dataStoreSummary.New[*types.CGMStats](summaryRepository)
						bgmStore := dataStoreSummary.New[*types.BGMStats](summaryRepository)

						var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)

						var cgmSummaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
							test.RandomBGMSummary(userIdTwo),
							test.RandomBGMSummary(userIdThree),
						}

						var continuousSummaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userIdFour),
							test.RandomContinousSummary(userIdFive),
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

			Context("CGM", func() {
				var userCGMSummary *types.Summary[*types.CGMStats, types.CGMStats]
				var cgmStore *dataStoreSummary.Repo[*types.CGMStats, types.CGMStats]

				BeforeEach(func() {
					cgmStore = dataStoreSummary.New[*types.CGMStats](summaryRepository)
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
						var userCGMSummaryTwo *types.Summary[*types.CGMStats, types.CGMStats]
						var userCGMSummaryWritten *types.Summary[*types.CGMStats, types.CGMStats]
						var userCGMSummaryWrittenTwo *types.Summary[*types.CGMStats, types.CGMStats]

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
						var userCGMSummaryWritten *types.Summary[*types.CGMStats, types.CGMStats]

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
						var summaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
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
						var summaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						summaries[0].Type = "bgm"

						_, err = cgmStore.CreateSummaries(ctx, summaries)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("invalid summary type 'bgm', expected 'cgm' at index 0"))
					})

					It("Create summaries with an empty userId", func() {
						var summaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
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
						var summaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
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
					var userCGMSummaryWritten *types.Summary[*types.CGMStats, types.CGMStats]

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

						outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonBackfill)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userCGMSummary, err = cgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userCGMSummary.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userCGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(userCGMSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonBackfill}))

						outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userCGMSummary, err = cgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userCGMSummary.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userCGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(userCGMSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonBackfill}))
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
						Expect(userCGMSummary.Stats.Buckets).ToNot(HaveLen(0))
						Expect(userCGMSummary.Stats.Periods).ToNot(HaveLen(0))

						err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
						Expect(err).ToNot(HaveOccurred())

						outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonSchemaMigration)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userCGMSummaryWritten, err = cgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userCGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userCGMSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(userCGMSummaryWritten.Stats.Buckets).To(HaveLen(0))
						Expect(userCGMSummaryWritten.Stats.Periods).To(HaveLen(0))
						Expect(userCGMSummaryWritten.Dates.LastData).To(BeNil())
						Expect(userCGMSummaryWritten.Dates.FirstData).To(BeNil())
						Expect(userCGMSummaryWritten.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
						Expect(userCGMSummaryWritten.Dates.LastUploadDate).To(BeNil())
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
						var summaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
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
						bgmStore := dataStoreSummary.New[*types.BGMStats](summaryRepository)
						continuousStore := dataStoreSummary.New[*types.ContinuousStats](summaryRepository)

						var cgmSummaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
							test.RandomBGMSummary(userId),
							test.RandomBGMSummary(userIdOther),
						}

						var continuousSummaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userId),
							test.RandomContinousSummary(userIdOther),
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

				Context("DistinctSummaryIDs", func() {
					var userIds []string

					It("With missing context", func() {
						userIds, err = cgmStore.DistinctSummaryIDs(nil)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("context is missing"))
						Expect(len(userIds)).To(Equal(0))
					})

					It("With no summaries", func() {
						userIds, err = cgmStore.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(userIds)).To(Equal(0))
					})

					It("With summaries", func() {
						var cgmSummaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
						Expect(err).ToNot(HaveOccurred())

						userIds, err = cgmStore.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(userIds)).To(Equal(2))
						Expect(userIds).To(ConsistOf([]string{userId, userIdOther}))
					})

					It("With summaries of all types", func() {
						userIdTwo := userTest.RandomID()
						userIdThree := userTest.RandomID()
						userIdFour := userTest.RandomID()
						userIdFive := userTest.RandomID()
						continuousStore := dataStoreSummary.New[*types.ContinuousStats](summaryRepository)
						bgmStore := dataStoreSummary.New[*types.BGMStats](summaryRepository)

						var cgmSummaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
							test.RandomBGMSummary(userIdTwo),
							test.RandomBGMSummary(userIdThree),
						}

						var continuousSummaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userIdFour),
							test.RandomContinousSummary(userIdFive),
						}

						_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
						Expect(err).ToNot(HaveOccurred())
						_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
						Expect(err).ToNot(HaveOccurred())
						_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
						Expect(err).ToNot(HaveOccurred())

						userIds, err = cgmStore.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(userIds)).To(Equal(2))
						Expect(userIds).To(ConsistOf([]string{userId, userIdOther}))
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
						var cgmSummaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
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
						var cgmSummaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
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
						var cgmSummaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
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
						userIdFour := userTest.RandomID()
						userIdFive := userTest.RandomID()
						continuousStore := dataStoreSummary.New[*types.ContinuousStats](summaryRepository)
						bgmStore := dataStoreSummary.New[*types.BGMStats](summaryRepository)

						var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)

						var cgmSummaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
							test.RandomBGMSummary(userIdTwo),
							test.RandomBGMSummary(userIdThree),
						}

						var continuousSummaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userIdFour),
							test.RandomContinousSummary(userIdFive),
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

			Context("BGM", func() {
				var bgmStore *dataStoreSummary.Repo[*types.BGMStats, types.BGMStats]
				var userBGMSummary *types.Summary[*types.BGMStats, types.BGMStats]

				BeforeEach(func() {
					bgmStore = dataStoreSummary.New[*types.BGMStats](summaryRepository)
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
						var userBGMSummaryTwo *types.Summary[*types.BGMStats, types.BGMStats]
						var userBGMSummaryWritten *types.Summary[*types.BGMStats, types.BGMStats]
						var userBGMSummaryWrittenTwo *types.Summary[*types.BGMStats, types.BGMStats]

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
						var userBGMSummaryWritten *types.Summary[*types.BGMStats, types.BGMStats]

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
						var summaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
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
						var summaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
							test.RandomBGMSummary(userId),
							test.RandomBGMSummary(userIdOther),
						}

						summaries[0].Type = "cgm"

						_, err = bgmStore.CreateSummaries(ctx, summaries)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("invalid summary type 'cgm', expected 'bgm' at index 0"))
					})

					It("Create summaries with an invalid type", func() {
						var summaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
							test.RandomBGMSummary(userId),
							test.RandomBGMSummary(userIdOther),
						}

						summaries[0].Type = "cgm"

						_, err = bgmStore.CreateSummaries(ctx, summaries)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("invalid summary type 'cgm', expected 'bgm' at index 0"))
					})

					It("Create summaries with an empty userId", func() {
						var summaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
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
						var summaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
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
					var userBGMSummaryWritten *types.Summary[*types.BGMStats, types.BGMStats]

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

						outdatedSince, err = bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonBackfill)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userBGMSummary, err = bgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userBGMSummary.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userBGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(userBGMSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonBackfill}))

						outdatedSince, err = bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userBGMSummary, err = bgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userBGMSummary.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userBGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(userBGMSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonBackfill}))
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
						Expect(userBGMSummary.Stats.Buckets).ToNot(HaveLen(0))
						Expect(userBGMSummary.Stats.Periods).ToNot(HaveLen(0))

						err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
						Expect(err).ToNot(HaveOccurred())

						outdatedSince, err = bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonSchemaMigration)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userBGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userBGMSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(userBGMSummaryWritten.Stats.Buckets).To(HaveLen(0))
						Expect(userBGMSummaryWritten.Stats.Periods).To(HaveLen(0))
						Expect(userBGMSummaryWritten.Dates.LastData).To(BeNil())
						Expect(userBGMSummaryWritten.Dates.FirstData).To(BeNil())
						Expect(userBGMSummaryWritten.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
						Expect(userBGMSummaryWritten.Dates.LastUploadDate).To(BeNil())
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
						var summaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
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

					It("Get with multiple summaries of different type", func() {
						cgmStore := dataStoreSummary.New[*types.CGMStats](summaryRepository)
						continuousStore := dataStoreSummary.New[*types.ContinuousStats](summaryRepository)

						var cgmSummaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
							test.RandomBGMSummary(userId),
							test.RandomBGMSummary(userIdOther),
						}

						var continuousSummaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userId),
							test.RandomContinousSummary(userIdOther),
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

				Context("DistinctSummaryIDs", func() {
					var userIds []string

					It("With missing context", func() {
						userIds, err = bgmStore.DistinctSummaryIDs(nil)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("context is missing"))
						Expect(len(userIds)).To(Equal(0))
					})

					It("With no summaries", func() {
						userIds, err = bgmStore.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(userIds)).To(Equal(0))
					})

					It("With summaries", func() {
						var cgmSummaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
							test.RandomBGMSummary(userId),
							test.RandomBGMSummary(userIdOther),
						}

						_, err = bgmStore.CreateSummaries(ctx, cgmSummaries)
						Expect(err).ToNot(HaveOccurred())

						userIds, err = bgmStore.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(userIds)).To(Equal(2))
						Expect(userIds).To(ConsistOf([]string{userId, userIdOther}))
					})

					It("With summaries of all types", func() {
						userIdTwo := userTest.RandomID()
						userIdThree := userTest.RandomID()
						userIdFour := userTest.RandomID()
						userIdFive := userTest.RandomID()
						continuousStore := dataStoreSummary.New[*types.ContinuousStats](summaryRepository)
						cgmStore := dataStoreSummary.New[*types.CGMStats](summaryRepository)

						var cgmSummaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
							test.RandomBGMSummary(userIdTwo),
							test.RandomBGMSummary(userIdThree),
						}

						var continuousSummaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userIdFour),
							test.RandomContinousSummary(userIdFive),
						}

						_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
						Expect(err).ToNot(HaveOccurred())
						_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
						Expect(err).ToNot(HaveOccurred())
						_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
						Expect(err).ToNot(HaveOccurred())

						userIds, err = bgmStore.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(userIds)).To(Equal(2))
						Expect(userIds).To(ConsistOf([]string{userIdTwo, userIdThree}))
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
						var bgmSummaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
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
						var bgmSummaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
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
						var bgmSummaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
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
						continuousStore := dataStoreSummary.New[*types.ContinuousStats](summaryRepository)
						cgmStore := dataStoreSummary.New[*types.CGMStats](summaryRepository)

						var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)

						var cgmSummaries = []*types.Summary[*types.CGMStats, types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[*types.BGMStats, types.BGMStats]{
							test.RandomBGMSummary(userIdTwo),
							test.RandomBGMSummary(userIdThree),
						}

						var continuousSummaries = []*types.Summary[*types.ContinuousStats, types.ContinuousStats]{
							test.RandomContinousSummary(userIdFour),
							test.RandomContinousSummary(userIdFive),
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
			})

			Context("Typeless", func() {
				var userBGMSummary *types.Summary[*types.BGMStats, types.BGMStats]
				var userCGMSummary *types.Summary[*types.CGMStats, types.CGMStats]
				var userContinuousSummary *types.Summary[*types.ContinuousStats, types.ContinuousStats]
				var bgmStore *dataStoreSummary.Repo[*types.BGMStats, types.BGMStats]
				var cgmStore *dataStoreSummary.Repo[*types.CGMStats, types.CGMStats]
				var continuousStore *dataStoreSummary.Repo[*types.ContinuousStats, types.ContinuousStats]

				BeforeEach(func() {
					bgmStore = dataStoreSummary.New[*types.BGMStats](summaryRepository)
					cgmStore = dataStoreSummary.New[*types.CGMStats](summaryRepository)
					continuousStore = dataStoreSummary.New[*types.ContinuousStats](summaryRepository)
				})

				Context("DeleteSummary", func() {
					It("Delete All Summaries for User", func() {
						var userCGMSummaryWritten *types.Summary[*types.CGMStats, types.CGMStats]
						var userBGMSummaryWritten *types.Summary[*types.BGMStats, types.BGMStats]
						var userContinuousSummaryWritten *types.Summary[*types.ContinuousStats, types.ContinuousStats]

						userCGMSummary = test.RandomCGMSummary(userId)
						Expect(userCGMSummary.Type).To(Equal("cgm"))

						err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
						Expect(err).ToNot(HaveOccurred())

						userBGMSummary = test.RandomBGMSummary(userId)
						Expect(userBGMSummary.Type).To(Equal("bgm"))

						err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
						Expect(err).ToNot(HaveOccurred())

						userContinuousSummary = test.RandomContinousSummary(userId)
						Expect(userContinuousSummary.Type).To(Equal("continuous"))

						err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
						Expect(err).ToNot(HaveOccurred())

						// confirm writes
						userCGMSummaryWritten, err = cgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userCGMSummaryWritten).ToNot(BeNil())

						userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userBGMSummaryWritten).ToNot(BeNil())

						userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userContinuousSummaryWritten).ToNot(BeNil())

						// delete
						err = typelessStore.DeleteSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())

						// confirm delete
						userCGMSummaryWritten, err = cgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userCGMSummaryWritten).To(BeNil())

						userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userBGMSummaryWritten).To(BeNil())

						userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userContinuousSummaryWritten).To(BeNil())
					})
				})
			})

		})
	})
})
