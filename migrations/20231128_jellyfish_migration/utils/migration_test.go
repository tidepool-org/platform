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
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

type fakeMigrator struct {
	rollback            bool
	rollbackSectionName string
	ctx                 context.Context
	dataC               *mongo.Collection
	updates             []mongo.WriteModel
	rawData             []bson.M
	errorsCount         int
	updatedCount        int
	lastUpdatedId       string
}

func newMigrationUtil(dataC *mongo.Collection, rollback bool) utils.Migration {
	return &fakeMigrator{
		ctx:                 context.Background(),
		dataC:               dataC,
		rollback:            rollback,
		rollbackSectionName: "_testRollback",
		updates:             []mongo.WriteModel{},
		rawData:             []bson.M{},
		errorsCount:         0,
		updatedCount:        0,
	}
}

func (m *fakeMigrator) Initialize() error {
	return nil
}
func (m *fakeMigrator) Execute(selector bson.M, selectorOpt *options.FindOptions, queryFn utils.MigrationQueryFn, updateFn utils.MigrationUpdateFn) error {
	settings := m.GetSettings()
	for queryFn(m, selector, selectorOpt) {
		updated, err := updateFn(m)
		if err != nil {
			return err
		}
		m.updatedCount += updated
		if settings.Cap != nil {
			if m.GetStats().Fetched >= *settings.Cap {
				break
			}
		}
	}
	return nil
}
func (m *fakeMigrator) OnError(data utils.ErrorData) {
	m.errorsCount++
}
func (m *fakeMigrator) SetUpdates(data utils.UpdateData) {
	m.updates = append(m.updates, data.GetMongoUpdates(m.rollback, m.rollbackSectionName)...)
}
func (m *fakeMigrator) GetUpdates() []mongo.WriteModel {
	return m.updates
}

func (m *fakeMigrator) GetSettings() utils.Settings {
	writeBatchSize := int64(20)
	return utils.Settings{
		DryRun:              false,
		Rollback:            m.rollback,
		RollbackSectionName: m.rollbackSectionName,
		StopOnErr:           false,
		Cap:                 nil,
		WriteBatchSize:      &writeBatchSize,
	}
}

func (m *fakeMigrator) ResetUpdates() {
	m.updates = []mongo.WriteModel{}
}

func (m *fakeMigrator) GetCtx() context.Context {
	return context.Background()
}

func (m *fakeMigrator) GetLastID() string {
	return m.lastUpdatedId
}

func (m *fakeMigrator) SetLastProcessed(lastID string) {
	m.lastUpdatedId = lastID
}

func (m *fakeMigrator) GetDataCollection() *mongo.Collection {
	return m.dataC
}

func (m *fakeMigrator) UpdateChecks() error {
	return nil
}

func (m *fakeMigrator) SetFetched(raw []bson.M) {
	m.rawData = append(m.rawData, raw...)
}

func (m *fakeMigrator) GetStats() utils.MigrationStats {
	return utils.MigrationStats{
		Errored: m.errorsCount,
		Fetched: len(m.rawData),
		Applied: m.updatedCount,
	}
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
		var migration utils.Migration
		const datumCount = 50
		var store *dataStoreMongo.Store
		var ctx context.Context

		BeforeEach(func() {
			logger := logTest.NewLogger()
			ctx = platformLog.NewContextWithLogger(context.Background(), logger)
			testData = test.BulkJellyfishData("test-device-88x89", "test-group-id", "test-user-id-123", datumCount)
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
		It("apply migration", func() {
			collection := store.GetCollection("testMigration")
			Expect(setCollectionData(ctx, collection, testData)).To(Succeed())

			migration = newMigrationUtil(collection, false)
			Expect(testData).ToNot(BeNil())
			Expect(len(testData)).To(Equal(datumCount))
			allDocs, err := collection.CountDocuments(ctx, bson.D{})
			Expect(err).To(BeNil())
			Expect(allDocs).To(Equal(int64(datumCount)))
			selector, opt := utils.JellyfishUpdatesQuery(nil, nil, 50, 100)
			Expect(migration.Execute(selector, opt, utils.ProcessJellyfishQueryFn, utils.WriteJellyfishUpdatesFn)).To(Succeed())
			stats := migration.GetStats()
			Expect(stats.Errored).To(Equal(0))
			Expect(stats.Fetched).To(Equal(datumCount))
			Expect(stats.Applied).To(Equal(datumCount * 3))

			cur, err := collection.Find(ctx, bson.D{})
			Expect(err).To(BeNil())
			migrated := []map[string]interface{}{}
			cur.All(ctx, &migrated)

			Expect(len(migrated)).To(Equal(datumCount))

			for _, item := range migrated {
				Expect(item).Should(HaveKey("_deduplicator"))
				Expect(item).Should(HaveKey(migration.GetSettings().RollbackSectionName))
				Expect(item).ShouldNot(HaveKey("localTime"))
			}

		})

		It("apply then rollback migration will return the data to its orginal state", func() {

			collection := store.GetCollection("testRollback")
			Expect(setCollectionData(ctx, collection, testData)).To(Succeed())

			findOptions := options.Find()
			findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})

			migration = newMigrationUtil(collection, false)
			Expect(testData).ToNot(BeNil())
			Expect(len(testData)).To(Equal(datumCount))

			cur, err := collection.Find(ctx, bson.D{}, findOptions)
			Expect(err).To(BeNil())

			original := []map[string]interface{}{}
			cur.All(ctx, &original)
			Expect(len(original)).To(Equal(datumCount))

			selector, opt := utils.JellyfishUpdatesQuery(nil, nil, 50, 100)
			Expect(migration.Execute(selector, opt, utils.ProcessJellyfishQueryFn, utils.WriteJellyfishUpdatesFn)).To(Succeed())

			cur, err = collection.Find(ctx, bson.D{}, findOptions)
			Expect(err).To(BeNil())
			migrated := []map[string]interface{}{}
			cur.All(ctx, &migrated)
			Expect(len(migrated)).To(Equal(datumCount))

			rollback := newMigrationUtil(collection, true)

			rollbackSelector, rollbackOpt := utils.JellyfishRollbackQuery(rollback.GetSettings().RollbackSectionName, nil, nil, 50, 100)
			Expect(rollback.Execute(rollbackSelector, rollbackOpt, utils.ProcessJellyfishQueryFn, utils.WriteJellyfishUpdatesFn)).To(Succeed())

			cur, err = collection.Find(ctx, bson.D{}, findOptions)
			Expect(err).To(BeNil())
			rolledback := []map[string]interface{}{}
			cur.All(ctx, &rolledback)
			Expect(len(rolledback)).To(Equal(datumCount))

			for i, item := range rolledback {
				Expect(original[i]["_id"]).To(Equal(item["_id"]))
				Expect(migrated[i]["_id"]).To(Equal(item["_id"]))

				Expect(migrated[i]).Should(HaveKey("_deduplicator"))
				Expect(original[i]).ShouldNot(HaveKey("_deduplicator"))
				Expect(item).ShouldNot(HaveKey("_deduplicator"))

				Expect(migrated[i]).Should(HaveKey((migration.GetSettings().RollbackSectionName)))
				Expect(original[i]).ShouldNot(HaveKey((migration.GetSettings().RollbackSectionName)))
				Expect(item).ShouldNot(HaveKey((migration.GetSettings().RollbackSectionName)))
			}

		})
	})
})
