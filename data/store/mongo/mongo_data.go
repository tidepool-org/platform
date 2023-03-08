package mongo

import (
	"context"
	"strings"
	"time"

	baseDatum "github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/data/summary/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

// DataRepository implements the platform/data/store.DataRepository inteface.
// It mostly just uses embedding to forward the method calls, but implements
// a few methods that makes use of both repositories.
type DataRepository struct {
	*DatumRepository
	*DataSetRepository
}

func (d *DataRepository) EnsureIndexes() error {
	if err := d.DatumRepository.EnsureIndexes(); err != nil {
		return err
	}
	return d.DataSetRepository.EnsureIndexes()
}

func (d *DataRepository) GetDataSetsForUserByID(ctx context.Context, userID string, filter *store.Filter, pagination *page.Pagination) ([]*upload.Upload, error) {
	// Try reading from both new and old collections that hold dataSets,
	// starting with the new one. Can read only from the new deviceDataSets
	// collection via DataSetRepository when migration completed.
	newUploads, err := d.getDataSetsForUserByID(ctx, d.DataSetRepository.Repository, userID, filter, pagination)
	if err != nil {
		return nil, err
	}

	// Read from old deviceData collection for Uploads. Can delete this code
	// when migration is complete.
	prevUploads, err := d.getDataSetsForUserByID(ctx, d.DatumRepository.Repository, userID, filter, pagination)
	if err != nil {
		return nil, err
	}

	// Because there may be some dataSets in the old deviceData collection we
	// must read from both and merge the results while migration isn't
	// complete. Can delete this code when migration is complete.
	merged := mergeSortedUploads(newUploads, prevUploads)
	if pagination != nil && len(merged) > pagination.Size {
		merged = merged[:pagination.Size]
	}
	return merged, nil
}

func (d *DataRepository) getDataSetsForUserByID(ctx context.Context, repo *storeStructuredMongo.Repository, userID string, filter *store.Filter, pagination *page.Pagination) ([]*upload.Upload, error) {
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
	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": -1})
	cursor, err := repo.Find(ctx, selector, opts)

	loggerFields := log.Fields{"userId": userID, "dataSetsCount": len(dataSets), "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("getDataSetsForUserByID")

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

func (d *DataRepository) ListUserDataSets(ctx context.Context, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error) {
	// Try reading from both new and old collections that hold dataSets,
	// starting with the new one. Can read only from the new deviceDataSets
	// collection via DataSetRepository when migration completed.
	newDataSets, err := d.listUserDataSets(ctx, d.DataSetRepository.Repository, userID, filter, pagination)
	if err != nil {
		return nil, err
	}

	// Read from old deviceData collection for DataSets. Can delete this code
	// when migration is complete.
	prevDataSets, err := d.listUserDataSets(ctx, d.DatumRepository.Repository, userID, filter, pagination)
	if err != nil {
		return nil, err
	}

	// Because there may be some dataSets in the old deviceData collection we
	// must read from both and merge the results while migration isn't
	// complete. Can delete this code when migration is complete.
	merged := mergeSortedDataSets(newDataSets, prevDataSets)
	if pagination != nil && len(merged) > pagination.Size {
		merged = merged[:pagination.Size]
	}
	return merged, nil
}

func (d *DataRepository) listUserDataSets(ctx context.Context, repo *storeStructuredMongo.Repository, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error) {
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
	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": -1})
	cursor, err := repo.Find(ctx, selector, opts)
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
	// Try reading from both new and old collections that hold dataSets, starting with the new one.
	// Can read only from the new deviceDataSets collection via DataSetRepository when migration completed.
	dataSet, err := d.DataSetRepository.GetDataSet(ctx, id)
	if err != nil {
		return nil, err
	}
	if dataSet != nil {
		return dataSet, nil
	}
	return d.DatumRepository.GetDataSet(ctx, id)
}

func (d *DataRepository) GetDataSetByID(ctx context.Context, dataSetID string) (*upload.Upload, error) {
	// Try reading from both new and old collections that hold dataSets, starting with the new one.
	// Can read only from the new deviceDataSets collection via DataSetRepository when migration completed.
	dataSet, err := d.DataSetRepository.GetDataSetByID(ctx, dataSetID)
	if err != nil {
		return nil, err
	}
	if dataSet != nil {
		return dataSet, nil
	}

	return d.DatumRepository.GetDataSetByID(ctx, dataSetID)
}

func (d *DataRepository) CreateDataSet(ctx context.Context, dataSet *upload.Upload) error {
	// Until everything is migrated over to the new collection, some old
	// clients may still be reading from the old collection so we must write
	// to both old and new collection.
	steps := func(sessCtx mongo.SessionContext) (interface{}, error) {
		if err := d.DatumRepository.CreateDataSet(sessCtx, dataSet); err != nil {
			return nil, err
		}
		return nil, d.DataSetRepository.CreateDataSet(sessCtx, dataSet)
	}

	_, err := d.transact(ctx, steps)
	return err
}

func (d *DataRepository) transact(ctx context.Context, steps func(sessCtx mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	sess, err := d.mongoClient().StartSession()
	if err != nil {
		return nil, err
	}
	defer sess.EndSession(ctx)

	wc := writeconcern.New(writeconcern.WMajority(), writeconcern.J(true))
	rc := readconcern.Majority()
	txOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	return sess.WithTransaction(ctx, steps, txOpts)

}

func (d *DataRepository) UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*upload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	steps := func(sessCtx mongo.SessionContext) (interface{}, error) {
		if doc, err := d.DatumRepository.UpdateDataSet(sessCtx, id, update); err != nil {
			return nil, err
		} else if doc == nil {
			// if document doesn't exist in the old collection, then it
			// shouldn't exist in the new one either. Once migration is
			// complete, can delete this checking code and just use the
			// DataSetRepository.upsertDataSet (but changing it to be named
			// UpdateDataSet with no upsert option ) by itself.
			return nil, nil
		}
		return d.DataSetRepository.upsertDataSet(sessCtx, id, update)
	}

	dataSet, err := d.transact(ctx, steps)
	if err != nil {
		return nil, err
	}
	if dataSet == nil {
		return nil, nil
	}
	return dataSet.(*upload.Upload), nil
}

// DeleteDataSet will actually delete all non upload data and not actually
// delete the dataSet/upload but rather mark it as deleted by setting the
// deletedTime field.
func (d *DataRepository) DeleteDataSet(ctx context.Context, dataSet *upload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if err := validateDataSet(dataSet); err != nil {
		return err
	}

	now := time.Now().UTC()
	timestamp := now.Truncate(time.Millisecond)

	var err error
	var removeInfo *mongo.DeleteResult
	var updateInfoDeviceData *mongo.UpdateResult    // updating of DataSets in the old deviceData collection
	var updateInfoDeviceDataSet *mongo.UpdateResult // updating of DataSets in the new deviceDataSets collection

	selector := bson.M{
		"_userId":  dataSet.UserID,
		"uploadId": dataSet.UploadID,
		"type":     bson.M{"$ne": "upload"},
	}
	removeInfo, err = d.DatumRepository.DeleteMany(ctx, selector)
	if err == nil {
		selector = bson.M{
			"_userId":       dataSet.UserID,
			"uploadId":      dataSet.UploadID,
			"type":          "upload",
			"deletedTime":   bson.M{"$exists": false},
			"deletedUserId": bson.M{"$exists": false},
		}
		set := bson.M{
			"deletedTime":  timestamp,
			"modifiedTime": timestamp,
		}
		unset := bson.M{}

		var sessErr error
		steps := func(sessCtx mongo.SessionContext) (interface{}, error) {
			updateInfoDeviceDataSet, sessErr = d.DataSetRepository.UpdateMany(sessCtx, selector, d.DataSetRepository.ConstructUpdate(set, unset))
			if sessErr != nil {
				return nil, sessErr
			}

			updateInfoDeviceData, sessErr = d.DatumRepository.UpdateMany(sessCtx, selector, d.DataSetRepository.ConstructUpdate(set, unset))
			if sessErr != nil {
				return nil, sessErr
			}
			return nil, nil
		}

		_, err = d.transact(ctx, steps)
	}

	loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "removeInfo": removeInfo, "updateInfoDeviceData": updateInfoDeviceData, "updateInfoDeviceDataSet": updateInfoDeviceDataSet, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DeleteDataSet")

	if err != nil {
		return errors.Wrap(err, "unable to delete data set")
	}

	dataSet.SetDeletedTime(&timestamp)
	dataSet.SetModifiedTime(&timestamp)
	return nil
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

	now := time.Now().UTC()
	timestamp := now.Truncate(time.Millisecond)

	var err error
	var removeInfo *mongo.DeleteResult
	var updateInfoDeviceData *mongo.UpdateResult
	var updateInfoDeviceDataSet *mongo.UpdateResult

	selector := bson.M{
		"_userId":  dataSet.UserID,
		"deviceId": *dataSet.DeviceID,
		"uploadId": bson.M{"$ne": dataSet.UploadID},
		"type":     bson.M{"$ne": "upload"},
	}
	removeInfo, err = d.DatumRepository.DeleteMany(ctx, selector)
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
			// this upload's records has been deleted but we don't need to set the modifiedTime of the upload
			"deletedTime": timestamp,
		}
		unset := bson.M{}

		var sessErr error
		steps := func(sessCtx mongo.SessionContext) (interface{}, error) {
			updateInfoDeviceDataSet, sessErr = d.DataSetRepository.UpdateMany(sessCtx, selector, d.DataSetRepository.ConstructUpdate(set, unset))
			if sessErr != nil {
				return nil, sessErr
			}
			updateInfoDeviceData, sessErr = d.DatumRepository.UpdateMany(sessCtx, selector, d.DataSetRepository.ConstructUpdate(set, unset))
			if sessErr != nil {
				return nil, sessErr
			}
			return nil, nil
		}
		_, err = d.transact(ctx, steps)
	}

	loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "removeInfo": removeInfo, "updateInfoDeviceData": updateInfoDeviceData, "updateInfoDeviceDataSet": updateInfoDeviceDataSet, "duration": time.Since(now) / time.Microsecond}
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
	var removeDatumInfo *mongo.DeleteResult
	var removeDeviceDataSetInfo *mongo.DeleteResult
	var removeDataSetInfo *mongo.DeleteResult
	var err error

	removeDatumInfo, err = d.DatumRepository.DeleteMany(ctx, selector)
	if err == nil {
		var sessErr error
		steps := func(sessCtx mongo.SessionContext) (interface{}, error) {
			removeDeviceDataSetInfo, sessErr = d.DataSetRepository.DeleteMany(sessCtx, selector)
			if sessErr != nil {
				return nil, sessErr
			}

			removeDataSetInfo, sessErr = d.DatumRepository.DeleteMany(sessCtx, selector)
			if sessErr != nil {
				return nil, sessErr
			}

			return nil, nil
		}

		_, err = d.transact(ctx, steps)
	}
	loggerFields := log.Fields{"userId": userID, "removeDatumInfo": removeDatumInfo, "removeDataSetInfo": removeDataSetInfo, "removeDeviceDataSetInfo": removeDeviceDataSetInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroyDataForUserByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy data for user by id")
	}
	return nil
}

func (d *DataRepository) mongoClient() *mongo.Client {
	return d.DatumRepository.Database().Client()
}

// GetDataRange be careful when calling this, as if dataRecords isn't a pointer underneath, it will silently not
// result in any results being returned.
func (d *DataRepository) GetDataRange(ctx context.Context, dataRecords interface{}, userId string, typ string, startTime time.Time, endTime time.Time) error {
	if !isTypeUpload(typ) {
		return d.getDataRange(ctx, d.DatumRepository.Repository, dataRecords, userId, typ, startTime, endTime)
	}
	return nil // xxx temp
}

// getDataRange be careful when calling this, as if dataRecords isn't a pointer underneath, it will silently not
// result in any results being returned.
func (d *DataRepository) getDataRange(ctx context.Context, repo *storeStructuredMongo.Repository, dataRecords interface{}, userId string, typ string, startTime time.Time, endTime time.Time) error {

	// quit early if range is 0
	if startTime.Equal(endTime) {
		return nil
	}

	// return error if ranges are inverted, as this can produce unexpected results
	if startTime.After(endTime) {
		return errors.Newf("startTime (%s) after endTime (%s) for user %s", startTime, endTime, userId)
	}

	selector := bson.M{
		"_active": true,
		"_userId": userId,
		"type":    typ,
		"time": bson.M{"$gt": startTime,
			"$lte": endTime},
	}

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "time", Value: 1}})

	cursor, err := repo.Find(ctx, selector, opts)
	if err != nil {
		return errors.Wrap(err, "unable to get cgm data in date range for user")
	}

	if err = cursor.All(ctx, dataRecords); err != nil {
		return errors.Wrap(err, "unable to decode data sets")
	}

	return nil
}

func (d *DataRepository) GetLastUpdatedForUser(ctx context.Context, id string, typ string) (status *types.UserLastUpdated, err error) {
	if !isTypeUpload(typ) {
		return d.getLastUpdatedForUser(ctx, d.DatumRepository.Repository, id, typ)
	}

	// if typ is "upload", read from both deviceData and deviceDataSets
	// collection and get the more recent one as migration of uploads to
	// deviceDataSets happens.
	lastUpdatedDatum, err := d.getLastUpdatedForUser(ctx, d.DatumRepository.Repository, id, typ)
	if err != nil {
		return nil, err
	}
	lastUpdatedDataSet, err := d.getLastUpdatedForUser(ctx, d.DataSetRepository.Repository, id, typ)
	if err != nil {
		return nil, err
	}
	if lastUpdatedDatum == nil {
		return lastUpdatedDataSet, nil
	}
	if lastUpdatedDataSet == nil {
		return lastUpdatedDatum, nil
	}
	if lastUpdatedDatum.LastData.After(lastUpdatedDataSet.LastData) {
		return lastUpdatedDatum, nil
	}
	return lastUpdatedDataSet, nil
}

func (d *DataRepository) getLastUpdatedForUser(ctx context.Context, repo *storeStructuredMongo.Repository, id string, typ string) (*types.UserLastUpdated, error) {
	var err error
	var cursor *mongo.Cursor
	var status = &types.UserLastUpdated{}
	var dataSet []*baseDatum.Base

	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	if id == "" {
		return nil, errors.New("id is missing")
	}

	futureCutoff := time.Now().AddDate(0, 0, 1).UTC()
	pastCutoff := time.Now().AddDate(-2, 0, 0).UTC()

	selector := bson.M{
		"_active": true,
		"_userId": id,
		"type":    typ,
		"time": bson.M{"$lte": futureCutoff,
			"$gte": pastCutoff},
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "time", Value: -1}})
	findOptions.SetLimit(1)

	cursor, err = repo.Find(ctx, selector, findOptions)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get last %s date", typ)
	}

	if err = cursor.All(ctx, &dataSet); err != nil {
		return nil, errors.Wrapf(err, "unable to decode last %s date", typ)
	}

	// if we have no record
	if len(dataSet) < 1 {
		return status, nil
	}

	status.LastUpload = *dataSet[0].CreatedTime
	status.LastUpload = status.LastUpload.UTC()

	status.LastData = *dataSet[0].Time
	status.LastData = status.LastData.UTC()

	return status, nil
}

func (d *DataRepository) DistinctUserIDs(ctx context.Context, typ string) ([]string, error) {
	if !isTypeUpload(typ) {
		return d.distinctUserIDs(ctx, d.DatumRepository.Repository, typ)
	}

	// If type is upload read from both deviceData and deviceDataSets while
	// migration of uploads from deviceData to deviceDataSets is happening.
	datumIDs, err := d.distinctUserIDs(ctx, d.DatumRepository.Repository, typ)
	if err != nil {
		return nil, err
	}
	dataSetIDs, err := d.distinctUserIDs(ctx, d.DataSetRepository.Repository, typ)
	if err != nil {
		return nil, err
	}

	distinctUserIDs := make(map[string]struct{})
	for _, userID := range datumIDs {
		distinctUserIDs[userID] = struct{}{}
	}
	for _, userID := range dataSetIDs {
		distinctUserIDs[userID] = struct{}{}
	}
	userIDs := make([]string, 0, len(distinctUserIDs))
	for userID := range distinctUserIDs {
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

func (d *DataRepository) distinctUserIDs(ctx context.Context, repo *storeStructuredMongo.Repository, typ string) ([]string, error) {
	var distinctUserIDMap = make(map[string]struct{})
	var empty struct{}

	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	// allow for a small margin on the pastCutoff to allow for calculation delay
	pastCutoff := time.Now().AddDate(0, -23, -20).UTC()
	futureCutoff := time.Now().AddDate(0, 0, 1).UTC()

	selector := bson.M{
		"_userId": bson.M{"$ne": -1111},
		"_active": true,
		"type":    typ,
		"time":    bson.M{"$gte": pastCutoff, "$lte": futureCutoff},
	}

	result, err := repo.Distinct(ctx, "_userId", selector)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching distinct userIDs")
	}

	for _, v := range result {
		distinctUserIDMap[v.(string)] = empty
	}

	userIDs := make([]string, 0, len(distinctUserIDMap))
	for k := range distinctUserIDMap {
		userIDs = append(userIDs, k)
	}

	return userIDs, nil
}

// repo gets the correct repo for data depending on its type.
func (d *DataRepository) repo(typ string) *storeStructuredMongo.Repository {
	if strings.ToLower(typ) == strings.ToLower(upload.Type) {
		return d.DataSetRepository.Repository
	}
	return d.DatumRepository.Repository
}

// mergeSortedUploads combines the unique Uploads by UploadID into a new slice.
func mergeSortedUploads(newUploads, prevUploads []*upload.Upload) []*upload.Upload {
	combined := make([]*upload.Upload, 0, len(newUploads)+len(prevUploads))

	// Merge the two datasets like the merge step in merge sort. Note we don't
	// use sort.Slice/sort.SliceStable from the standard library as the
	// sorting criteria may change (?) in the Repositories in the future.
	newCounter := 0
	// Prefer the uploads in prevUploads as that will maintain proper
	// Pagination in the case that not all records are in the new collection
	// yet because all existing uploads are already in the old collection but
	// might not be in the new one.
	for _, dataSet := range prevUploads {
		for newCounter < len(newUploads) && *newUploads[newCounter].UploadID < *dataSet.UploadID {
			combined = append(combined, newUploads[newCounter])
			newCounter++
		}
		combined = append(combined, prevUploads[newCounter])
		// Skip duplicate of newUploads in prevUploads if it exists.
		if newCounter < len(newUploads) && *newUploads[newCounter].UploadID == *dataSet.UploadID {
			newCounter++
		}
	}
	combined = append(combined, newUploads[newCounter:]...)
	return combined
}

// mergeSortedDataSets combines the unique Uploads by UploadID into a new slice.
func mergeSortedDataSets(newDataSets, prevDataSets data.DataSets) data.DataSets {
	combined := make(data.DataSets, 0, len(newDataSets)+len(prevDataSets))

	// Merge the two datasets like the merge step in merge sort. Note we don't
	// use sort.Slice/sort.SliceStable from the standard library as the
	// sorting criteria may change (?) in the Repositories in the future.
	newCounter := 0
	// Prefer the dataSets in prevDataSets as that will maintain proper
	// Pagination in the case that not all records are in the new collection
	// yet because all existing DataSets are already in the old collection but
	// might not be in the new one.
	for _, dataSet := range prevDataSets {
		for newCounter < len(newDataSets) && *newDataSets[newCounter].UploadID < *dataSet.UploadID {
			combined = append(combined, newDataSets[newCounter])
			newCounter++
		}
		combined = append(combined, prevDataSets[newCounter])
		// Skip duplicate of newDataSets in prevDataSets if it exists.
		if newCounter < len(newDataSets) && *newDataSets[newCounter].UploadID == *dataSet.UploadID {
			newCounter++
		}
	}
	combined = append(combined, newDataSets[newCounter:]...)
	return combined
}

func isTypeUpload(typ string) bool {
	return strings.ToLower(typ) == strings.ToLower(upload.Type)
}
