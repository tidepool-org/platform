package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"net/http"
	"time"

	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/id"
	metricClient "github.com/tidepool-org/platform/metric/client"
	"github.com/tidepool-org/platform/version"
)

var _ = Describe("Client", func() {
	var name string
	var versionReporter version.Reporter

	BeforeEach(func() {
		name = id.New()
		var err error
		versionReporter, err = version.NewReporter("1.2.3", "4567890", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn")
		Expect(err).ToNot(HaveOccurred())
		Expect(versionReporter).ToNot(BeNil())
	})

	Context("NewClient", func() {
		var config *client.Config

		BeforeEach(func() {
			config = client.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = "http://localhost:1234"
			config.Timeout = 30 * time.Second
		})

		It("returns an error if config is missing", func() {
			clnt, err := metricClient.NewClient(nil, name, versionReporter)
			Expect(err).To(MatchError("client: config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if name is missing", func() {
			clnt, err := metricClient.NewClient(config, "", versionReporter)
			Expect(err).To(MatchError("client: name is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if version reporter is missing", func() {
			clnt, err := metricClient.NewClient(config, name, nil)
			Expect(err).To(MatchError("client: version reporter is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config address is missing", func() {
			config.Address = ""
			clnt, err := metricClient.NewClient(config, name, versionReporter)
			Expect(err).To(MatchError("client: config is invalid; client: address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			clnt, err := metricClient.NewClient(config, name, versionReporter)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})
	})

	Context("with started server and new client", func() {
		var server *Server
		var clnt metricClient.Client
		var context *testAuth.Context

		BeforeEach(func() {
			server = NewServer()
			config := client.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = server.URL()
			config.Timeout = 30 * time.Second
			var err error
			clnt, err = metricClient.NewClient(config, name, versionReporter)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
			context = testAuth.NewContext()
			Expect(context).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
			Expect(context.UnusedOutputsCount()).To(Equal(0))
		})

		Context("RecordMetric", func() {
			var metric string
			var data map[string]string

			BeforeEach(func() {
				metric = id.New()
				data = map[string]string{
					"left":  "handed",
					"right": "correct",
				}
			})

			It("returns error if context is missing", func() {
				Expect(clnt.RecordMetric(nil, metric, data)).To(MatchError("client: context is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if metric is missing", func() {
				Expect(clnt.RecordMetric(context, "", data)).To(MatchError("client: metric is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			Context("with auth token", func() {
				var token string

				BeforeEach(func() {
					token = id.New()
					context.AuthDetailsImpl.TokenOutputs = []string{token}
				})

				Context("as user", func() {
					BeforeEach(func() {
						context.AuthDetailsImpl.IsServerOutputs = []bool{false}
					})

					Context("with an unauthorized response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/metrics/thisuser/"+metric, "left=handed&right=correct&sourceVersion=1.2.3"),
									VerifyHeaderKV("X-Tidepool-Session-Token", token),
									VerifyBody([]byte{}),
									RespondWith(http.StatusUnauthorized, nil, nil)),
							)
						})

						It("returns an error", func() {
							err := clnt.RecordMetric(context, metric, data)
							Expect(err).To(MatchError("client: unauthorized"))
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a successful response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/metrics/thisuser/"+metric, "left=handed&right=correct&sourceVersion=1.2.3"),
									VerifyHeaderKV("X-Tidepool-Session-Token", token),
									VerifyBody([]byte{}),
									RespondWith(http.StatusOK, nil, nil)),
							)
						})

						It("returns success", func() {
							err := clnt.RecordMetric(context, metric, data)
							Expect(err).ToNot(HaveOccurred())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a successful response without data", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/metrics/thisuser/"+metric, "sourceVersion=1.2.3"),
									VerifyHeaderKV("X-Tidepool-Session-Token", token),
									VerifyBody([]byte{}),
									RespondWith(http.StatusOK, nil, nil)),
							)
						})

						It("returns success", func() {
							err := clnt.RecordMetric(context, metric)
							Expect(err).ToNot(HaveOccurred())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})
				})

				Context("as server", func() {
					BeforeEach(func() {
						context.AuthDetailsImpl.IsServerOutputs = []bool{true}
					})

					Context("with an unauthorized response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/metrics/server/"+name+"/"+metric, "left=handed&right=correct&sourceVersion=1.2.3"),
									VerifyHeaderKV("X-Tidepool-Session-Token", token),
									VerifyBody([]byte{}),
									RespondWith(http.StatusUnauthorized, nil, nil)),
							)
						})

						It("returns an error", func() {
							err := clnt.RecordMetric(context, metric, data)
							Expect(err).To(MatchError("client: unauthorized"))
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a successful response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/metrics/server/"+name+"/"+metric, "left=handed&right=correct&sourceVersion=1.2.3"),
									VerifyHeaderKV("X-Tidepool-Session-Token", token),
									VerifyBody([]byte{}),
									RespondWith(http.StatusOK, nil, nil)),
							)
						})

						It("returns success", func() {
							err := clnt.RecordMetric(context, metric, data)
							Expect(err).ToNot(HaveOccurred())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a successful response without data", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/metrics/server/"+name+"/"+metric, "sourceVersion=1.2.3"),
									VerifyHeaderKV("X-Tidepool-Session-Token", token),
									VerifyBody([]byte{}),
									RespondWith(http.StatusOK, nil, nil)),
							)
						})

						It("returns success", func() {
							err := clnt.RecordMetric(context, metric)
							Expect(err).ToNot(HaveOccurred())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})
				})
			})
		})
	})
})
