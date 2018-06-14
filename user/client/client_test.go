package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"context"
	"net/http"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
	"github.com/tidepool-org/platform/user"
	userClient "github.com/tidepool-org/platform/user/client"
)

var _ = Describe("Client", func() {
	var config *platform.Config
	var authorizeAs platform.AuthorizeAs

	BeforeEach(func() {
		config = platform.NewConfig()
		config.UserAgent = testHttp.NewUserAgent()
		authorizeAs = platform.AuthorizeAsService
	})

	Context("New", func() {
		BeforeEach(func() {
			config.Address = testHttp.NewAddress()
		})

		It("returns an error when the config is missing", func() {
			config = nil
			client, err := userClient.New(nil, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the authorize as is invalid", func() {
			authorizeAs = platform.AuthorizeAs(-1)
			client, err := userClient.New(config, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("authorize as is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns success", func() {
			Expect(userClient.New(config, authorizeAs)).ToNot(BeNil())
		})
	})

	Context("with server and new client", func() {
		var server *Server
		var requestHandlers []http.HandlerFunc
		var responseHeaders http.Header
		var client *userClient.Client
		var logger *logTest.Logger
		var sessionToken string
		var details request.Details
		var ctx context.Context

		BeforeEach(func() {
			server = NewServer()
			requestHandlers = nil
			responseHeaders = http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
			logger = logTest.NewLogger()
			sessionToken = authTest.NewSessionToken()
			details = request.NewDetails(request.MethodSessionToken, "", sessionToken)
			ctx = context.Background()
			ctx = log.NewContextWithLogger(ctx, logger)
			ctx = auth.NewContextWithServerSessionToken(ctx, sessionToken)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(CombineHandlers(requestHandlers...))
			var err error
			config.Address = server.URL()
			client, err = userClient.New(config, authorizeAs)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
			ctx = request.NewContextWithDetails(ctx, details)
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
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
					ctx = request.NewContextWithDetails(ctx, nil)
					errorsTest.ExpectEqual(client.EnsureAuthorizedService(ctx), request.ErrorUnauthorized())
				})

				It("returns an error when the details are for not a service", func() {
					ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, user.NewID(), sessionToken))
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
			var permission string

			BeforeEach(func() {
				requestUserID = user.NewID()
				targetUserID = user.NewID()
				details = request.NewDetails(request.MethodSessionToken, requestUserID, sessionToken)
				permission = test.RandomStringFromArray([]string{user.UploadPermission, user.ViewPermission})
			})

			Context("without server response", func() {
				AfterEach(func() {
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, permission)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(userID).To(BeEmpty())
				})

				It("returns an error when the target user id is missing", func() {
					targetUserID = ""
					userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, permission)
					errorsTest.ExpectEqual(err, errors.New("target user id is missing"))
					Expect(userID).To(BeEmpty())
				})

				It("returns an error when the permission is missing", func() {
					permission = ""
					userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, permission)
					errorsTest.ExpectEqual(err, errors.New("permission is missing"))
					Expect(userID).To(BeEmpty())
				})

				It("returns an error when the details are missing", func() {
					ctx = request.NewContextWithDetails(ctx, nil)
					userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, permission)
					errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
					Expect(userID).To(BeEmpty())
				})

				It("returns successfully when the details are for a service and permission is custodian", func() {
					ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, "", sessionToken))
					permission = user.CustodianPermission
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, permission)).To(Equal(""))
				})

				It("returns successfully when the details are for a service and permission is owner", func() {
					ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, "", sessionToken))
					permission = user.OwnerPermission
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, permission)).To(Equal(""))
				})

				It("returns successfully when the details are for a service and permission is upload", func() {
					ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, "", sessionToken))
					permission = user.UploadPermission
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, permission)).To(Equal(""))
				})

				It("returns successfully when the details are for a service and permission is view", func() {
					ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, "", sessionToken))
					permission = user.ViewPermission
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, permission)).To(Equal(""))
				})

				It("returns an error when the details are for the target user and permission is custodian", func() {
					ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, targetUserID, sessionToken))
					permission = user.CustodianPermission
					userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, permission)
					errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
					Expect(userID).To(BeEmpty())
				})

				It("returns successfully when the details are for the target user and permission is owner", func() {
					ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, targetUserID, sessionToken))
					permission = user.OwnerPermission
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, permission)).To(Equal(targetUserID))
				})

				It("returns successfully when the details are for the target user and permission is upload", func() {
					ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, targetUserID, sessionToken))
					permission = user.UploadPermission
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, permission)).To(Equal(targetUserID))
				})

				It("returns successfully when the details are for the target user and permission is view", func() {
					ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, targetUserID, sessionToken))
					permission = user.ViewPermission
					Expect(client.EnsureAuthorizedUser(ctx, targetUserID, permission)).To(Equal(targetUserID))
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
						userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, permission)
						errorsTest.ExpectEqual(err, errors.New("unable to get user permissions"))
						Expect(userID).To(Equal(""))
					})
				})

				Context("with a not found response, which is the same as unauthorized", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusNotFound, nil, responseHeaders))
					})

					It("returns an error", func() {
						userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, permission)
						errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
						Expect(userID).To(Equal(""))
					})
				})

				Context("with a successful response, but with no permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, "{}", responseHeaders))
					})

					It("returns an error", func() {
						userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, permission)
						errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
						Expect(userID).To(Equal(""))
					})
				})

				Context("with a successful response with incorrect permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, `{"view": {}}`, responseHeaders))
					})

					It("returns an error", func() {
						permission = user.UploadPermission
						userID, err := client.EnsureAuthorizedUser(ctx, targetUserID, permission)
						errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
						Expect(userID).To(Equal(""))
					})
				})

				Context("with a successful response with correct permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, `{"upload": {}, "view": {}}`, responseHeaders))
					})

					It("returns successfully with expected permissions", func() {
						permission = user.UploadPermission
						Expect(client.EnsureAuthorizedUser(ctx, targetUserID, permission)).To(Equal(requestUserID))
					})
				})
			})
		})

		Context("GetUserPermissions", func() {
			var requestUserID string
			var targetUserID string

			BeforeEach(func() {
				requestUserID = user.NewID()
				targetUserID = user.NewID()
			})

			Context("without server response", func() {
				AfterEach(func() {
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					permissions, err := client.GetUserPermissions(ctx, requestUserID, targetUserID)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(permissions).To(BeNil())
				})

				It("returns an error when the request user id is missing", func() {
					requestUserID = ""
					permissions, err := client.GetUserPermissions(ctx, requestUserID, targetUserID)
					errorsTest.ExpectEqual(err, errors.New("request user id is missing"))
					Expect(permissions).To(BeNil())
				})

				It("returns an error when the target user id is missing", func() {
					targetUserID = ""
					permissions, err := client.GetUserPermissions(ctx, requestUserID, targetUserID)
					errorsTest.ExpectEqual(err, errors.New("target user id is missing"))
					Expect(permissions).To(BeNil())
				})
			})

			Context("with server response", func() {
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
						permissions, err := client.GetUserPermissions(ctx, requestUserID, targetUserID)
						errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
						Expect(permissions).To(BeNil())
					})
				})

				Context("with a not found response, which is the same as unauthorized", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusNotFound, nil, responseHeaders))
					})

					It("returns an unauthorized error", func() {
						permissions, err := client.GetUserPermissions(ctx, requestUserID, targetUserID)
						errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
						Expect(permissions).To(BeNil())
					})
				})

				Context("with a successful response, but with no permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, "{}", responseHeaders))
					})

					It("returns successfully with expected permissions", func() {
						Expect(client.GetUserPermissions(ctx, requestUserID, targetUserID)).To(BeEmpty())
					})
				})

				Context("with a successful response with upload and view permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, `{"upload": {}, "view": {}}`, responseHeaders))
					})

					It("returns successfully with expected permissions", func() {
						Expect(client.GetUserPermissions(ctx, requestUserID, targetUserID)).To(Equal(user.Permissions{
							user.UploadPermission: user.Permission{},
							user.ViewPermission:   user.Permission{},
						}))
					})
				})

				Context("with a successful response with owner permissions that already includes upload permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, `{"root": {"root-inner": "unused"}, "upload": {}}`, responseHeaders))
					})

					It("returns successfully with expected permissions", func() {
						Expect(client.GetUserPermissions(ctx, requestUserID, targetUserID)).To(Equal(user.Permissions{
							user.OwnerPermission:  user.Permission{"root-inner": "unused"},
							user.UploadPermission: user.Permission{},
							user.ViewPermission:   user.Permission{"root-inner": "unused"},
						}))
					})
				})

				Context("with a successful response with owner permissions that already includes view permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, `{"root": {"root-inner": "unused"}, "view": {}}`, responseHeaders))
					})

					It("returns successfully with expected permissions", func() {
						Expect(client.GetUserPermissions(ctx, requestUserID, targetUserID)).To(Equal(user.Permissions{
							user.OwnerPermission:  user.Permission{"root-inner": "unused"},
							user.UploadPermission: user.Permission{"root-inner": "unused"},
							user.ViewPermission:   user.Permission{},
						}))
					})
				})

				Context("with a successful response with owner permissions that already includes upload and view permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, `{"root": {"root-inner": "unused"}, "upload": {}, "view": {}}`, responseHeaders))
					})

					It("returns successfully with expected permissions", func() {
						Expect(client.GetUserPermissions(ctx, requestUserID, targetUserID)).To(Equal(user.Permissions{
							user.OwnerPermission:  user.Permission{"root-inner": "unused"},
							user.UploadPermission: user.Permission{},
							user.ViewPermission:   user.Permission{},
						}))
					})
				})
			})
		})
	})
})
