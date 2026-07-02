package v1_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"

	"github.com/ant0ine/go-json-rest/rest"
	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	authServiceApiV1 "github.com/tidepool-org/platform/auth/service/api/v1"
	authServiceTest "github.com/tidepool-org/platform/auth/service/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testRest "github.com/tidepool-org/platform/test/rest"
	"github.com/tidepool-org/platform/times"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("ProviderSession", func() {
	var lgr log.Logger
	var ctx context.Context
	var mockController *gomock.Controller
	var mockService *authServiceTest.MockService
	var router *authServiceApiV1.Router
	var req *rest.Request
	var res *testRest.MockResponse
	var handlerFunc rest.HandlerFunc

	BeforeEach(func() {
		lgr = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), lgr)
		mockController, ctx = gomock.WithContext(ctx, GinkgoT())
		mockService = authServiceTest.NewMockService(mockController)
		router = test.Must(authServiceApiV1.NewRouter(mockService))
		req = testRest.NewRequest()
		req.Request = req.WithContext(ctx)
		res = testRest.NewMockResponse(mockController)
		res.EXPECT().Header().Return(http.Header{}).AnyTimes()
	})

	JustBeforeEach(func() {
		handlerFunc = test.Must(rest.MakeRouter(router.Routes()...)).AppFunc()
	})

	Context("RefreshProviderSession", func() {
		var providerSessionID string

		BeforeEach(func() {
			providerSessionID = authTest.RandomProviderSessionID()
			req.Method = http.MethodPost
			req.URL.Path = fmt.Sprintf("/v1/provider_sessions/%s/refresh", providerSessionID)
		})

		It("returns an error if not authenticated", func() {
			expectErr := request.ErrorUnauthenticated()
			res.EXPECT().WriteHeader(http.StatusUnauthorized)
			res.EXPECT().Write(test.MarshalRequestBody(errors.NewSerializable(errors.Sanitize(expectErr))))
			handlerFunc(res, req)
		})

		It("returns an error if authenticated as user", func() {
			expectErr := request.ErrorUnauthorized()
			req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), request.NewAuthDetails(request.MethodSessionToken, userTest.RandomUserID(), authTest.RandomSessionToken())))
			res.EXPECT().WriteHeader(http.StatusForbidden)
			res.EXPECT().Write(test.MarshalRequestBody(errors.NewSerializable(errors.Sanitize(expectErr))))
			handlerFunc(res, req)
		})

		When("authenticated as service", func() {
			BeforeEach(func() {
				req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), request.NewAuthDetails(request.MethodSessionToken, "", authTest.RandomSessionToken())))
			})

			It("returns an error if the id is missing", func() {
				expectErr := request.ErrorParameterMissing("id")
				req.URL.Path = "/v1/provider_sessions//refresh"
				res.EXPECT().WriteHeader(http.StatusBadRequest)
				res.EXPECT().Write(test.MarshalRequestBody(errors.NewSerializable(errors.Sanitize(expectErr))))
				handlerFunc(res, req)
			})

			It("returns an error if the body is invalid", func() {
				expectErr := errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timeRange/from")
				req.Body = io.NopCloser(bytes.NewBuffer(test.MarshalRequestBody(&auth.ProviderSessionRefresh{TimeRange: &times.TimeRange{From: pointer.From(time.Time{})}})))
				res.EXPECT().WriteHeader(http.StatusBadRequest)
				res.EXPECT().Write(test.MarshalRequestBody(errors.NewSerializable(errors.Sanitize(expectErr))))
				handlerFunc(res, req)
			})

			Context("with auth client and refresh", func() {
				var mockAuthClient *authTest.MockClient
				var refresh *auth.ProviderSessionRefresh

				BeforeEach(func() {
					mockAuthClient = authTest.NewMockClient(mockController)
					mockService.EXPECT().AuthClient().Return(mockAuthClient).AnyTimes()
					refresh = authTest.RandomProviderSessionRefresh(test.AllowOptionals())
					req.Body = io.NopCloser(bytes.NewBuffer(test.MarshalRequestBody(refresh)))
				})

				It("returns an error if the auth client returns an error", func() {
					testErr := errorsTest.RandomError()
					mockAuthClient.EXPECT().RefreshProviderSession(gomock.Not(gomock.Nil()), providerSessionID, refresh).Return(nil, testErr)
					res.EXPECT().WriteHeader(http.StatusInternalServerError)
					res.EXPECT().Write(test.MarshalRequestBody(errors.NewSerializable(errors.Sanitize(testErr))))
					handlerFunc(res, req)
				})

				It("returns not found error if the provider session is not found", func() {
					mockAuthClient.EXPECT().RefreshProviderSession(gomock.Not(gomock.Nil()), providerSessionID, refresh).Return(nil, nil)
					res.EXPECT().WriteHeader(http.StatusNotFound)
					res.EXPECT().Write(test.MarshalRequestBody(errors.NewSerializable(errors.Sanitize(request.ErrorResourceNotFoundWithID(providerSessionID)))))
					handlerFunc(res, req)
				})

				It("returns the refreshed provider session", func() {
					providerSession := authTest.RandomProviderSession()
					providerSession.ID = providerSessionID
					mockAuthClient.EXPECT().RefreshProviderSession(gomock.Not(gomock.Nil()), providerSessionID, refresh).Return(providerSession, nil)
					res.EXPECT().WriteHeader(http.StatusOK)
					res.EXPECT().Write(test.MarshalRequestBody(providerSession))
					handlerFunc(res, req)
				})
			})
		})
	})
})
