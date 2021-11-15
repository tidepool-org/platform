package middleware_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mdblp/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/service/middleware"
	serviceTest "github.com/tidepool-org/platform/service/test"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Auth", func() {
	var serviceSecret string
	var authClient *authTest.Client

	BeforeEach(func() {
		serviceSecret = authTest.NewServiceSecret()
		authClient = authTest.NewClient()
	})

	AfterEach(func() {
		authClient.AssertOutputsEmpty()
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
				res.AssertOutputsEmpty()
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
					serverSessionToken = authTest.NewSessionToken()
					authClient.ServerSessionTokenOutputs = []authTest.ServerSessionTokenOutput{{Token: serverSessionToken, Error: nil}}
				})

				It("does not set the server session token if error", func() {
					authClient.ServerSessionTokenOutputs = []authTest.ServerSessionTokenOutput{{Token: serverSessionToken, Error: errorsTest.RandomError()}}
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
						res.HeaderOutput = &http.Header{}
						res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						req.Header.Add("X-Tidepool-Service-Secret", serviceSecret)
						middlewareFunc(res, req)
						Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					})

					It("returns unauthorized if the server secret does not match", func() {
						res.HeaderOutput = &http.Header{}
						res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						req.Header.Set("X-Tidepool-Service-Secret", authTest.NewServiceSecret())
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
						accessToken = authTest.NewAccessToken()
						req.Header.Add("Authorization", fmt.Sprintf("bEaReR %s", accessToken))
					})

					It("returns unauthorized if multiple values", func() {
						res.HeaderOutput = &http.Header{}
						res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
						middlewareFunc(res, req)
						Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					})

					It("returns unauthorized if not valid header", func() {
						res.HeaderOutput = &http.Header{}
						res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						req.Header.Set("Authorization", accessToken)
						middlewareFunc(res, req)
						Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					})

					It("returns unauthorized if not Bearer token", func() {
						res.HeaderOutput = &http.Header{}
						res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						req.Header.Set("Authorization", fmt.Sprintf("NotBearer %s", accessToken))
						middlewareFunc(res, req)
						Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					})

					It("returns successfully", func() {
						userID := serviceTest.NewUserID()
						authClient.ValidateSessionTokenOutputs = []authTest.ValidateSessionTokenOutput{
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
						Expect(authClient.ValidateSessionTokenInputs).To(Equal([]string{accessToken}))
					})

					It("returns successfully with no details if access token is not valid", func() {
						authClient.ValidateSessionTokenOutputs = []authTest.ValidateSessionTokenOutput{{Details: nil, Error: errorsTest.RandomError()}}
						handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
							details := request.DetailsFromContext(req.Context())
							Expect(details).To(BeNil())
							Expect(service.GetRequestAuthDetails(req)).To(BeNil())
							Expect(log.LoggerFromContext(req.Context())).To(Equal(lgr))
							Expect(service.GetRequestLogger(req)).To(Equal(lgr))
						}
						middlewareFunc(res, req)
						Expect(authClient.ValidateSessionTokenInputs).To(Equal([]string{accessToken}))
					})
				})

				Context("with session token", func() {
					var sessionToken string

					BeforeEach(func() {
						sessionToken = authTest.NewSessionToken()
						req.Header.Add("X-Tidepool-Session-Token", sessionToken)
					})

					It("returns unauthorized if multiple values", func() {
						res.HeaderOutput = &http.Header{}
						res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						req.Header.Add("X-Tidepool-Session-Token", sessionToken)
						middlewareFunc(res, req)
						Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					})

					It("returns successfully", func() {
						userID := serviceTest.NewUserID()
						authClient.ValidateSessionTokenOutputs = []authTest.ValidateSessionTokenOutput{
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
						Expect(authClient.ValidateSessionTokenInputs).To(Equal([]string{sessionToken}))
					})

					It("returns successfully as service", func() {
						authClient.ValidateSessionTokenOutputs = []authTest.ValidateSessionTokenOutput{
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
						Expect(authClient.ValidateSessionTokenInputs).To(Equal([]string{sessionToken}))
					})

					It("returns successfully with no details if session token is not valid", func() {
						authClient.ValidateSessionTokenOutputs = []authTest.ValidateSessionTokenOutput{{Details: nil, Error: errorsTest.RandomError()}}
						handlerFunc = func(res rest.ResponseWriter, req *rest.Request) {
							details := request.DetailsFromContext(req.Context())
							Expect(details).To(BeNil())
							Expect(service.GetRequestAuthDetails(req)).To(BeNil())
							Expect(log.LoggerFromContext(req.Context())).To(Equal(lgr))
							Expect(service.GetRequestLogger(req)).To(Equal(lgr))
						}
						middlewareFunc(res, req)
						Expect(authClient.ValidateSessionTokenInputs).To(Equal([]string{sessionToken}))
					})
				})

				Context("with restricted token", func() {
					var restrictedToken string

					BeforeEach(func() {
						restrictedToken = authTest.NewRestrictedToken()
						query := req.URL.Query()
						query.Add("restricted_token", restrictedToken)
						req.URL.RawQuery = query.Encode()
					})

					It("returns unauthorized if multiple values", func() {
						res.HeaderOutput = &http.Header{}
						res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						query := req.URL.Query()
						query.Add("restricted_token", restrictedToken)
						req.URL.RawQuery = query.Encode()
						middlewareFunc(res, req)
						Expect(res.WriteHeaderInputs).To(Equal([]int{403}))
					})

					It("returns successfully", func() {
						userID := serviceTest.NewUserID()
						restrictedTokenObject := &auth.RestrictedToken{
							ID:             restrictedToken,
							UserID:         userID,
							ExpirationTime: time.Now().Add(time.Hour),
						}
						authClient.GetRestrictedTokenOutputs = []authTest.GetRestrictedTokenOutput{{RestrictedToken: restrictedTokenObject, Error: nil}}
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
						authClient.GetRestrictedTokenOutputs = []authTest.GetRestrictedTokenOutput{{RestrictedToken: nil, Error: errorsTest.RandomError()}}
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
						authClient.GetRestrictedTokenOutputs = []authTest.GetRestrictedTokenOutput{{RestrictedToken: nil, Error: nil}}
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
						userID := serviceTest.NewUserID()
						restrictedTokenObject := &auth.RestrictedToken{
							ID:             restrictedToken,
							UserID:         userID,
							ExpirationTime: time.Now().Add(-time.Hour),
						}
						authClient.GetRestrictedTokenOutputs = []authTest.GetRestrictedTokenOutput{{RestrictedToken: restrictedTokenObject, Error: nil}}
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
