package events

import (
	"context"
	"log/slog"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/test"
)

var _ = Describe("SaramaRunner", func() {
	Context("has a lifecycle", func() {
		newTestRunner := func() *SaramaRunner {
			return &SaramaRunner{
				Config:       SaramaRunnerConfig{},
				EventsRunner: &mockEventsRunner{},
			}
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

	Describe("logger", func() {
		It("prefers a context's logger", func() {
			testLogger := test.NewLogger()
			ctxLogger := test.NewLogger()
			r := &SaramaRunner{
				Config: SaramaRunnerConfig{Logger: testLogger},
			}

			ctx := log.NewContextWithLogger(context.Background(), ctxLogger)
			got := r.logger(ctx)

			goCommonLogger, ok := got.(*log.GoCommonAdapter)
			Expect(ok).To(BeTrue())
			Expect(goCommonLogger.Logger).To(Equal(ctxLogger))
		})

		Context("without a context logger", func() {
			It("uses the configured logger", func() {
				testLogger := test.NewLogger()
				r := &SaramaRunner{
					Config: SaramaRunnerConfig{
						Logger: testLogger,
					},
				}

				got := r.logger(context.Background())

				goCommonLogger, ok := got.(*log.GoCommonAdapter)
				Expect(ok).To(BeTrue())
				Expect(goCommonLogger.Logger).To(Equal(testLogger))
			})

			Context("or any configured logger", func() {
				It("doesn't panic", func() {
					r := &SaramaRunner{Config: SaramaRunnerConfig{}}
					ctx := context.Background()
					got := r.logger(ctx)

					Expect(func() {
						got.Log(ctx, slog.LevelInfo, "testing")
					}).ToNot(Panic())
				})
			})
		})
	})

	DescribeTable("CappedExponentialBinaryDelay",
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
})

type mockEventsRunner struct {
	Err error
}

func (r *mockEventsRunner) Run(ctx context.Context) error {
	return r.Err
}
