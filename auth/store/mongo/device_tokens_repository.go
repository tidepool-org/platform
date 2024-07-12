package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/errors"
	structuredmongo "github.com/tidepool-org/platform/store/structured/mongo"
)

// deviceTokenRepo implements devicetokens.Repository, writing data to a
// MongoDB collection.
type deviceTokenRepo structuredmongo.Repository

func (r *deviceTokenRepo) GetAllByUserID(ctx context.Context, userID string) ([]*devicetokens.Document, error) {
	f := bson.M{"userId": userID}
	cursor, err := r.Find(ctx, f, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []*devicetokens.Document
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// Upsert will create or update the given Config.
func (r *deviceTokenRepo) Upsert(ctx context.Context, doc *devicetokens.Document) error {
	// The presence of UserID and TokenID should be enforced with a mongodb
	// index, but better safe than sorry.
	if doc.UserID == "" {
		return errors.New("UserID is empty")
	}
	if doc.TokenKey == "" {
		return errors.New("TokenID is empty")
	}

	opts := options.Update().SetUpsert(true)
	f := bson.M{"tokenKey": doc.TokenKey, "userId": doc.UserID}
	_, err := r.UpdateOne(ctx, f, bson.M{"$set": doc}, opts)
	if err != nil {
		return errors.Wrap(err, "unable to upsert device token")
	}
	return nil
}

// EnsureIndexes to maintain index constraints.
func (r *deviceTokenRepo) EnsureIndexes() error {
	repo := structuredmongo.Repository(*r)
	return (&repo).CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "tokenKey", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("UserIdTokenKeyUnique"),
		},
	})
}
