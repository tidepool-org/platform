package asyncevents

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/IBM/sarama"
)

// SaramaEventsConsumer consumes Kafka messages for asynchronous event
// handling.
type SaramaEventsConsumer struct {
	Handler       sarama.ConsumerGroupHandler
	ConsumerGroup sarama.ConsumerGroup
	Topics        []string
}

func NewSaramaEventsConsumer(consumerGroup sarama.ConsumerGroup,
	handler sarama.ConsumerGroupHandler, topics ...string) *SaramaEventsConsumer {

	return &SaramaEventsConsumer{
		ConsumerGroup: consumerGroup,
		Handler:       handler,
		Topics:        topics,
	}
}

// Run the consumer, to begin consuming Kafka messages.
//
// Run is stopped by its context being canceled. When its context is canceled,
// it returns nil.
func (p *SaramaEventsConsumer) Run(ctx context.Context) (err error) {
	for {
		err := p.ConsumerGroup.Consume(ctx, p.Topics, p.Handler)
		if err != nil {
			return err
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil
		}
	}
}

// SaramaConsumerGroupHandler implements sarama.ConsumerGroupHandler.
type SaramaConsumerGroupHandler struct {
	Consumer        SaramaMessageConsumer
	ConsumerTimeout time.Duration
	Logger          Logger
}

// NewSaramaConsumerGroupHandler builds a consumer group handler.
//
// A timeout of 0 will use DefaultMessageConsumptionTimeout.
func NewSaramaConsumerGroupHandler(logger Logger, consumer SaramaMessageConsumer,
	timeout time.Duration) *SaramaConsumerGroupHandler {

	if timeout == 0 {
		timeout = DefaultMessageConsumptionTimeout
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SaramaConsumerGroupHandler{
		Consumer:        consumer,
		ConsumerTimeout: timeout,
		Logger:          logger,
	}
}

const (
	// DefaultMessageConsumptionTimeout is the default time to allow
	// SaramaMessageConsumer.Consume to work before canceling.
	DefaultMessageConsumptionTimeout = 30 * time.Second
)

// Setup implements sarama.ConsumerGroupHandler.
func (h *SaramaConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

// Cleanup implements sarama.ConsumerGroupHandler.
func (h *SaramaConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim implements sarama.ConsumerGroupHandler.
func (h *SaramaConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {

	done := session.Context().Done()
	for {
		select {
		case <-done:
			return nil
		case message, more := <-claim.Messages():
			if !more {
				return nil
			}
			err := func() error {
				ctx, cancel := context.WithTimeout(session.Context(), h.ConsumerTimeout)
				defer cancel()
				return h.Consumer.Consume(ctx, session, message)
			}()
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				h.Logger.Log(session.Context(), slog.LevelDebug, err.Error())
			case !errors.Is(err, nil):
				return err
			}
		}
	}
}

// Close implements sarama.ConsumerGroupHandler.
func (h *SaramaConsumerGroupHandler) Close() error { return nil }

// SaramaMessageConsumer processes Kafka messages.
type SaramaMessageConsumer interface {
	// Consume should process a message.
	//
	// Consume is responsible for marking the message consumed, unless the
	// context is canceled, in which case the caller should retry, or mark the
	// message as appropriate.
	Consume(ctx context.Context, session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) error
}

var ErrRetriesLimitExceeded = errors.New("retry limit exceeded")

// NTimesRetryingConsumer enhances a SaramaMessageConsumer with a finite
// number of immediate retries.
//
// The delay between each retry can be controlled via the Delay property. If
// no Delay property is specified, a delay based on the Fibonacci sequence is
// used.
//
// Logger is intentionally minimal. The slog.Log function is used by default.
type NTimesRetryingConsumer struct {
	Times    int
	Consumer SaramaMessageConsumer
	Delay    func(tries int) time.Duration
	Logger   Logger
}

// Logger is an intentionally minimal interface for basic logging.
//
// It matches the signature of slog.Log.
type Logger interface {
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
}

func (c *NTimesRetryingConsumer) Consume(ctx context.Context,
	session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) (err error) {

	var joinedErrors error
	var tries int = 0
	var delay time.Duration = 0
	if c.Delay == nil {
		c.Delay = DelayFibonacci
	}
	if c.Logger == nil {
		c.Logger = slog.Default()
	}
	done := ctx.Done()
	for tries < c.Times {
		select {
		case <-done:
			return nil
		case <-time.After(delay):
			err := c.Consumer.Consume(ctx, session, message)
			if errors.Is(err, nil) || errors.Is(err, context.Canceled) {
				return nil
			}
			delay = c.Delay(tries)
			c.Logger.Log(ctx, slog.LevelInfo, "failure consuming Kafka message, will retry",
				slog.Attr{Key: "tries", Value: slog.IntValue(tries)},
				slog.Attr{Key: "times", Value: slog.IntValue(c.Times)},
				slog.Attr{Key: "delay", Value: slog.DurationValue(delay)},
				slog.Attr{Key: "err", Value: slog.AnyValue(err)},
			)
			joinedErrors = errors.Join(joinedErrors, err)
			tries++
		}
	}

	return errors.Join(joinedErrors, c.retryLimitError())
}

func (c *NTimesRetryingConsumer) retryLimitError() error {
	return fmt.Errorf("%w (%d)", ErrRetriesLimitExceeded, c.Times)
}

// DelayNone is a function returning a constant "no delay" of 0 seconds.
var DelayNone = func(_ int) time.Duration { return DelayConstant(0) }

// DelayConstant is a function returning a constant number of seconds.
func DelayConstant(n int) time.Duration { return time.Duration(n) * time.Second }

// DelayExponentialBinary returns a binary exponential delay.
//
// The delay is 2**tries seconds.
func DelayExponentialBinary(tries int) time.Duration {
	return time.Second * time.Duration(math.Pow(2, float64(tries)))
}

// DelayFibonacci returns a delay based on the Fibonacci sequence.
func DelayFibonacci(tries int) time.Duration {
	return time.Second * time.Duration(Fib(tries))
}

// Fib returns the nth number in the Fibonacci sequence.
func Fib(n int) int {
	if n == 0 {
		return 0
	} else if n < 3 {
		return 1
	}

	n1, n2 := 1, 1
	for i := 3; i <= n; i++ {
		n1, n2 = n1+n2, n1
	}

	return n1
}
