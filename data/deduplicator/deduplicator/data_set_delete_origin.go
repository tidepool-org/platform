package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/pointer"
)

const (
	DataSetDeleteOriginName    = "org.tidepool.deduplicator.dataset.delete.origin"
	DataSetDeleteOriginVersion = "1.0.0"
)

type DataSetDeleteOrigin struct {
	*DataSetDeleteOriginBase
}

func NewDataSetDeleteOrigin() (*DataSetDeleteOrigin, error) {
	dataSetDeleteOriginBase, err := NewDataSetDeleteOriginBase(DataSetDeleteOriginName, DataSetDeleteOriginVersion, &dataSetDeleteOriginProvider{})
	if err != nil {
		return nil, err
	}

	return &DataSetDeleteOrigin{
		DataSetDeleteOriginBase: dataSetDeleteOriginBase,
	}, nil
}

func (d *DataSetDeleteOrigin) New(ctx context.Context, dataSet *dataTypesUpload.Upload) (bool, error) {
	return d.Get(ctx, dataSet)
}

func (d *DataSetDeleteOrigin) Get(ctx context.Context, dataSet *dataTypesUpload.Upload) (bool, error) {
	if found, err := d.DataSetDeleteOriginBase.Get(ctx, dataSet); err != nil || found {
		return found, err
	}

	return dataSet.HasDeduplicatorNameMatch("org.tidepool.continuous.origin"), nil // TODO: DEPRECATED
}

type dataSetDeleteOriginProvider struct{}

func (d *dataSetDeleteOriginProvider) FilterData(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload, dataSetData data.Data) (data.Data, error) {
	return dataSetData, nil
}

func (d *dataSetDeleteOriginProvider) GetDataSelectors(dataSetData data.Data) *data.Selectors {
	selectors := data.Selectors{}
	for _, dataSetDatum := range dataSetData {
		if selector := d.getDatumSelector(dataSetDatum); selector != nil {
			selectors = append(selectors, selector)
		}
	}
	if len(selectors) == 0 {
		return nil
	}
	return &selectors
}

func (d *dataSetDeleteOriginProvider) getDatumSelector(dataSetDatum data.Datum) *data.Selector {
	if origin := dataSetDatum.GetOrigin(); origin != nil && origin.ID != nil {
		return &data.Selector{
			Origin: &data.SelectorOrigin{
				ID: pointer.CloneString(origin.ID),
			},
		}
	}
	return nil
}
