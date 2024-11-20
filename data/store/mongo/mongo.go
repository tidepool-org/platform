package mongo

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	goComMgo "github.com/mdblp/go-db/mongo"

	"github.com/tidepool-org/platform/data/types/activity/physical"
	"github.com/tidepool-org/platform/data/types/basal/automated"
	"github.com/tidepool-org/platform/data/types/basal/scheduled"
	"github.com/tidepool-org/platform/data/types/basal/temporary"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/bolus/biphasic"
	"github.com/tidepool-org/platform/data/types/bolus/normal"
	"github.com/tidepool-org/platform/data/types/bolus/pen"
	"github.com/tidepool-org/platform/data/types/calculator"
	"github.com/tidepool-org/platform/data/types/device/alarm"
	"github.com/tidepool-org/platform/data/types/device/calibration"
	"github.com/tidepool-org/platform/data/types/device/flush"
	"github.com/tidepool-org/platform/data/types/device/mode"
	"github.com/tidepool-org/platform/data/types/device/prime"
	"github.com/tidepool-org/platform/data/types/device/reservoirchange"
	"github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/data/types/settings/basalsecurity"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/schema"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var dataWriteToReadStoreMetrics = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:      "write_to_read_store_duration_seconds",
	Help:      "A histogram for writing cbg data to read store execution time in ms",
	Subsystem: "data",
	Namespace: "dblp",
}, []string{"type"})

var datumWriteToDeviceDataStoreMetrics = promauto.NewHistogram(prometheus.HistogramOpts{
	Name:      "write_datum_to_device_data_duration_seconds",
	Help:      "A histogram for writing a datum to the device data execution time (ms)",
	Subsystem: "data",
	Namespace: "dblp",
})

var datumWriteToDeviceDataArchiveStoreMetrics = promauto.NewHistogram(prometheus.HistogramOpts{
	Name:      "write_datum_to_device_data_archive_duration_seconds",
	Help:      "A histogram for writing a datum to the device data archive execution time (ms)",
	Subsystem: "data",
	Namespace: "dblp",
})

type Stores struct {
	*storeStructuredMongo.Store
	BucketStore           *MongoBucketStoreClient
	DataTypesArchived     []string
	DataTypesBucketed     []string
	DataTypesKeptInLegacy []string
}

type BucketMigrationConfig struct {
	DataTypesArchived     []string
	DataTypesBucketed     []string
	DataTypesKeptInLegacy []string
}

var (
	dataSourcesIndexes = map[string][]mongo.IndexModel{
		"deviceData": {
			// Additional indexes are also created in `tide-whisperer` and `jellyfish`
			{
				Keys: bson.D{
					{Key: "_userId", Value: 1},
					{Key: "uploadId", Value: 1},
					{Key: "type", Value: 1},
				},
				Options: options.Index().
					SetName("UserIdUploadIdType"),
			},
			{
				Keys: bson.D{
					{Key: "uploadId", Value: 1},
				},
				Options: options.Index().
					SetUnique(true).
					SetPartialFilterExpression(bson.D{{Key: "type", Value: "upload"}}).
					SetName("UniqueUploadId"),
			},
			{
				Keys: bson.D{
					{Key: "_userId", Value: 1},
					{Key: "guid", Value: 1},
					{Key: "deviceId", Value: 1},
				},
				Options: options.Index().
					SetPartialFilterExpression(bson.M{"$and": []bson.M{
						{"deviceId": bson.M{"$exists": true}},
						{"guid": bson.M{"$exists": true}},
					}}).
					SetName("UserIdGuidDeviceId"),
			},
		},
		"deviceData_archive": {
			{
				Keys: bson.D{
					{Key: "_userId", Value: 1},
					{Key: "guid", Value: 1},
					{Key: "deviceId", Value: 1},
				},
				Options: options.Index().
					SetPartialFilterExpression(bson.M{"$and": []bson.M{
						{"deviceId": bson.M{"$exists": true}},
						{"guid": bson.M{"$exists": true}},
					}}).
					SetName("UserIdGuidDeviceId"),
			},
		},
	}
)

func NewStores(cfg *storeStructuredMongo.Config, config *goComMgo.Config, lg *logrus.Logger, migrateConfig BucketMigrationConfig, minimalYearSupportedForData int) (*Stores, error) {
	if config != nil {
		cfg.Indexes = dataSourcesIndexes
	}

	baseStore, err := storeStructuredMongo.NewStore(cfg)
	if err != nil {
		return nil, err
	}

	var bucketStore *MongoBucketStoreClient
	bucketStore, err = NewMongoBucketStoreClient(config, lg, minimalYearSupportedForData)
	if err != nil {
		return nil, err
	}
	bucketStore.Start()

	return &Stores{
		Store:                 baseStore,
		BucketStore:           bucketStore,
		DataTypesArchived:     migrateConfig.DataTypesArchived,
		DataTypesBucketed:     migrateConfig.DataTypesBucketed,
		DataTypesKeptInLegacy: migrateConfig.DataTypesKeptInLegacy,
	}, nil
}

func (s *Stores) NewDataRepository() store.DataRepository {
	return &DataRepository{
		Repository:            s.Store.GetRepository("deviceData"),
		BucketStore:           s.BucketStore,
		DataTypesArchived:     s.DataTypesArchived,
		DataTypesBucketed:     s.DataTypesBucketed,
		DataTypesKeptInLegacy: s.DataTypesKeptInLegacy,
	}
}

type DataRepository struct {
	*storeStructuredMongo.Repository
	BucketStore           *MongoBucketStoreClient
	DataTypesArchived     []string
	DataTypesBucketed     []string
	DataTypesKeptInLegacy []string
}

func (d *DataRepository) GetDataSetsForUserByID(ctx context.Context, userID string, filter *store.Filter, pagination *page.Pagination) ([]*upload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = store.NewFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	now := time.Now()

	var dataSets []*upload.Upload
	selector := bson.M{
		"_active": true,
		"_userId": userID,
		"type":    "upload",
	}
	if !filter.Deleted {
		selector["deletedTime"] = bson.M{"$exists": false}
	}
	if filter.State != nil {
		selector["_state"] = *filter.State
	}
	if filter.DataSetType != nil {
		selector["dataSetType"] = *filter.DataSetType
	}
	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"time": -1})
	cursor, err := d.Find(ctx, selector, opts)

	loggerFields := log.Fields{"userId": userID, "dataSetsCount": len(dataSets), "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("GetDataSetsForUserByID")

	if err != nil {
		return nil, errors.Wrap(err, "unable to get data sets for user by id")
	}

	if err = cursor.All(ctx, &dataSets); err != nil {
		return nil, errors.Wrap(err, "unable to decode data sets for user by id")
	}

	if dataSets == nil {
		dataSets = []*upload.Upload{}
	}
	return dataSets, nil
}

func (d *DataRepository) GetDataSetByID(ctx context.Context, dataSetID string) (*upload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if dataSetID == "" {
		return nil, errors.New("data set id is missing")
	}

	now := time.Now()

	var dataSet *upload.Upload
	selector := bson.M{
		"uploadId": dataSetID,
		"type":     "upload",
	}
	err := d.FindOne(ctx, selector).Decode(&dataSet)

	loggerFields := log.Fields{"dataSetId": dataSetID, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("GetDataSetByID")

	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get data set by id")
	}

	return dataSet, nil
}

func (d *DataRepository) CreateDataSet(ctx context.Context, dataSet *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)

	dataSet.CreatedTime = pointer.FromString(timestamp)

	dataSet.ByUser = dataSet.CreatedUserID

	var err error
	if _, err = d.InsertOne(ctx, dataSet); storeStructuredMongo.IsDup(err) {
		err = errors.New("data set already exists")
	}

	loggerFields := log.Fields{"userId": dataSet.UserID, "dataSetId": dataSet.UploadID, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("CreateDataSet")

	if err != nil {
		return errors.Wrap(err, "unable to create data set")
	}
	return nil
}

func (d *DataRepository) UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*upload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	} else if !data.IsValidSetID(id) {
		return nil, errors.New("id is invalid")
	}
	if update == nil {
		return nil, errors.New("update is missing")
	} else if err := structureValidator.New().Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"id": id, "update": update})

	set := bson.M{
		"modifiedTime": timestamp,
	}
	unset := bson.M{}
	if update.Active != nil {
		set["_active"] = *update.Active
	}
	if update.DeviceID != nil {
		set["deviceId"] = *update.DeviceID
	}
	if update.DeviceModel != nil {
		set["deviceModel"] = *update.DeviceModel
	}
	if update.DeviceSerialNumber != nil {
		set["deviceSerialNumber"] = *update.DeviceSerialNumber
	}
	if update.Deduplicator != nil {
		set["_deduplicator"] = update.Deduplicator
	}
	if update.State != nil {
		set["_state"] = *update.State
	}
	if update.Time != nil {
		set["time"] = (*update.Time).Format(data.TimeFormat)
	}
	if update.TimeZoneName != nil {
		set["timezone"] = *update.TimeZoneName
	}
	if update.TimeZoneOffset != nil {
		set["timezoneOffset"] = *update.TimeZoneOffset
	}
	changeInfo, err := d.UpdateMany(ctx, bson.M{"type": "upload", "uploadId": id}, d.ConstructUpdate(set, unset))
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("UpdateDataSet")
	if err != nil {
		return nil, errors.Wrap(err, "unable to update data set")
	}

	return d.GetDataSetByID(ctx, id)
}

func (d *DataRepository) DeleteDataSet(ctx context.Context, dataSet *upload.Upload, doPurge bool) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)

	var err error
	var selector bson.M
	var removeInfo *mongo.DeleteResult
	var updateInfo *mongo.UpdateResult

	if doPurge {
		selector = bson.M{
			"_userId":  dataSet.UserID,
			"uploadId": dataSet.UploadID,
		}
	} else {
		selector = bson.M{
			"_userId":  dataSet.UserID,
			"uploadId": dataSet.UploadID,
			"type":     bson.M{"$ne": "upload"},
		}
	}

	removeInfo, err = d.DeleteMany(ctx, selector)
	if err == nil && doPurge == false {
		selector = bson.M{
			"_userId":       dataSet.UserID,
			"uploadId":      dataSet.UploadID,
			"type":          "upload",
			"deletedTime":   bson.M{"$exists": false},
			"deletedUserId": bson.M{"$exists": false},
		}
		set := bson.M{
			"deletedTime": timestamp,
		}
		unset := bson.M{}
		updateInfo, err = d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset))
	}

	loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "doPurge": doPurge, "removeInfo": removeInfo, "updateInfo": updateInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DeleteDataSet")

	if err != nil {
		return errors.Wrap(err, "unable to delete data set")
	}

	dataSet.SetDeletedTime(&timestamp)
	return nil
}

// Create data
func (d *DataRepository) CreateDataSetData(ctx context.Context, dataSet *upload.Upload, dataSetData []data.Datum) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	if dataSetData == nil {
		return errors.New("data set data is missing")
	}

	if len(dataSetData) == 0 {
		return nil
	}

	now := time.Now()
	creationTimestamp := now.Truncate(time.Millisecond)
	strTimestamp := creationTimestamp.Format(time.RFC3339Nano)

	var insertData []mongo.WriteModel
	var archiveData []mongo.WriteModel

	allSamples := make(map[string][]schema.ISample)
	var err error
	var incomingUserMetadata *schema.Metadata
	strUserId := *dataSet.UserID

	for _, datum := range dataSetData {
		datum.SetUserID(dataSet.UserID)
		datum.SetDataSetID(dataSet.UploadID)
		datum.SetCreatedTime(&strTimestamp)
		writeToArchive := d.isDatumToArchive(datum)
		writeToBucket := d.isDatumToBucket(datum)
		writeToLegacy := d.isDatumToLegacy(datum)
		guid := datum.GetGUID()
		deviceId := datum.GetDeviceID()
		/*If data type is in write to bucket ENV VAR, we write it to bucket*/
		if writeToBucket {
			// Prepare  to be pushed into data read db
			loggerFields := log.Fields{"datum": datum}
			switch event := datum.(type) {
			case *continuous.Continuous:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a cbg entry")
				// mapping
				var s = &schema.CbgSample{}
				s.Map(event)
				allSamples["Cbg"] = append(allSamples["Cbg"], *s)
			case *automated.Automated:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a automated basal entry")
				var s = &schema.BasalSample{}
				s.MapForAutomatedBasal(event)
				allSamples["Basal"] = append(allSamples["Basal"], *s)
			case *scheduled.Scheduled:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a scheduled basal entry")
				var s = &schema.BasalSample{}
				s.MapForScheduledBasal(event)
				allSamples["Basal"] = append(allSamples["Basal"], *s)
			case *temporary.Temporary:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a temp basal entry")
				var s = &schema.BasalSample{}
				s.MapForTempBasal(event)
				allSamples["Basal"] = append(allSamples["Basal"], *s)
			case *normal.Normal:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a bolus entry")
				var s = &schema.BolusSample{}
				s.MapForNormalBolus(event)
				log.LoggerFromContext(ctx).WithFields(log.Fields{"sample": s}).Debug("add normal bolus in bucket")
				allSamples["Bolus"] = append(allSamples["Bolus"], *s)
			case *biphasic.Biphasic:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a bolus entry")
				var s = &schema.BolusSample{}
				s.MapForBiphasicBolus(event)
				log.LoggerFromContext(ctx).WithFields(log.Fields{"sample": s}).Debug("add biphasic bolus in bucket")
				allSamples["Bolus"] = append(allSamples["Bolus"], *s)
			case *pen.Pen:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a pen bolus entry")
				var s = &schema.BolusSample{}
				s.MapForPenBolus(event)
				log.LoggerFromContext(ctx).WithFields(log.Fields{"sample": s}).Debug("add pen bolus in bucket")
				allSamples["Bolus"] = append(allSamples["Bolus"], *s)
			case *alarm.Alarm:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a alarm entry")
				var s = &schema.AlarmSample{}
				s.MapForAlarm(event)
				log.LoggerFromContext(ctx).WithFields(log.Fields{"sample": s}).Debug("add alarm in bucket")
				allSamples["Alarm"] = append(allSamples["Alarm"], *s)
			case *mode.Mode:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a mode entry")
				var s = &schema.Mode{}
				s.MapForMode(event)
				log.LoggerFromContext(ctx).WithFields(log.Fields{"sample": s}).Debug("add mode in bucket")
				if s.SubType == mode.LoopMode {
					// this condition allow us to store loop mode in a dedicated collection
					// because of the special processing with have with this data type
					// see tide-whisperer-v2 for more info
					allSamples["loopMode"] = append(allSamples["loopMode"], *s)
				} else {
					allSamples["Mode"] = append(allSamples["Mode"], *s)
				}
			case *calibration.Calibration:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a mode entry")
				var s = &schema.Calibration{}
				s.MapForCalibration(event)
				log.LoggerFromContext(ctx).WithFields(log.Fields{"sample": s}).Debug("add Calibration in bucket")
				allSamples["Calibration"] = append(allSamples["Calibration"], *s)
			case *flush.Flush:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a mode entry")
				var s = &schema.Flush{}
				s.MapForFlush(event)
				log.LoggerFromContext(ctx).WithFields(log.Fields{"sample": s}).Debug("add Flush in bucket")
				allSamples["Flush"] = append(allSamples["Flush"], *s)
			case *prime.Prime:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a Prime entry")
				var s = &schema.Prime{}
				s.MapForPrime(event)
				log.LoggerFromContext(ctx).WithFields(log.Fields{"sample": s}).Debug("add Prime in bucket")
				allSamples["Prime"] = append(allSamples["Prime"], *s)
			case *reservoirchange.ReservoirChange:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a ReservoirChange entry")
				var s = &schema.ReservoirChange{}
				s.MapForReservoirChange(event)
				log.LoggerFromContext(ctx).WithFields(log.Fields{"sample": s}).Debug("add ReservoirChange in bucket")
				allSamples["ReservoirChange"] = append(allSamples["ReservoirChange"], *s)
			case *calculator.Calculator:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a wizard entry")
				var s = &schema.Wizard{}
				s.MapForWizard(event)
				log.LoggerFromContext(ctx).WithFields(log.Fields{"sample": s}).Debug("add wizard in bucket")
				allSamples["Wizard"] = append(allSamples["Wizard"], *s)
			case *food.Food:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a food entry")
				var s = &schema.Food{}
				s.MapForFood(event)
				log.LoggerFromContext(ctx).WithFields(log.Fields{"sample": s}).Debug("add food in bucket")
				allSamples["Food"] = append(allSamples["Food"], *s)
			case *physical.Physical:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a PhysicalActivity entry")
				var s = &schema.PhysicalActivity{}
				s.MapForPhysical(event)
				log.LoggerFromContext(ctx).WithFields(log.Fields{"sample": s}).Debug("add PhysicalActivity in bucket")
				allSamples["PhysicalActivity"] = append(allSamples["PhysicalActivity"], *s)
			case *basalsecurity.BasalSecurity:
				log.LoggerFromContext(ctx).WithFields(loggerFields).Debug("add a SecurityBasal entry")
				var s = &schema.SecurityBasals{}
				s.MapForBasalSecurity(event)
				log.LoggerFromContext(ctx).WithFields(log.Fields{"sample": s}).Debug("add SecurityBasal in bucket")
				allSamples["SecurityBasal"] = append(allSamples["SecurityBasal"], *s)
			default:
				d.BucketStore.log.Infof("object ignored %+v", event)
			}
		}

		incomingUserMetadata = d.BucketStore.BuildUserMetadata(incomingUserMetadata, creationTimestamp, strUserId, datum.GetTime())

		var writeOp mongo.WriteModel
		if guid != nil && deviceId != nil {
			writeOp = mongo.NewReplaceOneModel().SetFilter(bson.M{"guid": *guid, "_userId": strUserId, "deviceId": *deviceId}).SetReplacement(datum).SetUpsert(true)
		} else {
			writeOp = mongo.NewInsertOneModel().SetDocument(datum)
		}

		/*If data type is in write to archive ENV VAR, we write it to the archive*/
		if writeToArchive {
			archiveData = append(archiveData, writeOp)
		}

		/*If data type is not in write to bucket ENV VAR, we write it to legacy deviceData*/
		/*We also write it if write to legacy is set alongside write to bucket ENV VAR*/
		if !writeToBucket || (writeToBucket && writeToLegacy) {
			insertData = append(insertData, writeOp)
		}
	}

	for dataType, samples := range allSamples {
		start := time.Now()
		err := d.BucketStore.UpsertMany(ctx, dataSet.UserID, creationTimestamp, samples, dataType)
		if err != nil {
			return errors.Wrapf(err, "unable to create %v bucket", dataType)
		}

		elapsedTime := time.Since(start).Seconds()
		dataWriteToReadStoreMetrics.WithLabelValues(dataType).Observe(float64(elapsedTime))
	}
	// update meta data
	if incomingUserMetadata != nil {
		err = d.BucketStore.UpsertMetaData(ctx, dataSet.UserID, incomingUserMetadata)
		if err != nil {
			return errors.Wrap(err, "unable to update metadata")
		}
	}

	opts := options.BulkWrite().SetOrdered(false)

	if len(insertData) > 0 {
		_, err = d.BulkWrite(ctx, insertData, opts)
		loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "dataCount": len(insertData), "duration": time.Since(now) / time.Microsecond}
		log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("CreateDataSetData")
		if err != nil {
			return errors.Wrap(err, "unable to create data set data")
		}
	}

	if len(archiveData) > 0 {
		_, err = d.ArchiveCollection.BulkWrite(ctx, archiveData, opts)
		loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "dataCount": len(archiveData), "duration": time.Since(now) / time.Microsecond}
		log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("CreateDataSetArchiveData")
		if err != nil {
			return errors.Wrap(err, "unable to create data set archive data")
		}
	}

	return nil
}

func (d *DataRepository) isDatumToArchive(datum data.Datum) bool {
	datumType := datum.GetType()
	for _, archivedType := range d.DataTypesArchived {
		if archivedType == datumType {
			return true
		}
	}
	return false
}

func (d *DataRepository) isDatumToBucket(datum data.Datum) bool {
	datumType := datum.GetType()
	for _, bucketedType := range d.DataTypesBucketed {
		if bucketedType == datumType {
			return true
		}
	}
	return false
}

func (d *DataRepository) isDatumToLegacy(datum data.Datum) bool {
	datumType := datum.GetType()
	for _, legacyType := range d.DataTypesKeptInLegacy {
		if legacyType == datumType {
			return true
		}
	}
	return false
}

func (d *DataRepository) ActivateDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, err := validateAndTranslateSelectors(selectors)
	if err != nil {
		return err
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["type"] = bson.M{"$ne": "upload"}
	selector["_active"] = false
	selector["deletedTime"] = bson.M{"$exists": false}
	set := bson.M{
		"_active":      true,
		"modifiedTime": timestamp,
	}
	unset := bson.M{
		"archivedDatasetId": 1,
		"archivedTime":      1,
		"modifiedUserId":    1,
	}
	changeInfo, err := d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to activate data set data")
		return errors.Wrap(err, "unable to activate data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("ActivateDataSetData")
	return nil
}

func (d *DataRepository) ArchiveDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, err := validateAndTranslateSelectors(selectors)
	if err != nil {
		return err
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["type"] = bson.M{"$ne": "upload"}
	selector["_active"] = true
	selector["deletedTime"] = bson.M{"$exists": false}
	set := bson.M{
		"_active":      false,
		"archivedTime": timestamp,
		"modifiedTime": timestamp,
	}
	unset := bson.M{
		"archivedDatasetId": 1,
		"modifiedUserId":    1,
	}
	changeInfo, err := d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to archive data set data")
		return errors.Wrap(err, "unable to archive data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("ArchiveDataSetData")
	return nil
}

func (d *DataRepository) DeleteDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, err := validateAndTranslateSelectors(selectors)
	if err != nil {
		return err
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["type"] = bson.M{"$ne": "upload"}
	selector["deletedTime"] = bson.M{"$exists": false}
	set := bson.M{
		"_active":      false,
		"archivedTime": timestamp,
		"deletedTime":  timestamp,
		"modifiedTime": timestamp,
	}
	unset := bson.M{
		"archivedDatasetId": 1,
		"deletedUserId":     1,
		"modifiedUserId":    1,
	}
	changeInfo, err := d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset))
	if err != nil {
		logger.WithError(err).Error("Unable to delete data set data")
		return errors.Wrap(err, "unable to delete data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DeleteDataSetData")
	return nil
}

func (d *DataRepository) DestroyDeletedDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, err := validateAndTranslateSelectors(selectors)
	if err != nil {
		return err
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["type"] = bson.M{"$ne": "upload"}
	selector["deletedTime"] = bson.M{"$exists": true}
	changeInfo, err := d.DeleteMany(ctx, selector)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy deleted data set data")
		return errors.Wrap(err, "unable to destroy deleted data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DestroyDeletedDataSetData")
	return nil
}

func (d *DataRepository) DestroyDataSetData(ctx context.Context, dataSet *upload.Upload, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	selector, err := validateAndTranslateSelectors(selectors)
	if err != nil {
		return err
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("dataSetId", *dataSet.UploadID)

	selector["_userId"] = dataSet.UserID
	selector["uploadId"] = dataSet.UploadID
	selector["type"] = bson.M{"$ne": "upload"}
	changeInfo, err := d.DeleteMany(ctx, selector)
	if err != nil {
		logger.WithError(err).Error("Unable to destroy data set data")
		return errors.Wrap(err, "unable to destroy data set data")
	}

	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).Debug("DestroyDataSetData")
	return nil
}

func (d *DataRepository) ArchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	if dataSet.DeviceID == nil || *dataSet.DeviceID == "" {
		return errors.New("data set device id is missing")
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)

	var updateInfo *mongo.UpdateResult

	selector := bson.M{
		"_userId":  dataSet.UserID,
		"uploadId": dataSet.UploadID,
		"type":     bson.M{"$ne": "upload"},
	}
	hashes, err := d.Distinct(ctx, "_deduplicator.hash", selector)
	if err == nil && len(hashes) > 0 {
		selector = bson.M{
			"_userId":            dataSet.UserID,
			"deviceId":           *dataSet.DeviceID,
			"type":               bson.M{"$ne": "upload"},
			"_active":            true,
			"_deduplicator.hash": bson.M{"$in": hashes},
		}
		set := bson.M{
			"_active":           false,
			"archivedDatasetId": dataSet.UploadID,
			"archivedTime":      timestamp,
			"modifiedTime":      timestamp,
		}
		unset := bson.M{}
		updateInfo, err = d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset))
	}

	loggerFields := log.Fields{"userId": dataSet.UserID, "deviceId": *dataSet.DeviceID, "updateInfo": updateInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("ArchiveDeviceDataUsingHashesFromDataSet")

	if err != nil {
		return errors.Wrap(err, "unable to archive device data using hashes from data set")
	}
	return nil
}

func (d *DataRepository) UnarchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, dataSet *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	if dataSet.DeviceID == nil || *dataSet.DeviceID == "" {
		return errors.New("data set device id is missing")
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"uploadId": dataSet.UploadID,
				"type":     bson.M{"$ne": "upload"},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"_active":           "$_active",
					"archivedDatasetId": "$archivedDatasetId",
					"archivedTime":      "$archivedTime",
				},
				"archivedHashes": bson.M{"$push": "$_deduplicator.hash"},
			},
		},
	}
	cursor, _ := d.Aggregate(ctx, pipeline)

	var overallUpdateInfo mongo.UpdateResult
	var overallErr error

	result := struct {
		ID struct {
			Active            bool   `bson:"_active"`
			ArchivedDataSetID string `bson:"archivedDatasetId"`
			ArchivedTime      string `bson:"archivedTime"`
		} `bson:"_id"`
		ArchivedHashes []string `bson:"archivedHashes"`
	}{}
	for cursor.Next(ctx) {
		err := cursor.Decode(&result)
		if err != nil {
			loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "result": result}
			log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Error("Unable to decode result for UnarchiveDeviceDataUsingHashesFromDataSet")
			if overallErr == nil {
				overallErr = errors.Wrap(err, "unable to decode device data results")
			}
		}
		if result.ID.Active != (result.ID.ArchivedDataSetID == "") || result.ID.Active != (result.ID.ArchivedTime == "") {
			loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "result": result}
			log.LoggerFromContext(ctx).WithFields(loggerFields).Error("Unexpected pipe result for UnarchiveDeviceDataUsingHashesFromDataSet")
			continue
		}

		selector := bson.M{
			"_userId":            dataSet.UserID,
			"deviceId":           dataSet.DeviceID,
			"archivedDatasetId":  dataSet.UploadID,
			"_deduplicator.hash": bson.M{"$in": result.ArchivedHashes},
		}
		set := bson.M{
			"_active":      result.ID.Active,
			"modifiedTime": timestamp,
		}
		unset := bson.M{}
		if result.ID.Active {
			unset["archivedDatasetId"] = true
			unset["archivedTime"] = true
		} else {
			set["archivedDatasetId"] = result.ID.ArchivedDataSetID
			set["archivedTime"] = result.ID.ArchivedTime
		}
		updateInfo, err := d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset))
		if err != nil {
			loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "result": result}
			log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Error("Unable to update result for UnarchiveDeviceDataUsingHashesFromDataSet")
			if overallErr == nil {
				overallErr = errors.Wrap(err, "unable to transfer device data active")
			}
		} else {
			overallUpdateInfo.ModifiedCount += updateInfo.ModifiedCount
		}
	}

	if err := cursor.Err(); err != nil {
		if overallErr == nil {
			overallErr = errors.Wrap(err, "unable to iterate to transfer device data active")
		}
	}

	loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "updateInfo": overallUpdateInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(overallErr).Debug("UnarchiveDeviceDataUsingHashesFromDataSet")

	return overallErr
}

func (d *DataRepository) DeleteOtherDataSetData(ctx context.Context, dataSet *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if err := validateDataSet(dataSet); err != nil {
		return err
	}
	if dataSet.DeviceID == nil || *dataSet.DeviceID == "" {
		return errors.New("data set device id is missing")
	}

	now := time.Now()
	timestamp := now.Truncate(time.Millisecond).Format(time.RFC3339Nano)

	var err error
	var removeInfo *mongo.DeleteResult
	var updateInfo *mongo.UpdateResult

	selector := bson.M{
		"_userId":  dataSet.UserID,
		"deviceId": *dataSet.DeviceID,
		"uploadId": bson.M{"$ne": dataSet.UploadID},
		"type":     bson.M{"$ne": "upload"},
	}
	removeInfo, err = d.DeleteMany(ctx, selector)
	if err == nil {
		selector = bson.M{
			"_userId":       dataSet.UserID,
			"deviceId":      *dataSet.DeviceID,
			"uploadId":      bson.M{"$ne": dataSet.UploadID},
			"type":          "upload",
			"deletedTime":   bson.M{"$exists": false},
			"deletedUserId": bson.M{"$exists": false},
		}
		set := bson.M{
			"deletedTime": timestamp,
		}
		unset := bson.M{}
		updateInfo, err = d.UpdateMany(ctx, selector, d.ConstructUpdate(set, unset))
	}

	loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "removeInfo": removeInfo, "updateInfo": updateInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DeleteOtherDataSetData")

	if err != nil {
		return errors.Wrap(err, "unable to remove other data set data")
	}
	return nil
}

func (d *DataRepository) DestroyDataForUserByID(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	now := time.Now()

	selector := bson.M{
		"_userId": userID,
	}
	removeInfo, err := d.DeleteMany(ctx, selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroyDataForUserByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy data for user by id")
	}

	return nil
}

func (d *DataRepository) ListUserDataSets(ctx context.Context, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = data.NewDataSetFilter()
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "filter": filter, "pagination": pagination})

	dataSets := data.DataSets{}
	selector := bson.M{
		"_active": true,
		"_userId": userID,
		"type":    "upload",
	}
	if filter.ClientName != nil {
		selector["client.name"] = *filter.ClientName
	}
	if filter.Deleted == nil || !*filter.Deleted {
		selector["deletedTime"] = bson.M{"$exists": false}
	}
	if filter.DeviceID != nil {
		selector["deviceId"] = *filter.DeviceID
	}
	if filter.State != nil {
		selector["_state"] = *filter.State
	}
	if filter.DataSetType != nil {
		selector["dataSetType"] = *filter.DataSetType
	}
	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": -1})
	cursor, err := d.Find(ctx, selector, opts)
	logger.WithFields(log.Fields{"count": len(dataSets), "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListUserDataSets")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list user data sets")
	}

	if err = cursor.All(ctx, &dataSets); err != nil {
		return nil, errors.Wrap(err, "unable to decode user data sets")
	}

	if dataSets == nil {
		dataSets = data.DataSets{}
	}

	return dataSets, nil
}

func (d *DataRepository) GetDataSet(ctx context.Context, id string) (*data.DataSet, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("id is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", id)

	var dataSet *data.DataSet
	selector := bson.M{
		"uploadId": id,
		"type":     "upload",
	}
	err := d.FindOne(ctx, selector).Decode(&dataSet)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("GetDataSet")
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get data set")
	}

	return dataSet, nil
}

func validateDataSet(dataSet *upload.Upload) error {
	if dataSet == nil {
		return errors.New("data set is missing")
	}
	if dataSet.UserID == nil {
		return errors.New("data set user id is missing")
	}
	if *dataSet.UserID == "" {
		return errors.New("data set user id is empty")
	}
	if dataSet.UploadID == nil {
		return errors.New("data set upload id is missing")
	}
	if *dataSet.UploadID == "" {
		return errors.New("data set upload id is empty")
	}
	return nil
}

func validateAndTranslateSelectors(selectors *data.Selectors) (bson.M, error) {
	if selectors == nil {
		return bson.M{}, nil
	} else if err := structureValidator.New().Validate(selectors); err != nil {
		return nil, errors.Wrap(err, "selectors is invalid")
	}

	var selectorIDs []string
	var selectorOriginIDs []string
	for _, selector := range *selectors {
		if selector != nil {
			if selector.ID != nil {
				selectorIDs = append(selectorIDs, *selector.ID)
			} else if selector.Origin != nil && selector.Origin.ID != nil {
				selectorOriginIDs = append(selectorOriginIDs, *selector.Origin.ID)
			}
		}
	}

	selector := bson.M{}
	if len(selectorIDs) > 0 && len(selectorOriginIDs) > 0 {
		selector["$or"] = []bson.M{
			{"id": bson.M{"$in": selectorIDs}},
			{"origin.id": bson.M{"$in": selectorOriginIDs}},
		}
	} else if len(selectorIDs) > 0 {
		selector["id"] = bson.M{"$in": selectorIDs}
	} else if len(selectorOriginIDs) > 0 {
		selector["origin.id"] = bson.M{"$in": selectorOriginIDs}
	}

	if len(selector) == 0 {
		return nil, errors.New("selectors is invalid")
	}

	return selector, nil
}
