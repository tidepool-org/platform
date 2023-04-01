package store_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	dataStoreSummary "github.com/tidepool-org/platform/data/summary/store"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Mongo", func() {
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

		Context("With a new repository", func() {
			BeforeEach(func() {
				summaryRepository = store.NewBareSummaryRepository()
				Expect(summaryRepository).ToNot(BeNil())
			})

			AfterEach(func() {
				if summaryRepository != nil {
					_, _ = summaryCollection.DeleteMany(context.Background(), bson.D{})
				}
			})

			Context("With new typed Stores", func() {
				var userId string
				var cgmStore *dataStoreSummary.Repo[types.CGMStats, *types.CGMStats]
				var bgmStore *dataStoreSummary.Repo[types.BGMStats, *types.BGMStats]

				var userCGMSummary *types.Summary[types.CGMStats, *types.CGMStats]
				var userBGMSummary *types.Summary[types.BGMStats, *types.BGMStats]

				BeforeEach(func() {
					ctx = log.NewContextWithLogger(context.Background(), logger)
					userId = userTest.RandomID()

					cgmStore = dataStoreSummary.New[types.CGMStats](summaryRepository)
					bgmStore = dataStoreSummary.New[types.BGMStats](summaryRepository)
				})

				Context("UpsertSummary", func() {

					It("Insert CGM Summary", func() {
						userCGMSummary = types.Create[types.CGMStats](userId)
						Expect(userCGMSummary.Type).To(Equal("cgm"))

						err = cgmStore.UpsertSummary(ctx, userCGMSummary)
						Expect(err).ToNot(HaveOccurred())

						userCGMSummaryWritten, err := cgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())

						// copy id, as that was mongo generated
						userCGMSummary.ID = userCGMSummaryWritten.ID
						Expect(userCGMSummaryWritten).To(Equal(userCGMSummary))
					})

					It("Insert BGM Summary", func() {
						userBGMSummary = types.Create[types.BGMStats](userId)
						Expect(userBGMSummary.Type).To(Equal("bgm"))

						err = bgmStore.UpsertSummary(ctx, userBGMSummary)
						Expect(err).ToNot(HaveOccurred())

						userBGMSummaryWritten, err := bgmStore.GetSummary(ctx, userId)
						Expect(err).ToNot(HaveOccurred())

						// copy id, as that was mongo generated
						userBGMSummary.ID = userBGMSummaryWritten.ID
						Expect(userBGMSummaryWritten).To(Equal(userBGMSummary))
					})
				})
			})
		})
	})
})
