package mongo

import (
	"context"
	stdErrors "errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/errors"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

type FallbackUserProfileRepository struct {
	*storeStructuredMongo.Repository
}

func (p *FallbackUserProfileRepository) EnsureIndexes() error {
	return nil
}

func (p *FallbackUserProfileRepository) FindUserProfile(ctx context.Context, userID string) (*user.LegacyUserProfile, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	selector := bson.M{
		"userId": userID,
	}
	var doc user.LegacySeagullDocument
	if err := p.FindOne(ctx, selector).Decode(&doc); err != nil {
		if stdErrors.Is(err, mongo.ErrNoDocuments) {
			return nil, user.ErrUserProfileNotFound
		}
		return nil, err
	}

	return doc.ToLegacyProfile()
}

func (p *FallbackUserProfileRepository) UpdateUserProfile(ctx context.Context, userID string, profile *user.LegacyUserProfile) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}
	if err := structureValidator.New().Validate(profile); err != nil {
		return err
	}
	var doc user.LegacySeagullDocument
	selector := bson.M{"userId": userID}
	err := p.FindOne(ctx, selector).Decode(&doc)

	// A user can have no profile set - see seagull/lib/routes/seagullApi.js `if (err.statusCode == 404 && addIfNotThere)`
	if err != nil && !stdErrors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	// This will create a new value even if doc.Value is empty
	updatedValueRaw, err := user.AddProfileToSeagullValue(doc.Value, profile)
	if err != nil {
		return err
	}

	uselector := bson.M{"userId": userID}
	update := bson.M{
		"$set": bson.M{
			"value":  updatedValueRaw,
			"userId": userID, // Set because of possible upsert
		},
	}
	uopts := options.Update().SetUpsert(true)
	_, err = p.UpdateOne(ctx, uselector, update, uopts)
	return err
}

func (p *FallbackUserProfileRepository) DeleteUserProfile(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	_, err := p.DeleteOne(ctx, bson.M{"userId": userID})
	if err != nil {
		return errors.Wrap(err, "unable to delete user profile")
	}
	return nil
}
