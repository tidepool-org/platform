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
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
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
	logger           log.Logger
	dataStoreSession store.Session
	dataset          *upload.Upload
	config           Config
}

func (f *Factory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if dataset == nil {
		return false, app.Error("truncate", "dataset upload is nil")
	}
	if config := dataset.Deduplicator; config != nil {
		if configAsMap, configAsMapOk := config.(map[string]interface{}); configAsMapOk {
			return configAsMap["name"] == "truncate", nil
		} else if configAsM, configAsMOk := config.(bson.M); configAsMOk {
			return configAsM["name"] == "truncate", nil
		} else {
			return false, nil
		}
	} else if deviceModel := dataset.DeviceModel; deviceModel != nil {
		if deviceID := dataset.DeviceID; deviceID != nil {
			return true, nil
		}
	}
	return false, nil
}

func (f *Factory) NewDeduplicator(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (deduplicator.Deduplicator, error) {
	if logger == nil {
		return nil, app.Error("truncate", "logger is nil")
	}
	if dataStoreSession == nil {
		return nil, app.Error("truncate", "store session is nil")
	}
	if dataset == nil {
		return nil, app.Error("truncate", "dataset upload is nil")
	}

	return &Deduplicator{
		logger:           logger,
		dataStoreSession: dataStoreSession,
		dataset:          dataset,
		config: Config{
			Name: "truncate",
		},
	}, nil
}

func (d *Deduplicator) InitializeDataset() error {
	d.dataset.SetDeduplicator(d.config)

	if err := d.dataStoreSession.UpdateDataset(d.dataset); err != nil {
		return app.ExtError(err, "truncate", "unable to initialize dataset")
	}

	return nil
}

func (d *Deduplicator) AddDataToDataset(datasetData []data.Datum) error {
	return d.dataStoreSession.CreateDatasetData(d.dataset, datasetData)
}

func (d *Deduplicator) FinalizeDataset() error {
	// TODO: Technically, ActivateAllDatasetData could succeed, but RemoveAllOtherDatasetData fail. This would
	// result in duplicate (and possible incorrect) data. Is there a way to resolve this? Would be nice to have transactions.

	if err := d.dataStoreSession.ActivateAllDatasetData(d.dataset); err != nil {
		return app.ExtErrorf(err, "truncate", "unable to activate data in dataset with id '%s'", d.dataset.UploadID)
	}
	if err := d.dataStoreSession.RemoveAllOtherDatasetData(d.dataset); err != nil {
		return app.ExtErrorf(err, "truncate", "unable to remove all other data except dataset with id '%s'", d.dataset.UploadID)
	}

	return nil
}
