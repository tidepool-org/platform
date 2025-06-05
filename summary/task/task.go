package task

import (
	"context"
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/task/queue"
)

const (
	ConfigMinInterval = "minInterval"
	ConfigMaxInterval = "maxInterval"
	ConfigBatch       = "batch"
)

var SummaryTypes = []string{"cgm", "bgm", "con"}

func NewSummaryRunners(authClient auth.Client, dataClient dataClient.Client, logger log.Logger) ([]queue.Runner, error) {
	var runners []queue.Runner

	for _, typ := range SummaryTypes {
		logger.Debugf("Creating %s summary update runner", typ)
		summaryUpdateRnnr, summaryUpdateRnnrErr := NewUpdateRunner(logger, authClient, dataClient, typ)
		if summaryUpdateRnnrErr != nil {
			return nil, errors.Wrapf(summaryUpdateRnnrErr, "unable to create %s summary update runner", typ)
		}
		runners = append(runners, summaryUpdateRnnr)

		logger.Debugf("Creating %s summary migration runner", typ)
		summaryMigrationRnnr, summaryMigrationRnnrErr := NewMigrationRunner(logger, authClient, dataClient, typ)
		if summaryMigrationRnnrErr != nil {
			return nil, errors.Wrapf(summaryMigrationRnnrErr, "unable to create %s summary migration runner", typ)
		}
		runners = append(runners, summaryMigrationRnnr)
	}

	return runners, nil
}

func GenerateNextTime(minSeconds int, maxSeconds int) time.Duration {
	Min := time.Duration(minSeconds) * time.Second
	Max := time.Duration(maxSeconds) * time.Second

	randTime := time.Duration(rand.Int63n(int64(Max - Min + 1)))
	return Min + randTime
}

func updateSummaries(ctx context.Context, logger log.Logger, dataClient dataClient.Client, typ string, outdatedUserIds []string, workerCount int64, deadline time.Time, callerType string) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(workerCount)

	for _, userId := range outdatedUserIds {
		if time.Now().After(deadline) {
			return context.DeadlineExceeded
		}
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}

		// we can't pass arguments to errgroup goroutines
		// we need to explicitly redefine the variables,
		// because we're launching the goroutines in a loop
		userId := userId
		eg.Go(func() error {
			defer sem.Release(1)
			logger.WithField("userId", userId).Debugf("%s User %s Summary", callerType, typ)

			// update summary
			err := updateSummary(ctx, dataClient, typ, userId)
			if err != nil {
				return err
			}

			logger.WithField("userId", userId).Debugf("Finished %s User %s Summary", callerType, typ)

			return nil
		})
	}
	return eg.Wait()
}

func updateSummary(ctx context.Context, dataClient dataClient.Client, typ string, userId string) (err error) {
	switch typ {
	case "cgm":
		_, err = dataClient.UpdateCGMSummary(ctx, userId)
	case "bgm":
		_, err = dataClient.UpdateBGMSummary(ctx, userId)
	case "con":
		_, err = dataClient.UpdateContinuousSummary(ctx, userId)
	default:
		err = errors.New("summary type unsupported by updateSummary")
	}
	return err
}

func isGtZero[T int | int32 | int64](v T) bool {
	return v > 0
}

func ValueFromTaskDataWithFallback[T any](data map[string]any, property string, isValid func(T) bool, fallback T) T {
	value, ok := data[property].(T)
	if !ok || !isValid(value) {
		value = fallback
		data[property] = value
	}

	return value
}
