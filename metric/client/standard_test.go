package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/metric/client"
	userClient "github.com/tidepool-org/platform/user/client"
	"github.com/tidepool-org/platform/version"
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

func (t *TestContext) AuthenticationDetails() userClient.AuthenticationDetails {
	return t.TestAuthenticationDetails
}

func (t *TestContext) ValidateTest() bool {
	return t.TestAuthenticationDetails.ValidateTest()
}

var _ = Describe("Standard", func() {
	var versionReporter version.Reporter
	var context *TestContext

	BeforeEach(func() {
		var err error
		versionReporter, err = version.NewReporter("1.2.3", "4567890", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn")
		Expect(err).ToNot(HaveOccurred())
		Expect(versionReporter).ToNot(BeNil())
		context = &TestContext{
			TestLogger:                null.NewLogger(),
			TestRequest:               &rest.Request{},
			TestAuthenticationDetails: &TestAuthenticationDetails{},
		}
	})

	Context("NewStandard", func() {
		var config *client.Config

		BeforeEach(func() {
			config = &client.Config{
				Address: "http://localhost:1234",
				Timeout: 30 * time.Second,
			}
		})

		It("returns an error if version reporter is missing", func() {
			standard, err := client.NewStandard(nil, "test", config)
			Expect(err).To(MatchError("client: version reporter is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if name is missing", func() {
			standard, err := client.NewStandard(versionReporter, "", config)
			Expect(err).To(MatchError("client: name is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config is missing", func() {
			standard, err := client.NewStandard(versionReporter, "test", nil)
			Expect(err).To(MatchError("client: config is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config address is missing", func() {
			config.Address = ""
			standard, err := client.NewStandard(versionReporter, "test", config)
			Expect(err).To(MatchError("client: config is invalid; client: address is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config timeout is invalid", func() {
			config.Timeout = 0
			standard, err := client.NewStandard(versionReporter, "test", config)
			Expect(err).To(MatchError("client: config is invalid; client: timeout is invalid"))
			Expect(standard).To(BeNil())
		})

		It("returns success", func() {
			standard, err := client.NewStandard(versionReporter, "test", config)
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
				Address: server.URL(),
				Timeout: 30 * time.Second,
			}
			data = map[string]string{
				"left":  "handed",
				"right": "correct",
			}
		})

		JustBeforeEach(func() {
			var err error
			standard, err = client.NewStandard(versionReporter, "test", config)
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
								ghttp.VerifyRequest("GET", "/metrics/thisuser/test-metric", "left=handed&right=correct&sourceVersion=1.2.3"),
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
								ghttp.VerifyRequest("GET", "/metrics/thisuser/test-metric", "left=handed&right=correct&sourceVersion=1.2.3"),
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
								ghttp.VerifyRequest("GET", "/metrics/thisuser/test-metric", "left=handed&right=correct&sourceVersion=1.2.3"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, nil, nil)),
						)
					})

					It("returns success", func() {
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
								ghttp.VerifyRequest("GET", "/metrics/thisuser/test-metric", "sourceVersion=1.2.3"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, nil, nil)),
						)
					})

					It("returns success", func() {
						err := standard.RecordMetric(context, "test-metric")
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
								ghttp.VerifyRequest("GET", "/metrics/server/test/test-metric", "left=handed&right=correct&sourceVersion=1.2.3"),
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
								ghttp.VerifyRequest("GET", "/metrics/server/test/test-metric", "left=handed&right=correct&sourceVersion=1.2.3"),
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
								ghttp.VerifyRequest("GET", "/metrics/server/test/test-metric", "left=handed&right=correct&sourceVersion=1.2.3"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, nil, nil)),
						)
					})

					It("returns success", func() {
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
								ghttp.VerifyRequest("GET", "/metrics/server/test/test-metric", "sourceVersion=1.2.3"),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", "test-authentication-token"),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, nil, nil)),
						)
					})

					It("returns success", func() {
						err := standard.RecordMetric(context, "test-metric")
						Expect(err).ToNot(HaveOccurred())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(context.ValidateTest()).To(BeTrue())
					})
				})
			})
		})
	})
})
