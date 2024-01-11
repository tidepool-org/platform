package summary_test

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data/types/blood/glucose"

	baseDatum "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/pointer"

	"go.mongodb.org/mongo-driver/bson"

	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/data/summary"
	dataStoreSummary "github.com/tidepool-org/platform/data/summary/store"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"

	userTest "github.com/tidepool-org/platform/user/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/summary/types"
)

const units = "mmol/L"

func NewDataSetData(typ string, deviceId string, userId string, startTime time.Time, hours float64, glucoseValue float64) []mongo.WriteModel {
	requiredRecords := int(hours * 1)
	var dataSetData = make([]mongo.WriteModel, requiredRecords)

	for count := 0; count < requiredRecords; count++ {
		datumTime := startTime.Add(time.Duration(-(count + 1)) * time.Minute * 60)
		datum := NewGlucose(typ, units, &datumTime, deviceId, userId, glucoseValue)
		dataSetData[count] = mongo.NewInsertOneModel().SetDocument(datum)
	}
	return dataSetData
}

func NewDatum(typ string) *baseDatum.Base {
	datum := baseDatum.New(typ)
	datum.Time = pointer.FromAny(time.Now().UTC())
	datum.Active = true
	Expect(datum.GetType()).To(Equal(typ))
	return &datum
}

func NewOldDatum(typ string) *baseDatum.Base {
	datum := NewDatum(typ)
	datum.Active = true
	datum.Time = pointer.FromAny(time.Now().UTC().AddDate(0, -24, -1))
	return datum
}

func NewNewDatum(typ string) *baseDatum.Base {
	datum := NewDatum(typ)
	datum.Active = true
	datum.Time = pointer.FromAny(time.Now().UTC().AddDate(0, 0, 2))
	return datum
}

func NewGlucose(typ string, units string, datumTime *time.Time, deviceID string, userID string, value float64) *glucose.Glucose {
	timestamp := time.Now()

	datum := glucose.New(typ)
	datum.Units = &units

	datum.Active = true
	datum.ArchivedDataSetID = nil
	datum.ArchivedTime = nil
	datum.CreatedTime = &timestamp
	datum.CreatedUserID = nil
	datum.DeletedTime = nil
	datum.DeletedUserID = nil
	datum.DeviceID = &deviceID
	datum.ModifiedTime = &timestamp
	datum.ModifiedUserID = nil
	datum.Time = datumTime
	datum.UserID = &userID
	datum.Value = &value

	return &datum
}

var _ = Describe("Summary", func() {
	Context("MaybeUpdateSummary", func() {
		var err error
		var empty struct{}
		var logger log.Logger
		var ctx context.Context
		var registry *summary.SummarizerRegistry
		var config *storeStructuredMongo.Config
		var store *dataStoreMongo.Store
		var summaryRepository *storeStructuredMongo.Repository
		var dataStore dataStore.DataRepository
		var userId string
		var cgmStore *dataStoreSummary.Repo[types.CGMStats, *types.CGMStats]
		var bgmStore *dataStoreSummary.Repo[types.BGMStats, *types.BGMStats]

		BeforeEach(func() {
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
			config = storeStructuredMongoTest.NewConfig()

			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())

			summaryRepository = store.NewSummaryRepository().GetStore()
			dataStore = store.NewDataRepository()
			registry = summary.New(summaryRepository, dataStore)
			userId = userTest.RandomID()

			cgmStore = dataStoreSummary.New[types.CGMStats, *types.CGMStats](summaryRepository)
			bgmStore = dataStoreSummary.New[types.BGMStats, *types.BGMStats](summaryRepository)
		})

		AfterEach(func() {
			_, err = summaryRepository.DeleteMany(ctx, bson.D{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("with all summary types outdated", func() {
			updatesSummary := map[string]struct{}{
				"cgm": empty,
				"bgm": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(HaveLen(2))
			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeCGM))
			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeBGM))

			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userCgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeCGM]))

			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userBgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeBGM]))
		})

		It("with cgm summary type outdated", func() {
			updatesSummary := map[string]struct{}{
				"cgm": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(HaveLen(1))
			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeCGM))

			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userCgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeCGM]))

			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBgmSummary).To(BeNil())
		})

		It("with bgm summary type outdated", func() {
			updatesSummary := map[string]struct{}{
				"bgm": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(HaveLen(1))
			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeBGM))

			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userCgmSummary).To(BeNil())

			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userBgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeBGM]))
		})

		It("with unknown summary type outdated", func() {
			updatesSummary := map[string]struct{}{
				"food": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(BeEmpty())
		})
	})

	Context("CheckDatumUpdatesSummary", func() {
		It("with non-summary type", func() {
			var updatesSummary map[string]struct{}
			datum := NewDatum(food.Type)

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(BeEmpty())
		})

		It("with too old summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewOldDatum(continuous.Type)

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})

		It("with future summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewNewDatum(continuous.Type)

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})

		It("with CGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewDatum(continuous.Type)

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(1))
			Expect(updatesSummary).To(HaveKey(types.SummaryTypeCGM))
		})

		It("with BGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewDatum(selfmonitored.Type)

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(1))
			Expect(updatesSummary).To(HaveKey(types.SummaryTypeBGM))
		})

		It("with inactive BGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewDatum(selfmonitored.Type)
			datum.Active = false

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})

		It("with inactive CGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewDatum(continuous.Type)
			datum.Active = false

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})
	})

	Context("end to end summary calculation", func() {
		var err error
		var logger log.Logger
		var ctx context.Context
		var registry *summary.SummarizerRegistry
		var config *storeStructuredMongo.Config
		var store *dataStoreMongo.Store
		var summaryRepository *storeStructuredMongo.Repository
		var dataStore dataStore.DataRepository
		var userId string
		//var cgmStore *dataStoreSummary.Repo[types.CGMStats, *types.CGMStats]
		//var bgmStore *dataStoreSummary.Repo[types.BGMStats, *types.BGMStats]
		var cgmSummarizer summary.Summarizer[types.CGMStats, *types.CGMStats]
		var bgmSummarizer summary.Summarizer[types.BGMStats, *types.BGMStats]
		var dataCollection *mongo.Collection
		var deviceId string
		var datumTime time.Time

		BeforeEach(func() {
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
			config = storeStructuredMongoTest.NewConfig()

			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())

			summaryRepository = store.NewSummaryRepository().GetStore()
			dataCollection = store.GetCollection("deviceData")
			dataStore = store.NewDataRepository()
			registry = summary.New(summaryRepository, dataStore)
			userId = userTest.RandomID()
			deviceId = "SummaryTestDevice"

			//cgmStore = dataStoreSummary.New[types.CGMStats, *types.CGMStats](summaryRepository)
			//bgmStore = dataStoreSummary.New[types.BGMStats, *types.BGMStats](summaryRepository)

			cgmSummarizer = summary.GetSummarizer[types.CGMStats, *types.CGMStats](registry)
			bgmSummarizer = summary.GetSummarizer[types.BGMStats, *types.BGMStats](registry)

			datumTime = time.Now().UTC().Truncate(time.Hour)
		})

		It("repeat summary calc", func() {
			var userSummary *types.Summary[types.CGMStats, *types.CGMStats]
			var deviceData []mongo.WriteModel
			opts := options.BulkWrite().SetOrdered(false)

			deviceData = NewDataSetData("cbg", deviceId, userId, datumTime, 5, 5)
			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
			Expect(err).ToNot(HaveOccurred())

			userSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userSummary).ToNot(BeNil())
			Expect(len(userSummary.Stats.Buckets)).To(Equal(5))
			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(5))

			deviceData = NewDataSetData("cbg", deviceId, userId, datumTime.Add(5*time.Hour), 5, 10)
			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
			Expect(err).ToNot(HaveOccurred())

			userSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userSummary).ToNot(BeNil())
			Expect(len(userSummary.Stats.Buckets)).To(Equal(10))
			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(10))

			deviceData = NewDataSetData("cbg", deviceId, userId, datumTime.Add(15*time.Hour), 5, 2)
			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
			Expect(err).ToNot(HaveOccurred())

			userSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userSummary).ToNot(BeNil())
			Expect(len(userSummary.Stats.Buckets)).To(Equal(20))
			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(15))

			deviceData = NewDataSetData("cbg", deviceId, userId, datumTime.Add(20*time.Hour), 5, 7)
			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
			Expect(err).ToNot(HaveOccurred())

			deviceData = NewDataSetData("cbg", deviceId, userId, datumTime.Add(23*time.Hour), 2, 7)
			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
			Expect(err).ToNot(HaveOccurred())

			userSummary, err = cgmSummarizer.UpdateSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userSummary).ToNot(BeNil())
			Expect(len(userSummary.Stats.Buckets)).To(Equal(28))
			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(22))
		})

		It("repeat summary calc", func() {
			var userSummary *types.Summary[types.BGMStats, *types.BGMStats]
			var deviceData []mongo.WriteModel
			opts := options.BulkWrite().SetOrdered(false)

			deviceData = NewDataSetData("smbg", deviceId, userId, datumTime, 5, 5)
			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
			Expect(err).ToNot(HaveOccurred())

			userSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userSummary).ToNot(BeNil())
			Expect(len(userSummary.Stats.Buckets)).To(Equal(5))
			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(5))

			deviceData = NewDataSetData("smbg", deviceId, userId, datumTime.Add(5*time.Hour), 5, 10)
			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
			Expect(err).ToNot(HaveOccurred())

			userSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userSummary).ToNot(BeNil())
			Expect(len(userSummary.Stats.Buckets)).To(Equal(10))
			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(10))

			deviceData = NewDataSetData("smbg", deviceId, userId, datumTime.Add(15*time.Hour), 5, 2)
			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
			Expect(err).ToNot(HaveOccurred())

			userSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userSummary).ToNot(BeNil())
			Expect(len(userSummary.Stats.Buckets)).To(Equal(20))
			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(15))

			deviceData = NewDataSetData("smbg", deviceId, userId, datumTime.Add(20*time.Hour), 5, 7)
			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
			Expect(err).ToNot(HaveOccurred())

			deviceData = NewDataSetData("smbg", deviceId, userId, datumTime.Add(23*time.Hour), 2, 7)
			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
			Expect(err).ToNot(HaveOccurred())

			userSummary, err = bgmSummarizer.UpdateSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userSummary).ToNot(BeNil())
			Expect(len(userSummary.Stats.Buckets)).To(Equal(28))
			Expect(*userSummary.Stats.Periods["7d"].TotalRecords).To(Equal(22))
		})
	})
})
