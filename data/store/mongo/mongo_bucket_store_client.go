package mongo

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	goComMgo "github.com/mdblp/go-db/mongo"
	"go.mongodb.org/mongo-driver/bson"
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
	log                         *log.Logger
	minimalYearSupportedForData int
}

// Create a new bucket store client for a mongo DB if active is set to true, nil otherwise
func NewMongoBucketStoreClient(config *goComMgo.Config, logger *log.Logger, minimalYearSupportedForData int) (*MongoBucketStoreClient, error) {
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
	client.minimalYearSupportedForData = minimalYearSupportedForData
	return &client, err
}

/* bucket methods */

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
		switch dataType {
		case "Cbg":
			ops, _ := buildCbgUpdateOneModel(sample, userId, ts, creationTimestamp)
			operations = append(operations, ops...)
		case "Basal":
			ops, _ := buildBasalUpdateOneModel(sample, userId, ts, creationTimestamp)
			operations = append(operations, ops...)
		case "Bolus":
			ops, _ := buildBolusUpdateOneModel(sample, userId, ts, creationTimestamp)
			operations = append(operations, ops...)
		case "Alarm":
			ops, _ := buildAlarmUpdateOneModel(sample, userId, ts, creationTimestamp)
			operations = append(operations, ops...)
		case "Mode":
			ops, _ := buildModeUpdateOneModel(sample, userId, ts, creationTimestamp)
			operations = append(operations, ops...)
		case "loopMode":
			ops, _ := buildLoopModeWriteModel(sample, userId)
			operations = append(operations, ops...)
		case "Calibration":
			ops, _ := buildCalibrationUpdateOneModel(sample, userId, ts, creationTimestamp)
			operations = append(operations, ops...)
		case "Flush":
			ops, _ := buildFlushUpdateOneModel(sample, userId, ts, creationTimestamp)
			operations = append(operations, ops...)
		case "Prime":
			ops, _ := buildPrimeUpdateOneModel(sample, userId, ts, creationTimestamp)
			operations = append(operations, ops...)
		case "ReservoirChange":
			ops, _ := buildReservoirChangeUpdateOneModel(sample, userId, ts, creationTimestamp)
			operations = append(operations, ops...)
		}
	}

	// Specify an option to turn the bulk insertion with no order of operation
	bulkOption := options.BulkWriteOptions{}
	bulkOption.SetOrdered(false)

	if dataType == "loopMode" {
		// loop mode event is recorded with out hot/cold collection
		_, err := c.Collection("loopMode").BulkWrite(ctx, operations, &bulkOption)
		if err != nil {
			return err
		}
	} else {
		// update or insert in Hot Daily and Cold Daily
		for _, collectionPrefix := range dailyPrefixCollections {
			var collectionName string
			//TODO: to enhance
			if dataType == "Alarm" || dataType == "Mode" ||
				dataType == "Calibration" || dataType == "Flush" ||
				dataType == "Prime" || dataType == "ReservoirChange" {
				collectionName = collectionPrefix + "DeviceEvent"
			} else {
				collectionName = collectionPrefix + dataType
			}
			_, err := c.Collection(collectionName).BulkWrite(ctx, operations, &bulkOption)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func buildCbgUpdateOneModel(sample schema.ISample, userId *string, date string, creationTimestamp time.Time) ([]mongo.WriteModel, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, ErrUnableToParseBucketDayTime
	}

	strUserId := *userId
	var updates []mongo.WriteModel

	op := mongo.NewUpdateOneModel()
	op.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + date}})
	op.SetUpdate(bson.D{ // update
		{Key: "$addToSet", Value: bson.D{
			{Key: "samples", Value: sample}}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "creationTimestamp", Value: creationTimestamp},
			{Key: "day", Value: day},
			{Key: "userId", Value: strUserId}}},
	})
	op.SetUpsert(true)
	updates = append(updates, op)
	return updates, nil
}

func buildBasalUpdateOneModel(sample schema.ISample, userId *string, date string, creationTimestamp time.Time) ([]mongo.WriteModel, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, ErrUnableToParseBucketDayTime
	}

	strUserId := *userId
	var updates []mongo.WriteModel

	// Insert the bucket if not exist and then insert the sample in it
	basalFirstOp := mongo.NewUpdateOneModel()
	var array []schema.ISample
	basalFirstOp.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + date}})
	basalFirstOp.SetUpdate(bson.D{ // update
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "creationTimestamp", Value: creationTimestamp},
			{Key: "day", Value: day},
			{Key: "userId", Value: strUserId},
			{Key: "samples", Value: append(array, sample)},
		},
		},
	})
	basalFirstOp.SetUpsert(true)

	// Update the basal
	basalSecondOp := mongo.NewUpdateOneModel()
	elemfilter := sample.(schema.BasalSample)
	if elemfilter.Guid != "" {
		// All fields update based on guid
		basalSecondOp.SetFilter(bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "samples", Value: bson.D{
				{Key: "$elemMatch", Value: bson.D{
					{Key: "guid", Value: elemfilter.Guid},
				},
				},
			},
			},
		})
		basalSecondOp.SetUpdate(bson.D{ // update
			{Key: "$set", Value: bson.D{
				{Key: "samples.$.internalId", Value: elemfilter.InternalID},
				{Key: "samples.$.duration", Value: elemfilter.Duration},
				{Key: "samples.$.rate", Value: elemfilter.Rate},
				{Key: "samples.$.deliveryType", Value: elemfilter.DeliveryType},
				{Key: "samples.$.timestamp", Value: elemfilter.Timestamp},
			},
			},
		})
	} else {
		// Duration update based on rate/deliveryType/timestamp (nil guid)
		basalSecondOp.SetFilter(bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "samples", Value: bson.D{
				{Key: "$elemMatch", Value: bson.D{
					{Key: "guid", Value: nil},
					{Key: "rate", Value: elemfilter.Rate},
					{Key: "deliveryType", Value: elemfilter.DeliveryType},
					{Key: "timestamp", Value: elemfilter.Timestamp},
				},
				},
			},
			},
		})
		basalSecondOp.SetUpdate(bson.D{ // update
			{Key: "$set", Value: bson.D{
				{Key: "samples.$.internalId", Value: elemfilter.InternalID},
				{Key: "samples.$.duration", Value: elemfilter.Duration},
			},
			},
		})
	}

	// Otherwise we know that we did not update the basal so we guarantee an insertion
	// in the array
	basalThirdOp := mongo.NewUpdateOneModel()
	basalThirdOp.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + date}})
	basalThirdOp.SetUpdate(bson.D{ // update
		{Key: "$addToSet", Value: bson.D{
			{Key: "samples", Value: sample}}},
	})
	updates = append(updates, basalFirstOp, basalSecondOp, basalThirdOp)

	return updates, nil
}

func buildBolusUpdateOneModel(sample schema.ISample, userId *string, date string, creationTimestamp time.Time) ([]mongo.WriteModel, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, ErrUnableToParseBucketDayTime
	}

	strUserId := *userId
	var updates []mongo.WriteModel

	// Insert the bucket if not exist and then insert the sample in it
	bolusFirstOp := mongo.NewUpdateOneModel()
	var array []schema.ISample
	bolusFirstOp.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + date}})
	bolusFirstOp.SetUpdate(bson.D{ // update
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "creationTimestamp", Value: creationTimestamp},
			{Key: "day", Value: day},
			{Key: "userId", Value: strUserId},
			{Key: "samples", Value: append(array, sample)},
		},
		},
	})
	bolusFirstOp.SetUpsert(true)
	updates = append(updates, bolusFirstOp)

	// Update the bolus
	elemfilter := sample.(schema.BolusSample)
	if elemfilter.Guid != "" && elemfilter.DeviceId != "" {
		bolusSecondOp := mongo.NewUpdateOneModel()
		bolusSecondOp.SetFilter(bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "samples", Value: bson.D{
				{Key: "$elemMatch", Value: bson.D{
					{Key: "guid", Value: elemfilter.Guid},
					{Key: "deviceId", Value: elemfilter.DeviceId},
				},
				},
			},
			},
		})
		bolusSecondOp.SetUpdate(bson.D{ // update
			{Key: "$set", Value: bson.D{
				{Key: "samples.$.normal", Value: elemfilter.Normal},
				{Key: "samples.$.uuid", Value: elemfilter.Uuid},
			},
			},
		})
		updates = append(updates, bolusSecondOp)
	}
	// Otherwise we know that we did not update, so we guarantee an insertion
	// in the array
	bolusThirdOp := mongo.NewUpdateOneModel()
	bolusThirdOp.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + date}})
	bolusThirdOp.SetUpdate(bson.D{ // update
		{Key: "$addToSet", Value: bson.D{
			{Key: "samples", Value: sample}}},
	})
	updates = append(updates, bolusThirdOp)

	return updates, nil
}

func buildAlarmUpdateOneModel(sample schema.ISample, userId *string, date string, creationTimestamp time.Time) ([]mongo.WriteModel, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, ErrUnableToParseBucketDayTime
	}

	strUserId := *userId
	var updates []mongo.WriteModel

	// Insert the bucket if not exist and then insert the sample in it
	firstOp := mongo.NewUpdateOneModel()
	var array []schema.ISample
	firstOp.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + date}})
	firstOp.SetUpdate(bson.D{ // update
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "creationTimestamp", Value: creationTimestamp},
			{Key: "day", Value: day},
			{Key: "userId", Value: strUserId},
			{Key: "alarms", Value: append(array, sample)},
		},
		},
	})
	firstOp.SetUpsert(true)
	updates = append(updates, firstOp)

	// Update the bolus
	elemfilter := sample.(schema.AlarmSample)
	if elemfilter.Guid != "" && elemfilter.DeviceId != "" {
		secondOp := mongo.NewUpdateOneModel()
		secondOp.SetFilter(bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "alarms", Value: bson.D{
				{Key: "$elemMatch", Value: bson.D{
					{Key: "guid", Value: elemfilter.Guid},
					{Key: "deviceId", Value: elemfilter.DeviceId},
				},
				},
			},
			},
		})
		secondOp.SetUpdate(bson.D{ // update
			{Key: "$set", Value: bson.D{
				{Key: "alarms.$.level", Value: elemfilter.Level},
				{Key: "alarms.$.ackStatus", Value: elemfilter.AckStatus},
				{Key: "alarms.$.updateTimestamp", Value: elemfilter.UpdateTimestamp},
			},
			},
		})
		updates = append(updates, secondOp)
	}
	// Otherwise we know that we did not update, so we guarantee an insertion
	// in the array
	thirdOp := mongo.NewUpdateOneModel()
	thirdOp.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + date}})
	thirdOp.SetUpdate(bson.D{ // update
		{Key: "$addToSet", Value: bson.D{
			{Key: "alarms", Value: sample}}},
	})
	updates = append(updates, thirdOp)

	return updates, nil
}

func buildModeUpdateOneModel(sample schema.ISample, userId *string, date string, creationTimestamp time.Time) ([]mongo.WriteModel, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, ErrUnableToParseBucketDayTime
	}

	strUserId := *userId
	var updates []mongo.WriteModel

	// Insert the bucket if not exist and then insert the sample in it
	firstOp := mongo.NewUpdateOneModel()
	var array []schema.ISample
	firstOp.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + date}})
	firstOp.SetUpdate(bson.D{ // update
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "creationTimestamp", Value: creationTimestamp},
			{Key: "day", Value: day},
			{Key: "userId", Value: strUserId},
			{Key: "modes", Value: append(array, sample)},
		},
		},
	})
	firstOp.SetUpsert(true)
	updates = append(updates, firstOp)

	// Update the bolus
	elemfilter := sample.(schema.Mode)
	if elemfilter.Guid != "" && elemfilter.DeviceId != "" {
		secondOp := mongo.NewUpdateOneModel()
		secondOp.SetFilter(bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "modes", Value: bson.D{
				{Key: "$elemMatch", Value: bson.D{
					{Key: "guid", Value: elemfilter.Guid},
					{Key: "deviceId", Value: elemfilter.DeviceId},
				},
				},
			},
			},
		})
		secondOp.SetUpdate(bson.D{ // update
			{Key: "$set", Value: bson.D{
				{Key: "modes.$.duration", Value: elemfilter.Duration},
				{Key: "modes.$.inputTimestamp", Value: elemfilter.InputTimestamp},
			},
			},
		})
		updates = append(updates, secondOp)
	}
	// Otherwise we know that we did not update, so we guarantee an insertion
	// in the array
	thirdOp := mongo.NewUpdateOneModel()
	thirdOp.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + date}})
	thirdOp.SetUpdate(bson.D{ // update
		{Key: "$addToSet", Value: bson.D{
			{Key: "modes", Value: sample}}},
	})
	updates = append(updates, thirdOp)

	return updates, nil
}

func buildLoopModeWriteModel(sample schema.ISample, userId *string) ([]mongo.WriteModel, error) {

	strUserId := *userId

	elem := sample.(schema.Mode)
	//hack: a mapping here is required to add the user id in to the saved document
	// as the mode struct is a shared model with the bucket elements
	doc := bson.M{}
	doc["timestamp"] = elem.Timestamp
	doc["timezone"] = elem.Timezone
	doc["timezoneOffset"] = elem.TimezoneOffset
	doc["subType"] = elem.SubType
	doc["deviceId"] = elem.DeviceId
	doc["guid"] = elem.Guid
	doc["duration"] = elem.Duration
	doc["inputTimestamp"] = elem.InputTimestamp
	doc["userId"] = strUserId

	var updates []mongo.WriteModel
	var writeOp mongo.WriteModel
	if elem.Guid != "" && elem.DeviceId != "" {
		writeOp = mongo.NewReplaceOneModel().SetFilter(bson.M{"guid": elem.Guid, "userId": strUserId, "deviceId": elem.DeviceId}).SetReplacement(doc).SetUpsert(true)
	} else {
		writeOp = mongo.NewInsertOneModel().SetDocument(doc)
	}
	updates = append(updates, writeOp)
	return updates, nil
}

func buildCalibrationUpdateOneModel(sample schema.ISample, userId *string, date string, creationTimestamp time.Time) ([]mongo.WriteModel, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, ErrUnableToParseBucketDayTime
	}

	strUserId := *userId
	var updates []mongo.WriteModel

	op := mongo.NewUpdateOneModel()
	op.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + date}})
	op.SetUpdate(bson.D{ // update
		{Key: "$addToSet", Value: bson.D{
			{Key: "calibrations", Value: sample}}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "creationTimestamp", Value: creationTimestamp},
			{Key: "day", Value: day},
			{Key: "userId", Value: strUserId}}},
	})
	op.SetUpsert(true)
	updates = append(updates, op)

	return updates, nil
}

func buildFlushUpdateOneModel(sample schema.ISample, userId *string, date string, creationTimestamp time.Time) ([]mongo.WriteModel, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, ErrUnableToParseBucketDayTime
	}

	strUserId := *userId
	var updates []mongo.WriteModel

	op := mongo.NewUpdateOneModel()
	op.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + date}})
	op.SetUpdate(bson.D{ // update
		{Key: "$addToSet", Value: bson.D{
			{Key: "flushs", Value: sample}}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "creationTimestamp", Value: creationTimestamp},
			{Key: "day", Value: day},
			{Key: "userId", Value: strUserId}}},
	})
	op.SetUpsert(true)
	updates = append(updates, op)

	return updates, nil
}

func buildPrimeUpdateOneModel(sample schema.ISample, userId *string, date string, creationTimestamp time.Time) ([]mongo.WriteModel, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, ErrUnableToParseBucketDayTime
	}

	strUserId := *userId
	var updates []mongo.WriteModel

	op := mongo.NewUpdateOneModel()
	op.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + date}})
	op.SetUpdate(bson.D{ // update
		{Key: "$addToSet", Value: bson.D{
			{Key: "primes", Value: sample}}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "creationTimestamp", Value: creationTimestamp},
			{Key: "day", Value: day},
			{Key: "userId", Value: strUserId}}},
	})
	op.SetUpsert(true)
	updates = append(updates, op)

	return updates, nil
}

func buildReservoirChangeUpdateOneModel(sample schema.ISample, userId *string, date string, creationTimestamp time.Time) ([]mongo.WriteModel, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, ErrUnableToParseBucketDayTime
	}

	strUserId := *userId
	var updates []mongo.WriteModel

	// one operation because normally the event is sent once
	op := mongo.NewUpdateOneModel()
	op.SetFilter(bson.D{{Key: "_id", Value: strUserId + "_" + date}})
	op.SetUpdate(bson.D{ // update
		{Key: "$addToSet", Value: bson.D{
			{Key: "reservoirChanges", Value: sample}}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: strUserId + "_" + date},
			{Key: "creationTimestamp", Value: creationTimestamp},
			{Key: "day", Value: day},
			{Key: "userId", Value: strUserId}}},
	})
	op.SetUpsert(true)
	updates = append(updates, op)

	return updates, nil
}

// UpsertMetaData update or insert in MetaData
func (c *MongoBucketStoreClient) UpsertMetaData(ctx context.Context, userId *string, incomingUserMetadata *schema.Metadata) error {

	var dbUserMetadata *schema.Metadata
	var performUpdate bool

	opts := options.FindOne()
	if err := c.Collection("metadata").FindOne(ctx, bson.M{"userId": userId}, opts).Decode(&dbUserMetadata); err != nil && err != mongo.ErrNoDocuments {
		c.log.WithError(err)
		return err
	}

	dbUserMetadata, performUpdate = c.RefreshUserMetadata(dbUserMetadata, incomingUserMetadata)
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

func (c *MongoBucketStoreClient) RefreshUserMetadata(dbUserMetadata *schema.Metadata, incomingUserMetadata *schema.Metadata) (*schema.Metadata, bool) {
	if dbUserMetadata != nil {
		var performUpdate = false
		//Linked to YLP-1981, in some situation the DBLG1 is sending a data with a timestamp in the near 1970's ...
		//We do not want to update our metadata with this value. The CBG will be recorded with the 1970's date to keep
		//a trace of it, but it won't be displayed since the data is erroneous.
		if dbUserMetadata.OldestDataTimestamp.After(incomingUserMetadata.OldestDataTimestamp) && incomingUserMetadata.OldestDataTimestamp.Year() > c.minimalYearSupportedForData {
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
