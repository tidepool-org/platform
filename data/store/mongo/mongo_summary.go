package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data/types/blood/glucose/summary"
	"github.com/tidepool-org/platform/errors"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type SummaryRepository struct {
	*storeStructuredMongo.Repository
}

func (d *SummaryRepository) EnsureIndexes() error {
	return d.CreateAllIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
			},
			Options: options.Index().
				SetBackground(true).
				SetUnique(true).
				SetName("UserID"),
		},
		{
			Keys: bson.D{
				{Key: "lastUpdated", Value: 1},
			},
			Options: options.Index().
				SetBackground(true).
				SetName("LastUpdated"),
		},
	})
}

func (d *SummaryRepository) GetSummary(ctx context.Context, id string) (*summary.Summary, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if id == "" {
		return nil, errors.New("summary UserID is missing")
	}

	var summary summary.Summary
	selector := bson.M{
		"userId": id,
	}

	err := d.FindOne(ctx, selector).Decode(&summary)

	if err == mongo.ErrNoDocuments {
		return nil, errors.New("summary not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get summary")
	}

	return &summary, nil
}

func (d *SummaryRepository) UpdateSummary(ctx context.Context, summary *summary.Summary) (*summary.Summary, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if summary == nil {
		return nil, errors.New("summary object is missing")
	}

	if summary.UserID == "" {
		return nil, errors.New("summary missing UserID")
	}

	opts := options.Replace().SetUpsert(true)
	filter := bson.M{"userId": summary.UserID}

	_, err := d.ReplaceOne(ctx, filter, summary, opts)

	return summary, err
}

func (d *SummaryRepository) GetUsersWithSummariesBefore(ctx context.Context, lastUpdated time.Time) ([]string, error) {
	var userIDs []string

	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	// find results on a brand new set
	if lastUpdated.IsZero() {
		lastUpdated = time.Now()
	}

	var summaries []*summary.Summary
	selector := bson.M{
		"lastUpdated": bson.M{"$lte": lastUpdated.Add(-60 * time.Minute)},
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "lastUpdated", Value: 1}})

	cursor, err := d.Find(ctx, selector, findOptions)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get aged summaries")
	}

	if err = cursor.All(ctx, &summaries); err != nil {
		return nil, errors.Wrap(err, "unable to decode aged summaries")
	}

	for _, v := range summaries {
		userIDs = append(userIDs, v.UserID)
	}

	return userIDs, nil
}

func (d *SummaryRepository) GetLastUpdated(ctx context.Context) (*time.Time, error) {
	var lastUpdated time.Time
	var summaries []*summary.Summary

	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	selector := bson.M{}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "lastUpdated", Value: 1}})
	findOptions.SetLimit(1)

	cursor, err := d.Find(ctx, selector, findOptions)

	if err != nil {
		return nil, errors.Wrap(err, "unable to get last cbg date")
	}

	if err = cursor.All(ctx, &summaries); err != nil {
		return nil, errors.Wrap(err, "unable to decode last cbg date")
	}

	if len(summaries) > 0 {
		if summaries[0].LastUpdated != nil {
			lastUpdated = *summaries[0].LastUpdated
		}
	} else {
		lastUpdated = time.Now().UTC().Truncate(time.Millisecond)
	}

	return &lastUpdated, nil
}

func (d *SummaryRepository) UpdateLastUpdated(ctx context.Context, id string) (*time.Time, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	if id == "" {
		return nil, errors.New("user id is missing")
	}

	selector := bson.M{"userId": id}

	timestamp := time.Now().UTC().Truncate(time.Millisecond)
	update := bson.M{
		"$set": bson.M{
			"lastUpdated": &timestamp,
		},
	}

	_, err := d.UpdateOne(ctx, selector, update)

	if err != nil {
		return nil, errors.Wrap(err, "unable to update lastUpdated date")
	}

	return &timestamp, nil
}

func (d *SummaryRepository) DistinctSummaryIDs(ctx context.Context) ([]string, error) {
	var userIDs []string

	if ctx == nil {
		return userIDs, errors.New("context is missing")
	}

	selector := bson.M{}

	result, err := d.Distinct(ctx, "userId", selector)
	if err != nil {
		return userIDs, errors.New("error fetching distinct userIDs")
	}

	for _, v := range result {
		userIDs = append(userIDs, v.(string))
	}

	return userIDs, nil
}

func (d *SummaryRepository) CreateSummaries(ctx context.Context, summaries []*summary.Summary) (int64, error) {
	if ctx == nil {
		return 0, errors.New("context is missing")
	}
	if len(summaries) == 0 {
		return 0, errors.New("summaries for create missing")
	}

	var insertData []mongo.WriteModel

	for _, userSummary := range summaries {
		insertData = append(insertData, mongo.NewInsertOneModel().SetDocument(userSummary))
	}

	opts := options.BulkWrite().SetOrdered(false)

	writeResult, err := d.BulkWrite(ctx, insertData, opts)
	count := writeResult.InsertedCount

	if err != nil {
		if count > 0 {
			return count, errors.Wrap(err, "failed to create some summaries")
		}
		return count, errors.Wrap(err, "unable to create summaries")
	}
	return count, nil
}
