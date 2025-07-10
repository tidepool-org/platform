package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/summary/types"
)

// EventsRepository abstracts the MongoDB interface to make testing easier.
type EventsRepository interface {
	CountDocuments(context.Context, any, ...*options.CountOptions) (int64, error)
	Find(context.Context, any, ...*options.FindOptions) (*mongo.Cursor, error)
	FindOneAndUpdate(ctx context.Context, filter any, update any, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult
}

type Events struct {
	repo EventsRepository
}

func NewEvents(repo EventsRepository) *Events {
	return &Events{
		repo: repo,
	}
}

func (e *Events) GetOutdatedUserIDs(ctx context.Context, summaryType string,
	limit int) (*types.OutdatedSummariesResponse, error) {

	filter := bson.M{
		"time": bson.M{"$lte": time.Now().UTC()},
		"type": summaryType,
	}
	const mongoSortAsc = 1
	opts := options.Find().
		SetSort(bson.M{"time": mongoSortAsc}). // makes finding start/end easier
		SetLimit(int64(limit))
	cursor, err := e.repo.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get outdated summaries")
	}
	event := &SummaryEvent{}
	total := cursor.RemainingBatchLength()
	userIDs := make([]string, total)
	start, end := time.Time{}, time.Time{}
	for i := 0; cursor.Next(ctx); i++ {
		if err := cursor.Decode(event); err != nil {
			return nil, errors.Wrapf(err, "unable to decode summary event %d", i)
		}
		userIDs[i] = event.UserID
		if i == 0 {
			start = event.Time
		} else if i == total-1 {
			end = event.Time
		}
	}

	lag := time.Since(start)
	defer e.updateMetrics(ctx, summaryType, lag)

	return &types.OutdatedSummariesResponse{
		UserIds: userIDs,
		Start:   start,
		End:     end,
	}, nil
}

func (e *Events) SetOutdated(ctx context.Context, userId, summaryType, reason string) (*time.Time, error) {
	filter := bson.M{
		"userId": userId,
		"type":   summaryType,
	}
	outdated := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"userId": userId,
			"type":   summaryType,
			"time":   outdated,
		},
		"$addToSet": bson.M{
			"reasons": reason,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	if err := e.repo.FindOneAndUpdate(ctx, filter, update, opts).Err(); err != nil {
		return nil, errors.Wrap(err, "unable to set summary as outdated")
	}
	return &outdated, nil
}

// updateMetrics for visibility via Prometheus.
//
// Intended to be called from a defer so errors won't interrupt work, but will log if
// possible.
func (e *Events) updateMetrics(ctx context.Context, summaryType string, lag time.Duration) {
	QueueLag.WithLabelValues(summaryType).Observe(lag.Minutes())
	filter := bson.M{
		"type": summaryType,
		"time": bson.M{"$lte": time.Now().UTC()},
	}
	count, err := e.repo.CountDocuments(ctx, filter)
	if err != nil {
		e.log(ctx).WithError(err).Info("unable to update summary event queue length metric")
		return
	}
	QueueLength.WithLabelValues(summaryType).Set(float64(count))
}

func (e *Events) log(ctx context.Context) log.Logger {
	if ctxLog := log.LoggerFromContext(ctx); ctxLog != nil {
		return ctxLog
	}
	return null.NewLogger()
}

type SummaryEvent struct {
	UserID  string    `json:"userId" bson:"user_id"`
	Time    time.Time `json:"time" bson:"time"`
	Type    string    `json:"type" bson:"type"`
	Reasons []string  `json:"reasons" bson:"reasons"`
}
