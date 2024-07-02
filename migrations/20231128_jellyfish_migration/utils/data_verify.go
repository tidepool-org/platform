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

func CompareDatasets(platformData []map[string]interface{}, jellyfishData []map[string]interface{}) ([]string, error) {

	batch := 100
	differences := []string{}

	var processBatch = func(batchPlatform, batchJellyfish []map[string]interface{}) error {

		cleanedJellyfish := []map[string]interface{}{}
		cleanedPlatform := []map[string]interface{}{}

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
			"payload",
			"time",
			"provenance", //provenance.byUserID
		}

		for _, datum := range batchPlatform {
			for _, key := range doNotCompare {
				delete(datum, key)
			}
			cleanedPlatform = append(cleanedPlatform, datum)
		}

		for _, datum := range batchJellyfish {
			for _, key := range doNotCompare {
				delete(datum, key)
			}
			cleanedJellyfish = append(cleanedJellyfish, datum)
		}

		changelog, err := diff.Diff(cleanedPlatform, cleanedJellyfish, diff.StructMapKeySupport(), diff.AllowTypeMismatch(true), diff.FlattenEmbeddedStructs(), diff.SliceOrdering(false))
		if err != nil {
			return err
		}

		for _, change := range changelog {
			differences = append(differences, fmt.Sprintf("[%s] %s => expected:[%v] actual:[%v]", change.Type, strings.Join(change.Path, "."), change.From, change.To))
		}
		return nil
	}

	for i := 0; i < len(platformData); i += batch {
		j := i + batch
		if j > len(platformData) {
			j = len(platformData)
		}
		if err := processBatch(platformData[i:j], jellyfishData[i:j]); err != nil {
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
		Sort: bson.D{{Key: "time", Value: 1}, {Key: "type", Value: -1}},
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

func (m *DataVerify) FetchBlobIDs() ([]map[string]interface{}, error) {
	if m.dataC == nil {
		return nil, errors.New("missing data collection")
	}

	blobData := []map[string]interface{}{}

	dDataCursor, err := m.dataC.Find(m.ctx, bson.M{
		"deviceManufacturers":   bson.M{"$in": []string{"Tandem", "Insulet"}},
		"client.private.blobId": bson.M{"$exists": true},
	}, &options.FindOptions{
		Sort:       bson.M{"time": 1},
		Projection: bson.M{"client.private.blobId": 1, "time": 1, "deviceManufacturers": 1},
	})
	if err != nil {
		return nil, err
	}
	defer dDataCursor.Close(m.ctx)

	if err := dDataCursor.All(m.ctx, &blobData); err != nil {
		return nil, err
	}
	return blobData, nil
}

func (m *DataVerify) Verify(ref string, platformUploadID string, jellyfishUploadID string) error {

	platformDataset, err := m.fetchDataSet(platformUploadID)
	if err != nil {
		return err
	}

	jellyfishDataset, err := m.fetchDataSet(jellyfishUploadID)
	if err != nil {
		return err
	}

	log.Printf("Compare platform[%s] vs jellyfish[%s]", platformUploadID, jellyfishUploadID)
	differences, err := CompareDatasets(platformDataset, jellyfishDataset)
	if err != nil {
		return err
	}

	for _, v := range differences {
		log.Println(v)
	}

	return nil
}
