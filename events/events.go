package events

import (
	"sync"

	ev "github.com/tidepool-org/go-common/events"
)

type Runner interface {
	Initialize() error
	Run() error
	Terminate() error
}

type runner struct {
	consumer      *ev.FaultTolerantConsumerGroup
	handlers      []ev.EventHandler
	consumerGroup string
	mu            sync.Mutex
	stopped       bool
}

type Option func(*runner)

// WithConsumerGroup gives each component an independent subscription when
// multiple event runners share the same process environment.
func WithConsumerGroup(group string) Option {
	return func(r *runner) { r.consumerGroup = group }
}

func NewRunner(handlers []ev.EventHandler, options ...Option) Runner {
	r := &runner{
		handlers: handlers,
	}
	for _, option := range options {
		option(r)
	}
	return r
}

func (r *runner) loadConfig() (*ev.CloudEventsConfig, error) {
	config := ev.NewConfig()
	if err := config.LoadFromEnv(); err != nil {
		return nil, err
	}
	if r.consumerGroup != "" {
		config.KafkaConsumerGroup = r.consumerGroup
	}
	return config, nil
}

func (r *runner) Initialize() error {
	config, err := r.loadConfig()
	if err != nil {
		return err
	}
	consumer, err := ev.NewFaultTolerantConsumerGroup(config, func() (ev.MessageConsumer, error) {
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.stopped {
			return nil, ev.ErrConsumerStopped
		}
		return ev.NewCloudEventsMessageHandler(r.handlers)
	})
	if err != nil {
		return err
	}

	r.consumer = consumer
	return nil
}

func (r *runner) Run() error {
	return r.consumer.Start()
}

func (r *runner) Terminate() error {
	// The shared consumer returns early from Stop if Start has not created a
	// delegate yet. Guard its factory too, so a late Start cannot reopen it.
	r.mu.Lock()
	r.stopped = true
	r.mu.Unlock()
	if r.consumer != nil {
		return r.consumer.Stop()
	}
	return nil
}
