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

func CompareDatasetDatums(platformData []map[string]interface{}, jellyfishData []map[string]interface{}, ignoredPaths ...string) (map[string]interface{}, error) {
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

func (m *DataVerify) WriteBlobIDs() error {
	if m.dataC == nil {
		return errors.New("missing data collection")
	}

	blobData := []map[string]interface{}{}

	dDataCursor, err := m.dataC.Find(m.ctx, bson.M{
		"deviceManufacturers":   bson.M{"$in": []string{"Tandem", "Insulet"}},
		"client.private.blobId": bson.M{"$exists": true},
		"_active":               true,
	}, &options.FindOptions{
		Sort:       bson.D{{Key: "deviceId", Value: 1}, {Key: "time", Value: 1}},
		Projection: bson.M{"_id": 0, "deviceId": 1, "blobId": "$client.private.blobId", "time": 1},
	})
	if err != nil {
		return err
	}
	defer dDataCursor.Close(m.ctx)

	if err := dDataCursor.All(m.ctx, &blobData); err != nil {
		return err
	}

	type Blob struct {
		DeviceID string `json:"deviceId"`
		BlobID   string `json:"blobId"`
	}

	blobs := []Blob{}

	for _, v := range blobData {
		blobs = append(blobs, Blob{
			BlobID:   fmt.Sprintf("%v", v["blobId"]),
			DeviceID: fmt.Sprintf("%v", v["deviceId"])})
	}

	blobPath := filepath.Join(".", "_blobs")
	log.Printf("blob data written to %s", blobPath)
	writeFileData(blobs, blobPath, "device_blobs.json", true)
	return nil
}

const (
	PlatformExtra     = "extra"
	PlatformDuplicate = "duplicate"
	PlatformMissing   = "missing"
)

func CompareDatasets(platformSet []map[string]interface{}, jellyfishSet []map[string]interface{}) map[string][]map[string]interface{} {

	diffs := map[string][]map[string]interface{}{
		PlatformExtra:     {},
		PlatformDuplicate: {},
		PlatformMissing:   {},
	}
	const deviceTimeName = "deviceTime"
	type deviceTimeDatums map[string][]map[string]interface{}

	pfCounts := deviceTimeDatums{}
	jfCounts := deviceTimeDatums{}

	for _, jDatum := range jellyfishSet {
		strDatumTime := fmt.Sprintf("%v", jDatum[deviceTimeName])

		if len(jfCounts[strDatumTime]) == 0 {
			jfCounts[strDatumTime] = []map[string]interface{}{jDatum}
		} else if len(jfCounts[strDatumTime]) >= 1 {
			jfCounts[strDatumTime] = append(jfCounts[strDatumTime], jDatum)
		}
	}

	for _, pDatum := range platformSet {

		strDatumTime := fmt.Sprintf("%v", pDatum[deviceTimeName])

		if len(pfCounts[strDatumTime]) == 0 {
			pfCounts[strDatumTime] = []map[string]interface{}{pDatum}
		} else if len(pfCounts[strDatumTime]) >= 1 {

			currentItems := pfCounts[strDatumTime]
			for _, item := range currentItems {
				if fmt.Sprintf("%v", item) == fmt.Sprintf("%v", pDatum) {
					diffs[PlatformDuplicate] = append(diffs[PlatformDuplicate], pDatum)
					continue
				} else {
					diffs[PlatformExtra] = append(diffs[PlatformExtra], pDatum)
					break
				}
			}
			pfCounts[strDatumTime] = append(pfCounts[strDatumTime], pDatum)
		}
		if len(jfCounts[fmt.Sprintf("%v", pDatum[deviceTimeName])]) == 0 {
			diffs[PlatformExtra] = append(diffs[PlatformExtra], pDatum)
		}
	}

	for jfDeviceTimeStr, jDatums := range jfCounts {
		if len(pfCounts[jfDeviceTimeStr]) < len(jfCounts[jfDeviceTimeStr]) {
			//NOTE: more of an indicator there are missing records ...
			for i := len(pfCounts[jfDeviceTimeStr]); i < len(jfCounts[jfDeviceTimeStr]); i++ {
				diffs[PlatformMissing] = append(diffs[PlatformMissing], jDatums[i])
			}
		}
	}
	return diffs
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
		setDifferences := CompareDatasets(pfSet, jfSet)
		if len(setDifferences[PlatformMissing]) > 0 {
			writeFileData(setDifferences[PlatformMissing], comparePath, fmt.Sprintf("%s_platform_missing.json", dType), true)
		}
		if len(setDifferences[PlatformDuplicate]) > 0 {
			writeFileData(setDifferences[PlatformDuplicate], comparePath, fmt.Sprintf("%s_platform_duplicates.json", dType), true)
		}
		if len(setDifferences[PlatformExtra]) > 0 {
			writeFileData(setDifferences[PlatformExtra], comparePath, fmt.Sprintf("%s_platform_extra.json", dType), true)
		}
		if len(pfSet) != len(jfSet) {
			log.Printf("NOTE: datasets mismatch platform (%d) vs jellyfish (%d)", len(pfSet), len(jfSet))
			writeFileData(jfSet, comparePath, fmt.Sprintf("%s_jellyfish_datums.json", dType), true)
			writeFileData(pfSet, comparePath, fmt.Sprintf("%s_platform_datums.json", dType), true)
			break
		}
		differences, err := CompareDatasetDatums(pfSet, jfSet, dataTypePathIgnored[dType]...)
		if err != nil {
			return err
		}
		if len(differences) > 0 {
			writeFileData(differences, comparePath, fmt.Sprintf("%s_datum_diff.json", dType), true)
		}
	}
	return nil
}
