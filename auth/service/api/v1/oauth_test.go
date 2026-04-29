package v1_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/golang-jwt/jwt/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/auth"
	authServiceApiV1 "github.com/tidepool-org/platform/auth/service/api/v1"
	authServiceTest "github.com/tidepool-org/platform/auth/service/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/permission"
	permissionTest "github.com/tidepool-org/platform/permission/test"
	"github.com/tidepool-org/platform/provider"
	providerTest "github.com/tidepool-org/platform/provider/test"
	"github.com/tidepool-org/platform/request"
	serviceTest "github.com/tidepool-org/platform/service/test"
	testRest "github.com/tidepool-org/platform/test/rest"
)

// mockOAuthProvider is a minimal mock of oauth.Provider that supports user-initiated unlinking
type mockOAuthProvider struct {
	provider.Provider
	typeField              string
	nameField              string
	useCookieField         bool
	supportsUnlinkingField bool
}

func (m *mockOAuthProvider) Type() string    { return m.typeField }
func (m *mockOAuthProvider) Name() string    { return m.nameField }
func (m *mockOAuthProvider) UseCookie() bool { return m.useCookieField }
func (m *mockOAuthProvider) SupportsUserInitiatedAccountUnlinking() bool {
	return m.supportsUnlinkingField
}
func (m *mockOAuthProvider) OnCreate(context.Context, *auth.ProviderSession) error { return nil }
func (m *mockOAuthProvider) OnDelete(context.Context, *auth.ProviderSession) error { return nil }
func (m *mockOAuthProvider) TokenSource(context.Context, *auth.OAuthToken) (oauth2.TokenSource, error) {
	return nil, nil
}
func (m *mockOAuthProvider) ParseToken(string, jwt.Claims) error            { return nil }
func (m *mockOAuthProvider) CalculateStateForRestrictedToken(string) string { return "" }
func (m *mockOAuthProvider) GetAuthorizationCodeURLWithState(string) string { return "" }
func (m *mockOAuthProvider) ExchangeAuthorizationCodeForToken(context.Context, string) (*auth.OAuthToken, error) {
	return nil, nil
}
func (m *mockOAuthProvider) IsErrorCodeAccessDenied(string) bool { return false }

var _ = Describe("OAuth", func() {
	var svc *authServiceTest.Service
	var res *testRest.ResponseWriter
	var req *rest.Request
	var ctx context.Context
	var handlerFunc rest.HandlerFunc

	BeforeEach(func() {
		svc = authServiceTest.NewService()
		svc.ProviderFactoryImpl.GetOutputs = []providerTest.GetOutput{{
			Provider: &mockOAuthProvider{
				typeField:              "example",
				nameField:              "example",
				useCookieField:         false,
				supportsUnlinkingField: true,
			},
			Error: nil,
		}}
		res = testRest.NewResponseWriter()
		res.HeaderOutput = &http.Header{}
		req = testRest.NewRequest()
		ctx = log.NewContextWithLogger(req.Context(), logTest.NewLogger())
		req.Request = req.WithContext(ctx)
	})

	JustBeforeEach(func() {
		router, err := authServiceApiV1.NewRouter(svc)
		Expect(err).ToNot(HaveOccurred())
		Expect(router).ToNot(BeNil())
		app, err := rest.MakeRouter(router.Routes()...)
		Expect(err).ToNot(HaveOccurred())
		Expect(app).ToNot(BeNil())
		handlerFunc = app.AppFunc()
	})

	Describe("UserOAuthProviderAuthorizeDelete", func() {
		var userID string

		BeforeEach(func() {
			userID = serviceTest.NewUserID()
			req.Method = http.MethodDelete
			req.URL.Path = fmt.Sprintf("/v1/users/%s/oauth/example/authorize", userID)
		})

		Context("with user details (same user)", func() {
			BeforeEach(func() {
				svc.AuthClientImpl.GetUserPermissionsOutputs = []permissionTest.GetUserPermissionsOutput{{
					Permissions: permission.Permissions{
						permission.Owner: permission.Permission{},
					},
					Error: nil,
				}}
				svc.AuthClientImpl.ListProviderSessionsOutputs = []authTest.ListProviderSessionsOutput{{
					ProviderSessions: auth.ProviderSessions{},
					Error:            nil,
				}}
				details := request.NewAuthDetails(request.MethodSessionToken, userID, authTest.NewSessionToken())
				req.Request = req.WithContext(request.NewContextWithAuthDetails(ctx, details))
			})

			It("allows access when user has write permissions to themselves", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				handlerFunc(res, req)
				Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
			})
		})

		Context("with user details (different user)", func() {
			BeforeEach(func() {
				targetUserID := serviceTest.NewUserID()
				svc.AuthClientImpl.GetUserPermissionsOutputs = []permissionTest.GetUserPermissionsOutput{{
					Permissions: permission.Permissions{
						permission.Custodian: permission.Permission{},
					},
					Error: nil,
				}}
				svc.AuthClientImpl.ListProviderSessionsOutputs = []authTest.ListProviderSessionsOutput{{
					ProviderSessions: auth.ProviderSessions{},
					Error:            nil,
				}}
				details := request.NewAuthDetails(request.MethodSessionToken, userID, authTest.NewSessionToken())
				req.URL.Path = fmt.Sprintf("/v1/users/%s/oauth/example/authorize", targetUserID)
				req.Request = req.WithContext(request.NewContextWithAuthDetails(ctx, details))
			})

			It("allows access when user has write permissions to the target user", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				handlerFunc(res, req)
				Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
			})
		})

		Context("with service details", func() {
			BeforeEach(func() {
				svc.AuthClientImpl.ListProviderSessionsOutputs = []authTest.ListProviderSessionsOutput{{
					ProviderSessions: auth.ProviderSessions{},
					Error:            nil,
				}}
				details := request.NewAuthDetails(request.MethodSessionToken, "", authTest.NewSessionToken())
				req.Request = req.WithContext(request.NewContextWithAuthDetails(ctx, details))
			})

			It("allows access for service tokens", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				handlerFunc(res, req)
				Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
			})
		})

		Context("with user details but no write permissions", func() {
			BeforeEach(func() {
				targetUserID := serviceTest.NewUserID()
				svc.AuthClientImpl.GetUserPermissionsOutputs = []permissionTest.GetUserPermissionsOutput{{
					Permissions: permission.Permissions{},
					Error:       nil,
				}}
				details := request.NewAuthDetails(request.MethodSessionToken, userID, authTest.NewSessionToken())
				req.URL.Path = fmt.Sprintf("/v1/users/%s/oauth/example/authorize", targetUserID)
				req.Request = req.WithContext(request.NewContextWithAuthDetails(ctx, details))
			})

			It("denies access when user has no write permissions to the target user", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				handlerFunc(res, req)
				Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
			})
		})

		Context("with user details trying to access a different user without self-access", func() {
			BeforeEach(func() {
				targetUserID := serviceTest.NewUserID()
				svc.AuthClientImpl.GetUserPermissionsOutputs = []permissionTest.GetUserPermissionsOutput{{
					Permissions: permission.Permissions{},
					Error:       nil,
				}}
				details := request.NewAuthDetails(request.MethodSessionToken, targetUserID, authTest.NewSessionToken())
				req.URL.Path = fmt.Sprintf("/v1/users/%s/oauth/example/authorize", userID)
				req.Request = req.WithContext(request.NewContextWithAuthDetails(ctx, details))
			})

			It("denies access when user tries to access another user without permissions", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				handlerFunc(res, req)
				Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
			})
		})

		Context("when unlinking is not supported by the provider", func() {
			BeforeEach(func() {
				svc.ProviderFactoryImpl.GetOutputs = []providerTest.GetOutput{{
					Provider: &mockOAuthProvider{
						typeField:              "example",
						nameField:              "example",
						useCookieField:         false,
						supportsUnlinkingField: false,
					},
					Error: nil,
				}}
				svc.AuthClientImpl.GetUserPermissionsOutputs = []permissionTest.GetUserPermissionsOutput{{
					Permissions: permission.Permissions{
						permission.Owner: permission.Permission{},
					},
					Error: nil,
				}}
				details := request.NewAuthDetails(request.MethodSessionToken, userID, authTest.NewSessionToken())
				req.Request = req.WithContext(request.NewContextWithAuthDetails(ctx, details))
			})

			It("responds with forbidden when provider does not support user-initiated unlinking", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				handlerFunc(res, req)
				Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
			})
		})

		Context("when provider session deletion fails", func() {
			BeforeEach(func() {
				svc.AuthClientImpl.GetUserPermissionsOutputs = []permissionTest.GetUserPermissionsOutput{{
					Permissions: permission.Permissions{
						permission.Owner: permission.Permission{},
					},
					Error: nil,
				}}
				svc.AuthClientImpl.ListProviderSessionsOutputs = []authTest.ListProviderSessionsOutput{{
					ProviderSessions: auth.ProviderSessions{
						&auth.ProviderSession{ID: "session-1", Type: "example"},
					},
					Error: nil,
				}}
				svc.AuthClientImpl.DeleteProviderSessionOutputs = []error{fmt.Errorf("deletion failed")}
				details := request.NewAuthDetails(request.MethodSessionToken, userID, authTest.NewSessionToken())
				req.Request = req.WithContext(request.NewContextWithAuthDetails(ctx, details))
			})

			It("responds with internal server error when session deletion fails", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				handlerFunc(res, req)
				Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
			})
		})

		Context("when the userId is invalid", func() {
			BeforeEach(func() {
				req.URL.Path = "/v1/users//oauth/example/authorize"
			})

			It("responds with bad request when userId is empty", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				handlerFunc(res, req)
				Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
			})
		})
	})

	Describe("OAuthProviderAuthorizeDelete", func() {
		BeforeEach(func() {
			req.Method = http.MethodDelete
			req.URL.Path = "/v1/oauth/example/authorize"
		})

		When("the auth details are missing", func() {
			It("responds with unauthenticated error", func() {
				res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
				req.Request = req.WithContext(request.NewContextWithAuthDetails(ctx, nil))
				handlerFunc(res, req)
				Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
			})
		})
	})
})
