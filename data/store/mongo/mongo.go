package mongo

import (
	"github.com/tidepool-org/platform/data/store"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func NewStore(config *storeStructuredMongo.Config) (*Store, error) {
	baseStore, err := storeStructuredMongo.NewStore(config)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: baseStore,
	}, nil
}

type Store struct {
	*storeStructuredMongo.Store
}

func (s *Store) EnsureIndexes() error {
	dataRepository := s.NewDataRepository()
	summaryRepository := s.NewSummaryRepository()

	err := dataRepository.EnsureIndexes()
	if err != nil {
		return err
	}

	err = summaryRepository.EnsureIndexes()

	return err
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
