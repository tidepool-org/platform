package events

import (
	"context"
	"fmt"
	"os"
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
	logjson "github.com/tidepool-org/platform/log/json"
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

// SaramaRunner interfaces between [events.Runner] and go-common's
// [asyncevents.SaramaEventsConsumer].
//
// This means providing Initialize(), Run(), and Terminate() to satisfy events.Runner, while
// under the hood calling SaramaEventConsumer's Run(), and canceling its Context as
// appropriate.
type SaramaRunner struct {
	eventsRunner SaramaEventsRunner
	cancelCtx    context.CancelFunc
	cancelMu     sync.Mutex
}

func NewSaramaRunner(eventsRunner SaramaEventsRunner) *SaramaRunner {
	return &SaramaRunner{
		eventsRunner: eventsRunner,
	}
}

// SaramaEventsRunner is implemented by go-common's [asyncevents.SaramaEventsRunner].
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
	Topics          []string
	MessageConsumer asyncevents.SaramaMessageConsumer

	Sarama *sarama.Config
}

func (r *SaramaRunner) Initialize() error { return nil }

// Run adapts platform's event.Runner to work with go-common's
// asyncevents.SaramaEventsConsumer.
func (r *SaramaRunner) Run() error {
	if r.eventsRunner == nil {
		return errors.New("Unable to run SaramaRunner, eventsRunner is nil")
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
	if err := r.eventsRunner.Run(ctx); err != nil {
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

// CascadingSaramaEventsRunner manages multiple sarama consumer groups to execute a
// topic-cascading retry process.
//
// The topic names are generated from Config.Topics combined with Delays. If given a single
// topic "updates", and delays: 0s, 1s, and 5s, then the following topics will be consumed:
// updates, updates-retry-1s, updates-retry-5s. The consumer of the updates-retry-5s topic
// will write failed messages to updates-dead.
//
// The inspiration for this system was drawn from
// https://www.uber.com/blog/reliable-reprocessing/
type CascadingSaramaEventsRunner struct {
	Config         SaramaRunnerConfig
	Delays         []time.Duration
	Logger         log.Logger
	SaramaBuilders SaramaBuilders
}

func NewCascadingSaramaEventsRunner(config SaramaRunnerConfig, logger log.Logger,
	delays []time.Duration) *CascadingSaramaEventsRunner {

	return &CascadingSaramaEventsRunner{
		Config:         config,
		Delays:         delays,
		Logger:         logger,
		SaramaBuilders: DefaultSaramaBuilders{},
	}
}

// LimitedAsyncProducer restricts the [sarama.AsyncProducer] interface to ensure that its
// recipient isn't able to call Close(), thereby opening the potential for a panic when
// writing to a closed channel.
type LimitedAsyncProducer interface {
	AbortTxn() error
	BeginTxn() error
	CommitTxn() error
	Input() chan<- *sarama.ProducerMessage
}

func (r *CascadingSaramaEventsRunner) Run(ctx context.Context) error {
	if len(r.Config.Topics) == 0 {
		return errors.New("no topics")
	}
	if len(r.Delays) == 0 {
		return errors.New("no delays")
	}

	producersCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	errs := make(chan error, len(r.Config.Topics)*len(r.Delays))
	defer func() {
		r.logger(ctx).Debug("CascadingSaramaEventsRunner: waiting for consumers")
		wg.Wait()
		r.logger(ctx).Debug("CascadingSaramaEventsRunner: all consumers returned")
		close(errs)
	}()

	for _, topic := range r.Config.Topics {
		for idx, delay := range r.Delays {
			producerCfg := r.producerConfig(idx, delay)
			// The producer is built here rather than in buildConsumer() to control when
			// producer is closed. Were the producer to be closed before consumer.Run()
			// returns, it would be possible for consumer to write to the producer's
			// Inputs() channel, which if closed, would cause a panic.
			producer, err := r.SaramaBuilders.NewAsyncProducer(r.Config.Brokers, producerCfg)
			if err != nil {
				return errors.Wrapf(err, "Unable to build async producer: %s", r.Config.GroupID)
			}

			consumer, err := r.buildConsumer(producersCtx, idx, producer, delay, topic)
			if err != nil {
				return err
			}

			wg.Add(1)
			go func(topic string) {
				defer func() { wg.Done(); producer.Close() }()
				if err := consumer.Run(producersCtx); err != nil {
					errs <- fmt.Errorf("topics[%q]: %s", topic, err)
				}
				r.logger(ctx).WithField("topic", topic).
					Debug("CascadingSaramaEventsRunner: consumer go proc returning")
			}(topic)
		}
	}

	select {
	case <-ctx.Done():
		r.logger(ctx).Debug("CascadingSaramaEventsRunner: context is done")
		return nil
	case err := <-errs:
		r.logger(ctx).WithError(err).
			Debug("CascadingSaramaEventsRunner: Run(): error from consumer")
		return err
	}
}

func (r *CascadingSaramaEventsRunner) producerConfig(idx int, delay time.Duration) *sarama.Config {
	uniqueConfig := *r.Config.Sarama
	hostID := os.Getenv("HOSTNAME") // set by default in kubernetes pods
	if hostID == "" {
		hostID = fmt.Sprintf("%d-%d", time.Now().UnixNano()/int64(time.Second), os.Getpid())
	}
	txnID := fmt.Sprintf("%s-%s-%d-%s", r.Config.GroupID, delay.String(), idx, hostID)
	uniqueConfig.Producer.Transaction.ID = txnID
	uniqueConfig.Producer.Idempotent = true
	uniqueConfig.Producer.RequiredAcks = sarama.WaitForAll
	uniqueConfig.Net.MaxOpenRequests = 1
	uniqueConfig.Consumer.IsolationLevel = sarama.ReadCommitted
	return &uniqueConfig
}

// SaramaBuilders allows tests to inject mock objects.
type SaramaBuilders interface {
	NewAsyncProducer([]string, *sarama.Config) (sarama.AsyncProducer, error)
	NewConsumerGroup([]string, string, *sarama.Config) (sarama.ConsumerGroup, error)
}

// DefaultSaramaBuilders implements SaramaBuilders for normal, non-test use.
type DefaultSaramaBuilders struct{}

func (DefaultSaramaBuilders) NewAsyncProducer(brokers []string, config *sarama.Config) (
	sarama.AsyncProducer, error) {

	return sarama.NewAsyncProducer(brokers, config)
}

func (DefaultSaramaBuilders) NewConsumerGroup(brokers []string, groupID string,
	config *sarama.Config) (sarama.ConsumerGroup, error) {

	return sarama.NewConsumerGroup(brokers, groupID, config)
}

func (r *CascadingSaramaEventsRunner) buildConsumer(ctx context.Context, idx int,
	producer LimitedAsyncProducer, delay time.Duration, baseTopic string) (
	*asyncevents.SaramaEventsConsumer, error) {

	groupID := r.Config.GroupID
	if delay > 0 {
		groupID += "-retry-" + delay.String()
	}
	group, err := r.SaramaBuilders.NewConsumerGroup(r.Config.Brokers, groupID,
		r.Config.Sarama)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to build sarama consumer group: %s", groupID)
	}

	var consumer asyncevents.SaramaMessageConsumer = r.Config.MessageConsumer
	if len(r.Delays) > 0 {
		nextTopic := baseTopic + "-dead"
		if idx+1 < len(r.Delays) {
			nextTopic = baseTopic + "-retry-" + r.Delays[idx+1].String()
		}
		consumer = &CascadingConsumer{
			Consumer:  consumer,
			NextTopic: nextTopic,
			Producer:  producer,
			Logger:    r.Logger,
		}
	}
	if delay > 0 {
		consumer = &DelayingConsumer{
			Consumer: consumer,
			Delay:    delay,
			Logger:   r.Logger,
		}
	}
	handler := asyncevents.NewSaramaConsumerGroupHandler(consumer,
		AlertsEventConsumptionTimeout)
	topic := baseTopic
	if delay > 0 {
		topic += "-retry-" + delay.String()
	}
	r.logger(ctx).WithField("topic", topic).Debug("creating consumer")

	return asyncevents.NewSaramaEventsConsumer(group, handler, topic), nil
}

func (r *CascadingSaramaEventsRunner) logger(ctx context.Context) log.Logger {
	// A context logger might have more fields or ... context. So prefer that if availble.
	if ctxLogger := log.LoggerFromContext(ctx); ctxLogger != nil {
		return ctxLogger
	}
	if r.Logger == nil {
		// logjson.NewLogger will only fail if an argument is missing.
		r.Logger, _ = logjson.NewLogger(os.Stderr, log.DefaultLevelRanks(), log.DefaultLevel())
	}
	return r.Logger
}

// DelayingConsumer injects a delay before consuming a message.
type DelayingConsumer struct {
	Consumer asyncevents.SaramaMessageConsumer
	Delay    time.Duration
	Logger   log.Logger
}

func (c *DelayingConsumer) Consume(ctx context.Context, session sarama.ConsumerGroupSession,
	msg *sarama.ConsumerMessage) error {

	select {
	case <-ctx.Done():
		if ctxErr := ctx.Err(); ctxErr != context.Canceled {
			return ctxErr
		}
		return nil
	case <-time.After(c.Delay):
		c.Logger.WithFields(log.Fields{"topic": msg.Topic, "delay": c.Delay}).Debugf("delayed")
		return c.Consumer.Consume(ctx, session, msg)
	}
}

// CascadingConsumer cascades messages that failed to be consumed to another topic.
type CascadingConsumer struct {
	Consumer  asyncevents.SaramaMessageConsumer
	NextTopic string
	Producer  LimitedAsyncProducer
	Logger    log.Logger
}

func (c *CascadingConsumer) Consume(ctx context.Context, session sarama.ConsumerGroupSession,
	msg *sarama.ConsumerMessage) (err error) {

	if err := c.Consumer.Consume(ctx, session, msg); err != nil {
		txnErr := c.withTxn(func() error {
			select {
			case <-ctx.Done():
				if ctxErr := ctx.Err(); ctxErr != context.Canceled {
					return ctxErr
				}
				return nil
			case c.Producer.Input() <- c.cascadeMessage(msg):
				fields := log.Fields{"from": msg.Topic, "to": c.NextTopic}
				c.Logger.WithFields(fields).Debug("cascaded")
				return nil
			}
		})
		if txnErr != nil {
			c.Logger.WithError(txnErr).Info("Unable to complete cascading transaction")
			return err
		}
	}
	return nil
}

// withTxn wraps a function with a transaction that is aborted if an error is returned.
func (c *CascadingConsumer) withTxn(f func() error) (err error) {
	if err := c.Producer.BeginTxn(); err != nil {
		return errors.Wrap(err, "Unable to begin transaction")
	}
	defer func(err *error) {
		if err != nil && *err != nil {
			if abortErr := c.Producer.AbortTxn(); abortErr != nil {
				c.Logger.WithError(abortErr).Info("Unable to abort transaction")
			}
			return
		}
		if commitErr := c.Producer.CommitTxn(); commitErr != nil {
			c.Logger.WithError(commitErr).Info("Unable to commit transaction")
		}
	}(&err)
	return f()
}

func (c *CascadingConsumer) cascadeMessage(msg *sarama.ConsumerMessage) *sarama.ProducerMessage {
	pHeaders := make([]sarama.RecordHeader, len(msg.Headers))
	for idx, header := range msg.Headers {
		pHeaders[idx] = *header
	}
	return &sarama.ProducerMessage{
		Key:     sarama.ByteEncoder(msg.Key),
		Value:   sarama.ByteEncoder(msg.Value),
		Topic:   c.NextTopic,
		Headers: pHeaders,
	}
}
