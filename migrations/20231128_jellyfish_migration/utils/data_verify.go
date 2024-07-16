package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/r3labs/diff/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DataVerify struct {
	ctx   context.Context
	dataC *mongo.Collection
}

func CompareDatasets(platformData []map[string]interface{}, jellyfishData []map[string]interface{}, ignoredPaths ...string) (map[string]interface{}, error) {
	diffs := map[string]interface{}{}
	for id, platformDatum := range platformData {
		if jellyfishData[id] == nil {
			log.Println("no matching value in the jellyfish data")
			break
		}
		changelog, err := diff.Diff(platformDatum, jellyfishData[id], diff.ConvertCompatibleTypes(), diff.StructMapKeySupport(), diff.AllowTypeMismatch(true), diff.FlattenEmbeddedStructs(), diff.SliceOrdering(false))
		if err != nil {
			return nil, err
		}
		if len(changelog) > 0 {
			if ignoredPaths != nil {
				for _, path := range ignoredPaths {
					changelog = changelog.FilterOut([]string{path})
				}
				if len(changelog) == 0 {
					continue
				}
			}
			diffs[fmt.Sprintf("platform_%d", id)] = changelog
		}
	}
	return diffs, nil
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

var DatasetTypes = []string{"cbg", "smbg", "basal", "bolus", "deviceEvent", "wizard", "pumpSettings"}

func (m *DataVerify) fetchDataSet(uploadID string, dataTypes []string) (map[string][]map[string]interface{}, error) {
	if m.dataC == nil {
		return nil, errors.New("missing data collection")
	}

	typeSet := map[string][]map[string]interface{}{}

	for _, dType := range dataTypes {

		dset := []map[string]interface{}{}

		filter := bson.M{
			"uploadId": uploadID,
			"type":     dType,
		}

		sort := bson.D{{Key: "time", Value: 1}}

		if dType == "deviceEvent" || dType == "bolus" {
			sort = bson.D{{Key: "time", Value: 1}, {Key: "subType", Value: 1}}
		}

		excludedFeilds := bson.M{
			"_active":          0,
			"_archivedTime":    0,
			"createdTime":      0,
			"clockDriftOffset": 0,
			"conversionOffset": 0,
			"deduplicator":     0,
			"_deduplicator":    0,
			"_groupId":         0,
			"guid":             0,
			"_id":              0,
			"id":               0,
			"modifiedTime":     0,
			"payload":          0,
			"provenance":       0,
			"revision":         0,
			"_schemaVersion":   0,
			"time":             0,
			"timezoneOffset":   0,
			"type":             0,
			"_userId":          0,
			"uploadId":         0,
			"_version":         0,
		}

		dDataCursor, err := m.dataC.Find(m.ctx, filter, &options.FindOptions{
			Sort:       sort,
			Projection: excludedFeilds,
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
		Sort:       bson.D{{Key: "deviceId", Value: 1}, {Key: "time", Value: 1}},
		Projection: bson.M{"_id": 0, "deviceId": 1, "blobId": "$client.private.blobId", "_userId": 1, "time": 1},
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

func getMissing(a []map[string]interface{}, b []map[string]interface{}) []map[string]interface{} {
	missing := []map[string]interface{}{}

	more := a
	less := b

	if len(b) > len(a) {
		more = b
		less = a
	}

	ma := make(map[string]bool, len(less))
	for _, ka := range less {
		ma[fmt.Sprintf("%v", ka["deviceTime"])] = true
	}
	for _, kb := range more {
		if !ma[fmt.Sprintf("%v", kb["deviceTime"])] {
			missing = append(missing, kb)
		}
	}
	return missing

}

var dataTypePathIgnored = map[string][]string{
	"smbg":  {"raw", "value"},
	"cbg":   {"value"},
	"basal": {"rate"},
	"bolus": {"normal"},
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
		pfSet := platformDataset[dType]
		comparePath := filepath.Join(".", "_compare", fmt.Sprintf("%s_%s", platformUploadID, jellyfishUploadID))
		log.Printf("data written to %s", comparePath)
		if len(pfSet) != len(jfSet) {
			log.Printf("NOTE: datasets mismatch platform (%d) vs jellyfish (%d)", len(pfSet), len(jfSet))
			missing := getMissing(pfSet, jfSet)
			writeFileData(missing, comparePath, fmt.Sprintf("missing_%s.json", dType), true)
			writeFileData(jfSet, comparePath, fmt.Sprintf("raw_%s_jf_%s.json", dType, jellyfishUploadID), true)
			writeFileData(pfSet, comparePath, fmt.Sprintf("raw_%s_pf_%s.json", dType, platformUploadID), true)
			break
		}

		differences, err := CompareDatasets(pfSet, jfSet, dataTypePathIgnored[dType]...)
		if err != nil {
			return err
		}
		writeFileData(differences, comparePath, fmt.Sprintf("%s_diff.json", dType), true)
	}

	return nil
}
