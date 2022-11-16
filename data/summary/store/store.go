package store

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tidepool-org/platform/data/summary/types"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repo[T types.Stats] struct {
	*storeStructuredMongo.Repository
}

func New[T types.Stats](delegate *storeStructuredMongo.Repository) *Repo[T] {
	return &Repo[T]{
		delegate,
	}
}

func (r *Repo[T]) GetSummary(ctx context.Context, userId string) (*types.Summary[T], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	summary := types.Create[T]()
	selector := bson.M{
		"userId": userId,
		"type":   summary.Type,
	}

	err := r.FindOne(ctx, selector).Decode(&summary)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get summary")
	}

	return &summary, nil
}

func (r *Repo[T]) DeleteSummary(ctx context.Context, userId string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	selector := bson.M{
		"userId": userId,
	}

	_, err := r.DeleteOne(ctx, selector)
	if err != nil {
		return errors.Wrap(err, "unable to delete summary")
	}

	return nil
}

func (r *Repo[T]) UpsertSummary(ctx context.Context, summary *types.Summary[T]) (*types.Summary[T], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if summary == nil {
		return nil, errors.New("summary object is missing")
	}

	s := types.Create[T]()
	if summary.Type != s.Type {
		return nil, fmt.Errorf("invalid summary type %v, expected %v", summary.Type, s.Type)
	}

	if summary.UserID == "" {
		return nil, errors.New("summary missing UserID")
	}

	opts := options.Update().SetUpsert(true)
	selector := bson.M{
		"userId": summary.UserID,
		"type":   summary.Type,
	}

	_, err := r.UpdateOne(ctx, selector, bson.M{"$set": summary}, opts)

	return summary, err
}
