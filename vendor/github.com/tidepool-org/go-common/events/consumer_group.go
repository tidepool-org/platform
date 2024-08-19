package events

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/tidepool-org/go-common/errors"
	"log"
	"sync"
)

var ErrConsumerStopped = errors.New("consumer has been stopped")

type SaramaConsumerGroup struct {
	config        *CloudEventsConfig
	consumerGroup sarama.ConsumerGroup
	consumer      MessageConsumer
	ready         chan bool
	stop          chan struct{}
	stopOnce      *sync.Once
	topic         string
	wg            *sync.WaitGroup
}

func NewSaramaConsumerGroup(config *CloudEventsConfig, consumer MessageConsumer) (EventConsumer, error) {
	if err := validateConsumerConfig(config); err != nil {
		return nil, err
	}

	return &SaramaConsumerGroup{
		config:   config,
		consumer: consumer,
		topic:    config.GetPrefixedTopic(),
		wg:       &sync.WaitGroup{},
		ready:    make(chan bool, 1),
		stop:     make(chan struct{}),
		stopOnce: &sync.Once{},
	}, nil
}

func (s *SaramaConsumerGroup) Setup(session sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(s.ready)
	return nil
}

func (s *SaramaConsumerGroup) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (s *SaramaConsumerGroup) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if err := s.consumer.HandleKafkaMessage(message); err != nil {
			log.Printf("failed to process kafka message: %v", err)
			return err
		}
		session.MarkMessage(message, "")
	}

	return nil
}

func (s *SaramaConsumerGroup) Start() error {
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

func (s *SaramaConsumerGroup) Stop() error {
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

func (s *SaramaConsumerGroup) initialize() error {
	cg, err := sarama.NewConsumerGroup(
		s.config.KafkaBrokers,
		s.config.KafkaConsumerGroup,
		s.config.SaramaConfig,
	)
	if err != nil {
		return err
	}

	if err := s.consumer.Initialize(s.config); err != nil {
		return err
	}

	s.consumerGroup = cg
	return nil
}
