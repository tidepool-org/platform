package task

import (
	"context"
	"errors"
	"fmt"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/summary/types"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"math/rand"
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/task"
)

const ConfigVersion = 1

var SummaryTypes = []string{"cgm", "bgm", "con"}

type MinuteRange struct {
	Min int
	Max int
}

type Configuration struct {
	Interval MinuteRange `json:"interval" bson:"interval"`
	Batch    *int        `json:"batch,omitempty" bson:"batch,omitempty"`
	Version  int         `json:"version" bson:"version"`
}

type AuthClient interface {
	ServerSessionToken() (string, error)
}

type DataClient interface {
	GetMigratableUserIDs(ctx context.Context, typ string, pagination *page.Pagination) ([]string, error)
	UpdateCGMSummary(ctx context.Context, id string) (*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket], error)
	UpdateBGMSummary(ctx context.Context, id string) (*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket], error)
	UpdateContinuousSummary(ctx context.Context, id string) (*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket], error)
	GetOutdatedUserIDs(ctx context.Context, typ string, pagination *page.Pagination) (*types.OutdatedSummariesResponse, error)
}

type Provider interface {
	AuthClient() AuthClient
	DataClient() DataClient
	SummaryType() string
	GetRunnerDurationMaximum() time.Duration
}

func ValidateConfig(config Configuration, withBatch bool) error {
	if config.Version != ConfigVersion {
		return errors.New("old version number, must be remade")
	}
	if config.Interval.Min < 1 {
		return errors.New("minimum Interval cannot be <1 minute")
	}
	if config.Interval.Max < config.Interval.Min {
		return errors.New("maximum Interval cannot be less than Minimum Interval")
	}

	if withBatch == true {
		if config.Batch == nil {
			return errors.New("batch is required but not provided")
		}
		if *config.Batch < 1 {
			return errors.New("batch can not be <1")
		}
	} else {
		if config.Batch != nil {
			return errors.New("batch is not required, but was provided")
		}
	}

	return nil
}

func GenerateNextTime(interval MinuteRange) time.Duration {
	Min := time.Duration(interval.Min) * time.Second
	Max := time.Duration(interval.Max) * time.Second

	randTime := time.Duration(rand.Int63n(int64(Max - Min + 1)))
	return Min + randTime
}

func NewDefaultUpdateConfig() Configuration {
	return Configuration{
		Interval: MinuteRange{
			int(DefaultUpdateAvailableAfterDurationMinimum.Seconds()),
			int(DefaultUpdateAvailableAfterDurationMaximum.Seconds())},
		Batch:   pointer.FromAny(DefaultUpdateWorkerBatchSize),
		Version: ConfigVersion,
	}
}

func NewDefaultUpdateTaskCreate(summaryType string) *task.TaskCreate {
	return &task.TaskCreate{
		Name:          pointer.FromAny(fmt.Sprintf(UpdateType, summaryType)),
		Type:          fmt.Sprintf(UpdateType, summaryType),
		Priority:      5,
		AvailableTime: pointer.FromAny(time.Now().UTC()),
		Data: map[string]interface{}{
			"config": NewDefaultUpdateConfig(),
		},
	}
}

func NewDefaultMigrationConfig() Configuration {
	return Configuration{
		Interval: MinuteRange{
			int(DefaultMigrationAvailableAfterDurationMinimum.Seconds()),
			int(DefaultMigrationAvailableAfterDurationMaximum.Seconds())},
		Batch:   pointer.FromAny(DefaultMigrationWorkerBatchSize),
		Version: ConfigVersion,
	}
}

func NewDefaultMigrationTaskCreate(summaryType string) *task.TaskCreate {
	return &task.TaskCreate{
		Name:          pointer.FromAny(fmt.Sprintf(MigrationType, summaryType)),
		Type:          fmt.Sprintf(MigrationType, summaryType),
		Priority:      5,
		AvailableTime: pointer.FromAny(time.Now().UTC()),
		Data: map[string]interface{}{
			"config": NewDefaultMigrationConfig(),
		},
	}
}

func updateSummaries(ctx context.Context, dataClient DataClient, typ string, outdatedUserIds []string) error {
	eg, ctx := errgroup.WithContext(ctx)
	logger := log.LoggerFromContext(ctx)

	sem := semaphore.NewWeighted(UpdateWorkerCount)
	for _, userId := range outdatedUserIds {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}

		// we can't pass arguments to errgroup goroutines
		// we need to explicitly redefine the variables,
		// because we're launching the goroutines in a loop
		userId := userId
		eg.Go(func() error {
			defer sem.Release(1)
			logger.WithField("userId", userId).Debugf("Migrating User %s Summary", typ)

			// update summary
			err := updateSummary(ctx, dataClient, typ, userId)
			if err != nil {
				return err
			}

			logger.WithField("userId", userId).Debugf("Finished Migrating User %s Summary", typ)

			return nil
		})
	}
	return eg.Wait()
}

func updateSummary(ctx context.Context, dataClient DataClient, typ string, userId string) (err error) {
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
