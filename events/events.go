package events

import (
	"context"

	"github.com/Shopify/sarama"
	ev "github.com/tidepool-org/go-common/events"
)

type Runner interface {
	Initialize() error
	Run(context.Context) error
	Terminate()
}

type runner struct {
	cancel   context.CancelFunc
	consumer ev.EventConsumer
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
	consumer, err := ev.NewSaramaCloudEventsConsumer(config)
	if err != nil {
		return err
	}
	for _, handler := range u.handlers {
		consumer.RegisterHandler(handler)
	}
	u.consumer = consumer
	return nil
}

func (u *runner) Run(ctx context.Context) error {
	var consumerCtx context.Context
	consumerCtx, u.cancel = context.WithCancel(ctx)

	// blocks until Terminate() is invoked
	return u.consumer.Start(consumerCtx)
}

func (u *runner) Terminate() {
	if u.cancel != nil {
		u.cancel()
	}
}
