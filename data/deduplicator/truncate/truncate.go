package truncate

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/upload"
	"github.com/tidepool-org/platform/store"
)

func NewFactory() deduplicator.Factory {
	return &Factory{}
}

type Factory struct {
}

type Config struct {
	Name string `bson:"name"`
}

type Deduplicator struct {
	logger        log.Logger
	storeSession  store.Session
	datasetUpload *upload.Upload
	config        Config
}

func (f *Factory) CanDeduplicateDataset(datasetUpload *upload.Upload) (bool, error) {
	if datasetUpload == nil {
		return false, app.Error("truncate", "dataset upload is nil")
	}
	if config := datasetUpload.Deduplicator; config != nil {
		if configAsMap, configAsMapOk := config.(map[string]interface{}); configAsMapOk {
			return configAsMap["name"] == "truncate", nil
		} else if configAsM, configAsMOk := config.(bson.M); configAsMOk {
			return configAsM["name"] == "truncate", nil
		} else {
			return false, nil
		}
	} else if deviceModel := datasetUpload.DeviceModel; deviceModel != nil {
		if deviceID := datasetUpload.DeviceID; deviceID != nil {
			return true, nil
		}
	}
	return false, nil
}

func (f *Factory) NewDeduplicator(logger log.Logger, storeSession store.Session, datasetUpload *upload.Upload) (deduplicator.Deduplicator, error) {
	if logger == nil {
		return nil, app.Error("truncate", "logger is nil")
	}
	if storeSession == nil {
		return nil, app.Error("truncate", "store session is nil")
	}
	if datasetUpload == nil {
		return nil, app.Error("truncate", "dataset upload is nil")
	}

	return &Deduplicator{
		logger:        logger,
		storeSession:  storeSession,
		datasetUpload: datasetUpload,
		config: Config{
			Name: "truncate",
		},
	}, nil
}

func (d *Deduplicator) InitializeDataset() error {
	d.datasetUpload.SetDeduplicator(d.config)
	query := map[string]interface{}{"uploadId": d.datasetUpload.UploadID, "type": d.datasetUpload.Type}
	return d.storeSession.Update(query, d.datasetUpload)
}

func (d *Deduplicator) AddDataToDataset(datumArray []data.Datum) error {
	// TODO: FIXME: Lame Go array conversion
	insertArray := make([]interface{}, len(datumArray))
	for index, datum := range datumArray {
		insertArray[index] = datum
	}
	return d.storeSession.InsertAll(insertArray...)
}

func (d *Deduplicator) FinalizeDataset() error {
	datasetID := d.datasetUpload.UploadID
	deviceID := *d.datasetUpload.DeviceID

	// TODO: Technically, UpdateAll could succeed, but RemoveAll fail. This which result in duplicate (and possible incorrect) data.
	// TODO: Is there a way to resolve this?

	if err := d.storeSession.UpdateAll(bson.M{"uploadId": datasetID}, bson.M{"$set": bson.M{"_active": true}}); err != nil {
		return app.ExtErrorf(err, "truncate", "unable to activate data in dataset with id '%s'", datasetID)
	}

	if err := d.storeSession.RemoveAll(bson.M{"uploadId": bson.M{"$ne": datasetID}, "deviceId": deviceID, "type": bson.M{"$ne": "upload"}}); err != nil {
		return app.ExtErrorf(err, "truncate", "unable to delete data in datasets with device ID '%s' other than with id '%s'", deviceID, datasetID)
	}

	return nil
}
