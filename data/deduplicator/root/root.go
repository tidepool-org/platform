package root

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
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/deduplicator/truncate"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
)

func NewFactory() deduplicator.Factory {
	return &Factory{
		[]deduplicator.Factory{
			truncate.NewFactory(),
		},
	}
}

type Factory struct {
	factories []deduplicator.Factory
}

func (f *Factory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	for _, factory := range f.factories {
		if can, err := factory.CanDeduplicateDataset(dataset); err != nil {
			return false, err
		} else if can {
			return true, nil
		}
	}
	return false, nil
}

func (f *Factory) NewDeduplicator(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (deduplicator.Deduplicator, error) {
	for _, factory := range f.factories {
		if can, err := factory.CanDeduplicateDataset(dataset); err != nil {
			return nil, err
		} else if can {
			return factory.NewDeduplicator(logger, dataStoreSession, dataset)
		}
	}
	return nil, app.Error("root", "deduplicator not found")
}
