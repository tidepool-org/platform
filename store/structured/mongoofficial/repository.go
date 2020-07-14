package mongoofficial

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/errors"
)

type Repository struct {
	collection *mongo.Collection
}

func NewRepository(collection *mongo.Collection) *Repository {
	return &Repository{
		collection: collection,
	}
}

func (r *Repository) C() *mongo.Collection {
	return r.collection
}

func (r *Repository) CreateAllIndexes(ctx context.Context, indexes []mongo.IndexModel) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if _, err := r.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		return errors.Wrap(err, "unable to create indexes")
	}

	return nil
}

func (r *Repository) FindOneByID(ctx context.Context, id primitive.ObjectID, model interface{}) error {
	return r.C().FindOne(ctx, bson.M{"_id": id}).Decode(model)
}
