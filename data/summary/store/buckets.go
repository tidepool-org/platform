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
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type Buckets[B types.BucketDataPt[A], A types.BucketData] struct {
	*storeStructuredMongo.Repository
	Type string
}

func NewBuckets[B types.BucketDataPt[A], A types.BucketData](delegate *storeStructuredMongo.Repository, typ string) *Buckets[B, A] {
	return &Buckets[B, A]{
		Repository: delegate,
		Type:       typ,
	}
}

func (r *Buckets[B, A]) GetBuckets(ctx context.Context, userId string, startTime, endTime time.Time) (types.BucketsByTime[B, A], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	buckets := make([]types.Bucket[B, A], 1)
	transformed := make(types.BucketsByTime[B, A], 1)

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

func (r *Buckets[B, A]) GetAllBuckets(ctx context.Context, userId string) (*mongo.Cursor, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	return r.GetBucketsRange(ctx, userId, nil, nil)
}

func (r *Buckets[B, A]) GetBucketsRange(ctx context.Context, userId string, startTime *time.Time, endTime *time.Time) (*mongo.Cursor, error) {
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

	if startTime != nil && !startTime.IsZero() {
		selector["time"].(bson.M)["$gte"] = startTime
	}

	if endTime != nil && !endTime.IsZero() {
		selector["time"].(bson.M)["$lte"] = endTime
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

func (r *Buckets[B, A]) TrimExcessBuckets(ctx context.Context, userId string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	bucket, err := r.GetEnd(ctx, userId, -1)
	if err != nil {
		return err
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

func (r *Buckets[B, A]) ClearInvalidatedBuckets(ctx context.Context, userId string, earliestModified time.Time) (firstData time.Time, err error) {
	selector := bson.M{
		"userId": userId,
		"type":   r.Type,
		"time":   bson.M{"$gt": earliestModified},
	}

	_, err = r.DeleteMany(ctx, selector)
	if err != nil {
		return time.Time{}, err
	}

	return r.GetNewestRecordTime(ctx, userId)
}

func (r *Buckets[B, A]) GetEnd(ctx context.Context, userId string, side int) (*types.Bucket[B, A], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	buckets := make([]types.Bucket[B, A], 1)
	selector := bson.M{
		"userId": userId,
		"type":   r.Type,
	}
	opts := options.Find()
	opts.SetSort(bson.D{{Key: "time", Value: side}})
	opts.SetLimit(1)

	cur, err := r.Find(ctx, selector, opts)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get buckets: %w", err)
	}

	if err = cur.All(ctx, &buckets); err != nil {
		return nil, fmt.Errorf("unable to decode buckets: %w", err)
	}

	return &buckets[0], nil
}

func (r *Buckets[B, A]) GetNewestRecordTime(ctx context.Context, userId string) (time.Time, error) {
	if ctx == nil {
		return time.Time{}, errors.New("context is missing")
	}
	if userId == "" {
		return time.Time{}, errors.New("userId is missing")
	}

	bucket, err := r.GetEnd(ctx, userId, -1)
	if err != nil {
		return time.Time{}, err
	}

	return bucket.LastData, nil
}

func (r *Buckets[B, A]) GetOldestRecordTime(ctx context.Context, userId string) (time.Time, error) {
	if ctx == nil {
		return time.Time{}, errors.New("context is missing")
	}
	if userId == "" {
		return time.Time{}, errors.New("userId is missing")
	}

	bucket, err := r.GetEnd(ctx, userId, 1)
	if err != nil {
		return time.Time{}, err
	}

	return bucket.FirstData, nil
}

func (r *Buckets[B, A]) GetTotalHours(ctx context.Context, userId string) (int, error) {
	if ctx == nil {
		return 0, errors.New("context is missing")
	}
	if userId == "" {
		return 0, errors.New("userId is missing")
	}

	firstBucket, err := r.GetEnd(ctx, userId, 1)
	if err != nil {
		return 0, err
	}

	lastBucket, err := r.GetEnd(ctx, userId, 1)
	if err != nil {
		return 0, err
	}

	return int(lastBucket.LastData.Sub(firstBucket.FirstData).Hours()), nil
}

func (r *Buckets[B, A]) writeBuckets(ctx context.Context, buckets []interface{}) error {
	opts := options.InsertMany()
	opts.SetOrdered(false)
	_, err := r.InsertMany(ctx, buckets, opts)
	return err
}

func (r *Buckets[B, A]) WriteModifiedBuckets(ctx context.Context, startTime time.Time, buckets types.BucketsByTime[B, A]) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if startTime.IsZero() {
		return errors.New("startTime is missing")
	}
	if len(buckets) == 0 {
		return nil
	}

	bucketsInt := make([]interface{}, 0, len(buckets))
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
		bucketsInt = append(bucketsInt, v)
	}

	return r.writeBuckets(ctx, bucketsInt)
}
