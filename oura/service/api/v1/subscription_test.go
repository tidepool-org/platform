package v1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"
	"go.uber.org/mock/gomock"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	ouraServiceApiV1 "github.com/tidepool-org/platform/oura/service/api/v1"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("subscription", func() {
	It("QueryVerificationToken is expected", func() {
		Expect(ouraServiceApiV1.QueryVerificationToken).To(Equal("verification_token"))
	})

	It("QueryChallenge is expected", func() {
		Expect(ouraServiceApiV1.QueryChallenge).To(Equal("challenge"))
	})

	Context("with request, response, and router", func() {
		var (
			logger            *logTest.Logger
			ctx               context.Context
			req               *rest.Request
			res               *testHttp.ResponseWriter
			mockController    *gomock.Controller
			mockAuthClient    *authTest.MockClient
			mockOuraClient    *ouraTest.MockClient
			mockWorkClient    *workTest.MockClient
			router            *ouraServiceApiV1.Router
			verificationToken string
			challenge         string
			query             url.Values
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
			verificationToken = ouraTest.RandomVerificationToken()
			challenge = ouraTest.RandomChallenge()
			query = url.Values{
				ouraServiceApiV1.QueryVerificationToken: []string{verificationToken},
				ouraServiceApiV1.QueryChallenge:         []string{challenge},
			}
		})

		It("returns http.StatusForbidden when the verification token is missing", func() {
			query.Del(ouraServiceApiV1.QueryVerificationToken)
			req.Request = httptest.NewRequestWithContext(ctx, "GET", "/?"+query.Encode(), nil)
			router.Subscription(res, req)
			Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusForbidden))
			Expect(res.ResponseRecorder).To(HaveHTTPBody("verification token is missing"))
			logger.AssertError("verification token is missing")
		})

		It("returns http.StatusForbidden when the challenge is missing", func() {
			query.Del(ouraServiceApiV1.QueryChallenge)
			req.Request = httptest.NewRequestWithContext(ctx, "GET", "/?"+query.Encode(), nil)
			router.Subscription(res, req)
			Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusForbidden))
			Expect(res.ResponseRecorder).To(HaveHTTPBody("challenge is missing"))
			logger.AssertError("challenge is missing")
		})

		It("returns http.StatusForbidden when the verification token is invalid", func() {
			mockOuraClient.EXPECT().PartnerSecret().Return(test.RandomString())
			req.Request = httptest.NewRequestWithContext(ctx, "GET", "/?"+query.Encode(), nil)
			router.Subscription(res, req)
			Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusForbidden))
			Expect(res.ResponseRecorder).To(HaveHTTPBody("verification token is invalid"))
			logger.AssertError("verification token is invalid", log.Fields{"challenge": challenge})
		})

		It("returns http.StatusOK with the challenge when the verification token is valid", func() {
			expectedBody, err := json.Marshal(map[string]string{"challenge": challenge})
			Expect(err).ToNot(HaveOccurred())
			Expect(expectedBody).ToNot(BeEmpty())
			mockOuraClient.EXPECT().PartnerSecret().Return(verificationToken)
			req.Request = httptest.NewRequestWithContext(ctx, "GET", "/?"+query.Encode(), nil)
			router.Subscription(res, req)
			Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusOK))
			Expect(res.ResponseRecorder).To(HaveHTTPBody(MatchJSON(expectedBody)))
		})
	})
})
