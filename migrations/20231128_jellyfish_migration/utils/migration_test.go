package utils_test

import (
	"context"
	"crypto/sha1"
	"encoding/base32"
	"io"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	dataTypesBloodGlucoseTest "github.com/tidepool-org/platform/data/types/blood/glucose/test"
	dataTypesPumpSettingsTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	"github.com/tidepool-org/platform/data/types/upload"
	dataTypesUploadTest "github.com/tidepool-org/platform/data/types/upload/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"

	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils"
)

type fakeMigrationImpl struct {
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
	return &fakeMigrationImpl{
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

func (m *fakeMigrationImpl) Initialize() error {
	return nil
}
func (m *fakeMigrationImpl) Execute(selector bson.M, opt *options.FindOptions, queryFn utils.MigrationQueryFn, updateFn utils.MigrationUpdateFn) error {
	return nil
}
func (m *fakeMigrationImpl) OnError(data utils.ErrorData) {
	m.errorsCount++
}
func (m *fakeMigrationImpl) SetUpdates(data utils.UpdateData) {
	m.updates = append(m.updates, data.GetMongoUpdates(m.rollback, m.rollbackSectionName)...)
}

func (m *fakeMigrationImpl) GetUpdates() []mongo.WriteModel {
	return m.updates
}

func (m *fakeMigrationImpl) GetSettings() utils.Settings {
	cap := 10000
	writeBatchSize := int64(1000)
	return utils.Settings{
		DryRun:              false,
		Rollback:            m.rollback,
		RollbackSectionName: m.rollbackSectionName,
		StopOnErr:           false,
		Cap:                 &cap,
		WriteBatchSize:      &writeBatchSize,
	}
}

func (m *fakeMigrationImpl) ResetUpdates() {
	m.updates = []mongo.WriteModel{}
}

func (m *fakeMigrationImpl) GetCtx() context.Context {
	return context.Background()
}

func (m *fakeMigrationImpl) GetLastID() string {
	return m.lastUpdatedId
}

func (m *fakeMigrationImpl) SetLastProcessed(lastID string) {
	m.lastUpdatedId = lastID
}

func (m *fakeMigrationImpl) GetDataCollection() *mongo.Collection {
	return m.dataC
}

func (m *fakeMigrationImpl) UpdateChecks() error {
	return nil
}

func (m *fakeMigrationImpl) SetFetched(raw []bson.M) {
	m.rawData = append(m.rawData, raw...)
}

var _ = Describe("back-37", func() {
	var _ = Describe("migrationUtil", func() {

		var testData data.Data
		var migration utils.Migration

		var makeJellyfishID = func(fields ...string) *string {
			h := sha1.New()
			hashFields := append(fields, "bootstrap")
			for _, field := range hashFields {
				io.WriteString(h, field)
				io.WriteString(h, "_")
			}
			sha1 := h.Sum(nil)
			id := strings.ToLower(base32.HexEncoding.WithPadding('-').EncodeToString(sha1))
			return &id
		}

		var newData = func(deviceID string, requiredRecords int) data.Data {
			units := pointer.FromString("mg/dL")
			testData := data.Data{}

			for count := 0; count < requiredRecords; count++ {
				typ := test.RandomChoice([]string{"cbg", "smbg", "basal", "pumpSettings"})

				switch typ {
				case "cbg":
					datum := continuous.New()
					datum.Glucose = *dataTypesBloodGlucoseTest.NewGlucose(units)
					datum.Type = "cbg"
					datum.Deduplicator = nil
					datum.DeviceID = pointer.FromAny(deviceID)
					datum.ID = makeJellyfishID(datum.Type, *datum.DeviceID, datum.Time.Format(time.RFC3339))
					testData = append(testData, datum)
				case "smbg":
					datum := selfmonitored.New()
					datum.Glucose = *dataTypesBloodGlucoseTest.NewGlucose(units)
					datum.Type = "smbg"
					datum.SubType = pointer.FromString(test.RandomStringFromArray(selfmonitored.SubTypes()))
					datum.Deduplicator = nil
					datum.DeviceID = pointer.FromAny(deviceID)
					datum.ID = makeJellyfishID(datum.Type, *datum.DeviceID, datum.Time.Format(time.RFC3339))
					testData = append(testData, datum)
				case "basal":
					datum := dataTypesBasalTest.RandomBasal()
					datum.Deduplicator = nil
					datum.DeviceID = pointer.FromAny(deviceID)
					datum.ID = makeJellyfishID(datum.Type, *datum.DeviceID, datum.Time.Format(time.RFC3339))
					testData = append(testData, datum)
				case "pumpSettings":
					datum := dataTypesPumpSettingsTest.NewPump(units)
					datum.Deduplicator = nil
					datum.DeviceID = pointer.FromAny(deviceID)
					datum.ID = makeJellyfishID(datum.Type, *datum.DeviceID, datum.Time.Format(time.RFC3339))
					testData = append(testData, datum)
				}

			}
			return testData
		}

		var newDataSet = func(userID string, deviceID string) *upload.Upload {
			dataSet := dataTypesUploadTest.RandomUpload()
			dataSet.Active = true
			dataSet.ArchivedDataSetID = nil
			dataSet.ArchivedTime = nil
			dataSet.CreatedTime = nil
			dataSet.CreatedUserID = nil
			dataSet.DeletedTime = nil
			dataSet.DeletedUserID = nil
			dataSet.DeviceID = pointer.FromAny(deviceID)
			dataSet.Location.GPS.Origin.Time = nil
			dataSet.ModifiedTime = nil
			dataSet.ModifiedUserID = nil
			dataSet.Origin.Time = nil
			dataSet.UserID = pointer.FromAny(userID)
			return dataSet
		}

		var store *dataStoreMongo.Store
		var repository dataStore.DataRepository
		var ctx context.Context

		var collection *mongo.Collection

		BeforeEach(func() {
			logger := logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
			deviceID := "test-device-88x89"
			testData = newData("test-device-88x89", 2000)
			uploadDataSet := newDataSet("test-user-id", deviceID)
			var err error
			store, err = dataStoreMongo.NewStore(storeStructuredMongoTest.NewConfig())
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			collection = store.GetCollection("deviceData")
			repository = store.NewDataRepository()
			Expect(repository.CreateDataSetData(ctx, uploadDataSet, testData)).To(Succeed())
		})
		AfterEach(func() {
			if collection != nil {
				collection.Database().Drop(ctx)
			}
			if store != nil {
				_ = store.Terminate(ctx)
			}
		})
		It("apply migration", func() {
			migration = newMigrationUtil(collection, false)
			Expect(testData).ToNot(BeNil())
			Expect(len(testData)).To(Equal(2000))
			selector, opt := utils.JellyfishUpdatesQuery(nil, nil, 50, 100)
			Expect(migration.Execute(selector, opt, utils.ProcessJellyfishQueryFn, utils.WriteJellyfishUpdatesFn)).To(Succeed())

			Expect(true).To(Equal(false))
		})
	})
})
