package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/errors"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type AppValidateRepository struct {
	*storeStructuredMongo.Repository
}

func (r *AppValidateRepository) EnsureIndexes() error {
	return r.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "userId", Value: 1}, {Key: "keyId", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetBackground(true),
		},
	})
}

func (r *AppValidateRepository) Upsert(ctx context.Context, v *appvalidate.AppValidation) error {
	selector := bson.M{
		"userId": v.UserID,
		"keyId":  v.KeyID,
	}
	update := bson.M{
		"$set": v,
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.UpdateOne(ctx, selector, update, opts)
	return err
}

func (r *AppValidateRepository) GetAttestationChallenge(ctx context.Context, f appvalidate.Filter) (string, error) {
	selector := selectorFromFilter(f)
	selector["attestationChallenge"] = bson.M{
		"$nin": []interface{}{"", nil},
	}

	opts := options.FindOne().SetProjection(bson.D{{Key: "attestationChallenge", Value: 1}})
	var av appvalidate.AppValidation
	if err := r.FindOne(ctx, selector, opts).Decode(&av); err != nil {
		return "", err
	}
	return av.AttestationChallenge, nil
}

func (r *AppValidateRepository) Get(ctx context.Context, f appvalidate.Filter) (*appvalidate.AppValidation, error) {
	selector := selectorFromFilter(f)

	var validation appvalidate.AppValidation
	if err := r.FindOne(ctx, selector).Decode(&validation); err != nil {
		return nil, err
	}
	return &validation, nil
}

func (r *AppValidateRepository) UpdateAssertion(ctx context.Context, f appvalidate.Filter, u appvalidate.AssertionUpdate) error {
	selector := selectorFromFilter(f)
	update := bson.M{
		"$set": u,
	}
	res, err := r.UpdateOne(ctx, selector, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("unable to find app validation object")
	}
	return nil
}

func (r *AppValidateRepository) IsVerified(ctx context.Context, f appvalidate.Filter) (bool, error) {
	selector := selectorFromFilter(f)
	opts := options.FindOne().SetProjection(bson.D{{Key: "verified", Value: 1}})
	var av appvalidate.AppValidation
	if err := r.FindOne(ctx, selector, opts).Decode(&av); err != nil {
		return false, err
	}
	return av.Verified, nil
}

func (r *AppValidateRepository) UpdateAttestation(ctx context.Context, f appvalidate.Filter, u appvalidate.AttestationUpdate) error {
	selector := selectorFromFilter(f)
	update := bson.M{
		"$set": u,
	}
	res, err := r.UpdateOne(ctx, selector, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("unable to find app validation object")
	}
	return nil
}

func selectorFromFilter(f appvalidate.Filter) bson.M {
	return bson.M{
		"userId": f.UserID,
		"keyId":  f.KeyID,
	}
}
