package events

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/avast/retry-go"
)

var (
	DefaultAttempts  = uint(1000)
	DefaultDelay     = 30 * time.Second
	DefaultDelayType = retry.FixedDelay
)

type FaultTolerantConsumerGroup struct {
	config         *CloudEventsConfig
	createConsumer ConsumerFactory
	m              sync.Mutex
	delegate       EventConsumer
	isShuttingDown bool
	attempts       uint
	delay          time.Duration
	delayType      retry.DelayTypeFunc
}

var _ EventConsumer = &FaultTolerantConsumerGroup{}

func NewFaultTolerantConsumerGroup(config *CloudEventsConfig, createConsumer ConsumerFactory) (*FaultTolerantConsumerGroup, error) {
	return &FaultTolerantConsumerGroup{
		config:         config,
		createConsumer: createConsumer,
		attempts:       DefaultAttempts,
		delay:          DefaultDelay,
		delayType:      DefaultDelayType,
	}, nil
}

func (f *FaultTolerantConsumerGroup) Start() error {
	return retry.Do(
		f.restart,
		retry.Attempts(f.attempts),
		retry.Delay(f.delay),
		retry.DelayType(f.delayType),
	)
}

func (f *FaultTolerantConsumerGroup) restart() error {
	if err := f.recreateConsumer(); err != nil {
		return err
	}

	err := f.delegate.Start()
	log.Printf("Consumer exited. Reason: %v", err)

	if errors.Is(err, ErrConsumerStopped) {
		return retry.Unrecoverable(err)
	}

	return err
}

func (f *FaultTolerantConsumerGroup) recreateConsumer() error {
	f.m.Lock()
	defer f.m.Unlock()

	// Do not try to restart the consumer if we're trying to shut it down
	if f.isShuttingDown {
		return retry.Unrecoverable(ErrConsumerStopped)
	}

	consumer, err := f.createConsumer()
	if err != nil {
		return retry.Unrecoverable(err)
	}

	delegate, err := NewSaramaConsumerGroup(f.config, consumer)
	if err != nil {
		return retry.Unrecoverable(err)
	}

	f.delegate = delegate

	return nil
}

func (f *FaultTolerantConsumerGroup) Stop() error {
	f.m.Lock()
	defer f.m.Unlock()

	if f.delegate == nil {
		return nil
	}

	f.isShuttingDown = true
	err := f.delegate.Stop()
	return err
}
