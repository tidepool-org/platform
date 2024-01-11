package store_test

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"

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
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Summary Stats Mongo", func() {
	var logger *logTest.Logger
	var err error
	var ctx context.Context
	var config *storeStructuredMongo.Config
	var store *dataStoreMongo.Store
	var summaryRepository *storeStructuredMongo.Repository

	BeforeEach(func() {
		logger = logTest.NewLogger()
		config = storeStructuredMongoTest.NewConfig()
	})

	AfterEach(func() {
		if store != nil {
			_ = store.Terminate(context.Background())
		}
	})

	Context("Create Stores", func() {
		It("CGM Repo", func() {
			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())

			summaryRepository = store.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			cgmStore := dataStoreSummary.New[types.CGMStats, *types.CGMStats](summaryRepository)
			Expect(cgmStore).ToNot(BeNil())
		})

		It("BGM Repo", func() {
			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())

			summaryRepository = store.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			bgmStore := dataStoreSummary.New[types.BGMStats, *types.BGMStats](summaryRepository)
			Expect(bgmStore).ToNot(BeNil())
		})

		It("Typeless Repo", func() {
			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())

			summaryRepository = store.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			typelessStore := dataStoreSummary.NewTypeless(summaryRepository)
			Expect(typelessStore).ToNot(BeNil())
		})
	})

	Context("With a new store", func() {
		var summaryCollection *mongo.Collection

		BeforeEach(func() {
			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())

			summaryCollection = store.GetCollection("summary")
			Expect(store.EnsureIndexes()).To(Succeed())
		})

		AfterEach(func() {
			if summaryCollection != nil {
				_ = summaryCollection.Database().Drop(context.Background())
			}
		})

		Context("With a repository", func() {
			BeforeEach(func() {
				summaryRepository = store.NewSummaryRepository().GetStore()
				Expect(summaryRepository).ToNot(BeNil())
			})

			AfterEach(func() {
				_, err = summaryCollection.DeleteMany(ctx, bson.D{})
				Expect(err).ToNot(HaveOccurred())
			})

			Context("With typed Stores", func() {
				var userId string
				var userIdOther string
				var cgmStore *dataStoreSummary.Repo[types.CGMStats, *types.CGMStats]
				var bgmStore *dataStoreSummary.Repo[types.BGMStats, *types.BGMStats]
				var typelessStore *dataStoreSummary.TypelessRepo

				var userCGMSummary *types.Summary[types.CGMStats, *types.CGMStats]
				var userBGMSummary *types.Summary[types.BGMStats, *types.BGMStats]

				BeforeEach(func() {
					ctx = log.NewContextWithLogger(context.Background(), logger)
					userId = userTest.RandomID()
					userIdOther = userTest.RandomID()

					cgmStore = dataStoreSummary.New[types.CGMStats, *types.CGMStats](summaryRepository)
					bgmStore = dataStoreSummary.New[types.BGMStats, *types.BGMStats](summaryRepository)
					typelessStore = dataStoreSummary.NewTypeless(summaryRepository)
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

					It("Insert CGM Summary with missing Type", func() {
						userCGMSummary = test.RandomCGMSummary(userId)
						userCGMSummary.Type = ""

						err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("invalid summary type '', expected 'cgm'"))
					})

					It("Insert CGM Summary with invalid Type", func() {
						userCGMSummary = test.RandomCGMSummary(userId)
						userCGMSummary.Type = "bgm"

						err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("invalid summary type 'bgm', expected 'cgm'"))
					})

					It("Insert BGM Summary with missing Type", func() {
						userBGMSummary = test.RandomBGMSummary(userId)
						userBGMSummary.Type = ""

						err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("invalid summary type '', expected 'bgm'"))
					})

					It("Insert BGM Summary with invalid Type", func() {
						userBGMSummary = test.RandomBGMSummary(userId)
						userBGMSummary.Type = "asdf"

						err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("invalid summary type 'asdf', expected 'bgm'"))
					})

					It("Insert CGM Summary", func() {
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

					It("Insert BGM Summary", func() {
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

					It("Update CGM Summary", func() {
						var userCGMSummaryTwo *types.Summary[types.CGMStats, *types.CGMStats]
						var userCGMSummaryWritten *types.Summary[types.CGMStats, *types.CGMStats]
						var userCGMSummaryWrittenTwo *types.Summary[types.CGMStats, *types.CGMStats]

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

					It("Update BGM Summary", func() {
						var userBGMSummaryTwo *types.Summary[types.BGMStats, *types.BGMStats]
						var userBGMSummaryWritten *types.Summary[types.BGMStats, *types.BGMStats]
						var userBGMSummaryWrittenTwo *types.Summary[types.BGMStats, *types.BGMStats]

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
						err = cgmStore.DeleteSummary(nil, userId)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("context is missing"))
					})

					It("Delete Summary with empty userId", func() {
						err = cgmStore.DeleteSummary(ctx, "")
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("userId is missing"))
					})

					It("Delete CGM Summary", func() {
						var userCGMSummaryWritten *types.Summary[types.CGMStats, *types.CGMStats]

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

					It("Delete BGM Summary", func() {
						var userBGMSummaryWritten *types.Summary[types.BGMStats, *types.BGMStats]

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

					It("Delete All Summaries for User", func() {
						var userCGMSummaryWritten *types.Summary[types.CGMStats, *types.CGMStats]
						var userBGMSummaryWritten *types.Summary[types.BGMStats, *types.BGMStats]

						userCGMSummary = test.RandomCGMSummary(userId)
						Expect(userCGMSummary.Type).To(Equal("cgm"))

						err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
						Expect(err).ToNot(HaveOccurred())

						userBGMSummary = test.RandomBGMSummary(userId)
						Expect(userBGMSummary.Type).To(Equal("bgm"))

						err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
						Expect(err).ToNot(HaveOccurred())

						// confirm writes
						userCGMSummaryWritten, err = cgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userCGMSummaryWritten).ToNot(BeNil())

						userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userBGMSummaryWritten).ToNot(BeNil())

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
					})

				})

				Context("CreateSummaries", func() {

					It("Create summaries with missing context", func() {
						var summaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
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

					It("Create CGM summaries with an invalid type", func() {
						var summaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						summaries[0].Type = "bgm"

						_, err = cgmStore.CreateSummaries(ctx, summaries)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("invalid summary type 'bgm', expected 'cgm' at index 0"))
					})

					It("Create BGM summaries with an invalid type", func() {
						var summaries = []*types.Summary[types.BGMStats, *types.BGMStats]{
							test.RandomBGMSummary(userId),
							test.RandomBGMSummary(userIdOther),
						}

						summaries[0].Type = "cgm"

						_, err = bgmStore.CreateSummaries(ctx, summaries)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("invalid summary type 'cgm', expected 'bgm' at index 0"))
					})

					It("Create summaries with an empty userId", func() {
						var summaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						summaries[0].UserID = ""

						_, err = cgmStore.CreateSummaries(ctx, summaries)
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("userId is missing at index 0"))
					})

					It("Create CGM summaries", func() {
						var count int
						var summaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
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

					It("Create BGM summaries", func() {
						var count int
						var summaries = []*types.Summary[types.BGMStats, *types.BGMStats]{
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
						Expect(*userCGMSummary.Dates.OutdatedSinceLimit).To(Equal(outdatedSinceOriginal.Add(30 * time.Minute)))
						Expect(userCGMSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded}))

						outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonBackfill)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userCGMSummary, err = cgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userCGMSummary.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userCGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(*userCGMSummary.Dates.OutdatedSinceLimit).To(Equal(outdatedSinceOriginal.Add(30 * time.Minute)))
						Expect(userCGMSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonBackfill}))

						outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userCGMSummary, err = cgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userCGMSummary.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userCGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(*userCGMSummary.Dates.OutdatedSinceLimit).To(Equal(outdatedSinceOriginal.Add(30 * time.Minute)))
						Expect(userCGMSummary.Dates.OutdatedReason).To(ConsistOf([]string{types.OutdatedReasonDataAdded, types.OutdatedReasonBackfill}))
					})

					It("With no existing CGM summary", func() {
						outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userCGMSummary, err = cgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userCGMSummary.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userCGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(*userCGMSummary.Dates.OutdatedSinceLimit).To(Equal(outdatedSince.Add(30 * time.Minute)))
					})

					It("With an existing non-outdated CGM summary", func() {
						var userCGMSummaryWritten *types.Summary[types.CGMStats, *types.CGMStats]

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
						Expect(*userCGMSummaryWritten.Dates.OutdatedSinceLimit).To(Equal(outdatedSince.Add(30 * time.Minute)))

					})

					It("With an existing outdated CGM summary", func() {
						var userCGMSummaryWritten *types.Summary[types.CGMStats, *types.CGMStats]
						var fiveMinutesAgo = time.Now().Add(time.Duration(-5) * time.Minute).UTC().Truncate(time.Millisecond)

						userCGMSummary = test.RandomCGMSummary(userId)
						userCGMSummary.Dates.OutdatedSince = &fiveMinutesAgo
						userCGMSummary.Dates.OutdatedSinceLimit = pointer.FromAny(fiveMinutesAgo.Add(28 * time.Minute))
						err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
						Expect(err).ToNot(HaveOccurred())

						outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userCGMSummaryWritten, err = cgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userCGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userCGMSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(*userCGMSummaryWritten.Dates.OutdatedSinceLimit).To(Equal(fiveMinutesAgo.Add(28 * time.Minute)))
					})

					It("With an existing outdated CGM summary beyond the outdatedSinceLimit", func() {
						var userCGMSummaryWritten *types.Summary[types.CGMStats, *types.CGMStats]
						now := time.Now().UTC().Truncate(time.Millisecond)

						userCGMSummary = test.RandomCGMSummary(userId)
						userCGMSummary.Dates.OutdatedSince = &now
						userCGMSummary.Dates.OutdatedSinceLimit = &now
						err = cgmStore.ReplaceSummary(ctx, userCGMSummary)
						Expect(err).ToNot(HaveOccurred())

						outdatedSince, err = cgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userCGMSummaryWritten, err = cgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userCGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
						Expect(*userCGMSummaryWritten.Dates.OutdatedSinceLimit).To(Equal(now))
					})

					It("With an existing outdated CGM summary with schema migration reason", func() {
						var userCGMSummaryWritten *types.Summary[types.CGMStats, *types.CGMStats]
						now := time.Now().UTC().Truncate(time.Millisecond)
						fiveMinutesAgo := now.Add(time.Duration(-5) * time.Minute)

						userCGMSummary = test.RandomCGMSummary(userId)
						userCGMSummary.Dates.OutdatedSince = &fiveMinutesAgo
						userCGMSummary.Dates.OutdatedReason = []string{types.OutdatedReasonUploadCompleted}
						userCGMSummary.Dates.OutdatedSinceLimit = pointer.FromAny(fiveMinutesAgo.Add(28 * time.Minute))
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
						Expect(*userCGMSummaryWritten.Dates.OutdatedSinceLimit).To(Equal(outdatedSince.Add(30 * time.Minute)))
						Expect(userCGMSummaryWritten.Stats.Buckets).To(HaveLen(0))
						Expect(userCGMSummaryWritten.Stats.Periods).To(HaveLen(0))
						Expect(userCGMSummaryWritten.Dates.LastData).To(BeNil())
						Expect(userCGMSummaryWritten.Dates.FirstData).To(BeNil())
						Expect(userCGMSummaryWritten.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
						Expect(userCGMSummaryWritten.Dates.LastUploadDate).To(BeNil())
					})

					It("With no existing BGM summary", func() {
						outdatedSince, err = bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userBGMSummary, err = bgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userBGMSummary.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userBGMSummary.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(*userBGMSummary.Dates.OutdatedSinceLimit).To(Equal(outdatedSince.Add(30 * time.Minute)))
					})

					It("With an existing non-outdated BGM summary", func() {
						var userBGMSummaryWritten *types.Summary[types.BGMStats, *types.BGMStats]

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
						Expect(*userBGMSummaryWritten.Dates.OutdatedSinceLimit).To(Equal(outdatedSince.Add(30 * time.Minute)))

					})

					It("With an existing outdated BGM summary", func() {
						var userBGMSummaryWritten *types.Summary[types.BGMStats, *types.BGMStats]
						var fiveMinutesAgo = time.Now().Add(time.Duration(-5) * time.Minute).UTC().Truncate(time.Millisecond)

						userBGMSummary = test.RandomBGMSummary(userId)
						userBGMSummary.Dates.OutdatedSince = &fiveMinutesAgo
						userBGMSummary.Dates.OutdatedSinceLimit = pointer.FromAny(fiveMinutesAgo.Add(30 * time.Minute))
						err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
						Expect(err).ToNot(HaveOccurred())

						outdatedSince, err = bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userBGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
						Expect(userBGMSummaryWritten.Dates.OutdatedSince).To(Equal(outdatedSince))
						Expect(*userBGMSummaryWritten.Dates.OutdatedSinceLimit).To(Equal(fiveMinutesAgo.Add(30 * time.Minute)))

					})

					It("With an existing outdated BGM summary beyond the outdatedSinceLimit", func() {
						var userBGMSummaryWritten *types.Summary[types.BGMStats, *types.BGMStats]
						now := time.Now().UTC().Truncate(time.Millisecond)

						userBGMSummary = test.RandomBGMSummary(userId)
						userBGMSummary.Dates.OutdatedSince = &now
						userBGMSummary.Dates.OutdatedSinceLimit = &now
						err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
						Expect(err).ToNot(HaveOccurred())

						outdatedSince, err = bgmStore.SetOutdated(ctx, userId, types.OutdatedReasonDataAdded)
						Expect(err).ToNot(HaveOccurred())
						Expect(outdatedSince).ToNot(BeNil())

						userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userBGMSummaryWritten.Dates.OutdatedSince).ToNot(BeNil())
						Expect(*userBGMSummaryWritten.Dates.OutdatedSince).To(Equal(now))
						Expect(*userBGMSummaryWritten.Dates.OutdatedSinceLimit).To(Equal(now))
					})

					It("With an existing outdated BGM summary with schema migration reason", func() {
						var userBGMSummaryWritten *types.Summary[types.BGMStats, *types.BGMStats]
						now := time.Now().UTC().Truncate(time.Millisecond)
						fiveMinutesAgo := now.Add(time.Duration(-5) * time.Minute)

						userBGMSummary = test.RandomBGMSummary(userId)
						userBGMSummary.Dates.OutdatedSince = &fiveMinutesAgo
						userBGMSummary.Dates.OutdatedReason = []string{types.OutdatedReasonUploadCompleted}
						userBGMSummary.Dates.OutdatedSinceLimit = pointer.FromAny(fiveMinutesAgo.Add(30 * time.Minute))
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
						Expect(*userBGMSummaryWritten.Dates.OutdatedSinceLimit).To(Equal(outdatedSince.Add(30 * time.Minute)))
						Expect(userBGMSummaryWritten.Stats.Buckets).To(HaveLen(0))
						Expect(userBGMSummaryWritten.Stats.Periods).To(HaveLen(0))
						Expect(userBGMSummaryWritten.Dates.LastData).To(BeNil())
						Expect(userBGMSummaryWritten.Dates.FirstData).To(BeNil())
						Expect(userBGMSummaryWritten.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
						Expect(userBGMSummaryWritten.Dates.LastUploadDate).To(BeNil())
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

					It("With multiple CGM summaries", func() {
						var summaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
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

					It("With multiple BGM summaries", func() {
						var summaries = []*types.Summary[types.BGMStats, *types.BGMStats]{
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

					It("Get CGM with multiple summaries of different type", func() {
						var cgmSummaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[types.BGMStats, *types.BGMStats]{
							test.RandomBGMSummary(userId),
							test.RandomBGMSummary(userIdOther),
						}

						_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
						Expect(err).ToNot(HaveOccurred())

						_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
						Expect(err).ToNot(HaveOccurred())

						userCGMSummary, err = cgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())
						Expect(userCGMSummary).ToNot(BeNil())

						cgmSummaries[0].ID = userCGMSummary.ID
						Expect(userCGMSummary).To(Equal(cgmSummaries[0]))
					})

					It("Get BGM with multiple summaries of different type", func() {
						var cgmSummaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[types.BGMStats, *types.BGMStats]{
							test.RandomBGMSummary(userId),
							test.RandomBGMSummary(userIdOther),
						}

						_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
						Expect(err).ToNot(HaveOccurred())

						_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
						Expect(err).ToNot(HaveOccurred())

						userBGMSummary, err = bgmStore.GetSummary(ctx, userIdOther)
						Expect(err).ToNot(HaveOccurred())
						Expect(userBGMSummary).ToNot(BeNil())

						bgmSummaries[1].ID = userBGMSummary.ID
						Expect(userBGMSummary).To(Equal(bgmSummaries[1]))
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

					It("With CGM summaries", func() {
						var cgmSummaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
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

					It("With BGM summaries", func() {
						var bgmSummaries = []*types.Summary[types.BGMStats, *types.BGMStats]{
							test.RandomBGMSummary(userId),
							test.RandomBGMSummary(userIdOther),
						}

						_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
						Expect(err).ToNot(HaveOccurred())

						userIds, err = bgmStore.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(userIds)).To(Equal(2))
						Expect(userIds).To(ConsistOf([]string{userId, userIdOther}))
					})

					It("Get CGM with summaries of both types", func() {
						userIdTwo := userTest.RandomID()
						userIdThree := userTest.RandomID()
						var cgmSummaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[types.BGMStats, *types.BGMStats]{
							test.RandomBGMSummary(userIdTwo),
							test.RandomBGMSummary(userIdThree),
						}

						_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
						Expect(err).ToNot(HaveOccurred())
						_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
						Expect(err).ToNot(HaveOccurred())

						userIds, err = cgmStore.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(userIds)).To(Equal(2))
						Expect(userIds).To(ConsistOf([]string{userId, userIdOther}))
					})

					It("Get BGM with summaries of both types", func() {
						userIdTwo := userTest.RandomID()
						userIdThree := userTest.RandomID()
						var cgmSummaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[types.BGMStats, *types.BGMStats]{
							test.RandomBGMSummary(userIdTwo),
							test.RandomBGMSummary(userIdThree),
						}

						_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
						Expect(err).ToNot(HaveOccurred())
						_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
						Expect(err).ToNot(HaveOccurred())

						userIds, err = bgmStore.DistinctSummaryIDs(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(userIds)).To(Equal(2))
						Expect(userIds).To(ConsistOf([]string{userIdTwo, userIdThree}))
					})

				})

				Context("GetOutdatedUserIDs", func() {
					var userIds []string
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

					It("With outdated CGM summaries", func() {
						var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
						var cgmSummaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
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
						Expect(userIds).To(ConsistOf([]string{userId, userIdTwo}))
					})

					It("With outdated BGM summaries", func() {
						var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
						var bgmSummaries = []*types.Summary[types.BGMStats, *types.BGMStats]{
							test.RandomBGMSummary(userId),
							test.RandomBGMSummary(userIdOther),
							test.RandomBGMSummary(userIdTwo),
						}

						// mark 2/3 summaries outdated
						bgmSummaries[0].Dates.OutdatedSince = nil
						bgmSummaries[1].Dates.OutdatedSince = &outdatedTime
						bgmSummaries[2].Dates.OutdatedSince = &outdatedTime
						_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
						Expect(err).ToNot(HaveOccurred())

						userIds, err = bgmStore.GetOutdatedUserIDs(ctx, page.NewPagination())
						Expect(err).ToNot(HaveOccurred())
						Expect(userIds).To(ConsistOf([]string{userIdOther, userIdTwo}))
					})

					It("Get outdated CGM summaries with both types present", func() {
						var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
						var cgmSummaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[types.BGMStats, *types.BGMStats]{
							test.RandomBGMSummary(userIdTwo),
							test.RandomBGMSummary(userIdThree),
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

						userIds, err = cgmStore.GetOutdatedUserIDs(ctx, page.NewPagination())
						Expect(err).ToNot(HaveOccurred())
						Expect(userIds).To(ConsistOf([]string{userId}))
					})

					It("Get outdated BGM summaries with both types present", func() {
						var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
						var cgmSummaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
						}

						var bgmSummaries = []*types.Summary[types.BGMStats, *types.BGMStats]{
							test.RandomBGMSummary(userIdTwo),
							test.RandomBGMSummary(userIdThree),
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

						userIds, err = bgmStore.GetOutdatedUserIDs(ctx, page.NewPagination())
						Expect(err).ToNot(HaveOccurred())
						Expect(userIds).To(ConsistOf([]string{userIdThree}))
					})

					It("With a specific pagination size", func() {
						var pagination = page.NewPagination()
						var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
						var cgmSummaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
							test.RandomCGMSummary(userId),
							test.RandomCGMSummary(userIdOther),
							test.RandomCGMSummary(userIdTwo),
							test.RandomCGMSummary(userIdThree),
						}

						pagination.Size = 3

						for i := 0; i < len(cgmSummaries); i++ {
							cgmSummaries[i].Dates.OutdatedSince = &outdatedTime
						}
						_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
						Expect(err).ToNot(HaveOccurred())

						userIds, err = cgmStore.GetOutdatedUserIDs(ctx, pagination)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(userIds)).To(Equal(3))
						Expect(userIds).To(ConsistOf([]string{userId, userIdOther, userIdTwo}))
					})

					It("Check sort order", func() {
						var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
						var cgmSummaries = []*types.Summary[types.CGMStats, *types.CGMStats]{
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
						Expect(len(userIds)).To(Equal(3))

						// we expect these to come back in reverse order than inserted
						for i := 0; i < len(userIds); i++ {
							Expect(userIds[i]).To(Equal(cgmSummaries[len(cgmSummaries)-i-1].UserID))
						}
					})

				})

			})
		})
	})
})
