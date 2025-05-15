package work_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	work_load "github.com/tidepool-org/platform/work/load"
	workService "github.com/tidepool-org/platform/work/service"
	workServiceTest "github.com/tidepool-org/platform/work/service/test"
	workTestLoad "github.com/tidepool-org/platform/work/test/load"
)

var _ = Describe("Work System", func() {
	var authClient *authTest.Client
	var workClient work.Client
	var workStoreController *gomock.Controller
	var workStore *workServiceTest.MockStore
	var ctx context.Context
	var coordinator *workTestLoad.CoordinatorClient
	var err error

	BeforeEach(func() {
		authClient = authTest.NewClient()
		workStoreController = gomock.NewController(GinkgoT())
		workStore = workServiceTest.NewMockStore(workStoreController)
		workClient, err = workService.NewClient(workStore)
		ctx = context.Background()
		coordinator, err = workTestLoad.NewCoordinatorClient(authClient, workClient)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		authClient.AssertOutputsEmpty()
		workStoreController.Finish()
		workStore.EXPECT().DeleteAllByGroupID(ctx, coordinator.GetWorkGroupID()).Times(1)
		err = coordinator.CleanUp(ctx)
		Expect(err).To(BeNil())
	})

	Context("Coordinator", func() {
		It("runs for basic", func() {
			Expect(coordinator.Initialize(ctx, nil, nil, &work.Create{Type: work_load.TypeDopey, ProcessingTimeout: 5})).To(BeNil())
			for _, create := range coordinator.GetCreate() {
				workStore.EXPECT().Create(ctx, create).Return(&work.Work{}, nil).Times(1)
			}
			Expect(coordinator.Run(ctx)).To(BeNil())
		})
		It("runs for duplicates", func() {
			Expect(coordinator.Initialize(ctx, pointer.FromString("duplicate-id"), nil, &work.Create{Type: work_load.TypeDopey, ProcessingTimeout: 5}, &work.Create{Type: work_load.TypeDopey, ProcessingTimeout: 5})).To(BeNil())
			created := coordinator.GetCreate()
			Expect(len(created)).To(Equal(2))
			first := created[0]
			workStore.EXPECT().Create(ctx, first).Return(&work.Work{}, nil).Times(1)
			second := created[1]
			workStore.EXPECT().Create(ctx, second).Return(nil, nil).Times(1)
			Expect(coordinator.Run(ctx)).To(BeNil())
		})
		It("runs for serialize", func() {
			Expect(coordinator.Initialize(ctx, nil, pointer.FromString("serial-id"), &work.Create{Type: work_load.TypeSleepy, ProcessingTimeout: 5}, &work.Create{Type: work_load.TypeSleepy, ProcessingTimeout: 5})).To(BeNil())
			for _, create := range coordinator.GetCreate() {
				workStore.EXPECT().Create(ctx, create).Return(&work.Work{}, nil).Times(1)
			}
			Expect(coordinator.Run(ctx)).To(BeNil())
		})
	})

})
