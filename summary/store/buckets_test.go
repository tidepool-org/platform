package store_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	dataStoreSummary "github.com/tidepool-org/platform/summary/store"
	. "github.com/tidepool-org/platform/summary/test"
	"github.com/tidepool-org/platform/summary/types"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Buckets", Label("mongodb", "slow", "integration"), func() {
	var logger *logTest.Logger
	var ctx context.Context
	var store *dataStoreMongo.Store
	var bucketsRepository *storeStructuredMongo.Repository

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
	})

	Context("Create repo and store", func() {
		var createStore *dataStoreMongo.Store

		It("CGM Buckets Repo", func() {
			createStore = GetSuiteStore()

			bucketsRepository = createStore.NewBucketsRepository().GetStore()
			Expect(bucketsRepository).ToNot(BeNil())

			cgmBucketsStore := dataStoreSummary.NewBuckets[*types.GlucoseBucket](bucketsRepository, types.SummaryTypeCGM)
			Expect(cgmBucketsStore).ToNot(BeNil())
		})

		It("BGM Buckets Repo", func() {
			createStore = GetSuiteStore()

			bucketsRepository = createStore.NewBucketsRepository().GetStore()
			Expect(bucketsRepository).ToNot(BeNil())

			bgmBucketsStore := dataStoreSummary.NewBuckets[*types.GlucoseBucket](bucketsRepository, types.SummaryTypeBGM)
			Expect(bgmBucketsStore).ToNot(BeNil())
		})

		It("Continuous Buckets Repo", func() {
			createStore = GetSuiteStore()

			bucketsRepository = createStore.NewBucketsRepository().GetStore()
			Expect(bucketsRepository).ToNot(BeNil())

			bgmBucketsStore := dataStoreSummary.NewBuckets[*types.GlucoseBucket](bucketsRepository, types.SummaryTypeBGM)
			Expect(bgmBucketsStore).ToNot(BeNil())

			cgmBucketsStore := dataStoreSummary.NewBuckets[*types.GlucoseBucket](bucketsRepository, types.SummaryTypeCGM)
			Expect(cgmBucketsStore).ToNot(BeNil())

			conBucketsStore := dataStoreSummary.NewBuckets[*types.ContinuousBucket](bucketsRepository, types.SummaryTypeContinuous)
			Expect(conBucketsStore).ToNot(BeNil())
		})
	})

	Context("Store", func() {
		var userId string
		var conStore *dataStoreSummary.Buckets[*types.ContinuousBucket, types.ContinuousBucket]
		var bgmStore *dataStoreSummary.Buckets[*types.GlucoseBucket, types.GlucoseBucket]
		var cgmStore *dataStoreSummary.Buckets[*types.GlucoseBucket, types.GlucoseBucket]
		var bucketTime time.Time

		BeforeEach(func() {
			store = GetSuiteStore()
			bucketsRepository = store.NewBucketsRepository().GetStore()
			Expect(bucketsRepository).ToNot(BeNil())

			bgmStore = dataStoreSummary.NewBuckets[*types.GlucoseBucket](bucketsRepository, types.SummaryTypeBGM)
			Expect(bgmStore).ToNot(BeNil())

			cgmStore = dataStoreSummary.NewBuckets[*types.GlucoseBucket](bucketsRepository, types.SummaryTypeCGM)
			Expect(cgmStore).ToNot(BeNil())

			conStore = dataStoreSummary.NewBuckets[*types.ContinuousBucket](bucketsRepository, types.SummaryTypeContinuous)
			Expect(conStore).ToNot(BeNil())

			userId = userTest.RandomID()
			bucketTime = time.Now().UTC().Truncate(time.Hour)
		})

		AfterEach(func() {
			if bucketsRepository != nil {
				_, err := bucketsRepository.DeleteMany(ctx, bson.D{})
				Expect(err).To(Succeed())
			}
		})

		Context("GetBucketsByTime", func() {
			It("With two buckets in time range, two outside", func() {
				buckets := []types.Bucket[*types.GlucoseBucket, types.GlucoseBucket]{
					// A bucket that's too old by 1s
					{BaseBucket: types.BaseBucket{
						UserId: userId,
						Type:   types.SummaryTypeCGM,
						Time:   bucketTime.Add(-(time.Hour*24 + time.Second)),
					}},
					// A bucket that's right on the lower edge
					{BaseBucket: types.BaseBucket{
						UserId: userId,
						Type:   types.SummaryTypeCGM,
						Time:   bucketTime.Add(-time.Hour * 24),
					}},
					// A bucket that's right on the upper edge
					{BaseBucket: types.BaseBucket{
						UserId: userId,
						Type:   types.SummaryTypeCGM,
						Time:   bucketTime,
					}},
					// A bucket that's too new by 1s
					{BaseBucket: types.BaseBucket{
						UserId: userId,
						Type:   types.SummaryTypeCGM,
						Time:   bucketTime.Add(time.Second),
					}},
				}
				opts := options.BulkWrite().SetOrdered(false)
				_, err := bucketsRepository.BulkWrite(ctx, SliceToInsertWriteModel(buckets), opts)
				Expect(err).ToNot(HaveOccurred())

				r, err := cgmStore.GetBucketsByTime(ctx, userId, bucketTime.Add(-time.Hour*24), bucketTime)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(r)).To(Equal(2))
				Expect(r).To(HaveKey(bucketTime.Truncate(time.Millisecond)))
				Expect(r).To(HaveKey(bucketTime.Add(-time.Hour * 24).Truncate(time.Millisecond)))
			})
		})
	})
})
