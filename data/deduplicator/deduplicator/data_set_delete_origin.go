package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/pointer"
)

const (
	DataSetDeleteOriginName    = "org.tidepool.deduplicator.dataset.delete.origin"
	DataSetDeleteOriginVersion = "1.0.0"
)

type DataSetDeleteOrigin struct {
	*DataSetDeleteOriginBase
}

func NewDataSetDeleteOrigin(dependencies Dependencies) (*DataSetDeleteOrigin, error) {
	dataSetDeleteOriginDependencies := DataSetDeleteOriginDependencies{
		Dependencies: dependencies,
		DataFilter:   &dataSetDeleteOriginDataFilter{},
	}
	dataSetDeleteOriginBase, err := NewDataSetDeleteOriginBase(dataSetDeleteOriginDependencies, DataSetDeleteOriginName, DataSetDeleteOriginVersion)
	if err != nil {
		return nil, err
	}

	return &DataSetDeleteOrigin{
		DataSetDeleteOriginBase: dataSetDeleteOriginBase,
	}, nil
}

func (d *DataSetDeleteOrigin) New(ctx context.Context, dataSet *data.DataSet) (bool, error) {
	return d.Get(ctx, dataSet)
}

func (d *DataSetDeleteOrigin) Get(ctx context.Context, dataSet *data.DataSet) (bool, error) {
	if found, err := d.DataSetDeleteOriginBase.Get(ctx, dataSet); err != nil || found {
		return found, err
	}

	return dataSet.HasDeduplicatorNameMatch("org.tidepool.continuous.origin"), nil // TODO: DEPRECATED
}

type dataSetDeleteOriginDataFilter struct{}

func (d *dataSetDeleteOriginDataFilter) FilterData(ctx context.Context, dataSet *data.DataSet, dataSetData data.Data) (data.Data, error) {
	return dataSetData, nil
}

func (d *dataSetDeleteOriginDataFilter) GetDataSelectors(dataSetData data.Data) *data.Selectors {
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

func (d *dataSetDeleteOriginDataFilter) getDatumSelector(dataSetDatum data.Datum) *data.Selector {
	if origin := dataSetDatum.GetOrigin(); origin != nil && origin.ID != nil {
		return &data.Selector{
			Origin: &data.SelectorOrigin{
				ID: pointer.CloneString(origin.ID),
			},
		}
	}
	return nil
}
