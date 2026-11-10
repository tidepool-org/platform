package v1_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"
	"go.uber.org/mock/gomock"

	v1 "github.com/tidepool-org/platform/data/service/api/v1"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	testRest "github.com/tidepool-org/platform/test/rest"
	userTest "github.com/tidepool-org/platform/user/test"
	"github.com/tidepool-org/platform/work"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("SyncSummaries", func() {
	var controller *gomock.Controller
	var workClient *workTest.MockClient
	var res *testRest.ResponseWriter
	var svcCtx *mockDataServiceContext
	var userID string

	statusCode := func() int {
		Expect(res.WriteHeaderInputs).To(HaveLen(1))
		return res.WriteHeaderInputs[0]
	}

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		workClient = workTest.NewMockClient(controller)
		userID = userTest.RandomUserID()
		res = testRest.NewResponseWriter()
		res.HeaderOutput = &http.Header{}
		res.WriteStub = func(bites []byte) (int, error) { return len(bites), nil }
		req := httptest.NewRequest(http.MethodPost, "/v1/summaries/"+userID+"/sync", nil)
		req = req.WithContext(log.NewContextWithLogger(req.Context(), logTest.NewLogger()))
		svcCtx = &mockDataServiceContext{
			RestRequest:    &rest.Request{Request: req, PathParams: map[string]string{"userId": userID}, Env: map[string]interface{}{}},
			RestResponse:   res,
			TestWorkClient: workClient,
		}
	})

	AfterEach(func() {
		controller.Finish()
	})

	It("creates postprocess work for the user requesting a clinic sync", func() {
		workClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, create *work.Create) (*work.Work, error) {
				Expect(create.Type).To(Equal("org.tidepool.data.upload.postprocess"))
				Expect(create.Metadata).To(HaveKeyWithValue("userId", userID))
				Expect(create.Metadata).To(HaveKeyWithValue("reasons", ConsistOf("CLINIC_SYNC")))
				return &work.Work{ID: "test-work-id", Type: create.Type}, nil
			})

		v1.SyncSummaries(svcCtx)

		Expect(statusCode()).To(Equal(http.StatusAccepted))
	})

	It("responds with internal server error when the work cannot be created", func() {
		workClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errorsTest.RandomError())

		v1.SyncSummaries(svcCtx)

		Expect(statusCode()).To(Equal(http.StatusInternalServerError))
	})
})
