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

	configReporter config.Reporter
	logger         log.Logger
}

type Params struct {
	fx.In

	ConfigReporter config.Reporter
	Logger         log.Logger

	Lifestyle fx.Lifecycle
}

// NewStoreStatusReporter explicitly casts the store to status.StoreStatusReporter
// as required by fx.Provide()
func NewStoreStatusReporter(str store.Store) status.StoreStatusReporter {
	return str
}

func NewStore(p Params) (store.Store, error) {
	if p.Logger == nil {
		return nil, errors.New("logger is missing")
	}
	if p.ConfigReporter == nil {
		return nil, errors.New("config reporter is missing")
	}

	prescriptionStore := &PrescriptionStore{
		configReporter: p.ConfigReporter,
		logger:         p.Logger,
		Store:          nil,
	}

	p.Lifestyle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			return prescriptionStore.Initialize()
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

func (p *PrescriptionStore) Initialize() error {
	p.logger.Debug("Initializing prescription store")
	cfg := storeStructuredMongo.NewConfig()
	if err := cfg.Load(p.configReporter.WithScopes("prescription", "store")); err != nil {
		return errors.Wrap(err, "unable to load prescription store config")
	}

	str, err := storeStructuredMongo.NewStore(cfg, p.logger)
	if err != nil {
		return err
	}

	p.Store = str

	p.logger.Debug("Creating prescription store indexes")
	if err := p.EnsureIndexes(); err != nil {
		return errors.Wrap(err, "unable to ensure prescription store indexes")
	}
	return nil
}

func (p *PrescriptionStore) EnsureIndexes() error {
	session := p.prescriptionSession()
	defer session.Close()
	return session.EnsureIndexes()
}

func (p *PrescriptionStore) NewPrescriptionSession() store.PrescriptionSession {
	return p.prescriptionSession()
}

func (p *PrescriptionStore) prescriptionSession() *PrescriptionSession {
	return &PrescriptionSession{
		Session: p.Store.NewSession("prescriptions"),
	}
}
