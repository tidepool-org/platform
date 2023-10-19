package store

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type Repo[T types.Stats, A types.StatsPt[T]] struct {
	*storeStructuredMongo.Repository
}

type TypelessRepo struct {
	*storeStructuredMongo.Repository
}

func New[T types.Stats, A types.StatsPt[T]](delegate *storeStructuredMongo.Repository) *Repo[T, A] {
	return &Repo[T, A]{
		delegate,
	}
}

func NewTypeless(delegate *storeStructuredMongo.Repository) *TypelessRepo {
	return &TypelessRepo{
		delegate,
	}
}

func (r *Repo[T, A]) GetSummary(ctx context.Context, userId string) (*types.Summary[T, A], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	summary := types.Create[T, A](userId)
	selector := bson.M{
		"userId": userId,
		"type":   summary.Type,
	}

	err := r.FindOne(ctx, selector).Decode(&summary)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get summary")
	}

	return summary, nil
}

func (r *TypelessRepo) DeleteSummary(ctx context.Context, userId string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userId == "" {
		return errors.New("userId is missing")
	}

	selector := bson.M{
		"userId": userId,
	}

	_, err := r.DeleteMany(ctx, selector)
	if err != nil {
		return errors.Wrap(err, "unable to delete summary")
	}

	return nil
}

func (r *Repo[T, A]) DeleteSummary(ctx context.Context, userId string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userId == "" {
		return errors.New("userId is missing")
	}

	selector := bson.M{
		"userId": userId,
		"type":   types.GetTypeString[T, A](),
	}

	_, err := r.DeleteMany(ctx, selector)
	if err != nil {
		return errors.Wrap(err, "unable to delete summary")
	}

	return nil
}

func (r *Repo[T, A]) UpsertSummary(ctx context.Context, userSummary *types.Summary[T, A]) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userSummary == nil {
		return errors.New("summary object is missing")
	}

	var expectedType = types.GetTypeString[T, A]()
	if userSummary.Type != expectedType {
		return fmt.Errorf("invalid summary type '%v', expected '%v'", userSummary.Type, expectedType)
	}

	if userSummary.UserID == "" {
		return errors.New("summary is missing UserID")
	}

	opts := options.Update().SetUpsert(true)
	selector := bson.M{
		"userId": userSummary.UserID,
		"type":   userSummary.Type,
	}

	_, err := r.UpdateOne(ctx, selector, bson.M{"$set": userSummary}, opts)

	return err
}

func (r *Repo[T, A]) DistinctSummaryIDs(ctx context.Context) ([]string, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	selector := bson.M{"type": types.GetTypeString[T, A]()}

	result, err := r.Distinct(ctx, "userId", selector)
	if err != nil {
		return nil, errors.New("error fetching distinct userIDs")
	}

	var userIDs []string
	for _, v := range result {
		userIDs = append(userIDs, v.(string))
	}

	return userIDs, nil
}

func (r *Repo[T, A]) CreateSummaries(ctx context.Context, summaries []*types.Summary[T, A]) (int, error) {
	if ctx == nil {
		return 0, errors.New("context is missing")
	}
	if len(summaries) == 0 {
		return 0, errors.New("summaries for create missing")
	}

	var expectedType = types.GetTypeString[T, A]()

	insertData := make([]interface{}, 0, len(summaries))

	for i, userSummary := range summaries {
		// we don't guard against duplicates, as they fail to insert safely, we only worry about unfilled fields
		if userSummary.UserID == "" {
			return 0, errors.Errorf("userId is missing at index %d", i)
		} else if userSummary.Type != expectedType {
			return 0, fmt.Errorf("invalid summary type '%v', expected '%v' at index %d", userSummary.Type, expectedType, i)
		}

		insertData = append(insertData, *userSummary)
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

func (r *Repo[T, A]) SetOutdated(ctx context.Context, userId, reason string) (*time.Time, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	// we need to get the summary first, as there is multiple possible operations, and we do not want to replace
	// the existing field, but also want to upsert if no summary exists.
	userSummary, err := r.GetSummary(ctx, userId)
	if err != nil {
		return nil, err
	}

	if userSummary == nil {
		userSummary = types.Create[T, A](userId)
	}

	userSummary.SetOutdated(reason)
	err = r.UpsertSummary(ctx, userSummary)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to update user %s outdatedSince date for type %s", userId, userSummary.Type)
	}

	return userSummary.Dates.OutdatedSince, nil
}

func (r *Repo[T, A]) GetOutdatedUserIDs(ctx context.Context, page *page.Pagination) ([]string, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if page == nil {
		return nil, errors.New("pagination is missing")
	}

	selector := bson.M{
		"dates.outdatedSince": bson.M{"$lte": time.Now().UTC()},
		"type":                types.GetTypeString[T, A](),
	}

	opts := options.Find()
	opts.SetSort(bson.D{
		{Key: "dates.outdatedSince", Value: 1},
	})
	opts.SetLimit(int64(page.Size))
	opts.SetProjection(bson.M{"stats": 0})

	cursor, err := r.Find(ctx, selector, opts)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get outdated summaries")
	}

	var summaries []*types.Summary[T, A]
	if err = cursor.All(ctx, &summaries); err != nil {
		return nil, errors.Wrap(err, "unable to decode outdated summaries")
	}

	var userIDs = make([]string, len(summaries))
	for i := 0; i < len(summaries); i++ {
		userIDs[i] = summaries[i].UserID
	}

	return userIDs, nil
}

func (r *Repo[T, A]) GetMigratableUserIDs(ctx context.Context, page *page.Pagination) ([]string, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if page == nil {
		return nil, errors.New("pagination is missing")
	}

	selector := bson.M{
		"config.schemaVersion": bson.M{"$ne": types.SchemaVersion},
		"type":                 types.GetTypeString[T, A](),
	}

	opts := options.Find()
	opts.SetSort(bson.D{
		{Key: "dates.lastUpdatedDate", Value: 1},
	})
	opts.SetLimit(int64(page.Size))
	opts.SetProjection(bson.M{"stats": 0})

	cursor, err := r.Find(ctx, selector, opts)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get outdated summaries")
	}

	var summaries []*types.Summary[T, A]
	if err = cursor.All(ctx, &summaries); err != nil {
		return nil, errors.Wrap(err, "unable to decode outdated summaries")
	}

	var userIDs = make([]string, len(summaries))
	for i := 0; i < len(summaries); i++ {
		userIDs[i] = summaries[i].UserID
	}

	return userIDs, nil
}
