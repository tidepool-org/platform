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

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
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
	return d.DataSetRepository.GetDataSetsForUserByID(ctx, userID, filter, pagination)
}

func (d *DataRepository) ListUserDataSets(ctx context.Context, userID string, filter *data.DataSetFilter, pagination *page.Pagination) (data.DataSets, error) {
	return d.DataSetRepository.ListUserDataSets(ctx, userID, filter, pagination)
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
	return d.DataSetRepository.GetDataSetByID(ctx, dataSetID)
}

func (d *DataRepository) CreateDataSet(ctx context.Context, dataSet *upload.Upload) error {
	return d.DataSetRepository.CreateDataSet(ctx, dataSet)
}

func (d *DataRepository) UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*upload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	return d.DataSetRepository.UpdateDataSet(ctx, id, update)
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
	var updateInfo *mongo.UpdateResult // updating of DataSets in the new deviceDataSets collection

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
		// Note setting updateInfo and err as defined above
		updateInfo, err = d.DataSetRepository.UpdateMany(ctx, selector, d.DataSetRepository.ConstructUpdate(set, unset))
	}

	loggerFields := log.Fields{"dataSetId": dataSet.UploadID, "removeInfo": removeInfo, "updateInfo": updateInfo, "duration": time.Since(now) / time.Microsecond}
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
	var updateInfo *mongo.UpdateResult

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
		updateInfo, err = d.DataSetRepository.UpdateMany(ctx, selector, d.DataSetRepository.ConstructUpdate(set, unset))
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
	var removeDatumInfo *mongo.DeleteResult
	var removeDeviceDataSetInfo *mongo.DeleteResult
	var err error

	removeDatumInfo, err = d.DatumRepository.DeleteMany(ctx, selector)
	if err == nil {
		removeDeviceDataSetInfo, err = d.DataSetRepository.DeleteMany(ctx, selector)
	}
	loggerFields := log.Fields{"userId": userID, "removeDatumInfo": removeDatumInfo, "removeDeviceDataSetInfo": removeDeviceDataSetInfo, "duration": time.Since(now) / time.Microsecond}
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
	return d.getDataRange(ctx, d.repo(typ), dataRecords, userId, typ, startTime, endTime)
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
	return d.getLastUpdatedForUser(ctx, d.repo(typ), id, typ)
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
	return d.distinctUserIDs(ctx, d.repo(typ), typ)
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
