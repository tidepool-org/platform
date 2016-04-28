package alignment

import (
	"fmt"
	"reflect"

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

func NewFactory() deduplicator.Factory {
	return &Factory{}
}

type Factory struct {
}

type Config struct {
	Name    string   `bson:"name"`
	Sort    []string `bson:"sort"`
	Match   []string `bson:"match"`
	Unified bool     `bson:"unified"`
}

type Deduplicator struct {
	logger        log.Logger
	datasetUpload *upload.Upload
	storeSession  store.Session
	config        Config
}

func (f *Factory) CanDeduplicateDataset(datasetUpload *upload.Upload) (bool, error) {
	if datasetUpload == nil {
		return false, fmt.Errorf("Dataset upload is nil")
	}
	if config := datasetUpload.Deduplicator; config != nil {
		if configAsMap, ok := config.(map[string]interface{}); ok {
			return configAsMap["name"] == "alignment", nil
		} else if configAsM, ok := config.(bson.M); ok {
			return configAsM["name"] == "alignment", nil
		} else {
			return false, nil
		}
	} else if deviceModel := datasetUpload.DeviceModel; deviceModel != nil {
		// switch *deviceModel {
		// case "G4Receiver":
		// 	return true, nil
		// }
		return true, nil
	}
	return false, nil
}

func (f *Factory) NewDeduplicator(datasetUpload *upload.Upload, storeSession store.Session, logger log.Logger) (deduplicator.Deduplicator, error) {
	if datasetUpload == nil {
		return nil, fmt.Errorf("Dataset datum is nil")
	}
	if storeSession == nil {
		return nil, fmt.Errorf("Dataset datum is nil")
	}
	return &Deduplicator{
		logger:        logger,
		datasetUpload: datasetUpload,
		storeSession:  storeSession,
		config: Config{
			Name:    "alignment",
			Sort:    []string{"payload.internalTime"},
			Match:   []string{"deviceTime"},
			Unified: false,
		},
	}, nil
}

func (d *Deduplicator) InitializeDataset() error {
	d.datasetUpload.Deduplicator = d.config
	query := map[string]interface{}{"type": d.datasetUpload.Type, "uploadId": d.datasetUpload.UploadID}
	return d.storeSession.Update(query, d.datasetUpload)
}

func (d *Deduplicator) AddDataToDataset(datumArray data.BuiltDatumArray) error {
	// Go is a STUPID language! I've used it long enough to decide this. I've never thought a modern language was such
	// a piece of crap as Go. I think Google is purposefully trying to convince developers that a "simple" language is
	// *SO* much better that these more complicated languages. Yea, like I want to write my own copy functions ten thousand
	// times.  What a piece of bullshit Google is peddling.
	saveArray := make([]interface{}, len(datumArray))
	for index, datum := range datumArray {
		saveArray[index] = datum
	}
	return d.storeSession.InsertAll(saveArray...)
}

func (d *Deduplicator) FinalizeDataset() error {
	newDatasetID := *d.datasetUpload.UploadID

	if previousDatasetUpload, err := d.findPreviousDataset(); err != nil {
		return err
	} else if previousDatasetUpload != nil {
		previousDatasetID := *previousDatasetUpload.UploadID
		return d.deduplicateDataset(previousDatasetID, newDatasetID)
	}

	d.logger.Warn("No previous dataset found; activating all data")
	return d.activateDataInDataset(newDatasetID)
}

func (d *Deduplicator) findPreviousDataset() (*upload.Upload, error) {
	groupID := d.datasetUpload.GroupID
	datasetID := d.datasetUpload.UploadID

	// TODO: Updated query to pull first, need order
	iter := d.storeSession.FindAll(store.Query{"_groupId": groupID, "type": "upload", "uploadId": map[string]interface{}{"$ne": datasetID}}, []string{}, store.Filter{})
	defer iter.Close()

	d.logger.WithField("iter.Err()", iter.Err()).Warn("findPreviousDataset iter.Err()")

	// TODO Check error here
	d.logger.WithField("d.datasetUpload", d.datasetUpload).Warn("findPreviousDataset")

	var previousDatasetUpload *upload.Upload
	datasetUpload := upload.Upload{}
	for iter.Next(&datasetUpload) {
		d.logger.WithField("datasetUpload", datasetUpload).Warn("findPreviousDataset possible")
		if previousDatasetUpload == nil || *previousDatasetUpload.Time < *datasetUpload.Time {
			d.logger.WithField("datasetUpload", datasetUpload).Warn("findPreviousDataset found")
			previousDatasetUpload = &datasetUpload
		}
	}

	return previousDatasetUpload, nil
}

func (d *Deduplicator) deduplicateDataset(previousDatasetID string, newDatasetID string) error {
	previousDatumArray, err := d.readDatumArrayForDataset(previousDatasetID)
	if err != nil {
		return err
	}
	newDatumArray, err := d.readDatumArrayForDataset(newDatasetID)
	if err != nil {
		return err
	}
	filter := bson.M{"uploadId": newDatasetID}
	if d.config.Unified {
		return d.deduplicateDatasetUnifed(previousDatumArray, newDatumArray, filter)
	}
	return d.deduplicateDatasetByDatumType(previousDatumArray, newDatumArray, filter)
}

func (d *Deduplicator) readDatumArrayForDataset(datasetID string) ([]bson.M, error) {
	filter := store.Filter{"_id": false, "id": true, "type": true}
	for _, match := range d.config.Match {
		filter[match] = true
	}

	iter := d.storeSession.FindAll(store.Query{"uploadId": datasetID}, d.config.Sort, filter)
	if err := iter.Err(); err != nil {
		return nil, err
	}

	var datumArray []bson.M
	if err := iter.All(&datumArray); err != nil {
		return nil, err
	}

	return datumArray, nil
}

func (d *Deduplicator) deduplicateDatasetUnifed(previousDatumArray []bson.M, newDatumArray []bson.M, filter bson.M) error {
	return d.deduplicateDatumArray(previousDatumArray, newDatumArray, filter)
}

func (d *Deduplicator) deduplicateDatasetByDatumType(previousDatumArray []bson.M, newDatumArray []bson.M, filter bson.M) error {
	filter = shallowCloneMap(filter)
	datumTypes := d.calculateDatumTypes(newDatumArray)
	for _, datumType := range datumTypes {
		d.logger.Warn(fmt.Sprintf("Deduplicating type: %s", datumType))
		previousDatumArrayByType := d.filterDatumArrayByDatumType(previousDatumArray, datumType)
		newDatumArrayByType := d.filterDatumArrayByDatumType(newDatumArray, datumType)
		filter["type"] = datumType
		if err := d.deduplicateDatumArray(previousDatumArrayByType, newDatumArrayByType, filter); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deduplicator) calculateDatumTypes(datumArray []bson.M) []string {
	datumMap := make(map[string]bool)
	for _, datum := range datumArray {
		if datumType, ok := datum["type"].(string); ok {
			datumMap[datumType] = true
		}
	}

	datumTypes := []string{}
	for datumType := range datumMap {
		datumTypes = append(datumTypes, datumType)
	}

	return datumTypes
}

func (d *Deduplicator) filterDatumArrayByDatumType(datumArray []bson.M, datumType string) []bson.M {
	filteredDatumArray := []bson.M{}
	for _, datum := range datumArray {
		if datum["type"] == datumType {
			filteredDatumArray = append(filteredDatumArray, datum)
		}
	}
	return filteredDatumArray
}

func (d *Deduplicator) deduplicateDatumArray(previousDatumArray []bson.M, newDatumArray []bson.M, filter bson.M) error {
	filter = shallowCloneMap(filter)
	previousLength := len(previousDatumArray)
	for previousIndex := range previousDatumArray {
		if d.datumArrayAlignment(previousDatumArray[previousIndex:], newDatumArray[:(previousLength-previousIndex)]) {
			var ids []string
			for _, newDatum := range newDatumArray[(previousLength - previousIndex):] {
				ids = append(ids, newDatum["id"].(string))
			}

			// TODO: Cleanup
			d.logger.Warn(fmt.Sprintf("Deduplicating against previous dataset; activating ids: %d", len(ids)))
			filter["id"] = bson.M{"$in": ids}
			return d.storeSession.UpdateAll(filter, bson.M{"$set": bson.M{"_active": true}})
		}
	}

	// TODO: Cleanup
	d.logger.Warn("Deduplicating against previous dataset; no overlap; activating ids: all")
	return d.storeSession.UpdateAll(filter, bson.M{"$set": bson.M{"_active": true}})
}

func (d *Deduplicator) datumArrayAlignment(leftDatumArray []bson.M, rightDatumArray []bson.M) bool {
	if len(leftDatumArray) != len(rightDatumArray) {
		return false
	}
	for index, leftDatum := range leftDatumArray {
		if !d.datumMatch(leftDatum, rightDatumArray[index]) {
			return false
		}
	}
	return true
}

func (d *Deduplicator) datumMatch(leftDatum bson.M, rightDatum bson.M) bool {
	for _, match := range d.config.Match {
		leftValue, leftOK := leftDatum[match]
		rightValue, rightOK := rightDatum[match]
		if leftOK != rightOK {
			return false
		}
		if leftOK && rightOK && !reflect.DeepEqual(leftValue, rightValue) {
			return false
		}
	}
	return true
}

func (d *Deduplicator) activateDataInDataset(datasetID string) error {
	return d.storeSession.UpdateAll(bson.M{"uploadId": datasetID}, bson.M{"$set": bson.M{"_active": true}})
}

func shallowCloneMap(source map[string]interface{}) map[string]interface{} {
	destination := make(map[string]interface{})
	for key, value := range source {
		destination[key] = value
	}
	return destination
}
