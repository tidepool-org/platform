package mongo

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/data/summary"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/page"

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
				{Key: "lastUpdatedDate", Value: 1},
			},
			Options: options.Index().
				SetBackground(true).
				SetName("LastUpdatedDate"),
		},
		{
			Keys: bson.D{
				{Key: "outdatedSince", Value: 1},
			},
			Options: options.Index().
				SetBackground(true).
				SetName("OutdatedSince"),
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

	var userSummary = summary.New(id)
	selector := bson.M{
		"userId": id,
	}

	err := d.FindOne(ctx, selector).Decode(userSummary)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get summary")
	}

	return userSummary, nil
}

func (d *SummaryRepository) DeleteSummary(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if id == "" {
		return errors.New("summary UserID is missing")
	}

	selector := bson.M{
		"userId": id,
	}

	_, err := d.DeleteOne(ctx, selector)

	if err != nil {
		return errors.Wrap(err, "unable to delete summary")
	}

	return nil
}

func (d *SummaryRepository) GetOutdatedUserIDs(ctx context.Context, page *page.Pagination) ([]string, error) {
	var summaries []*summary.Summary

	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	if page == nil {
		return nil, errors.New("pagination is missing")
	}

	selector := bson.M{
		"outdatedSince": bson.M{"$ne": nil},
	}

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "outdatedSince", Value: 1}})
	opts.SetLimit(int64(page.Size))

	cursor, err := d.Find(ctx, selector, opts)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get outdated summaries")
	}

	if err = cursor.All(ctx, &summaries); err != nil {
		return nil, errors.Wrap(err, "unable to decode outdated summaries")
	}

	var userIDs = make([]string, len(summaries))
	for i := 0; i < len(summaries); i++ {
		userIDs[i] = summaries[i].UserID
	}

	return userIDs, nil
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

	opts := options.Update().SetUpsert(true)
	selector := bson.M{"userId": summary.UserID}

	_, err := d.UpdateOne(ctx, selector, bson.M{"$set": summary}, opts)

	return summary, err
}

func (d *SummaryRepository) SetOutdated(ctx context.Context, id string) (*time.Time, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	if id == "" {
		return nil, errors.New("user id is missing")
	}

	// we need to get the summary first, as there is multiple possible operations, and we do not want to replace
	// the existing field, but also want to upsert if no summary exists.
	var summary summary.Summary
	opts := options.Update().SetUpsert(true)
	timestamp := time.Now().UTC().Truncate(time.Millisecond)

	update := bson.M{
		"$set": bson.M{
			"outdatedSince": &timestamp,
		},
	}

	selector := bson.M{
		"userId": id,
	}

	err := d.FindOne(ctx, selector).Decode(&summary)

	if err != nil && err != mongo.ErrNoDocuments {
		return nil, errors.Wrap(err, "unable to get summary")
	}

	if summary.OutdatedSince != nil {
		return summary.OutdatedSince, nil
	}

	_, err = d.UpdateOne(ctx, selector, update, opts)
	if err != nil {
		return nil, errors.Wrap(err, "unable to update outdatedSince date")
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

func (d *SummaryRepository) CreateSummaries(ctx context.Context, summaries []*summary.Summary) (int, error) {
	if ctx == nil {
		return 0, errors.New("context is missing")
	}
	if len(summaries) == 0 {
		return 0, errors.New("summaries for create missing")
	}

	insertData := make([]interface{}, len(summaries))

	for i := 0; i < len(summaries); i++ {
		insertData[i] = *summaries[i]
	}

	opts := options.InsertMany().SetOrdered(false)

	writeResult, err := d.InsertMany(ctx, insertData, opts)
	count := len(writeResult.InsertedIDs)

	if err != nil {
		if count > 0 {
			return count, errors.Wrap(err, "failed to create some summaries")
		}
		return count, errors.Wrap(err, "unable to create summaries")
	}
	return count, nil
}
