package mongo

import (
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type Store struct {
	*storeStructuredMongo.Store
}

func NewStore(config *storeStructuredMongo.Config) (*Store, error) {
	store, err := storeStructuredMongo.NewStore(config)
	if err != nil {
		return nil, err
	}
	return NewStoreFromBase(store), nil
}

func NewStoreFromBase(base *storeStructuredMongo.Store) *Store {
	return &Store{Store: base}
}

func (s *Store) EnsureIndexes() error {
	repository := s.newRepository()
	err := repository.EnsureIndexes()
	if err != nil {
		return err
	}
	deviceLogsrepository := s.newDeviceLogsRepository()
	return deviceLogsrepository.EnsureIndexes()
}

func (s *Store) NewBlobRepository() blobStoreStructured.BlobRepository {
	return s.newRepository()
}

func (s *Store) newRepository() *BlobRepository {
	return &BlobRepository{
		s.Store.GetRepository("blobs"),
	}
}

func (s *Store) NewDeviceLogsRepository() blobStoreStructured.DeviceLogsRepository {
	return s.newDeviceLogsRepository()
}

func (s *Store) newDeviceLogsRepository() *DeviceLogsRepository {
	return &DeviceLogsRepository{
		s.Store.GetRepository("deviceLogs"),
	}
}
