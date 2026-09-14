package v1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"
	"go.uber.org/mock/gomock"

	v1 "github.com/tidepool-org/platform/data/service/api/v1"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	testRest "github.com/tidepool-org/platform/test/rest"
	"github.com/tidepool-org/platform/work"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("Work", func() {
	var controller *gomock.Controller
	var workClient *workTest.MockClient
	var res *testRest.ResponseWriter
	var svcCtx *mockDataServiceContext

	newRestRequest := func(method string, body string, pathParams map[string]string) *rest.Request {
		req := httptest.NewRequest(method, "/v1/work", strings.NewReader(body))
		req = req.WithContext(log.NewContextWithLogger(req.Context(), logTest.NewLogger()))
		return &rest.Request{Request: req, PathParams: pathParams, Env: map[string]interface{}{}}
	}

	statusCode := func() int {
		Expect(res.WriteHeaderInputs).To(HaveLen(1))
		return res.WriteHeaderInputs[0]
	}

	respondedObject := func() map[string]interface{} {
		Expect(res.WriteInputs).To(HaveLen(1))
		var object map[string]interface{}
		Expect(json.Unmarshal(res.WriteInputs[0], &object)).To(Succeed())
		return object
	}

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		workClient = workTest.NewMockClient(controller)
		res = testRest.NewResponseWriter()
		res.HeaderOutput = &http.Header{}
		res.WriteStub = func(bites []byte) (int, error) { return len(bites), nil }
		svcCtx = &mockDataServiceContext{
			RestResponse:   res,
			TestWorkClient: workClient,
		}
	})

	AfterEach(func() {
		controller.Finish()
	})

	Context("CreateWork", func() {
		createBody := func(availableTime *time.Time) string {
			body := map[string]interface{}{
				"type":              "org.tidepool.data.upload.postprocess",
				"groupId":           "org.tidepool.data.upload.postprocess:test-user-id",
				"serialId":          "org.tidepool.data.upload.postprocess:test-user-id",
				"processingTimeout": 300,
				"metadata":          map[string]interface{}{"userId": "test-user-id", "reasons": []string{"LEGACY_DATA_ADDED"}},
			}
			if availableTime != nil {
				body["processingAvailableTime"] = availableTime.Format(time.RFC3339Nano)
			}
			bites, err := json.Marshal(body)
			Expect(err).ToNot(HaveOccurred())
			return string(bites)
		}

		It("creates the work and responds with the created document", func() {
			availableTime := time.Now().Add(90 * time.Second)
			svcCtx.RestRequest = newRestRequest(http.MethodPost, createBody(&availableTime), nil)
			workClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, create *work.Create) (*work.Work, error) {
					Expect(create.Type).To(Equal("org.tidepool.data.upload.postprocess"))
					Expect(create.GroupID).To(HaveValue(Equal("org.tidepool.data.upload.postprocess:test-user-id")))
					Expect(create.SerialID).To(HaveValue(Equal("org.tidepool.data.upload.postprocess:test-user-id")))
					Expect(create.ProcessingTimeout).To(Equal(300))
					Expect(create.ProcessingAvailableTime).To(BeTemporally("~", availableTime, time.Second))
					Expect(create.Metadata).To(HaveKeyWithValue("userId", "test-user-id"))
					Expect(create.Metadata).To(HaveKeyWithValue("reasons", ConsistOf("LEGACY_DATA_ADDED")))
					return &work.Work{ID: "test-work-id", Type: create.Type}, nil
				})

			v1.CreateWork(svcCtx)

			Expect(statusCode()).To(Equal(http.StatusCreated))
			Expect(respondedObject()).To(HaveKeyWithValue("id", "test-work-id"))
		})

		It("responds with no content when the work is deduplicated", func() {
			svcCtx.RestRequest = newRestRequest(http.MethodPost, createBody(nil), nil)
			workClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil)

			v1.CreateWork(svcCtx)

			Expect(statusCode()).To(Equal(http.StatusNoContent))
			Expect(res.WriteInputs).To(BeEmpty())
		})

		It("responds with bad request when the body is malformed", func() {
			svcCtx.RestRequest = newRestRequest(http.MethodPost, "{malformed", nil)

			v1.CreateWork(svcCtx)

			Expect(statusCode()).To(Equal(http.StatusBadRequest))
		})

		It("responds with bad request when the create is invalid", func() {
			svcCtx.RestRequest = newRestRequest(http.MethodPost, "{}", nil)

			v1.CreateWork(svcCtx)

			Expect(statusCode()).To(Equal(http.StatusBadRequest))
		})

		It("responds with internal server error when the work cannot be created", func() {
			svcCtx.RestRequest = newRestRequest(http.MethodPost, createBody(nil), nil)
			workClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errorsTest.RandomError())

			v1.CreateWork(svcCtx)

			Expect(statusCode()).To(Equal(http.StatusInternalServerError))
		})
	})

	Context("GetWork", func() {
		It("responds with the work", func() {
			svcCtx.RestRequest = newRestRequest(http.MethodGet, "", map[string]string{"id": "test-work-id"})
			workClient.EXPECT().Get(gomock.Any(), "test-work-id", gomock.Nil()).
				Return(&work.Work{ID: "test-work-id"}, nil)

			v1.GetWork(svcCtx)

			Expect(statusCode()).To(Equal(http.StatusOK))
			Expect(respondedObject()).To(HaveKeyWithValue("id", "test-work-id"))
		})

		It("responds with not found when the work does not exist", func() {
			svcCtx.RestRequest = newRestRequest(http.MethodGet, "", map[string]string{"id": "test-work-id"})
			workClient.EXPECT().Get(gomock.Any(), "test-work-id", gomock.Nil()).Return(nil, nil)

			v1.GetWork(svcCtx)

			Expect(statusCode()).To(Equal(http.StatusNotFound))
		})

		It("responds with bad request when the id is missing", func() {
			svcCtx.RestRequest = newRestRequest(http.MethodGet, "", nil)

			v1.GetWork(svcCtx)

			Expect(statusCode()).To(Equal(http.StatusBadRequest))
		})

		It("responds with internal server error when the work cannot be gotten", func() {
			svcCtx.RestRequest = newRestRequest(http.MethodGet, "", map[string]string{"id": "test-work-id"})
			workClient.EXPECT().Get(gomock.Any(), "test-work-id", gomock.Nil()).Return(nil, errorsTest.RandomError())

			v1.GetWork(svcCtx)

			Expect(statusCode()).To(Equal(http.StatusInternalServerError))
		})
	})
})
