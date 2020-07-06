package mailer

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/kelseyhightower/envconfig"
	"github.com/tidepool-org/mailer/mailer"
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/log"
)

const DefaultCloseTimeoutMs = 30000 // 30 seconds

var Module = fx.Provide(NewEmailDeliveryChannel, NewDeliveryErrorLogger, NewKafkaMailer)

type EmailDeliveryChannel chan kafka.Event

func NewEmailDeliveryChannel() EmailDeliveryChannel {
	return make(chan kafka.Event)
}

type KafkaMailerParams struct {
	fx.In
	fx.Lifecycle

	DeliveryChan        EmailDeliveryChannel
	DeliveryErrorLogger *DeliveryErrorLogger
}

func NewKafkaMailer(params KafkaMailerParams) (mailer.Mailer, error) {
	kafkaMailerConfig := &mailer.KafkaMailerConfig{}
	err := envconfig.Process("", kafkaMailerConfig)
	if err != nil {
		return nil, err
	}

	kafkaMailer, err := mailer.NewKafkaMailer(kafkaMailerConfig, params.DeliveryChan)
	if err != nil {
		return nil, err
	}

	errorLoggerCtx, errorLoggerCtxCancel := context.WithCancel(context.Background())
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go params.DeliveryErrorLogger.Run(errorLoggerCtx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			timeoutMs := DefaultCloseTimeoutMs
			deadline, ok := ctx.Deadline()
			if ok {
				timeoutMs = int(time.Now().Sub(deadline).Seconds() * 1000)
			}
			err := kafkaMailer.Close(timeoutMs)
			errorLoggerCtxCancel()
			return err
		},
	})

	return kafkaMailer, err
}

type DeliveryErrorLoggerParams struct {
	fx.In
	fx.Lifecycle

	DeliveryChan EmailDeliveryChannel
	Logger       log.Logger
}

type DeliveryErrorLogger struct {
	deliveryChan EmailDeliveryChannel
	logger       log.Logger
}

func NewDeliveryErrorLogger(params DeliveryErrorLoggerParams) *DeliveryErrorLogger {
	return &DeliveryErrorLogger{
		deliveryChan: params.DeliveryChan,
		logger:       params.Logger,
	}
}

func (d *DeliveryErrorLogger) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-d.deliveryChan:
			msg, ok := event.(*kafka.Message)
			if !ok {
				d.logger.WithField("event", event).Warn("unexpected kafka event on delivery channel")
			} else if msg.TopicPartition.Error != nil {
				d.logger.WithField("message", msg).Errorf("error delivering kafka email message: %v", msg.TopicPartition.Error.Error())
			}
		}
	}
}
