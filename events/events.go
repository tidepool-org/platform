package events

import (
	"github.com/Shopify/sarama"
	ev "github.com/tidepool-org/go-common/events"
)

type Runner interface {
	Initialize() error
	Run() error
	Terminate() error
}

type runner struct {
	consumer *ev.FaultTolerantConsumer
	handlers []ev.EventHandler
}

func NewRunner(handlers []ev.EventHandler) Runner {
	return &runner{
		handlers: handlers,
	}
}

func (u *runner) Initialize() error {
	config := ev.NewConfig()
	if err := config.LoadFromEnv(); err != nil {
		return err
	}
	config.SaramaConfig.Version = sarama.V2_6_0_0
	consumer, err := ev.NewFaultTolerantCloudEventsConsumer(config)
	if err != nil {
		return err
	}
	for _, handler := range u.handlers {
		consumer.RegisterHandler(handler)
	}
	u.consumer = consumer
	return nil
}

func (u *runner) Run() error {
	return u.consumer.Start()
}

func (u *runner) Terminate() error {
	if u.consumer != nil {
		return u.consumer.Stop()
	}
	return nil
}
