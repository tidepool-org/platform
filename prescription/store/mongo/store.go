package mongo

import (
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/prescription/store"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type Store struct {
	*storeStructuredMongo.Store
}

func NewStore(cfg *storeStructuredMongo.Config, lgr log.Logger) (*Store, error) {
	str, err := storeStructuredMongo.NewStore(cfg, lgr)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: str,
	}, nil
}

func (s *Store) EnsureIndexes() error {
	prescriptionTokenSession := s.prescriptionSession()
	defer prescriptionTokenSession.Close()
	return prescriptionTokenSession.EnsureIndexes()
}

func (s *Store) NewPrescriptionSession() store.PrescriptionSession {
	return s.prescriptionSession()
}

func (s *Store) prescriptionSession() *PrescriptionSession {
	return &PrescriptionSession{
		Session: s.Store.NewSession("prescriptions"),
	}
}
