package store

import (
	"context"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"

	"github.com/tidepool-org/platform/errors"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

const (
	CollectionName = "shopify_order_events"

	OrderEventTypeCreated   = "created"
	OrderEventTypeDelivered = "delivered"
)

var (
	orderIDRegExp = regexp.MustCompile(`^gid://shopify/Order/\d+$`)
)

type ShopifyOrderEvent struct {
	OrderGID    string    `bson:"orderGID"`
	UserID      string    `bson:"userId"`
	Type        string    `bson:"type"`
	CreatedTime time.Time `bson:"createdTime"`
}

func (s *ShopifyOrderEvent) Validate(validator structure.Validator) {
	validator.String("orderGID", &s.OrderGID).NotEmpty().Matches(orderIDRegExp)
	validator.String("userId", &s.UserID).NotEmpty().Using(user.IDValidator)
	validator.String("type", &s.Type).OneOf(OrderEventTypes()...)
	validator.Time("createdTime", &s.CreatedTime).NotZero()
}

func OrderEventTypes() []string {
	return []string{
		OrderEventTypeCreated,
		OrderEventTypeDelivered,
	}
}

type Store interface {
	GetShopifyOrderEvent(ctx context.Context, orderGID, typ string) (*ShopifyOrderEvent, error)
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
			Keys: bson.D{{Key: "orderGID", Value: 1}, {Key: "type", Value: 1}},
			Options: options.Index().
				SetUnique(true),
		},
	})
}

func (s *store) GetShopifyOrderEvent(ctx context.Context, orderGID, typ string) (*ShopifyOrderEvent, error) {
	var event ShopifyOrderEvent
	err := s.FindOne(ctx, bson.M{"orderGID": orderGID, "type": typ}).Decode(&event)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to get shopify order event")
	}

	return &event, nil
}

func (s *store) CreateShopifyOrderEvent(ctx context.Context, event ShopifyOrderEvent) error {
	validator := structureValidator.New(log.LoggerFromContext(ctx))
	if err := validator.Validate(&event); err != nil {
		return errors.Wrap(err, "event is not valid")
	}

	_, err := s.InsertOne(ctx, event)
	if err != nil {
		return errors.Wrap(err, "unable to create shopify order event")
	}

	return nil
}
