package deduplicator

import (
	"context"

	"github.com/tidepool-org/platform/data"
)

const (
	NoneName    = "org.tidepool.deduplicator.none"
	NoneVersion = "1.0.0"
)

type None struct {
	*Base
}

func NewNone(dependencies Dependencies) (*None, error) {
	base, err := NewBase(dependencies, NoneName, NoneVersion)
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
