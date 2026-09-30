package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataTypes "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
)

const (
	DataSetDropHashReplaceIDName    = "org.tidepool.deduplicator.dataset.drop.hash.replace.id"
	DataSetDropHashReplaceIDVersion = "1.0.0"
)

// DataSetDropHashReplaceID drops a datum whose identity hash is already in the data set, as DataSetDropHash does,
// unless the datum has the id of the datum there with that hash: then it replaces that datum. A client that
// updates a datum it read from the data set, such as a duration that was not known then, gives the update its id.
type DataSetDropHashReplaceID struct {
	*Base
}

func NewDataSetDropHashReplaceID(dependencies Dependencies) (*DataSetDropHashReplaceID, error) {
	base, err := NewBase(dependencies, DataSetDropHashReplaceIDName, DataSetDropHashReplaceIDVersion)
	if err != nil {
		return nil, err
	}

	return &DataSetDropHashReplaceID{
		Base: base,
	}, nil
}

func (d *DataSetDropHashReplaceID) AddData(ctx context.Context, dataSet *data.DataSet, dataSetData data.Data) error {
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

	// Drop data already in the data set before duplicates, so a replacement wins over a duplicate without its id
	var replacedSelectors data.Selectors
	if selectors := MapDataSetDataToSelectors(dataSetData, GetDatumDeduplicatorSelector); selectors != nil {
		existingSelectors, err := d.DataStore.ExistingDataSetData(ctx, dataSet, selectors)
		if err != nil {
			return err
		} else if existingSelectorsCount := len(*existingSelectors); existingSelectorsCount > 0 {

			existingSelectorsMap := make(map[string]*data.Selector, existingSelectorsCount)
			for _, existingSelector := range *existingSelectors {
				existingSelectorsMap[*existingSelector.Deduplicator.Hash] = existingSelector
			}

			dataSetData = dataSetData.Filter(func(datum data.Datum) bool {
				datumHash := GetDatumDeduplicatorHash(datum)
				if datumHash == nil {
					return true
				}
				existingSelector, ok := existingSelectorsMap[*datumHash]
				if !ok {
					return true
				}
				if datumID := datum.GetID(); datumID != nil && existingSelector.ID != nil && *datumID == *existingSelector.ID {
					replacedSelectors = append(replacedSelectors, &data.Selector{ID: pointer.CloneString(datumID)})
					return true
				}
				return false
			})
		}
	}

	dataSetData = DeduplicateDataSetDataByIdentity(dataSetData, GetDatumDeduplicatorHash)

	if len(replacedSelectors) == 0 {
		return d.Base.AddData(ctx, dataSet, dataSetData)
	}

	// Delete before adding, as the replacement has the same id, then destroy only what was deleted
	if err := d.DataStore.DeleteDataSetData(ctx, dataSet, &replacedSelectors); err != nil {
		return err
	}
	if err := d.Base.AddData(ctx, dataSet, dataSetData); err != nil {
		return err
	}
	return d.DataStore.DestroyDeletedDataSetData(ctx, dataSet, &replacedSelectors)
}
