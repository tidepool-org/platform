package events

import (
	ev "github.com/tidepool-org/go-common/events"
)

type Runner interface {
	Initialize() error
	Run() error
	Terminate() error
}

type runner struct {
	consumer *ev.FaultTolerantConsumerGroup
	handlers []ev.EventHandler
}

func NewRunner(handlers []ev.EventHandler) Runner {
	return &runner{
		handlers: handlers,
	}
}

func (r *runner) Initialize() error {
	config := ev.NewConfig()
	if err := config.LoadFromEnv(); err != nil {
		return err
	}
	consumer, err := ev.NewFaultTolerantConsumerGroup(config, func() (ev.MessageConsumer, error) {
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
	if r.consumer != nil {
		return r.consumer.Stop()
	}
	return nil
}
