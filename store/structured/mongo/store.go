package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/tidepool-org/platform/errors"
)

type Store struct {
	client          *mongoDriver.Client
	config          *Config
	initializeGroup sync.WaitGroup
	closingChannel  chan bool
	clientMux       sync.Mutex
}

type Status struct {
	Ping string
}

func (s *Store) getClient() *mongoDriver.Client {
	s.clientMux.Lock()
	defer s.clientMux.Unlock()
	return s.client
}
func (s *Store) setClient(cli *mongoDriver.Client) {
	s.clientMux.Lock()
	defer s.clientMux.Unlock()
	s.client = cli
}

func (s *Store) Close() error {
	if s.closingChannel != nil {
		s.closingChannel <- true
	}
	s.initializeGroup.Wait()
	return s.getClient().Disconnect(context.Background())
}

func NewStore(c *Config) (*Store, error) {
	if c == nil {
		return nil, errors.New("database config is empty")
	} else if err := c.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	store := &Store{
		config: c,
	}
	store.Start()
	return store, nil
}

func (s *Store) Start() {
	if s.closingChannel == nil {
		s.initializeGroup.Add(1)
		go s.connectionRoutine()
	}
}

func (s *Store) connectionRoutine() {

	var attempts int64 = 1
	var err error
	s.closingChannel = make(chan bool, 1)
	for {
		var timer <-chan time.Time
		if attempts == int64(0) {
			timer = time.After(0)
		} else {
			timer = time.After(s.config.WaitConnectionInterval)
		}
		select {
		case <-s.closingChannel:
			close(s.closingChannel)
			s.closingChannel = nil
			s.initializeGroup.Done()
			return
		case <-timer:
			err = initConnexion(s)
			if err == nil {
				s.closingChannel <- true
			} else {
				if s.config.MaxConnectionAttempts > 0 && s.config.MaxConnectionAttempts <= attempts {
					s.closingChannel <- true
					panic(err)
				} else {
					attempts++
				}
			}
		}
	}
}

func initConnexion(store *Store) error {
	var err error
	cs := store.config.AsConnectionString()
	clientOptions := options.Client().
		ApplyURI(cs).
		SetConnectTimeout(store.config.Timeout).
		SetServerSelectionTimeout(store.config.Timeout)
	store.client, err = mongoDriver.Connect(context.Background(), clientOptions)
	if err != nil {
		fmt.Println("connection options are invalid")
		return errors.Wrap(err, "connection options are invalid")
	}
	ctx, cancel := context.WithTimeout(context.Background(), store.config.Timeout)
	defer cancel()
	err = store.client.Ping(ctx, readpref.PrimaryPreferred())
	if err != nil {
		fmt.Println("cannot ping store")
		return errors.Wrap(err, "cannot ping store")
	}

	store.createIndexesFromConfig()
	return nil
}

func (s *Store) createIndexesFromConfig() {
	if s.config.Indexes != nil {
		for collection, idxs := range s.config.Indexes {
			repository := s.GetRepository(collection)
			err := repository.CreateAllIndexes(context.Background(), idxs)
			if err != nil {
				fmt.Printf("unable to ensure indexes on %s : %v", collection, err)
			}
		}
	}
}

func (s *Store) WaitUntilStarted() {
	s.initializeGroup.Wait()
}

func (s *Store) GetRepository(collection string) *Repository {
	return NewRepository(s.GetCollectionWithArchive(collection))
}

func (s *Store) GetCollectionWithArchive(collection string) (*mongoDriver.Collection, *mongoDriver.Collection) {
	db := s.getClient().Database(s.config.Database)
	prefixed := fmt.Sprintf("%s%s", s.config.CollectionPrefix, collection)
	prefixedArchive := fmt.Sprintf("%s%s_archive", s.config.CollectionPrefix, collection)
	return db.Collection(prefixed), db.Collection(prefixedArchive)
}

func (s *Store) GetCollection(collection string) *mongoDriver.Collection {
	db := s.getClient().Database(s.config.Database)
	prefixed := fmt.Sprintf("%s%s", s.config.CollectionPrefix, collection)
	return db.Collection(prefixed)
}

func (s *Store) Ping(ctx context.Context) error {
	if s.getClient() == nil {
		return errors.New("store has not been initialized")
	}

	return s.getClient().Ping(ctx, readpref.Primary())
}

func (s *Store) Status(ctx context.Context) *Status {
	status := &Status{
		Ping: "FAILED",
	}

	if s.Ping(ctx) == nil {
		status.Ping = "OK"
	}

	return status
}

func (s *Store) Terminate(ctx context.Context) error {
	if s.getClient() == nil {
		return nil
	}

	return s.getClient().Disconnect(ctx)
}
