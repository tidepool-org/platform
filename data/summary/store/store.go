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

const RealtimeUserThreshold = 16

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

// GetNumberOfDaysWithRealtimeData processes two slices of Item and returns an int count of days with realtime records.
// this currently doesn't handle N slices, only 1-2, might need adjustment for more types.
func GetNumberOfDaysWithRealtimeData(a, b []*types.Bucket[*types.CGMBucketData, types.CGMBucketData]) int {
	var count int

	// Calculate the offset in hours between the first items of each list
	offset := 0
	startA, startB := 0, 0
	combinedLength := len(a) + startA

	if b != nil {
		offset = int(b[0].Date.Sub(a[0].Date).Hours())

		if offset >= 0 {
			startB = offset
		} else {
			startA = -offset
		}

		if temp := len(b) + startB; temp > combinedLength {
			combinedLength = temp
		}
	}

	for i := 0; i < combinedLength; i++ {
		indexA := i - startA
		indexB := i - startB

		// If the A list has a flagged item at this time, count it and advance to the next day.
		if indexA >= 0 && indexA < len(a) && a[indexA].Data.RealtimeRecords > 0 {
			count += 1
			i += 23 - a[indexA].Date.Hour()
			continue
		}

		if b != nil {
			// Likewise with the B list
			if indexB >= 0 && indexB < len(b) && b[indexB].Data.RealtimeRecords > 0 {
				count += 1
				i += 23 - b[indexB].Date.Hour()
				continue
			}
		}

		if b != nil {
			// If neither list has an item at this index, we've exhausted one list, and they don't overlap.
			// We need to jump to the start of the later list.
			if (indexA < 0 || indexA >= len(a)) && (indexB < 0 || indexB >= len(b)) {
				if indexA > 0 {
					i -= indexB + 1
				} else {
					i -= indexA + 1
				}
			}
		}
	}

	return count
}

func (r *TypelessRepo) GetPatientsWithRealtimeData(ctx context.Context, userIds []string, startTime time.Time, endTime time.Time) (map[string]int, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userIds == nil {
		return nil, errors.New("userIds is missing")
	}
	if len(userIds) == 0 {
		return nil, errors.New("no userIds provided")
	}
	if startTime.IsZero() {
		return nil, errors.New("startTime is missing")
	}
	if endTime.IsZero() {
		return nil, errors.New("startTime is missing")
	}

	if startTime.After(endTime) {
		return nil, errors.New("startTime is after endTime")
	}

	if startTime.Before(time.Now().AddDate(0, 0, -60)) {
		return nil, errors.New("startTime is too old ( >60d ago ) ")
	}

	if int(endTime.Sub(startTime).Hours()/24) < RealtimeUserThreshold {
		return nil, errors.New("time range smaller than threshold, impossible")
	}

	typs := []string{types.SummaryTypeBGM, types.SummaryTypeCGM}
	oldestPossibleLastData := startTime.AddDate(0, 0, RealtimeUserThreshold/len(typs))
	newestPossibleFirstData := endTime.AddDate(0, 0, RealtimeUserThreshold/len(typs))
	opts := options.Find()
	opts.SetProjection(bson.M{"stats.buckets": 1})

	var realtimeUsers map[string]int

	for _, userId := range userIds {
		selector := bson.M{
			"userId":          userId,
			"type":            bson.M{"$in": typs},
			"dates.lastData":  bson.M{"$gte": oldestPossibleLastData},
			"dates.firstData": bson.M{"$lte": newestPossibleFirstData},
			// maybe filter period too? we don't care if offset and regular 30d aren't over 16d of realtime records
		}
		cursor, err := r.Find(ctx, selector)
		if err != nil {
			return nil, fmt.Errorf("unable to get realtime summaries for %s:  %w", userId, err)
		}

		var userSummaries []types.Summary[types.CGMStats, *types.CGMStats]
		if err = cursor.All(ctx, &userSummaries); err != nil {
			return nil, fmt.Errorf("unable to decode summaries for user %s: %w", userId, err)
		}

		var buckets [][]*types.Bucket[*types.CGMBucketData, types.CGMBucketData]
		for i := 0; i < len(userSummaries); i++ {
			if len(userSummaries[i].Stats.Buckets) > 0 {
				startOffset := int(startTime.Sub(userSummaries[i].Stats.Buckets[0].Date).Hours())
				endOffset := int(endTime.Sub(userSummaries[i].Stats.Buckets[0].Date).Hours())
				buckets = append(buckets, userSummaries[i].Stats.Buckets[startOffset:endOffset])
			}
		}

		realtimeDays := 0
		if len(buckets) > 1 {
			realtimeDays = GetNumberOfDaysWithRealtimeData(buckets[0], buckets[1])
		} else if len(buckets) > 0 {
			realtimeDays = GetNumberOfDaysWithRealtimeData(buckets[0], nil)
		}

		realtimeUsers[userId] = realtimeDays

	}

	return realtimeUsers, nil
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
