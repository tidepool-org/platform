package factory

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataDeduplicator "github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/errors"
)

type Deduplicator interface {
	dataDeduplicator.Deduplicator

	New(ctx context.Context, dataSet *data.DataSet) (bool, error)
	Get(ctx context.Context, dataSet *data.DataSet) (bool, error)
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

func (f *Factory) New(ctx context.Context, dataSet *data.DataSet) (dataDeduplicator.Deduplicator, error) {
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	}

	if dataSet.HasDeduplicatorName() {
		return f.get(ctx, dataSet)
	}

	for _, deduplicator := range f.deduplicators {
		if found, err := deduplicator.New(ctx, dataSet); err != nil {
			return nil, err
		} else if found {
			return deduplicator, nil
		}
	}

	return nil, errors.New("deduplicator not found")
}

func (f *Factory) Get(ctx context.Context, dataSet *data.DataSet) (dataDeduplicator.Deduplicator, error) {
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	}

	if dataSet.HasDeduplicatorName() {
		return f.get(ctx, dataSet)
	}

	return nil, nil
}

func (f *Factory) get(ctx context.Context, dataSet *data.DataSet) (dataDeduplicator.Deduplicator, error) {
	for _, deduplicator := range f.deduplicators {
		if found, err := deduplicator.Get(ctx, dataSet); err != nil {
			return nil, err
		} else if found {
			return deduplicator, nil
		}
	}

	return nil, errors.New("deduplicator not found")
}
