package client_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metricservices/client"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
)

type TestAuthenticationDetails struct {
	TokenOutputs    []string
	IsServerOutputs []bool
}

func (t *TestAuthenticationDetails) Token() string {
	output := t.TokenOutputs[0]
	t.TokenOutputs = t.TokenOutputs[1:]
	return output
}

func (t *TestAuthenticationDetails) IsServer() bool {
	output := t.IsServerOutputs[0]
	t.IsServerOutputs = t.IsServerOutputs[1:]
	return output
}

func (t *TestAuthenticationDetails) UserID() string {
	panic("Unexpected invocation of UserID on TestAuthenticationDetails")
}

func (t *TestAuthenticationDetails) ValidateTest() bool {
	return len(t.TokenOutputs) == 0 && len(t.IsServerOutputs) == 0
}

type TestContext struct {
	TestLogger                log.Logger
	TestRequest               *rest.Request
	TestAuthenticationDetails *TestAuthenticationDetails
}

func (t *TestContext) Logger() log.Logger {
	return t.TestLogger
}

func (t *TestContext) Request() *rest.Request {
	return t.TestRequest
}

func (t *TestContext) AuthenticationDetails() userservicesClient.AuthenticationDetails {
	return t.TestAuthenticationDetails
}

func (t *TestContext) ValidateTest() bool {
	return t.TestAuthenticationDetails.ValidateTest()
}

var _ = Describe("Standard", func() {
	var logger log.Logger
	var context *TestContext

	BeforeEach(func() {
		logger = log.NewNull()
		context = &TestContext{
			TestLogger:                logger,
			TestRequest:               &rest.Request{},
			TestAuthenticationDetails: &TestAuthenticationDetails{},
		}
	})

	Context("NewStandard", func() {
		var config *client.Config

		BeforeEach(func() {
			config = &client.Config{
				Address:        "http://localhost:1234",
				RequestTimeout: 30,
			}
		})

		It("returns an error if logger is missing", func() {
			standard, err := client.NewStandard(nil, "testservices", config)
			Expect(err).To(MatchError("client: logger is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if name is missing", func() {
			standard, err := client.NewStandard(logger, "", config)
			Expect(err).To(MatchError("client: name is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config is missing", func() {
			standard, err := client.NewStandard(logger, "testservices", nil)
			Expect(err).To(MatchError("client: config is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config address is invalid", func() {
			config.Address = ""
			standard, err := client.NewStandard(logger, "testservices", config)
			Expect(err).To(MatchError("client: config is invalid; client: address is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config request timeout is invalid", func() {
			config.RequestTimeout = -1
			standard, err := client.NewStandard(logger, "testservices", config)
			Expect(err).To(MatchError("client: config is invalid; client: request timeout is invalid"))
			Expect(standard).To(BeNil())
		})

		It("returns success", func() {
			standard, err := client.NewStandard(logger, "testservices", config)
			Expect(err).ToNot(HaveOccurred())
			Expect(standard).ToNot(BeNil())
		})
	})

	Context("with server", func() {
		var server *ghttp.Server
		var config *client.Config
		var standard *client.Standard
		var data map[string]string

		BeforeEach(func() {
			server = ghttp.NewServer()
			config = &client.Config{
				Address:        server.URL(),
				RequestTimeout: 30,
			}
			data = map[string]string{
				"left":  "handed",
				"right": "correct",
			}
		})

		JustBeforeEach(func() {
			var err error
			standard, err = client.NewStandard(logger, "testservices", config)
			Expect(err).ToNot(HaveOccurred())
			Expect(standard).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("RecordMetric", func() {
			It("returns error if context is missing", func() {
				Expect(standard.RecordMetric(nil, "test-metric", data)).To(MatchError("client: context is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
				Expect(context.ValidateTest()).To(BeTrue())
			})

			It("returns error if metric is missing", func() {
				Expect(standard.RecordMetric(context, "", data)).To(MatchError("client: metric is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
				Expect(context.ValidateTest()).To(BeTrue())
			})

			It("returns error if the context request is missing", func() {
				context.TestAuthenticationDetails.IsServerOutputs = []bool{false}
				context.TestRequest = nil
				err := standard.RecordMetric(context, "test-metric", data)
				Expect(err).To(MatchError("client: unable to copy request trace; service: source request is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
				Expect(context.ValidateTest()).To(BeTrue())
			})

			It("returns error if the server is not reachable", func() {
				context.TestAuthenticationDetails.TokenOutputs = []string{"test-authentication-token"}
				context.TestAuthenticationDetails.IsServerOutputs = []bool{false}
				server.Close()
				server = nil
				err := standard.RecordMetric(context, "test-metric", data)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("client: unable to perform request GET "))
				Expect(context.ValidateTest()).To(BeTrue())
			})

			Context("as user", func() {
				BeforeEach(func() {
					context.TestAuthenticationDetails.TokenOutputs = []string{"test-authentication-token"}
					context.TestAuthenticationDetails.IsServerOutputs = []bool{false}
				})

				Context("with an unexpected response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/metrics/thisuser/test-metric", "left=handed&right=correct"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := standard.RecordMetric(context, "test-metric", data)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(HavePrefix("client: unexpected response status code 400 from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(context.ValidateTest()).To(BeTrue())
					})
				})

				Context("with an unauthorized response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/metrics/thisuser/test-metric", "left=handed&right=correct"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusUnauthorized, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := standard.RecordMetric(context, "test-metric", data)
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(context.ValidateTest()).To(BeTrue())
					})
				})

				Context("with a successful response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/metrics/thisuser/test-metric", "left=handed&right=correct"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, nil, nil)),
						)
					})

					It("returns the user id", func() {
						err := standard.RecordMetric(context, "test-metric", data)
						Expect(err).ToNot(HaveOccurred())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(context.ValidateTest()).To(BeTrue())
					})
				})

				Context("with a successful response without data", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/metrics/thisuser/test-metric", ""),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, nil, nil)),
						)
					})

					It("returns the user id", func() {
						err := standard.RecordMetric(context, "test-metric", nil)
						Expect(err).ToNot(HaveOccurred())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(context.ValidateTest()).To(BeTrue())
					})
				})
			})

			Context("as server", func() {
				BeforeEach(func() {
					context.TestAuthenticationDetails.TokenOutputs = []string{"test-authentication-token"}
					context.TestAuthenticationDetails.IsServerOutputs = []bool{true}
				})

				Context("with an unexpected response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/metrics/server/testservices/test-metric", "left=handed&right=correct"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := standard.RecordMetric(context, "test-metric", data)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(HavePrefix("client: unexpected response status code 400 from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(context.ValidateTest()).To(BeTrue())
					})
				})

				Context("with an unauthorized response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/metrics/server/testservices/test-metric", "left=handed&right=correct"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusUnauthorized, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := standard.RecordMetric(context, "test-metric", data)
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(context.ValidateTest()).To(BeTrue())
					})
				})

				Context("with a successful response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/metrics/server/testservices/test-metric", "left=handed&right=correct"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, nil, nil)),
						)
					})

					It("returns the user id", func() {
						err := standard.RecordMetric(context, "test-metric", data)
						Expect(err).ToNot(HaveOccurred())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(context.ValidateTest()).To(BeTrue())
					})
				})

				Context("with a successful response without data", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/metrics/server/testservices/test-metric", ""),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, nil, nil)),
						)
					})

					It("returns the user id", func() {
						err := standard.RecordMetric(context, "test-metric", nil)
						Expect(err).ToNot(HaveOccurred())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(context.ValidateTest()).To(BeTrue())
					})
				})
			})
		})
	})
})
