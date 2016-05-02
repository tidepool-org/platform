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
	Name string `bson:"name"`
}

type Deduplicator struct {
	logger        log.Logger
	datasetUpload *upload.Upload
	storeSession  store.Session
	config        Config
}

func (f *Factory) CanDeduplicateDataset(datasetUpload *upload.Upload) (bool, error) {
	if datasetUpload == nil {
		return false, app.Error("truncate", "dataset upload is nil")
	}
	if config := datasetUpload.Deduplicator; config != nil {
		if configAsMap, ok := config.(map[string]interface{}); ok {
			return configAsMap["name"] == "truncate", nil
		} else if configAsM, ok := config.(bson.M); ok {
			return configAsM["name"] == "truncate", nil
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
		return nil, app.Error("truncate", "dataset upload is nil")
	}
	if storeSession == nil {
		return nil, app.Error("truncate", "store session is nil")
	}
	if logger == nil {
		return nil, app.Error("truncate", "logger is nil")
	}
	return &Deduplicator{
		logger:        logger,
		datasetUpload: datasetUpload,
		storeSession:  storeSession,
		config: Config{
			Name: "truncate",
		},
	}, nil
}

func (d *Deduplicator) InitializeDataset() error {
	d.datasetUpload.Deduplicator = d.config
	query := map[string]interface{}{"uploadId": d.datasetUpload.UploadID, "type": d.datasetUpload.Type}
	return d.storeSession.Update(query, d.datasetUpload)
}

func (d *Deduplicator) AddDataToDataset(datumArray data.BuiltDatumArray) error {
	return d.storeSession.InsertAll(datumArray...)
}

func (d *Deduplicator) FinalizeDataset() error {
	datasetID := *d.datasetUpload.UploadID

	// TODO: Technically, UpdateAll could succeed, but RemoveAll fail. This which result in duplicate (and possible incorrect) data.
	// TODO: Is there a way to resolve this?

	if err := d.storeSession.UpdateAll(bson.M{"uploadId": datasetID}, bson.M{"$set": bson.M{"_active": true}}); err != nil {
		return app.ExtErrorf(err, "truncate", "unable to activate data in dataset with id %s", datasetID)
	}

	if err := d.storeSession.RemoveAll(bson.M{"uploadId": bson.M{"$ne": datasetID}, "type": bson.M{"$ne": "upload"}}); err != nil {
		return app.ExtErrorf(err, "truncate", "unable to delete data in datasets other than with id %s", datasetID)
	}

	return nil
}
