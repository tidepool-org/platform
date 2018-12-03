package factory

import (
	dataDeduplicator "github.com/tidepool-org/platform/data/deduplicator"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Deduplicator interface {
	dataDeduplicator.Deduplicator

	New(dataSet *dataTypesUpload.Upload) (bool, error)
	Get(dataSet *dataTypesUpload.Upload) (bool, error)
}

type Factory struct {
	deduplicators []Deduplicator
}

func New(deduplicators []Deduplicator) (*Factory, error) {
	if deduplicators == nil {
		return nil, errors.New("deduplicators is missing")
	}

	return &Factory{
		deduplicators: deduplicators,
	}, nil
}

func (f *Factory) New(dataSet *dataTypesUpload.Upload) (dataDeduplicator.Deduplicator, error) {
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	} else if err := structureValidator.New().WithOrigin(structure.OriginStore).Validate(dataSet); err != nil {
		return nil, errors.Wrap(err, "data set is invalid")
	}

	if dataSet.HasDeduplicatorName() {
		return f.get(dataSet)
	}

	for _, deduplicator := range f.deduplicators {
		if found, err := deduplicator.New(dataSet); err != nil {
			return nil, err
		} else if found {
			return deduplicator, nil
		}
	}

	return nil, errors.New("deduplicator not found")
}

func (f *Factory) Get(dataSet *dataTypesUpload.Upload) (dataDeduplicator.Deduplicator, error) {
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	} else if err := structureValidator.New().WithOrigin(structure.OriginStore).Validate(dataSet); err != nil {
		return nil, errors.Wrap(err, "data set is invalid")
	}

	if dataSet.HasDeduplicatorName() {
		return f.get(dataSet)
	}

	return nil, nil
}

func (f *Factory) get(dataSet *dataTypesUpload.Upload) (dataDeduplicator.Deduplicator, error) {
	for _, deduplicator := range f.deduplicators {
		if found, err := deduplicator.Get(dataSet); err != nil {
			return nil, err
		} else if found {
			return deduplicator, nil
		}
	}

	return nil, errors.New("deduplicator not found")
}
