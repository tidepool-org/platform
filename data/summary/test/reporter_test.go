package test_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/data/summary"
	"github.com/tidepool-org/platform/data/summary/reporters"
	. "github.com/tidepool-org/platform/data/summary/test/generators"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Reporters", func() {
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
	var dataCollection *mongo.Collection
	var continuousSummarizer summary.Summarizer[*types.ContinuousStats, *types.ContinuousBucket, types.ContinuousStats, types.ContinuousBucket]
	var realtimeReporter *reporters.PatientRealtimeDaysReporter
	var deviceData []mongo.WriteModel

	Context("PatientRealtimeDaysReporter", func() {

		BeforeEach(func() {
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
			config = storeStructuredMongoTest.NewConfig()

			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store.EnsureIndexes()).To(Succeed())

			dataCollection = store.GetCollection("deviceData")
			summaryRepo = store.NewSummaryRepository().GetStore()
			bucketsRepo = store.NewBucketsRepository().GetStore()
			dataRepo = store.NewDataRepository()
			registry = summary.New(summaryRepo, bucketsRepo, dataRepo)
			userId = userTest.RandomID()

			continuousSummarizer = summary.GetSummarizer[*types.ContinuousStats, *types.ContinuousBucket](registry)
			realtimeReporter = reporters.NewReporter(registry)
		})

		AfterEach(func() {
			_, err = summaryRepo.DeleteMany(ctx, bson.D{})
			Expect(err).ToNot(HaveOccurred())

			_, err = dataCollection.DeleteMany(ctx, bson.D{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("realtime reporter run with mix of users", func() {
			realtimeDatumTime := time.Now().UTC().Truncate(24 * time.Hour)
			userIdTwo := userTest.RandomID()

			uploadRecord := NewDataSet(userId, data.DataSetTypeContinuous)
			err = dataRepo.CreateDataSet(ctx, uploadRecord)
			Expect(err).ToNot(HaveOccurred())

			opts := options.BulkWrite().SetOrdered(false)
			deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, realtimeDatumTime, 10, true)
			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
			Expect(err).ToNot(HaveOccurred())

			uploadRecord = NewDataSet(userIdTwo, data.DataSetTypeContinuous)
			err = dataRepo.CreateDataSet(ctx, uploadRecord)
			Expect(err).ToNot(HaveOccurred())

			deviceData = NewDataSetData("smbg", userIdTwo, realtimeDatumTime, 10, 5)
			_, err = dataCollection.BulkWrite(ctx, deviceData, opts)
			Expect(err).ToNot(HaveOccurred())

			_, err = continuousSummarizer.UpdateSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			_, err = continuousSummarizer.UpdateSummary(ctx, userIdTwo)
			Expect(err).ToNot(HaveOccurred())

			result, err := realtimeReporter.GetRealtimeDaysForUsers(ctx, []string{userId, userIdTwo}, realtimeDatumTime.AddDate(0, -1, 0), realtimeDatumTime)
			Expect(err).ToNot(HaveOccurred())

			Expect(result[userId]).To(Equal(1))
			Expect(result[userIdTwo]).To(Equal(0))
		})

		It("run with a user that doesnt have a summary at all", func() {
			realtimeDatumTime := time.Now().UTC().Truncate(24 * time.Hour)
			userIdTwo := userTest.RandomID()

			uploadRecord := NewDataSet(userId, data.DataSetTypeContinuous)
			err = dataRepo.CreateDataSet(ctx, uploadRecord)
			Expect(err).ToNot(HaveOccurred())

			opts := options.BulkWrite().SetOrdered(false)
			deviceData = NewDataSetDataRealtime("smbg", userId, *uploadRecord.UploadID, realtimeDatumTime, 10, true)
			_, err := dataCollection.BulkWrite(ctx, deviceData, opts)
			Expect(err).ToNot(HaveOccurred())

			_, err = continuousSummarizer.UpdateSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())

			result, err := realtimeReporter.GetRealtimeDaysForUsers(ctx, []string{userId, userIdTwo}, realtimeDatumTime.AddDate(0, -1, 0), realtimeDatumTime)
			Expect(err).ToNot(HaveOccurred())

			Expect(result[userId]).To(Equal(1))
			Expect(result[userIdTwo]).To(Equal(0))
		})
	})

})
