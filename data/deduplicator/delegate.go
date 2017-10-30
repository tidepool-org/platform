package deduplicator

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type DelegateFactory struct {
	factories []Factory
}

func NewDelegateFactory(factories []Factory) (*DelegateFactory, error) {
	if len(factories) == 0 {
		return nil, errors.New("factories is missing")
	}

	return &DelegateFactory{
		factories: factories,
	}, nil
}

func (d *DelegateFactory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	if dataset == nil {
		return false, errors.New("dataset is missing")
	}

	for _, factory := range d.factories {
		if can, err := factory.CanDeduplicateDataset(dataset); err != nil {
			return false, err
		} else if can {
			return true, nil
		}
	}
	return false, nil
}

func (d *DelegateFactory) NewDeduplicatorForDataset(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataset *upload.Upload) (data.Deduplicator, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if dataSession == nil {
		return nil, errors.New("data store session is missing")
	}
	if dataset == nil {
		return nil, errors.New("dataset is missing")
	}

	for _, factory := range d.factories {
		if can, err := factory.CanDeduplicateDataset(dataset); err != nil {
			return nil, err
		} else if can {
			return factory.NewDeduplicatorForDataset(logger, dataSession, dataset)
		}
	}
	return nil, errors.New("deduplicator not found")
}

func (d *DelegateFactory) IsRegisteredWithDataset(dataset *upload.Upload) (bool, error) {
	if dataset == nil {
		return false, errors.New("dataset is missing")
	}

	for _, factory := range d.factories {
		if is, err := factory.IsRegisteredWithDataset(dataset); err != nil {
			return false, err
		} else if is {
			return true, nil
		}
	}
	return false, nil
}

func (d *DelegateFactory) NewRegisteredDeduplicatorForDataset(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataset *upload.Upload) (data.Deduplicator, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if dataSession == nil {
		return nil, errors.New("data store session is missing")
	}
	if dataset == nil {
		return nil, errors.New("dataset is missing")
	}

	deduplicatorDescriptor := dataset.DeduplicatorDescriptor()
	if deduplicatorDescriptor == nil || !deduplicatorDescriptor.IsRegisteredWithAnyDeduplicator() {
		return nil, errors.Newf("dataset not registered with deduplicator")
	}

	for _, factory := range d.factories {
		if is, err := factory.IsRegisteredWithDataset(dataset); err != nil {
			return nil, err
		} else if is {
			return factory.NewRegisteredDeduplicatorForDataset(logger, dataSession, dataset)
		}
	}
	return nil, errors.New("deduplicator not found")
}
