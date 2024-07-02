package utils_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	platformLog "github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils"
	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils/test"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

type mongoInstance struct {
	writeBatchSize *int64
}

func newMongoInstanceChecker() utils.MongoInstanceCheck {
	return &mongoInstance{
		writeBatchSize: pointer.FromInt64(100),
	}
}

func (m *mongoInstance) SetWriteBatchSize(ctx context.Context) error {
	return nil
}
func (m *mongoInstance) CheckFreeSpace(ctx context.Context, dataC *mongo.Collection) error {
	return nil
}
func (m *mongoInstance) BlockUntilDBReady(ctx context.Context) error {
	return nil
}
func (m *mongoInstance) GetWriteBatchSize() *int64 {
	return m.writeBatchSize
}

func setCollectionData(ctx context.Context, dataC *mongo.Collection, dataSetData []map[string]interface{}) error {
	insertData := make([]mongo.WriteModel, 0, len(dataSetData))
	for _, datum := range dataSetData {
		insertData = append(insertData, mongo.NewInsertOneModel().SetDocument(datum))
	}
	opts := options.BulkWrite().SetOrdered(false)
	_, err := dataC.BulkWrite(ctx, insertData, opts)
	return err
}

var _ = Describe("back-37", func() {
	var _ = Describe("migrationUtil", func() {

		var testJFData []map[string]interface{}
		var testJFUploads []map[string]interface{}
		const datumCount = 1000
		const uploadCount = 99
		var store *dataStoreMongo.Store
		var ctx context.Context
		var migrationSettings *utils.Settings

		BeforeEach(func() {
			logger := logTest.NewLogger()
			ctx = platformLog.NewContextWithLogger(context.Background(), logger)
			testJFData = test.BulkJellyfishData("test-device-88x89", "test-group-id", "test-user-id-123", datumCount)
			testJFUploads = test.BulkJellyfishData("test-device-other", "test-group-id_2", "test-user-id-987", uploadCount)

			migrationSettings = utils.NewSettings(
				pointer.FromBool(false),
				pointer.FromBool(false),
				nil,
				pointer.FromInt(50),
				pointer.FromInt(100),
				pointer.FromBool(false),
			)
			var err error
			store, err = dataStoreMongo.NewStore(storeStructuredMongoTest.NewConfig())
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
		AfterEach(func() {
			if store != nil {
				_ = store.Terminate(ctx)
			}
		})
		It("will set _deduplicator.hash to be the datum _id for jellyfish data", func() {
			collection := store.GetCollection("testJFDatum")
			Expect(setCollectionData(ctx, collection, testJFData)).To(Succeed())

			migration, err := utils.NewMigration(ctx, migrationSettings, newMongoInstanceChecker(), collection, nil)
			Expect(err).To(BeNil())

			Expect(testJFData).ToNot(BeNil())
			Expect(len(testJFData)).To(Equal(datumCount))
			allDocs, err := collection.CountDocuments(ctx, bson.D{})
			Expect(err).To(BeNil())
			Expect(allDocs).To(Equal(int64(datumCount)))
			Expect(migration.Execute(utils.JellyfishDataQueryFn, utils.JellyfishDataUpdatesFn)).To(Succeed())
			stats := migration.GetStats()
			Expect(stats.Errored).To(Equal(0))
			Expect(stats.Fetched).To(Equal(datumCount))
			Expect(stats.Applied).To(Equal(datumCount))
			cur, err := collection.Find(ctx, bson.D{})
			Expect(err).To(BeNil())
			migrated := []map[string]interface{}{}
			cur.All(ctx, &migrated)

			Expect(len(migrated)).To(Equal(datumCount))

			for _, item := range migrated {
				Expect(item).Should(HaveKey("_deduplicator"))
				Expect(item["_deduplicator"]).Should(HaveLen(1))
				Expect(item["_deduplicator"]).Should(HaveKeyWithValue("hash", item["_id"]))
			}
		})

		It("will set _deduplicator to be the datum _id for jellyfish uploads", func() {
			collection := store.GetCollection("testJFUploads")
			Expect(setCollectionData(ctx, collection, testJFUploads)).To(Succeed())

			migration, err := utils.NewMigration(ctx, migrationSettings, newMongoInstanceChecker(), collection, nil)
			Expect(err).To(BeNil())

			Expect(testJFUploads).ToNot(BeNil())
			Expect(len(testJFUploads)).To(Equal(uploadCount))
			allUploadDocs, err := collection.CountDocuments(ctx, bson.D{})
			Expect(err).To(BeNil())
			Expect(allUploadDocs).To(Equal(int64(uploadCount)))
			Expect(migration.Execute(utils.JellyfishUploadQueryFn, utils.JellyfishDataUpdatesFn)).To(Succeed())
			stats := migration.GetStats()
			Expect(stats.Errored).To(Equal(0))
			Expect(stats.Fetched).To(Equal(uploadCount))
			Expect(stats.Applied).To(Equal(uploadCount))
			cur, err := collection.Find(ctx, bson.D{})
			Expect(err).To(BeNil())
			migrated := []map[string]interface{}{}
			cur.All(ctx, &migrated)

			Expect(len(migrated)).To(Equal(uploadCount))

			for _, item := range migrated {
				Expect(item).Should(HaveKey("_deduplicator"))
				Expect(item["_deduplicator"]).Should(HaveLen(2))
				Expect(item["_deduplicator"]).Should(HaveKeyWithValue("name", deduplicator.DeviceDeactivateLegacyHashName))
				Expect(item["_deduplicator"]).Should(HaveKeyWithValue("version", "0.0.0"))
			}
		})
	})
})
