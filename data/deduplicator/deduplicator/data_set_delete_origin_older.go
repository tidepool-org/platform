package deduplicator

import (
	"context"
	"slices"

	"github.com/tidepool-org/platform/data"
	dataTypesBolus "github.com/tidepool-org/platform/data/types/bolus"
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/pointer"
)

const (
	DataSetDeleteOriginOlderName    = "org.tidepool.deduplicator.dataset.delete.origin.older"
	DataSetDeleteOriginOlderVersion = "1.0.0"
)

type DataSetDeleteOriginOlder struct {
	*DataSetDeleteOriginBase
}

func NewDataSetDeleteOriginOlder(dependencies Dependencies) (*DataSetDeleteOriginOlder, error) {
	dataSetDeleteOriginDependencies := DataSetDeleteOriginDependencies{
		Dependencies: dependencies,
		DataFilter: &dataSetDeleteOriginOlderDataFilter{
			dataStore: dependencies.DataStore,
		},
	}
	dataSetDeleteOriginBase, err := NewDataSetDeleteOriginBase(dataSetDeleteOriginDependencies, DataSetDeleteOriginOlderName, DataSetDeleteOriginOlderVersion)
	if err != nil {
		return nil, err
	}

	return &DataSetDeleteOriginOlder{
		DataSetDeleteOriginBase: dataSetDeleteOriginBase,
	}, nil
}

type dataSetDeleteOriginOlderDataFilter struct {
	dataStore DataStore
}

func (d *dataSetDeleteOriginOlderDataFilter) FilterData(ctx context.Context, dataSet *data.DataSet, dataSetData data.Data) (data.Data, error) {
	filterableDataSetData := dataSetData.Filter(func(datum data.Datum) bool {
		return slices.Contains(FilterableDataSetDataTypes, datum.GetType())
	})

	if selectors := MapDataSetDataToSelectors(filterableDataSetData, d.getDatumSelector); selectors != nil {
		existingSelectors, err := d.dataStore.ExistingDataSetData(ctx, dataSet, selectors)
		if err != nil {
			return nil, err
		}

		existingSelectorsMap := make(map[string]*data.Selector, len(*existingSelectors))
		for _, existingSelector := range *existingSelectors {
			existingSelectorsMap[*existingSelector.Origin.ID] = existingSelector
		}

		dataSetData = dataSetData.Filter(func(datum data.Datum) bool {
			if datumSelector := d.getDatumSelector(datum); datumSelector != nil {
				if existingSelector, ok := existingSelectorsMap[*datumSelector.Origin.ID]; ok {
					return datumSelector.NewerThan(existingSelector)
				}
			}
			return true
		})
	}

	return dataSetData, nil
}

func (d *dataSetDeleteOriginOlderDataFilter) GetDataSelectors(dataSetData data.Data) *data.Selectors {
	return MapDataSetDataToSelectors(dataSetData, d.getDatumSelector)
}

func (d *dataSetDeleteOriginOlderDataFilter) getDatumSelector(dataSetDatum data.Datum) *data.Selector {
	if origin := dataSetDatum.GetOrigin(); origin != nil && origin.ID != nil {
		return &data.Selector{
			Origin: &data.SelectorOrigin{
				ID:   pointer.CloneString(origin.ID),
				Time: pointer.CloneString(origin.Time),
			},
		}
	}
	return nil
}

var FilterableDataSetDataTypes = []string{
	dataTypesBolus.Type,
	dataTypesFood.Type,
}
