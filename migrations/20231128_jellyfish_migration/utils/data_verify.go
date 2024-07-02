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

	if len(platformData) != len(jellyfishData) {
		log.Printf("NOTE: datasets mismatch platform (%d) vs jellyfish (%d)", len(platformData), len(jellyfishData))
	}

	// small batches or the diff takes to long
	var processBatch = func(batchPlatform, batchJellyfish []map[string]interface{}) ([]string, error) {

		cleanedJellyfish := []map[string]interface{}{}
		cleanedPlatform := []map[string]interface{}{}

		doNotCompare := []string{
			"_active",
			"_archivedTime",
			"createdTime",
			"deduplicator",
			"_deduplicator",
			"_groupId",
			"guid",
			"_id",
			"id",
			"modifiedTime",
			"payload",
			"provenance",
			"revision",
			"_schemaVersion",
			"time",
			"_userId",
			"uploadId",
			"_version",
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
			return nil, err
		}
		diffs := []string{}
		for _, change := range changelog {
			diffs = append(diffs, fmt.Sprintf("%s => platform:[%v] jellyfish:[%v]", strings.Join(change.Path, "."), change.From, change.To))
		}
		return diffs, nil
	}

	var processAllData = func() ([]string, error) {
		batch := 100
		differences := []string{}
		for i := 0; i < len(platformData); i += batch {
			j := i + batch
			if j > len(platformData) {
				j = len(platformData)
			}
			if batchDiff, err := processBatch(platformData[i:j], jellyfishData[i:j]); err != nil {
				return nil, err
			} else {
				differences = append(differences, batchDiff...)
			}
		}
		return differences, nil
	}

	var processSubsetOfData = func() ([]string, error) {

		differences := []string{}
		batch := 20
		quater := len(platformData) / 4
		batchStarts := []int{
			0,
			quater,
			quater * 2,
			quater * 3,
		}

		log.Printf("NOTE: comparing a subset of all [%d] datum with a batch size [%d] starting at [%v] ", len(platformData), batch, batchStarts)

		for _, startAt := range batchStarts {
			j := startAt + batch

			if j > len(platformData) {
				j = len(platformData)
			}
			if batchDiff, err := processBatch(platformData[startAt:j], jellyfishData[startAt:j]); err != nil {
				return nil, err
			} else {
				differences = append(differences, batchDiff...)
			}
		}
		return differences, nil
	}

	if len(platformData) <= 100 {
		return processAllData()
	}
	return processSubsetOfData()
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

var DatasetTypes = []string{"cbg", "basal", "bolus", "deviceEvent", "wizard", "pumpSettings"}

func (m *DataVerify) fetchDataSet(uploadID string, dataTypes []string) (map[string][]map[string]interface{}, error) {
	if m.dataC == nil {
		return nil, errors.New("missing data collection")
	}

	typeSet := map[string][]map[string]interface{}{}

	for _, dType := range dataTypes {

		dset := []map[string]interface{}{}

		dDataCursor, err := m.dataC.Find(m.ctx, bson.M{
			"uploadId": uploadID,
			"type":     dType,
		}, &options.FindOptions{
			Sort: bson.M{"time": 1},
		})
		if err != nil {
			return nil, err
		}
		defer dDataCursor.Close(m.ctx)

		if err := dDataCursor.All(m.ctx, &dset); err != nil {
			return nil, err
		}
		log.Printf("got dataset [%s][%s][%d] results", uploadID, dType, len(dset))
		typeSet[dType] = dset
	}
	return typeSet, nil
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
		Sort:       bson.D{{"deviceId", 1}, {"time", 1}},
		Projection: bson.M{"_id": 0, "deviceId": 1, "deviceManufacturers": 1, "client.private.blobId": 1},
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

func (m *DataVerify) Verify(ref string, platformUploadID string, jellyfishUploadID string, dataTyes []string) error {

	if len(dataTyes) == 0 {
		dataTyes = DatasetTypes
	}

	platformDataset, err := m.fetchDataSet(platformUploadID, dataTyes)
	if err != nil {
		return err
	}

	jellyfishDataset, err := m.fetchDataSet(jellyfishUploadID, dataTyes)
	if err != nil {
		return err
	}

	log.Printf("Compare platform[%s] vs jellyfish[%s]", platformUploadID, jellyfishUploadID)

	for dType, jfSet := range jellyfishDataset {

		differences, err := CompareDatasets(platformDataset[dType], jfSet)
		if err != nil {
			return err
		}
		for _, v := range differences {
			log.Printf("%s.%s", dType, v)
		}
	}

	return nil
}
