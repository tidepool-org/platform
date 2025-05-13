package summary_test

import (
	"context"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/summary"
	"github.com/tidepool-org/platform/summary/store"
	. "github.com/tidepool-org/platform/summary/test"
	. "github.com/tidepool-org/platform/summary/types"
	userTest "github.com/tidepool-org/platform/user/test"
)

func GetBuckets[B BucketDataPt[A], A BucketData](ctx context.Context, userId string, bucketsStore *store.Buckets[B, A]) []*Bucket[B, A] {
	buckets := []*Bucket[B, A]{}

	bucketsCursor, err := bucketsStore.GetAllBuckets(ctx, userId)
	Expect(err).ToNot(HaveOccurred())
	err = bucketsCursor.All(ctx, &buckets)
	Expect(err).ToNot(HaveOccurred())

	return buckets
}

var _ = Describe("End to end summary calculations", func() {
	var err error
	var logger log.Logger
	var ctx context.Context
	var registry *summary.SummarizerRegistry
	var config *storeStructuredMongo.Config
	var mongoStore *dataStoreMongo.Store
	var summaryRepo *storeStructuredMongo.Repository
	var bucketsRepo *storeStructuredMongo.Repository
	var dataRepo dataStore.DataRepository
	var userId string
	var datumTime time.Time
	var deviceData []mongo.WriteModel
	//var cgmStore *dataStoreSummary.Summaries[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]
	var bgmStore *store.Summaries[*BGMPeriods, *GlucoseBucket, BGMPeriods, GlucoseBucket]
	var cgmBucketsStore *store.Buckets[*GlucoseBucket, GlucoseBucket]
	var bgmBucketsStore *store.Buckets[*GlucoseBucket, GlucoseBucket]
	var conBucketsStore *store.Buckets[*ContinuousBucket, ContinuousBucket]
	//var conStore *dataStoreSummary.Summaries[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]
	var cgmSummarizer summary.Summarizer[*CGMPeriods, *GlucoseBucket, CGMPeriods, GlucoseBucket]
	var bgmSummarizer summary.Summarizer[*BGMPeriods, *GlucoseBucket, BGMPeriods, GlucoseBucket]
	var continuousSummarizer summary.Summarizer[*ContinuousPeriods, *ContinuousBucket, ContinuousPeriods, ContinuousBucket]
	var cgmSummary *Summary[*CGMPeriods, *GlucoseBucket, CGMPeriods, GlucoseBucket]
	var bgmSummary *Summary[*BGMPeriods, *GlucoseBucket, BGMPeriods, GlucoseBucket]
	var conSummary *Summary[*ContinuousPeriods, *ContinuousBucket, ContinuousPeriods, ContinuousBucket]
	var dataCollection *mongo.Collection

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
		config = storeStructuredMongoTest.NewConfig()

		mongoStore, err = dataStoreMongo.NewStore(config)
		Expect(err).ToNot(HaveOccurred())
		Expect(mongoStore.EnsureIndexes()).To(Succeed())

		summaryRepo = mongoStore.NewSummaryRepository().GetStore()
		bucketsRepo = mongoStore.NewBucketsRepository().GetStore()
		dataRepo = mongoStore.NewDataRepository()
		registry = summary.New(summaryRepo, bucketsRepo, dataRepo, mongoStore.GetClient())
		userId = userTest.RandomID()
		datumTime = time.Now().UTC().Truncate(time.Hour)
		dataCollection = mongoStore.GetCollection("deviceData")

		//cgmStore = dataStoreSummary.NewSummaries[*types.CGMPeriods, *types.GlucoseBucket](summaryRepo)
		bgmStore = store.NewSummaries[*BGMPeriods, *GlucoseBucket](summaryRepo)
		cgmBucketsStore = store.NewBuckets[*GlucoseBucket](bucketsRepo, SummaryTypeCGM)
		bgmBucketsStore = store.NewBuckets[*GlucoseBucket](bucketsRepo, SummaryTypeBGM)
		conBucketsStore = store.NewBuckets[*ContinuousBucket](bucketsRepo, SummaryTypeContinuous)
		//conStore = dataStoreSummary.NewSummaries[*types.ContinuousPeriods, *types.ContinuousBucket](summaryRepo)

		cgmSummarizer = summary.GetSummarizer[*CGMPeriods, *GlucoseBucket](registry)
		bgmSummarizer = summary.GetSummarizer[*BGMPeriods, *GlucoseBucket](registry)
		continuousSummarizer = summary.GetSummarizer[*ContinuousPeriods, *ContinuousBucket](registry)
	})

	AfterEach(func() {
		_, err = summaryRepo.DeleteMany(ctx, bson.D{})
		Expect(err).ToNot(HaveOccurred())

		_, err = dataCollection.DeleteMany(ctx, bson.D{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("repeat out of order cgm summary calc", func() {
		opts := options.BulkWrite().SetOrdered(false)

		deviceData = NewDataSetData("cbg", userId, datumTime, 5, 5)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())

		buckets := GetBuckets(ctx, userId, cgmBucketsStore)

		Expect(len(buckets)).To(Equal(5))
		Expect(cgmSummary.Periods.GlucosePeriods["7d"].Total.Records).To(Equal(5))

		deviceData = NewDataSetData("cbg", userId, datumTime.Add(5*time.Hour), 5, 10)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())

		buckets = GetBuckets(ctx, userId, cgmBucketsStore)

		Expect(len(buckets)).To(Equal(10))
		Expect(cgmSummary.Periods.GlucosePeriods["7d"].Total.Records).To(Equal(10))

		deviceData = NewDataSetData("cbg", userId, datumTime.Add(15*time.Hour), 5, 2)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())

		buckets = GetBuckets(ctx, userId, cgmBucketsStore)

		Expect(len(buckets)).To(Equal(15))
		Expect(cgmSummary.Periods.GlucosePeriods["7d"].Total.Records).To(Equal(15))

		deviceData = NewDataSetData("cbg", userId, datumTime.Add(20*time.Hour), 5, 7)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		deviceData = NewDataSetData("cbg", userId, datumTime.Add(23*time.Hour), 2, 7)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())

		buckets = GetBuckets(ctx, userId, cgmBucketsStore)

		Expect(len(buckets)).To(Equal(22))
		Expect(cgmSummary.Periods.GlucosePeriods["7d"].Total.Records).To(Equal(22))
	})

	It("repeat out of order bgm summary calc", func() {
		opts := options.BulkWrite().SetOrdered(false)

		deviceData = NewDataSetData("smbg", userId, datumTime, 5, 5)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())

		buckets := GetBuckets(ctx, userId, bgmBucketsStore)

		Expect(len(buckets)).To(Equal(5))
		Expect(bgmSummary.Periods.GlucosePeriods["7d"].Total.Records).To(Equal(5))

		deviceData = NewDataSetData("smbg", userId, datumTime.Add(5*time.Hour), 5, 10)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())

		buckets = GetBuckets(ctx, userId, bgmBucketsStore)

		Expect(len(buckets)).To(Equal(10))
		Expect(bgmSummary.Periods.GlucosePeriods["7d"].Total.Records).To(Equal(10))

		deviceData = NewDataSetData("smbg", userId, datumTime.Add(15*time.Hour), 5, 2)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())

		buckets = GetBuckets(ctx, userId, bgmBucketsStore)

		Expect(len(buckets)).To(Equal(15))
		Expect(bgmSummary.Periods.GlucosePeriods["7d"].Total.Records).To(Equal(15))

		deviceData = NewDataSetData("smbg", userId, datumTime.Add(20*time.Hour), 5, 7)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		deviceData = NewDataSetData("smbg", userId, datumTime.Add(23*time.Hour), 2, 7)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())

		buckets = GetBuckets(ctx, userId, bgmBucketsStore)

		Expect(len(buckets)).To(Equal(22))
		Expect(bgmSummary.Periods.GlucosePeriods["7d"].Total.Records).To(Equal(22))
	})

	It("summary calc with very old data", func() {
		opts := options.BulkWrite().SetOrdered(false)

		deviceData = NewDataSetData("smbg", userId, datumTime.AddDate(-3, 0, 0), 5, 5)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).To(BeNil())
	})

	It("summary calc with jellyfish created summary", func() {
		opts := options.BulkWrite().SetOrdered(false)

		deviceData = NewDataSetData("smbg", userId, datumTime, 5, 5)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		summaries := make([]*Summary[*BGMPeriods, *GlucoseBucket, BGMPeriods, GlucoseBucket], 1)

		// we don't use types.Create as we want to create a sparse jellyfish style upsert
		summaries[0] = &Summary[*BGMPeriods, *GlucoseBucket, BGMPeriods, GlucoseBucket]{
			BaseSummary: BaseSummary{
				Type:   SummaryTypeBGM,
				UserID: userId,
				Dates: Dates{
					OutdatedSince:  &time.Time{},
					OutdatedReason: []string{"LEGACY_DATA_ADDED"},
				},
			},
		}

		count, err := bgmStore.CreateSummaries(ctx, summaries)
		Expect(err).ToNot(HaveOccurred())
		Expect(count).To(Equal(1))

		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())

		buckets := GetBuckets(ctx, userId, bgmBucketsStore)

		Expect(len(buckets)).To(Equal(5))
		Expect(bgmSummary.Periods.GlucosePeriods["7d"].Total.Records).To(Equal(5))
		Expect(bgmSummary.Dates.LastUpdatedReason).To(ConsistOf("LEGACY_DATA_ADDED", OutdatedReasonSchemaMigration))
	})

	It("summary calc with no data correctly deletes summaries", func() {
		var t *time.Time

		// create bgm summary
		t, err = bgmSummarizer.SetOutdated(ctx, userId, OutdatedReasonUploadCompleted)
		Expect(err).ToNot(HaveOccurred())

		// check that it exists in the db
		bgmSummary, err = bgmSummarizer.GetSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())
		Expect(bgmSummary.Dates.OutdatedSince).To(Equal(t))

		// create cgm summary
		t, err = cgmSummarizer.SetOutdated(ctx, userId, OutdatedReasonUploadCompleted)
		Expect(err).ToNot(HaveOccurred())
		// check that it exists in the db
		cgmSummary, err = cgmSummarizer.GetSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())
		Expect(cgmSummary.Dates.OutdatedSince).To(Equal(t))

		// update bgm summary, which should delete it
		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).To(BeNil())

		// confirm its truly gone
		bgmSummary, err = bgmSummarizer.GetSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).To(BeNil())

		// update cgm summary, which should delete it
		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).To(BeNil())

		// confirm its truly gone
		cgmSummary, err = cgmSummarizer.GetSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).To(BeNil())
	})

	It("summary calc with no new data correctly leaves summary unchanged", func() {
		var cgmSummaryNew *Summary[*CGMPeriods, *GlucoseBucket, CGMPeriods, GlucoseBucket]

		opts := options.BulkWrite().SetOrdered(false)
		deviceData = NewDataSetData("cbg", userId, datumTime, 5, 5)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		// update once for real
		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())

		totalHours, err := cgmBucketsStore.GetTotalHours(ctx, userId)
		Expect(err).ToNot(HaveOccurred())

		Expect(totalHours).To(Equal(4))

		// get the real summary stored to the db
		cgmSummary, err = cgmSummarizer.GetSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())

		_, err = cgmSummarizer.SetOutdated(ctx, userId, OutdatedReasonUploadCompleted)
		Expect(err).ToNot(HaveOccurred())

		cgmSummaryNew, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummaryNew).ToNot(BeNil())

		// get the real summary stored to the db
		cgmSummaryNew, err = cgmSummarizer.GetSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())

		// ensure unchanged
		cmpOpts := cmpopts.IgnoreUnexported(GlucosePeriod{})
		Expect(cgmSummaryNew).To(BeComparableTo(cgmSummary, cmpOpts))
	})

	It("summary calc with realtime data", func() {
		realtimeDatumTime := time.Now().UTC().Truncate(24 * time.Hour)

		uploadRecord := NewDataSet(userId, data.DataSetTypeContinuous)
		err = dataRepo.CreateDataSet(ctx, uploadRecord)
		Expect(err).ToNot(HaveOccurred())

		opts := options.BulkWrite().SetOrdered(false)
		deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, realtimeDatumTime, 10, true)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		conSummary, err = continuousSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(conSummary).ToNot(BeNil())

		buckets := GetBuckets(ctx, userId, conBucketsStore)

		Expect(len(buckets)).To(Equal(10))

		for i := 0; i < len(buckets); i++ {
			Expect(buckets[i].Data.Realtime.Records).To(Equal(2))
			Expect(buckets[i].Data.Deferred.Records).To(Equal(0))
		}
	})

	It("summary calc with deferred data", func() {
		deferredDatumTime := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -2)

		uploadRecord := NewDataSet(userId, data.DataSetTypeContinuous)
		err = dataRepo.CreateDataSet(ctx, uploadRecord)
		Expect(err).ToNot(HaveOccurred())

		opts := options.BulkWrite().SetOrdered(false)
		deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, deferredDatumTime, 10, false)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		conSummary, err = continuousSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(conSummary).ToNot(BeNil())

		buckets := GetBuckets(ctx, userId, conBucketsStore)

		Expect(len(buckets)).To(Equal(10))

		for i := 0; i < len(buckets); i++ {
			Expect(buckets[i].Data.Realtime.Records).To(Equal(0))
			Expect(buckets[i].Data.Deferred.Records).To(Equal(2))
		}
	})

	It("summary calc with non-continuous data", func() {
		deferredDatumTime := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -2)

		uploadRecord := NewDataSet(userId, data.DataSetTypeNormal)
		err = dataRepo.CreateDataSet(ctx, uploadRecord)
		Expect(err).ToNot(HaveOccurred())

		opts := options.BulkWrite().SetOrdered(false)
		deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, deferredDatumTime, 10, true)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		conSummary, err = continuousSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(conSummary).ToNot(BeNil())
		Expect(conSummary.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
		Expect(conSummary.Dates.OutdatedSince).To(BeNil())
		Expect(conSummary.Dates.LastData).To(BeZero())
		Expect(conSummary.Periods).To(BeNil())
	})

	It("summary calc with non-continuous data multiple times", func() {
		deferredDatumTime := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -2)

		uploadRecord := NewDataSet(userId, data.DataSetTypeNormal)
		err = dataRepo.CreateDataSet(ctx, uploadRecord)
		Expect(err).ToNot(HaveOccurred())

		opts := options.BulkWrite().SetOrdered(false)
		deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, deferredDatumTime, 10, true)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		conSummary, err = continuousSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(conSummary).ToNot(BeNil())
		Expect(conSummary.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
		Expect(conSummary.Dates.OutdatedSince).To(BeNil())
		Expect(conSummary.Dates.LastData).To(BeZero())
		Expect(conSummary.Periods).To(BeNil())

		conSummary, err = continuousSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(conSummary).ToNot(BeNil())
		Expect(conSummary.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
		Expect(conSummary.Dates.OutdatedSince).To(BeNil())
		Expect(conSummary.Dates.LastData).To(BeZero())
		Expect(conSummary.Periods).To(BeNil())
	})

	It("continuous summary calc with >batch of realtime data", func() {
		realtimeDatumTime := time.Now().UTC().Truncate(24 * time.Hour)

		uploadRecord := NewDataSet(userId, data.DataSetTypeContinuous)
		err = dataRepo.CreateDataSet(ctx, uploadRecord)
		Expect(err).ToNot(HaveOccurred())

		opts := options.BulkWrite().SetOrdered(false)
		deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, realtimeDatumTime, 200, true)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		conSummary, err = continuousSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(conSummary).ToNot(BeNil())

		buckets := GetBuckets(ctx, userId, conBucketsStore)

		Expect(len(buckets)).To(Equal(200))
	})

	It("bgm summary calc with >batch of data", func() {
		opts := options.BulkWrite().SetOrdered(false)

		deviceData = NewDataSetData("smbg", userId, datumTime, 350, 5)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())

		buckets := GetBuckets(ctx, userId, bgmBucketsStore)

		Expect(len(buckets)).To(Equal(350))
	})

	It("cgm summary calc with >batch of data", func() {
		opts := options.BulkWrite().SetOrdered(false)

		deviceData = NewDataSetData("cbg", userId, datumTime, 350, 5)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())

		buckets := GetBuckets(ctx, userId, cgmBucketsStore)

		Expect(len(buckets)).To(Equal(350))
	})

	It("cgm summary calc with the same data range twice, with new modifiedTime", func() {
		opts := options.BulkWrite().SetOrdered(false)
		hourAgo := time.Now().UTC().Truncate(time.Millisecond).Add(-time.Hour)

		deviceData = NewDataSetDataModifiedTime("cbg", userId, datumTime, hourAgo, 7*24, 5)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())

		buckets := GetBuckets(ctx, userId, cgmBucketsStore)
		Expect(len(buckets)).To(Equal(7 * 24))

		deviceData = NewDataSetDataModifiedTime("cbg", userId, datumTime, cgmSummary.Dates.LastUpdatedDate.Add(time.Second), 5*24, 5)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())

		buckets = GetBuckets(ctx, userId, cgmBucketsStore)
		Expect(len(buckets)).To(Equal(7 * 24))
	})
})
