package mongo

import (
	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/data/store"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func NewStore(config *storeStructuredMongo.Config) (*Store, error) {
	baseStore, err := storeStructuredMongo.NewStore(config)
	if err != nil {
		return nil, err
	}

	return NewStoreFromBase(baseStore), nil
}

func NewStoreFromBase(base *storeStructuredMongo.Store) *Store {
	return &Store{
		Store: base,
	}
}

type Store struct {
	*storeStructuredMongo.Store
}

func (s *Store) EnsureIndexes() error {
	dataRepository := s.NewDataRepository()
	summaryRepository := s.NewSummaryRepository()
	alertsRepository := s.NewAlertsRepository()

	if err := dataRepository.EnsureIndexes(); err != nil {
		return err
	}

	if err := summaryRepository.EnsureIndexes(); err != nil {
		return err
	}

	if err := alertsRepository.EnsureIndexes(); err != nil {
		return err
	}

	return nil
}

func (s *Store) NewDataRepository() store.DataRepository {
	return &DataRepository{
		DatumRepository: &DatumRepository{
			s.Store.GetRepository("deviceData"),
		},
		DataSetRepository: &DataSetRepository{
			s.Store.GetRepository("deviceDataSets"),
		},
	}
}

func (s *Store) NewSummaryRepository() store.SummaryRepository {
	return &SummaryRepository{
		s.Store.GetRepository("summary"),
	}
}

func (s *Store) NewBucketsRepository() store.BucketsRepository {
	return &BucketsRepository{
		s.Store.GetRepository("buckets"),
	}
}

func (s *Store) NewAlertsRepository() alerts.Repository {
	r := alertsRepo(*s.Store.GetRepository("alerts"))
	return &r
}
