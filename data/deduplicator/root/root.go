package root

import (
	"fmt"

	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/deduplicator/alignment"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/store"
)

func NewFactory() deduplicator.Factory {
	return &Factory{
		[]deduplicator.Factory{
			alignment.NewFactory(),
		},
	}
}

type Factory struct {
	factories []deduplicator.Factory
}

func (f *Factory) CanDeduplicateDataset(datasetUpload *upload.Upload) (bool, error) {
	for _, factory := range f.factories {
		if can, err := factory.CanDeduplicateDataset(datasetUpload); err != nil {
			return false, err
		} else if can {
			return true, nil
		}
	}
	return false, nil
}

func (f *Factory) NewDeduplicator(datasetUpload *upload.Upload, storeSession store.Session, logger log.Logger) (deduplicator.Deduplicator, error) {
	for _, factory := range f.factories {
		if can, err := factory.CanDeduplicateDataset(datasetUpload); err != nil {
			return nil, err
		} else if can {
			return factory.NewDeduplicator(datasetUpload, storeSession, logger)
		}
	}
	return nil, fmt.Errorf("Deduplicator not found")
}
