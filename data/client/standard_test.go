package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"net/http"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	userClient "github.com/tidepool-org/platform/user/client"
)

type ServerTokenOutput struct {
	serverToken string
	err         error
}

type TestUserClient struct {
	ServerTokenOutputs []ServerTokenOutput
}

func (t *TestUserClient) UserID() string {
	panic("Unexpected invocation of UserID on TestUserClient")
}

func (t *TestUserClient) ValidateAuthenticationToken(context userClient.Context, authenticationToken string) (userClient.AuthenticationDetails, error) {
	panic("Unexpected invocation of ValidateAuthenticationToken on TestUserClient")
}

func (t *TestUserClient) GetUserPermissions(context userClient.Context, requestUserID string, targetUserID string) (userClient.Permissions, error) {
	panic("Unexpected invocation of GetUserPermissions on TestUserClient")
}

func (t *TestUserClient) ServerToken() (string, error) {
	output := t.ServerTokenOutputs[0]
	t.ServerTokenOutputs = t.ServerTokenOutputs[1:]
	return output.serverToken, output.err
}

func (t *TestUserClient) ValidateTest() bool {
	return len(t.ServerTokenOutputs) == 0
}

type TestContext struct {
	TestLogger     log.Logger
	TestRequest    *rest.Request
	TestUserClient *TestUserClient
}

func (t *TestContext) Logger() log.Logger {
	return t.TestLogger
}

func (t *TestContext) Request() *rest.Request {
	return t.TestRequest
}

func (t *TestContext) UserClient() userClient.Client {
	return t.TestUserClient
}

func (t *TestContext) ValidateTest() bool {
	return t.TestUserClient.ValidateTest()
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
				TestLogger:     null.NewLogger(),
				TestRequest:    &rest.Request{},
				TestUserClient: &TestUserClient{},
			}
			context.TestUserClient.ServerTokenOutputs = []ServerTokenOutput{{"test-server-token", nil}}
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("DestroyDataForUserByID", func() {
			It("returns error if context is missing", func() {
				context.TestUserClient.ServerTokenOutputs = nil
				Expect(standard.DestroyDataForUserByID(nil, "test-user-id")).To(MatchError("client: context is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
				Expect(context.ValidateTest()).To(BeTrue())
			})

			It("returns error if user id is missing", func() {
				context.TestUserClient.ServerTokenOutputs = nil
				Expect(standard.DestroyDataForUserByID(context, "")).To(MatchError("client: user id is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
				Expect(context.ValidateTest()).To(BeTrue())
			})

			It("returns error if the context request is missing", func() {
				context.TestRequest = nil
				context.TestUserClient.ServerTokenOutputs = nil
				Expect(standard.DestroyDataForUserByID(context, "test-user-id")).To(MatchError("client: unable to copy request trace; service: source request is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
				Expect(context.ValidateTest()).To(BeTrue())
			})

			It("returns error if the user client server token returns an error", func() {
				err := errors.New("test-error")
				context.TestUserClient.ServerTokenOutputs = []ServerTokenOutput{{"", err}}
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
							ghttp.VerifyRequest("DELETE", "/dataservice/v1/users/test-user-id/data"),
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
							ghttp.VerifyRequest("DELETE", "/dataservice/v1/users/test-user-id/data"),
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
							ghttp.VerifyRequest("DELETE", "/dataservice/v1/users/test-user-id/data"),
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
