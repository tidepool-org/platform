package mongo

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
	}

	var err error

	if p.Lifecycle != nil {
		p.Lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return store.Initialize(ctx)
			},
			OnStop: func(ctx context.Context) error {
				return store.Terminate(ctx)
			},
		})
	} else {
		// If we're not using `fx`, then do the Initialization as part of `NewStore`
		err = store.Initialize(context.Background())
	}

	return store, err
}

func (o *Store) Initialize(ctx context.Context) error {
	cs := o.config.AsConnectionString()
	clientOptions := options.Client().
		ApplyURI(cs).
		SetConnectTimeout(o.config.Timeout).
		SetServerSelectionTimeout(o.config.Timeout)
	mongoClient, err := mongoDriver.Connect(context.Background(), clientOptions)
	if err != nil {
		return errors.Wrap(err, "connection options are invalid")
	}

	o.client = mongoClient
	return o.Ping(ctx)
}

func (o *Store) GetRepository(collection string) *Repository {
	if o.client == nil {
		// TODO: TK - should this return an error instead?
		return nil
	}

	return NewRepository(o.GetCollection(collection))
}

func (o *Store) GetCollection(collection string) *mongoDriver.Collection {
	if o.client == nil {
		// TODO: TK - should this return an error instead?
		return nil
	}

	db := o.client.Database(o.config.Database)
	prefixed := fmt.Sprintf("%s%s", o.config.CollectionPrefix, collection)
	return db.Collection(prefixed)
}

func (o *Store) Ping(ctx context.Context) error {
	if o.client == nil {
		return errors.New("store has not been initialized")
	}

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
	if o.client == nil {
		return nil
	}

	return o.client.Disconnect(ctx)
}
