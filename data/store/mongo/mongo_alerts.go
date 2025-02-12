package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/errors"
	structuredmongo "github.com/tidepool-org/platform/store/structured/mongo"
)

// alertsRepo implements alerts.Repository, writing data to a MongoDB collection.
type alertsRepo structuredmongo.Repository

// Upsert will create or update the given Config.
//
// Once set, UploadID, UserID, and FollowedUserID cannot be changed. This is to prevent a
// user from granting themselves access to another data set.
func (r *alertsRepo) Upsert(ctx context.Context, conf *alerts.Config) error {
	opts := options.Update().SetUpsert(true)
	filter := bson.D{
		{Key: "userId", Value: conf.UserID},
		{Key: "followedUserId", Value: conf.FollowedUserID},
		{Key: "uploadId", Value: conf.UploadID},
	}
	doc := bson.M{
		"$set":         bson.M{"alerts": conf.Alerts, "activity": conf.Activity},
		"$setOnInsert": filter,
	}
	_, err := r.UpdateOne(ctx, filter, doc, opts)
	if err != nil {
		return fmt.Errorf("upserting alerts.Config: %w", err)
	}
	return nil
}

// Delete will delete the given Config.
func (r *alertsRepo) Delete(ctx context.Context, cfg *alerts.Config) error {
	_, err := r.DeleteMany(ctx, r.filter(cfg), nil)
	if err != nil {
		return fmt.Errorf("upserting alerts.Config: %w", err)
	}
	return nil
}

// List will retrieve any Configs that are defined by followers of the given user.
func (r *alertsRepo) List(ctx context.Context, followedUserID string) ([]*alerts.Config, error) {
	filter := bson.D{
		{Key: "followedUserId", Value: followedUserID},
	}
	cursor, err := r.Find(ctx, filter, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to list alerts.Config(s) for followed user %s", followedUserID)
	}
	defer cursor.Close(ctx)
	out := []*alerts.Config{}
	if err := cursor.All(ctx, &out); err != nil {
		return nil, errors.Wrapf(err, "Unable to decode alerts.Config(s) for followed user %s", followedUserID)
	}
	if err := cursor.Err(); err != nil {
		return nil, errors.Wrapf(err, "Unexpected error for followed user %s", followedUserID)
	}
	return out, nil
}

// Get will retrieve the given Config.
func (r *alertsRepo) Get(ctx context.Context, cfg *alerts.Config) (*alerts.Config, error) {
	res := r.FindOne(ctx, r.filter(cfg), nil)
	if res.Err() != nil {
		return nil, fmt.Errorf("getting alerts.Config: %w", res.Err())
	}
	out := &alerts.Config{}
	if err := res.Decode(out); err != nil {
		return nil, err
	}
	return out, nil
}

// EnsureIndexes to maintain index constraints.
func (r *alertsRepo) EnsureIndexes() error {
	repo := structuredmongo.Repository(*r)
	return (&repo).CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "followedUserId", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("UserIdFollowedUserIdTypeUnique"),
		},
	})
}

func (r *alertsRepo) filter(cfg *alerts.Config) interface{} {
	return bson.D{
		{Key: "userId", Value: cfg.UserID},
		{Key: "followedUserId", Value: cfg.FollowedUserID},
	}
}

type alertsDataRepo structuredmongo.Repository

func (d *alertsDataRepo) GetAlertableData(ctx context.Context,
	params alerts.GetAlertableDataParams) (*alerts.GetAlertableDataResponse, error) {

	if params.End.IsZero() {
		params.End = time.Now()
	}

	cursor, err := d.getAlertableData(ctx, params, dosingdecision.Type)
	if err != nil {
		return nil, err
	}
	dosingDecisions := []*dosingdecision.DosingDecision{}
	if err := cursor.All(ctx, &dosingDecisions); err != nil {
		return nil, errors.Wrap(err, "Unable to load alertable dosing documents")
	}
	cursor, err = d.getAlertableData(ctx, params, continuous.Type)
	if err != nil {
		return nil, err
	}
	glucoseData := []*glucose.Glucose{}
	if err := cursor.All(ctx, &glucoseData); err != nil {
		return nil, errors.Wrap(err, "Unable to load alertable glucose documents")
	}
	response := &alerts.GetAlertableDataResponse{
		DosingDecisions: dosingDecisions,
		Glucose:         glucoseData,
	}

	return response, nil
}

func (d *alertsDataRepo) getAlertableData(ctx context.Context,
	params alerts.GetAlertableDataParams, typ string) (*mongo.Cursor, error) {

	selector := bson.M{
		"_active":  true,
		"uploadId": params.UploadID,
		"type":     typ,
		"_userId":  params.UserID,
		"time":     bson.M{"$gte": params.Start, "$lte": params.End},
	}
	findOptions := options.Find().SetSort(bson.D{{Key: "time", Value: -1}})
	cursor, err := d.Find(ctx, selector, findOptions)
	if err != nil {
		format := "Unable to find alertable %s data in dataset %s"
		return nil, errors.Wrapf(err, format, typ, params.UploadID)
	}
	return cursor, nil
}
