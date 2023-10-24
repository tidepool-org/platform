package client_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/metric"
	metricClient "github.com/tidepool-org/platform/metric/client"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
	"github.com/tidepool-org/platform/version"
)

var _ = Describe("Client", func() {
	var name string
	var versionReporter version.Reporter

	BeforeEach(func() {
		name = test.RandomStringFromRangeAndCharset(1, 64, test.CharsetAlphaNumeric)
		var err error
		versionReporter, err = version.NewReporter("1.2.3", "4567890", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmn")
		Expect(err).ToNot(HaveOccurred())
		Expect(versionReporter).ToNot(BeNil())
	})

	Context("New", func() {
		var config *platform.Config

		BeforeEach(func() {
			config = platform.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = testHttp.NewAddress()
			config.UserAgent = testHttp.NewUserAgent()
		})

		It("returns an error if config is missing", func() {
			clnt, err := metricClient.New(nil, platform.AuthorizeAsUser, name, versionReporter)
			Expect(err).To(MatchError("config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if name is missing", func() {
			clnt, err := metricClient.New(config, platform.AuthorizeAsUser, "", versionReporter)
			Expect(err).To(MatchError("name is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if version reporter is missing", func() {
			clnt, err := metricClient.New(config, platform.AuthorizeAsUser, name, nil)
			Expect(err).To(MatchError("version reporter is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config address is missing", func() {
			config.Address = ""
			clnt, err := metricClient.New(config, platform.AuthorizeAsUser, name, versionReporter)
			Expect(err).To(MatchError("config is invalid; address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config user agent is missing", func() {
			config.UserAgent = ""
			clnt, err := metricClient.New(config, platform.AuthorizeAsUser, name, versionReporter)
			Expect(err).To(MatchError("config is invalid; user agent is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			clnt, err := metricClient.New(config, platform.AuthorizeAsUser, name, versionReporter)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})
	})

	Context("with started server and new client", func() {
		var server *Server
		var userAgent string
		var clnt metric.Client
		var ctx context.Context

		BeforeEach(func() {
			server = NewServer()
			userAgent = testHttp.NewUserAgent()
			config := platform.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = server.URL()
			config.UserAgent = userAgent
			var err error
			clnt, err = metricClient.New(config, platform.AuthorizeAsUser, name, versionReporter)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
			ctx = context.Background()
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("RecordMetric", func() {
			var metric string
			var data map[string]string

			BeforeEach(func() {
				metric = test.RandomStringFromRangeAndCharset(1, 32, test.CharsetAlphaNumeric)
				data = map[string]string{
					"left":  "handed",
					"right": "correct",
				}
			})

			It("returns error if context is missing", func() {
				Expect(clnt.RecordMetric(nil, metric, data)).To(MatchError("context is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if metric is missing", func() {
				Expect(clnt.RecordMetric(ctx, "", data)).To(MatchError("metric is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			Context("as user", func() {
				var token string

				BeforeEach(func() {
					token = test.RandomStringFromRangeAndCharset(64, 64, test.CharsetAlphaNumeric)
					ctx = log.NewContextWithLogger(ctx, logNull.NewLogger())
					ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexidecimalLowercase), token))
				})

				Context("as user", func() {
					Context("with an unauthorized response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/metrics/thisuser/"+metric, "left=handed&right=correct&sourceVersion=1.2.3"),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("X-Tidepool-Session-Token", token),
									VerifyBody(nil),
									RespondWith(http.StatusUnauthorized, nil)),
							)
						})

						It("returns an error", func() {
							err := clnt.RecordMetric(ctx, metric, data)
							Expect(err).To(MatchError("authentication token is invalid"))
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a forbidden response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/metrics/thisuser/"+metric, "left=handed&right=correct&sourceVersion=1.2.3"),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("X-Tidepool-Session-Token", token),
									VerifyBody(nil),
									RespondWith(http.StatusForbidden, nil)),
							)
						})

						It("returns an error", func() {
							err := clnt.RecordMetric(ctx, metric, data)
							Expect(err).To(MatchError("authentication token is not authorized for requested action"))
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a successful response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/metrics/thisuser/"+metric, "left=handed&right=correct&sourceVersion=1.2.3"),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("X-Tidepool-Session-Token", token),
									VerifyBody(nil),
									RespondWith(http.StatusOK, nil)),
							)
						})

						It("returns success", func() {
							err := clnt.RecordMetric(ctx, metric, data)
							Expect(err).ToNot(HaveOccurred())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a successful response without data", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/metrics/thisuser/"+metric, "sourceVersion=1.2.3"),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("X-Tidepool-Session-Token", token),
									VerifyBody(nil),
									RespondWith(http.StatusOK, nil)),
							)
						})

						It("returns success", func() {
							err := clnt.RecordMetric(ctx, metric)
							Expect(err).ToNot(HaveOccurred())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})
				})
			})

			Context("as server", func() {
				var token string

				BeforeEach(func() {
					token = test.RandomStringFromRangeAndCharset(64, 64, test.CharsetAlphaNumeric)
					ctx = log.NewContextWithLogger(ctx, logNull.NewLogger())
					ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodSessionToken, "", token))
				})

				Context("with an unauthorized response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/metrics/server/"+name+"/"+metric, "left=handed&right=correct&sourceVersion=1.2.3"),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyHeaderKV("X-Tidepool-Session-Token", token),
								VerifyBody(nil),
								RespondWith(http.StatusUnauthorized, nil)),
						)
					})

					It("returns an error", func() {
						err := clnt.RecordMetric(ctx, metric, data)
						Expect(err).To(MatchError("authentication token is invalid"))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a forbidden response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/metrics/server/"+name+"/"+metric, "left=handed&right=correct&sourceVersion=1.2.3"),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyHeaderKV("X-Tidepool-Session-Token", token),
								VerifyBody(nil),
								RespondWith(http.StatusForbidden, nil)),
						)
					})

					It("returns an error", func() {
						err := clnt.RecordMetric(ctx, metric, data)
						Expect(err).To(MatchError("authentication token is not authorized for requested action"))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a successful response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/metrics/server/"+name+"/"+metric, "left=handed&right=correct&sourceVersion=1.2.3"),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyHeaderKV("X-Tidepool-Session-Token", token),
								VerifyBody(nil),
								RespondWith(http.StatusOK, nil)),
						)
					})

					It("returns success", func() {
						err := clnt.RecordMetric(ctx, metric, data)
						Expect(err).ToNot(HaveOccurred())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a successful response without data", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/metrics/server/"+name+"/"+metric, "sourceVersion=1.2.3"),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyHeaderKV("X-Tidepool-Session-Token", token),
								VerifyBody(nil),
								RespondWith(http.StatusOK, nil)),
						)
					})

					It("returns success", func() {
						err := clnt.RecordMetric(ctx, metric)
						Expect(err).ToNot(HaveOccurred())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})
			})
		})
	})
})
