package store

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/summary/types"
)

type Summaries[PP types.PeriodsPt[P, PB, B], PB types.BucketDataPt[B], P types.Periods, B types.BucketData] struct {
	*storeStructuredMongo.Repository
}

type TypelessSummaries struct {
	*storeStructuredMongo.Repository
}

func NewSummaries[PP types.PeriodsPt[P, PB, B], PB types.BucketDataPt[B], P types.Periods, B types.BucketData](delegate *storeStructuredMongo.Repository) *Summaries[PP, PB, P, B] {
	return &Summaries[PP, PB, P, B]{
		delegate,
	}
}

func NewTypeless(delegate *storeStructuredMongo.Repository) *TypelessSummaries {
	return &TypelessSummaries{
		delegate,
	}
}

func (r *Summaries[PP, PB, P, B]) GetSummary(ctx context.Context, userId string) (*types.Summary[PP, PB, P, B], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	summary := types.Create[PP, PB](userId)
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

func (r *Summaries[PP, PB, P, B]) DeleteSummary(ctx context.Context, userId string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userId == "" {
		return errors.New("userId is missing")
	}

	selector := bson.M{
		"userId": userId,
		"type":   types.GetType[PP, PB, P, B](),
	}

	_, err := r.DeleteMany(ctx, selector)
	if err != nil {
		return fmt.Errorf("unable to delete summary: %w", err)
	}

	return nil
}

func (r *Summaries[PP, PB, P, B]) ReplaceSummary(ctx context.Context, userSummary *types.Summary[PP, PB, P, B]) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userSummary == nil {
		return errors.New("summary object is missing")
	}

	var expectedType = types.GetType[PP, PB]()
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

func (r *Summaries[PP, PB, P, B]) CreateSummaries(ctx context.Context, summaries []*types.Summary[PP, PB, P, B]) (int, error) {
	if ctx == nil {
		return 0, errors.New("context is missing")
	}
	if len(summaries) == 0 {
		return 0, errors.New("summaries for create missing")
	}

	var expectedType = types.GetType[PP, PB]()

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

	count := 0
	if writeResult != nil {
		count = len(writeResult.InsertedIDs)
	}

	if err != nil {
		if count > 0 {
			return count, fmt.Errorf("failed to create some summaries: %w", err)
		}
		return count, fmt.Errorf("unable to create summaries: %w", err)
	}
	return count, nil
}
