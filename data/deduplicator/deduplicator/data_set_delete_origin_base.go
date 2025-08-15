package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
)

type DataSetDeleteOriginDataFilter interface {
	FilterData(ctx context.Context, dataSet *data.DataSet, dataSetData data.Data) (data.Data, error)
	GetDataSelectors(datum data.Data) *data.Selectors
}

type DataSetDeleteOriginDependencies struct {
	Dependencies
	DataFilter DataSetDeleteOriginDataFilter
}

func (d DataSetDeleteOriginDependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.DataFilter == nil {
		return errors.New("data filter is missing")
	}
	return nil
}

type DataSetDeleteOriginBase struct {
	*Base
	DataFilter DataSetDeleteOriginDataFilter
}

func NewDataSetDeleteOriginBase(dependencies DataSetDeleteOriginDependencies, name string, version string) (*DataSetDeleteOriginBase, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	base, err := NewBase(dependencies.Dependencies, name, version)
	if err != nil {
		return nil, err
	}

	return &DataSetDeleteOriginBase{
		Base:       base,
		DataFilter: dependencies.DataFilter,
	}, nil
}

func (d *DataSetDeleteOriginBase) AddData(ctx context.Context, dataSet *data.DataSet, dataSetData data.Data) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}
	if dataSetData == nil {
		return errors.New("data set data is missing")
	}

	var err error
	if dataSetData, err = d.DataFilter.FilterData(ctx, dataSet, dataSetData); err != nil {
		return err
	}

	if selectors := d.DataFilter.GetDataSelectors(dataSetData); selectors != nil {
		if err := d.DataStore.DeleteDataSetData(ctx, dataSet, selectors); err != nil {
			return err
		}
		if err := d.Base.AddData(ctx, dataSet, dataSetData); err != nil {
			return err
		}
		return d.DataStore.DestroyDeletedDataSetData(ctx, dataSet, selectors)
	}

	return d.Base.AddData(ctx, dataSet, dataSetData)
}

func (d *DataSetDeleteOriginBase) DeleteData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}
	if selectors == nil {
		return errors.New("selectors is missing")
	}

	return d.DataStore.ArchiveDataSetData(ctx, dataSet, selectors)
}
