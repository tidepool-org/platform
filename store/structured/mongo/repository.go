package mongo

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
)

type Repository struct {
	*mongo.Collection
	config RepositoryConfig
}

type RepositoryConfig struct {
	DisableIndexCreation bool
}

func NewRepository(collection *mongo.Collection, config RepositoryConfig) *Repository {
	return &Repository{
		collection,
		config,
	}
}

func (r *Repository) CreateAllIndexes(ctx context.Context, indexes []mongo.IndexModel) error {
	if r.config.DisableIndexCreation {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if len(indexes) > 0 {
		if _, err := r.Indexes().CreateMany(ctx, indexes); err != nil {
			return errors.Wrap(err, "unable to create indexes")
		}
	}

	return nil
}

func (r *Repository) FindOneByID(ctx context.Context, id primitive.ObjectID, model interface{}) error {
	return r.FindOne(ctx, bson.M{"_id": id}).Decode(model)
}

func (r *Repository) ConstructUpdate(set bson.M, unset bson.M, operators ...map[string]bson.M) bson.M {
	update := bson.M{}
	if len(set) > 0 {
		update["$set"] = set
	}
	if len(unset) > 0 {
		update["$unset"] = unset
	}
	for _, operator := range operators {
		for fieldKey, fieldValues := range operator {
			update = mergeUpdateField(update, fieldKey, fieldValues)
		}
	}
	if len(update) > 0 {
		return mergeUpdateField(update, "$inc", bson.M{"revision": 1})
	}
	return nil
}

func mergeUpdateField(update bson.M, fieldKey string, fieldValues bson.M) bson.M {
	var mergedFieldValues bson.M
	if raw, ok := update[fieldKey]; ok {
		mergedFieldValues, _ = raw.(bson.M)
	}
	if mergedFieldValues == nil {
		mergedFieldValues = bson.M{}
	}
	for key, value := range fieldValues {
		mergedFieldValues[key] = value
	}
	if len(mergedFieldValues) > 0 {
		update[fieldKey] = mergedFieldValues
	} else {
		delete(update, fieldKey)
	}
	return update
}

type QueryModifier func(query bson.M) bson.M

func ModifyQuery(query bson.M, queryModifiers ...QueryModifier) bson.M {
	if query == nil {
		return nil
	}
	for _, queryModifier := range queryModifiers {
		query = queryModifier(query)
	}
	return query
}

func NotDeleted(query bson.M) bson.M {
	if query == nil {
		return nil
	}
	query["deletedTime"] = bson.M{
		"$exists": false,
	}
	return query
}

// IsDup tells us whether the error passed in occurred because a duplicate existed.
// See https://jira.mongodb.org/browse/GODRIVER-972
func IsDup(err error) bool {
	writeException, ok := err.(mongo.WriteException)

	if !ok {
		return false
	}

	for _, writeError := range writeException.WriteErrors {
		return writeError.Code == 11000 || writeError.Code == 11001 || writeError.Code == 12582 || writeError.Code == 16460 && strings.Contains(writeError.Message, " E11000 ")
	}

	return false
}

func FindWithPagination(pagination *page.Pagination) *options.FindOptions {
	return options.Find().
		SetSkip(int64(pagination.Page * pagination.Size)).
		SetLimit(int64(pagination.Size))
}
