package service

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	configTest "github.com/tidepool-org/platform/config/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
	workServiceTest "github.com/tidepool-org/platform/work/service/test"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("Coordinator", func() {
	var controller *gomock.Controller
	var logger *logTest.Logger
	var workClient *workServiceTest.MockWorkClient
	var coordinator *Coordinator

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		logger = logTest.NewLogger()
		workClient = workServiceTest.NewMockWorkClient(controller)
		workClient.EXPECT().QueueSizes(gomock.Any()).Return(nil, nil).AnyTimes()

		var err error
		coordinator, err = NewCoordinator(NewCoordinatorConfig(), logger, workServiceTest.NewMockServerSessionTokenProvider(controller), workClient)
		Expect(err).ToNot(HaveOccurred())
		Expect(coordinator).ToNot(BeNil())

		// Initialized directly as the contexts are otherwise only initialized by Start, which
		// defers the first request for work by CoordinatorDelayInitial
		coordinator.initializeContexts()
	})

	AfterEach(func() {
		controller.Finish()
	})

	It("CoordinatorFrequencyDefault is expected", func() {
		Expect(CoordinatorFrequencyDefault).To(Equal(5 * time.Second))
	})

	Context("CoordinatorConfig", func() {
		var cfg *CoordinatorConfig

		BeforeEach(func() {
			cfg = NewCoordinatorConfig()
		})

		It("NewCoordinatorConfig returns the defaults", func() {
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Frequency).To(Equal(CoordinatorFrequencyDefault))
		})

		Context("Load", func() {
			var configReporter *configTest.Reporter

			BeforeEach(func() {
				configReporter = configTest.NewReporter()
			})

			It("returns an error when the config reporter is missing", func() {
				Expect(cfg.Load(nil)).To(MatchError("config reporter is missing"))
			})

			It("returns an error when frequency is not parsable", func() {
				configReporter.Config["frequency"] = test.RandomStringFromCharset(test.CharsetAlpha)
				Expect(cfg.Load(configReporter)).To(MatchError("frequency is invalid"))
			})

			It("returns an error when frequency is not positive", func() {
				configReporter.Config["frequency"] = "0s"
				Expect(cfg.Load(configReporter)).To(MatchError("frequency is invalid"))
			})

			It("keeps the default when frequency is not set", func() {
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Frequency).To(Equal(CoordinatorFrequencyDefault))
			})

			It("loads frequency", func() {
				configReporter.Config["frequency"] = "2.5"
				Expect(cfg.Load(configReporter)).To(Succeed())
				Expect(cfg.Frequency).To(Equal(2500 * time.Millisecond))
			})
		})

		Context("Validate", func() {
			It("succeeds with the defaults", func() {
				Expect(cfg.Validate()).To(Succeed())
			})

			It("returns an error when frequency is not positive", func() {
				cfg.Frequency = 0
				Expect(cfg.Validate()).To(MatchError("frequency is invalid"))
			})
		})
	})

	Context("NewCoordinator", func() {
		It("returns an error when the config is missing", func() {
			coordinator, err := NewCoordinator(nil, logger, workServiceTest.NewMockServerSessionTokenProvider(controller), workClient)
			Expect(err).To(MatchError("config is missing"))
			Expect(coordinator).To(BeNil())
		})

		It("returns an error when the config is invalid", func() {
			cfg := NewCoordinatorConfig()
			cfg.Frequency = 0
			coordinator, err := NewCoordinator(cfg, logger, workServiceTest.NewMockServerSessionTokenProvider(controller), workClient)
			Expect(err).To(MatchError("config is invalid; frequency is invalid"))
			Expect(coordinator).To(BeNil())
		})

		It("uses the configured frequency", func() {
			cfg := NewCoordinatorConfig()
			cfg.Frequency = test.RandomDurationFromRange(time.Second, time.Minute)
			coordinator, err := NewCoordinator(cfg, logger, workServiceTest.NewMockServerSessionTokenProvider(controller), workClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(coordinator.frequency).To(Equal(cfg.Frequency))
		})
	})

	Context("requestAndDispatchWork", func() {
		Context("with a registered processor factory", func() {
			BeforeEach(func() {
				processorFactory := workTest.NewMockProcessorFactory(controller)
				processorFactory.EXPECT().Type().Return(netTest.RandomReverseDomain()).AnyTimes()
				processorFactory.EXPECT().Quantity().Return(test.RandomIntFromRange(1, 4)).AnyTimes()
				Expect(coordinator.RegisterProcessorFactory(processorFactory)).To(Succeed())
			})

			It("reaps expired processing work before polling for work", func() {
				gomock.InOrder(
					workClient.EXPECT().ReapExpiredProcessing(gomock.Any()).Return(0, nil),
					workClient.EXPECT().Poll(gomock.Any(), gomock.Any()).Return(nil, nil),
				)
				coordinator.requestAndDispatchWork()
			})

			// Work that was reaped is already delayed, so a failure to reap must not additionally
			// prevent work that is available from being polled
			It("polls for work even when reaping expired processing work fails", func() {
				err := errorsTest.RandomError()
				gomock.InOrder(
					workClient.EXPECT().ReapExpiredProcessing(gomock.Any()).Return(0, err),
					workClient.EXPECT().Poll(gomock.Any(), gomock.Any()).Return(nil, nil),
				)
				coordinator.requestAndDispatchWork()
				logger.AssertError("unable to reap expired processing work")

				// The coordinator is the single log site for the failure
				count := 0
				for _, fields := range logger.SerializedFields {
					if fields["message"] == "unable to reap expired processing work" {
						count++
					}
				}
				Expect(count).To(Equal(1))
			})

			It("does not reap expired processing work again within the reap interval", func() {
				workClient.EXPECT().ReapExpiredProcessing(gomock.Any()).Return(0, nil)
				workClient.EXPECT().Poll(gomock.Any(), gomock.Any()).Return(nil, nil).Times(2)
				coordinator.requestAndDispatchWork()
				coordinator.requestAndDispatchWork()
			})

			It("reaps expired processing work again once the reap interval passes", func() {
				now := time.Now()
				coordinator.NowFunc = func() time.Time { return now }
				workClient.EXPECT().ReapExpiredProcessing(gomock.Any()).Return(0, nil).Times(2)
				workClient.EXPECT().Poll(gomock.Any(), gomock.Any()).Return(nil, nil).Times(2)
				coordinator.requestAndDispatchWork()
				now = now.Add(ReapExpiredProcessingInterval)
				coordinator.requestAndDispatchWork()
			})

			It("reports the count when expired processing work is reaped", func() {
				count := test.RandomIntFromRange(1, 10)
				workClient.EXPECT().ReapExpiredProcessing(gomock.Any()).Return(count, nil)
				workClient.EXPECT().Poll(gomock.Any(), gomock.Any()).Return(nil, nil)
				coordinator.requestAndDispatchWork()
				logger.AssertWarn("reaped expired processing work", log.Fields{"count": count})
			})
		})

		// Reaping is not specific to any processor type, so it must not be prevented by the absence
		// of any processor with capacity, which stops work from being polled
		It("reaps expired processing work when no processor factory is registered", func() {
			workClient.EXPECT().ReapExpiredProcessing(gomock.Any()).Return(0, nil)
			coordinator.requestAndDispatchWork()
		})

		It("does not reap expired processing work before the coordinator is started", func() {
			coordinator.workersContext = nil
			coordinator.requestAndDispatchWork()
		})
	})

	Context("startManager", func() {
		BeforeEach(func() {
			coordinator.workersCompletionChannel = make(chan *coordinatorProcessingCompletion, 1)
			coordinator.startManager()
		})

		AfterEach(func() {
			coordinator.Stop()
		})

		// Work is requested only when the timer fires, so that the poll rate is bounded by the
		// frequency rather than by the completion rate
		It("completes work without requesting more work", func() {
			completion := randomCompletion()
			updated := make(chan struct{})
			workClient.EXPECT().Update(gomock.Any(), completion.Identifier.ID, gomock.Any(), gomock.Any()).
				DoAndReturn(func(context.Context, string, *request.Condition, *work.Update) (*work.Work, error) {
					close(updated)
					return workTest.RandomWork(), nil
				})
			coordinator.workersCompletionChannel <- completion
			Eventually(updated).Should(BeClosed())
		})
	})

	Context("completeAllWork", func() {
		BeforeEach(func() {
			coordinator.workersCompletionChannel = make(chan *coordinatorProcessingCompletion, 2)
		})

		It("completes all queued work without requesting more work", func() {
			for range 2 {
				completion := randomCompletion()
				workClient.EXPECT().Update(gomock.Any(), completion.Identifier.ID, gomock.Any(), gomock.Any()).Return(workTest.RandomWork(), nil)
				coordinator.workersCompletionChannel <- completion
			}
			coordinator.completeAllWork()
			Expect(coordinator.workersCompletionChannel).To(BeEmpty())
		})

		It("returns immediately when no work is queued", func() {
			coordinator.completeAllWork()
		})
	})
})

func randomCompletion() *coordinatorProcessingCompletion {
	return &coordinatorProcessingCompletion{
		Identifier: &coordinatorProcessingIdentifier{
			ID:       workTest.RandomID(),
			Type:     workTest.RandomType(),
			Revision: test.RandomIntFromRange(0, 10),
		},
		ProcessResult: workTest.RandomSuccessProcessResult(),
	}
}
