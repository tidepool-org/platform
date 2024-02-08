package mongo

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
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

func (d *DataRepository) GetDataSet(ctx context.Context, dataSetID string) (*data.DataSet, error) {
	return d.DataSetRepository.GetDataSet(ctx, dataSetID)
}

func (d *DataRepository) GetDataSetByID(ctx context.Context, dataSetID string) (*upload.Upload, error) {
	return d.DataSetRepository.GetDataSetByID(ctx, dataSetID)
}

func (d *DataRepository) CreateDataSet(ctx context.Context, dataSet *upload.Upload) error {
	return d.DataSetRepository.createDataSet(ctx, dataSet, time.Now().UTC())
}

func (d *DataRepository) UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*upload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	return d.DataSetRepository.updateDataSet(ctx, id, update, time.Now().UTC())
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

// MergeSortedUploads combines the unique Uploads by UploadID into a new slice.
func MergeSortedUploads(newUploads, prevUploads []*upload.Upload) []*upload.Upload {
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
		// Always add the dataSet/upload in the "old" deviceData collection
		// because the dataSet/upload may not have been finished migrating
		// into the new deviceDataSets collection
		combined = append(combined, dataSet)
		// Skip duplicate of newUploads in prevUploads if it exists.
		if newCounter < len(newUploads) && *newUploads[newCounter].UploadID == *dataSet.UploadID {
			newCounter++
		}
	}
	combined = append(combined, newUploads[newCounter:]...)
	return combined
}

// MergeSortedDataSets combines the unique Uploads by UploadID into a new slice.
func MergeSortedDataSets(newDataSets, prevDataSets data.DataSets) data.DataSets {
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
		// Always add the dataSet/upload in the "old" deviceData collection
		// because the dataSet/upload may not have been finished migrating
		// into the new deviceDataSets collection
		combined = append(combined, dataSet)
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
