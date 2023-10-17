package client_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/permission"
	permissionClient "github.com/tidepool-org/platform/permission/client"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	testHttp "github.com/tidepool-org/platform/test/http"
	userTest "github.com/tidepool-org/platform/user/test"
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
			client, err := permissionClient.New(nil, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the authorize as is invalid", func() {
			authorizeAs = platform.AuthorizeAs(-1)
			client, err := permissionClient.New(config, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("authorize as is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns success", func() {
			Expect(permissionClient.New(config, authorizeAs)).ToNot(BeNil())
		})
	})

	Context("with server and new client", func() {
		var server *Server
		var requestHandlers []http.HandlerFunc
		var responseHeaders http.Header
		var logger *logTest.Logger
		var sessionToken string
		var details request.AuthDetails
		var ctx context.Context
		var client *permissionClient.Client

		BeforeEach(func() {
			server = NewServer()
			requestHandlers = nil
			responseHeaders = http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
			logger = logTest.NewLogger()
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
			client, err = permissionClient.New(config, authorizeAs)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
			ctx = request.NewContextWithAuthDetails(ctx, details)
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("GetUserPermissions", func() {
			var requestUserID string
			var targetUserID string

			BeforeEach(func() {
				requestUserID = userTest.RandomID()
				targetUserID = userTest.RandomID()
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
						Expect(client.GetUserPermissions(ctx, requestUserID, targetUserID)).To(Equal(permission.Permissions{
							permission.Write: permission.Permission{},
							permission.Read:  permission.Permission{},
						}))
					})
				})

				Context("with a successful response with owner permissions that already includes upload permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, `{"root": {"root-inner": "unused"}, "upload": {}}`, responseHeaders))
					})

					It("returns successfully with expected permissions", func() {
						Expect(client.GetUserPermissions(ctx, requestUserID, targetUserID)).To(Equal(permission.Permissions{
							permission.Owner: permission.Permission{"root-inner": "unused"},
							permission.Write: permission.Permission{},
							permission.Read:  permission.Permission{"root-inner": "unused"},
						}))
					})
				})

				Context("with a successful response with owner permissions that already includes view permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, `{"root": {"root-inner": "unused"}, "view": {}}`, responseHeaders))
					})

					It("returns successfully with expected permissions", func() {
						Expect(client.GetUserPermissions(ctx, requestUserID, targetUserID)).To(Equal(permission.Permissions{
							permission.Owner: permission.Permission{"root-inner": "unused"},
							permission.Write: permission.Permission{"root-inner": "unused"},
							permission.Read:  permission.Permission{},
						}))
					})
				})

				Context("with a successful response with owner permissions that already includes upload and view permissions", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, `{"root": {"root-inner": "unused"}, "upload": {}, "view": {}}`, responseHeaders))
					})

					It("returns successfully with expected permissions", func() {
						Expect(client.GetUserPermissions(ctx, requestUserID, targetUserID)).To(Equal(permission.Permissions{
							permission.Owner: permission.Permission{"root-inner": "unused"},
							permission.Write: permission.Permission{},
							permission.Read:  permission.Permission{},
						}))
					})
				})
			})
		})
	})
})
