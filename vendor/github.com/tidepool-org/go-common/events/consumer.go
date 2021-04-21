package events

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/tidepool-org/go-common/errors"
	"log"
	"sync"
)

var ErrConsumerStopped = errors.New("consumer has been stopped")

type EventConsumer interface {
	RegisterHandler(handler EventHandler)
	Start() error
	Stop() error
}

type SaramaConsumer struct {
	config             *CloudEventsConfig
	consumerGroup      sarama.ConsumerGroup
	ready              chan bool
	stop               chan struct{}
	stopOnce           *sync.Once
	topic              string
	wg                 *sync.WaitGroup
	handlers           []EventHandler
	deadLetterProducer *KafkaCloudEventsProducer
}

func NewSaramaCloudEventsConsumer(config *CloudEventsConfig) (EventConsumer, error) {
	if err := validateConsumerConfig(config); err != nil {
		return nil, err
	}

	return &SaramaConsumer{
		config:   config,
		topic:    config.GetPrefixedTopic(),
		wg:       &sync.WaitGroup{},
		ready:    make(chan bool, 1),
		stop:     make(chan struct{}),
		stopOnce: &sync.Once{},
		handlers: make([]EventHandler, 0),
	}, nil
}

func (s *SaramaConsumer) Setup(session sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(s.ready)
	return nil
}

func (s *SaramaConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (s *SaramaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		m := kafka_sarama.NewMessageFromConsumerMessage(message)
		// just ignore non-cloud event messages
		if rs, rserr := binding.ToEvent(context.Background(), m); rserr == nil {
			s.handleCloudEvent(*rs)
		}
		session.MarkMessage(message, "")
	}

	return nil
}

func (s *SaramaConsumer) handleCloudEvent(ce cloudevents.Event) {
	var errors []error
	for _, handler := range s.handlers {
		if handler.CanHandle(ce) {
			if err := handler.Handle(ce); err != nil {
				errors = append(errors, err)
			}
		}
	}
	if len(errors) != 0 {
		log.Printf("Sending event %v to dead-letter topic due to handler error(s): %v", ce.ID(), errors)
		s.sendToDeadLetterTopic(ce)
	}
}

func (s *SaramaConsumer) sendToDeadLetterTopic(ce cloudevents.Event) {
	if err := s.deadLetterProducer.SendCloudEvent(context.Background(), ce); err != nil {
		log.Printf("Failed to send event %v to dead-letter topic: %v", ce, err)
	}
}

func (s *SaramaConsumer) RegisterHandler(handler EventHandler) {
	s.handlers = append(s.handlers, handler)
}

func (s *SaramaConsumer) Start() error {
	if err := s.initialize(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-s.stop
		cancel()
	}()

	errChan := make(chan error)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := s.consumerGroup.Consume(ctx, []string{s.topic}, s); err != nil {
				log.Printf("Error from consumer: %v", err)
				// It's not clear whether this condition can be true
				if err == context.Canceled {
					err = ErrConsumerStopped
				}
				errChan <- err
				return
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				errChan <- ErrConsumerStopped
				return
			}
			s.ready = make(chan bool)
		}
	}()

	err := <-errChan
	if err == ErrConsumerStopped {
		return err
	}

	// The consumer group was terminated with an unexpected error.
	// We need to call stop, so we cancel the context and stop the
	// go routine so it doesn't leak on restart.
	if e := s.Stop(); e != nil {
		err = fmt.Errorf("%w: %s", err, e.Error())
	}

	return err
}

func (s *SaramaConsumer) Stop() error {
	// Initialization failed
	if s.consumerGroup == nil {
		return nil
	}

	// Signal that the consumer group should be terminated
	s.stopOnce.Do(func() {
		s.stop <- struct{}{}
	})

	// Wait for the consumer group to exit
	s.wg.Wait()
	return s.consumerGroup.Close()
}

func (s *SaramaConsumer) initialize() error {
	cg, err := sarama.NewConsumerGroup(
		s.config.KafkaBrokers,
		s.config.KafkaConsumerGroup,
		s.config.SaramaConfig,
	)
	if err != nil {
		return err
	}

	if s.config.IsDeadLettersEnabled() {
		s.deadLetterProducer, err = NewKafkaCloudEventsProducerForDeadLetters(s.config)
		if err != nil {
			return err
		}
	}

	s.consumerGroup = cg
	return nil
}
