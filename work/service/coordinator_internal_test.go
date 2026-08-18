package service

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/test"
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

		var err error
		coordinator, err = NewCoordinator(logger, workServiceTest.NewMockServerSessionTokenProvider(controller), workClient)
		Expect(err).ToNot(HaveOccurred())
		Expect(coordinator).ToNot(BeNil())

		// Initialized directly as the contexts are otherwise only initialized by Start, which
		// defers the first request for work by CoordinatorDelayInitial
		coordinator.initializeContexts()
	})

	AfterEach(func() {
		controller.Finish()
	})

	Context("requestAndDispatchWork", func() {
		Context("with a registered processor factory", func() {
			BeforeEach(func() {
				processorFactory := workTest.NewMockProcessorFactory(controller)
				processorFactory.EXPECT().Type().Return(netTest.RandomReverseDomain()).AnyTimes()
				processorFactory.EXPECT().Quantity().Return(test.RandomIntFromRange(1, 4)).AnyTimes()
				processorFactory.EXPECT().Frequency().Return(CoordinatorFrequencyDefault).AnyTimes()
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
})
