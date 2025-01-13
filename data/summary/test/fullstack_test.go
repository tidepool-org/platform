package test_test

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/data/summary"
	dataStoreSummary "github.com/tidepool-org/platform/data/summary/store"
	. "github.com/tidepool-org/platform/data/summary/test/generators"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	userTest "github.com/tidepool-org/platform/user/test"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var _ = Describe("End to end summary calculations", func() {
	var err error
	var logger log.Logger
	var ctx context.Context
	var registry *summary.SummarizerRegistry
	var config *storeStructuredMongo.Config
	var store *dataStoreMongo.Store
	var summaryRepo *storeStructuredMongo.Repository
	var bucketsRepo *storeStructuredMongo.Repository
	var dataRepo dataStore.DataRepository
	var userId string
	var datumTime time.Time
	var deviceData []mongo.WriteModel
	//var cgmStore *dataStoreSummary.Summaries[*types.CGMStats, *types.GlucoseBucket, types.CGMStats, types.GlucoseBucket]
	var bgmStore *dataStoreSummary.Summaries[*types.BGMStats, *types.GlucoseBucket, types.BGMStats, types.GlucoseBucket]
	//var conStore *dataStoreSummary.Summaries[*types.ContinuousStats, *types.ContinuousBucket, types.ContinuousStats, types.ContinuousBucket]
	var cgmSummarizer summary.Summarizer[*types.CGMStats, *types.GlucoseBucket, types.CGMStats, types.GlucoseBucket]
	var bgmSummarizer summary.Summarizer[*types.BGMStats, *types.GlucoseBucket, types.BGMStats, types.GlucoseBucket]
	var continuousSummarizer summary.Summarizer[*types.ContinuousStats, *types.ContinuousBucket, types.ContinuousStats, types.ContinuousBucket]
	var cgmSummary *types.Summary[*types.CGMStats, *types.GlucoseBucket, types.CGMStats, types.GlucoseBucket]
	var bgmSummary *types.Summary[*types.BGMStats, *types.GlucoseBucket, types.BGMStats, types.GlucoseBucket]
	var conSummary *types.Summary[*types.ContinuousStats, *types.ContinuousBucket, types.ContinuousStats, types.ContinuousBucket]
	var dataCollection *mongo.Collection

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
		config = storeStructuredMongoTest.NewConfig()

		store, err = dataStoreMongo.NewStore(config)
		Expect(err).ToNot(HaveOccurred())
		Expect(store.EnsureIndexes()).To(Succeed())

		summaryRepo = store.NewSummaryRepository().GetStore()
		bucketsRepo = store.NewBucketsRepository().GetStore()
		dataRepo = store.NewDataRepository()
		registry = summary.New(summaryRepo, bucketsRepo, dataRepo)
		userId = userTest.RandomID()
		datumTime = time.Now().UTC().Truncate(time.Hour)
		dataCollection = store.GetCollection("deviceData")

		//cgmStore = dataStoreSummary.NewSummaries[*types.CGMStats, *types.GlucoseBucket](summaryRepo)
		bgmStore = dataStoreSummary.NewSummaries[*types.BGMStats, *types.GlucoseBucket](summaryRepo)
		//conStore = dataStoreSummary.NewSummaries[*types.ContinuousStats, *types.ContinuousBucket](summaryRepo)

		cgmSummarizer = summary.GetSummarizer[*types.CGMStats, *types.GlucoseBucket](registry)
		bgmSummarizer = summary.GetSummarizer[*types.BGMStats, *types.GlucoseBucket](registry)
		continuousSummarizer = summary.GetSummarizer[*types.ContinuousStats, *types.ContinuousBucket](registry)
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
		Expect(len(cgmSummary.Stats.Buckets)).To(Equal(5))
		Expect(*cgmSummary.Stats.Periods["7d"].TotalRecords).To(Equal(5))

		deviceData = NewDataSetData("cbg", userId, datumTime.Add(5*time.Hour), 5, 10)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())
		Expect(len(cgmSummary.Stats.Buckets)).To(Equal(10))
		Expect(*cgmSummary.Stats.Periods["7d"].TotalRecords).To(Equal(10))

		deviceData = NewDataSetData("cbg", userId, datumTime.Add(15*time.Hour), 5, 2)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())
		Expect(len(cgmSummary.Stats.Buckets)).To(Equal(20))
		Expect(*cgmSummary.Stats.Periods["7d"].TotalRecords).To(Equal(15))

		deviceData = NewDataSetData("cbg", userId, datumTime.Add(20*time.Hour), 5, 7)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		deviceData = NewDataSetData("cbg", userId, datumTime.Add(23*time.Hour), 2, 7)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())
		Expect(len(cgmSummary.Stats.Buckets)).To(Equal(28))
		Expect(*cgmSummary.Stats.Periods["7d"].TotalRecords).To(Equal(22))
	})

	It("repeat out of order bgm summary calc", func() {
		opts := options.BulkWrite().SetOrdered(false)

		deviceData = NewDataSetData("smbg", userId, datumTime, 5, 5)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())
		Expect(len(bgmSummary.Stats.Buckets)).To(Equal(5))
		Expect(*bgmSummary.Stats.Periods["7d"].TotalRecords).To(Equal(5))

		deviceData = NewDataSetData("smbg", userId, datumTime.Add(5*time.Hour), 5, 10)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())
		Expect(len(bgmSummary.Stats.Buckets)).To(Equal(10))
		Expect(*bgmSummary.Stats.Periods["7d"].TotalRecords).To(Equal(10))

		deviceData = NewDataSetData("smbg", userId, datumTime.Add(15*time.Hour), 5, 2)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())
		Expect(len(bgmSummary.Stats.Buckets)).To(Equal(20))
		Expect(*bgmSummary.Stats.Periods["7d"].TotalRecords).To(Equal(15))

		deviceData = NewDataSetData("smbg", userId, datumTime.Add(20*time.Hour), 5, 7)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		deviceData = NewDataSetData("smbg", userId, datumTime.Add(23*time.Hour), 2, 7)
		_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())
		Expect(len(bgmSummary.Stats.Buckets)).To(Equal(28))
		Expect(*bgmSummary.Stats.Periods["7d"].TotalRecords).To(Equal(22))
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

		summaries := make([]*types.Summary[*types.BGMStats, *types.GlucoseBucket, types.BGMStats, types.GlucoseBucket], 1)

		// we don't use types.Create as we want to create a sparse jellyfish style upsert
		summaries[0] = &types.Summary[*types.BGMStats, *types.GlucoseBucket, types.BGMStats, types.GlucoseBucket]{
			Type:   types.SummaryTypeBGM,
			UserID: userId,
			Dates: types.Dates{
				OutdatedSince:    &time.Time{},
				HasOutdatedSince: true,
				OutdatedReason:   []string{"LEGACY_DATA_ADDED"},
			},
		}

		count, err := bgmStore.CreateSummaries(ctx, summaries)
		Expect(err).ToNot(HaveOccurred())
		Expect(count).To(Equal(1))

		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())
		Expect(len(bgmSummary.Stats.Buckets)).To(Equal(5))
		Expect(*bgmSummary.Stats.Periods["7d"].TotalRecords).To(Equal(5))
		Expect(bgmSummary.Dates.LastUpdatedReason).To(ConsistOf("LEGACY_DATA_ADDED", types.OutdatedReasonSchemaMigration))
	})

	It("summary calc with no data correctly deletes summaries", func() {
		var t *time.Time

		// create bgm summary
		t, err = bgmSummarizer.SetOutdated(ctx, userId, types.OutdatedReasonUploadCompleted)
		Expect(err).ToNot(HaveOccurred())

		// check that it exists in the db
		bgmSummary, err = bgmSummarizer.GetSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())
		Expect(bgmSummary.Dates.OutdatedSince).To(Equal(t))

		// create cgm summary
		t, err = cgmSummarizer.SetOutdated(ctx, userId, types.OutdatedReasonUploadCompleted)
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
		var cgmSummaryNew *types.Summary[*types.CGMStats, *types.GlucoseBucket, types.CGMStats, types.GlucoseBucket]

		opts := options.BulkWrite().SetOrdered(false)
		deviceData = NewDataSetData("cbg", userId, datumTime, 5, 5)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		// update once for real
		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())
		Expect(cgmSummary.Stats.TotalHours).To(Equal(5))

		_, err = cgmSummarizer.SetOutdated(ctx, userId, types.OutdatedReasonUploadCompleted)
		Expect(err).ToNot(HaveOccurred())

		cgmSummaryNew, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummaryNew).ToNot(BeNil())

		// ensure unchanged
		cgmSummaryNew.ID = cgmSummary.ID
		Expect(cgmSummaryNew).To(BeComparableTo(cgmSummary))
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
		Expect(len(conSummary.Stats.Buckets)).To(Equal(10))

		for i := 0; i < len(conSummary.Stats.Buckets); i++ {
			Expect(conSummary.Stats.Buckets[i].Data.RealtimeRecords).To(Equal(2))
			Expect(conSummary.Stats.Buckets[i].Data.DeferredRecords).To(Equal(0))
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
		Expect(len(conSummary.Stats.Buckets)).To(Equal(10))

		for i := 0; i < len(conSummary.Stats.Buckets); i++ {
			Expect(conSummary.Stats.Buckets[i].Data.RealtimeRecords).To(Equal(0))
			Expect(conSummary.Stats.Buckets[i].Data.DeferredRecords).To(Equal(2))
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
		Expect(conSummary.Dates.LastData).To(BeNil())
		Expect(conSummary.Stats).To(BeNil())
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
		Expect(conSummary.Dates.LastData).To(BeNil())
		Expect(conSummary.Stats).To(BeNil())

		conSummary, err = continuousSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(conSummary).ToNot(BeNil())
		Expect(conSummary.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
		Expect(conSummary.Dates.OutdatedSince).To(BeNil())
		Expect(conSummary.Dates.LastData).To(BeNil())
		Expect(conSummary.Stats).To(BeNil())
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
		Expect(len(conSummary.Stats.Buckets)).To(Equal(200))
	})

	It("bgm summary calc with >batch of data", func() {
		opts := options.BulkWrite().SetOrdered(false)

		deviceData = NewDataSetData("smbg", userId, datumTime, 350, 5)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(bgmSummary).ToNot(BeNil())
		Expect(len(bgmSummary.Stats.Buckets)).To(Equal(350))
	})

	It("cgm summary calc with >batch of data", func() {
		opts := options.BulkWrite().SetOrdered(false)

		deviceData = NewDataSetData("cbg", userId, datumTime, 350, 5)
		_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
		Expect(err).ToNot(HaveOccurred())

		cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
		Expect(err).ToNot(HaveOccurred())
		Expect(cgmSummary).ToNot(BeNil())
		Expect(len(cgmSummary.Stats.Buckets)).To(Equal(350))
	})
})
