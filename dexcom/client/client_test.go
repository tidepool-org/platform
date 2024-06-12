package client_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/dexcom"
	dexcomClient "github.com/tidepool-org/platform/dexcom/client"
	dexcomTest "github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	oauthTest "github.com/tidepool-org/platform/oauth/test"
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Client", func() {
	var userAgent string
	var config *client.Config
	var tokenSourceSource *oauthTest.TokenSourceSource

	BeforeEach(func() {
		userAgent = testHttp.NewUserAgent()
		config = client.NewConfig()
		config.UserAgent = userAgent
		tokenSourceSource = oauthTest.NewTokenSourceSource()
	})

	AfterEach(func() {
		tokenSourceSource.AssertOutputsEmpty()
	})

	Context("New", func() {
		BeforeEach(func() {
			config.Address = testHttp.NewAddress()
		})

		It("returns an error when config is missing", func() {
			clnt, err := dexcomClient.New(nil, tokenSourceSource)
			Expect(err).To(MatchError("config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error when config is invalid", func() {
			config.Address = ""
			clnt, err := dexcomClient.New(config, tokenSourceSource)
			Expect(err).To(MatchError("config is invalid; address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error when token source source is missing", func() {
			clnt, err := dexcomClient.New(config, nil)
			Expect(err).To(MatchError("token source source is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(dexcomClient.New(config, tokenSourceSource)).ToNot(BeNil())
		})
	})

	Context("with started server and new client", func() {
		var server *Server
		var responseHeaders http.Header
		var ctx context.Context
		var startTime time.Time
		var endTime time.Time
		var requestQuery string
		var tokenSource *oauthTest.TokenSource
		var clnt *dexcomClient.Client

		BeforeEach(func() {
			server = NewServer()
			responseHeaders = http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			startTime = test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now())
			endTime = test.RandomTimeFromRange(startTime, time.Now())
			requestQuery = fmt.Sprintf("startDate=%s&endDate=%s", startTime.UTC().Format(dexcom.DateRangeTimeFormat), endTime.UTC().Format(dexcom.DateRangeTimeFormat))
			tokenSource = oauthTest.NewTokenSource()
		})

		JustBeforeEach(func() {
			config.Address = server.URL()
			var err error
			clnt, err = dexcomClient.New(config, tokenSourceSource)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
			tokenSource.AssertOutputsEmpty()
		})

		Context("GetAlerts", func() {
			var responseAlertsResponse *dexcom.AlertsResponse

			BeforeEach(func() {
				responseAlertsResponse = dexcomTest.RandomAlertsResponse()
			})

			It("returns error when http client source is missing", func() {
				alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, nil)
				Expect(err).To(MatchError("unable to get alerts; http client source is missing"))
				Expect(alertsResponse).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error when http client source returns an error", func() {
				responseErr := errorsTest.RandomError()
				tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: nil, Error: responseErr}}
				alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, tokenSource)
				Expect(err).To(MatchError(fmt.Sprintf("unable to get alerts; %s", responseErr)))
				Expect(alertsResponse).To(BeNil())
				Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error when http client source returns that indicates an oauth token failure", func() {
				responseErr := errors.New("oauth2: cannot fetch token: 400 Bad Request")
				tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: nil, Error: responseErr}}
				alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, tokenSource)
				Expect(err).To(MatchError("unable to get alerts; oauth2: cannot fetch token: 400 Bad Request; authentication token is invalid"))
				Expect(alertsResponse).To(BeNil())
				Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			When("http client source returns successfully", func() {
				var httpClient *http.Client

				BeforeEach(func() {
					httpClient = http.DefaultClient
					tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: httpClient, Error: nil}}
				})

				It("returns error when context is missing", func() {
					ctx = nil
					alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, tokenSource)
					Expect(err).To(MatchError("unable to get alerts; context is missing"))
					Expect(alertsResponse).To(BeNil())
					Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error when the server is not reachable", func() {
					server.Close()
					server = nil
					alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, tokenSource)
					Expect(err.Error()).To(MatchRegexp("unable to get alerts; unable to perform request to .*: connect: connection refused"))
					Expect(alertsResponse).To(BeNil())
					Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				})

				requestAssertions := func() {
					Context("with an bad request 400", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/alerts", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusBadRequest, []byte{255, 255, 255}, responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get alerts; bad request"))
							Expect(alertsResponse).To(BeNil())
						})
					})

					Context("with an forbidden response 403", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/alerts", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusForbidden, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get alerts; authentication token is not authorized for requested action"))
							Expect(alertsResponse).To(BeNil())
						})
					})

					Context("with an resource not found 404", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/alerts", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusNotFound, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get alerts; resource not found"))
							Expect(alertsResponse).To(BeNil())
						})
					})

					Context("with an unexpected response 500", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/alerts", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusInternalServerError, nil, responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, tokenSource)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(MatchRegexp("unable to get alerts; unexpected response status code 500 from"))
							Expect(alertsResponse).To(BeNil())
						})
					})

					Context("with an unparseable response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/alerts", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusOK, []byte("{"), responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get alerts; json is malformed"))
							Expect(alertsResponse).To(BeNil())
						})
					})

					Context("with a successful response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/alerts", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusOK, test.MarshalResponseBody(responseAlertsResponse), responseHeaders),
								),
							)
						})

						It("returns success", func() {
							alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, tokenSource)
							Expect(err).ToNot(HaveOccurred())
							Expect(alertsResponse).To(Equal(responseAlertsResponse))
						})
					})
				}

				When("the server responds directly to the one request", func() {
					AfterEach(func() {
						Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
						Expect(tokenSource.ExpireTokenInvocations).To(Equal(0))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
					requestAssertions()
				})

				When("the server responds with unauthorized, the token is expired and the request retried", func() {
					BeforeEach(func() {
						tokenSource.HTTPClientOutputs = append(tokenSource.HTTPClientOutputs, oauthTest.HTTPClientOutput{HTTPClient: httpClient, Error: nil})
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/v3/users/self/alerts", requestQuery),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyBody(nil),
								RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
							),
						)
					})

					AfterEach(func() {
						Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}, {Context: ctx, TokenSourceSource: tokenSourceSource}}))
						Expect(tokenSource.ExpireTokenInvocations).To(Equal(1))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})

					requestAssertions()

					Context("with an unauthorized response 401", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/alerts", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get alerts; authentication token is invalid"))
							Expect(alertsResponse).To(BeNil())
						})
					})
				})
			})
		})

		Context("GetCalibrations", func() {
			var responseCalibrationsResponse *dexcom.CalibrationsResponse

			BeforeEach(func() {
				responseCalibrationsResponse = dexcomTest.RandomCalibrationsResponse()
			})

			It("returns error when http client source is missing", func() {
				calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, nil)
				Expect(err).To(MatchError("unable to get calibrations; http client source is missing"))
				Expect(calibrationsResponse).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error when http client source returns an error", func() {
				responseErr := errorsTest.RandomError()
				tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: nil, Error: responseErr}}
				calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, tokenSource)
				Expect(err).To(MatchError(fmt.Sprintf("unable to get calibrations; %s", responseErr)))
				Expect(calibrationsResponse).To(BeNil())
				Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error when http client source returns that indicates an oauth token failure", func() {
				responseErr := errors.New("oauth2: cannot fetch token: 400 Bad Request")
				tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: nil, Error: responseErr}}
				calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, tokenSource)
				Expect(err).To(MatchError("unable to get calibrations; oauth2: cannot fetch token: 400 Bad Request; authentication token is invalid"))
				Expect(calibrationsResponse).To(BeNil())
				Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			When("http client source returns successfully", func() {
				var httpClient *http.Client

				BeforeEach(func() {
					httpClient = http.DefaultClient
					tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: httpClient, Error: nil}}
				})

				It("returns error when context is missing", func() {
					ctx = nil
					calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, tokenSource)
					Expect(err).To(MatchError("unable to get calibrations; context is missing"))
					Expect(calibrationsResponse).To(BeNil())
					Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error when the server is not reachable", func() {
					server.Close()
					server = nil
					calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, tokenSource)
					Expect(err.Error()).To(MatchRegexp("unable to get calibrations; unable to perform request to .*: connect: connection refused"))
					Expect(calibrationsResponse).To(BeNil())
					Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				})

				requestAssertions := func() {
					Context("with an bad request 400", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/calibrations", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusBadRequest, []byte{255, 255, 255}, responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get calibrations; bad request"))
							Expect(calibrationsResponse).To(BeNil())
						})
					})

					Context("with an forbidden response 403", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/calibrations", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusForbidden, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get calibrations; authentication token is not authorized for requested action"))
							Expect(calibrationsResponse).To(BeNil())
						})
					})

					Context("with an resource not found 404", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/calibrations", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusNotFound, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get calibrations; resource not found"))
							Expect(calibrationsResponse).To(BeNil())
						})
					})

					Context("with an unexpected response 500", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/calibrations", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusInternalServerError, nil, responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, tokenSource)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(MatchRegexp("unable to get calibrations; unexpected response status code 500 from"))
							Expect(calibrationsResponse).To(BeNil())
						})
					})

					Context("with an unparseable response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/calibrations", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusOK, []byte("{"), responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get calibrations; json is malformed"))
							Expect(calibrationsResponse).To(BeNil())
						})
					})

					Context("with a successful response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/calibrations", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusOK, test.MarshalResponseBody(responseCalibrationsResponse), responseHeaders),
								),
							)
						})

						It("returns success", func() {
							calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, tokenSource)
							Expect(err).ToNot(HaveOccurred())
							Expect(calibrationsResponse).To(Equal(responseCalibrationsResponse))
						})
					})
				}

				When("the server responds directly to the one request", func() {
					AfterEach(func() {
						Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
						Expect(tokenSource.ExpireTokenInvocations).To(Equal(0))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})

					requestAssertions()
				})

				When("the server responds with unauthorized, the token is expired and the request retried", func() {
					BeforeEach(func() {
						tokenSource.HTTPClientOutputs = append(tokenSource.HTTPClientOutputs, oauthTest.HTTPClientOutput{HTTPClient: httpClient, Error: nil})
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/v3/users/self/calibrations", requestQuery),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyBody(nil),
								RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
							),
						)
					})

					AfterEach(func() {
						Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}, {Context: ctx, TokenSourceSource: tokenSourceSource}}))
						Expect(tokenSource.ExpireTokenInvocations).To(Equal(1))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})

					requestAssertions()

					Context("with an unauthorized response 401", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/calibrations", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get calibrations; authentication token is invalid"))
							Expect(calibrationsResponse).To(BeNil())
						})
					})
				})
			})
		})

		Context("GetDevices", func() {
			var responseDevicesResponse *dexcom.DevicesResponse

			BeforeEach(func() {
				responseDevicesResponse = dexcomTest.RandomDevicesResponse()
			})

			It("returns error when http client source is missing", func() {
				devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, nil)
				Expect(err).To(MatchError("unable to get devices; http client source is missing"))
				Expect(devicesResponse).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error when http client source returns an error", func() {
				responseErr := errorsTest.RandomError()
				tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: nil, Error: responseErr}}
				devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, tokenSource)
				Expect(err).To(MatchError(fmt.Sprintf("unable to get devices; %s", responseErr)))
				Expect(devicesResponse).To(BeNil())
				Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error when http client source returns that indicates an oauth token failure", func() {
				responseErr := errors.New("oauth2: cannot fetch token: 400 Bad Request")
				tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: nil, Error: responseErr}}
				devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, tokenSource)
				Expect(err).To(MatchError("unable to get devices; oauth2: cannot fetch token: 400 Bad Request; authentication token is invalid"))
				Expect(devicesResponse).To(BeNil())
				Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			When("http client source returns successfully", func() {
				var httpClient *http.Client

				BeforeEach(func() {
					httpClient = http.DefaultClient
					tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: httpClient, Error: nil}}
				})

				It("returns error when context is missing", func() {
					ctx = nil
					devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, tokenSource)
					Expect(err).To(MatchError("unable to get devices; context is missing"))
					Expect(devicesResponse).To(BeNil())
					Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error when the server is not reachable", func() {
					server.Close()
					server = nil
					devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, tokenSource)
					Expect(err.Error()).To(MatchRegexp("unable to get devices; unable to perform request to .*: connect: connection refused"))
					Expect(devicesResponse).To(BeNil())
					Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				})

				requestAssertions := func() {
					Context("with an bad request 400", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/devices", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusBadRequest, []byte{255, 255, 255}, responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get devices; bad request"))
							Expect(devicesResponse).To(BeNil())
						})
					})

					Context("with an forbidden response 403", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/devices", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusForbidden, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get devices; authentication token is not authorized for requested action"))
							Expect(devicesResponse).To(BeNil())
						})
					})

					Context("with an resource not found 404", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/devices", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusNotFound, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get devices; resource not found"))
							Expect(devicesResponse).To(BeNil())
						})
					})

					Context("with an unexpected response 500", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/devices", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusInternalServerError, nil, responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, tokenSource)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(MatchRegexp("unable to get devices; unexpected response status code 500 from"))
							Expect(devicesResponse).To(BeNil())
						})
					})

					Context("with an unparseable response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/devices", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusOK, []byte("{"), responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get devices; json is malformed"))
							Expect(devicesResponse).To(BeNil())
						})
					})

					Context("with a successful response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/devices", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusOK, test.MarshalResponseBody(responseDevicesResponse), responseHeaders),
								),
							)
						})

						It("returns success", func() {
							devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, tokenSource)
							Expect(err).ToNot(HaveOccurred())
							Expect(structureNormalizer.New().Normalize(responseDevicesResponse)).To(Succeed())
							Expect(devicesResponse).To(Equal(responseDevicesResponse))
						})
					})
				}

				When("the server responds directly to the one request", func() {
					AfterEach(func() {
						Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
						Expect(tokenSource.ExpireTokenInvocations).To(Equal(0))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})

					requestAssertions()
				})

				When("the server responds with unauthorized, the token is expired and the request retried", func() {
					BeforeEach(func() {
						tokenSource.HTTPClientOutputs = append(tokenSource.HTTPClientOutputs, oauthTest.HTTPClientOutput{HTTPClient: httpClient, Error: nil})
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/v3/users/self/devices", requestQuery),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyBody(nil),
								RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
							),
						)
					})

					AfterEach(func() {
						Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}, {Context: ctx, TokenSourceSource: tokenSourceSource}}))
						Expect(tokenSource.ExpireTokenInvocations).To(Equal(1))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})

					requestAssertions()

					Context("with an unauthorized response 401", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/devices", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get devices; authentication token is invalid"))
							Expect(devicesResponse).To(BeNil())
						})
					})
				})
			})
		})

		Context("GetEGVs", func() {
			var responseEGVsResponse *dexcom.EGVsResponse

			BeforeEach(func() {
				responseEGVsResponse = dexcomTest.RandomEGVsResponse()
			})

			It("returns error when http client source is missing", func() {
				egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, nil)
				Expect(err).To(MatchError("unable to get egvs; http client source is missing"))
				Expect(egvsResponse).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error when http client source returns an error", func() {
				responseErr := errorsTest.RandomError()
				tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: nil, Error: responseErr}}
				egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, tokenSource)
				Expect(err).To(MatchError(fmt.Sprintf("unable to get egvs; %s", responseErr)))
				Expect(egvsResponse).To(BeNil())
				Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error when http client source returns that indicates an oauth token failure", func() {
				responseErr := errors.New("oauth2: cannot fetch token: 400 Bad Request")
				tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: nil, Error: responseErr}}
				egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, tokenSource)
				Expect(err).To(MatchError("unable to get egvs; oauth2: cannot fetch token: 400 Bad Request; authentication token is invalid"))
				Expect(egvsResponse).To(BeNil())
				Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			When("http client source returns successfully", func() {
				var httpClient *http.Client

				BeforeEach(func() {
					httpClient = http.DefaultClient
					tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: httpClient, Error: nil}}
				})

				It("returns error when context is missing", func() {
					ctx = nil
					egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, tokenSource)
					Expect(err).To(MatchError("unable to get egvs; context is missing"))
					Expect(egvsResponse).To(BeNil())
					Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error when the server is not reachable", func() {
					server.Close()
					server = nil
					egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, tokenSource)
					Expect(err.Error()).To(MatchRegexp("unable to get egvs; unable to perform request to .*: connect: connection refused"))
					Expect(egvsResponse).To(BeNil())
					Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				})

				requestAssertions := func() {
					Context("with an bad request 400", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/egvs", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusBadRequest, []byte{255, 255, 255}, responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get egvs; bad request"))
							Expect(egvsResponse).To(BeNil())
						})
					})

					Context("with an forbidden response 403", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/egvs", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusForbidden, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get egvs; authentication token is not authorized for requested action"))
							Expect(egvsResponse).To(BeNil())
						})
					})

					Context("with an resource not found 404", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/egvs", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusNotFound, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get egvs; resource not found"))
							Expect(egvsResponse).To(BeNil())
						})
					})

					Context("with an unexpected response 500", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/egvs", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusInternalServerError, nil, responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, tokenSource)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(MatchRegexp("unable to get egvs; unexpected response status code 500 from"))
							Expect(egvsResponse).To(BeNil())
						})
					})

					Context("with an unparseable response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/egvs", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusOK, []byte("{"), responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get egvs; json is malformed"))
							Expect(egvsResponse).To(BeNil())
						})
					})

					Context("with a successful response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/egvs", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusOK, test.MarshalResponseBody(responseEGVsResponse), responseHeaders),
								),
							)
						})

						It("returns success", func() {
							egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, tokenSource)
							Expect(err).ToNot(HaveOccurred())
							Expect(egvsResponse).To(Equal(responseEGVsResponse))
						})
					})
				}

				When("the server responds directly to the one request", func() {
					AfterEach(func() {
						Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
						Expect(tokenSource.ExpireTokenInvocations).To(Equal(0))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})

					requestAssertions()
				})

				When("the server responds with unauthorized, the token is expired and the request retried", func() {
					BeforeEach(func() {
						tokenSource.HTTPClientOutputs = append(tokenSource.HTTPClientOutputs, oauthTest.HTTPClientOutput{HTTPClient: httpClient, Error: nil})
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/v3/users/self/egvs", requestQuery),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyBody(nil),
								RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
							),
						)
					})

					AfterEach(func() {
						Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}, {Context: ctx, TokenSourceSource: tokenSourceSource}}))
						Expect(tokenSource.ExpireTokenInvocations).To(Equal(1))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})

					requestAssertions()

					Context("with an unauthorized response 401", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/egvs", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get egvs; authentication token is invalid"))
							Expect(egvsResponse).To(BeNil())
						})
					})
				})
			})
		})

		Context("GetEvents", func() {
			var responseEventsResponse *dexcom.EventsResponse

			BeforeEach(func() {
				responseEventsResponse = dexcomTest.RandomEventsResponse()
			})

			It("returns error when http client source is missing", func() {
				eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, nil)
				Expect(err).To(MatchError("unable to get events; http client source is missing"))
				Expect(eventsResponse).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error when http client source returns an error", func() {
				responseErr := errorsTest.RandomError()
				tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: nil, Error: responseErr}}
				eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, tokenSource)
				Expect(err).To(MatchError(fmt.Sprintf("unable to get events; %s", responseErr)))
				Expect(eventsResponse).To(BeNil())
				Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error when http client source returns that indicates an oauth token failure", func() {
				responseErr := errors.New("oauth2: cannot fetch token: 400 Bad Request")
				tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: nil, Error: responseErr}}
				eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, tokenSource)
				Expect(err).To(MatchError("unable to get events; oauth2: cannot fetch token: 400 Bad Request; authentication token is invalid"))
				Expect(eventsResponse).To(BeNil())
				Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			When("http client source returns successfully", func() {
				var httpClient *http.Client

				BeforeEach(func() {
					httpClient = http.DefaultClient
					tokenSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: httpClient, Error: nil}}
				})

				It("returns error when context is missing", func() {
					ctx = nil
					eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, tokenSource)
					Expect(err).To(MatchError("unable to get events; context is missing"))
					Expect(eventsResponse).To(BeNil())
					Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error when the server is not reachable", func() {
					server.Close()
					server = nil
					eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, tokenSource)
					Expect(err.Error()).To(MatchRegexp("unable to get events; unable to perform request to .*: connect: connection refused"))
					Expect(eventsResponse).To(BeNil())
					Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				})

				requestAssertions := func() {
					Context("with an bad request 400", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/events", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusBadRequest, []byte{255, 255, 255}, responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get events; bad request"))
							Expect(eventsResponse).To(BeNil())
						})
					})

					Context("with an forbidden response 403", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/events", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusForbidden, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get events; authentication token is not authorized for requested action"))
							Expect(eventsResponse).To(BeNil())
						})
					})

					Context("with an resource not found 404", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/events", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusNotFound, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get events; resource not found"))
							Expect(eventsResponse).To(BeNil())
						})
					})

					Context("with an unexpected response 500", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/events", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusInternalServerError, nil, responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, tokenSource)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(MatchRegexp("unable to get events; unexpected response status code 500 from"))
							Expect(eventsResponse).To(BeNil())
						})
					})

					Context("with an unparseable response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/events", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusOK, []byte("{"), responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get events; json is malformed"))
							Expect(eventsResponse).To(BeNil())
						})
					})

					Context("with a successful response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/events", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusOK, test.MarshalResponseBody(responseEventsResponse), responseHeaders),
								),
							)
						})

						It("returns success", func() {
							eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, tokenSource)
							Expect(err).ToNot(HaveOccurred())
							Expect(eventsResponse).To(Equal(responseEventsResponse))
						})
					})
				}

				When("the server responds directly to the one request", func() {
					AfterEach(func() {
						Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
						Expect(tokenSource.ExpireTokenInvocations).To(Equal(0))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})

					requestAssertions()
				})

				When("the server responds with unauthorized, the token is expired and the request retried", func() {
					BeforeEach(func() {
						tokenSource.HTTPClientOutputs = append(tokenSource.HTTPClientOutputs, oauthTest.HTTPClientOutput{HTTPClient: httpClient, Error: nil})
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/v3/users/self/events", requestQuery),
								VerifyHeaderKV("User-Agent", userAgent),
								VerifyBody(nil),
								RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
							),
						)
					})

					AfterEach(func() {
						Expect(tokenSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}, {Context: ctx, TokenSourceSource: tokenSourceSource}}))
						Expect(tokenSource.ExpireTokenInvocations).To(Equal(1))
						Expect(server.ReceivedRequests()).To(HaveLen(2))
					})

					requestAssertions()

					Context("with an unauthorized response 401", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/events", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, tokenSource)
							Expect(err).To(MatchError("unable to get events; authentication token is invalid"))
							Expect(eventsResponse).To(BeNil())
						})
					})
				})
			})
		})
	})
})
