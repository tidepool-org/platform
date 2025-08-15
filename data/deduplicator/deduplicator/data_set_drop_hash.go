package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataTypes "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
)

const (
	DataSetDropHashName    = "org.tidepool.deduplicator.dataset.drop.hash"
	DataSetDropHashVersion = "1.0.0"
)

type DataSetDropHash struct {
	*Base
}

func NewDataSetDropHash(dependencies Dependencies) (*DataSetDropHash, error) {
	base, err := NewBase(dependencies, DataSetDropHashName, DataSetDropHashVersion)
	if err != nil {
		return nil, err
	}

	return &DataSetDropHash{
		Base: base,
	}, nil
}

func (d *DataSetDropHash) AddData(ctx context.Context, dataSet *data.DataSet, dataSetData data.Data) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}
	if dataSetData == nil {
		return errors.New("data set data is missing")
	}

	dataSetData.SetUserID(dataSet.UserID)
	dataSetData.SetDataSetID(dataSet.ID)

	if err := AssignDataSetDataIdentityHashes(dataSetData, dataTypes.IdentityFieldsVersionDataSetID); err != nil {
		return err
	}

	if selectors := MapDataSetDataToSelectors(dataSetData, d.getDatumSelector); selectors != nil {
		existingSelectors, err := d.DataStore.ExistingDataSetData(ctx, dataSet, selectors)
		if err != nil {
			return err
		}

		existingSelectorsMap := make(map[string]*data.Selector, len(*existingSelectors))
		for _, existingSelector := range *existingSelectors {
			existingSelectorsMap[*existingSelector.Deduplicator.Hash] = existingSelector
		}

		dataSetData = dataSetData.Filter(func(datum data.Datum) bool {
			if datumSelector := d.getDatumSelector(datum); datumSelector != nil {
				_, ok := existingSelectorsMap[*datumSelector.Deduplicator.Hash]
				return !ok
			}
			return true
		})
	}

	return d.Base.AddData(ctx, dataSet, dataSetData)
}

func (d *DataSetDropHash) getDatumSelector(dataSetDatum data.Datum) *data.Selector {
	if deduplicator := dataSetDatum.DeduplicatorDescriptor(); deduplicator != nil && deduplicator.Hash != nil {
		return &data.Selector{
			Deduplicator: &data.SelectorDeduplicator{
				Hash: pointer.CloneString(deduplicator.Hash),
			},
		}
	}
	return nil
}
