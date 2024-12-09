package test_test

//
//import (
//	"context"
//	"time"
//
//	. "github.com/onsi/ginkgo/v2"
//	. "github.com/onsi/gomega"
//	"go.mongodb.org/mongo-driver/bson"
//	"go.mongodb.org/mongo-driver/mongo"
//	"go.mongodb.org/mongo-driver/mongo/options"
//
//	"github.com/tidepool-org/platform/data"
//	dataStore "github.com/tidepool-org/platform/data/store"
//	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
//	"github.com/tidepool-org/platform/data/summary"
//	"github.com/tidepool-org/platform/data/summary/reporters"
//	dataStoreSummary "github.com/tidepool-org/platform/data/summary/store"
//	"github.com/tidepool-org/platform/data/summary/types"
//	"github.com/tidepool-org/platform/data/test"
//	baseDatum "github.com/tidepool-org/platform/data/types"
//	"github.com/tidepool-org/platform/data/types/blood/glucose"
//	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
//	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
//	"github.com/tidepool-org/platform/data/types/food"
//	"github.com/tidepool-org/platform/data/types/upload"
//	dataTypesUploadTest "github.com/tidepool-org/platform/data/types/upload/test"
//	"github.com/tidepool-org/platform/log"
//	logTest "github.com/tidepool-org/platform/log/test"
//	"github.com/tidepool-org/platform/pointer"
//	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
//	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
//	userTest "github.com/tidepool-org/platform/user/test"
//)
//
//const units = "mmol/L"
//
//func NewDataSet(userID string, typ string) *upload.Upload {
//	var deviceId = "SummaryTestDevice"
//	var timestamp = time.Now().UTC().Truncate(time.Millisecond)
//
//	dataSet := dataTypesUploadTest.RandomUpload()
//	dataSet.DataSetType = &typ
//	dataSet.Active = true
//	dataSet.ArchivedDataSetID = nil
//	dataSet.ArchivedTime = nil
//	dataSet.CreatedTime = &timestamp
//	dataSet.CreatedUserID = nil
//	dataSet.DeletedTime = nil
//	dataSet.DeletedUserID = nil
//	dataSet.DeviceID = &deviceId
//	dataSet.Location.GPS.Origin.Time = nil
//	dataSet.ModifiedTime = &timestamp
//	dataSet.ModifiedUserID = nil
//	dataSet.Origin.Time = nil
//	dataSet.UserID = &userID
//	return dataSet
//}
//
//func NewDataSetData(typ string, userId string, startTime time.Time, hours float64, glucoseValue float64) []mongo.WriteModel {
//	requiredRecords := int(hours * 1)
//	var dataSetData = make([]mongo.WriteModel, requiredRecords)
//	var uploadId = test.RandomSetID()
//	var deviceId = "SummaryTestDevice"
//
//	for count := 0; count < requiredRecords; count++ {
//		datumTime := startTime.Add(time.Duration(-(count + 1)) * time.Minute * 60)
//		datum := NewGlucose(typ, units, &datumTime, deviceId, userId, uploadId, glucoseValue)
//		dataSetData[count] = mongo.NewInsertOneModel().SetDocument(datum)
//	}
//	return dataSetData
//}
//
//func NewDataSetDataRealtime(typ string, userId string, uploadId string, startTime time.Time, hours float64, realtime bool) []mongo.WriteModel {
//	requiredRecords := int(hours * 2)
//
//	var dataSetData = make([]mongo.WriteModel, requiredRecords)
//	var glucoseValue = 5.0
//	var deviceId = "SummaryTestDevice"
//
//	// generate X hours of data
//	for count := 0; count < requiredRecords; count += 1 {
//		datumTime := startTime.Add(time.Duration(count-requiredRecords) * time.Minute * 30)
//
//		datum := NewGlucose(typ, units, &datumTime, deviceId, userId, uploadId, glucoseValue)
//		datum.Value = pointer.FromFloat64(glucoseValue)
//
//		if realtime {
//			datum.CreatedTime = pointer.FromAny(datumTime.Add(5 * time.Minute))
//			datum.ModifiedTime = pointer.FromAny(datumTime.Add(10 * time.Minute))
//		}
//
//		dataSetData[count] = mongo.NewInsertOneModel().SetDocument(datum)
//	}
//
//	return dataSetData
//}
//
//func NewDatum(typ string) *baseDatum.Base {
//	datum := baseDatum.New(typ)
//	datum.Time = pointer.FromAny(time.Now().UTC())
//	datum.Active = true
//	Expect(datum.GetType()).To(Equal(typ))
//	return &datum
//}
//
//func NewOldDatum(typ string) *baseDatum.Base {
//	datum := NewDatum(typ)
//	datum.Active = true
//	datum.Time = pointer.FromAny(time.Now().UTC().AddDate(0, -24, -1))
//	return datum
//}
//
//func NewNewDatum(typ string) *baseDatum.Base {
//	datum := NewDatum(typ)
//	datum.Active = true
//	datum.Time = pointer.FromAny(time.Now().UTC().AddDate(0, 0, 2))
//	return datum
//}
//
//func NewGlucose(typ string, units string, datumTime *time.Time, deviceID string, userID string, uploadId string, value float64) *glucose.Glucose {
//	timestamp := time.Now().UTC().Truncate(time.Millisecond)
//
//	datum := glucose.New(typ)
//	datum.Units = &units
//
//	datum.Active = true
//	datum.ArchivedDataSetID = nil
//	datum.ArchivedTime = nil
//	datum.CreatedTime = &timestamp
//	datum.CreatedUserID = nil
//	datum.DeletedTime = nil
//	datum.DeletedUserID = nil
//	datum.DeviceID = &deviceID
//	datum.ModifiedTime = &timestamp
//	datum.ModifiedUserID = nil
//	datum.Time = datumTime
//	datum.UserID = &userID
//	datum.Value = &value
//	datum.UploadID = &uploadId
//
//	return &datum
//}
//
//var _ = Describe("Summary", func() {
//	Context("MaybeUpdateSummary", func() {
//		var err error
//		var empty struct{}
//		var logger log.Logger
//		var ctx context.Context
//		var registry *summary.SummarizerRegistry
//		var config *storeStructuredMongo.Config
//		var store *dataStoreMongo.Store
//		var summaryRepository *storeStructuredMongo.Repository
//		var dataStore dataStore.DataRepository
//		var userId string
//		var cgmStore *dataStoreSummary.Repo[*types.CGMStats, types.CGMStats]
//		var bgmStore *dataStoreSummary.Repo[*types.BGMStats, types.BGMStats]
//		var continuousStore *dataStoreSummary.Repo[*types.ContinuousStats, types.ContinuousStats]
//
//		BeforeEach(func() {
//			logger = logTest.NewLogger()
//			ctx = log.NewContextWithLogger(context.Background(), logger)
//			config = storeStructuredMongoTest.NewConfig()
//
//			store, err = dataStoreMongo.NewStore(config)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(store.EnsureIndexes()).To(Succeed())
//
//			summaryRepository = store.NewSummaryRepository().GetStore()
//			dataStore = store.NewDataRepository()
//			registry = summary.New(summaryRepository, dataStore)
//			userId = userTest.RandomID()
//
//			cgmStore = dataStoreSummary.New[*types.CGMStats](summaryRepository)
//			bgmStore = dataStoreSummary.New[*types.BGMStats](summaryRepository)
//			continuousStore = dataStoreSummary.New[*types.ContinuousStats](summaryRepository)
//		})
//
//		It("with all summary types outdated", func() {
//			updatesSummary := map[string]struct{}{
//				"cgm":        empty,
//				"bgm":        empty,
//				"continuous": empty,
//			}
//
//			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
//			Expect(outdatedSinceMap).To(HaveLen(3))
//			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeCGM))
//			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeBGM))
//			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeContinuous))
//
//			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(*userCgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeCGM]))
//
//			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(*userBgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeBGM]))
//
//			userContinuousSummary, err := continuousStore.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(*userContinuousSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeContinuous]))
//		})
//
//		It("with cgm summary type outdated", func() {
//			updatesSummary := map[string]struct{}{
//				"cgm": empty,
//			}
//
//			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
//			Expect(outdatedSinceMap).To(HaveLen(1))
//			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeCGM))
//
//			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(*userCgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeCGM]))
//
//			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userBgmSummary).To(BeNil())
//
//			userContinuousSummary, err := continuousStore.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userContinuousSummary).To(BeNil())
//		})
//
//		It("with bgm summary type outdated", func() {
//			updatesSummary := map[string]struct{}{
//				"bgm": empty,
//			}
//
//			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
//			Expect(outdatedSinceMap).To(HaveLen(1))
//			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeBGM))
//
//			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userCgmSummary).To(BeNil())
//
//			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(*userBgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeBGM]))
//
//			userContinuousSummary, err := continuousStore.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userContinuousSummary).To(BeNil())
//		})
//
//		It("with continuous summary type outdated", func() {
//			updatesSummary := map[string]struct{}{
//				"continuous": empty,
//			}
//
//			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
//			Expect(outdatedSinceMap).To(HaveLen(1))
//			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeContinuous))
//
//			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userCgmSummary).To(BeNil())
//
//			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userBgmSummary).To(BeNil())
//
//			userContinuousSummary, err := continuousStore.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(*userContinuousSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeContinuous]))
//		})
//
//		It("with unknown summary type outdated", func() {
//			updatesSummary := map[string]struct{}{
//				"food": empty,
//			}
//
//			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
//			Expect(outdatedSinceMap).To(BeEmpty())
//		})
//	})
//
//	Context("CheckDatumUpdatesSummary", func() {
//		It("with non-summary type", func() {
//			var updatesSummary map[string]struct{}
//			datum := NewDatum(food.Type)
//
//			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
//			Expect(updatesSummary).To(BeEmpty())
//		})
//
//		It("with too old summary affecting record", func() {
//			updatesSummary := make(map[string]struct{})
//			datum := NewOldDatum(continuous.Type)
//
//			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
//			Expect(updatesSummary).To(HaveLen(0))
//		})
//
//		It("with future summary affecting record", func() {
//			updatesSummary := make(map[string]struct{})
//			datum := NewNewDatum(continuous.Type)
//
//			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
//			Expect(updatesSummary).To(HaveLen(0))
//		})
//
//		It("with CGM summary affecting record", func() {
//			updatesSummary := make(map[string]struct{})
//			datum := NewDatum(continuous.Type)
//
//			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
//			Expect(updatesSummary).To(HaveLen(2))
//			Expect(updatesSummary).To(HaveKey(types.SummaryTypeCGM))
//			Expect(updatesSummary).To(HaveKey(types.SummaryTypeContinuous))
//		})
//
//		It("with BGM summary affecting record", func() {
//			updatesSummary := make(map[string]struct{})
//			datum := NewDatum(selfmonitored.Type)
//
//			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
//			Expect(updatesSummary).To(HaveLen(2))
//			Expect(updatesSummary).To(HaveKey(types.SummaryTypeBGM))
//			Expect(updatesSummary).To(HaveKey(types.SummaryTypeContinuous))
//		})
//
//		It("with inactive BGM summary affecting record", func() {
//			updatesSummary := make(map[string]struct{})
//			datum := NewDatum(selfmonitored.Type)
//			datum.Active = false
//
//			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
//			Expect(updatesSummary).To(HaveLen(0))
//		})
//
//		It("with inactive CGM summary affecting record", func() {
//			updatesSummary := make(map[string]struct{})
//			datum := NewDatum(continuous.Type)
//			datum.Active = false
//
//			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
//			Expect(updatesSummary).To(HaveLen(0))
//		})
//	})
//
//	Context("BackfillSummaries", func() {
//		var err error
//		var logger log.Logger
//		var ctx context.Context
//		var config *storeStructuredMongo.Config
//		var store *dataStoreMongo.Store
//		var dataCollection *mongo.Collection
//		var datumTime time.Time
//		var deviceData []mongo.WriteModel
//		var opts *options.BulkWriteOptions
//		var registry *summary.SummarizerRegistry
//		var dataStore dataStore.DataRepository
//		var summaryRepository *storeStructuredMongo.Repository
//		var continuousSummarizer summary.Summarizer[*types.ContinuousStats, types.ContinuousStats]
//
//		BeforeEach(func() {
//			logger = logTest.NewLogger()
//			ctx = log.NewContextWithLogger(context.Background(), logger)
//			config = storeStructuredMongoTest.NewConfig()
//			store, err = dataStoreMongo.NewStore(config)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(store.EnsureIndexes()).To(Succeed())
//
//			dataCollection = store.GetCollection("deviceData")
//			datumTime = time.Now().UTC().Truncate(time.Hour)
//			opts = options.BulkWrite().SetOrdered(false)
//			dataStore = store.NewDataRepository()
//			summaryRepository = store.NewSummaryRepository().GetStore()
//			registry = summary.New(summaryRepository, dataStore)
//			continuousSummarizer = summary.GetSummarizer[*types.ContinuousStats](registry)
//		})
//
//		AfterEach(func() {
//			_, err = summaryRepository.DeleteMany(ctx, bson.D{})
//			Expect(err).ToNot(HaveOccurred())
//
//			_, err = dataCollection.DeleteMany(ctx, bson.D{})
//			Expect(err).ToNot(HaveOccurred())
//		})
//
//		It("backfill continuous summaries", func() {
//			userIdOne := userTest.RandomID()
//			userIdTwo := userTest.RandomID()
//
//			deviceData = NewDataSetData(continuous.Type, userIdOne, datumTime, 2, 5)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			deviceData = NewDataSetData(selfmonitored.Type, userIdTwo, datumTime, 2, 5)
//			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			count, err := continuousSummarizer.BackfillSummaries(ctx)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(count).To(Equal(2))
//
//			userSummary, err := continuousSummarizer.GetSummary(ctx, userIdOne)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary.UserID).To(Equal(userIdOne))
//			Expect(userSummary.Type).To(Equal(types.SummaryTypeContinuous))
//
//			userSummary, err = continuousSummarizer.GetSummary(ctx, userIdTwo)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary.UserID).To(Equal(userIdTwo))
//			Expect(userSummary.Type).To(Equal(types.SummaryTypeContinuous))
//		})
//	})
//
//	Context("end to end summary calculation", func() {
//		var err error
//		var logger log.Logger
//		var ctx context.Context
//		var registry *summary.SummarizerRegistry
//		var config *storeStructuredMongo.Config
//		var store *dataStoreMongo.Store
//		var summaryRepository *storeStructuredMongo.Repository
//		var dataStore dataStore.DataRepository
//		var userId string
//		//var cgmStore *dataStoreSummary.Repo[types.CGMStats, *types.CGMStats]
//		var bgmStore *dataStoreSummary.Repo[*types.BGMStats, types.BGMStats]
//		var cgmSummarizer summary.Summarizer[*types.CGMStats, types.CGMStats]
//		var bgmSummarizer summary.Summarizer[*types.BGMStats, types.BGMStats]
//		var continuousSummarizer summary.Summarizer[*types.ContinuousStats, types.ContinuousStats]
//		var dataCollection *mongo.Collection
//		var datumTime time.Time
//		var deviceData []mongo.WriteModel
//
//		BeforeEach(func() {
//			logger = logTest.NewLogger()
//			ctx = log.NewContextWithLogger(context.Background(), logger)
//			config = storeStructuredMongoTest.NewConfig()
//
//			store, err = dataStoreMongo.NewStore(config)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(store.EnsureIndexes()).To(Succeed())
//
//			summaryRepository = store.NewSummaryRepository().GetStore()
//			dataCollection = store.GetCollection("deviceData")
//			dataStore = store.NewDataRepository()
//			registry = summary.New(summaryRepository, dataStore)
//			userId = userTest.RandomID()
//
//			//cgmStore = dataStoreSummary.New[types.CGMStats, *types.CGMStats](summaryRepository)
//			bgmStore = dataStoreSummary.New[*types.BGMStats](summaryRepository)
//
//			cgmSummarizer = summary.GetSummarizer[*types.CGMStats](registry)
//			bgmSummarizer = summary.GetSummarizer[*types.BGMStats](registry)
//			continuousSummarizer = summary.GetSummarizer[*types.ContinuousStats](registry)
//
//			datumTime = time.Now().UTC().Truncate(time.Hour)
//		})
//
//		AfterEach(func() {
//			_, err = summaryRepository.DeleteMany(ctx, bson.D{})
//			Expect(err).ToNot(HaveOccurred())
//
//			_, err = dataCollection.DeleteMany(ctx, bson.D{})
//			Expect(err).ToNot(HaveOccurred())
//		})
//
//		It("repeat out of order cgm summary calc", func() {
//			var userSummary *types.Summary[*types.CGMStats, types.CGMStats]
//			opts := options.BulkWrite().SetOrdered(false)
//
//			deviceData = NewDataSetData("cbg", userId, datumTime, 5, 5)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(5))
//			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(5))
//
//			deviceData = NewDataSetData("cbg", userId, datumTime.Add(5*time.Hour), 5, 10)
//			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(10))
//			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(10))
//
//			deviceData = NewDataSetData("cbg", userId, datumTime.Add(15*time.Hour), 5, 2)
//			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(20))
//			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(15))
//
//			deviceData = NewDataSetData("cbg", userId, datumTime.Add(20*time.Hour), 5, 7)
//			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			deviceData = NewDataSetData("cbg", userId, datumTime.Add(23*time.Hour), 2, 7)
//			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(28))
//			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(22))
//		})
//
//		It("repeat out of order bgm summary calc", func() {
//			var userSummary *types.Summary[*types.BGMStats, types.BGMStats]
//			opts := options.BulkWrite().SetOrdered(false)
//
//			deviceData = NewDataSetData("smbg", userId, datumTime, 5, 5)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(5))
//			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(5))
//
//			deviceData = NewDataSetData("smbg", userId, datumTime.Add(5*time.Hour), 5, 10)
//			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(10))
//			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(10))
//
//			deviceData = NewDataSetData("smbg", userId, datumTime.Add(15*time.Hour), 5, 2)
//			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(20))
//			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(15))
//
//			deviceData = NewDataSetData("smbg", userId, datumTime.Add(20*time.Hour), 5, 7)
//			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			deviceData = NewDataSetData("smbg", userId, datumTime.Add(23*time.Hour), 2, 7)
//			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(28))
//			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(22))
//		})
//
//		It("summary calc with very old data", func() {
//			var userSummary *types.Summary[*types.BGMStats, types.BGMStats]
//			opts := options.BulkWrite().SetOrdered(false)
//
//			deviceData = NewDataSetData("smbg", userId, datumTime.AddDate(-3, 0, 0), 5, 5)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).To(BeNil())
//		})
//
//		It("summary calc with jellyfish created summary", func() {
//			var userSummary *types.Summary[*types.BGMStats, types.BGMStats]
//			opts := options.BulkWrite().SetOrdered(false)
//
//			deviceData = NewDataSetData("smbg", userId, datumTime, 5, 5)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			summaries := make([]*types.Summary[*types.BGMStats, types.BGMStats], 1)
//
//			// we don't use types.Create as we want to create a sparse jellyfish style upsert
//			summaries[0] = &types.Summary[*types.BGMStats, types.BGMStats]{
//				Type:   types.SummaryTypeBGM,
//				UserID: userId,
//				Dates: types.Dates{
//					OutdatedSince:    &time.Time{},
//					HasOutdatedSince: true,
//					OutdatedReason:   []string{"LEGACY_DATA_ADDED"},
//				},
//			}
//
//			count, err := bgmStore.CreateSummaries(ctx, summaries)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(count).To(Equal(1))
//
//			userSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(5))
//			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(5))
//			Expect(userSummary.Dates.LastUpdatedReason).To(ConsistOf("LEGACY_DATA_ADDED", types.OutdatedReasonSchemaMigration))
//		})
//
//		It("summary calc with no data correctly deletes summaries", func() {
//			var cgmSummary *types.Summary[*types.CGMStats, types.CGMStats]
//			var bgmSummary *types.Summary[*types.BGMStats, types.BGMStats]
//			var t *time.Time
//
//			// create bgm summary
//			t, err = bgmSummarizer.SetOutdated(ctx, userId, types.OutdatedReasonUploadCompleted)
//			Expect(err).ToNot(HaveOccurred())
//
//			// check that it exists in the db
//			bgmSummary, err = bgmSummarizer.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(bgmSummary).ToNot(BeNil())
//			Expect(bgmSummary.Dates.OutdatedSince).To(Equal(t))
//
//			// create cgm summary
//			t, err = cgmSummarizer.SetOutdated(ctx, userId, types.OutdatedReasonUploadCompleted)
//			Expect(err).ToNot(HaveOccurred())
//			// check that it exists in the db
//			cgmSummary, err = cgmSummarizer.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(cgmSummary).ToNot(BeNil())
//			Expect(cgmSummary.Dates.OutdatedSince).To(Equal(t))
//
//			// update bgm summary, which should delete it
//			bgmSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(bgmSummary).To(BeNil())
//
//			// confirm its truly gone
//			bgmSummary, err = bgmSummarizer.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(bgmSummary).To(BeNil())
//
//			// update cgm summary, which should delete it
//			cgmSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(cgmSummary).To(BeNil())
//
//			// confirm its truly gone
//			cgmSummary, err = cgmSummarizer.GetSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(cgmSummary).To(BeNil())
//		})
//
//		It("summary calc with no new data correctly leaves summary unchanged", func() {
//			var userSummary *types.Summary[*types.CGMStats, types.CGMStats]
//			var userSummaryNew *types.Summary[*types.CGMStats, types.CGMStats]
//
//			opts := options.BulkWrite().SetOrdered(false)
//			deviceData = NewDataSetData("cbg", userId, datumTime, 5, 5)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			// update once for real
//			userSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(userSummary.Stats.TotalHours).To(Equal(5))
//
//			_, err = cgmSummarizer.SetOutdated(ctx, userId, types.OutdatedReasonUploadCompleted)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummaryNew, err = cgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummaryNew).ToNot(BeNil())
//
//			// ensure unchanged
//			userSummaryNew.ID = userSummary.ID
//			Expect(userSummaryNew).To(BeComparableTo(userSummary))
//		})
//
//		It("summary calc with realtime data", func() {
//			var userSummary *types.Summary[*types.ContinuousStats, types.ContinuousStats]
//			realtimeDatumTime := time.Now().UTC().Truncate(24 * time.Hour)
//
//			uploadRecord := NewDataSet(userId, data.DataSetTypeContinuous)
//			err = dataStore.CreateDataSet(ctx, uploadRecord)
//			Expect(err).ToNot(HaveOccurred())
//
//			opts := options.BulkWrite().SetOrdered(false)
//			deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, realtimeDatumTime, 10, true)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = continuousSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(10))
//
//			for i := 0; i < len(userSummary.Stats.Buckets); i++ {
//				Expect(userSummary.Stats.Buckets[i].Data.RealtimeRecords).To(Equal(2))
//				Expect(userSummary.Stats.Buckets[i].Data.DeferredRecords).To(Equal(0))
//			}
//		})
//
//		It("summary calc with deferred data", func() {
//			var userSummary *types.Summary[*types.ContinuousStats, types.ContinuousStats]
//			deferredDatumTime := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -2)
//
//			uploadRecord := NewDataSet(userId, data.DataSetTypeContinuous)
//			err = dataStore.CreateDataSet(ctx, uploadRecord)
//			Expect(err).ToNot(HaveOccurred())
//
//			opts := options.BulkWrite().SetOrdered(false)
//			deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, deferredDatumTime, 10, false)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = continuousSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(10))
//
//			for i := 0; i < len(userSummary.Stats.Buckets); i++ {
//				Expect(userSummary.Stats.Buckets[i].Data.RealtimeRecords).To(Equal(0))
//				Expect(userSummary.Stats.Buckets[i].Data.DeferredRecords).To(Equal(2))
//			}
//		})
//
//		It("summary calc with non-continuous data", func() {
//			var userSummary *types.Summary[*types.ContinuousStats, types.ContinuousStats]
//			deferredDatumTime := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -2)
//
//			uploadRecord := NewDataSet(userId, data.DataSetTypeNormal)
//			err = dataStore.CreateDataSet(ctx, uploadRecord)
//			Expect(err).ToNot(HaveOccurred())
//
//			opts := options.BulkWrite().SetOrdered(false)
//			deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, deferredDatumTime, 10, true)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = continuousSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(userSummary.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
//			Expect(userSummary.Dates.OutdatedSince).To(BeNil())
//			Expect(userSummary.Dates.LastData).To(BeNil())
//			Expect(userSummary.Stats).To(BeNil())
//		})
//
//		It("summary calc with non-continuous data multiple times", func() {
//			var userSummary *types.Summary[*types.ContinuousStats, types.ContinuousStats]
//			deferredDatumTime := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -2)
//
//			uploadRecord := NewDataSet(userId, data.DataSetTypeNormal)
//			err = dataStore.CreateDataSet(ctx, uploadRecord)
//			Expect(err).ToNot(HaveOccurred())
//
//			opts := options.BulkWrite().SetOrdered(false)
//			deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, deferredDatumTime, 10, true)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = continuousSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(userSummary.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
//			Expect(userSummary.Dates.OutdatedSince).To(BeNil())
//			Expect(userSummary.Dates.LastData).To(BeNil())
//			Expect(userSummary.Stats).To(BeNil())
//
//			userSummary, err = continuousSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(userSummary.Dates.LastUpdatedDate.IsZero()).To(BeTrue())
//			Expect(userSummary.Dates.OutdatedSince).To(BeNil())
//			Expect(userSummary.Dates.LastData).To(BeNil())
//			Expect(userSummary.Stats).To(BeNil())
//		})
//
//		It("continuous summary calc with >batch of realtime data", func() {
//			var userSummary *types.Summary[*types.ContinuousStats, types.ContinuousStats]
//			realtimeDatumTime := time.Now().UTC().Truncate(24 * time.Hour)
//
//			uploadRecord := NewDataSet(userId, data.DataSetTypeContinuous)
//			err = dataStore.CreateDataSet(ctx, uploadRecord)
//			Expect(err).ToNot(HaveOccurred())
//
//			opts := options.BulkWrite().SetOrdered(false)
//			deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, realtimeDatumTime, 200, true)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = continuousSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(200))
//		})
//
//		It("bgm summary calc with >batch of data", func() {
//			var userSummary *types.Summary[*types.BGMStats, types.BGMStats]
//			opts := options.BulkWrite().SetOrdered(false)
//
//			deviceData = NewDataSetData("smbg", userId, datumTime, 350, 5)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(350))
//		})
//
//		It("cgm summary calc with >batch of data", func() {
//			var userSummary *types.Summary[*types.CGMStats, types.CGMStats]
//			opts := options.BulkWrite().SetOrdered(false)
//
//			deviceData = NewDataSetData("cbg", userId, datumTime, 350, 5)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			userSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(userSummary).ToNot(BeNil())
//			Expect(len(userSummary.Stats.Buckets)).To(Equal(350))
//		})
//	})
//
//	Context("PatientRealtimeDaysReporter", func() {
//		var err error
//		var logger log.Logger
//		var ctx context.Context
//		var registry *summary.SummarizerRegistry
//		var config *storeStructuredMongo.Config
//		var store *dataStoreMongo.Store
//		var summaryRepository *storeStructuredMongo.Repository
//		var dataStore dataStore.DataRepository
//		var userId string
//		var continuousSummarizer summary.Summarizer[*types.ContinuousStats, types.ContinuousStats]
//		var dataCollection *mongo.Collection
//		var realtimeReporter *reporters.PatientRealtimeDaysReporter
//		var deviceData []mongo.WriteModel
//
//		BeforeEach(func() {
//			logger = logTest.NewLogger()
//			ctx = log.NewContextWithLogger(context.Background(), logger)
//			config = storeStructuredMongoTest.NewConfig()
//
//			store, err = dataStoreMongo.NewStore(config)
//			Expect(err).ToNot(HaveOccurred())
//			Expect(store.EnsureIndexes()).To(Succeed())
//
//			summaryRepository = store.NewSummaryRepository().GetStore()
//			dataCollection = store.GetCollection("deviceData")
//			dataStore = store.NewDataRepository()
//			registry = summary.New(summaryRepository, dataStore)
//			userId = userTest.RandomID()
//
//			continuousSummarizer = summary.GetSummarizer[*types.ContinuousStats](registry)
//			realtimeReporter = reporters.NewReporter(registry)
//		})
//
//		AfterEach(func() {
//			_, err = summaryRepository.DeleteMany(ctx, bson.D{})
//			Expect(err).ToNot(HaveOccurred())
//
//			_, err = dataCollection.DeleteMany(ctx, bson.D{})
//			Expect(err).ToNot(HaveOccurred())
//		})
//
//		It("realtime reporter run with mix of users", func() {
//			realtimeDatumTime := time.Now().UTC().Truncate(24 * time.Hour)
//			userIdTwo := userTest.RandomID()
//
//			uploadRecord := NewDataSet(userId, data.DataSetTypeContinuous)
//			err = dataStore.CreateDataSet(ctx, uploadRecord)
//			Expect(err).ToNot(HaveOccurred())
//
//			opts := options.BulkWrite().SetOrdered(false)
//			deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, realtimeDatumTime, 10, true)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			uploadRecord = NewDataSet(userIdTwo, data.DataSetTypeContinuous)
//			err = dataStore.CreateDataSet(ctx, uploadRecord)
//			Expect(err).ToNot(HaveOccurred())
//
//			deviceData = NewDataSetData("smbg", userIdTwo, realtimeDatumTime, 10, 5)
//			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			_, err = continuousSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//			_, err = continuousSummarizer.UpdateSummary(ctx, userIdTwo)
//			Expect(err).ToNot(HaveOccurred())
//
//			result, err := realtimeReporter.GetRealtimeDaysForUsers(ctx, []string{userId, userIdTwo}, realtimeDatumTime.AddDate(0, -1, 0), realtimeDatumTime)
//			Expect(err).ToNot(HaveOccurred())
//
//			Expect(result[userId]).To(Equal(1))
//			Expect(result[userIdTwo]).To(Equal(0))
//		})
//
//		It("run with a user that doesnt have a summary at all", func() {
//			realtimeDatumTime := time.Now().UTC().Truncate(24 * time.Hour)
//			userIdTwo := userTest.RandomID()
//
//			uploadRecord := NewDataSet(userId, data.DataSetTypeContinuous)
//			err = dataStore.CreateDataSet(ctx, uploadRecord)
//			Expect(err).ToNot(HaveOccurred())
//
//			opts := options.BulkWrite().SetOrdered(false)
//			deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, realtimeDatumTime, 10, true)
//			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
//			Expect(err).ToNot(HaveOccurred())
//
//			_, err = continuousSummarizer.UpdateSummary(ctx, userId)
//			Expect(err).ToNot(HaveOccurred())
//
//			result, err := realtimeReporter.GetRealtimeDaysForUsers(ctx, []string{userId, userIdTwo}, realtimeDatumTime.AddDate(0, -1, 0), realtimeDatumTime)
//			Expect(err).ToNot(HaveOccurred())
//
//			Expect(result[userId]).To(Equal(1))
//			Expect(result[userIdTwo]).To(Equal(0))
//		})
//	})
//})
