package utils_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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

		var testData []map[string]interface{}
		const datumCount = 1000
		var store *dataStoreMongo.Store
		var ctx context.Context
		var rollbackSettings *utils.Settings
		var migrationSettings *utils.Settings

		BeforeEach(func() {
			logger := logTest.NewLogger()
			ctx = platformLog.NewContextWithLogger(context.Background(), logger)
			testData = test.BulkJellyfishData("test-device-88x89", "test-group-id", "test-user-id-123", datumCount)
			rollbackSettings = utils.NewSettings(
				pointer.FromBool(false),
				pointer.FromBool(false),
				pointer.FromBool(true),
				pointer.FromString("_testRollback"),
				nil,
				pointer.FromInt(50),
				pointer.FromInt(100),
				pointer.FromBool(false),
			)
			migrationSettings = utils.NewSettings(
				pointer.FromBool(false),
				pointer.FromBool(false),
				pointer.FromBool(false),
				pointer.FromString("_testRollback"),
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
		It("apply migration will set _deduplicator and also rollback data", func() {
			collection := store.GetCollection("testMigration")
			Expect(setCollectionData(ctx, collection, testData)).To(Succeed())

			migration, err := utils.NewMigration(ctx, migrationSettings, newMongoInstanceChecker(), collection, nil)
			Expect(err).To(BeNil())

			Expect(testData).ToNot(BeNil())
			Expect(len(testData)).To(Equal(datumCount))
			allDocs, err := collection.CountDocuments(ctx, bson.D{})
			Expect(err).To(BeNil())
			Expect(allDocs).To(Equal(int64(datumCount)))
			Expect(migration.Execute(utils.JellyfishDataQueryFn, utils.JellyfishDataUpdatesFn)).To(Succeed())
			stats := migration.GetStats()
			Expect(stats.Errored).To(Equal(0))
			Expect(stats.Fetched).To(Equal(datumCount))
			Expect(stats.Applied).To(Equal(datumCount * 2))

			cur, err := collection.Find(ctx, bson.D{})
			Expect(err).To(BeNil())
			migrated := []map[string]interface{}{}
			cur.All(ctx, &migrated)

			Expect(len(migrated)).To(Equal(datumCount))

			for _, item := range migrated {
				Expect(item).Should(HaveKey("_deduplicator"))
				Expect(item["_deduplicator"]).Should(HaveKey("hash"))
				Expect(item).Should(HaveKey(migration.GetSettings().RollbackSectionName))
			}

		})

		It("apply then rollback migration will return the data to its orginal state", func() {
			collection := store.GetCollection("testRollback")
			Expect(setCollectionData(ctx, collection, testData)).To(Succeed())

			findOptions := options.Find()
			findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})

			migration, err := utils.NewMigration(ctx, migrationSettings, newMongoInstanceChecker(), collection, nil)
			Expect(err).To(BeNil())

			Expect(testData).ToNot(BeNil())
			Expect(len(testData)).To(Equal(datumCount))

			cur, err := collection.Find(ctx, bson.D{}, findOptions)
			Expect(err).To(BeNil())

			original := []map[string]interface{}{}
			cur.All(ctx, &original)
			Expect(len(original)).To(Equal(datumCount))

			Expect(migration.Execute(utils.JellyfishDataQueryFn, utils.JellyfishDataUpdatesFn)).To(Succeed())

			cur, err = collection.Find(ctx, bson.D{}, findOptions)
			Expect(err).To(BeNil())
			migrated := []map[string]interface{}{}
			cur.All(ctx, &migrated)
			Expect(len(migrated)).To(Equal(datumCount))

			rollback, err := utils.NewMigration(ctx, rollbackSettings, newMongoInstanceChecker(), collection, nil)
			Expect(err).To(BeNil())

			Expect(rollback.Execute(utils.JellyfishDataQueryFn, utils.JellyfishDataUpdatesFn)).To(Succeed())

			cur, err = collection.Find(ctx, bson.D{}, findOptions)
			Expect(err).To(BeNil())
			rolledback := []map[string]interface{}{}
			cur.All(ctx, &rolledback)
			Expect(len(rolledback)).To(Equal(datumCount))

			for i, rollbackItem := range rolledback {
				Expect(migrated[i]["_id"]).To(Equal(rollbackItem["_id"]))
				Expect(original[i]).To(Equal(rollbackItem))
			}

		})
	})
})
