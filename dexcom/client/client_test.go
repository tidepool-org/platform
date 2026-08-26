package client_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/dexcom"
	dexcomClient "github.com/tidepool-org/platform/dexcom/client"
	dexcomTest "github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	oauthTest "github.com/tidepool-org/platform/oauth/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("client", func() {
	It("RetrierRetries is expected", func() {
		Expect(dexcomClient.RetrierRetries).To(Equal(4))
	})

	It("RetrierDelay is expected", func() {
		Expect(dexcomClient.RetrierDelay).To(Equal(2 * time.Second))
	})

	It("RetrierJitter is expected", func() {
		Expect(dexcomClient.RetrierJitter).To(Equal(0.1))
	})

	Context("Client", func() {
		var userAgent string
		var config *client.Config
		var mockController *gomock.Controller
		var mockTokenSourceSource *oauthTest.MockTokenSourceSource

		BeforeEach(func() {
			userAgent = testHttp.NewUserAgent()
			config = client.NewConfig()
			config.UserAgent = userAgent
			mockController = gomock.NewController(GinkgoT())
			mockTokenSourceSource = oauthTest.NewMockTokenSourceSource(mockController)
		})

		Context("New", func() {
			BeforeEach(func() {
				config.Address = testHttp.NewAddress()
			})

			It("returns an error when config is missing", func() {
				clnt, err := dexcomClient.New(nil, nil, mockTokenSourceSource, request.RetryNone)
				Expect(err).To(MatchError("config is missing"))
				Expect(clnt).To(BeNil())
			})

			It("returns an error when config is invalid", func() {
				config.Address = ""
				clnt, err := dexcomClient.New(config, nil, mockTokenSourceSource, request.RetryNone)
				Expect(err).To(MatchError("config is invalid; address is missing"))
				Expect(clnt).To(BeNil())
			})

			It("returns an error when token source source is missing", func() {
				clnt, err := dexcomClient.New(config, nil, nil, request.RetryNone)
				Expect(err).To(MatchError("token source source is missing"))
				Expect(clnt).To(BeNil())
			})

			It("returns an error when retrier is missing", func() {
				clnt, err := dexcomClient.New(config, nil, mockTokenSourceSource, nil)
				Expect(err).To(MatchError("retrier is missing"))
				Expect(clnt).To(BeNil())
			})

			It("returns successfully", func() {
				Expect(dexcomClient.New(config, nil, mockTokenSourceSource, request.RetryNone)).ToNot(BeNil())
			})
		})

		Context("with started server and new client", func() {
			var server *Server
			var responseHeaders http.Header
			var ctx context.Context
			var mockTokenSource *oauthTest.MockTokenSource
			var clnt *dexcomClient.Client

			BeforeEach(func() {
				server = NewServer()
				responseHeaders = http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
				ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
				mockTokenSource = oauthTest.NewMockTokenSource(mockController)
			})

			JustBeforeEach(func() {
				config.Address = server.URL()
				var err error
				clnt, err = dexcomClient.New(config, nil, mockTokenSourceSource, request.RetryNone)
				Expect(err).ToNot(HaveOccurred())
				Expect(clnt).ToNot(BeNil())
			})

			AfterEach(func() {
				if server != nil {
					server.Close()
				}
			})

			Context("GetDataRange", func() {
				var lastSyncTime *time.Time
				var requestQuery string
				var responseDataRangesResponse *dexcom.DataRangesResponse

				BeforeEach(func() {
					lastSyncTime = nil
					requestQuery = ""
					responseDataRangesResponse = dexcomTest.RandomDataRangesResponse()
				})

				It("returns error when token source is missing", func() {
					dataRangeResponse, err := clnt.GetDataRange(ctx, lastSyncTime, nil)
					Expect(err).To(MatchError("unable to get data range; token source is missing"))
					Expect(dataRangeResponse).To(BeNil())
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error when context is missing", func() {
					dataRangeResponse, err := clnt.GetDataRange(context.Context(nil), lastSyncTime, mockTokenSource)
					Expect(err).To(MatchError("unable to get data range; context is missing"))
					Expect(dataRangeResponse).To(BeNil())
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error when token source returns an error", func() {
					responseErr := errorsTest.RandomError()
					mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(nil, responseErr)
					dataRangeResponse, err := clnt.GetDataRange(ctx, lastSyncTime, mockTokenSource)
					Expect(err).To(MatchError(fmt.Sprintf("unable to get data range; %s", responseErr)))
					Expect(dataRangeResponse).To(BeNil())
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error when token source returns that indicates an oauth token failure", func() {
					responseErr := errors.New(`oauth2: "invalid_grant"`)
					mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(nil, responseErr)
					dataRangeResponse, err := clnt.GetDataRange(ctx, lastSyncTime, mockTokenSource)
					Expect(err).To(MatchError(`unable to get data range; oauth2: "invalid_grant"; authentication token is invalid`))
					Expect(dataRangeResponse).To(BeNil())
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				When("token source returns successfully", func() {
					var httpClient *http.Client

					BeforeEach(func() {
						httpClient = http.DefaultClient
						mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(httpClient, nil)
						mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil)
					})

					It("returns error when the server is not reachable", func() {
						server.Close()
						server = nil
						dataRangeResponse, err := clnt.GetDataRange(ctx, lastSyncTime, mockTokenSource)
						Expect(err.Error()).To(MatchRegexp("unable to get data range; unable to perform request to .*: connect: connection refused"))
						Expect(dataRangeResponse).To(BeNil())
					})

					requestAssertions := func() {
						Context("with an bad request 400", func() {
							BeforeEach(func() {
								server.AppendHandlers(
									CombineHandlers(
										VerifyRequest("GET", "/v3/users/self/dataRange", requestQuery),
										VerifyHeaderKV("User-Agent", userAgent),
										VerifyBody(nil),
										RespondWith(http.StatusBadRequest, []byte{255, 255, 255}, responseHeaders),
									),
								)
							})

							It("returns an error", func() {
								dataRangeResponse, err := clnt.GetDataRange(ctx, lastSyncTime, mockTokenSource)
								Expect(err).To(MatchError("unable to get data range; bad request"))
								Expect(dataRangeResponse).To(BeNil())
							})
						})

						Context("with an forbidden response 403", func() {
							BeforeEach(func() {
								server.AppendHandlers(
									CombineHandlers(
										VerifyRequest("GET", "/v3/users/self/dataRange", requestQuery),
										VerifyHeaderKV("User-Agent", userAgent),
										VerifyBody(nil),
										RespondWith(http.StatusForbidden, "NOT JSON", responseHeaders),
									),
								)
							})

							It("returns an error", func() {
								dataRangeResponse, err := clnt.GetDataRange(ctx, lastSyncTime, mockTokenSource)
								Expect(err).To(MatchError("unable to get data range; authentication token is not authorized for requested action"))
								Expect(dataRangeResponse).To(BeNil())
							})
						})

						Context("with an resource not found 404", func() {
							BeforeEach(func() {
								server.AppendHandlers(
									CombineHandlers(
										VerifyRequest("GET", "/v3/users/self/dataRange", requestQuery),
										VerifyHeaderKV("User-Agent", userAgent),
										VerifyBody(nil),
										RespondWith(http.StatusNotFound, "NOT JSON", responseHeaders),
									),
								)
							})

							It("returns an error", func() {
								dataRangeResponse, err := clnt.GetDataRange(ctx, lastSyncTime, mockTokenSource)
								Expect(err).To(MatchError("unable to get data range; resource not found"))
								Expect(dataRangeResponse).To(BeNil())
							})
						})

						Context("with an unexpected response 500", func() {
							BeforeEach(func() {
								server.AppendHandlers(
									CombineHandlers(
										VerifyRequest("GET", "/v3/users/self/dataRange", requestQuery),
										VerifyHeaderKV("User-Agent", userAgent),
										VerifyBody(nil),
										RespondWith(http.StatusInternalServerError, nil, responseHeaders),
									),
								)
							})

							It("returns an error", func() {
								dataRangeResponse, err := clnt.GetDataRange(ctx, lastSyncTime, mockTokenSource)
								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(MatchRegexp("unable to get data range; unexpected response status code 500 from"))
								Expect(dataRangeResponse).To(BeNil())
							})
						})

						Context("with an unparsable response", func() {
							BeforeEach(func() {
								server.AppendHandlers(
									CombineHandlers(
										VerifyRequest("GET", "/v3/users/self/dataRange", requestQuery),
										VerifyHeaderKV("User-Agent", userAgent),
										VerifyBody(nil),
										RespondWith(http.StatusOK, []byte("{"), responseHeaders),
									),
								)
							})

							It("returns an error", func() {
								dataRangeResponse, err := clnt.GetDataRange(ctx, lastSyncTime, mockTokenSource)
								Expect(err).To(MatchError("unable to get data range; json is malformed"))
								Expect(dataRangeResponse).To(BeNil())
							})
						})

						Context("with a successful response", func() {
							BeforeEach(func() {
								server.AppendHandlers(
									CombineHandlers(
										VerifyRequest("GET", "/v3/users/self/dataRange", requestQuery),
										VerifyHeaderKV("User-Agent", userAgent),
										VerifyBody(nil),
										RespondWith(http.StatusOK, test.MarshalResponseBody(responseDataRangesResponse), responseHeaders),
									),
								)
							})

							It("returns success", func() {
								dataRangeResponse, err := clnt.GetDataRange(ctx, lastSyncTime, mockTokenSource)
								Expect(err).ToNot(HaveOccurred())
								Expect(dataRangeResponse).To(Equal(responseDataRangesResponse))
							})
						})
					}

					When("the server responds directly to the one request with last sync time", func() {
						BeforeEach(func() {
							lastSyncTime = pointer.FromTime(test.RandomTimeBeforeNow())
							requestQuery = fmt.Sprintf("lastSyncTime=%s", lastSyncTime.UTC().Format(time.RFC3339))
						})

						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})

						requestAssertions()
					})

					When("the server responds directly to the one request without last sync time", func() {
						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})

						requestAssertions()
					})

					When("the server responds with unauthorized, the token is expired and the request retried", func() {
						BeforeEach(func() {
							mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(httpClient, nil)
							mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil)
							mockTokenSource.EXPECT().ExpireToken(gomock.Not(gomock.Nil())).Return(true, nil)
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest("GET", "/v3/users/self/dataRange", requestQuery),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyBody(nil),
									RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
								),
							)
						})

						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(HaveLen(2))
						})

						requestAssertions()

						Context("with an unauthorized response 401", func() {
							BeforeEach(func() {
								server.AppendHandlers(
									CombineHandlers(
										VerifyRequest("GET", "/v3/users/self/dataRange", requestQuery),
										VerifyHeaderKV("User-Agent", userAgent),
										VerifyBody(nil),
										RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
									),
								)
							})

							It("returns an error", func() {
								dataRangeResponse, err := clnt.GetDataRange(ctx, lastSyncTime, mockTokenSource)
								Expect(err).To(MatchError("unable to get data range; authentication token is invalid"))
								Expect(dataRangeResponse).To(BeNil())
							})
						})
					})
				})
			})

			Context("with data range", func() {
				var startTime time.Time
				var endTime time.Time
				var requestQuery string

				BeforeEach(func() {
					startTime = test.RandomTimeBeforeNow()
					endTime = test.RandomTimeFromRange(startTime, time.Now())
					requestQuery = fmt.Sprintf("startDate=%s&endDate=%s", startTime.UTC().Format(dexcom.DateRangeTimeFormat), endTime.UTC().Format(dexcom.DateRangeTimeFormat))
				})

				Context("GetAlerts", func() {
					var responseAlertsResponse *dexcom.AlertsResponse

					BeforeEach(func() {
						responseAlertsResponse = dexcomTest.RandomAlertsResponse()
					})

					It("returns error when token source is missing", func() {
						alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, nil)
						Expect(err).To(MatchError("unable to get alerts; token source is missing"))
						Expect(alertsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when context is missing", func() {
						alertsResponse, err := clnt.GetAlerts(context.Context(nil), startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError("unable to get alerts; context is missing"))
						Expect(alertsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when token source returns an error", func() {
						responseErr := errorsTest.RandomError()
						mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(nil, responseErr)
						alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError(fmt.Sprintf("unable to get alerts; %s", responseErr)))
						Expect(alertsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when token source returns that indicates an oauth token failure", func() {
						responseErr := errors.New(`oauth2: "invalid_grant"`)
						mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(nil, responseErr)
						alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError(`unable to get alerts; oauth2: "invalid_grant"; authentication token is invalid`))
						Expect(alertsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					When("token source returns successfully", func() {
						var httpClient *http.Client

						BeforeEach(func() {
							httpClient = http.DefaultClient
							mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(httpClient, nil)
							mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil)
						})

						It("returns error when the server is not reachable", func() {
							server.Close()
							server = nil
							alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, mockTokenSource)
							Expect(err.Error()).To(MatchRegexp("unable to get alerts; unable to perform request to .*: connect: connection refused"))
							Expect(alertsResponse).To(BeNil())
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
									alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, mockTokenSource)
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
									alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, mockTokenSource)
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
									alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, mockTokenSource)
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
									alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, mockTokenSource)
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(MatchRegexp("unable to get alerts; unexpected response status code 500 from"))
									Expect(alertsResponse).To(BeNil())
								})
							})

							Context("with an unparsable response", func() {
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
									alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, mockTokenSource)
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
									alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, mockTokenSource)
									Expect(err).ToNot(HaveOccurred())
									Expect(alertsResponse).To(Equal(responseAlertsResponse))
								})
							})
						}

						When("the server responds directly to the one request", func() {
							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(HaveLen(1))
							})
							requestAssertions()
						})

						When("the server responds with unauthorized, the token is expired and the request retried", func() {
							BeforeEach(func() {
								mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(httpClient, nil)
								mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil)
								mockTokenSource.EXPECT().ExpireToken(gomock.Not(gomock.Nil())).Return(true, nil)
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
									alertsResponse, err := clnt.GetAlerts(ctx, startTime, endTime, mockTokenSource)
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

					It("returns error when token source is missing", func() {
						calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, nil)
						Expect(err).To(MatchError("unable to get calibrations; token source is missing"))
						Expect(calibrationsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when context is missing", func() {
						calibrationsResponse, err := clnt.GetCalibrations(context.Context(nil), startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError("unable to get calibrations; context is missing"))
						Expect(calibrationsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when token source returns an error", func() {
						responseErr := errorsTest.RandomError()
						mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(nil, responseErr)
						calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError(fmt.Sprintf("unable to get calibrations; %s", responseErr)))
						Expect(calibrationsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when token source returns that indicates an oauth token failure", func() {
						responseErr := errors.New(`oauth2: "invalid_grant"`)
						mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(nil, responseErr)
						calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError(`unable to get calibrations; oauth2: "invalid_grant"; authentication token is invalid`))
						Expect(calibrationsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					When("token source returns successfully", func() {
						var httpClient *http.Client

						BeforeEach(func() {
							httpClient = http.DefaultClient
							mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(httpClient, nil)
							mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil)
						})

						It("returns error when the server is not reachable", func() {
							server.Close()
							server = nil
							calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, mockTokenSource)
							Expect(err.Error()).To(MatchRegexp("unable to get calibrations; unable to perform request to .*: connect: connection refused"))
							Expect(calibrationsResponse).To(BeNil())
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
									calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, mockTokenSource)
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
									calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, mockTokenSource)
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
									calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, mockTokenSource)
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
									calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, mockTokenSource)
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(MatchRegexp("unable to get calibrations; unexpected response status code 500 from"))
									Expect(calibrationsResponse).To(BeNil())
								})
							})

							Context("with an unparsable response", func() {
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
									calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, mockTokenSource)
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
									calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, mockTokenSource)
									Expect(err).ToNot(HaveOccurred())
									Expect(calibrationsResponse).To(Equal(responseCalibrationsResponse))
								})
							})
						}

						When("the server responds directly to the one request", func() {
							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(HaveLen(1))
							})

							requestAssertions()
						})

						When("the server responds with unauthorized, the token is expired and the request retried", func() {
							BeforeEach(func() {
								mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(httpClient, nil)
								mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil)
								mockTokenSource.EXPECT().ExpireToken(gomock.Not(gomock.Nil())).Return(true, nil)
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
									calibrationsResponse, err := clnt.GetCalibrations(ctx, startTime, endTime, mockTokenSource)
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

					It("returns error when token source is missing", func() {
						devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, nil)
						Expect(err).To(MatchError("unable to get devices; token source is missing"))
						Expect(devicesResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when context is missing", func() {
						devicesResponse, err := clnt.GetDevices(context.Context(nil), startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError("unable to get devices; context is missing"))
						Expect(devicesResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when token source returns an error", func() {
						responseErr := errorsTest.RandomError()
						mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(nil, responseErr)
						devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError(fmt.Sprintf("unable to get devices; %s", responseErr)))
						Expect(devicesResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when token source returns that indicates an oauth token failure", func() {
						responseErr := errors.New(`oauth2: "invalid_grant"`)
						mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(nil, responseErr)
						devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError(`unable to get devices; oauth2: "invalid_grant"; authentication token is invalid`))
						Expect(devicesResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					When("token source returns successfully", func() {
						var httpClient *http.Client

						BeforeEach(func() {
							httpClient = http.DefaultClient
							mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(httpClient, nil)
							mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil)
						})

						It("returns error when the server is not reachable", func() {
							server.Close()
							server = nil
							devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, mockTokenSource)
							Expect(err.Error()).To(MatchRegexp("unable to get devices; unable to perform request to .*: connect: connection refused"))
							Expect(devicesResponse).To(BeNil())
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
									devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, mockTokenSource)
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
									devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, mockTokenSource)
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
									devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, mockTokenSource)
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
									devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, mockTokenSource)
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(MatchRegexp("unable to get devices; unexpected response status code 500 from"))
									Expect(devicesResponse).To(BeNil())
								})
							})

							Context("with an unparsable response", func() {
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
									devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, mockTokenSource)
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
									devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, mockTokenSource)
									Expect(err).ToNot(HaveOccurred())
									Expect(devicesResponse).To(Equal(responseDevicesResponse))
								})
							})
						}

						When("the server responds directly to the one request", func() {
							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(HaveLen(1))
							})

							requestAssertions()
						})

						When("the server responds with unauthorized, the token is expired and the request retried", func() {
							BeforeEach(func() {
								mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(httpClient, nil)
								mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil)
								mockTokenSource.EXPECT().ExpireToken(gomock.Not(gomock.Nil())).Return(true, nil)
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
									devicesResponse, err := clnt.GetDevices(ctx, startTime, endTime, mockTokenSource)
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

					It("returns error when token source is missing", func() {
						egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, nil)
						Expect(err).To(MatchError("unable to get egvs; token source is missing"))
						Expect(egvsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when context is missing", func() {
						egvsResponse, err := clnt.GetEGVs(context.Context(nil), startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError("unable to get egvs; context is missing"))
						Expect(egvsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when token source returns an error", func() {
						responseErr := errorsTest.RandomError()
						mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(nil, responseErr)
						egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError(fmt.Sprintf("unable to get egvs; %s", responseErr)))
						Expect(egvsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when token source returns that indicates an oauth token failure", func() {
						responseErr := errors.New(`oauth2: "invalid_grant"`)
						mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(nil, responseErr)
						egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError(`unable to get egvs; oauth2: "invalid_grant"; authentication token is invalid`))
						Expect(egvsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					When("token source returns successfully", func() {
						var httpClient *http.Client

						BeforeEach(func() {
							httpClient = http.DefaultClient
							mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(httpClient, nil)
							mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil)
						})

						It("returns error when the server is not reachable", func() {
							server.Close()
							server = nil
							egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, mockTokenSource)
							Expect(err.Error()).To(MatchRegexp("unable to get egvs; unable to perform request to .*: connect: connection refused"))
							Expect(egvsResponse).To(BeNil())
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
									egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, mockTokenSource)
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
									egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, mockTokenSource)
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
									egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, mockTokenSource)
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
									egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, mockTokenSource)
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(MatchRegexp("unable to get egvs; unexpected response status code 500 from"))
									Expect(egvsResponse).To(BeNil())
								})
							})

							Context("with an unparsable response", func() {
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
									egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, mockTokenSource)
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
									egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, mockTokenSource)
									Expect(err).ToNot(HaveOccurred())
									Expect(egvsResponse).To(Equal(responseEGVsResponse))
								})
							})
						}

						When("the server responds directly to the one request", func() {
							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(HaveLen(1))
							})

							requestAssertions()
						})

						When("the server responds with unauthorized, the token is expired and the request retried", func() {
							BeforeEach(func() {
								mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(httpClient, nil)
								mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil)
								mockTokenSource.EXPECT().ExpireToken(gomock.Not(gomock.Nil())).Return(true, nil)
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
									egvsResponse, err := clnt.GetEGVs(ctx, startTime, endTime, mockTokenSource)
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

					It("returns error when token source is missing", func() {
						eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, nil)
						Expect(err).To(MatchError("unable to get events; token source is missing"))
						Expect(eventsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when context is missing", func() {
						eventsResponse, err := clnt.GetEvents(context.Context(nil), startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError("unable to get events; context is missing"))
						Expect(eventsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when token source returns an error", func() {
						responseErr := errorsTest.RandomError()
						mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(nil, responseErr)
						eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError(fmt.Sprintf("unable to get events; %s", responseErr)))
						Expect(eventsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when token source returns that indicates an oauth token failure", func() {
						responseErr := errors.New(`oauth2: "invalid_grant"`)
						mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(nil, responseErr)
						eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, mockTokenSource)
						Expect(err).To(MatchError(`unable to get events; oauth2: "invalid_grant"; authentication token is invalid`))
						Expect(eventsResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					When("token source returns successfully", func() {
						var httpClient *http.Client

						BeforeEach(func() {
							httpClient = http.DefaultClient
							mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(httpClient, nil)
							mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil)
						})

						It("returns error when the server is not reachable", func() {
							server.Close()
							server = nil
							eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, mockTokenSource)
							Expect(err.Error()).To(MatchRegexp("unable to get events; unable to perform request to .*: connect: connection refused"))
							Expect(eventsResponse).To(BeNil())
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
									eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, mockTokenSource)
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
									eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, mockTokenSource)
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
									eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, mockTokenSource)
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
									eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, mockTokenSource)
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(MatchRegexp("unable to get events; unexpected response status code 500 from"))
									Expect(eventsResponse).To(BeNil())
								})
							})

							Context("with an unparsable response", func() {
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
									eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, mockTokenSource)
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
									eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, mockTokenSource)
									Expect(err).ToNot(HaveOccurred())
									Expect(eventsResponse).To(Equal(responseEventsResponse))
								})
							})
						}

						When("the server responds directly to the one request", func() {
							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(HaveLen(1))
							})

							requestAssertions()
						})

						When("the server responds with unauthorized, the token is expired and the request retried", func() {
							BeforeEach(func() {
								mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(httpClient, nil)
								mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil)
								mockTokenSource.EXPECT().ExpireToken(gomock.Not(gomock.Nil())).Return(true, nil)
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
									eventsResponse, err := clnt.GetEvents(ctx, startTime, endTime, mockTokenSource)
									Expect(err).To(MatchError("unable to get events; authentication token is invalid"))
									Expect(eventsResponse).To(BeNil())
								})
							})
						})
					})
				})
			})

			Context("with started server and new client with a retrier", func() {
				const retries = 2

				var server *Server
				var responseHeaders http.Header
				var ctx context.Context
				var mockTokenSource *oauthTest.MockTokenSource
				var clnt *dexcomClient.Client

				BeforeEach(func() {
					server = NewServer()
					responseHeaders = http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
					ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
					mockTokenSource = oauthTest.NewMockTokenSource(mockController)
				})

				JustBeforeEach(func() {
					config.Address = server.URL()
					var err error
					clnt, err = dexcomClient.New(config, nil, mockTokenSourceSource, request.NewRetrier(retries, time.Millisecond, 0))
					Expect(err).ToNot(HaveOccurred())
					Expect(clnt).ToNot(BeNil())
				})

				AfterEach(func() {
					if server != nil {
						server.Close()
					}
				})

				DescribeTable("retries only a transient failure",
					func(statusCode int, responseBody any, expectedRequests int, expectedError string) {
						for range expectedRequests {
							server.AppendHandlers(RespondWith(statusCode, responseBody, responseHeaders))
						}
						mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(http.DefaultClient, nil).Times(expectedRequests)
						mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil).Times(expectedRequests)
						dataRangeResponse, err := clnt.GetDataRange(ctx, nil, mockTokenSource)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring(expectedError))
						Expect(dataRangeResponse).To(BeNil())
						Expect(server.ReceivedRequests()).To(HaveLen(expectedRequests))
					},
					Entry("with too many requests 429", http.StatusTooManyRequests, "NOT JSON", retries+1, "too many requests"),
					Entry("with an internal server error 500", http.StatusInternalServerError, "NOT JSON", retries+1, "unexpected response status code 500"),
					Entry("with a service unavailable 503", http.StatusServiceUnavailable, "NOT JSON", retries+1, "unexpected response status code 503"),
					Entry("with a bad request 400", http.StatusBadRequest, "NOT JSON", 1, "bad request"),
					Entry("with a forbidden response 403", http.StatusForbidden, "NOT JSON", 1, "authentication token is not authorized for requested action"),
					Entry("with a resource not found 404", http.StatusNotFound, "NOT JSON", 1, "resource not found"),
					Entry("with a conflict 409", http.StatusConflict, "NOT JSON", 1, "unexpected response status code 409"),
					Entry("with an unparsable response", http.StatusOK, "{", 1, "json is malformed"),
				)

				It("does not retry an unauthorized response 401 beyond the token refresh within the oauth client", func() {
					for range 2 {
						server.AppendHandlers(RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders))
					}
					mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(http.DefaultClient, nil).Times(2)
					mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil).Times(2)
					mockTokenSource.EXPECT().ExpireToken(gomock.Not(gomock.Nil())).Return(true, nil)
					dataRangeResponse, err := clnt.GetDataRange(ctx, nil, mockTokenSource)
					Expect(err).To(MatchError("unable to get data range; authentication token is invalid"))
					Expect(dataRangeResponse).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(2))
				})

				It("retries when the server is not reachable", func() {
					server.Close()
					server = nil
					mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(http.DefaultClient, nil).Times(retries + 1)
					mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil).Times(retries + 1)
					dataRangeResponse, err := clnt.GetDataRange(ctx, nil, mockTokenSource)
					Expect(err.Error()).To(MatchRegexp("unable to get data range; unable to perform request to .*: connect: connection refused"))
					Expect(dataRangeResponse).To(BeNil())
				})

				It("does not retry a successful response", func() {
					responseDataRangesResponse := dexcomTest.RandomDataRangesResponse()
					server.AppendHandlers(RespondWith(http.StatusOK, test.MarshalResponseBody(responseDataRangesResponse), responseHeaders))
					mockTokenSource.EXPECT().HTTPClient(gomock.Not(gomock.Nil()), gomock.Eq(mockTokenSourceSource)).Return(http.DefaultClient, nil)
					mockTokenSource.EXPECT().UpdateToken(gomock.Not(gomock.Nil())).Return(true, nil)
					Expect(clnt.GetDataRange(ctx, nil, mockTokenSource)).To(Equal(responseDataRangesResponse))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
