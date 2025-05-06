package work_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/work"
	workTest "github.com/tidepool-org/platform/work/test"
	workLoadTest "github.com/tidepool-org/platform/work/test/load"
)

var _ = Describe("Work System", func() {
	var authClient *authTest.Client
	var workClient *workTest.MockClient
	var workController *gomock.Controller
	var ctx context.Context
	var coordinator *workLoadTest.CoordinatorClient
	var err error

	BeforeEach(func() {
		authClient = authTest.NewClient()
		workController = gomock.NewController(GinkgoT())
		workClient = workTest.NewMockClient(workController)
		ctx = context.Background()
		coordinator, err = workLoadTest.NewCoordinatorClient(authClient, workClient)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		authClient.AssertOutputsEmpty()
		workController.Finish()
		for groupID := range coordinator.GetCreatedWork() {
			workClient.EXPECT().DeleteAllByGroupID(ctx, groupID).Times(1)
		}
		err = coordinator.CleanUp(ctx)
		Expect(err).To(BeNil())
	})

	Context("Coordinator", func() {
		It("runs for basic load", func() {
			workClient.EXPECT().Create(ctx, gomock.Any()).Return(&work.Work{}, nil).Times(3)
			Expect(coordinator.Run(ctx, "./test/load/data/basic_load.json")).To(BeNil())
		})
	})

})
