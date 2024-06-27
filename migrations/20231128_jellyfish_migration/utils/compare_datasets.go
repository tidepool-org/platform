package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/r3labs/diff/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func CompareDatasets(a []map[string]interface{}, b []map[string]interface{}) (map[string]string, error) {

	// cleanedA := []map[string]interface{}{}
	// cleanedB := []map[string]interface{}{}

	// doNotCompare := []string{
	// 	"_active",
	// 	"_archivedTime",
	// 	"_groupId",
	// 	"_id",
	// 	"id",
	// 	"_schemaVersion",
	// 	"_userId",
	// 	"_version",
	// 	"createdTime",
	// 	"guid",
	// 	"modifiedTime",
	// 	"uploadId",
	// 	"deduplicator",
	// 	"time",
	// }

	// for _, datum := range b {
	// 	for _, key := range doNotCompare {
	// 		delete(datum, key)
	// 	}
	// 	cleanedB = append(cleanedB, datum)
	// }

	// for _, datum := range a {
	// 	for _, key := range doNotCompare {
	// 		delete(datum, key)
	// 	}
	// 	cleanedA = append(cleanedA, datum)
	// }

	changelog, err := diff.Diff(a, b, diff.StructMapKeySupport(), diff.AllowTypeMismatch(true))
	if err != nil {
		return nil, err
	}

	differences := map[string]string{}
	for _, change := range changelog {
		differences[fmt.Sprintf("[%s] %s", change.Type, strings.Join(change.Path, "."))] = fmt.Sprintf("expected:[%v] actual:[%v]", change.From, change.To)
	}
	return differences, nil
}

func fetchDataSet(ctx context.Context, dataC *mongo.Collection, uploadID string) ([]map[string]interface{}, error) {
	if dataC == nil {
		return nil, errors.New("missing data collection")
	}

	dataset := []map[string]interface{}{}

	log.Printf("fetch dataset [%s]", uploadID)

	dDataCursor, err := dataC.Find(ctx, bson.M{
		"uploadId": uploadID,
	})
	if err != nil {
		return nil, err
	}
	defer dDataCursor.Close(ctx)

	if err := dDataCursor.All(ctx, &dataset); err != nil {
		return nil, err
	}
	log.Printf("got dataset [%s][%d] results", uploadID, len(dataset))
	return dataset, nil
}
