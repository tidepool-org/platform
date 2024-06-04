package utils

import (
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	"github.com/tidepool-org/platform/pointer"
)

func jellyfishQuery(settings Settings, userID *string, lastFetchedID *string) (bson.M, *options.FindOptions) {
	selector := bson.M{
		"_deduplicator": bson.M{"$exists": false},
	}

	if userID != nil && *userID != "" {
		log.Printf("fetching for user %s", *userID)
		selector["_userId"] = *userID
	}
	idNotObjectID := bson.M{"$not": bson.M{"$type": "objectId"}}

	if lastFetchedID != nil && *lastFetchedID != "" {
		selector["$and"] = []interface{}{
			bson.M{"_id": bson.M{"$gt": *lastFetchedID}},
			bson.M{"_id": idNotObjectID},
		}
	} else {
		selector["_id"] = idNotObjectID
	}

	bSize := int32(settings.QueryBatchSize)
	limit := int64(settings.QueryBatchLimit)
	opts := &options.FindOptions{
		Sort:      bson.M{"_id": 1},
		BatchSize: &bSize,
		Limit:     &limit,
	}

	return selector, opts
}

var JellyfishDataQueryFn = func(m *DataMigration) bool {

	settings := m.GetSettings()

	if dataC := m.GetDataCollection(); dataC != nil {

		selector, opts := jellyfishQuery(
			settings,
			nil,
			pointer.FromString(m.GetLastID()),
		)

		opts.Projection = bson.M{"_id": 1}
		dDataCursor, err := dataC.Find(m.GetCtx(), selector, opts)
		if err != nil {
			log.Printf("failed to select data: %s", err)
			return false
		}
		defer dDataCursor.Close(m.GetCtx())

		count := 0

		for dDataCursor.Next(m.GetCtx()) {
			m.UpdateFetchedCount()
			var result struct {
				ID string `bson:"_id"`
			}
			if err := dDataCursor.Decode(&result); err != nil {
				m.OnError(ErrorData{Error: err})
				continue
			} else {
				setDeduplicator := bson.M{"$set": bson.M{"_deduplicator": bson.M{"hash": result.ID}}}
				m.SetUpdates(UpdateData{
					Filter: bson.M{"_id": result.ID},
					ItemID: result.ID,
					Apply:  []bson.M{setDeduplicator},
				})
				count++
				m.SetLastProcessed(result.ID)
			}
		}
		return count > 0
	}
	return false
}

var JellyfishDataUpdatesFn = func(m *DataMigration) (int, error) {
	settings := m.GetSettings()
	updates := m.GetUpdates()
	dataC := m.GetDataCollection()
	if dataC == nil {
		return 0, errors.New("missing required collection to write updates to")
	}
	if len(updates) == 0 {
		return 0, nil
	}

	var getBatches = func(chunkSize int) [][]mongo.WriteModel {
		batches := [][]mongo.WriteModel{}
		for i := 0; i < len(updates); i += chunkSize {
			end := i + chunkSize
			if end > len(updates) {
				end = len(updates)
			}
			batches = append(batches, updates[i:end])
		}
		return batches
	}
	writtenCount := 0
	for _, batch := range getBatches(int(*settings.WriteBatchSize)) {
		if err := m.CheckMongoInstance(); err != nil {
			return writtenCount, err
		}
		if settings.DryRun {
			writtenCount += len(batch)
			continue
		}
		results, err := dataC.BulkWrite(m.GetCtx(), batch)
		if err != nil {
			log.Printf("error writing batch updates %v", err)
			return writtenCount, err
		}
		writtenCount += int(results.ModifiedCount)
	}
	return writtenCount, nil
}

var JellyfishUploadQueryFn = func(m *DataMigration) bool {

	settings := m.GetSettings()

	if dataC := m.GetDataCollection(); dataC != nil {

		selector, opts := jellyfishQuery(
			settings,
			nil,
			pointer.FromString(m.GetLastID()),
		)

		opts.Projection = bson.M{"_id": 1}
		dDataCursor, err := dataC.Find(m.GetCtx(), selector, opts)
		if err != nil {
			log.Printf("failed to select upload data: %s", err)
			return false
		}
		defer dDataCursor.Close(m.GetCtx())

		count := 0

		for dDataCursor.Next(m.GetCtx()) {
			m.UpdateFetchedCount()
			var result struct {
				ID string `bson:"_id"`
			}
			if err := dDataCursor.Decode(&result); err != nil {
				m.OnError(ErrorData{Error: err})
				continue
			} else {
				m.SetUpdates(UpdateData{
					Filter: bson.M{"_id": result.ID},
					ItemID: result.ID,
					Apply:  []bson.M{{"$set": bson.M{"_deduplicator": bson.M{"name": deduplicator.DeviceDeactivateLegacyHashName, "version": "0.0.0"}}}},
				})
				count++
				m.SetLastProcessed(result.ID)
			}
		}
		return count > 0
	}
	return false
}
