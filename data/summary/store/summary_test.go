package store_test

import (
	"context"

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
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
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
		var err error

		AfterEach(func() {
			if store != nil {
				_ = store.Terminate(context.Background())
			}
		})

		It("Typeless Repo", func() {
			config := storeStructuredMongoTest.NewConfig()
			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			summaryRepository = store.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			typelessStore := dataStoreSummary.NewTypeless(summaryRepository)
			Expect(typelessStore).ToNot(BeNil())
		})
	})

	Context("With a new store", func() {
		var summaryCollection *mongo.Collection
		var userId string
		var typelessStore *dataStoreSummary.TypelessSummaries
		var store *dataStoreMongo.Store
		var err error

		BeforeEach(func() {
			config := storeStructuredMongoTest.NewConfig()
			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			summaryCollection = store.GetCollection("summary")
			summaryRepository = store.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			userId = userTest.RandomID()
			typelessStore = dataStoreSummary.NewTypeless(summaryRepository)
		})

		AfterEach(func() {
			if summaryCollection != nil {
				_, err := summaryCollection.DeleteMany(ctx, bson.D{})
				Expect(err).To(Succeed())
			}

			if store != nil {
				_ = store.Terminate(ctx)
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
		})

	})
})
