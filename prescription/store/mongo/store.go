package mongo

import (
	"context"

	"go.uber.org/fx"

	structuredMongo "github.com/tidepool-org/platform/store/structured/mongo"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/prescription/store"
	"github.com/tidepool-org/platform/status"
)

type PrescriptionStore struct {
	*structuredMongo.Store

	logger log.Logger
}

type Params struct {
	fx.In

	Logger log.Logger
	Store  *structuredMongo.Store

	Lifestyle fx.Lifecycle
}

// NewStatusReporter explicitly casts the store to status.StoreStatusReporter
// as required by fx.Provide()
func NewStatusReporter(str store.Store) status.StoreStatusReporter {
	return str
}

func NewStore(p Params) (store.Store, error) {
	if p.Logger == nil {
		return nil, errors.New("logger is missing")
	}
	if p.Store == nil {
		return nil, errors.New("store is missing")
	}

	prescriptionStore := NewStoreFromBase(p.Store, p.Logger)

	p.Lifestyle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return prescriptionStore.CreateIndexes(ctx)
		},
	})

	return prescriptionStore, nil
}

func NewStoreFromBase(base *structuredMongo.Store, logger log.Logger) *PrescriptionStore {
	return &PrescriptionStore{
		logger: logger,
		Store:  base,
	}
}

func (p *PrescriptionStore) CreateIndexes(ctx context.Context) error {
	p.logger.Debug("creating prescriptions repository indexes")
	return p.GetPrescriptionRepository().CreateIndexes(ctx)
}

func (p *PrescriptionStore) GetPrescriptionRepository() store.PrescriptionRepository {
	return &PrescriptionRepository{
		Repository: p.Store.GetRepository("prescriptions"),
	}
}

func (p *PrescriptionStore) Status(ctx context.Context) interface{} {
	return p.Store.Status(ctx)
}
