package client_test

import (
	"context"
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	permissionClient "github.com/tidepool-org/platform/permission/client"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
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

		It("returns an error when the config is missing", func() {
			config = nil
			client, err := permissionClient.New(nil)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns success when the config is present", func() {
			Expect(permissionClient.New(config)).ToNot(BeNil())
		})
	})

	Context("with server and coastguard client", func() {
		var server *Server
		var requestHandlers []http.HandlerFunc
		var responseHeaders http.Header
		var logger *logTest.Logger
		var sessionToken string
		var ctx context.Context
		var client *permissionClient.Client
		var req *rest.Request

		BeforeEach(func() {
			server = NewServer()
			requestHandlers = nil
			responseHeaders = http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
			logger = logTest.NewLogger()
			sessionToken = authTest.NewSessionToken()
			ctx = context.Background()
			ctx = log.NewContextWithLogger(ctx, logger)
			ctx = auth.NewContextWithServerSessionToken(ctx, sessionToken)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(CombineHandlers(requestHandlers...))
			var err error
			config.Address = server.URL()
			client, err = permissionClient.New(config)
			Expect(err).ToNot(HaveOccurred())
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
				data := request.NewDetails(request.MethodSessionToken, requestUserID, sessionToken)
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
					data := request.NewDetails(request.MethodSessionToken, "", sessionToken)
					ctx = request.NewContextWithDetails(ctx, data)
					httpReq, _ := http.NewRequestWithContext(ctx, "GET", "http://test.fr", nil)
					req = &rest.Request{
						Request: httpReq,
					}
					permissions, err := client.GetUserPermissions(req, targetUserID)
					Expect(err).To(BeNil())
					Expect(permissions).To(Equal(true))
				})
			})

			Context("with server response", func() {

				BeforeEach(func() {
					var requestBody permissionClient.CoastguardRequestBody
					url := *req.URL
					headers := make(map[string]string)
					for k := range req.Header {
						headers[strings.ToLower(k)] = req.Header.Get(k)
					}
					requestBody.Input.Request.Headers = headers
					requestBody.Input.Request.Method = req.Method
					requestBody.Input.Request.Protocol = req.Proto
					requestBody.Input.Request.Host = req.Host
					requestBody.Input.Request.Path = url.Path
					requestBody.Input.Request.Query = url.RawQuery
					requestBody.Input.Request.Service = "platform"
					requestBody.Input.Data.TargetUserID = targetUserID
					requestHandlers = append(requestHandlers,
						VerifyContentType("application/json; charset=utf-8"),
						VerifyHeaderKV("X-Tidepool-Session-Token", sessionToken),
						VerifyBody(test.MarshalRequestBody(&requestBody)),
						VerifyRequest("POST", "/v1/data/backloops/access"),
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
						permissions, err := client.GetUserPermissions(req, targetUserID)
						Expect(err).NotTo(BeNil())
						Expect(permissions).To(Equal(false))
					})
				})

				Context("with a not found response ", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusNotFound, nil, responseHeaders))
					})

					It("returns an error", func() {
						permissions, err := client.GetUserPermissions(req, targetUserID)
						Expect(err).NotTo(BeNil())
						Expect(permissions).To(Equal(false))
					})
				})

				Context("with a successful response, but with empty response", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, "{}", responseHeaders))
					})

					It("returns successfully with expected refused authorization", func() {
						permissions, err := client.GetUserPermissions(req, targetUserID)
						Expect(err).To(BeNil())
						Expect(permissions).To(Equal(false))
					})
				})

				Context("with a successful response with authorization set to false", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, `{"result":{"authorized": false, "route": "test"}}`, responseHeaders))
					})

					It("returns successfully with expected refused authorization", func() {
						permissions, err := client.GetUserPermissions(req, targetUserID)
						Expect(err).To(BeNil())
						Expect(permissions).To(Equal(false))
					})
				})

				Context("with a successful response with authorization set to true", func() {
					BeforeEach(func() {
						requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, `{"result":{"authorized": true, "route": "test"}}`, responseHeaders))
					})

					It("returns successfully with expected accepted authorization", func() {
						permissions, err := client.GetUserPermissions(req, targetUserID)
						Expect(err).To(BeNil())
						Expect(permissions).To(Equal(true))
					})
				})
			})
		})
	})
})
