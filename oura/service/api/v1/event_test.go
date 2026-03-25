package v1_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"
	"go.uber.org/mock/gomock"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	ouraServiceApiV1 "github.com/tidepool-org/platform/oura/service/api/v1"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	testHttp "github.com/tidepool-org/platform/test/http"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("event", func() {
	Context("with request, response, and router", func() {
		var (
			logger         *logTest.Logger
			ctx            context.Context
			req            *rest.Request
			res            *testHttp.ResponseWriter
			mockController *gomock.Controller
			mockAuthClient *authTest.MockClient
			mockOuraClient *ouraTest.MockClient
			mockWorkClient *workTest.MockClient
			router         *ouraServiceApiV1.Router
		)

		BeforeEach(func() {
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
			req = &rest.Request{}
			res = testHttp.NewResponseWriter()
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockAuthClient = authTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			mockWorkClient = workTest.NewMockClient(mockController)
			dependencies := ouraServiceApiV1.Dependencies{
				AuthClient: mockAuthClient,
				OuraClient: mockOuraClient,
				WorkClient: mockWorkClient,
			}
			var err error
			router, err = ouraServiceApiV1.NewRouter(dependencies)
			Expect(err).ToNot(HaveOccurred())
			Expect(router).ToNot(BeNil())
		})

		It("returns http.StatusOK", func() {
			req.Request = httptest.NewRequestWithContext(ctx, "POST", "/", nil)
			router.Event(res, req)
			Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusOK))
			Expect(res.ResponseRecorder).To(HaveHTTPBody(BeEmpty()))
		})
	})
})
