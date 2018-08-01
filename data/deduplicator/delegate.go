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

func (d *DelegateFactory) CanDeduplicateDataSet(dataSet *upload.Upload) (bool, error) {
	if dataSet == nil {
		return false, errors.New("data set is missing")
	}

	for _, factory := range d.factories {
		if can, err := factory.CanDeduplicateDataSet(dataSet); err != nil {
			return false, err
		} else if can {
			return true, nil
		}
	}
	return false, nil
}

func (d *DelegateFactory) NewDeduplicatorForDataSet(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataSet *upload.Upload) (data.Deduplicator, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if dataSession == nil {
		return nil, errors.New("data store session is missing")
	}
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	}

	if dataSet.Deduplicator != nil {
		if dataSet.Deduplicator.Name == "" {
			return nil, errors.New("data set deduplicator name is missing")
		}

		for _, factory := range d.factories {
			if is, err := factory.IsRegisteredWithDataSet(dataSet); err != nil {
				return nil, err
			} else if is {
				return factory.NewDeduplicatorForDataSet(logger, dataSession, dataSet)
			}
		}

		return nil, errors.New("data set deduplicator name is unknown")
	}

	for _, factory := range d.factories {
		if can, err := factory.CanDeduplicateDataSet(dataSet); err != nil {
			return nil, err
		} else if can {
			return factory.NewDeduplicatorForDataSet(logger, dataSession, dataSet)
		}
	}
	return nil, errors.New("deduplicator not found")
}

func (d *DelegateFactory) IsRegisteredWithDataSet(dataSet *upload.Upload) (bool, error) {
	if dataSet == nil {
		return false, errors.New("data set is missing")
	}

	for _, factory := range d.factories {
		if is, err := factory.IsRegisteredWithDataSet(dataSet); err != nil {
			return false, err
		} else if is {
			return true, nil
		}
	}
	return false, nil
}

func (d *DelegateFactory) NewRegisteredDeduplicatorForDataSet(logger log.Logger, dataSession storeDEPRECATED.DataSession, dataSet *upload.Upload) (data.Deduplicator, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if dataSession == nil {
		return nil, errors.New("data store session is missing")
	}
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	}

	deduplicatorDescriptor := dataSet.DeduplicatorDescriptor()
	if deduplicatorDescriptor == nil || !deduplicatorDescriptor.IsRegisteredWithAnyDeduplicator() {
		return nil, errors.Newf("data set not registered with deduplicator")
	}

	for _, factory := range d.factories {
		if is, err := factory.IsRegisteredWithDataSet(dataSet); err != nil {
			return nil, err
		} else if is {
			return factory.NewRegisteredDeduplicatorForDataSet(logger, dataSession, dataSet)
		}
	}
	return nil, errors.New("deduplicator not found")
}
