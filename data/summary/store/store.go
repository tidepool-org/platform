package store

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Repo[T types.Stats] struct {
	*storeStructuredMongo.Repository
}

func New[T types.Stats](delegate *storeStructuredMongo.Repository) *Repo[T] {
	return &Repo[T]{
		delegate,
	}
}

func (r *Repo[T]) GetSummary(ctx context.Context, userId string) (*types.Summary[T], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	summary := types.Create[T](userId)
	selector := bson.M{
		"userId": userId,
		"type":   summary.Type,
	}

	err := r.FindOne(ctx, selector).Decode(&summary)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get summary")
	}

	return &summary, nil
}

func (r *Repo[T]) DeleteSummary(ctx context.Context, userId string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	selector := bson.M{
		"userId": userId,
	}

	_, err := r.DeleteOne(ctx, selector)
	if err != nil {
		return errors.Wrap(err, "unable to delete summary")
	}

	return nil
}

func (r *Repo[T]) UpsertSummary(ctx context.Context, summary *types.Summary[T]) (*types.Summary[T], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if summary == nil {
		return nil, errors.New("summary object is missing")
	}

	// TODO do we need s here?
	s := types.Create[T](summary.UserID)
	if summary.Type != s.Type {
		return nil, fmt.Errorf("invalid summary type %v, expected %v", summary.Type, s.Type)
	}

	if summary.UserID == "" {
		return nil, errors.New("summary missing UserID")
	}

	opts := options.Update().SetUpsert(true)
	selector := bson.M{
		"userId": summary.UserID,
		"type":   summary.Type,
	}

	_, err := r.UpdateOne(ctx, selector, bson.M{"$set": summary}, opts)

	return summary, err
}

func (r *Repo[T]) DistinctSummaryIDs(ctx context.Context) ([]string, error) {
	var userIDs []string

	if ctx == nil {
		return userIDs, errors.New("context is missing")
	}

	selector := bson.M{}

	result, err := r.Distinct(ctx, "userId", selector)
	if err != nil {
		return userIDs, errors.New("error fetching distinct userIDs")
	}

	for _, v := range result {
		userIDs = append(userIDs, v.(string))
	}

	return userIDs, nil
}

func (r *Repo[T]) CreateSummaries(ctx context.Context, summaries []*types.Summary[T]) (int, error) {
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

	writeResult, err := r.InsertMany(ctx, insertData, opts)
	count := len(writeResult.InsertedIDs)

	if err != nil {
		if count > 0 {
			return count, errors.Wrap(err, "failed to create some summaries")
		}
		return count, errors.Wrap(err, "unable to create summaries")
	}
	return count, nil
}

func (r *Repo[T]) SetOutdated(ctx context.Context, userId string) (*time.Time, error) {
	var outdatedTime *time.Time
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	if userId == "" {
		return nil, errors.New("user id is missing")
	}

	// we need to get the summary first, as there is multiple possible operations, and we do not want to replace
	// the existing field, but also want to upsert if no summary exists.
	s := types.Create[T](userId)
	opts := options.Update().SetUpsert(true)

	selector := bson.M{
		"userId": userId,
		"type":   s.Type,
	}

	err := r.FindOne(ctx, selector).Decode(&s)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, errors.Wrap(err, "unable to get summary")
	}

	outdatedTime = s.Dates.OutdatedSince

	if outdatedTime == nil {
		outdatedTime = pointer.FromTime(time.Now().UTC().Truncate(time.Millisecond))
		_, err = r.UpdateOne(ctx, selector, bson.M{"$set": bson.M{"dates.outdatedSince": outdatedTime}}, opts)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to update user %s outdatedSince date for type %s", userId, s.Type)
		}
	}

	return outdatedTime, nil
}

func (r *Repo[T]) GetOutdatedUserIDs(ctx context.Context, page *page.Pagination) ([]string, error) {
	// we use a summary, instead of a type specific summary as we don't actually care about its extra data
	var summaries []*types.Summary[T]

	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	if page == nil {
		return nil, errors.New("pagination is missing")
	}

	selector := bson.M{"dates.outdatedSince": bson.M{"$ne": nil}}

	opts := options.Find()
	opts.SetSort(bson.D{
		{Key: "dates.outdatedSince", Value: 1},
	})
	opts.SetLimit(int64(page.Size))

	cursor, err := r.Find(ctx, selector, opts)

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
