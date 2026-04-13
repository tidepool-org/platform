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
	dataServiceTest "github.com/tidepool-org/platform/data/service/test"
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
			dependencies      ouraServiceApiV1.Dependencies
			handler           rest.HandlerFunc
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
			dependencies = ouraServiceApiV1.Dependencies{
				AuthClient: mockAuthClient,
				OuraClient: mockOuraClient,
				WorkClient: mockWorkClient,
			}
			verificationToken = ouraTest.RandomVerificationToken()
			challenge = ouraTest.RandomChallenge()
			query = url.Values{
				ouraServiceApiV1.QueryVerificationToken: []string{verificationToken},
				ouraServiceApiV1.QueryChallenge:         []string{challenge},
			}
			req.Request = httptest.NewRequestWithContext(ctx, "GET", "/?"+query.Encode(), nil)
		})

		withHandler := func() {
			It("returns http.StatusForbidden when the verification token is missing", func() {
				query.Del(ouraServiceApiV1.QueryVerificationToken)
				req.Request = httptest.NewRequestWithContext(ctx, "GET", "/?"+query.Encode(), nil)
				handler(res, req)
				Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusForbidden))
				Expect(res.ResponseRecorder).To(HaveHTTPBody("verification token is missing"))
				logger.AssertError("verification token is missing")
			})

			It("returns http.StatusForbidden when the challenge is missing", func() {
				query.Del(ouraServiceApiV1.QueryChallenge)
				req.Request = httptest.NewRequestWithContext(ctx, "GET", "/?"+query.Encode(), nil)
				handler(res, req)
				Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusForbidden))
				Expect(res.ResponseRecorder).To(HaveHTTPBody("challenge is missing"))
				logger.AssertError("challenge is missing")
			})

			It("returns http.StatusForbidden when the verification token is invalid", func() {
				mockOuraClient.EXPECT().PartnerSecret().Return(test.RandomString())
				handler(res, req)
				Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusForbidden))
				Expect(res.ResponseRecorder).To(HaveHTTPBody("verification token is invalid"))
				logger.AssertError("verification token is invalid", log.Fields{"challenge": challenge})
			})

			It("returns http.StatusOK with the challenge when the verification token is valid", func() {
				expectedBody, err := json.Marshal(map[string]string{"challenge": challenge})
				Expect(err).ToNot(HaveOccurred())
				Expect(expectedBody).ToNot(BeEmpty())
				mockOuraClient.EXPECT().PartnerSecret().Return(verificationToken)
				handler(res, req)
				Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusOK))
				Expect(res.ResponseRecorder).To(HaveHTTPBody(MatchJSON(expectedBody)))
			})
		}

		Context("with modern router", func() {
			BeforeEach(func() {
				router, err := ouraServiceApiV1.NewRouter(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(router).ToNot(BeNil())
				handler = func(res rest.ResponseWriter, req *rest.Request) {
					router.Subscription(res, req)
				}
			})

			withHandler()
		})

		Context("with legacy router", func() {
			var mockDataServiceContext *dataServiceTest.MockContext

			BeforeEach(func() {
				mockDataServiceContext = dataServiceTest.NewMockContext(mockController)
				handler = func(res rest.ResponseWriter, req *rest.Request) {
					mockDataServiceContext.EXPECT().Request().Return(req)
					mockDataServiceContext.EXPECT().Response().Return(res)
					ouraServiceApiV1.Subscription(mockDataServiceContext)
				}
			})

			It("returns http.StatusInternalServerError when the dependencies are invalid", func() {
				mockDataServiceContext.EXPECT().AuthClient().Return(nil)
				mockDataServiceContext.EXPECT().OuraClient().Return(nil)
				mockDataServiceContext.EXPECT().WorkClient().Return(nil)
				handler(res, req)
				Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusInternalServerError))
				Expect(res.ResponseRecorder).To(HaveHTTPBody(MatchJSON(internalServerErrorJSON)))
			})

			Context("with valid dependencies", func() {
				BeforeEach(func() {
					mockDataServiceContext.EXPECT().AuthClient().Return(mockAuthClient)
					mockDataServiceContext.EXPECT().OuraClient().Return(mockOuraClient)
					mockDataServiceContext.EXPECT().WorkClient().Return(mockWorkClient)
				})

				withHandler()
			})
		})
	})
})
