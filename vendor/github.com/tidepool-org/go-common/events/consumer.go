package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"log"
)

type EventConsumer interface {
	Start() error
	Stop() error
}

type ConsumerFactory func() (MessageConsumer, error)

type MessageConsumer interface {
	Initialize(config *CloudEventsConfig) error
	HandleKafkaMessage(cm *sarama.ConsumerMessage) error
}

type CloudEventsMessageConsumer struct {
	handlers           []EventHandler
	deadLetterProducer *KafkaCloudEventsProducer
}

func NewCloudEventsMessageHandler(handlers []EventHandler) (*CloudEventsMessageConsumer, error) {
	return &CloudEventsMessageConsumer{
		handlers: handlers,
	}, nil
}

func (c *CloudEventsMessageConsumer) Initialize(config *CloudEventsConfig) error {
	if config.IsDeadLettersEnabled() {
		producer, err := NewKafkaCloudEventsProducerForDeadLetters(config)
		if err != nil {
			return err
		}
		c.deadLetterProducer = producer
	}

	return nil
}

func (c *CloudEventsMessageConsumer) HandleKafkaMessage(cm *sarama.ConsumerMessage) error {
	message := kafka_sarama.NewMessageFromConsumerMessage(cm)
	if rs, rserr := binding.ToEvent(context.Background(), message); rserr == nil {
		c.handleCloudEvent(*rs)
	}
	return nil
}

func (c *CloudEventsMessageConsumer) handleCloudEvent(ce cloudevents.Event) {
	var errors []error
	for _, handler := range c.handlers {
		if handler.CanHandle(ce) {
			if err := handler.Handle(ce); err != nil {
				errors = append(errors, err)
			}
		}
	}
	if len(errors) != 0 {
		log.Printf("Sending event %v to dead-letter topic due to handler error(c): %v", ce.ID(), errors)
		c.sendToDeadLetterTopic(ce)
	}
}

func (c *CloudEventsMessageConsumer) sendToDeadLetterTopic(ce cloudevents.Event) {
	if err := c.deadLetterProducer.SendCloudEvent(context.Background(), ce); err != nil {
		log.Printf("Failed to send event %v to dead-letter topic: %v", ce, err)
	}
}
