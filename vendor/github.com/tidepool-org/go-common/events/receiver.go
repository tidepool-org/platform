package events

import (
	"context"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type EventHandler interface {
	CanHandle(ce cloudevents.Event) bool
	Handle(ce cloudevents.Event) error
}

type EventConsumer interface {
	RegisterHandler(handler EventHandler)
	Start(ctx context.Context) error
}

var _ EventConsumer = &KafkaCloudEventsConsumer{}

type KafkaCloudEventsConsumer struct {
	client   cloudevents.Client
	consumer *kafka_sarama.Consumer
	handlers []EventHandler
}

func NewKafkaCloudEventsConsumer(config *CloudEventsConfig) (*KafkaCloudEventsConsumer, error) {
	if err := validateConsumerConfig(config); err != nil {
		return nil, err
	}

	consumer, err := kafka_sarama.NewConsumer(config.KafkaBrokers, config.SaramaConfig, config.KafkaConsumerGroup, config.GetPrefixedTopic())
	if err != nil {
		return nil, err
	}

	client, err := cloudevents.NewClient(consumer, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		return nil, err
	}

	return &KafkaCloudEventsConsumer{
		client:   client,
		consumer: consumer,
		handlers: make([]EventHandler, 0),
	}, nil
}

func (k *KafkaCloudEventsConsumer) RegisterHandler(handler EventHandler) {
	k.handlers = append(k.handlers, handler)
}

func (k *KafkaCloudEventsConsumer) Start(ctx context.Context) error {
	defer k.consumer.Close(context.Background())
	return k.client.StartReceiver(ctx, k.receive)
}

func (k *KafkaCloudEventsConsumer) receive(ce cloudevents.Event) {
	for _, handler := range k.handlers {
		if handler.CanHandle(ce) {
			_ = handler.Handle(ce)
		}
	}
}
