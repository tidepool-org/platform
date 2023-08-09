package mongo

import (
	"context"
	"strings"
	"time"

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
		now := time.Now().UTC()
		if err := d.DatumRepository.createDataSet(sessCtx, dataSet, now); err != nil {
			return nil, err
		}
		return nil, d.DataSetRepository.createDataSet(sessCtx, dataSet, now)
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
		now := time.Now().UTC()
		if doc, err := d.DatumRepository.updateDataSet(sessCtx, id, update, now); err != nil {
			return nil, err
		} else if doc == nil {
			// if document doesn't exist in the old collection, then it
			// shouldn't exist in the new one either. Once migration is
			// complete, can delete this checking code and just use the
			// DataSetRepository.upsertDataSet (but changing it to be named
			// UpdateDataSet with no upsert option ) by itself.
			return nil, nil
		}
		return d.DataSetRepository.upsertDataSet(sessCtx, id, update, now)
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
