package client_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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
	prometheusTest "github.com/tidepool-org/platform/prometheus/test"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Client", func() {
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
			clnt, err := dexcomClient.New(nil, mockTokenSourceSource)
			Expect(err).To(MatchError("config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error when config is invalid", func() {
			config.Address = ""
			clnt, err := dexcomClient.New(config, mockTokenSourceSource)
			Expect(err).To(MatchError("config is invalid; address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error when token source source is missing", func() {
			clnt, err := dexcomClient.New(config, nil)
			Expect(err).To(MatchError("token source source is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(dexcomClient.New(config, mockTokenSourceSource)).ToNot(BeNil())
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
			clnt, err = dexcomClient.New(config, mockTokenSourceSource)
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
	})

	It("RequestTimeHeaderName is expected", func() {
		Expect(dexcomClient.RequestTimeHeaderName).To(Equal("request-time"))
	})

	Context("PrometheusRequestMetricsRoundTripper", func() {
		Context("NewPrometheusRequestMetricsRoundTripper", func() {
			It("returns successfully", func() {
				roundTripper := dexcomClient.NewPrometheusRequestMetricsRoundTripper(prometheusTest.RandomMetricName(), prometheusTest.RandomMetricHelp())
				Expect(roundTripper).ToNot(BeNil())
				Expect(roundTripper.PrometheusRequestMetricsRoundTripper).ToNot(BeNil())
			})
		})

		Context("RoundTrip", func() {
			var testRoundTripper *testHttp.RoundTripper
			var name string
			var roundTripper *dexcomClient.PrometheusRequestMetricsRoundTripper
			var request *http.Request

			BeforeEach(func() {
				testRoundTripper = testHttp.NewRoundTripper()
				name = prometheusTest.RandomMetricName()
				roundTripper = dexcomClient.NewPrometheusRequestMetricsRoundTripper(name, prometheusTest.RandomMetricHelp())
				roundTripper.WithRoundTripper(testRoundTripper)
				request = testHttp.NewRequest()
			})

			It("returns the response from the resolved round tripper", func() {
				testRoundTripper.Response = &http.Response{StatusCode: testHttp.NewStatusCode()}

				result := test.Must(roundTripper.RoundTrip(request))
				Expect(result).To(BeIdenticalTo(testRoundTripper.Response))
				Expect(testRoundTripper.Request).To(BeIdenticalTo(request))
			})

			It("returns the error from the resolved round tripper", func() {
				testErr := errorsTest.RandomError()
				testRoundTripper.Error = testErr

				result, err := roundTripper.RoundTrip(request)
				Expect(err).To(Equal(testErr))
				Expect(result).To(BeNil())
			})

			It("does not record a request time metric when the resolved round tripper returns an error", func() {
				testRoundTripper.Error = errorsTest.RandomError()

				_, _ = roundTripper.RoundTrip(request)

				Expect(prometheusTest.MetricFamilyFromName(name + "_request_time_seconds")).To(BeNil())
			})

			It("records a request time metric when the response has a valid request-time header", func() {
				statusCode := testHttp.NewStatusCode()
				requestTime := time.Duration(test.RandomIntFromRange(1, 60*1000)) * time.Millisecond
				header := http.Header{}
				header.Set(dexcomClient.RequestTimeHeaderName, requestTime.String())
				testRoundTripper.Response = &http.Response{StatusCode: statusCode, Header: header}

				_ = test.Must(roundTripper.RoundTrip(request))

				family := prometheusTest.MetricFamilyFromName(name + "_request_time_seconds")
				Expect(family).ToNot(BeNil())
				Expect(family.GetMetric()).To(HaveLen(1))
				metric := family.GetMetric()[0]
				Expect(metric.GetHistogram().GetSampleCount()).To(Equal(uint64(1)))
				Expect(metric.GetHistogram().GetSampleSum()).To(Equal(requestTime.Seconds()))
				Expect(prometheusTest.LabelPairsToMap(metric.GetLabel())).To(Equal(map[string]string{
					client.PrometheusLabelNameMethod: request.Method,
					client.PrometheusLabelNamePath:   request.URL.Path,
					client.PrometheusLabelNameStatus: strconv.Itoa(statusCode),
				}))
			})

			It("does not record a request time metric when the response does not have a request-time header", func() {
				testRoundTripper.Response = &http.Response{StatusCode: testHttp.NewStatusCode(), Header: http.Header{}}

				_ = test.Must(roundTripper.RoundTrip(request))

				Expect(prometheusTest.MetricFamilyFromName(name + "_request_time_seconds")).To(BeNil())
			})

			It("does not record a request time metric when the request-time header is not a valid duration", func() {
				header := http.Header{}
				header.Set(dexcomClient.RequestTimeHeaderName, test.RandomString())
				testRoundTripper.Response = &http.Response{StatusCode: testHttp.NewStatusCode(), Header: header}

				_ = test.Must(roundTripper.RoundTrip(request))

				Expect(prometheusTest.MetricFamilyFromName(name + "_request_time_seconds")).To(BeNil())
			})
		})
	})
})
