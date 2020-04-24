package mongo

import (
	"context"

	"go.uber.org/fx"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/prescription/store"
	"github.com/tidepool-org/platform/status"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type PrescriptionStore struct {
	*storeStructuredMongo.Store
}

type Params struct {
	fx.In

	ConfigReporter config.Reporter
	Logger         log.Logger

	Lifestyle fx.Lifecycle
}

// fx.Provide() requires explicit conversion to the status reporter interface
func NewStoreStatusReporter(str store.Store) status.StoreStatusReporter {
	return str
}

func NewStore(p Params) (store.Store, error) {
	p.Logger.Debug("Initializing prescription store")
	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(p.ConfigReporter.WithScopes("prescription", "store")); err != nil {
		return nil, errors.Wrap(err, "unable to load prescription store config")
	}

	str, err := storeStructuredMongo.NewStore(cfg, p.Logger)
	if err != nil {
		return nil, err
	}

	prescriptionStore := &PrescriptionStore{
		Store: str,
	}

	p.Lifestyle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			p.Logger.Debug("Creating prescription store indexes")
			if err := prescriptionStore.EnsureIndexes(); err != nil {
				return errors.Wrap(err, "unable to ensure prescription store indexes")
			}
			return nil
		},
		OnStop: func(context.Context) error {
			p.Logger.Debug("Terminating prescription store")
			if err := prescriptionStore.Close(); err != nil {
				p.Logger.WithError(err).Warn("Unable to terminate prescription store")
				return err
			}
			return nil
		},
	})

	return prescriptionStore, nil
}

func (s *PrescriptionStore) EnsureIndexes() error {
	session := s.prescriptionSession()
	defer session.Close()
	return session.EnsureIndexes()
}

func (s *PrescriptionStore) NewPrescriptionSession() store.PrescriptionSession {
	return s.prescriptionSession()
}

func (s *PrescriptionStore) prescriptionSession() *PrescriptionSession {
	return &PrescriptionSession{
		Session: s.Store.NewSession("prescriptions"),
	}
}
