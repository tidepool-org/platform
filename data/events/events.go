package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/IBM/sarama"

	"github.com/tidepool-org/go-common/asyncevents"
	ev "github.com/tidepool-org/go-common/events"

	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
	dataStore "github.com/tidepool-org/platform/data/store"
	summaryStore "github.com/tidepool-org/platform/data/summary/store"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type userDeletionEventsHandler struct {
	ev.NoopUserEventsHandler

	ctx             context.Context
	dataStore       dataStore.Store
	dataSourceStore dataSourceStoreStructured.Store
}

func NewUserDataDeletionHandler(ctx context.Context, dataStore dataStore.Store, dataSourceStore dataSourceStoreStructured.Store) ev.EventHandler {
	return ev.NewUserEventsHandler(&userDeletionEventsHandler{
		ctx:             ctx,
		dataStore:       dataStore,
		dataSourceStore: dataSourceStore,
	})
}

func (u *userDeletionEventsHandler) HandleDeleteUserEvent(payload ev.DeleteUserEvent) error {
	var errs []error
	logger := log.LoggerFromContext(u.ctx).WithField("userId", payload.UserID)

	logger.Infof("Deleting data for user")
	dataRepository := u.dataStore.NewDataRepository()
	if err := dataRepository.DestroyDataForUserByID(u.ctx, payload.UserID); err != nil {
		errs = append(errs, err)
		logger.WithError(err).Error("unable to delete data for user")
	}

	logger.Infof("Deleting data source for user")
	dataSourceRepository := u.dataSourceStore.NewDataSourcesRepository()
	if _, err := dataSourceRepository.DestroyAll(u.ctx, payload.UserID); err != nil {
		errs = append(errs, err)
		logger.WithError(err).Error("unable to delete data sources for user")
	}

	logger.Infof("Deleting summary for user")
	summaryRepository := summaryStore.NewTypeless(u.dataStore.NewSummaryRepository().GetStore())
	if err := summaryRepository.DeleteSummary(u.ctx, payload.UserID); err != nil {
		errs = append(errs, err)
		logger.WithError(err).Error("unable to delete summary for user")
	}

	if len(errs) != 0 {
		return errors.New("Unable to delete device data for user")
	}
	return nil
}

// AlertsEventRetryDelayMaximum is the maximum delay between consumption
// retries.
const AlertsEventRetryDelayMaximum = time.Minute

// AlertsEventRetries is the maximum consumption attempts before giving up.
const AlertsEventRetries = 1000

// AlertsEventConsumptionTimeout is the maximum time to process an alerts event.
const AlertsEventConsumptionTimeout = 30 * time.Second

// SaramaRunner interfaces between events.Runner and go-common's
// asyncevents.SaramaEventsConsumer.
type SaramaRunner struct {
	EventsRunner SaramaEventsRunner
	Config       SaramaRunnerConfig
	cancelCtx    context.CancelFunc
	cancelMu     sync.Mutex
}

// SaramaEventsRunner is implemented by go-common's
// asyncevents.SaramaEventsRunner.
type SaramaEventsRunner interface {
	Run(ctx context.Context) error
}

// SaramaRunnerConfig collects values needed to initialize a SaramaRunner.
//
// This provides isolation for the SaramaRunner from ConfigReporter,
// envconfig, or any of the other options in platform for reading config
// values.
type SaramaRunnerConfig struct {
	Brokers         []string
	GroupID         string
	Logger          log.Logger
	Topics          []string
	MessageConsumer asyncevents.SaramaMessageConsumer

	Sarama *sarama.Config
}

func (r *SaramaRunner) Initialize() error {
	group, err := sarama.NewConsumerGroup(r.Config.Brokers, r.Config.GroupID, r.Config.Sarama)
	if err != nil {
		return errors.Wrap(err, "Unable to build sarama consumer group")
	}
	handler := asyncevents.NewSaramaConsumerGroupHandler(&asyncevents.NTimesRetryingConsumer{
		Consumer: r.Config.MessageConsumer,
		Delay:    CappedExponentialBinaryDelay(AlertsEventRetryDelayMaximum),
		Times:    AlertsEventRetries,
		Logger:   r.logger,
	}, AlertsEventConsumptionTimeout)
	r.EventsRunner = asyncevents.NewSaramaEventsConsumer(group, handler, r.Config.Topics...)
	return nil
}

func (r *SaramaRunner) logger(ctx context.Context) asyncevents.Logger {
	// Prefer a logger from the context.
	if ctxLogger := log.LoggerFromContext(ctx); ctxLogger != nil {
		return &log.GoCommonAdapter{Logger: ctxLogger}
	}
	if r.Config.Logger != nil {
		return &log.GoCommonAdapter{Logger: r.Config.Logger}
	}
	// No known log.Logger could be found, default to slog.
	return slog.Default()
}

// Run adapts platform's event.Runner to work with go-common's
// asyncevents.SaramaEventsConsumer.
func (r *SaramaRunner) Run() error {
	if r.EventsRunner == nil {
		return errors.New("Unable to run SaramaRunner, EventsRunner is nil")
	}

	r.cancelMu.Lock()
	ctx, err := func() (context.Context, error) {
		defer r.cancelMu.Unlock()
		if r.cancelCtx != nil {
			return nil, errors.New("Unable to Run SaramaRunner, it's already initialized")
		}
		var ctx context.Context
		ctx, r.cancelCtx = context.WithCancel(context.Background())
		return ctx, nil
	}()
	if err != nil {
		return err
	}
	if err := r.EventsRunner.Run(ctx); err != nil {
		return errors.Wrap(err, "Unable to Run SaramaRunner")
	}
	return nil
}

// Terminate adapts platform's event.Runner to work with go-common's
// asyncevents.SaramaEventsConsumer.
func (r *SaramaRunner) Terminate() error {
	r.cancelMu.Lock()
	defer r.cancelMu.Unlock()
	if r.cancelCtx == nil {
		return errors.New("Unable to Terminate SaramaRunner, it's not running")
	}
	r.cancelCtx()
	return nil
}

// CappedExponentialBinaryDelay builds delay functions that use exponential
// binary backoff with a maximum duration.
func CappedExponentialBinaryDelay(cap time.Duration) func(int) time.Duration {
	return func(tries int) time.Duration {
		b := asyncevents.DelayExponentialBinary(tries)
		if b > cap {
			return cap
		}
		return b
	}
}

// TODO: implement me!!
type AlertsEventsConsumer struct {
	Consumer asyncevents.SaramaMessageConsumer
}

func (c *AlertsEventsConsumer) Consume(ctx context.Context,
	session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) error {
	err := c.Consumer.Consume(ctx, session, message)
	if err != nil {
		session.MarkMessage(message, fmt.Sprintf("I have given up after error: %s", err))
		return err
	}
	return nil
}
