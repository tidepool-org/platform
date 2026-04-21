package history

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Repository struct {
	*storeStructuredMongo.Repository
}

func NewHistoryRepository(store *storeStructuredMongo.Store) *Repository {
	return &Repository{
		Repository: store.GetRepository("notification_history"),
	}
}

func (p *Repository) EnsureIndexes() error {
	return p.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "type", Value: 1},
				{Key: "createdTime", Value: 1},
			},
		},
	})
}

func (p *Repository) Create(ctx context.Context, entry Entry) error {
	entry.CreatedTime = time.Now()
	if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(&entry); err != nil {
		return errors.Wrap(err, "entry is invalid")
	}

	_, err := p.InsertOne(ctx, entry)
	if err != nil {
		return errors.Wrap(err, "unable to persist notification history entry")
	}

	return nil
}

func (p *Repository) List(ctx context.Context, filter Filter, pagination *page.Pagination) ([]Entry, error) {
	if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(&filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}
	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	selector := bson.M{}
	if filter.UserID != "" {
		selector["userId"] = filter.UserID
	}
	if filter.ProcessorType != "" {
		selector["processorType"] = filter.ProcessorType
	}
	if filter.EventType != "" {
		selector["eventType"] = filter.EventType
	}
	if filter.DataSourceID != "" {
		selector["dataSourceId"] = filter.DataSourceID
	}
	if filter.GroupID != "" {
		selector["groupId"] = filter.GroupID
	}
	opts := storeStructuredMongo.FindWithPagination(pagination).
		SetSort(bson.M{"createdTime": -1})
	cursor, err := p.Find(ctx, selector, opts)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
