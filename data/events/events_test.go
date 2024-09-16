package events

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/devlog"
	lognull "github.com/tidepool-org/platform/log/null"
	logtest "github.com/tidepool-org/platform/log/test"
)

var _ = Describe("SaramaRunner", func() {
	Context("has a lifecycle", func() {
		newTestRunner := func() *SaramaRunner {
			return NewSaramaRunner(&mockEventsRunner{})
		}
		It("starts with Run() and stops with Terminate()", func() {
			r := newTestRunner()
			var runErr error
			var errMu sync.Mutex
			launched := make(chan struct{}, 1)
			go func() {
				errMu.Lock()
				func() {
					defer errMu.Unlock()
					runErr = r.Run()
					launched <- struct{}{}
				}()
			}()
			<-launched
			time.Sleep(time.Millisecond)
			errMu.Lock()
			defer errMu.Unlock()

			Expect(r.Terminate()).To(Succeed())
			Eventually(runErr).WithTimeout(10 * time.Millisecond).Should(Succeed())
		})

		Describe("Run()", func() {
			var errMu sync.Mutex

			It("can be started only once", func() {
				r := newTestRunner()
				var firstRunErr, secondRunErr error
				launched := make(chan struct{}, 2)
				go func() {
					errMu.Lock()
					func() {
						defer errMu.Unlock()
						firstRunErr = r.Run()
						launched <- struct{}{}
					}()
				}()
				go func() {
					errMu.Lock()
					func() {
						defer errMu.Unlock()
						secondRunErr = r.Run()
						launched <- struct{}{}
					}()

				}()
				<-launched
				<-launched
				errMu.Lock()
				defer errMu.Unlock()

				// The above doesn't _guarantee_ that Run has been called twice,
				// but... it should work.

				Expect(r.Terminate()).To(Succeed())
				if firstRunErr != nil {
					Expect(firstRunErr).To(MatchError(ContainSubstring("it's already initialized")))
					Expect(secondRunErr).To(Succeed())
				} else {
					Expect(firstRunErr).To(Succeed())
					Expect(secondRunErr).To(MatchError(ContainSubstring("it's already initialized")))
				}
			})

			It("can't be Terminate()'d before it's Run()", func() {
				r := newTestRunner()
				Expect(r.Terminate()).To(MatchError(ContainSubstring("it's not running")))
			})
		})
	})
})

var _ = DescribeTable("CappedExponentialBinaryDelay",
	func(cap time.Duration, input int, output time.Duration) {
		f := CappedExponentialBinaryDelay(cap)
		Expect(f(input)).To(Equal(output))
	},
	Entry("cap: 1m; tries: 0", time.Minute, 0, time.Second),
	Entry("cap: 1m; tries: 1", time.Minute, 1, 2*time.Second),
	Entry("cap: 1m; tries: 2", time.Minute, 2, 4*time.Second),
	Entry("cap: 1m; tries: 3", time.Minute, 3, 8*time.Second),
	Entry("cap: 1m; tries: 4", time.Minute, 4, 16*time.Second),
	Entry("cap: 1m; tries: 5", time.Minute, 5, 32*time.Second),
	Entry("cap: 1m; tries: 6", time.Minute, 6, time.Minute),
	Entry("cap: 1m; tries: 20", time.Minute, 20, time.Minute),
)

var _ = Describe("NotBeforeConsumer", func() {
	Describe("Consume", func() {
		var newTestMsg = func(notBefore time.Time) *sarama.ConsumerMessage {
			headers := []*sarama.RecordHeader{}
			if !notBefore.IsZero() {
				headers = append(headers, &sarama.RecordHeader{
					Key:   HeaderNotBefore,
					Value: []byte(notBefore.Format(NotBeforeTimeFormat)),
				})
			}
			return &sarama.ConsumerMessage{Topic: "test.topic", Headers: headers}
		}

		It("delays based on the x-tidepool-not-before header", func() {
			logger := newTestDevlog()
			testDelay := 10 * time.Millisecond
			ctx := context.Background()
			start := time.Now()
			notBefore := start.Add(testDelay)
			msg := newTestMsg(notBefore)
			dc := &NotBeforeConsumer{
				Consumer: &mockSaramaMessageConsumer{Logger: logger},
				Logger:   logger,
			}

			err := dc.Consume(ctx, nil, msg)

			Expect(err).To(BeNil())
			Expect(time.Since(start)).To(BeNumerically(">", testDelay))
		})

		It("aborts if canceled", func() {
			logger := newTestDevlog()
			testDelay := 10 * time.Millisecond
			abortAfter := 1 * time.Millisecond
			notBefore := time.Now().Add(testDelay)
			msg := newTestMsg(notBefore)
			dc := &NotBeforeConsumer{
				Consumer: &mockSaramaMessageConsumer{Delay: time.Minute, Logger: logger},
				Logger:   logger,
			}
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				defer cancel()
				<-time.After(abortAfter)
			}()
			start := time.Now()

			err := dc.Consume(ctx, nil, msg)

			Expect(err).To(BeNil())
			Expect(time.Since(start)).To(BeNumerically(">", abortAfter))
		})

	})
})

var _ = Describe("CascadingConsumer", func() {
	Describe("Consume", func() {
		var testMsg = &sarama.ConsumerMessage{
			Topic: "test.topic",
		}

		Context("on failure", func() {
			It("cascades topics", func() {
				t := GinkgoT()
				logger := newTestDevlog()
				ctx := context.Background()
				testConfig := mocks.NewTestConfig()
				mockProducer := mocks.NewAsyncProducer(t, testConfig)
				msg := &sarama.ConsumerMessage{}
				nextTopic := "text-next"
				sc := &CascadingConsumer{
					Consumer: &mockSaramaMessageConsumer{
						Err:    fmt.Errorf("test error"),
						Logger: logger,
					},
					NextTopic: nextTopic,
					Producer:  mockProducer,
					Logger:    logger,
				}

				cf := func(msg *sarama.ProducerMessage) error {
					if msg.Topic != nextTopic {
						return fmt.Errorf("expected topic to be %q, got %q", nextTopic, msg.Topic)
					}
					return nil
				}
				mockProducer.ExpectInputWithMessageCheckerFunctionAndSucceed(cf)

				err := sc.Consume(ctx, nil, msg)
				Expect(mockProducer.Close()).To(Succeed())
				Expect(err).To(BeNil())
			})

			It("increments the failures header", func() {
				t := GinkgoT()
				logger := newTestDevlog()
				ctx := context.Background()
				testConfig := mocks.NewTestConfig()
				mockProducer := mocks.NewAsyncProducer(t, testConfig)
				msg := &sarama.ConsumerMessage{
					Headers: []*sarama.RecordHeader{
						{
							Key: HeaderFailures, Value: []byte("3"),
						},
					},
				}
				nextTopic := "text-next"
				sc := &CascadingConsumer{
					Consumer: &mockSaramaMessageConsumer{
						Err:    fmt.Errorf("test error"),
						Logger: logger,
					},
					NextTopic: nextTopic,
					Producer:  mockProducer,
					Logger:    logger,
				}

				cf := func(msg *sarama.ProducerMessage) error {
					failures := 0
					for _, header := range msg.Headers {
						if !bytes.Equal(header.Key, HeaderFailures) {
							continue
						}
						parsed, err := strconv.ParseInt(string(header.Value), 10, 32)
						Expect(err).To(Succeed())
						failures = int(parsed)
						if failures != 4 {
							return fmt.Errorf("expected failures == 4, got %d", failures)
						}
						return nil
					}
					return fmt.Errorf("expected failures header wasn't found")
				}
				mockProducer.ExpectInputWithMessageCheckerFunctionAndSucceed(cf)

				err := sc.Consume(ctx, nil, msg)
				Expect(mockProducer.Close()).To(Succeed())
				Expect(err).To(BeNil())
			})

			It("updates the not before header", func() {
				t := GinkgoT()
				logger := newTestDevlog()
				ctx := context.Background()
				testConfig := mocks.NewTestConfig()
				mockProducer := mocks.NewAsyncProducer(t, testConfig)
				msg := &sarama.ConsumerMessage{
					Headers: []*sarama.RecordHeader{
						{
							Key: HeaderFailures, Value: []byte("2"),
						},
					},
				}
				nextTopic := "text-next"
				sc := &CascadingConsumer{
					Consumer: &mockSaramaMessageConsumer{
						Err:    fmt.Errorf("test error"),
						Logger: logger,
					},
					NextTopic: nextTopic,
					Producer:  mockProducer,
					Logger:    logger,
				}

				cf := func(msg *sarama.ProducerMessage) error {
					for _, header := range msg.Headers {
						if !bytes.Equal(header.Key, HeaderNotBefore) {
							continue
						}
						parsed, err := time.Parse(NotBeforeTimeFormat, string(header.Value))
						if err != nil {
							return err
						}
						until := time.Until(parsed)
						delta := 10 * time.Millisecond
						if until < 2*time.Second-delta || until > 2*time.Second+delta {
							return fmt.Errorf("expected 2 seconds' delay, got: %s", until)
						}
						return nil
					}
					return fmt.Errorf("expected failures header wasn't found")
				}
				mockProducer.ExpectInputWithMessageCheckerFunctionAndSucceed(cf)

				err := sc.Consume(ctx, nil, msg)
				Expect(mockProducer.Close()).To(Succeed())
				Expect(err).To(BeNil())
			})
		})

		Context("on success", func() {
			It("doesn't produce a new message", func() {
				t := GinkgoT()
				logger := newTestDevlog()
				ctx := context.Background()
				testConfig := mocks.NewTestConfig()
				mockProducer := mocks.NewAsyncProducer(t, testConfig)
				msg := &sarama.ConsumerMessage{}
				nextTopic := "text-next"
				sc := &CascadingConsumer{
					Consumer:  &mockSaramaMessageConsumer{Logger: logger},
					NextTopic: nextTopic,
					Producer:  mockProducer,
					Logger:    logger,
				}

				err := sc.Consume(ctx, nil, msg)
				Expect(mockProducer.Close()).To(Succeed())
				Expect(err).To(BeNil())
			})
		})

		Context("when canceled", func() {
			It("aborts", func() {
				logger := newTestDevlog()
				abortAfter := 1 * time.Millisecond
				p := newMockSaramaAsyncProducer(nil)
				sc := &CascadingConsumer{
					Consumer: &mockSaramaMessageConsumer{Delay: time.Minute, Logger: logger},
					Logger:   lognull.NewLogger(),
					Producer: p,
				}
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					defer cancel()
					time.Sleep(abortAfter)
				}()
				start := time.Now()

				err := sc.Consume(ctx, nil, testMsg)
				Expect(err).To(BeNil())
				Expect(time.Since(start)).To(BeNumerically(">", abortAfter))
			})
		})
	})
})

var _ = Describe("CascadingSaramaEventsRunner", func() {
	It("cascades through configured delays", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		testDelays := []time.Duration{0, 1, 2, 3, 5}
		testLogger := newTestDevlog()
		testMessageConsumer := &mockSaramaMessageConsumer{
			Delay:  time.Millisecond,
			Err:    fmt.Errorf("test error"),
			Logger: testLogger,
		}
		testConfig := SaramaRunnerConfig{
			Topics:          []string{"test.cascading"},
			MessageConsumer: testMessageConsumer,
			Sarama:          mocks.NewTestConfig(),
		}
		producers := []*mockSaramaAsyncProducer{}
		var msgsReceived atomic.Int32
		prodFunc := func(_ []string, config *sarama.Config) (sarama.AsyncProducer, error) {
			prod := newMockSaramaAsyncProducer(func(msg *sarama.ProducerMessage) {
				msgsReceived.Add(1)
				if int(msgsReceived.Load()) == len(testDelays) {
					// Once all messages are entered, the test is complete. Cancel the
					// context to shut it all down properly.
					cancel()
				}
			})
			producers = append(producers, prod)
			return prod, nil
		}
		sser := NewCascadingSaramaEventsRunner(testConfig, testLogger, testDelays)
		sser.SaramaBuilders = newTestSaramaBuilders(nil, prodFunc)

		err := sser.Run(ctx)
		Expect(err).To(Succeed())
		for pIdx, p := range producers {
			Expect(p.isClosed()).To(BeTrue())
			Expect(p.messages).To(HaveLen(1))
			topic := p.messages[0].Topic
			switch {
			case pIdx+1 < len(testDelays):
				Expect(topic).To(MatchRegexp(fmt.Sprintf(".*-retry-%s$", testDelays[pIdx+1])))
			default:
				Expect(topic).To(MatchRegexp(".*-dead$"))
			}
		}
	})

	Describe("logger", func() {
		It("prefers a context's logger", func() {
			testLogger := logtest.NewLogger()
			ctxLogger := logtest.NewLogger()
			testDelays := []time.Duration{0}
			testConfig := SaramaRunnerConfig{}
			r := NewCascadingSaramaEventsRunner(testConfig, testLogger, testDelays)

			ctx := log.NewContextWithLogger(context.Background(), ctxLogger)
			got := r.logger(ctx)

			Expect(got).To(Equal(ctxLogger))
		})

		Context("without a context logger", func() {
			It("uses the configured logger", func() {
				testLogger := logtest.NewLogger()
				testDelays := []time.Duration{0}
				testConfig := SaramaRunnerConfig{}
				r := NewCascadingSaramaEventsRunner(testConfig, testLogger, testDelays)

				ctx := context.Background()
				got := r.logger(ctx)

				Expect(got).To(Equal(testLogger))
			})

			Context("or any configured logger", func() {
				It("doesn't panic", func() {
					testLogger := logtest.NewLogger()
					testDelays := []time.Duration{0}
					testConfig := SaramaRunnerConfig{}
					r := NewCascadingSaramaEventsRunner(testConfig, testLogger, testDelays)

					ctx := context.Background()
					got := r.logger(ctx)

					Expect(func() {
						got.Debug("testing")
					}).ToNot(Panic())
				})
			})
		})
	})
})

// testSaramaBuilders injects mocks into the CascadingSaramaEventsRunner
type testSaramaBuilders struct {
	consumerGroup func([]string, string, *sarama.Config) (sarama.ConsumerGroup, error)
	producer      func([]string, *sarama.Config) (sarama.AsyncProducer, error)
}

func newTestSaramaBuilders(
	cgFunc func([]string, string, *sarama.Config) (sarama.ConsumerGroup, error),
	prodFunc func([]string, *sarama.Config) (sarama.AsyncProducer, error)) *testSaramaBuilders {

	if cgFunc == nil {
		cgFunc = func(_ []string, groupID string, config *sarama.Config) (sarama.ConsumerGroup, error) {
			logger := newTestDevlog()
			return &mockSaramaConsumerGroup{
				Logger: logger,
			}, nil
		}
	}
	if prodFunc == nil {
		prodFunc = func(_ []string, config *sarama.Config) (sarama.AsyncProducer, error) {
			return mocks.NewAsyncProducer(GinkgoT(), config), nil
		}
	}
	return &testSaramaBuilders{
		consumerGroup: cgFunc,
		producer:      prodFunc,
	}
}

func (b testSaramaBuilders) NewAsyncProducer(brokers []string, config *sarama.Config) (
	sarama.AsyncProducer, error) {

	return b.producer(brokers, config)
}

func (b testSaramaBuilders) NewConsumerGroup(brokers []string, groupID string,
	config *sarama.Config) (sarama.ConsumerGroup, error) {

	return b.consumerGroup(brokers, groupID, config)
}

type mockEventsRunner struct {
	Err error
}

func (r *mockEventsRunner) Run(ctx context.Context) error {
	return r.Err
}

type mockSaramaMessageConsumer struct {
	Delay  time.Duration
	Err    error
	Logger log.Logger
}

func (c *mockSaramaMessageConsumer) Consume(ctx context.Context,
	session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) (err error) {

	c.Logger.Debugf("mockSaramaMessageConsumer[%q] is consuming %+v", msg.Topic, msg)
	defer func(err *error) {
		c.Logger.Debugf("mockSaramaMessageConsumer[%q] returns %s", msg.Topic, *err)
	}(&err)

	done := ctx.Done()
	select {
	case <-time.After(c.Delay):
		// no op
	case <-done:
		return ctx.Err()
	}

	if c.Err != nil {
		return c.Err
	}
	return nil
}

type mockSaramaConsumerGroup struct {
	Messages   chan *sarama.ConsumerMessage
	ConsumeErr error
	Logger     log.Logger
}

func (g *mockSaramaConsumerGroup) Consume(ctx context.Context,
	topics []string, handler sarama.ConsumerGroupHandler) error {

	if g.ConsumeErr != nil {
		return g.ConsumeErr
	}

	g.Logger.Debugf("mockSaramaConsumerGroup%s consuming", topics)
	session := &mockSaramaConsumerGroupSession{}
	if g.Messages == nil {
		g.Messages = make(chan *sarama.ConsumerMessage)
		go func() { <-ctx.Done(); close(g.Messages) }()
		go g.feedYourClaim(ctx, topics[0])
	}
	claim := &mockSaramaConsumerGroupClaim{
		topic:    topics[0],
		messages: g.Messages,
	}

	err := handler.ConsumeClaim(session, claim)
	if err != nil {
		return err
	}
	return nil
}

func (g *mockSaramaConsumerGroup) feedYourClaim(ctx context.Context, topic string) {
	msg := &sarama.ConsumerMessage{Topic: topic}
	select {
	case <-ctx.Done():
		return
	case g.Messages <- msg:
		// no op
	}
}

func (g *mockSaramaConsumerGroup) Errors() <-chan error {
	panic("not implemented") // implement if needed
}

func (g *mockSaramaConsumerGroup) Close() error {
	panic("not implemented") // implement if needed
}

func (g *mockSaramaConsumerGroup) Pause(partitions map[string][]int32) {
	panic("not implemented") // implement if needed
}

func (g *mockSaramaConsumerGroup) Resume(partitions map[string][]int32) {
	panic("not implemented") // implement if needed
}

func (g *mockSaramaConsumerGroup) PauseAll() {
	panic("not implemented") // implement if needed
}

func (g *mockSaramaConsumerGroup) ResumeAll() {
	panic("not implemented") // implement if needed}
}

type mockSaramaConsumerGroupSession struct{}

func (s *mockSaramaConsumerGroupSession) Claims() map[string][]int32 {
	panic("not implemented") // implement if needed
}

func (s *mockSaramaConsumerGroupSession) MemberID() string {
	panic("not implemented") // implement if needed
}

func (s *mockSaramaConsumerGroupSession) GenerationID() int32 {
	panic("not implemented") // implement if needed
}

func (s *mockSaramaConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
	panic("not implemented") // implement if needed
}

func (s *mockSaramaConsumerGroupSession) Commit() {
	panic("not implemented") // implement if needed
}

func (s *mockSaramaConsumerGroupSession) ResetOffset(topic string, partition int32, offset int64, metadata string) {
	panic("not implemented") // implement if needed
}

func (s *mockSaramaConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {
	panic("not implemented") // implement if needed
}

func (s *mockSaramaConsumerGroupSession) Context() context.Context {
	panic("not implemented") // implement if needed
}

type mockSaramaConsumerGroupClaim struct {
	messages <-chan *sarama.ConsumerMessage
	topic    string
}

func (c *mockSaramaConsumerGroupClaim) Topic() string {
	return c.topic
}

func (c *mockSaramaConsumerGroupClaim) Partition() int32 {
	panic("not implemented") // implement if needed
}

func (c *mockSaramaConsumerGroupClaim) InitialOffset() int64 {
	panic("not implemented") // implement if needed
}

func (c *mockSaramaConsumerGroupClaim) HighWaterMarkOffset() int64 {
	panic("not implemented") // implement if needed
}

func (c *mockSaramaConsumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage {
	return c.messages
}

type mockSaramaAsyncProducer struct {
	input              chan *sarama.ProducerMessage
	messages           []*sarama.ProducerMessage
	mu                 sync.Mutex
	setupCallbacksOnce sync.Once
	closeOnce          sync.Once
	msgCallback        func(*sarama.ProducerMessage)
}

func newMockSaramaAsyncProducer(msgCallback func(*sarama.ProducerMessage)) *mockSaramaAsyncProducer {
	return &mockSaramaAsyncProducer{
		input:       make(chan *sarama.ProducerMessage),
		messages:    []*sarama.ProducerMessage{},
		msgCallback: msgCallback,
	}
}

func (p *mockSaramaAsyncProducer) AsyncClose() {
	panic("not implemented") // implement if needed
}

func (p *mockSaramaAsyncProducer) Close() error {
	p.closeOnce.Do(func() { close(p.input) })
	return nil
}

func (p *mockSaramaAsyncProducer) setupCallbacks() {
	if p.msgCallback == nil {
		return
	}
	p.setupCallbacksOnce.Do(func() {
		go func(callback func(*sarama.ProducerMessage)) {
			for msg := range p.input {
				p.messages = append(p.messages, msg)
				go callback(msg)
			}
		}(p.msgCallback)
	})
}

func (p *mockSaramaAsyncProducer) Input() chan<- *sarama.ProducerMessage {
	defer p.setupCallbacks()
	return p.input
}

func (p *mockSaramaAsyncProducer) Successes() <-chan *sarama.ProducerMessage {
	panic("not implemented") // implement if needed
}

func (p *mockSaramaAsyncProducer) Errors() <-chan *sarama.ProducerError {
	panic("not implemented") // implement if needed
}

func (p *mockSaramaAsyncProducer) IsTransactional() bool {
	panic("not implemented") // implement if needed
}

func (p *mockSaramaAsyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	panic("not implemented") // implement if needed
}

func (p *mockSaramaAsyncProducer) BeginTxn() error {
	return nil
}

func (p *mockSaramaAsyncProducer) CommitTxn() error {
	return nil
}

func (p *mockSaramaAsyncProducer) AbortTxn() error {
	return nil
}

func (p *mockSaramaAsyncProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	panic("not implemented") // implement if needed
}

func (p *mockSaramaAsyncProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	panic("not implemented") // implement if needed
}

func (p *mockSaramaAsyncProducer) isClosed() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	select {
	case _, open := <-p.input:
		return !open
	default:
		return false
	}
}

func newTestDevlog() log.Logger {
	GinkgoHelper()
	l, err := devlog.NewWithDefaults(GinkgoWriter)
	Expect(err).To(Succeed())
	return l
}
