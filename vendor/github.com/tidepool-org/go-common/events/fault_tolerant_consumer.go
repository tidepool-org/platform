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

type FaultTolerantConsumer struct {
	config         *CloudEventsConfig
	handlers       []EventHandler
	m              sync.Mutex
	delegate       EventConsumer
	isShuttingDown bool
	attempts       uint
	delay          time.Duration
	delayType      retry.DelayTypeFunc
}

var _ EventConsumer = &FaultTolerantConsumer{}

func NewFaultTolerantCloudEventsConsumer(config *CloudEventsConfig) (*FaultTolerantConsumer, error) {
	return &FaultTolerantConsumer{
		config:    config,
		attempts:  DefaultAttempts,
		delay:     DefaultDelay,
		delayType: DefaultDelayType,
	}, nil
}

func (f *FaultTolerantConsumer) RegisterHandler(handler EventHandler) {
	f.handlers = append(f.handlers, handler)
}

func (f *FaultTolerantConsumer) Start() error {
	return retry.Do(
		f.restart,
		retry.Attempts(f.attempts),
		retry.Delay(f.delay),
		retry.DelayType(f.delayType),
	)
}

func (f *FaultTolerantConsumer) restart() error {
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

func (f *FaultTolerantConsumer) recreateConsumer() error {
	f.m.Lock()
	defer f.m.Unlock()

	// Do not try to restart the consumer if we're trying to shut it down
	if f.isShuttingDown {
		return retry.Unrecoverable(ErrConsumerStopped)
	}

	delegate, err := NewSaramaCloudEventsConsumer(f.config)
	if err != nil {
		return retry.Unrecoverable(err)
	}

	f.delegate = delegate
	for _, h := range f.handlers {
		f.delegate.RegisterHandler(h)
	}
	return nil
}

func (f *FaultTolerantConsumer) Stop() error {
	f.m.Lock()
	defer f.m.Unlock()

	if f.delegate == nil {
		return nil
	}

	f.isShuttingDown = true
	err := f.delegate.Stop()
	return err
}
