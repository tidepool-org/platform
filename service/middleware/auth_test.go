package middleware_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"time"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	testAuth "github.com/tidepool-org/platform/auth/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/middleware"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Auth", func() {
	var serviceSecret string
	var authClient *testAuth.Client

	BeforeEach(func() {
		serviceSecret = testAuth.NewServiceSecret()
		authClient = testAuth.NewClient()
	})

	AfterEach(func() {
		authClient.Expectations()
	})

	Context("NewAuth", func() {
		It("returns an error if service secret is missing", func() {
			authMiddleware, err := middleware.NewAuth("", authClient)
			Expect(err).To(MatchError("service secret is missing"))
			Expect(authMiddleware).To(BeNil())
		})

		It("returns an error if auth client is missing", func() {
			authMiddleware, err := middleware.NewAuth(serviceSecret, nil)
			Expect(err).To(MatchError("auth client is missing"))
			Expect(authMiddleware).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(middleware.NewAuth(serviceSecret, authClient)).ToNot(BeNil())
		})
	})

	Context("with auth middleware", func() {
		var authMiddleware *middleware.Auth

		BeforeEach(func() {
			var err error
			authMiddleware, err = middleware.NewAuth(serviceSecret, authClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(authMiddleware).ToNot(BeNil())
		})

		Context("MiddlewareFunc", func() {
			It("does nothing if handlerFunc is nil", func() {
				middlewareFunc := authMiddleware.MiddlewareFunc(nil)
				Expect(middlewareFunc).ToNot(BeNil())
				middlewareFunc(testRest.NewResponseWriter(), testRest.NewRequest())
			})

			It("returns successfully", func() {
				Expect(authMiddleware.MiddlewareFunc(func(res rest.ResponseWriter, req *rest.Request) {})).ToNot(BeNil())
			})
		})

		Context("with response, request, and middleware func", func() {
			var lgr log.Logger
			var res *testRest.ResponseWriter
			var req *rest.Request
			var handlerFunc rest.HandlerFunc
			var middlewareFunc rest.HandlerFunc

			BeforeEach(func() {
				lgr = logNull.NewLogger()
				res = testRest.NewResponseWriter()
				req = testRest.NewRequest()
				req.Request = req.WithContext(log.NewContextWithLogger(req.Context(), lgr))
				service.SetRequestLogger(req, lgr)
				handlerFunc = nil
				middlewareFunc = authMiddleware.MiddlewareFunc(func(res rest.ResponseWriter, req *rest.Request) {
					Expect(res).ToNot(BeNil())
					Expect(req).ToNot(BeNil())
					handlerFunc(res, req)
				})
				Expect(middlewareFunc).ToNot(BeNil())
			})

			AfterEach(func() {
				res.Expectations()
				Expect(log.LoggerFromContext(req.Context())).To(Equal(lgr))
				Expect(service.GetRequestLogger(req)).To(Equal(lgr))
			})

			It("does nothing if response is nil", func() {
				middlewareFunc(nil, req)
				Expect(res.WriteHeaderInputs).To(BeEmpty())
			})

			It("does nothing if request is nil", func() {
				middlewareFunc(res, nil)
				Expect(res.WriteHeaderInputs).To(BeEmpty())
			})

			Context("with server session token", func() {
				var serverSessionToken string

				BeforeEach(func() {
					serverSessionToken = testAuth.NewSessionToken()
					authClient.ServerSessionTokenOutputs = []testAuth.ServerSessionTokenOutput{{Token: serverSessionToken, Error: nil}}
				})

				It("does not set the server session token if error", func() {
					authClient.ServerSessionTokenOutputs = []testAuth.ServerSessionTokenOutput{{Token: serverSessionToken, Error: testErrors.NewError()}}
					handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
						Expect(auth.ServerSessionTokenFromContext(req.Context())).To(BeEmpty())
					}
					middlewareFunc(res, req)
				})

				It("sets the server session token successfully", func() {
					handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
						Expect(auth.ServerSessionTokenFromContext(req.Context())).To(Equal(serverSessionToken))
					}
					middlewareFunc(res, req)
				})

				It("returns successfully with no details", func() {
					handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
						details := request.DetailsFromContext(req.Context())
						Expect(details).To(BeNil())
						Expect(service.GetRequestAuthDetails(req)).To(BeNil())
						Expect(log.LoggerFromContext(req.Context())).To(Equal(lgr))
						Expect(service.GetRequestLogger(req)).To(Equal(lgr))
					}
					middlewareFunc(res, req)
				})

				Context("with service secret", func() {
					BeforeEach(func() {
						req.Header.Add("X-Tidepool-Service-Secret", serviceSecret)
					})

					It("returns unauthorized if multiple values", func() {
						res.WriteJsonOutputs = []error{nil}
						req.Header.Add("X-Tidepool-Service-Secret", serviceSecret)
						middlewareFunc(res, req)
						Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					})

					It("returns unauthorized if the server secret does not match", func() {
						res.WriteJsonOutputs = []error{nil}
						req.Header.Set("X-Tidepool-Service-Secret", testAuth.NewServiceSecret())
						middlewareFunc(res, req)
						Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					})

					It("returns successfully", func() {
						handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
							details := request.DetailsFromContext(req.Context())
							Expect(details).ToNot(BeNil())
							Expect(details.Method()).To(Equal(request.MethodServiceSecret))
							Expect(details.IsService()).To(BeTrue())
							Expect(details.HasToken()).To(BeFalse())
							Expect(service.GetRequestAuthDetails(req)).To(Equal(details))
							Expect(log.LoggerFromContext(req.Context())).To(Equal(lgr))
							Expect(service.GetRequestLogger(req)).To(Equal(lgr))
						}
						middlewareFunc(res, req)
					})
				})

				Context("with access token", func() {
					var accessToken string

					BeforeEach(func() {
						accessToken = testAuth.NewAccessToken()
						req.Header.Add("Authorization", fmt.Sprintf("bEaReR %s", accessToken))
					})

					It("returns unauthorized if multiple values", func() {
						res.WriteJsonOutputs = []error{nil}
						req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
						middlewareFunc(res, req)
						Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					})

					It("returns unauthorized if not valid header", func() {
						res.WriteJsonOutputs = []error{nil}
						req.Header.Set("Authorization", accessToken)
						middlewareFunc(res, req)
						Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					})

					It("returns unauthorized if not Bearer token", func() {
						res.WriteJsonOutputs = []error{nil}
						req.Header.Set("Authorization", fmt.Sprintf("NotBearer %s", accessToken))
						middlewareFunc(res, req)
						Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					})

					It("returns successfully", func() {
						userID := id.New()
						authClient.ValidateSessionTokenOutputs = []testAuth.ValidateSessionTokenOutput{
							{Details: request.NewDetails(request.MethodSessionToken, userID, accessToken), Error: nil},
						}
						handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
							details := request.DetailsFromContext(req.Context())
							Expect(details).ToNot(BeNil())
							Expect(details.Method()).To(Equal(request.MethodAccessToken))
							Expect(details.IsUser()).To(BeTrue())
							Expect(details.UserID()).To(Equal(userID))
							Expect(details.Token()).To(Equal(accessToken))
							Expect(service.GetRequestAuthDetails(req)).To(Equal(details))
							Expect(log.LoggerFromContext(req.Context())).ToNot(BeNil())
							Expect(log.LoggerFromContext(req.Context())).ToNot(Equal(lgr))
							Expect(service.GetRequestLogger(req)).ToNot(BeNil())
							Expect(service.GetRequestLogger(req)).ToNot(Equal(lgr))
						}
						middlewareFunc(res, req)
						Expect(authClient.ValidateSessionTokenInputs).To(HaveLen(1))
						Expect(authClient.ValidateSessionTokenInputs[0].Token).To(Equal(accessToken))
					})

					It("returns successfully with no details if access token is not valid", func() {
						authClient.ValidateSessionTokenOutputs = []testAuth.ValidateSessionTokenOutput{{Details: nil, Error: testErrors.NewError()}}
						handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
							details := request.DetailsFromContext(req.Context())
							Expect(details).To(BeNil())
							Expect(service.GetRequestAuthDetails(req)).To(BeNil())
							Expect(log.LoggerFromContext(req.Context())).To(Equal(lgr))
							Expect(service.GetRequestLogger(req)).To(Equal(lgr))
						}
						middlewareFunc(res, req)
						Expect(authClient.ValidateSessionTokenInputs).To(HaveLen(1))
						Expect(authClient.ValidateSessionTokenInputs[0].Token).To(Equal(accessToken))
					})
				})

				Context("with session token", func() {
					var sessionToken string

					BeforeEach(func() {
						sessionToken = testAuth.NewSessionToken()
						req.Header.Add("X-Tidepool-Session-Token", sessionToken)
					})

					It("returns unauthorized if multiple values", func() {
						res.WriteJsonOutputs = []error{nil}
						req.Header.Add("X-Tidepool-Session-Token", sessionToken)
						middlewareFunc(res, req)
						Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					})

					It("returns successfully", func() {
						userID := id.New()
						authClient.ValidateSessionTokenOutputs = []testAuth.ValidateSessionTokenOutput{
							{Details: request.NewDetails(request.MethodSessionToken, userID, sessionToken), Error: nil},
						}
						handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
							details := request.DetailsFromContext(req.Context())
							Expect(details).ToNot(BeNil())
							Expect(details.Method()).To(Equal(request.MethodSessionToken))
							Expect(details.IsUser()).To(BeTrue())
							Expect(details.UserID()).To(Equal(userID))
							Expect(details.Token()).To(Equal(sessionToken))
							Expect(service.GetRequestAuthDetails(req)).To(Equal(details))
							Expect(log.LoggerFromContext(req.Context())).ToNot(BeNil())
							Expect(log.LoggerFromContext(req.Context())).ToNot(Equal(lgr))
							Expect(service.GetRequestLogger(req)).ToNot(BeNil())
							Expect(service.GetRequestLogger(req)).ToNot(Equal(lgr))
						}
						middlewareFunc(res, req)
						Expect(authClient.ValidateSessionTokenInputs).To(HaveLen(1))
						Expect(authClient.ValidateSessionTokenInputs[0].Token).To(Equal(sessionToken))
					})

					It("returns successfully as service", func() {
						authClient.ValidateSessionTokenOutputs = []testAuth.ValidateSessionTokenOutput{
							{Details: request.NewDetails(request.MethodSessionToken, "", sessionToken), Error: nil},
						}
						handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
							details := request.DetailsFromContext(req.Context())
							Expect(details).ToNot(BeNil())
							Expect(details.Method()).To(Equal(request.MethodSessionToken))
							Expect(details.IsService()).To(BeTrue())
							Expect(details.Token()).To(Equal(sessionToken))
							Expect(service.GetRequestAuthDetails(req)).To(Equal(details))
							Expect(log.LoggerFromContext(req.Context())).ToNot(BeNil())
							Expect(log.LoggerFromContext(req.Context())).ToNot(Equal(lgr))
							Expect(service.GetRequestLogger(req)).ToNot(BeNil())
							Expect(service.GetRequestLogger(req)).ToNot(Equal(lgr))
						}
						middlewareFunc(res, req)
						Expect(authClient.ValidateSessionTokenInputs).To(HaveLen(1))
						Expect(authClient.ValidateSessionTokenInputs[0].Token).To(Equal(sessionToken))
					})

					It("returns successfully with no details if session token is not valid", func() {
						authClient.ValidateSessionTokenOutputs = []testAuth.ValidateSessionTokenOutput{{Details: nil, Error: testErrors.NewError()}}
						handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
							details := request.DetailsFromContext(req.Context())
							Expect(details).To(BeNil())
							Expect(service.GetRequestAuthDetails(req)).To(BeNil())
							Expect(log.LoggerFromContext(req.Context())).To(Equal(lgr))
							Expect(service.GetRequestLogger(req)).To(Equal(lgr))
						}
						middlewareFunc(res, req)
						Expect(authClient.ValidateSessionTokenInputs).To(HaveLen(1))
						Expect(authClient.ValidateSessionTokenInputs[0].Token).To(Equal(sessionToken))
					})
				})

				Context("with restricted token", func() {
					var restrictedToken string

					BeforeEach(func() {
						restrictedToken = testAuth.NewRestrictedToken()
						query := req.URL.Query()
						query.Add("restricted_token", restrictedToken)
						req.URL.RawQuery = query.Encode()
					})

					It("returns unauthorized if multiple values", func() {
						res.WriteJsonOutputs = []error{nil}
						query := req.URL.Query()
						query.Add("restricted_token", restrictedToken)
						req.URL.RawQuery = query.Encode()
						middlewareFunc(res, req)
						Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					})

					It("returns successfully", func() {
						userID := id.New()
						restrictedTokenObject := &auth.RestrictedToken{
							ID:             restrictedToken,
							UserID:         userID,
							ExpirationTime: time.Now().Add(time.Hour),
						}
						authClient.GetRestrictedTokenOutputs = []testAuth.GetRestrictedTokenOutput{{RestrictedToken: restrictedTokenObject, Error: nil}}
						handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
							details := request.DetailsFromContext(req.Context())
							Expect(details).ToNot(BeNil())
							Expect(details.Method()).To(Equal(request.MethodRestrictedToken))
							Expect(details.IsUser()).To(BeTrue())
							Expect(details.UserID()).To(Equal(userID))
							Expect(details.Token()).To(Equal(restrictedToken))
							Expect(service.GetRequestAuthDetails(req)).To(Equal(details))
							Expect(log.LoggerFromContext(req.Context())).ToNot(BeNil())
							Expect(log.LoggerFromContext(req.Context())).ToNot(Equal(lgr))
							Expect(service.GetRequestLogger(req)).ToNot(BeNil())
							Expect(service.GetRequestLogger(req)).ToNot(Equal(lgr))
						}
						middlewareFunc(res, req)
						Expect(authClient.GetRestrictedTokenInputs).To(HaveLen(1))
						Expect(authClient.GetRestrictedTokenInputs[0].ID).To(Equal(restrictedToken))
					})

					It("returns successfully with no details if restricted token is not valid", func() {
						authClient.GetRestrictedTokenOutputs = []testAuth.GetRestrictedTokenOutput{{RestrictedToken: nil, Error: testErrors.NewError()}}
						handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
							details := request.DetailsFromContext(req.Context())
							Expect(details).To(BeNil())
							Expect(service.GetRequestAuthDetails(req)).To(BeNil())
							Expect(log.LoggerFromContext(req.Context())).To(Equal(lgr))
							Expect(service.GetRequestLogger(req)).To(Equal(lgr))
						}
						middlewareFunc(res, req)
						Expect(authClient.GetRestrictedTokenInputs).To(HaveLen(1))
						Expect(authClient.GetRestrictedTokenInputs[0].ID).To(Equal(restrictedToken))
					})

					It("returns successfully with no details if restricted token is missing", func() {
						authClient.GetRestrictedTokenOutputs = []testAuth.GetRestrictedTokenOutput{{RestrictedToken: nil, Error: nil}}
						handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
							details := request.DetailsFromContext(req.Context())
							Expect(details).To(BeNil())
							Expect(service.GetRequestAuthDetails(req)).To(BeNil())
							Expect(log.LoggerFromContext(req.Context())).To(Equal(lgr))
							Expect(service.GetRequestLogger(req)).To(Equal(lgr))
						}
						middlewareFunc(res, req)
						Expect(authClient.GetRestrictedTokenInputs).To(HaveLen(1))
						Expect(authClient.GetRestrictedTokenInputs[0].ID).To(Equal(restrictedToken))
					})

					It("returns successfully with no details if restricted token does not authenticate request", func() {
						userID := id.New()
						restrictedTokenObject := &auth.RestrictedToken{
							ID:             restrictedToken,
							UserID:         userID,
							ExpirationTime: time.Now().Add(-time.Hour),
						}
						authClient.GetRestrictedTokenOutputs = []testAuth.GetRestrictedTokenOutput{{RestrictedToken: restrictedTokenObject, Error: nil}}
						handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
							details := request.DetailsFromContext(req.Context())
							Expect(details).To(BeNil())
							Expect(service.GetRequestAuthDetails(req)).To(BeNil())
							Expect(log.LoggerFromContext(req.Context())).To(Equal(lgr))
							Expect(service.GetRequestLogger(req)).To(Equal(lgr))
						}
						middlewareFunc(res, req)
						Expect(authClient.GetRestrictedTokenInputs).To(HaveLen(1))
						Expect(authClient.GetRestrictedTokenInputs[0].ID).To(Equal(restrictedToken))
					})
				})
			})
		})
	})
})
