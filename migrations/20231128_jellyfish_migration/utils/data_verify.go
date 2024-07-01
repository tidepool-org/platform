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
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DataVerify struct {
	ctx   context.Context
	dataC *mongo.Collection
}

func CompareDatasets(setA []map[string]interface{}, setB []map[string]interface{}) ([]string, error) {

	batch := 100
	differences := []string{}

	var processBatch = func(batchA, batchB []map[string]interface{}) error {

		cleanedA := []map[string]interface{}{}
		cleanedB := []map[string]interface{}{}

		doNotCompare := []string{
			"_active",
			"_archivedTime",
			"_groupId",
			"_id",
			"id",
			"_schemaVersion",
			"_userId",
			"_version",
			"createdTime",
			"guid",
			"modifiedTime",
			"uploadId",
			"deduplicator",
			"_deduplicator",
			"time",
			"provenance", //provenance.byUserID
		}

		for _, datum := range batchA {
			for _, key := range doNotCompare {
				delete(datum, key)
			}
			cleanedB = append(cleanedB, datum)
		}

		for _, datum := range batchB {
			for _, key := range doNotCompare {
				delete(datum, key)
			}
			cleanedA = append(cleanedA, datum)
		}

		changelog, err := diff.Diff(cleanedA, cleanedB, diff.StructMapKeySupport(), diff.AllowTypeMismatch(true), diff.FlattenEmbeddedStructs(), diff.SliceOrdering(false))
		if err != nil {
			return err
		}

		for _, change := range changelog {
			differences = append(differences, fmt.Sprintf("[%s] %s => expected:[%v] actual:[%v]", change.Type, strings.Join(change.Path, "."), change.From, change.To))
		}
		return nil
	}

	for i := 0; i < len(setA); i += batch {
		j := i + batch
		if j > len(setA) {
			j = len(setA)
		}
		if err := processBatch(setA[i:j], setB[i:j]); err != nil {
			return nil, err
		}

	}
	return differences, nil

}

func NewVerifier(ctx context.Context, dataC *mongo.Collection) (*DataVerify, error) {

	if dataC == nil {
		return nil, errors.New("missing required data collection")
	}

	m := &DataVerify{
		ctx:   ctx,
		dataC: dataC,
	}

	return m, nil
}

func (m *DataVerify) fetchDataSet(uploadID string) ([]map[string]interface{}, error) {
	if m.dataC == nil {
		return nil, errors.New("missing data collection")
	}

	dataset := []map[string]interface{}{}

	dDataCursor, err := m.dataC.Find(m.ctx, bson.M{
		"uploadId": uploadID,
	}, &options.FindOptions{
		Sort: bson.M{"time": 1},
	})
	if err != nil {
		return nil, err
	}
	defer dDataCursor.Close(m.ctx)

	if err := dDataCursor.All(m.ctx, &dataset); err != nil {
		return nil, err
	}
	log.Printf("got dataset [%s][%d] results", uploadID, len(dataset))
	return dataset, nil
}

func (m *DataVerify) Verify(ref string, a string, b string) error {

	datasetA, err := m.fetchDataSet(a)
	if err != nil {
		return err
	}

	datasetB, err := m.fetchDataSet(b)
	if err != nil {
		return err
	}

	log.Printf("Compare [%s] vs [%s]", a, b)
	differences, err := CompareDatasets(datasetA, datasetB)
	if err != nil {
		return err
	}

	for _, v := range differences {
		log.Println(v)
	}

	return nil
}
