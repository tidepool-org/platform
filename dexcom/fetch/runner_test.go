package fetch_test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"
	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/data"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/dexcom"
	dexcomFetch "github.com/tidepool-org/platform/dexcom/fetch"
	dexcomFetchTest "github.com/tidepool-org/platform/dexcom/fetch/test"
	dexcomTest "github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/task"
)

var _ = Describe("Runner", func() {
	var authClient *dexcomFetchTest.MockAuthClient
	var dataClient *dexcomFetchTest.MockDataClient
	var dataSourceClient *dexcomFetchTest.MockDataSourceClient
	var dexcomClient *dexcomFetchTest.MockDexcomClient

	BeforeEach(func() {
		authClient = dexcomFetchTest.NewMockAuthClient(gomock.NewController(GinkgoT()))
		dataClient = dexcomFetchTest.NewMockDataClient(gomock.NewController(GinkgoT()))
		dataSourceClient = dexcomFetchTest.NewMockDataSourceClient(gomock.NewController(GinkgoT()))
		dexcomClient = dexcomFetchTest.NewMockDexcomClient(gomock.NewController(GinkgoT()))
	})

	Context("NewRunner", func() {
		It("returns an error if the auth client is missing", func() {
			runner, err := dexcomFetch.NewRunner(nil, dataClient, dataSourceClient, dexcomClient)
			Expect(err).To(MatchError("auth client is missing"))
			Expect(runner).To(BeNil())
		})

		It("returns an error if the data client is missing", func() {
			runner, err := dexcomFetch.NewRunner(authClient, nil, dataSourceClient, dexcomClient)
			Expect(err).To(MatchError("data client is missing"))
			Expect(runner).To(BeNil())
		})

		It("returns an error if the data source client is missing", func() {
			runner, err := dexcomFetch.NewRunner(authClient, dataClient, nil, dexcomClient)
			Expect(err).To(MatchError("data source client is missing"))
			Expect(runner).To(BeNil())
		})

		It("returns an error if the dexcom client is missing", func() {
			runner, err := dexcomFetch.NewRunner(authClient, dataClient, dataSourceClient, nil)
			Expect(err).To(MatchError("dexcom client is missing"))
			Expect(runner).To(BeNil())
		})

		It("succeeds", func() {
			runner, err := dexcomFetch.NewRunner(authClient, dataClient, dataSourceClient, dexcomClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(runner).ToNot(BeNil())
		})
	})

	Context("with runner", func() {
		var runner *dexcomFetch.Runner

		BeforeEach(func() {
			var err error
			runner, err = dexcomFetch.NewRunner(authClient, dataClient, dataSourceClient, dexcomClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(runner).ToNot(BeNil())
		})

		It("returns the auth client", func() {
			Expect(runner.AuthClient()).To(Equal(authClient))
		})

		It("returns the data client", func() {
			Expect(runner.DataClient()).To(Equal(dataClient))
		})

		It("returns the data source client", func() {
			Expect(runner.DataSourceClient()).To(Equal(dataSourceClient))
		})

		It("returns the dexcom client", func() {
			Expect(runner.DexcomClient()).To(Equal(dexcomClient))
		})

		It("returns the runner type", func() {
			Expect(runner.GetRunnerType()).To(Equal("org.tidepool.oauth.dexcom.fetch"))
		})

		It("returns the runner deadline", func() {
			Expect(runner.GetRunnerDeadline()).Should(BeTemporally("~", time.Now().Add(45*time.Minute), time.Second))
		})

		It("returns the runner timeout", func() {
			Expect(runner.GetRunnerTimeout()).To(Equal(30 * time.Minute))
		})

		It("returns the runner duration maximum", func() {
			Expect(runner.GetRunnerDurationMaximum()).To(Equal(15 * time.Minute))
		})

		Context("with context", func() {
			var logger *logTest.Logger
			var ctx context.Context

			BeforeEach(func() {
				logger = logTest.NewLogger()
				ctx = log.NewContextWithLogger(context.Background(), logger)
			})

			Context("Run", func() {
				It("logs a warning if the task is missing", func() {
					runner.Run(ctx, nil)
					logger.AssertWarn("Unable to create task runner")
				})

			})
		})
	})

	Context("with provider and task", func() {
		var provider *dexcomFetchTest.MockProvider
		var tsk *task.Task

		BeforeEach(func() {
			provider = dexcomFetchTest.NewMockProvider(gomock.NewController(GinkgoT()))
			provider.EXPECT().AuthClient().Return(authClient).AnyTimes()
			provider.EXPECT().DataClient().Return(dataClient).AnyTimes()
			provider.EXPECT().DataSourceClient().Return(dataSourceClient).AnyTimes()
			provider.EXPECT().DexcomClient().Return(dexcomClient).AnyTimes()
			provider.EXPECT().GetRunnerDurationMaximum().Return(time.Second).AnyTimes()
			tsk = &task.Task{
				State: task.TaskStateRunning,
				Data: map[string]any{
					dexcom.DataKeyDataSourceID:      "test-data-source-id",
					dexcom.DataKeyProviderSessionID: "test-provider-session-id",
					dexcom.DataKeyDeviceHashes: map[string]any{
						"test-device-1": "test-device-hash-1",
						"test-device-2": "test-device-hash-2",
					},
				},
			}
		})

		Context("NewTaskRunner", func() {
			It("returns an error if the provider is missing", func() {
				taskRunner, err := dexcomFetch.NewTaskRunner(nil, tsk)
				Expect(err).To(MatchError("provider is missing"))
				Expect(taskRunner).To(BeNil())
			})

			It("returns an error if the task is missing", func() {
				taskRunner, err := dexcomFetch.NewTaskRunner(provider, nil)
				Expect(err).To(MatchError("task is missing"))
				Expect(taskRunner).To(BeNil())
			})

			It("succeeds", func() {
				taskRunner, err := dexcomFetch.NewTaskRunner(provider, tsk)
				Expect(err).ToNot(HaveOccurred())
				Expect(taskRunner).ToNot(BeNil())
			})
		})

		Context("with task runner and context", func() {
			var taskRunner *dexcomFetch.TaskRunner
			var ctx context.Context

			BeforeEach(func() {
				var err error
				taskRunner, err = dexcomFetch.NewTaskRunner(provider, tsk)
				Expect(err).ToNot(HaveOccurred())
				Expect(taskRunner).ToNot(BeNil())
				ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			})

			assertTaskState := func(state string) {
				Expect(tsk.State).To(Equal(state))

				if state == task.TaskStatePending {
					Expect(tsk.AvailableTime).ToNot(BeNil())
					Expect(*tsk.AvailableTime).To(BeTemporally(">", time.Now()))
				} else {
					Expect(tsk.AvailableTime).To(BeNil())
				}
			}

			assertTaskRetryCount := func(retryCount int) {
				Expect(tsk.Data[dexcom.DataKeyRetryCount]).To(Equal(retryCount))
			}

			assertTaskRetryCountNotPresent := func() {
				Expect(tsk.Data[dexcom.DataKeyRetryCount]).To(BeNil())
			}

			assertTaskError := func(code string, description string) {
				Expect(tsk.HasError()).To(BeTrue())
				Expect(errors.Code(errors.Last(tsk.GetError()))).To(Equal(code))
				Expect(errors.Last(tsk.GetError())).To(MatchError(ContainSubstring(description)))
			}

			assertTaskErrorMissing := func() {
				Expect(tsk.HasError()).To(BeFalse())
			}

			It("fails if data is missing", func() {
				tsk.Data = nil
				taskRunner.Run(ctx)
				assertTaskState(task.TaskStateFailed)
				assertTaskRetryCountNotPresent()
				assertTaskError(dexcomFetch.ErrorCodeInvalidState, "data is missing")
			})

			It("fails if data is empty", func() {
				tsk.Data = map[string]any{}
				taskRunner.Run(ctx)
				assertTaskState(task.TaskStateFailed)
				assertTaskRetryCountNotPresent()
				assertTaskError(dexcomFetch.ErrorCodeInvalidState, "data is missing")
			})

			It("fails if data source id is missing", func() {
				delete(tsk.Data, dexcom.DataKeyDataSourceID)
				taskRunner.Run(ctx)
				assertTaskState(task.TaskStateFailed)
				assertTaskRetryCountNotPresent()
				assertTaskError(dexcomFetch.ErrorCodeInvalidState, "data source id is missing")
			})

			It("fails if data source id is empty", func() {
				tsk.Data[dexcom.DataKeyDataSourceID] = ""
				taskRunner.Run(ctx)
				assertTaskState(task.TaskStateFailed)
				assertTaskRetryCountNotPresent()
				assertTaskError(dexcomFetch.ErrorCodeInvalidState, "data source id is missing")
			})

			It("fails if getting the data source fails", func() {
				testErr := errorsTest.RandomError()
				dataSourceClient.EXPECT().Get(matchContext(), "test-data-source-id").Return(nil, testErr).Times(1)
				taskRunner.Run(ctx)
				assertTaskState(task.TaskStatePending)
				assertTaskRetryCountNotPresent()
				assertTaskError(dexcomFetch.ErrorCodeResourceFailure, "unable to get data source")
			})

			It("fails if the data source is missing", func() {
				dataSourceClient.EXPECT().Get(matchContext(), "test-data-source-id").Return(nil, nil).Times(1)
				taskRunner.Run(ctx)
				assertTaskState(task.TaskStateFailed)
				assertTaskRetryCountNotPresent()
				assertTaskError(dexcomFetch.ErrorCodeInvalidState, "data source is missing")
			})

			Context("with data source", func() {
				var dataSrc *dataSource.Source

				BeforeEach(func() {
					dataSrc = &dataSource.Source{
						ID:                pointer.FromString("test-data-source-id"),
						ProviderSessionID: pointer.FromString("test-provider-session-id"),
						State:             pointer.FromString(dataSource.StateConnected),
					}
					dataSourceClient.EXPECT().Get(matchContext(), "test-data-source-id").Return(dataSrc, nil).Times(1)
				})

				assertTaskAndDataSourceState := func(state string) {
					assertTaskState(state)

					Expect(dataSrc.State).ToNot(BeNil())
					if state == task.TaskStatePending {
						Expect(*dataSrc.State).To(Equal(dataSource.StateConnected))
					} else {
						Expect(*dataSrc.State).To(Equal(dataSource.StateError))
					}
				}

				assertTaskAndDataSourceError := func(code string, description string) {
					assertTaskError(code, description)

					Expect(dataSrc.HasError()).To(BeTrue())
					Expect(errors.Last(dataSrc.GetError())).To(MatchError(ContainSubstring(description)))
				}

				assertTaskAndDataSourceErrorNotPresent := func() {
					assertTaskErrorMissing()

					Expect(tsk.HasError()).To(BeFalse())
				}

				It("fails if provider session id is missing and update data source returns an error", func() {
					testErr := errorsTest.RandomError()
					delete(tsk.Data, dexcom.DataKeyProviderSessionID)
					dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).Return(nil, testErr).Times(1)
					taskRunner.Run(ctx)
					assertTaskState(task.TaskStatePending)
					assertTaskRetryCountNotPresent()
					assertTaskError(dexcomFetch.ErrorCodeResourceFailure, "unable to update data source")
				})

				It("fails if provider session id is missing", func() {
					delete(tsk.Data, dexcom.DataKeyProviderSessionID)
					dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(1)
					taskRunner.Run(ctx)
					assertTaskAndDataSourceState(task.TaskStateFailed)
					assertTaskRetryCountNotPresent()
					assertTaskAndDataSourceError(dexcomFetch.ErrorCodeInvalidState, "provider session id is missing")
				})

				It("fails if provider session id is empty", func() {
					tsk.Data[dexcom.DataKeyProviderSessionID] = ""
					dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(1)
					taskRunner.Run(ctx)
					assertTaskAndDataSourceState(task.TaskStateFailed)
					assertTaskRetryCountNotPresent()
					assertTaskAndDataSourceError(dexcomFetch.ErrorCodeInvalidState, "provider session id is missing")
				})

				It("fails if getting the provider session fails", func() {
					testErr := errorsTest.RandomError()
					authClient.EXPECT().GetProviderSession(matchContext(), "test-provider-session-id").Return(nil, testErr).Times(1)
					dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(1)
					taskRunner.Run(ctx)
					assertTaskAndDataSourceState(task.TaskStatePending)
					assertTaskRetryCountNotPresent()
					assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, "unable to get provider session")
				})

				It("fails if the provider session is missing", func() {
					authClient.EXPECT().GetProviderSession(matchContext(), "test-provider-session-id").Return(nil, nil).Times(1)
					dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(1)
					taskRunner.Run(ctx)
					assertTaskAndDataSourceState(task.TaskStateFailed)
					assertTaskRetryCountNotPresent()
					assertTaskAndDataSourceError(dexcomFetch.ErrorCodeInvalidState, "provider session is missing")
				})

				Context("with provider session", func() {
					var oauthToken *oauth.Token
					var providerSession *auth.ProviderSession

					BeforeEach(func() {
						oauthToken = &oauth.Token{
							AccessToken:    "test-access-token-1",
							TokenType:      "Bearer",
							RefreshToken:   "test-refresh-token-1",
							ExpirationTime: time.Now().Add(time.Minute),
						}
						providerSession = &auth.ProviderSession{
							ID:         "test-provider-session-id",
							UserID:     "test-user-id",
							OAuthToken: oauthToken,
						}
						authClient.EXPECT().GetProviderSession(matchContext(), "test-provider-session-id").Return(providerSession, nil).Times(1)
						dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(1)
					})

					assertProviderSessionRefreshedTimes := func(times int) {
						Expect(strings.Count(providerSession.OAuthToken.RefreshToken, "*")).To(Equal(times))
					}

					assertProviderSessionNotRefreshed := func() {
						assertProviderSessionRefreshedTimes(0)
					}

					It("fails if provider session oauth token is missing", func() {
						providerSession.OAuthToken = nil
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStateFailed)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeInvalidState, "token is missing")
					})

					It("fails if device hashes is invalid", func() {
						tsk.Data[dexcom.DataKeyDeviceHashes] = true
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStateFailed)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeInvalidState, "device hashes is invalid")
						assertProviderSessionNotRefreshed()
					})

					It("fails if a device hash is invalid", func() {
						tsk.Data[dexcom.DataKeyDeviceHashes] = map[string]any{"invalid-device-hash": true}
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStateFailed)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeInvalidState, "device hash is invalid")
						assertProviderSessionNotRefreshed()
					})

					It("fails if get data ranges returns a general error", func() {
						testErr := errorsTest.RandomError()
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(nil, nil, testErr)).Times(1)
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStatePending)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, testErr.Error())
						assertProviderSessionNotRefreshed()
					})

					It("fails if get data ranges returns a general error with latest data time", func() {
						latestDataTime := pointer.FromTime(time.Now().Add(-Day))
						dataSrc.LatestDataTime = latestDataTime
						testErr := errorsTest.RandomError()
						dexcomClient.EXPECT().GetDataRange(matchContext(), latestDataTime, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(nil, nil, testErr)).Times(1)
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStatePending)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, testErr.Error())
						assertProviderSessionNotRefreshed()
					})

					It("fails if get data ranges refreshes the token and returns an error when updating the provider session", func() {
						testErr := errorsTest.RandomError()
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(&MockTokenSource{Refresh: true}, nil, nil)).Times(1)
						authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).Return(nil, testErr).Times(1)
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStatePending)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, "unable to update provider session")
						assertProviderSessionNotRefreshed()
					})

					It("fails if get data ranges refreshes the token and returns no provider session when updating the provider session", func() {
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(&MockTokenSource{Refresh: true}, nil, nil)).Times(1)
						authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).Return(nil, nil).Times(1)
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStateFailed)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeInvalidState, "provider session is missing")
						assertProviderSessionNotRefreshed()
					})

					It("fails if get data ranges refreshes the token and returns a general error", func() {
						testErr := errorsTest.RandomError()
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(&MockTokenSource{Refresh: true}, nil, testErr)).Times(1)
						authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(1)
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStatePending)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, testErr.Error())
						assertProviderSessionRefreshedTimes(1)
					})

					It("fails if get data ranges refreshes the token and returns an authentication error", func() {
						testErr := request.ErrorUnauthenticated()
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(&MockTokenSource{Refresh: true}, nil, testErr)).Times(1)
						authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(1)
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStatePending)
						assertTaskRetryCount(1)
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeAuthenticationFailure, testErr.Error())
						assertProviderSessionRefreshedTimes(1)
					})

					Context("with Dexcom data ranges response", func() {
						var startTime time.Time
						var endTime time.Time
						var dataRangeResponse *dexcom.DataRangesResponse

						BeforeEach(func() {
							startTime = time.Now().Add(-7 * Day)
							endTime = time.Now().Add(-3 * Day)
							dataRangeResponse = &dexcom.DataRangesResponse{
								Calibrations: &dexcom.DataRange{
									Start: &dexcom.Moment{SystemTime: &dexcom.Time{Time: startTime}},
									End:   &dexcom.Moment{SystemTime: &dexcom.Time{Time: endTime}},
								},
							}
							dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(&MockTokenSource{Refresh: true}, dataRangeResponse, nil)).Times(1)
							authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(1)
						})

						It("is successful if the Dexcom data ranges is not valid", func() {
							dataRangeResponse.Calibrations.Start = nil
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStatePending)
							assertTaskRetryCountNotPresent()
							assertTaskAndDataSourceErrorNotPresent()
							assertProviderSessionRefreshedTimes(1)
						})

						It("is successful if the Dexcom data ranges start is not before end", func() {
							dataRangeResponse.Calibrations.Start = &dexcom.Moment{SystemTime: &dexcom.Time{Time: time.Now().Add(-2 * Day)}}
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStatePending)
							assertTaskRetryCountNotPresent()
							assertTaskAndDataSourceErrorNotPresent()
							assertProviderSessionRefreshedTimes(1)
						})

						It("fails if get alerts returns a general error", func() {
							testErr := errorsTest.RandomError()
							dexcomClient.EXPECT().GetAlerts(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData[dexcom.AlertsResponse](nil, nil, testErr)).Times(1)
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStatePending)
							assertTaskRetryCountNotPresent()
							assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, testErr.Error())
							assertProviderSessionRefreshedTimes(1)
						})

						It("fails if get alerts refreshes the token and returns an error when updating the provider session", func() {
							testErr := errorsTest.RandomError()
							dexcomClient.EXPECT().GetAlerts(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData[dexcom.AlertsResponse](&MockTokenSource{Refresh: true}, nil, nil)).Times(1)
							authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).Return(nil, testErr).Times(1)
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStatePending)
							assertTaskRetryCountNotPresent()
							assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, "unable to update provider session")
							assertProviderSessionRefreshedTimes(1)
						})

						It("fails if get alerts refreshes the token and returns no provider session when updating the provider session", func() {
							dexcomClient.EXPECT().GetAlerts(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData[dexcom.AlertsResponse](&MockTokenSource{Refresh: true}, nil, nil)).Times(1)
							authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).Return(nil, nil).Times(1)
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStateFailed)
							assertTaskRetryCountNotPresent()
							assertTaskAndDataSourceError(dexcomFetch.ErrorCodeInvalidState, "provider session is missing")
							assertProviderSessionRefreshedTimes(1)
						})

						It("fails if get alerts refreshes the token and returns a general error", func() {
							testErr := errorsTest.RandomError()
							dexcomClient.EXPECT().GetAlerts(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData[dexcom.AlertsResponse](&MockTokenSource{Refresh: true}, nil, testErr)).Times(1)
							authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(1)
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStatePending)
							assertTaskRetryCountNotPresent()
							assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, testErr.Error())
							assertProviderSessionRefreshedTimes(2)
						})

						It("fails if get alerts refreshes the token and returns an authentication error", func() {
							testErr := request.ErrorUnauthenticated()
							dexcomClient.EXPECT().GetAlerts(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData[dexcom.AlertsResponse](&MockTokenSource{Refresh: true}, nil, testErr)).Times(1)
							authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(1)
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStatePending)
							assertTaskRetryCount(1)
							assertTaskAndDataSourceError(dexcomFetch.ErrorCodeAuthenticationFailure, testErr.Error())
							assertProviderSessionRefreshedTimes(2)
						})

						Context("with Dexcom data responses", func() {
							var alertsResponse *dexcom.AlertsResponse
							var calibrationsResponse *dexcom.CalibrationsResponse
							var devicesResponse *dexcom.DevicesResponse
							var egvsResponse *dexcom.EGVsResponse
							var eventsResponse *dexcom.EventsResponse

							BeforeEach(func() {
								alertsResponse = &dexcom.AlertsResponse{Records: &dexcom.Alerts{}}
								dexcomClient.EXPECT().GetAlerts(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, alertsResponse, nil)).Times(1)
								calibrationsResponse = &dexcom.CalibrationsResponse{Records: &dexcom.Calibrations{}}
								dexcomClient.EXPECT().GetCalibrations(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, calibrationsResponse, nil)).Times(1)
								devicesResponse = &dexcom.DevicesResponse{
									Records: &dexcom.Devices{
										{
											LastUploadDate:        &dexcom.Time{Time: time.Now().Add(-4 * Day)},
											AlertSchedules:        &dexcom.AlertSchedules{},
											TransmitterID:         pointer.FromString(dexcomTest.RandomTransmitterID()),
											TransmitterGeneration: pointer.FromString(dexcom.DeviceTransmitterGenerationG6),
											DisplayDevice:         pointer.FromString(dexcom.DeviceDisplayDeviceIOS),
											DisplayApp:            pointer.FromString(dexcom.DeviceDisplayAppG6),
										},
									},
								}
								dexcomClient.EXPECT().GetDevices(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, devicesResponse, nil)).Times(1)
								egvsResponse = &dexcom.EGVsResponse{Records: &dexcom.EGVs{}}
								dexcomClient.EXPECT().GetEGVs(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, egvsResponse, nil)).Times(1)
								eventsResponse = &dexcom.EventsResponse{Records: &dexcom.Events{}}
								dexcomClient.EXPECT().GetEvents(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, eventsResponse, nil)).Times(1)
							})

							assertTaskDeviceHashesCount := func(count int) {
								deviceHashesRaw, ok := tsk.Data[dexcom.DataKeyDeviceHashes]
								Expect(ok).To(BeTrue())
								Expect(deviceHashesRaw).ToNot(BeNil())
								deviceHashes, ok := deviceHashesRaw.(map[string]string)
								Expect(ok).To(BeTrue())
								Expect(len(deviceHashes)).To(Equal(count))
							}

							It("succeeds", func() {
								dataSet := &data.DataSet{
									ID:       pointer.FromString("test-data-set-id"),
									UploadID: pointer.FromString("test-data-set-upload-id"),
								}
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(3)
								dataClient.EXPECT().CreateUserDataSet(matchContext(), "test-user-id", matchNotNil()).DoAndReturn(mockDataClientCreateUserDataSet(dataSet, nil)).Times(1)
								dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(nil)).Times(1)
								taskRunner.Run(ctx)
								assertTaskAndDataSourceState(task.TaskStatePending)
								assertTaskDeviceHashesCount(3)
								assertTaskRetryCountNotPresent()
								assertTaskAndDataSourceErrorNotPresent()
								assertProviderSessionRefreshedTimes(1)
							})
						})
					})

					// ALTERNATES:
					// deviceHashes - not in data
					// dataSource.LatestDataTime - not nil (recent)
					// refresh token
					// data ranges multiple 30 day segments
				})
			})
		})
	})
})

func mockAuthClientUpdateProviderSession(providerSession *auth.ProviderSession) func(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error) {
	return func(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error) {
		providerSession.OAuthToken = update.OAuthToken
		return providerSession, nil
	}
}

func mockDataClientCreateUserDataSet(dataSet *data.DataSet, err error) func(ctx context.Context, userID string, create *data.DataSetCreate) (*data.DataSet, error) {
	return func(ctx context.Context, userID string, create *data.DataSetCreate) (*data.DataSet, error) {
		return dataSet, err
	}
}

func mockDataClientCreateDataSetsData(err error) func(ctx context.Context, dataSetID string, datumArray []data.Datum) error {
	return func(ctx context.Context, dataSetID string, datumArray []data.Datum) error {
		return err
	}
}

func mockDataSourceClientUpdate(dataSrc *dataSource.Source) func(context.Context, string, *request.Condition, *dataSource.Update) (*dataSource.Source, error) {
	localDataSrc := dataSrc
	return func(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error) {
		if update.ProviderSessionID != nil {
			localDataSrc.ProviderSessionID = update.ProviderSessionID
		}
		if update.State != nil {
			localDataSrc.State = update.State
		}
		if update.Error != nil {
			localDataSrc.Error = update.Error
		}
		if update.DataSetIDs != nil {
			localDataSrc.DataSetIDs = update.DataSetIDs
		}
		if update.EarliestDataTime != nil {
			localDataSrc.EarliestDataTime = update.EarliestDataTime
		}
		if update.LatestDataTime != nil {
			localDataSrc.LatestDataTime = update.LatestDataTime
		}
		if update.LastImportTime != nil {
			localDataSrc.LastImportTime = update.LastImportTime
		}
		return localDataSrc, nil
	}
}

func mockDexcomClientGetDataRange(mockTokenSource *MockTokenSource, response *dexcom.DataRangesResponse, err error) func(ctx context.Context, lastSyncTime *time.Time, tokenSource oauth.TokenSource) (*dexcom.DataRangesResponse, error) {
	if mockTokenSource == nil {
		mockTokenSource = &MockTokenSource{}
	}
	return func(ctx context.Context, lastSyncTime *time.Time, tokenSource oauth.TokenSource) (*dexcom.DataRangesResponse, error) {
		tokenSource.HTTPClient(ctx, mockTokenSource)
		return response, err
	}
}

func mockDexcomClientGetData[T any](mockTokenSource *MockTokenSource, response *T, err error) func(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*T, error) {
	if mockTokenSource == nil {
		mockTokenSource = &MockTokenSource{}
	}
	return func(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*T, error) {
		tokenSource.HTTPClient(ctx, mockTokenSource)
		return response, err
	}
}

type MockTokenSource struct {
	Refresh bool
	token   *oauth.Token
}

func (m *MockTokenSource) TokenSource(ctx context.Context, token *oauth.Token) (oauth2.TokenSource, error) {
	m.token = token
	return m, nil
}

func (m *MockTokenSource) Token() (*oauth2.Token, error) {
	if !m.Refresh {
		return m.token.RawToken(), nil
	} else {
		return &oauth2.Token{
			AccessToken:  fmt.Sprintf("%s*", m.token.AccessToken),
			TokenType:    m.token.TokenType,
			RefreshToken: fmt.Sprintf("%s*", m.token.RefreshToken),
			Expiry:       time.Now().Add(time.Minute),
		}, nil
	}
}

func matchContext() gomock.Matcher {
	return gomock.AssignableToTypeOf(reflect.TypeOf((*context.Context)(nil)).Elem())
}

func matchNotNil() gomock.Matcher {
	return gomock.Not(gomock.Nil())
}

func matchNil() gomock.Matcher {
	return gomock.Nil()
}

const Day = 24 * time.Hour
