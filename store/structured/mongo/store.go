package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/errors"
)

var StoreModule = fx.Options(
	fx.Provide(LoadConfig),
	fx.Provide(NewStore),
	fx.Invoke(AppendLifecycleHooksToStore),
)

type Store struct {
	client *mongo.Client
	config *Config
}

type Status struct {
	Ping string
}

func NewStore(c *Config) (*Store, error) {
	if c == nil {
		return nil, errors.New("database config is empty")
	}
	return NewStoreFromClient(c, nil)
}

func NewStoreFromClient(c *Config, clientOptions *options.ClientOptions) (*Store, error) {
	var err error
	store := &Store{
		config: c,
	}

	if clientOptions == nil {
		clientOptions = options.Client()
	}
	clientOptions.ApplyURI(c.AsConnectionString()).
		SetConnectTimeout(store.config.Timeout).
		SetServerSelectionTimeout(store.config.Timeout)
	store.client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "connection options are invalid")
	}

	return store, nil
}

func AppendLifecycleHooksToStore(store *Store, lifecycle fx.Lifecycle) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return store.Ping(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return store.Terminate(ctx)
		},
	})
}

func (o *Store) GetRepository(collection string) *Repository {
	return NewRepository(o.GetCollection(collection))
}

func (o *Store) GetCollection(collection string) *mongo.Collection {
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
