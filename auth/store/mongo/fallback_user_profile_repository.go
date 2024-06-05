package mongo

import (
	"context"
	"encoding/json"
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
	var profile user.LegacyUserProfile
	if err := p.FindOne(ctx, selector).Decode(&profile); err != nil {
		if stdErrors.Is(err, mongo.ErrNoDocuments) {
			return nil, user.ErrUserProfileNotFound
		}
		return nil, err
	}
	return &profile, nil

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
	var doc user.LegacyRawSeagullProfile
	// The original seagull code just had a JSONified string as the value
	// so we have to Unmarshal that to a an actual object, add any
	// updates to the profile, and then Marshal it back to JSON to store
	// in the database.
	opts := options.FindOne().SetProjection(bson.M{"_id": 0, "value": 1})
	selector := bson.M{"userId": userID}
	err := p.FindOne(ctx, selector, opts).Decode(&doc)

	// A user can have no profile set - see seagull/lib/routes/seagullApi.js `if (err.statusCode == 404 && addIfNotThere)`
	var noDocument bool
	if stdErrors.Is(err, mongo.ErrNoDocuments) {
		noDocument = true
	}
	if err != nil && !noDocument {
		return err
	}
	// Since the legacy seagull is actually a stringified JSON object
	// we need to JSON parse it so we can then add the profile as a regular object, then restrigify it
	value := make(map[string]any)
	if !noDocument {
		if err := json.Unmarshal([]byte(doc.Value), &value); err != nil {
			return err
		}
	}
	value["profile"] = profile
	valueRaw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	uselector := bson.M{"userId": userID}
	update := bson.M{
		"$set": bson.M{
			"value": string(valueRaw),
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
