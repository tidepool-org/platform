package mongoofficial

import (
	"context"
	"fmt"

	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/errors"
)

var StoreModule = fx.Provide(
	NewConfig,
	NewStore,
)

type Store struct {
	client *mongoDriver.Client
	config *Config
	ctx    context.Context
}

type Status struct {
	Ping string
}

type Params struct {
	fx.In

	DatabaseConfig *Config

	Lifecycle fx.Lifecycle
}

func NewStore(p Params) (*Store, error) {
	if p.DatabaseConfig == nil {
		return nil, errors.New("database config is empty")
	}

	store := &Store{
		config: p.DatabaseConfig,
		ctx:    context.Background(),
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return store.Initialize(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return store.Terminate(ctx)
		},
	})

	return store, nil
}

func (o *Store) Initialize(ctx context.Context) error {
	cs := o.config.AsConnectionString()
	clientOptions := options.Client().ApplyURI(cs)
	mongoClient, err := mongoDriver.Connect(context.Background(), clientOptions)
	if err != nil {
		return errors.Wrap(err, "connection options are invalid")
	}

	o.client = mongoClient
	return o.Ping(ctx)
}

func (o *Store) GetRepository(collection string) *Repository {
	return NewRepository(o.GetCollection(collection))
}

func (o *Store) GetCollection(collection string) *mongoDriver.Collection {
	db := o.client.Database(o.config.Database)
	prefixed := fmt.Sprintf("%s%s", o.config.CollectionPrefix, collection)
	return db.Collection(prefixed)
}

func (o *Store) Ping(ctx context.Context) error {
	return o.client.Ping(ctx, readpref.Primary())
}

func (o *Store) Status(ctx context.Context) *Status {
	status := &Status{
		Ping: "FAILED",
	}

	if o.Ping(ctx) == nil {
		status.Ping = "OK"
	}

	return status
}

func (o *Store) Terminate(ctx context.Context) error {
	return o.client.Disconnect(ctx)
}
