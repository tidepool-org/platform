package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type Summaries[A types.StatsPt[T, P, B], P types.BucketDataPt[B], T types.Stats, B types.BucketData] struct {
	*storeStructuredMongo.Repository
}

type TypelessSummaries struct {
	*storeStructuredMongo.Repository
}

func NewSummaries[A types.StatsPt[T, P, B], P types.BucketDataPt[B], T types.Stats, B types.BucketData](delegate *storeStructuredMongo.Repository) *Summaries[A, P, T, B] {
	return &Summaries[A, P, T, B]{
		delegate,
	}
}

func NewTypeless(delegate *storeStructuredMongo.Repository) *TypelessSummaries {
	return &TypelessSummaries{
		delegate,
	}
}

func (r *Summaries[A, P, T, B]) GetSummary(ctx context.Context, userId string) (*types.Summary[A, P, T, B], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	summary := types.Create[A, P](userId)
	selector := bson.M{
		"userId": userId,
		"type":   summary.Type,
	}

	err := r.FindOne(ctx, selector).Decode(&summary)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get summary: %w", err)
	}

	return summary, nil
}

func (r *TypelessSummaries) DeleteSummary(ctx context.Context, userId string) error {
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
		return fmt.Errorf("unable to delete summary: %w", err)
	}

	return nil
}

func (r *Summaries[A, P, T, B]) DeleteSummary(ctx context.Context, userId string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userId == "" {
		return errors.New("userId is missing")
	}

	selector := bson.M{
		"userId": userId,
		"type":   types.GetTypeString[A, P, T, B](),
	}

	_, err := r.DeleteMany(ctx, selector)
	if err != nil {
		return fmt.Errorf("unable to delete summary: %w", err)
	}

	return nil
}

func (r *Summaries[A, P, T, B]) ReplaceSummary(ctx context.Context, userSummary *types.Summary[A, P, T, B]) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userSummary == nil {
		return errors.New("summary object is missing")
	}

	var expectedType = types.GetTypeString[A, P]()
	if userSummary.Type != expectedType {
		return fmt.Errorf("invalid summary type '%v', expected '%v'", userSummary.Type, expectedType)
	}

	if userSummary.UserID == "" {
		return errors.New("summary is missing UserID")
	}

	opts := options.Replace().SetUpsert(true)
	selector := bson.M{
		"userId": userSummary.UserID,
		"type":   userSummary.Type,
	}

	_, err := r.ReplaceOne(ctx, selector, userSummary, opts)

	return err
}

func (r *Summaries[A, P, T, B]) DistinctSummaryIDs(ctx context.Context) ([]string, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	selector := bson.M{"type": types.GetTypeString[A, P]()}

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

func (r *Summaries[A, P, T, B]) CreateSummaries(ctx context.Context, summaries []*types.Summary[A, P, T, B]) (int, error) {
	if ctx == nil {
		return 0, errors.New("context is missing")
	}
	if len(summaries) == 0 {
		return 0, errors.New("summaries for create missing")
	}

	var expectedType = types.GetTypeString[A, P]()

	insertData := make([]interface{}, 0, len(summaries))

	for i, userSummary := range summaries {
		// we don't guard against duplicates, as they fail to insert safely, we only worry about unfilled fields
		if userSummary.UserID == "" {
			return 0, fmt.Errorf("userId is missing at index %d", i)
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
			return count, fmt.Errorf("failed to create some summaries: %w", err)
		}
		return count, fmt.Errorf("unable to create summaries: %w", err)
	}
	return count, nil
}

func (r *Summaries[A, P, T, B]) SetOutdated(ctx context.Context, userId, reason string) (*time.Time, error) {
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
		userSummary = types.Create[A, P](userId)
	}

	userSummary.SetOutdated(reason)
	err = r.ReplaceSummary(ctx, userSummary)
	if err != nil {
		return nil, fmt.Errorf("unable to update user %s outdatedSince date for type %s: %w", userId, userSummary.Type, err)
	}

	return userSummary.Dates.OutdatedSince, nil
}

func (r *Summaries[A, P, T, B]) GetOutdatedUserIDs(ctx context.Context, page *page.Pagination) (*types.OutdatedSummariesResponse, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if page == nil {
		return nil, errors.New("pagination is missing")
	}

	selector := bson.M{
		"type":                types.GetTypeString[A, P](),
		"dates.outdatedSince": bson.M{"$lte": time.Now().UTC()},
	}

	opts := options.Find()
	opts.SetSort(bson.D{
		{Key: "dates.outdatedSince", Value: 1},
	})
	opts.SetLimit(int64(page.Size))
	opts.SetProjection(bson.M{"stats": 0})

	cursor, err := r.Find(ctx, selector, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to get outdated summaries: %w", err)
	}

	response := &types.OutdatedSummariesResponse{
		UserIds: make([]string, 0, cursor.RemainingBatchLength()),
	}

	userSummary := &types.Summary[A, P, T, B]{}
	for cursor.Next(ctx) {
		if err = cursor.Decode(userSummary); err != nil {
			return nil, fmt.Errorf("unable to decode Summary: %w", err)
		}

		response.UserIds = append(response.UserIds, userSummary.UserID)

		if response.Start.IsZero() {
			response.Start = *userSummary.Dates.OutdatedSince
		}
	}

	// if we saw at least one summary
	if !response.Start.IsZero() {
		response.End = *userSummary.Dates.OutdatedSince
	}

	return response, nil
}

func (r *Summaries[A, P, T, B]) GetMigratableUserIDs(ctx context.Context, page *page.Pagination) ([]string, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if page == nil {
		return nil, errors.New("pagination is missing")
	}

	selector := bson.M{
		"type":                 types.GetTypeString[A, P](),
		"dates.outdatedSince":  nil,
		"config.schemaVersion": bson.M{"$ne": types.SchemaVersion},
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
		return nil, fmt.Errorf("unable to get outdated summaries: %w", err)
	}

	var summaries []*types.Summary[A, P, T, B]
	if err = cursor.All(ctx, &summaries); err != nil {
		return nil, fmt.Errorf("unable to decode outdated summaries: %w", err)
	}

	var userIDs = make([]string, len(summaries))
	for i := 0; i < len(summaries); i++ {
		userIDs[i] = summaries[i].UserID
	}

	return userIDs, nil
}
