package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
)

const NoneName = "org.tidepool.deduplicator.none"

type None struct {
	*Base
}

func NewNone() (*None, error) {
	base, err := NewBase(NoneName, "1.0.0")
	if err != nil {
		return nil, err
	}

	return &None{
		Base: base,
	}, nil
}

func (n *None) New(dataSet *dataTypesUpload.Upload) (bool, error) {
	return n.Get(dataSet)
}

func (n *None) Get(dataSet *dataTypesUpload.Upload) (bool, error) {
	if found, err := n.Base.Get(dataSet); err != nil || found {
		return found, err
	}

	return dataSet.HasDeduplicatorNameMatch("org.tidepool.continuous"), nil // TODO: DEPRECATED
}

func (n *None) Open(ctx context.Context, session dataStore.DataRepository, dataSet *dataTypesUpload.Upload) (*dataTypesUpload.Upload, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if session == nil {
		return nil, errors.New("session is missing")
	}
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	}

	if dataSet.HasDataSetTypeContinuous() {
		dataSet.SetActive(true)
	}

	return n.Base.Open(ctx, session, dataSet)
}

func (n *None) AddData(ctx context.Context, session dataStore.DataRepository, dataSet *dataTypesUpload.Upload, dataSetData data.Data) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if session == nil {
		return errors.New("session is missing")
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

	return n.Base.AddData(ctx, session, dataSet, dataSetData)
}

func (n *None) Close(ctx context.Context, session dataStore.DataRepository, dataSet *dataTypesUpload.Upload) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if session == nil {
		return errors.New("session is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}

	if dataSet.HasDataSetTypeContinuous() {
		return nil
	}

	return n.Base.Close(ctx, session, dataSet)
}
