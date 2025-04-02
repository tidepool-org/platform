package deduplicator

import (
	"context"
	"slices"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataTypesBolus "github.com/tidepool-org/platform/data/types/bolus"
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/pointer"
)

const (
	DataSetDeleteOriginOlderName    = "org.tidepool.deduplicator.dataset.delete.origin.older"
	DataSetDeleteOriginOlderVersion = "1.0.0"
)

type DataSetDeleteOriginOlder struct {
	*DataSetDeleteOriginBase
}

func NewDataSetDeleteOriginOlder() (*DataSetDeleteOriginOlder, error) {
	dataSetDeleteOriginBase, err := NewDataSetDeleteOriginBase(DataSetDeleteOriginOlderName, DataSetDeleteOriginOlderVersion, &dataSetDeleteOriginOlderProvider{})
	if err != nil {
		return nil, err
	}

	return &DataSetDeleteOriginOlder{
		DataSetDeleteOriginBase: dataSetDeleteOriginBase,
	}, nil
}

type dataSetDeleteOriginOlderProvider struct{}

func (d *dataSetDeleteOriginOlderProvider) FilterData(ctx context.Context, repository dataStore.DataRepository, dataSet *dataTypesUpload.Upload, dataSetData data.Data) (data.Data, error) {
	filterableDataSetData := dataSetData.Filter(func(datum data.Datum) bool {
		return slices.Contains(filterableDataSetDataTypes, datum.GetType())
	})

	if selectors := d.GetDataSelectors(filterableDataSetData); selectors != nil {
		if existingSelectors, err := repository.NewerDataSetData(ctx, dataSet, selectors); err != nil {
			return nil, err
		} else if existingSelectors != nil && len(*existingSelectors) > 0 {
			dataSetData = dataSetData.Filter(func(datum data.Datum) bool {
				return !slices.ContainsFunc(*existingSelectors, d.getDatumSelector(datum).Includes)
			})
		}
	}

	return dataSetData, nil
}

func (d *dataSetDeleteOriginOlderProvider) GetDataSelectors(dataSetData data.Data) *data.Selectors {
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

func (d *dataSetDeleteOriginOlderProvider) getDatumSelector(dataSetDatum data.Datum) *data.Selector {
	if origin := dataSetDatum.GetOrigin(); origin != nil && origin.ID != nil {
		return &data.Selector{
			Origin: &data.SelectorOrigin{
				ID:   pointer.CloneString(origin.ID),
				Time: origin.Time,
			},
		}
	}
	return nil
}

var filterableDataSetDataTypes = []string{
	dataTypesBolus.Type,
	dataTypesFood.Type,
}
