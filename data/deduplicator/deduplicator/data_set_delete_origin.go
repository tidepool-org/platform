package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
)

const DataSetDeleteOriginName = "org.tidepool.deduplicator.dataset.delete.origin"

type DataSetDeleteOrigin struct {
	*Base
}

func NewDataSetDeleteOrigin() (*DataSetDeleteOrigin, error) {
	base, err := NewBase(DataSetDeleteOriginName, "1.0.0")
	if err != nil {
		return nil, err
	}

	return &DataSetDeleteOrigin{
		Base: base,
	}, nil
}

func (d *DataSetDeleteOrigin) New(ctx context.Context, dataSet *data.DataSet) (bool, error) {
	return d.Get(ctx, dataSet)
}

func (d *DataSetDeleteOrigin) Get(ctx context.Context, dataSet *data.DataSet) (bool, error) {
	if found, err := d.Base.Get(ctx, dataSet); err != nil || found {
		return found, err
	}

	return dataSet.HasDeduplicatorNameMatch("org.tidepool.continuous.origin"), nil // TODO: DEPRECATED
}

func (d *DataSetDeleteOrigin) Open(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet) (*data.DataSet, error) {
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

func (d *DataSetDeleteOrigin) AddData(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet, dataSetData data.Data) error {
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

	if selectors := d.getSelectors(dataSetData); selectors != nil {
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

func (d *DataSetDeleteOrigin) DeleteData(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet, selectors *data.Selectors) error {
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

func (d *DataSetDeleteOrigin) Close(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet) error {
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

func (d *DataSetDeleteOrigin) getSelectors(dataSetData data.Data) *data.Selectors {
	selectors := data.Selectors{}
	for _, dataSetDatum := range dataSetData {
		if origin := dataSetDatum.GetOrigin(); origin != nil && origin.ID != nil {
			selector := &data.Selector{
				Origin: &data.SelectorOrigin{
					ID: pointer.CloneString(origin.ID),
				},
			}
			selectors = append(selectors, selector)
		}
	}
	if len(selectors) == 0 {
		return nil
	}
	return &selectors
}
