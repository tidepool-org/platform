package mongo

import (
	"context"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/prescription/store"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"go.uber.org/fx"
)

type Store struct {
	*storeStructuredMongo.Store
}

type Params struct {
	fx.In

	ConfigReporter config.Reporter
	Logger log.Logger

	Lifestyle fx.Lifecycle
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

	prescriptionStore := &Store{
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

func (s *Store) EnsureIndexes() error {
	session := s.prescriptionSession()
	defer session.Close()
	return session.EnsureIndexes()
}

func (s *Store) NewPrescriptionSession() store.PrescriptionSession {
	return s.prescriptionSession()
}

func (s *Store) prescriptionSession() *PrescriptionSession {
	return &PrescriptionSession{
		Session: s.Store.NewSession("prescriptions"),
	}
}
