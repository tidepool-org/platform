package utils

import (
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func commonQuery(selector bson.M, userID *string, lastFetchedID *string) {
	if userID != nil {
		log.Printf("fetching for user %s", *userID)
		selector["_userId"] = *userID
	}
	idNotObjectID := bson.M{"$not": bson.M{"$type": "objectId"}}

	if lastFetchedID != nil {
		selector["$and"] = []interface{}{
			bson.M{"_id": bson.M{"$gt": *lastFetchedID}},
			bson.M{"_id": idNotObjectID},
		}
	} else {
		selector["_id"] = idNotObjectID
	}
}

func opt(batchSize int32, queryLimit int64) *options.FindOptions {
	return &options.FindOptions{
		Sort:      bson.M{"_id": 1},
		BatchSize: &batchSize,
		Limit:     &queryLimit,
	}
}

func JellyfishUpdatesQuery(userID *string, lastFetchedID *string, batchSize int64, queryLimit int64) (bson.M, *options.FindOptions) {
	selector := bson.M{
		"_deduplicator": bson.M{"$exists": false},
	}
	commonQuery(selector, userID, lastFetchedID)
	return selector, opt(int32(batchSize), queryLimit)
}

func JellyfishRollbackQuery(rollbackSectionName string, userID *string, lastFetchedID *string, batchSize int64, queryLimit int64) (bson.M, *options.FindOptions) {
	selector := bson.M{
		rollbackSectionName: bson.M{"$exists": true},
	}
	commonQuery(selector, userID, lastFetchedID)
	return selector, opt(int32(batchSize), queryLimit)
}

var ProcessJellyfishQueryFn = func(m Migration, selector bson.M, opts ...*options.FindOptions) bool {
	if dataC := m.GetDataCollection(); dataC != nil {
		dDataCursor, err := dataC.Find(m.GetCtx(), selector, opts...)
		if err != nil {
			log.Printf("failed to select data: %s", err)
			return false
		}
		defer dDataCursor.Close(m.GetCtx())

		all := []bson.M{}

		for dDataCursor.Next(m.GetCtx()) {
			item := bson.M{}
			if err := dDataCursor.Decode(&item); err != nil {
				log.Printf("error decoding data: %s", err)
				return false
			}
			itemID := fmt.Sprintf("%v", item["_id"])
			userID := fmt.Sprintf("%v", item["_userId"])
			itemType := fmt.Sprintf("%v", item["type"])
			updates, revert, err := ProcessDatum(itemID, itemType, item)
			if err != nil {
				m.OnError(ErrorData{Error: err, ItemID: itemID, ItemType: itemType})
			} else if len(updates) > 0 {
				m.SetUpdates(UpdateData{
					Filter:   bson.M{"_id": itemID, "modifiedTime": item["modifiedTime"]},
					ItemID:   itemID,
					UserID:   userID,
					ItemType: itemType,
					Apply:    updates,
					Revert:   revert,
				})
			}
			m.SetLastProcessed(itemID)
			all = append(all, item)
		}
		m.SetFetched(all)
		return len(all) > 0
	}
	return false
}

var WriteJellyfishUpdatesFn = func(m Migration) (int, error) {
	settings := m.GetSettings()
	updates := m.GetUpdates()
	dataC := m.GetDataCollection()
	if dataC == nil {
		return 0, errors.New("missing required collection to write updates to")
	}
	if len(m.GetUpdates()) == 0 {
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

		if err := m.UpdateChecks(); err != nil {
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
	m.ResetUpdates()
	return writtenCount, nil
}
