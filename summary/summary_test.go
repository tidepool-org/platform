package summary_test

import (
	"context"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
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
	var mongoStore *dataStoreMongo.Store
	var summaryRepo *storeStructuredMongo.Repository
	var bucketsRepo *storeStructuredMongo.Repository
	var eventsRepo *storeStructuredMongo.Repository
	var dataRepo dataStore.DataRepository
	var userId string
	var datumTime time.Time
	var deviceData []mongo.WriteModel
	var cgmStore *store.CGMSummaries
	var bgmStore *store.BGMSummaries
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
		mongoStore = GetSuiteStore()

		summaryRepo = mongoStore.NewSummaryRepository().GetStore()
		bucketsRepo = mongoStore.NewBucketsRepository().GetStore()
		eventsRepo = mongoStore.NewEventsRepository().GetStore()
		dataRepo = mongoStore.NewDataRepository()
		registry = summary.New(summaryRepo, bucketsRepo, eventsRepo, dataRepo, mongoStore.GetClient())
		userId = userTest.RandomID()
		datumTime = time.Now().UTC().Truncate(time.Hour)
		dataCollection = mongoStore.GetCollection("deviceData")

		cgmStore = store.NewSummaries[*CGMPeriods, *GlucoseBucket](summaryRepo)
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
		bgmSummary := RandomBGMSummary(userId)
		t = bgmSummary.Dates.OutdatedSince
		_, err = bgmStore.CreateSummaries(ctx, []*BGMSummary{bgmSummary})
		Expect(err).To(Succeed())

		// check that it exists in the db
		bgmSummary, err = bgmSummarizer.GetSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())
		Expect(bgmSummary.Dates.OutdatedSince).To(Equal(t))

		// create cgm summary
		cgmSummary := RandomCGMSummary(userId)
		t = cgmSummary.Dates.OutdatedSince
		_, err = cgmStore.CreateSummaries(ctx, []*CGMSummary{cgmSummary})
		Expect(err).To(Succeed())

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

var _ = Describe("GetOutdateUserIDs", func() {

	It("succeeds", func() {
		time1, time2 := time.Now().Add(-5*time.Minute), time.Now()
		eventsRepo := newMockEventsRepository(
			&store.SummaryEvent{UserID: "foo", Time: time1},
			&store.SummaryEvent{UserID: "bar", Time: time2})
		registry := summary.New(nil, nil, eventsRepo, nil, nil)
		summarizer := summary.GetSummarizer[*CGMPeriods, *GlucoseBucket](registry)

		resp, err := summarizer.GetOutdatedUserIDs(GinkgoT().Context(), nil)
		Expect(err).To(Succeed())
		Expect(resp.UserIds).To(Equal([]string{"foo", "bar"}))
		Expect(resp.Start.Equal(truncTimeLikeMongoDB(time1))).To(BeTrue())
		Expect(resp.End.Equal(truncTimeLikeMongoDB(time2))).To(BeTrue())
	})

	It("updates metrics", func() {
		reg := prepPromRegistry()
		eventsRepo := newMockEventsRepository(&store.SummaryEvent{UserID: "foo"})
		registry := summary.New(nil, nil, eventsRepo, nil, nil)
		summarizer := summary.GetSummarizer[*CGMPeriods, *GlucoseBucket](registry)

		_, err := summarizer.GetOutdatedUserIDs(GinkgoT().Context(), nil)
		Expect(err).To(Succeed())

		families, err := reg.Gather()
		Expect(err).To(Succeed())
		Expect(len(families)).To(Equal(2))
	})

	It("errors if Pagination.Page is not 0", func() {
		// This value isn't used, so don't let anyone think it's somehow useful.
		eventsRepo := newMockEventsRepository(&store.SummaryEvent{UserID: "foo"})
		registry := summary.New(nil, nil, eventsRepo, nil, nil)
		summarizer := summary.GetSummarizer[*CGMPeriods, *GlucoseBucket](registry)

		_, err := summarizer.GetOutdatedUserIDs(GinkgoT().Context(), &page.Pagination{Page: 1})
		Expect(err.Error()).To(ContainSubstring("page is not supported"))
	})
})

// truncTimeLikeMongoDB adjusts the time, truncating to milliseconds, just like MongoDB does
// when it serializes a time.Time into BSON.
func truncTimeLikeMongoDB(t time.Time) time.Time {
	return t.Truncate(time.Millisecond)
}

func prepPromRegistry() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(store.QueueLag)
	reg.MustRegister(store.QueueLength)
	return reg
}

type mockEventsRepository struct {
	Count    *int64
	CountErr error

	FindCursor *mongo.Cursor
	FindErr    error
}

func newMockEventsRepository(docs ...any) *mockEventsRepository {
	r := &mockEventsRepository{}
	r.setFindResponse(docs)
	return r
}

func (m *mockEventsRepository) CountDocuments(_ context.Context, _ any, _ ...*options.CountOptions) (int64, error) {
	if m.CountErr != nil {
		return 0, m.CountErr
	}
	if m.Count != nil {
		return *m.Count, nil
	}
	return 0, nil
}

func (m *mockEventsRepository) Find(_ context.Context, _ any, _ ...*options.FindOptions) (*mongo.Cursor, error) {
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	if m.FindCursor != nil {
		return m.FindCursor, nil
	}
	return nil, nil
}

func (m *mockEventsRepository) setFindResponse(docs []any) {
	GinkgoHelper()
	cursor, err := mongo.NewCursorFromDocuments(docs, nil, nil)
	Expect(err).To(Succeed())
	m.FindCursor = cursor
}

func (m *mockEventsRepository) FindOneAndUpdate(ctx context.Context, filter any, update any, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	panic("not implemented") // TODO: Implement
}
