package client_test

import (
	"errors"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/dataservices/client"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
)

type ServerTokenOutput struct {
	serverToken string
	err         error
}

type TestUserServicesClient struct {
	ServerTokenOutputs []ServerTokenOutput
}

func (t *TestUserServicesClient) UserID() string {
	panic("Unexpected invocation of UserID on TestUserServicesClient")
}

func (t *TestUserServicesClient) ValidateAuthenticationToken(context service.Context, authenticationToken string) (userservicesClient.AuthenticationDetails, error) {
	panic("Unexpected invocation of ValidateAuthenticationToken on TestUserServicesClient")
}

func (t *TestUserServicesClient) GetUserPermissions(context service.Context, requestUserID string, targetUserID string) (userservicesClient.Permissions, error) {
	panic("Unexpected invocation of GetUserPermissions on TestUserServicesClient")
}

func (t *TestUserServicesClient) ServerToken() (string, error) {
	output := t.ServerTokenOutputs[0]
	t.ServerTokenOutputs = t.ServerTokenOutputs[1:]
	return output.serverToken, output.err
}

func (t *TestUserServicesClient) ValidateTest() bool {
	return len(t.ServerTokenOutputs) == 0
}

type TestContext struct {
	TestLogger             log.Logger
	TestRequest            *rest.Request
	TestUserServicesClient *TestUserServicesClient
}

func (t *TestContext) Logger() log.Logger {
	return t.TestLogger
}

func (t *TestContext) Request() *rest.Request {
	return t.TestRequest
}

func (t *TestContext) UserServicesClient() userservicesClient.Client {
	return t.TestUserServicesClient
}

func (t *TestContext) ValidateTest() bool {
	return t.TestUserServicesClient.ValidateTest()
}

var _ = Describe("Standard", func() {
	Context("NewStandard", func() {
		var config *client.Config

		BeforeEach(func() {
			config = &client.Config{
				Address: "http://localhost:1234",
				Timeout: 30 * time.Second,
			}
		})

		It("returns an error if config is missing", func() {
			standard, err := client.NewStandard(nil)
			Expect(err).To(MatchError("client: config is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config address is missing", func() {
			config.Address = ""
			standard, err := client.NewStandard(config)
			Expect(err).To(MatchError("client: config is invalid; client: address is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config timeout is invalid", func() {
			config.Timeout = 0
			standard, err := client.NewStandard(config)
			Expect(err).To(MatchError("client: config is invalid; client: timeout is invalid"))
			Expect(standard).To(BeNil())
		})

		It("returns success", func() {
			standard, err := client.NewStandard(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(standard).ToNot(BeNil())
		})
	})

	Context("with server", func() {
		var server *ghttp.Server
		var config *client.Config
		var standard *client.Standard
		var context *TestContext

		BeforeEach(func() {
			server = ghttp.NewServer()
			config = &client.Config{
				Address: server.URL(),
				Timeout: 30 * time.Second,
			}
			var err error
			standard, err = client.NewStandard(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(standard).ToNot(BeNil())
			context = &TestContext{
				TestLogger:             log.NewNull(),
				TestRequest:            &rest.Request{},
				TestUserServicesClient: &TestUserServicesClient{},
			}
			context.TestUserServicesClient.ServerTokenOutputs = []ServerTokenOutput{{"test-server-token", nil}}
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("DestroyDataForUserByID", func() {
			It("returns error if context is missing", func() {
				context.TestUserServicesClient.ServerTokenOutputs = nil
				Expect(standard.DestroyDataForUserByID(nil, "test-user-id")).To(MatchError("client: context is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
				Expect(context.ValidateTest()).To(BeTrue())
			})

			It("returns error if user id is missing", func() {
				context.TestUserServicesClient.ServerTokenOutputs = nil
				Expect(standard.DestroyDataForUserByID(context, "")).To(MatchError("client: user id is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
				Expect(context.ValidateTest()).To(BeTrue())
			})

			It("returns error if the context request is missing", func() {
				context.TestRequest = nil
				context.TestUserServicesClient.ServerTokenOutputs = nil
				Expect(standard.DestroyDataForUserByID(context, "test-user-id")).To(MatchError("client: unable to copy request trace; service: source request is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
				Expect(context.ValidateTest()).To(BeTrue())
			})

			It("returns error if the user services client server token returns an error", func() {
				err := errors.New("test-error")
				context.TestUserServicesClient.ServerTokenOutputs = []ServerTokenOutput{{"", err}}
				Expect(standard.DestroyDataForUserByID(context, "test-user-id")).To(Equal(err))
				Expect(server.ReceivedRequests()).To(BeEmpty())
				Expect(context.ValidateTest()).To(BeTrue())
			})

			It("returns error if the server is not reachable", func() {
				server.Close()
				server = nil
				err := standard.DestroyDataForUserByID(context, "test-user-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("client: unable to perform request DELETE "))
				Expect(context.ValidateTest()).To(BeTrue())
			})

			Context("with an unexpected response", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("DELETE", "/dataservices/v1/users/test-user-id/data"),
							ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-server-token"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
					)
				})

				It("returns an error", func() {
					err := standard.DestroyDataForUserByID(context, "test-user-id")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("client: unexpected response status code 400 from DELETE "))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
					Expect(context.ValidateTest()).To(BeTrue())
				})
			})

			Context("with an unauthorized response", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("DELETE", "/dataservices/v1/users/test-user-id/data"),
							ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-server-token"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusUnauthorized, nil, nil)),
					)
				})

				It("returns an error", func() {
					err := standard.DestroyDataForUserByID(context, "test-user-id")
					Expect(err).To(MatchError("client: unauthorized"))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
					Expect(context.ValidateTest()).To(BeTrue())
				})
			})

			Context("with a successful response", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("DELETE", "/dataservices/v1/users/test-user-id/data"),
							ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-server-token"),
							ghttp.VerifyBody([]byte{}),
							ghttp.RespondWith(http.StatusOK, nil, nil)),
					)
				})

				It("returns success", func() {
					Expect(standard.DestroyDataForUserByID(context, "test-user-id")).To(Succeed())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
					Expect(context.ValidateTest()).To(BeTrue())
				})
			})
		})
	})
})
