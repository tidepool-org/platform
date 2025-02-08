package mongo

import (
	"context"
	stdErrs "errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type DataSetRepository struct {
	*storeStructuredMongo.Repository
}

func (d *DataSetRepository) EnsureIndexes() error {
	// Note "type" field isn't really needed because datasets/uploads are
	// always type == "upload" but this is just to keep the original queries
	// untouched.
	return d.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		// Additional indexes are also created in `tide-whisperer` and `jellyfish`
		{
			Keys: bson.D{
				{Key: "_userId", Value: 1},
				{Key: "_active", Value: 1},
				{Key: "type", Value: 1},
				{Key: "time", Value: -1},
			},
			Options: options.Index().
				SetName("UserIdTypeWeighted_v2"),
		},
		{
			Keys: bson.D{
				{Key: "origin.id", Value: 1},
				{Key: "type", Value: 1},
				{Key: "deletedTime", Value: -1},
				{Key: "_active", Value: 1},
			},
			Options: options.Index().
				SetName("OriginId"),
		},
		{
			Keys: bson.D{
				{Key: "uploadId", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("UniqueUploadId"),
		},
		{
			Keys: bson.D{
				{Key: "uploadId", Value: 1},
				{Key: "type", Value: 1},
				{Key: "deletedTime", Value: -1},
				{Key: "_active", Value: 1},
			},
			Options: options.Index().
				SetName("UploadId"),
		},
		{
			Keys: bson.D{
				{Key: "_userId", Value: 1},
				{Key: "client.name", Value: 1},
				{Key: "type", Value: 1},
				{Key: "createdTime", Value: -1},
			},
			Options: options.Index().
				SetName("ListUserDataSets").
				SetPartialFilterExpression(bson.D{
					{Key: "_active", Value: true},
				}),
		},
		{
			Keys: bson.D{
				{Key: "_userId", Value: 1},
				{Key: "deviceId", Value: 1},
				{Key: "type", Value: 1},
				{Key: "createdTime", Value: -1},
			},
			Options: options.Index().
				SetName("ListUserDataSetsDeviceId").
				SetPartialFilterExpression(bson.D{
					{Key: "_active", Value: true},
				}),
		},
	})
}

func (d *DataSetRepository) GetDataSetByID(ctx context.Context, dataSetID string) (*data.DataSet, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if dataSetID == "" {
		return nil, errors.New("data set id is missing")
	}

	now := time.Now().UTC()

	var dataSet *data.DataSet
	selector := bson.M{
		"uploadId": dataSetID,
	}
	err := d.FindOne(ctx, selector).Decode(&dataSet)

	loggerFields := log.Fields{"dataSetId": dataSetID, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DataSet.GetDataSetByID")

	if stdErrs.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get data set by id")
	}

	return dataSet, nil
}

func (d *DataSetRepository) createDataSet(ctx context.Context, dataSet *data.DataSet, now time.Time) (*data.DataSet, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return nil, err
	}

	now = now.UTC()
	timestamp := now.Truncate(time.Millisecond)

	dataSet.CreatedTime = pointer.FromTime(timestamp)
	dataSet.ModifiedTime = pointer.FromTime(timestamp)

	dataSet.ByUser = dataSet.CreatedUserID

	var err error
	if _, err = d.InsertOne(ctx, dataSet); storeStructuredMongo.IsDup(err) {
		err = errors.New("data set already exists")
	}

	loggerFields := log.Fields{"userId": dataSet.UserID, "dataSetId": dataSet.UploadID, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DataSet.CreateDataSet")

	if err != nil {
		return nil, errors.Wrap(err, "unable to create data set")
	}

	return d.GetDataSetByID(ctx, *dataSet.ID)
}

func (d *DataSetRepository) updateDataSet(ctx context.Context, id string, update *data.DataSetUpdate, now time.Time) (*data.DataSet, error) {
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
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(update); err != nil {
		return nil, errors.Wrap(err, "update is invalid")
	}

	now = now.UTC()
	timestamp := now.Truncate(time.Millisecond)
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
		set["time"] = *update.Time
	}
	if update.TimeZoneName != nil {
		set["timezone"] = *update.TimeZoneName
	}
	if update.TimeZoneOffset != nil {
		set["timezoneOffset"] = *update.TimeZoneOffset
	}
	opts := options.Update().SetUpsert(false) // Can only update an existing document
	changeInfo, err := d.UpdateMany(ctx, bson.M{"uploadId": id}, d.ConstructUpdate(set, unset), opts)
	logger.WithFields(log.Fields{"changeInfo": changeInfo, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DataSetRepository.UpdateDataSet")
	if err != nil {
		return nil, errors.Wrap(err, "unable to upsert data set")
	}

	return d.GetDataSetByID(ctx, id)
}

func (d *DataSetRepository) GetDataSet(ctx context.Context, dataSetID string) (*data.DataSet, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if dataSetID == "" {
		return nil, errors.New("id is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithField("id", dataSetID)

	var dataSet *data.DataSet
	selector := bson.M{
		"uploadId": dataSetID,
	}

	err := d.FindOne(ctx, selector).Decode(&dataSet)
	logger.WithField("duration", time.Since(now)/time.Microsecond).WithError(err).Debug("DataSet.GetDataSet")
	if stdErrs.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get data set")
	}

	return dataSet, nil
}

func (d *DataSetRepository) ListUserDataSets(ctx context.Context, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = data.NewDataSetFilter()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(pagination); err != nil {
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

func (d *DataSetRepository) CreateUserDataSet(ctx context.Context, userID string, create *data.DataSetCreate) (*data.DataSet, error) {
	dataSet := data.NewDataSet()
	dataSet.Client = create.Client
	dataSet.DataSetType = create.DataSetType
	dataSet.Deduplicator = create.Deduplicator
	dataSet.DeviceID = create.DeviceID
	dataSet.DeviceManufacturers = create.DeviceManufacturers
	dataSet.DeviceModel = create.DeviceModel
	dataSet.DeviceSerialNumber = create.DeviceSerialNumber
	dataSet.DeviceTags = create.DeviceTags
	dataSet.Time = create.Time
	dataSet.TimeProcessing = create.TimeProcessing
	dataSet.TimeZoneName = create.TimeZoneName
	dataSet.TimeZoneOffset = create.TimeZoneOffset

	dataSet.DataState = pointer.FromString("open") // TODO: Deprecated DataState (after data migration)
	dataSet.ID = pointer.FromString(data.NewID())
	dataSet.State = pointer.FromString("open")
	dataSet.UserID = pointer.FromString(userID)
	dataSet.UploadID = dataSet.ID

	return d.createDataSet(ctx, dataSet, time.Now().UTC())
}

func (d *DataSetRepository) GetDataSetsForUserByID(ctx context.Context, userID string, filter *store.Filter, pagination *page.Pagination) ([]*data.DataSet, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if filter == nil {
		filter = store.NewFilter()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	now := time.Now()

	var dataSets []*data.DataSet
	selector := bson.M{
		"_active": true,
		"_userId": userID,
		"type":    "upload",
	}
	if !filter.Deleted {
		selector["deletedTime"] = bson.M{"$exists": false}
	}
	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": -1})
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
		dataSets = []*data.DataSet{}
	}
	return dataSets, nil
}

func validateDataSet(dataSet *data.DataSet) error {
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
