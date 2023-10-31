package client_test

import (
	"context"
	"net/http"

	"github.com/mdblp/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	permissionClient "github.com/tidepool-org/platform/permission/client"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	testHttp "github.com/tidepool-org/platform/test/http"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Client", func() {
	var config *platform.Config

	BeforeEach(func() {
		config = platform.NewConfig()
		config.UserAgent = testHttp.NewUserAgent()
	})

	Context("New", func() {
		BeforeEach(func() {
			config.Address = testHttp.NewAddress()
		})

		It("returns success", func() {
			Expect(permissionClient.New()).ToNot(BeNil())
		})
	})

	Context("with server", func() {
		var server *Server
		var requestHandlers []http.HandlerFunc
		var logger *logTest.Logger
		var sessionToken string
		var ctx context.Context
		var client *permissionClient.Client
		var req *rest.Request

		BeforeEach(func() {
			server = NewServer()
			requestHandlers = nil
			logger = logTest.NewLogger()
			sessionToken = authTest.NewSessionToken()
			ctx = context.Background()
			ctx = log.NewContextWithLogger(ctx, logger)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(CombineHandlers(requestHandlers...))
			config.Address = server.URL()
			client = permissionClient.New()
			Expect(client).ToNot(BeNil())
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
				data := request.NewDetails(request.MethodSessionToken, requestUserID, sessionToken, "patient")
				ctx = request.NewContextWithDetails(ctx, data)
				httpReq, _ := http.NewRequestWithContext(ctx, "GET", "http://test.fr", nil)
				req = &rest.Request{
					Request: httpReq,
				}

			})

			Context("without server response", func() {
				AfterEach(func() {
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns an error when the target user id is missing", func() {
					targetUserID = ""
					permissions, err := client.GetUserPermissions(req, targetUserID)
					errorsTest.ExpectEqual(err, errors.New("target user id is missing"))
					Expect(permissions).To(Equal(false))
				})

				It("returns successfully when request is target with expected accepted authorization without calling authorization service", func() {
					permissions, err := client.GetUserPermissions(req, requestUserID)
					Expect(err).To(BeNil())
					Expect(permissions).To(Equal(true))
				})

				It("returns successfully when the requester is a service", func() {
					// Service don't have userId set
					data := request.NewDetails(request.MethodSessionToken, "", sessionToken, "server")
					ctx = request.NewContextWithDetails(ctx, data)
					httpReq, _ := http.NewRequestWithContext(ctx, "GET", "http://test.fr", nil)
					req = &rest.Request{
						Request: httpReq,
					}
					permissions, err := client.GetUserPermissions(req, targetUserID)
					Expect(err).To(BeNil())
					Expect(permissions).To(Equal(true))
				})

				It("returns an error when the requester is not a patient", func() {
					data := request.NewDetails(request.MethodSessionToken, requestUserID, sessionToken, "hcp")
					ctx = request.NewContextWithDetails(ctx, data)
					httpReq, _ := http.NewRequestWithContext(ctx, "GET", "http://test.fr", nil)
					req = &rest.Request{
						Request: httpReq,
					}
					permissions, err := client.GetUserPermissions(req, targetUserID)
					Expect(err).To(BeNil())
					Expect(permissions).To(Equal(false))
				})

				It("returns an error when the requester is not the same UserId", func() {
					data := request.NewDetails(request.MethodSessionToken, requestUserID, sessionToken, "patient")
					ctx = request.NewContextWithDetails(ctx, data)
					httpReq, _ := http.NewRequestWithContext(ctx, "GET", "http://test.fr", nil)
					req = &rest.Request{
						Request: httpReq,
					}
					permissions, err := client.GetUserPermissions(req, targetUserID)
					Expect(err).To(BeNil())
					Expect(permissions).To(Equal(false))
				})
			})

		})
	})
})
