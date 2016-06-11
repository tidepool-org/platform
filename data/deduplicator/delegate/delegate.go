package delegate

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
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
)

type factory struct {
	factories []deduplicator.Factory
}

func NewFactory(factories []deduplicator.Factory) (deduplicator.Factory, error) {
	if len(factories) == 0 {
		return nil, app.Error("delegate", "factories is missing")
	}

	return &factory{
		factories: factories,
	}, nil
}

func (f *factory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if dataset == nil {
		return false, app.Error("delegate", "dataset is missing")
	}

	for _, factory := range f.factories {
		if can, err := factory.CanDeduplicateDataset(dataset); err != nil {
			return false, err
		} else if can {
			return true, nil
		}
	}
	return false, nil
}

func (f *factory) NewDeduplicator(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (deduplicator.Deduplicator, error) {
	if logger == nil {
		return nil, app.Error("delegate", "logger is missing")
	}
	if dataStoreSession == nil {
		return nil, app.Error("delegate", "data store session is missing")
	}
	if dataset == nil {
		return nil, app.Error("delegate", "dataset is missing")
	}

	for _, factory := range f.factories {
		if can, err := factory.CanDeduplicateDataset(dataset); err != nil {
			return nil, err
		} else if can {
			return factory.NewDeduplicator(logger, dataStoreSession, dataset)
		}
	}
	return nil, app.Error("delegate", "deduplicator not found")
}
