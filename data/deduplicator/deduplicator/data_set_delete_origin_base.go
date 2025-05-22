package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/errors"
)

type DataSetDeleteOriginProvider interface {
	FilterData(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet, dataSetData data.Data) (data.Data, error)
	GetDataSelectors(datum data.Data) *data.Selectors
}

type DataSetDeleteOriginBase struct {
	*Base
	provider DataSetDeleteOriginProvider
}

func NewDataSetDeleteOriginBase(name string, version string, provider DataSetDeleteOriginProvider) (*DataSetDeleteOriginBase, error) {
	base, err := NewBase(name, version)
	if err != nil {
		return nil, err
	}

	return &DataSetDeleteOriginBase{
		Base:     base,
		provider: provider,
	}, nil
}

func (d *DataSetDeleteOriginBase) Open(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet) (*data.DataSet, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if repository == nil {
		return nil, errors.New("repository is missing")
	}
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	}

	if dataSet.HasDataSetTypeContinuous() {
		dataSet.Active = true
	}

	return d.Base.Open(ctx, repository, dataSet)
}

func (d *DataSetDeleteOriginBase) AddData(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet, dataSetData data.Data) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if repository == nil {
		return errors.New("repository is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}
	if dataSetData == nil {
		return errors.New("data set data is missing")
	}

	if dataSet.HasDataSetTypeContinuous() {
		dataSetData.SetActive(true)
	}

	var err error
	if dataSetData, err = d.provider.FilterData(ctx, repository, dataSet, dataSetData); err != nil {
		return err
	}

	if selectors := d.provider.GetDataSelectors(dataSetData); selectors != nil {
		if err := repository.DeleteDataSetData(ctx, dataSet, selectors); err != nil {
			return err
		}
		if err := d.Base.AddData(ctx, repository, dataSet, dataSetData); err != nil {
			return err
		}
		return repository.DestroyDeletedDataSetData(ctx, dataSet, selectors)
	}

	return d.Base.AddData(ctx, repository, dataSet, dataSetData)
}

func (d *DataSetDeleteOriginBase) DeleteData(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if repository == nil {
		return errors.New("repository is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}
	if selectors == nil {
		return errors.New("selectors is missing")
	}

	return repository.ArchiveDataSetData(ctx, dataSet, selectors)
}

func (d *DataSetDeleteOriginBase) Close(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if repository == nil {
		return errors.New("repository is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}

	if dataSet.HasDataSetTypeContinuous() {
		return nil
	}

	return d.Base.Close(ctx, repository, dataSet)
}
