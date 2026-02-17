package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/errors"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

const CollectionName = "shopify_order_events"

type ShopifyOrderEvent struct {
	OrderID    string    `bson:"orderId"`
	UserID     string    `bson:"userId"`
	Type       string    `bson:"type"`
	CreateTime time.Time `bson:"createdTime"`
}

type Store interface {
	GetShopifyOrderEvent(ctx context.Context, orderID, typ string) (*ShopifyOrderEvent, error)
	CreateShopifyOrderEvent(ctx context.Context, event ShopifyOrderEvent) error
}

type store struct {
	*storeStructuredMongo.Repository
}

func NewStore(mongoStore *storeStructuredMongo.Store) (Store, error) {
	if mongoStore == nil {
		return nil, errors.New("mongo store is missing")
	}

	s := &store{
		Repository: mongoStore.GetRepository(CollectionName),
	}

	if err := s.EnsureIndexes(context.Background()); err != nil {
		return nil, errors.Wrap(err, "unable to ensure indexes")
	}

	return s, nil
}

func (s *store) EnsureIndexes(ctx context.Context) error {
	return s.CreateAllIndexes(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "orderId", Value: 1}, {Key: "type", Value: 1}},
			Options: options.Index().
				SetUnique(true),
		},
	})
}

func (s *store) GetShopifyOrderEvent(ctx context.Context, orderID, typ string) (*ShopifyOrderEvent, error) {
	if orderID == "" {
		return nil, errors.New("submission ID is missing")
	}
	if typ == "" {
		return nil, errors.New("submission ID is missing")
	}

	var event ShopifyOrderEvent
	err := s.FindOne(ctx, bson.M{
		"orderId": orderID,
		"type":    typ,
	}).Decode(&event)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to get shopify order event")
	}

	return &event, nil
}

func (s *store) CreateShopifyOrderEvent(ctx context.Context, event ShopifyOrderEvent) error {
	if event.OrderID == "" {
		return errors.New("order id is missing")
	}
	if event.Type == "" {
		return errors.New("type is missing")
	}

	_, err := s.InsertOne(ctx, event)
	if err != nil {
		if storeStructuredMongo.IsDup(err) {
			return errors.New("event already exists")
		}
		return errors.Wrap(err, "unable to create shopify order event")
	}

	return nil
}
