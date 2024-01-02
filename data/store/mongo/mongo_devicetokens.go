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

// deviceTokensRepo implements devicetokens.Repository, writing data to a
// MongoDB collection.
type deviceTokensRepo structuredmongo.Repository

// Upsert will create or update the given Config.
func (r *deviceTokensRepo) Upsert(ctx context.Context, doc *devicetokens.Document) error {
	// The presence of UserID and TokenID should be enforced with a mongodb
	// index, but better safe than sorry.
	if doc.UserID == "" {
		return errors.New("UserID may not be empty")
	}
	if doc.TokenKey == "" {
		return errors.New("TokenID may not be empty")
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.UpdateOne(ctx, r.filter(doc), bson.M{"$set": doc}, opts)
	if err != nil {
		return errors.Wrap(err, "upserting device token")
	}
	return nil
}

// EnsureIndexes to maintain index constraints.
func (r *deviceTokensRepo) EnsureIndexes() error {
	repo := structuredmongo.Repository(*r)
	return (&repo).CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "tokenKey", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("UserIdTokenKeyTypeUnique"),
		},
	})
}

func (r *deviceTokensRepo) filter(doc *devicetokens.Document) interface{} {
	return &devicetokens.Document{
		UserID:   doc.UserID,
		TokenKey: doc.TokenKey,
	}
}
