package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/errors"
)

const NoneName = "org.tidepool.deduplicator.none"
const NoneVersion = "1.0.0"

type None struct {
	*Base
}

func NewNone() (*None, error) {
	base, err := NewBase(NoneName, NoneVersion)
	if err != nil {
		return nil, err
	}

	return &None{
		Base: base,
	}, nil
}

func (n *None) New(ctx context.Context, dataSet *data.DataSet) (bool, error) {
	return n.Get(ctx, dataSet)
}

func (n *None) Get(ctx context.Context, dataSet *data.DataSet) (bool, error) {
	if found, err := n.Base.Get(ctx, dataSet); err != nil || found {
		return found, err
	}

	return dataSet.HasDeduplicatorNameMatch("org.tidepool.continuous"), nil // TODO: DEPRECATED
}

func (n *None) Open(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet) (*data.DataSet, error) {
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

	return n.Base.Open(ctx, repository, dataSet)
}

func (n *None) AddData(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet, dataSetData data.Data) error {
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

	return n.Base.AddData(ctx, repository, dataSet, dataSetData)
}

func (n *None) Close(ctx context.Context, repository dataStore.DataRepository, dataSet *data.DataSet) error {
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

	return n.Base.Close(ctx, repository, dataSet)
}
