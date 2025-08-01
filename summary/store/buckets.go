package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/summary/types"
)

type Buckets[PB types.BucketDataPt[B], B types.BucketData] struct {
	*storeStructuredMongo.Repository
	Type string
}

func NewBuckets[PB types.BucketDataPt[B], B types.BucketData](delegate *storeStructuredMongo.Repository, typ string) *Buckets[PB, B] {
	return &Buckets[PB, B]{
		Repository: delegate,
		Type:       typ,
	}
}

func (r *Buckets[PB, B]) GetBucketsByTime(ctx context.Context, userId string, startTime, endTime time.Time) (types.BucketsByTime[PB, B], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	buckets := make([]types.Bucket[PB, B], 1)
	transformed := make(types.BucketsByTime[PB, B], 1)

	selector := bson.M{
		"userId": userId,
		"type":   r.Type,
		"time":   bson.M{"$gte": startTime, "$lte": endTime},
	}
	opts := options.Find()
	opts.SetSort(bson.D{{Key: "time", Value: -1}})

	cur, err := r.Find(ctx, selector, opts)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return transformed, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get buckets: %w", err)
	}

	if err = cur.All(ctx, &buckets); err != nil {
		return nil, fmt.Errorf("unable to decode buckets: %w", err)
	}

	for _, bucket := range buckets {
		transformed[bucket.Time] = &bucket
	}

	return transformed, nil
}

func (r *Buckets[PB, B]) GetAllBuckets(ctx context.Context, userId string) (*mongo.Cursor, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	return r.GetBucketsRange(ctx, userId, nil, nil)
}

func (r *Buckets[PB, B]) GetBucketsRange(ctx context.Context, userId string, startTime *time.Time, endTime *time.Time) (*mongo.Cursor, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	selector := bson.M{
		"userId": userId,
		"type":   r.Type,
	}

	timeSelector := bson.M{}
	if startTime != nil && !startTime.IsZero() {
		timeSelector["$gte"] = startTime
	}
	if endTime != nil && !endTime.IsZero() {
		timeSelector["$lte"] = endTime
	}
	if len(timeSelector) > 0 {
		selector["time"] = timeSelector
	}

	opts := options.Find()
	// many functions depend on working in reverse, if this sort is changed, many changes will be needed.
	opts.SetSort(bson.D{{Key: "time", Value: -1}})
	opts.SetBatchSize(200)

	cur, err := r.Find(ctx, selector, opts)
	if err == nil || errors.Is(err, mongo.ErrNoDocuments) {
		return cur, nil
	}

	return nil, fmt.Errorf("unable to get buckets: %w", err)
}

func (r *Buckets[PB, B]) TrimExcessBuckets(ctx context.Context, userId string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	bucket, err := r.GetNewest(ctx, userId)
	if err != nil {
		return err
	}

	// we have no buckets
	if bucket == nil {
		return nil
	}

	oldestTimeToKeep := bucket.Time.Add(-time.Hour * types.HoursAgoToKeep)

	selector := bson.M{
		"userId": userId,
		"type":   r.Type,
		"time":   bson.M{"$lt": oldestTimeToKeep},
	}

	_, err = r.DeleteMany(ctx, selector)
	return err
}

func (r *Buckets[PB, B]) ClearInvalidatedBuckets(ctx context.Context, userId string, earliestModified time.Time) (newFirstData time.Time, err error) {
	selector := bson.M{
		"userId": userId,
		"type":   r.Type,
		// round earliestModified to the hour, to ensure it correctly deletes a bucket if only the final bucket is modified
		"time": bson.M{"$gte": earliestModified.UTC().Truncate(time.Hour)},
	}

	result, err := r.DeleteMany(ctx, selector)
	if err != nil {
		return time.Time{}, err
	}

	// If the query did not delete anything, we should return a 0 just in case this was called unconditionally.
	if result.DeletedCount == 0 {
		return time.Time{}, nil
	}

	return r.GetNewestRecordTime(ctx, userId)
}

func (r *Buckets[PB, B]) Reset(ctx context.Context, userId string) error {
	selector := bson.M{
		"userId": userId,
		"type":   r.Type,
	}

	_, err := r.DeleteMany(ctx, selector)
	return err
}

func (r *Buckets[PB, B]) getOne(ctx context.Context, userId string, sort int) (*types.Bucket[PB, B], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	buckets := make([]types.Bucket[PB, B], 1)
	selector := bson.M{
		"userId": userId,
		"type":   r.Type,
	}
	opts := options.Find()
	opts.SetSort(bson.D{{Key: "time", Value: sort}})
	opts.SetLimit(1)

	cur, err := r.Find(ctx, selector, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to get buckets: %w", err)
	}

	if err = cur.All(ctx, &buckets); err != nil {
		return nil, fmt.Errorf("unable to decode buckets: %w", err)
	}

	if len(buckets) == 0 {
		return nil, nil
	}
	return &buckets[0], nil
}

func (r *Buckets[PB, B]) GetNewest(ctx context.Context, userId string) (*types.Bucket[PB, B], error) {
	return r.getOne(ctx, userId, -1)
}

func (r *Buckets[PB, B]) GetOldest(ctx context.Context, userId string) (*types.Bucket[PB, B], error) {
	return r.getOne(ctx, userId, 1)
}

func (r *Buckets[PB, B]) GetNewestRecordTime(ctx context.Context, userId string) (time.Time, error) {
	if ctx == nil {
		return time.Time{}, errors.New("context is missing")
	}
	if userId == "" {
		return time.Time{}, errors.New("userId is missing")
	}

	bucket, err := r.GetNewest(ctx, userId)
	if err != nil || bucket == nil {
		return time.Time{}, err
	}

	return bucket.LastData, nil
}

func (r *Buckets[PB, B]) GetOldestRecordTime(ctx context.Context, userId string) (time.Time, error) {
	if ctx == nil {
		return time.Time{}, errors.New("context is missing")
	}
	if userId == "" {
		return time.Time{}, errors.New("userId is missing")
	}

	bucket, err := r.GetOldest(ctx, userId)
	if err != nil || bucket == nil {
		return time.Time{}, err
	}

	return bucket.FirstData, nil
}

func (r *Buckets[PB, B]) GetTotalHours(ctx context.Context, userId string) (int, error) {
	if ctx == nil {
		return 0, errors.New("context is missing")
	}
	if userId == "" {
		return 0, errors.New("userId is missing")
	}

	firstBucket, err := r.GetOldest(ctx, userId)
	if err != nil {
		return 0, err
	}

	// we have no buckets, no point in continuing
	if firstBucket == nil {
		return 0, nil
	}

	lastBucket, err := r.GetNewest(ctx, userId)
	if err != nil {
		return 0, err
	}

	return int(lastBucket.LastData.Sub(firstBucket.FirstData).Hours()), nil
}

func (r *Buckets[PB, B]) WriteModifiedBuckets(ctx context.Context, buckets types.BucketsByTime[PB, B]) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if len(buckets) == 0 {
		return nil
	}

	modifiedBuckets := make([]mongo.WriteModel, 0, len(buckets))
	for _, v := range buckets {
		if !v.IsModified() {
			continue
		}
		if v.UserId == "" {
			return errors.New("userId is missing")
		}
		if v.Type == "" {
			return errors.New("type is missing")
		}
		if v.Time.IsZero() {
			return errors.New("time is missing")
		}
		modifiedBuckets = append(modifiedBuckets, mongo.NewReplaceOneModel().SetFilter(bson.M{"userId": v.UserId, "type": v.Type, "time": v.Time}).SetReplacement(v).SetUpsert(true))
	}

	return r.writeBuckets(ctx, modifiedBuckets)
}

func (r *Buckets[PB, B]) writeBuckets(ctx context.Context, buckets []mongo.WriteModel) error {
	if len(buckets) == 0 {
		return nil
	}
	opts := options.BulkWrite()
	opts.SetOrdered(false)
	_, err := r.BulkWrite(ctx, buckets, opts)
	return err
}
