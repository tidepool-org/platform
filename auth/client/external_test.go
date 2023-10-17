package client_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/auth"
	authClient "github.com/tidepool-org/platform/auth/client"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("External", func() {
	var config *authClient.ExternalConfig
	var authorizeAs platform.AuthorizeAs
	var name string
	var logger *logTest.Logger

	BeforeEach(func() {
		config = authClient.NewExternalConfig()
		config.UserAgent = testHttp.NewUserAgent()
		config.ServerSessionTokenSecret = authTest.NewServiceSecret()
		authorizeAs = platform.AuthorizeAsService
		name = test.RandomString()
		logger = logTest.NewLogger()
	})

	Context("NewExternal", func() {
		BeforeEach(func() {
			config.Address = testHttp.NewAddress()
		})

		It("returns an error when the config is missing", func() {
			config = nil
			client, err := authClient.NewExternal(config, authorizeAs, name, logger)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the authorize as is invalid", func() {
			authorizeAs = platform.AuthorizeAs(-1)
			client, err := authClient.NewExternal(config, authorizeAs, name, logger)
			errorsTest.ExpectEqual(err, errors.New("authorize as is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the name is missing", func() {
			name = ""
			client, err := authClient.NewExternal(config, authorizeAs, name, logger)
			errorsTest.ExpectEqual(err, errors.New("name is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the logger is missing", func() {
			logger = nil
			client, err := authClient.NewExternal(config, authorizeAs, name, nil)
			errorsTest.ExpectEqual(err, errors.New("logger is missing"))
			Expect(client).To(BeNil())
		})

		It("returns success", func() {
			Expect(authClient.NewExternal(config, authorizeAs, name, logger)).ToNot(BeNil())
		})
	})

	Context("with server and new client", func() {
		var server *Server
		var requestHandlers []http.HandlerFunc
		var responseHeaders http.Header
		var client *authClient.External
		var sessionToken string
		var details request.AuthDetails
		var ctx context.Context

		BeforeEach(func() {
			server = NewServer()
			requestHandlers = nil
			responseHeaders = http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
			sessionToken = authTest.NewSessionToken()
			details = request.NewAuthDetails(request.MethodSessionToken, "", sessionToken)
			ctx = context.Background()
			ctx = log.NewContextWithLogger(ctx, logger)
			ctx = auth.NewContextWithServerSessionToken(ctx, sessionToken)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(CombineHandlers(requestHandlers...))
			var err error
			config.Address = server.URL()
			client, err = authClient.NewExternal(config, authorizeAs, name, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
			ctx = request.NewContextWithAuthDetails(ctx, details)
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("EnsureAuthorized", func() {
			Context("without server response", func() {
				AfterEach(func() {
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					errorsTest.ExpectEqual(client.EnsureAuthorized(ctx), errors.New("context is missing"))
				})

				It("returns an error when the details are missing", func() {
					ctx = request.NewContextWithAuthDetails(ctx, nil)
					errorsTest.ExpectEqual(client.EnsureAuthorized(ctx), request.ErrorUnauthorized())
				})

				It("returns successfully when the details are for a user", func() {
					ctx = request.NewContextWithAuthDetails(ctx, request.NewAuthDetails(request.MethodSessionToken, authTest.RandomUserID(), sessionToken))
					Expect(client.EnsureAuthorized(ctx)).To(Succeed())
				})

				It("returns successfully when the details are for a service", func() {
					Expect(client.EnsureAuthorized(ctx)).To(Succeed())
				})
			})
		})

		Context("EnsureAuthorizedService", func() {
			Context("without server response", func() {
				AfterEach(func() {
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					errorsTest.ExpectEqual(client.EnsureAuthorizedService(ctx), errors.New("context is missing"))
				})

				It("returns an error when the details are missing", func() {
					ctx = request.NewContextWithAuthDetails(ctx, nil)
					errorsTest.ExpectEqual(client.EnsureAuthorizedService(ctx), request.ErrorUnauthorized())
				})

				It("returns an error when the details are for not a service", func() {
					ctx = request.NewContextWithAuthDetails(ctx, request.NewAuthDetails(request.MethodSessionToken, authTest.RandomUserID(), sessionToken))
					errorsTest.ExpectEqual(client.EnsureAuthorizedService(ctx), request.ErrorUnauthorized())
				})

				It("returns successfully when the details are for a service", func() {
					Expect(client.EnsureAuthorizedService(ctx)).To(Succeed())
				})
			})
		})

		Context("EnsureAuthorizedUser", func() {
			var requestUserID string
			var targetUserID string
			var authorizedPermission string

			BeforeEach(func() {
				requestUserID = authTest.RandomUserID()
				targetUserID = authTest.RandomUserID()
				details = request.NewAuthDetails(request.MethodSessionToken, requestUserID, sessionToken)
				authorizedPermission = test.RandomStringFromArray([]string{permission.Write, permission.Read})
			})

			Context("without server response", func() {
				AfterEach(func() {
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(userID).To(BeEmpty())
				})

				It("returns an error when the target user id is missing", func() {
					targetUserID = ""
					userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)
					errorsTest.ExpectEqual(err, errors.New("target user id is missing"))
					Expect(userID).To(BeEmpty())
				})

				It("returns an error when the authorized permission is missing", func() {
					authorizedPermission = ""
					userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)
					errorsTest.ExpectEqual(err, errors.New("authorized permission is missing"))
					Expect(userID).To(BeEmpty())
				})

				It("returns an error when the details are missing", func() {
					ctx = request.NewContextWithAuthDetails(ctx, nil)
					userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)
					errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
					Expect(userID).To(BeEmpty())
				})

				It("returns successfully when the details are for a service and authorized permission is custodian", func() {
					ctx = request.NewContextWithAuthDetails(ctx, request.NewAuthDetails(request.MethodSessionToken, "", sessionToken))
					authorizedPermission = permission.Custodian
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)).To(Equal(""))
				})

				It("returns successfully when the details are for a service and authorized permission is owner", func() {
					ctx = request.NewContextWithAuthDetails(ctx, request.NewAuthDetails(request.MethodSessionToken, "", sessionToken))
					authorizedPermission = permission.Owner
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)).To(Equal(""))
				})

				It("returns successfully when the details are for a service and authorized permission is upload", func() {
					ctx = request.NewContextWithAuthDetails(ctx, request.NewAuthDetails(request.MethodSessionToken, "", sessionToken))
					authorizedPermission = permission.Write
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)).To(Equal(""))
				})

				It("returns successfully when the details are for a service and authorized permission is view", func() {
					ctx = request.NewContextWithAuthDetails(ctx, request.NewAuthDetails(request.MethodSessionToken, "", sessionToken))
					authorizedPermission = permission.Read
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)).To(Equal(""))
				})

				It("returns an error when the details are for the target user and authorized permission is custodian", func() {
					ctx = request.NewContextWithAuthDetails(ctx, request.NewAuthDetails(request.MethodSessionToken, targetUserID, sessionToken))
					authorizedPermission = permission.Custodian
					userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)
					errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
					Expect(userID).To(BeEmpty())
				})

				It("returns successfully when the details are for the target user and authorized permission is owner", func() {
					ctx = request.NewContextWithAuthDetails(ctx, request.NewAuthDetails(request.MethodSessionToken, targetUserID, sessionToken))
					authorizedPermission = permission.Owner
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)).To(Equal(targetUserID))
				})

				It("returns successfully when the details are for the target user and authorized permission is upload", func() {
					ctx = request.NewContextWithAuthDetails(ctx, request.NewAuthDetails(request.MethodSessionToken, targetUserID, sessionToken))
					authorizedPermission = permission.Write
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)).To(Equal(targetUserID))
				})

				It("returns successfully when the details are for the target user and authorized permission is view", func() {
					ctx = request.NewContextWithAuthDetails(ctx, request.NewAuthDetails(request.MethodSessionToken, targetUserID, sessionToken))
					authorizedPermission = permission.Read
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)).To(Equal(targetUserID))
				})
			})

			Context("with server response when the details are not for the target user", func() {
				BeforeEach(func() {
					requestHandlers = append(requestHandlers,
						VerifyContentType(""),
						VerifyHeaderKV("X-Tidepool-Session-Token", sessionToken),
						VerifyBody(nil),
						VerifyRequest("GET", "/access/"+targetUserID+"/"+requestUserID),
					)
				})

				AfterEach(func() {
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				Context("with an unauthenticated response", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusUnauthorized, nil, responseHeaders))
					})

					It("returns an error", func() {
						userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)
						errorsTest.ExpectEqual(err, errors.New("unable to get user permissions"))
						Expect(userID).To(Equal(""))
					})
				})

				Context("with a not found response, which is the same as unauthorized", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusNotFound, nil, responseHeaders))
					})

					It("returns an error", func() {
						userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)
						errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
						Expect(userID).To(Equal(""))
					})
				})

				Context("with a successful response, but with no permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, "{}", responseHeaders))
					})

					It("returns an error", func() {
						userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)
						errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
						Expect(userID).To(Equal(""))
					})
				})

				Context("with a successful response with incorrect permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, `{"view": {}}`, responseHeaders))
					})

					It("returns an error", func() {
						authorizedPermission = permission.Write
						userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)
						errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
						Expect(userID).To(Equal(""))
					})
				})

				Context("with a successful response with correct permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, `{"upload": {}, "view": {}}`, responseHeaders))
					})

					It("returns successfully with expected permissions", func() {
						authorizedPermission = permission.Write
						Expect(client.EnsureAuthorizedUser(ctx, targetUserID, authorizedPermission)).To(Equal(requestUserID))
					})
				})
			})
		})
	})
})
