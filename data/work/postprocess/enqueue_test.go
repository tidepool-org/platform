package postprocess_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"go.uber.org/mock/gomock"

	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	userTest "github.com/tidepool-org/platform/user/test"
	"github.com/tidepool-org/platform/work"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("Enqueue", func() {
	var controller *gomock.Controller
	var workClient *workTest.MockClient
	var ctx context.Context
	var userID string
	var id string

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		workClient = workTest.NewMockClient(controller)
		ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		userID = userTest.RandomUserID()
		id = dataWorkPostprocess.IDFromUserID(userID)
	})

	AfterEach(func() {
		controller.Finish()
	})

	expectCreate := func() func() *work.Create {
		var created *work.Create
		workClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, create *work.Create) (*work.Work, error) {
				created = create
				return &work.Work{ID: create.Type}, nil
			})
		return func() *work.Create { return created }
	}

	It("creates work available now", func() {
		created := expectCreate()

		Expect(dataWorkPostprocess.Enqueue(ctx, workClient, userID, dataWorkPostprocess.ReasonDataAdded)).To(Succeed())

		Expect(created().Type).To(Equal(dataWorkPostprocess.Type))
		Expect(created().GroupID).To(PointTo(Equal(id)))
		Expect(created().SerialID).To(PointTo(Equal(id)))
		Expect(created().ProcessingPriority).To(Equal(0))
		Expect(created().ProcessingTimeout).To(Equal(300))
		Expect(created().ProcessingAvailableTime).To(BeTemporally("~", time.Now(), time.Second))
		Expect(created().Metadata).To(HaveKeyWithValue("userId", userID))
		Expect(created().Metadata).To(HaveKeyWithValue("reasons", ConsistOf(dataWorkPostprocess.ReasonDataAdded)))
	})

	// Work is never merged into work already pending, so that reporting a change is a single insert.
	// The deduplication id is deliberately absent as it would instead discard the work reported.
	It("creates work that is neither deduplicated nor merged into work already pending", func() {
		created := expectCreate()

		Expect(dataWorkPostprocess.Enqueue(ctx, workClient, userID, dataWorkPostprocess.ReasonDataAdded)).To(Succeed())

		Expect(created().DeduplicationID).To(BeNil())
	})

	// Work is always available now. The time to defer a change until, which the legacy ingestion
	// service computes from its own batching, is expected to be reported by its caller rather than
	// derived here.
	It("creates work available now whatever the reasons report", func() {
		created := expectCreate()

		Expect(dataWorkPostprocess.Enqueue(ctx, workClient, userID, dataWorkPostprocess.ReasonLegacyDataAdded)).To(Succeed())

		Expect(created().ProcessingAvailableTime).To(BeTemporally("~", time.Now(), time.Second))
	})

	It("creates work reporting every reason once", func() {
		created := expectCreate()

		Expect(dataWorkPostprocess.Enqueue(ctx, workClient, userID, dataWorkPostprocess.ReasonDataAdded, dataWorkPostprocess.ReasonUploadCompleted, dataWorkPostprocess.ReasonDataAdded)).To(Succeed())

		Expect(created().Metadata).To(HaveKeyWithValue("reasons", ConsistOf(dataWorkPostprocess.ReasonDataAdded, dataWorkPostprocess.ReasonUploadCompleted)))
	})

	It("returns an error when the work cannot be created", func() {
		workClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errorsTest.RandomError())

		err := dataWorkPostprocess.Enqueue(ctx, workClient, userID, dataWorkPostprocess.ReasonDataAdded)
		Expect(err).To(MatchError(ContainSubstring("unable to create work")))
	})

	DescribeTable("returns an error when a parameter is missing",
		func(get func() (context.Context, work.Client, string, []string), expected string) {
			enqueueCtx, enqueueWorkClient, enqueueUserID, enqueueReasons := get()
			Expect(dataWorkPostprocess.Enqueue(enqueueCtx, enqueueWorkClient, enqueueUserID, enqueueReasons...)).To(MatchError(expected))
		},
		Entry("context", func() (context.Context, work.Client, string, []string) {
			return nil, workClient, userID, []string{dataWorkPostprocess.ReasonDataAdded}
		}, "context is missing"),
		Entry("work client", func() (context.Context, work.Client, string, []string) {
			return ctx, nil, userID, []string{dataWorkPostprocess.ReasonDataAdded}
		}, "work client is missing"),
		Entry("user id", func() (context.Context, work.Client, string, []string) {
			return ctx, workClient, "", []string{dataWorkPostprocess.ReasonDataAdded}
		}, "user id is missing"),
		Entry("reasons", func() (context.Context, work.Client, string, []string) {
			return ctx, workClient, userID, nil
		}, "reasons is missing"),
	)
})
