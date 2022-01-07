package mongo

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	goComMgo "github.com/mdblp/go-common/clients/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data/schema"
)

var ErrNoSamples = errors.New("impossible to bulk upsert an array of zero samples")
var ErrIncorrectTimestamp = errors.New("impossible to bulk upsert samples having a incorrect timestamp")
var ErrEmptyOrNilUserId = errors.New("impossible to upsert an array of sample for an empty or nil user id")
var ErrUnableToParseBucketDayTime = errors.New("unable to parse cbg day time")
var ErrInvalidDataType = errors.New("invalid empty data type")

var dailyPrefixCollections = []string{"coldDaily", "hotDaily"}

type MongoBucketStoreClient struct {
	*goComMgo.StoreClient
	log *log.Logger
}

// Create a new bucket store client for a mongo DB if active is set to true, nil otherwise
func NewMongoBucketStoreClient(config *goComMgo.Config, logger *log.Logger) (*MongoBucketStoreClient, error) {
	if config == nil {
		return nil, errors.New("bucket store mongo configuration is missing")
	}

	if logger == nil {
		return nil, errors.New("logger is missing for bucket store client")

	}

	client := MongoBucketStoreClient{}
	client.log = logger
	store, err := goComMgo.NewStoreClient(config, logger)
	client.StoreClient = store
	return &client, err
}

/* bucket methods */

// Look for a single bucket based on its Id
func (c *MongoBucketStoreClient) Find(ctx context.Context, bucket schema.IBucket) (result *schema.IBucket, err error) {

	if bucket.GetId() != "" {
		var query bson.M = bson.M{}
		query["_id"] = bucket.GetId()
		opts := options.FindOne()
		opts.SetSort(bson.D{primitive.E{Key: "_id", Value: -1}})
		if err = c.Collection("hotDailyCbg").FindOne(ctx, query, opts).Decode(&result); err != nil && err != mongo.ErrNoDocuments {
			c.log.WithError(err)
			return result, err
		}

		return result, nil
	}

	return nil, errors.New("Find called with an empty bucket.Id")
}

// Update a bucket record if found overwhise it will be created. The bucket is searched by its id.
func (c *MongoBucketStoreClient) Upsert(ctx context.Context, userId *string, creationTimestamp time.Time, sample schema.ISample, dataType string) error {

	if sample == nil {
		return errors.New("impossible to upsert a nil sample")
	}

	if sample.GetTimestamp().IsZero() {
		return errors.New("impossible to upsert a sample having a incorrect timestamp")
	}

	if userId == nil {
		return errors.New("impossible to upsert a sample for an empty user id")
	}

	// Extrat ISODate from sample timestamp
	ts := sample.GetTimestamp().Format("2006-01-02")
	day, err := time.Parse("2006-01-02", ts)
	if err != nil {
		return errors.New("unable to parse cbg day time")
	}
	valTrue := true
	strUserId := *userId

	c.log.Info("upsert cbg sample for: " + strUserId + "_" + ts)

	for _, collectionPrefix := range dailyPrefixCollections {
		collectionName := collectionPrefix + dataType
		_, err = c.Collection(collectionName).UpdateOne(
			ctx,
			bson.D{{Key: "_id", Value: strUserId + "_" + ts}}, // filter
			bson.D{ // update
				{Key: "$addToSet", Value: bson.D{
					{Key: "samples", Value: sample}}},
				{Key: "$setOnInsert", Value: bson.D{
					{Key: "_id", Value: strUserId + "_" + ts},
					{Key: "creationTimestamp", Value: creationTimestamp},
					{Key: "day", Value: day},
					{Key: "userId", Value: strUserId}}},
			},
			&options.UpdateOptions{Upsert: &valTrue}, //options
		)

		if err != nil {
			return err
		}
	}

	return err
}

// Perform a bulk of operations on bucket records based on the operation argument, update a record if found overwhise created it.
// The bucket is searched by its id.
func (c *MongoBucketStoreClient) UpsertMany(ctx context.Context, userId *string, creationTimestamp time.Time, samples []schema.ISample, dataType string) error {

	if userId == nil {
		return ErrEmptyOrNilUserId
	}

	if creationTimestamp.IsZero() {
		return ErrIncorrectTimestamp
	}

	if len(samples) == 0 {
		c.log.Debugf("no %v sample to write, nothing to add in bucket", dataType)
		return nil
	}

	if dataType == "" {
		return ErrInvalidDataType
	}

	var operations []mongo.WriteModel

	// transform as mongo operations
	// no data validation is done here as it is done in above layer in the Validate function
	for _, sample := range samples {
		ts := sample.GetTimestamp().Format("2006-01-02")
		//ts := "2021-12-22"

		day, err := time.Parse("2006-01-02", ts)
		if err != nil {
			return ErrUnableToParseBucketDayTime
		}

		strUserId := *userId
		operation := mongo.NewUpdateOneModel()
		operation.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + ts}})
		operation.SetUpdate(bson.D{ // update
			{Key: "$addToSet", Value: bson.D{
				{Key: "samples", Value: sample}}},
			{Key: "$setOnInsert", Value: bson.D{
				{Key: "_id", Value: strUserId + "_" + ts},
				{Key: "creationTimestamp", Value: creationTimestamp},
				{Key: "day", Value: day},
				{Key: "userId", Value: strUserId}}},
		})
		operation.SetUpsert(true)
		operations = append(operations, operation)

	}
	// Specify an option to turn the bulk insertion with no order of operation
	bulkOption := options.BulkWriteOptions{}
	bulkOption.SetOrdered(false)

	// update or insert in Hot Daily and Cold Daily
	for _, collectionPrefix := range dailyPrefixCollections {
		collectionName := collectionPrefix + dataType
		_, err := c.Collection(collectionName).BulkWrite(ctx, operations, &bulkOption)
		if err != nil {
			return err
		}
	}

	return nil
}

// update or insert in MetaData
func (c *MongoBucketStoreClient) UpsertMetaData(ctx context.Context, userId *string, incomingUserMetadata *schema.Metadata) error {

	var dbUserMetadata *schema.Metadata
	var performUpdate bool

	opts := options.FindOne()
	if err := c.Collection("metadata").FindOne(ctx, bson.M{"userId": userId}, opts).Decode(&dbUserMetadata); err != nil && err != mongo.ErrNoDocuments {
		c.log.WithError(err)
		return err
	}

	dbUserMetadata, performUpdate = c.refreshUserMetadata(dbUserMetadata, incomingUserMetadata)
	valTrue := true

	if performUpdate {
		c.log.Debug("perform update on metadata collection in data_read db")
		_, err := c.Collection("metadata").UpdateOne(ctx,
			bson.M{"userId": userId},
			bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "oldestDataTimestamp", Value: dbUserMetadata.OldestDataTimestamp},
					{Key: "newestDataTimestamp", Value: dbUserMetadata.NewestDataTimestamp}}},
				{Key: "$setOnInsert", Value: bson.D{
					{Key: "creationTimestamp", Value: dbUserMetadata.CreationTimestamp},
					{Key: "userId", Value: dbUserMetadata.UserId}}},
			},
			&options.UpdateOptions{Upsert: &valTrue},
		)
		return err
	}

	return nil
}

// Deletes a bucket record from the DB
func (c *MongoBucketStoreClient) Remove(ctx context.Context, bucket *schema.CbgBucket) error {

	if bucket.Id != "" {

		for _, collection := range dailyPrefixCollections {
			if _, err := c.Collection(collection).DeleteOne(ctx, bson.M{"_id": bucket.Id}); err != nil {
				return err
			}
		}
	}

	return errors.New("Remove called with an empty bucket.Id")
}

func (c *MongoBucketStoreClient) BuildUserMetadata(incomingUserMetadata *schema.Metadata, creationTimestamp time.Time, strUserId string, dataTimestamp time.Time) *schema.Metadata {
	if incomingUserMetadata == nil {
		incomingUserMetadata = &schema.Metadata{
			CreationTimestamp:   creationTimestamp,
			UserId:              strUserId,
			OldestDataTimestamp: dataTimestamp,
			NewestDataTimestamp: dataTimestamp,
		}
	} else {
		if incomingUserMetadata.OldestDataTimestamp.After(dataTimestamp) {
			incomingUserMetadata.OldestDataTimestamp = dataTimestamp
		} else if incomingUserMetadata.NewestDataTimestamp.Before(dataTimestamp) {
			incomingUserMetadata.NewestDataTimestamp = dataTimestamp
		}
	}
	return incomingUserMetadata
}

func (c *MongoBucketStoreClient) refreshUserMetadata(dbUserMetadata *schema.Metadata, incomingUserMetadata *schema.Metadata) (*schema.Metadata, bool) {
	if dbUserMetadata != nil {
		var performUpdate = false
		if dbUserMetadata.OldestDataTimestamp.After(incomingUserMetadata.OldestDataTimestamp) {
			c.log.WithField("oldestDataTimestamp", incomingUserMetadata.OldestDataTimestamp).Debug("set perform update to true and update OldestDataTimestamp db value")
			performUpdate = true
			dbUserMetadata.OldestDataTimestamp = incomingUserMetadata.OldestDataTimestamp
		}
		if dbUserMetadata.NewestDataTimestamp.Before(incomingUserMetadata.NewestDataTimestamp) {
			c.log.WithField("newestDataTimestamp", incomingUserMetadata.NewestDataTimestamp).Debug("set perform update to true and update NewestDataTimestamp db value")
			performUpdate = true
			dbUserMetadata.NewestDataTimestamp = incomingUserMetadata.NewestDataTimestamp
		}
		return dbUserMetadata, performUpdate
	} else {
		return incomingUserMetadata, true
	}
}
