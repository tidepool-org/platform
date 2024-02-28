package store

import (
	"context"
	"fmt"
	"time"

	"errors"

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

	summary := types.Create[A](userId)
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
		return fmt.Errorf("unable to delete summary: %w", err)
	}

	return nil
}

func (r *TypelessRepo) GetRealtimePatients(ctx context.Context, userIds []string, startTime time.Time, endTime time.Time) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userIds == nil {
		return errors.New("userIds is missing")
	}
	if len(userIds) == 0 {
		return errors.New("no userIds provided")
	}
	if startTime.IsZero() {
		return errors.New("startTime is missing")
	}
	if endTime.IsZero() {
		return errors.New("startTime is missing")
	}

	if startTime.After(endTime) {
		return errors.New("startTime is after endTime")
	}

	if startTime.Before(time.Now().AddDate(0, 0, -60)) {
		return errors.New("startTime is too old ( >60d ago ) ")
	}

	// todo const?
	threshold := 16

	if int(endTime.Sub(startTime).Hours()/24) < threshold {
		return errors.New("time range smaller than threshold, impossible")
	}

	typs := []string{types.SummaryTypeBGM, types.SummaryTypeCGM}
	oldestPossibleLastData := startTime.AddDate(0, 0, threshold)

	for _, userId := range userIds {
		selector := bson.M{
			"userId":         userId,
			"type":           bson.M{"$in": typs},
			"dates.lastData": bson.M{"$gt": oldestPossibleLastData},
		}
		r.Find(ctx, selector)
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
		return fmt.Errorf("unable to delete summary: %w", err)
	}

	return nil
}

func (r *Repo[T, A]) ReplaceSummary(ctx context.Context, userSummary *types.Summary[T, A]) error {
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

	opts := options.Replace().SetUpsert(true)
	selector := bson.M{
		"userId": userSummary.UserID,
		"type":   userSummary.Type,
	}

	_, err := r.ReplaceOne(ctx, selector, userSummary, opts)

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
		userSummary = types.Create[A](userId)
	}

	userSummary.SetOutdated(reason)
	err = r.ReplaceSummary(ctx, userSummary)
	if err != nil {
		return nil, fmt.Errorf("unable to update user %s outdatedSince date for type %s: %w", userId, userSummary.Type, err)
	}

	return userSummary.Dates.OutdatedSince, nil
}

func (r *Repo[T, A]) GetOutdatedUserIDs(ctx context.Context, page *page.Pagination) (*types.OutdatedSummariesResponse, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if page == nil {
		return nil, errors.New("pagination is missing")
	}

	selector := bson.M{
		"type":                types.GetTypeString[T, A](),
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

	userSummary := &types.Summary[T, A]{}
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

func (r *Repo[T, A]) GetMigratableUserIDs(ctx context.Context, page *page.Pagination) ([]string, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if page == nil {
		return nil, errors.New("pagination is missing")
	}

	selector := bson.M{
		"type":                 types.GetTypeString[T, A](),
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

	var summaries []*types.Summary[T, A]
	if err = cursor.All(ctx, &summaries); err != nil {
		return nil, fmt.Errorf("unable to decode outdated summaries: %w", err)
	}

	var userIDs = make([]string, len(summaries))
	for i := 0; i < len(summaries); i++ {
		userIDs[i] = summaries[i].UserID
	}

	return userIDs, nil
}
