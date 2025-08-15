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
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	dataStoreSummary "github.com/tidepool-org/platform/summary/store"
	"github.com/tidepool-org/platform/summary/test"
	"github.com/tidepool-org/platform/summary/types"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Summary Periods Mongo", Label("mongodb", "slow", "integration"), func() {
	var logger *logTest.Logger
	var ctx context.Context

	var summaryRepository *storeStructuredMongo.Repository

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
	})

	Context("Create Stores", func() {
		var store *dataStoreMongo.Store

		It("Typeless Repo", func() {
			store = GetSuiteStore()
			summaryRepository = store.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			typelessStore := dataStoreSummary.NewTypeless(summaryRepository)
			Expect(typelessStore).ToNot(BeNil())
		})
	})

	Context("With a new store", func() {
		var userId string
		var typelessStore *dataStoreSummary.TypelessSummaries
		var store *dataStoreMongo.Store
		var err error

		BeforeEach(func() {
			config := storeStructuredMongoTest.NewConfig()
			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			summaryRepository = store.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			userId = userTest.RandomUserID()
			typelessStore = dataStoreSummary.NewTypeless(summaryRepository)
		})

		AfterEach(func() {
			if summaryRepository != nil {
				_, err = summaryRepository.DeleteMany(ctx, bson.D{})
				Expect(err).To(Succeed())
			}
		})

		Context("Typeless", func() {
			var userBGMSummary *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]
			var userCGMSummary *types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]
			var userContinuousSummary *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]
			var bgmStore *dataStoreSummary.Summaries[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]
			var cgmStore *dataStoreSummary.Summaries[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]
			var continuousStore *dataStoreSummary.Summaries[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]

			BeforeEach(func() {
				bgmStore = dataStoreSummary.NewSummaries[*types.BGMPeriods, *types.GlucoseBucket](summaryRepository)
				cgmStore = dataStoreSummary.NewSummaries[*types.CGMPeriods, *types.GlucoseBucket](summaryRepository)
				continuousStore = dataStoreSummary.NewSummaries[*types.ContinuousPeriods, *types.ContinuousBucket](summaryRepository)
			})

			Context("DeleteSummary", func() {
				It("Delete All Summaries for User", func() {
					var userCGMSummaryWritten *types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]
					var userBGMSummaryWritten *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]
					var userContinuousSummaryWritten *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]

					userCGMSummary = test.RandomCGMSummary(userId)
					Expect(userCGMSummary.Type).To(Equal("cgm"))

					err := cgmStore.ReplaceSummary(ctx, userCGMSummary)
					Expect(err).ToNot(HaveOccurred())

					userBGMSummary = test.RandomBGMSummary(userId)
					Expect(userBGMSummary.Type).To(Equal("bgm"))

					err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
					Expect(err).ToNot(HaveOccurred())

					userContinuousSummary = test.RandomContinuousSummary(userId)
					Expect(userContinuousSummary.Type).To(Equal("con"))

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

			Context("GetMigratableUserIDs", func() {
				var userIds []string
				var userIdTwo string
				var userIdThree string
				var userIdOther string

				BeforeEach(func() {
					userIdTwo = userTest.RandomUserID()
					userIdThree = userTest.RandomUserID()
					userIdOther = userTest.RandomUserID()
				})

				It("With missing context", func() {
					userIds, err = continuousStore.GetMigratableUserIDs(nil, page.NewPagination())
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("context is missing"))
					Expect(userIds).To(BeNil())
				})

				It("With missing pagination", func() {
					userIds, err = continuousStore.GetMigratableUserIDs(ctx, nil)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("pagination is missing"))
					Expect(userIds).To(BeNil())
				})

				It("With no migratable summaries", func() {
					var pagination = page.NewPagination()

					userIds, err = continuousStore.GetMigratableUserIDs(ctx, pagination)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(userIds)).To(Equal(0))
				})

				It("With migratable CGM summaries", func() {
					var continuousSummaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
						test.RandomContinuousSummary(userId),
						test.RandomContinuousSummary(userIdOther),
						test.RandomContinuousSummary(userIdTwo),
					}

					// mark 2/3 summaries for migration
					continuousSummaries[0].Config.SchemaVersion = types.SchemaVersion - 1
					continuousSummaries[0].Dates.OutdatedSince = nil
					continuousSummaries[1].Config.SchemaVersion = types.SchemaVersion
					continuousSummaries[1].Dates.OutdatedSince = nil
					continuousSummaries[2].Config.SchemaVersion = types.SchemaVersion - 1
					continuousSummaries[2].Dates.OutdatedSince = nil
					_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
					Expect(err).ToNot(HaveOccurred())

					userIds, err = continuousStore.GetMigratableUserIDs(ctx, page.NewPagination())
					Expect(err).ToNot(HaveOccurred())
					Expect(userIds).To(ConsistOf([]string{userId, userIdTwo}))
				})

				It("With migratable and outdated CGM summaries", func() {
					var outdatedTime = time.Now().UTC().Truncate(time.Millisecond)
					var continuousSummaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
						test.RandomContinuousSummary(userId),
						test.RandomContinuousSummary(userIdOther),
						test.RandomContinuousSummary(userIdTwo),
					}

					// mark 2/3 summaries for migration, and 1/3 as outdated
					continuousSummaries[0].Config.SchemaVersion = types.SchemaVersion - 1
					continuousSummaries[0].Dates.OutdatedSince = nil
					continuousSummaries[1].Config.SchemaVersion = types.SchemaVersion - 1
					continuousSummaries[1].Dates.OutdatedSince = &outdatedTime
					continuousSummaries[2].Config.SchemaVersion = types.SchemaVersion - 1
					continuousSummaries[2].Dates.OutdatedSince = nil
					_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
					Expect(err).ToNot(HaveOccurred())

					userIds, err = continuousStore.GetMigratableUserIDs(ctx, page.NewPagination())
					Expect(err).ToNot(HaveOccurred())
					Expect(userIds).To(ConsistOf([]string{userId, userIdTwo}))
				})

				It("With a specific pagination size", func() {
					var lastUpdatedTime = time.Now().UTC().Truncate(time.Millisecond)
					var pagination = page.NewPagination()
					var continuousSummaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
						test.RandomContinuousSummary(userId),
						test.RandomContinuousSummary(userIdOther),
						test.RandomContinuousSummary(userIdTwo),
						test.RandomContinuousSummary(userIdThree),
					}

					pagination.Size = 3

					for i := len(continuousSummaries) - 1; i >= 0; i-- {
						continuousSummaries[i].Config.SchemaVersion = types.SchemaVersion - 1
						continuousSummaries[i].Dates.OutdatedSince = nil
						continuousSummaries[i].Dates.LastUpdatedDate = lastUpdatedTime.Add(time.Duration(-i) * time.Minute)
					}
					_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
					Expect(err).ToNot(HaveOccurred())

					userIds, err = continuousStore.GetMigratableUserIDs(ctx, pagination)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(userIds)).To(Equal(3))
					Expect(userIds).To(ConsistOf([]string{userIdThree, userIdTwo, userIdOther}))
				})

				It("Check sort order", func() {
					var lastUpdatedTime = time.Now().UTC().Truncate(time.Millisecond)
					var continuousSummaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
						test.RandomContinuousSummary(userId),
						test.RandomContinuousSummary(userIdOther),
						test.RandomContinuousSummary(userIdTwo),
					}

					for i := 0; i < len(continuousSummaries); i++ {
						continuousSummaries[i].Config.SchemaVersion = types.SchemaVersion - 1
						continuousSummaries[i].Dates.OutdatedSince = nil
						continuousSummaries[i].Dates.LastUpdatedDate = lastUpdatedTime.Add(time.Duration(-i) * time.Minute)
					}
					_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
					Expect(err).ToNot(HaveOccurred())

					userIds, err = continuousStore.GetMigratableUserIDs(ctx, page.NewPagination())
					Expect(err).ToNot(HaveOccurred())
					Expect(len(userIds)).To(Equal(3))

					// we expect these to come back in reverse order than inserted
					for i := 0; i < len(userIds); i++ {
						Expect(userIds[i]).To(Equal(continuousSummaries[len(continuousSummaries)-i-1].UserID))
					}
				})

				It("Get migratable summaries with all types present", func() {
					userIdFour := userTest.RandomUserID()
					userIdFive := userTest.RandomUserID()
					bgmStore := dataStoreSummary.NewSummaries[*types.BGMPeriods, *types.GlucoseBucket](summaryRepository)
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

					// mark 1 for migration per type
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

					userIds, err = continuousStore.GetMigratableUserIDs(ctx, page.NewPagination())
					Expect(err).ToNot(HaveOccurred())
					Expect(userIds).To(ConsistOf([]string{userIdFive}))
				})
			})
		})

	})
})
