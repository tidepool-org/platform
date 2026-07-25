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
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
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
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Runner", func() {
	var mockController *gomock.Controller
	var authClient *dexcomFetchTest.MockAuthClient
	var dataClient *dexcomFetchTest.MockDataClient
	var dataSourceClient *dataSourceTest.MockClient
	var dexcomClient *dexcomFetchTest.MockDexcomClient

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		authClient = dexcomFetchTest.NewMockAuthClient(mockController)
		dataClient = dexcomFetchTest.NewMockDataClient(mockController)
		dataSourceClient = dataSourceTest.NewMockClient(mockController)
		dexcomClient = dexcomFetchTest.NewMockDexcomClient(mockController)
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
			Expect(runner.GetRunnerDeadline()).To(Equal(45 * time.Minute))
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
		var runnerDurationMaximum time.Duration
		var tsk *task.Task

		BeforeEach(func() {
			provider = dexcomFetchTest.NewMockProvider(mockController)
			runnerDurationMaximum = time.Second
			provider.EXPECT().AuthClient().Return(authClient).AnyTimes()
			provider.EXPECT().DataClient().Return(dataClient).AnyTimes()
			provider.EXPECT().DataSourceClient().Return(dataSourceClient).AnyTimes()
			provider.EXPECT().DexcomClient().Return(dexcomClient).AnyTimes()
			provider.EXPECT().GetRunnerDurationMaximum().DoAndReturn(func() time.Duration { return runnerDurationMaximum }).AnyTimes()
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
			var logger *logTest.Logger
			var ctx context.Context

			BeforeEach(func() {
				var err error
				taskRunner, err = dexcomFetch.NewTaskRunner(provider, tsk)
				Expect(err).ToNot(HaveOccurred())
				Expect(taskRunner).ToNot(BeNil())
				logger = logTest.NewLogger()
				ctx = log.NewContextWithLogger(context.Background(), logger)
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

			assertTaskAvailableSoon := func() {
				Expect(tsk.AvailableTime).ToNot(BeNil())
				Expect(*tsk.AvailableTime).To(BeTemporally("~", time.Now().Add(time.Minute), 17*time.Second)) // 20% jitter + test execution time
			}

			assertTaskAvailableAfterStandardDuration := func() {
				Expect(tsk.AvailableTime).ToNot(BeNil())
				Expect(*tsk.AvailableTime).To(BeTemporally(">", time.Now().Add(dexcomFetch.AvailableAfterDuration-dexcomFetch.AvailableAfterDurationJitter-time.Second)))
				Expect(*tsk.AvailableTime).To(BeTemporally("<", time.Now().Add(dexcomFetch.AvailableAfterDuration+dexcomFetch.AvailableAfterDurationJitter)))
			}

			assertTaskRetryCount := func(retryCount int) {
				Expect(tsk.Data[dexcom.DataKeyRetryCount]).To(Equal(int32(retryCount)))
			}

			assertTaskRetryCountNotPresent := func() {
				Expect(tsk.Data[dexcom.DataKeyRetryCount]).To(BeNil())
			}

			setTaskResumeTime := func(resumeTime time.Time, resumeExpirationTime time.Time) {
				tsk.Data[dexcom.DataKeyResumeDataTime] = resumeTime.UTC().Format(time.RFC3339Nano)
				tsk.Data[dexcom.DataKeyResumeExpirationTime] = resumeExpirationTime.UTC().Format(time.RFC3339Nano)
			}

			assertTaskResumeTime := func(resumeTime time.Time) {
				Expect(tsk.Data[dexcom.DataKeyResumeDataTime]).To(Equal(resumeTime.UTC().Format(time.RFC3339Nano)))

				resumeExpirationTime, ok := tsk.Data[dexcom.DataKeyResumeExpirationTime].(string)
				Expect(ok).To(BeTrue())
				Expect(test.Must(time.Parse(time.RFC3339Nano, resumeExpirationTime))).To(BeTemporally("~", time.Now().Add(dexcomFetch.ResumeExpirationDuration), time.Minute))
			}

			assertTaskResumeTimeNotPresent := func() {
				Expect(tsk.Data[dexcom.DataKeyResumeDataTime]).To(BeNil())
				Expect(tsk.Data[dexcom.DataKeyResumeExpirationTime]).To(BeNil())
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
				dataSourceClient.EXPECT().Get(matchContext(), "test-data-source-id").Return(nil, testErr)
				taskRunner.Run(ctx)
				assertTaskState(task.TaskStatePending)
				assertTaskRetryCountNotPresent()
				assertTaskError(dexcomFetch.ErrorCodeResourceFailure, "unable to get data source")
			})

			It("fails if the data source is missing", func() {
				dataSourceClient.EXPECT().Get(matchContext(), "test-data-source-id").Return(nil, nil)
				taskRunner.Run(ctx)
				assertTaskState(task.TaskStateFailed)
				assertTaskRetryCountNotPresent()
				assertTaskError(dexcomFetch.ErrorCodeInvalidState, "data source is missing")
			})

			Context("with data source", func() {
				var dataSrc *dataSource.Source

				BeforeEach(func() {
					dataSrc = &dataSource.Source{
						ID:                "test-data-source-id",
						ProviderSessionID: pointer.FromString("test-provider-session-id"),
						State:             dataSource.StateConnected,
					}
					dataSourceClient.EXPECT().Get(matchContext(), "test-data-source-id").Return(dataSrc, nil)
				})

				assertTaskAndDataSourceState := func(state string) {
					assertTaskState(state)

					Expect(dataSrc.State).ToNot(BeNil())
					if state == task.TaskStatePending {
						Expect(dataSrc.State).To(Equal(dataSource.StateConnected))
					} else {
						Expect(dataSrc.State).To(Equal(dataSource.StateError))
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

				assertDataSourceLastImportTimePresent := func() {
					Expect(dataSrc.LastImportTime).ToNot(BeNil())
				}

				It("fails if provider session id is missing and update data source returns an error", func() {
					testErr := errorsTest.RandomError()
					delete(tsk.Data, dexcom.DataKeyProviderSessionID)
					dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).Return(nil, testErr)
					taskRunner.Run(ctx)
					assertTaskState(task.TaskStatePending)
					assertTaskRetryCountNotPresent()
					assertTaskError(dexcomFetch.ErrorCodeResourceFailure, "unable to update data source")
				})

				It("fails if provider session id is missing", func() {
					delete(tsk.Data, dexcom.DataKeyProviderSessionID)
					dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc))
					taskRunner.Run(ctx)
					assertTaskAndDataSourceState(task.TaskStateFailed)
					assertTaskRetryCountNotPresent()
					assertTaskAndDataSourceError(dexcomFetch.ErrorCodeInvalidState, "provider session id is missing")
				})

				It("fails if provider session id is empty", func() {
					tsk.Data[dexcom.DataKeyProviderSessionID] = ""
					dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc))
					taskRunner.Run(ctx)
					assertTaskAndDataSourceState(task.TaskStateFailed)
					assertTaskRetryCountNotPresent()
					assertTaskAndDataSourceError(dexcomFetch.ErrorCodeInvalidState, "provider session id is missing")
				})

				It("fails if getting the provider session fails", func() {
					testErr := errorsTest.RandomError()
					authClient.EXPECT().GetProviderSession(matchContext(), "test-provider-session-id").Return(nil, testErr)
					dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc))
					taskRunner.Run(ctx)
					assertTaskAndDataSourceState(task.TaskStatePending)
					assertTaskRetryCountNotPresent()
					assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, "unable to get provider session")
				})

				It("discards the run outcome if the task claim is lost", func() {
					claimContext, claimCancel := context.WithCancelCause(ctx)
					defer claimCancel(nil)
					testErr := errorsTest.RandomError()
					authClient.EXPECT().GetProviderSession(matchContext(), "test-provider-session-id").DoAndReturn(func(ctx context.Context, id string) (*auth.ProviderSession, error) {
						claimCancel(task.ErrClaimLost)
						return nil, testErr
					})
					taskRunner.Run(claimContext)
					assertTaskState(task.TaskStateRunning)
					Expect(dataSrc.State).To(Equal(dataSource.StateConnected))
					Expect(dataSrc.HasError()).To(BeFalse())
					logger.AssertWarn("Skipped updating data source and task because the task claim was lost")
				})

				It("fails if the provider session is missing", func() {
					authClient.EXPECT().GetProviderSession(matchContext(), "test-provider-session-id").Return(nil, nil)
					dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc))
					taskRunner.Run(ctx)
					assertTaskAndDataSourceState(task.TaskStateFailed)
					assertTaskRetryCountNotPresent()
					assertTaskAndDataSourceError(dexcomFetch.ErrorCodeInvalidState, "provider session is missing")
				})

				Context("with provider session", func() {
					var oauthToken *auth.OAuthToken
					var providerSession *auth.ProviderSession

					BeforeEach(func() {
						oauthToken = &auth.OAuthToken{
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
						authClient.EXPECT().GetProviderSession(matchContext(), "test-provider-session-id").Return(providerSession, nil)
						dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc))
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
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(nil, nil, testErr))
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStatePending)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, testErr.Error())
						assertProviderSessionNotRefreshed()
					})

					It("fails if get data ranges returns no response and no error", func() {
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(nil, nil, nil))
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStatePending)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, "data ranges response is missing")
						assertProviderSessionNotRefreshed()
					})

					It("fails if get data ranges returns a general error with latest data time", func() {
						latestDataTime := pointer.FromTime(time.Now().Add(-Day))
						dataSrc.LatestDataTime = latestDataTime
						testErr := errorsTest.RandomError()
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(nil, nil, testErr))
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStatePending)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, testErr.Error())
						assertProviderSessionNotRefreshed()
					})

					It("fails if get data ranges refreshes the token and returns a general error", func() {
						testErr := errorsTest.RandomError()
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(&MockTokenSource{Refresh: true}, nil, testErr))
						authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession))
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStatePending)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, testErr.Error())
						assertProviderSessionRefreshedTimes(1)
					})

					It("fails if get data ranges refreshes the token and returns an authentication error", func() {
						testErr := request.ErrorUnauthenticated()
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(&MockTokenSource{Refresh: true}, nil, testErr))
						authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession))
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStatePending)
						assertTaskRetryCount(1)
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeAuthenticationFailure, testErr.Error())
						assertProviderSessionRefreshedTimes(1)
					})

					// The refreshed token is lost if it cannot be persisted, so the run must stop even though the Dexcom
					// request itself succeeded
					It("fails if get data ranges refreshes the token and every provider session update fails", func() {
						testErr := errorsTest.RandomError()
						dataRangeResponse := &dexcom.DataRangesResponse{
							Calibrations: &dexcom.DataRange{
								Start: &dexcom.Moment{SystemTime: &dexcom.Time{Time: time.Now().Add(-7 * Day)}},
								End:   &dexcom.Moment{SystemTime: &dexcom.Time{Time: time.Now().Add(-3 * Day)}},
							},
						}
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(&MockTokenSource{Refresh: true}, dataRangeResponse, nil))
						authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).Return(nil, testErr).Times(4)
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStatePending)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, "unable to update provider session")
						assertProviderSessionNotRefreshed()
					})

					It("fails only with the Dexcom error if get data ranges refreshes the token and a provider session update retry succeeds", func() {
						testErr := errorsTest.RandomError()
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(&MockTokenSource{Refresh: true}, nil, testErr))
						authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).Return(nil, errorsTest.RandomError())
						authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession))
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStatePending)
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, testErr.Error())
						assertProviderSessionRefreshedTimes(1)
					})

					// The count round trips as int32 through BSON, but as float64 through JSON, and narrowing to one
					// would restart the count and defeat TaskRetryCountMaximum
					DescribeTable("continues an existing retry count stored as",
						func(storedRetryCount any) {
							testErr := request.ErrorUnauthenticated()
							tsk.Data[dexcom.DataKeyRetryCount] = storedRetryCount
							dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(nil, nil, testErr))
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStatePending)
							assertTaskRetryCount(3)
							assertTaskAndDataSourceError(dexcomFetch.ErrorCodeAuthenticationFailure, testErr.Error())
						},
						Entry("int32", int32(2)),
						Entry("int64", int64(2)),
						Entry("int", int(2)),
						Entry("float64", float64(2)),
					)

					It("fails permanently if the retry count is already at the maximum", func() {
						testErr := request.ErrorUnauthenticated()
						tsk.Data[dexcom.DataKeyRetryCount] = float64(dexcomFetch.TaskRetryCountMaximum)
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(nil, nil, testErr))
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStateFailed)
						assertTaskRetryCount(dexcomFetch.TaskRetryCountMaximum + 1)
						assertTaskAndDataSourceError(dexcomFetch.ErrorCodeAuthenticationFailure, testErr.Error())
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
							dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(&MockTokenSource{Refresh: true}, dataRangeResponse, nil))
							authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession))
						})

						It("is successful if the Dexcom data ranges is not valid", func() {
							dataRangeResponse.Calibrations.Start = nil
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStatePending)
							assertTaskRetryCountNotPresent()
							assertTaskResumeTimeNotPresent()
							assertTaskAndDataSourceErrorNotPresent()
							assertProviderSessionRefreshedTimes(1)
						})

						It("is successful if the Dexcom data ranges start is not before end", func() {
							dataRangeResponse.Calibrations.Start = &dexcom.Moment{SystemTime: &dexcom.Time{Time: time.Now().Add(-2 * Day)}}
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStatePending)
							assertTaskRetryCountNotPresent()
							assertTaskResumeTimeNotPresent()
							assertTaskAndDataSourceErrorNotPresent()
							assertProviderSessionRefreshedTimes(1)
						})

						It("fails if get alerts returns no response and no error", func() {
							dexcomClient.EXPECT().GetAlerts(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData[dexcom.AlertsResponse](nil, nil, nil))
							authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession))
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStatePending)
							assertTaskRetryCountNotPresent()
							assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, "alerts response is missing")
							assertProviderSessionRefreshedTimes(2)
						})

						It("fails if get alerts returns no records and no error", func() {
							dexcomClient.EXPECT().GetAlerts(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, &dexcom.AlertsResponse{}, nil))
							authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession))
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStatePending)
							assertTaskRetryCountNotPresent()
							assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, "alerts response is missing")
							assertProviderSessionRefreshedTimes(2)
						})

						It("fails if get alerts returns a general error", func() {
							testErr := errorsTest.RandomError()
							dexcomClient.EXPECT().GetAlerts(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData[dexcom.AlertsResponse](nil, nil, testErr))
							authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession))
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStatePending)
							assertTaskRetryCountNotPresent()
							assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, testErr.Error())
							assertProviderSessionRefreshedTimes(2)
						})

						It("fails if get alerts refreshes the token and returns a general error", func() {
							testErr := errorsTest.RandomError()
							dexcomClient.EXPECT().GetAlerts(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData[dexcom.AlertsResponse](&MockTokenSource{Refresh: true}, nil, testErr))
							authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession))
							taskRunner.Run(ctx)
							assertTaskAndDataSourceState(task.TaskStatePending)
							assertTaskRetryCountNotPresent()
							assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, testErr.Error())
							assertProviderSessionRefreshedTimes(2)
						})

						It("fails if get alerts refreshes the token and returns an authentication error", func() {
							testErr := request.ErrorUnauthenticated()
							dexcomClient.EXPECT().GetAlerts(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData[dexcom.AlertsResponse](&MockTokenSource{Refresh: true}, nil, testErr))
							authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession))
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
								dexcomClient.EXPECT().GetAlerts(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, alertsResponse, nil))
								calibrationsResponse = &dexcom.CalibrationsResponse{Records: &dexcom.Calibrations{}}
								dexcomClient.EXPECT().GetCalibrations(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, calibrationsResponse, nil))
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
								dexcomClient.EXPECT().GetDevices(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, devicesResponse, nil))
								egvsResponse = &dexcom.EGVsResponse{Records: &dexcom.EGVs{}}
								dexcomClient.EXPECT().GetEGVs(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, egvsResponse, nil))
								eventsResponse = &dexcom.EventsResponse{Records: &dexcom.Events{}}
								dexcomClient.EXPECT().GetEvents(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, eventsResponse, nil))
								authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(5)
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
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(2)
								dataClient.EXPECT().CreateUserDataSet(matchContext(), "test-user-id", matchNotNil()).DoAndReturn(mockDataClientCreateUserDataSet(dataSet, nil))
								dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(nil))
								taskRunner.Run(ctx)
								assertTaskAndDataSourceState(task.TaskStatePending)
								assertTaskAvailableAfterStandardDuration()
								assertTaskDeviceHashesCount(3)
								assertTaskRetryCountNotPresent()
								assertTaskResumeTimeNotPresent()
								assertTaskAndDataSourceErrorNotPresent()
								assertDataSourceLastImportTimePresent()
								assertProviderSessionRefreshedTimes(6)
							})

							It("completes the import if the runner duration maximum is exceeded and no segment remains", func() {
								runnerDurationMaximum = -time.Second
								dataSet := &data.DataSet{
									ID:       pointer.FromString("test-data-set-id"),
									UploadID: pointer.FromString("test-data-set-upload-id"),
								}
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(2)
								dataClient.EXPECT().CreateUserDataSet(matchContext(), "test-user-id", matchNotNil()).DoAndReturn(mockDataClientCreateUserDataSet(dataSet, nil))
								dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(nil))
								taskRunner.Run(ctx)
								assertTaskAndDataSourceState(task.TaskStatePending)
								assertTaskAvailableAfterStandardDuration()
								assertTaskDeviceHashesCount(3)
								assertTaskRetryCountNotPresent()
								assertTaskResumeTimeNotPresent()
								assertTaskAndDataSourceErrorNotPresent()
								assertDataSourceLastImportTimePresent()
								assertProviderSessionRefreshedTimes(6)
							})

							It("completes the import if the runner duration maximum is exceeded, no segment remains, and a later update fails", func() {
								runnerDurationMaximum = -time.Second
								testErr := errorsTest.RandomError()
								dataSet := &data.DataSet{
									ID:       pointer.FromString("test-data-set-id"),
									UploadID: pointer.FromString("test-data-set-upload-id"),
								}
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc))
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).Return(nil, testErr)
								dataClient.EXPECT().CreateUserDataSet(matchContext(), "test-user-id", matchNotNil()).DoAndReturn(mockDataClientCreateUserDataSet(dataSet, nil))
								dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(nil))
								taskRunner.Run(ctx)
								assertTaskState(task.TaskStatePending)
								assertTaskAvailableAfterStandardDuration()
								assertTaskRetryCountNotPresent()
								assertTaskError(dexcomFetch.ErrorCodeResourceFailure, "unable to update data source")
								assertDataSourceLastImportTimePresent()
								assertProviderSessionRefreshedTimes(6)
							})

							It("fails if the created data set has no upload id", func() {
								dataSet := &data.DataSet{ID: pointer.FromString("test-data-set-id")}
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc))
								dataClient.EXPECT().CreateUserDataSet(matchContext(), "test-user-id", matchNotNil()).DoAndReturn(mockDataClientCreateUserDataSet(dataSet, nil))
								taskRunner.Run(ctx)
								assertTaskAndDataSourceState(task.TaskStateFailed)
								assertTaskRetryCountNotPresent()
								assertTaskResumeTimeNotPresent()
								assertTaskAndDataSourceError(dexcomFetch.ErrorCodeInvalidState, "data set upload id is missing")
								assertProviderSessionRefreshedTimes(6)
							})
						})

						Context("with Dexcom data ranges response spanning multiple segments", func() {
							var transmitterID string
							var firstDevice *dexcom.Device
							var secondDevice *dexcom.Device

							expectSegment := func(segmentStartTime time.Time, segmentEndTime time.Time, devices ...*dexcom.Device) {
								dexcomClient.EXPECT().GetAlerts(matchContext(), segmentStartTime, segmentEndTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, &dexcom.AlertsResponse{Records: &dexcom.Alerts{}}, nil))
								dexcomClient.EXPECT().GetCalibrations(matchContext(), segmentStartTime, segmentEndTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, &dexcom.CalibrationsResponse{Records: &dexcom.Calibrations{}}, nil))
								dexcomClient.EXPECT().GetDevices(matchContext(), segmentStartTime, segmentEndTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, &dexcom.DevicesResponse{Records: pointer.From(dexcom.Devices(devices))}, nil))
								dexcomClient.EXPECT().GetEGVs(matchContext(), segmentStartTime, segmentEndTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, &dexcom.EGVsResponse{Records: &dexcom.EGVs{}}, nil))
								dexcomClient.EXPECT().GetEvents(matchContext(), segmentStartTime, segmentEndTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, &dexcom.EventsResponse{Records: &dexcom.Events{}}, nil))
							}

							assertTaskDeviceHash := func(deviceID string, deviceHash string) {
								deviceHashes, ok := tsk.Data[dexcom.DataKeyDeviceHashes].(map[string]string)
								Expect(ok).To(BeTrue())
								Expect(deviceHashes).To(HaveKeyWithValue(deviceID, deviceHash))
							}

							BeforeEach(func() {
								startTime = time.Now().Add(-45 * Day)
								dataRangeResponse.Calibrations.Start = &dexcom.Moment{SystemTime: &dexcom.Time{Time: startTime}}

								// The same device reported in both segments, with a change that alters its hash
								transmitterID = dexcomTest.RandomTransmitterID()
								firstDevice = &dexcom.Device{
									LastUploadDate:        &dexcom.Time{Time: startTime.Add(Day)},
									AlertSchedules:        &dexcom.AlertSchedules{},
									TransmitterID:         pointer.FromString(transmitterID),
									TransmitterGeneration: pointer.FromString(dexcom.DeviceTransmitterGenerationG6),
									DisplayDevice:         pointer.FromString(dexcom.DeviceDisplayDeviceIOS),
									DisplayApp:            pointer.FromString(dexcom.DeviceDisplayAppG6),
								}
								secondDevice = &dexcom.Device{
									LastUploadDate:        &dexcom.Time{Time: endTime.Add(-Day)},
									AlertSchedules:        &dexcom.AlertSchedules{},
									TransmitterID:         pointer.FromString(transmitterID),
									TransmitterGeneration: pointer.FromString(dexcom.DeviceTransmitterGenerationG6),
									DisplayDevice:         pointer.FromString(dexcom.DeviceDisplayDeviceAndroid),
									DisplayApp:            pointer.FromString(dexcom.DeviceDisplayAppG7),
								}
							})

							It("completes the import once every segment is fetched", func() {
								secondDeviceHash := test.Must(secondDevice.Hash())

								segmentStartTime := startTime.AddDate(0, 0, dexcomFetch.DataRangeDaysMaximum)
								expectSegment(startTime, segmentStartTime, firstDevice)
								expectSegment(segmentStartTime, endTime, secondDevice)
								authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(10)

								dataSet := &data.DataSet{
									ID:       pointer.FromString("test-data-set-id"),
									UploadID: pointer.FromString("test-data-set-upload-id"),
								}
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(3)
								dataClient.EXPECT().CreateUserDataSet(matchContext(), "test-user-id", matchNotNil()).DoAndReturn(mockDataClientCreateUserDataSet(dataSet, nil))
								dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(nil)).Times(2)

								taskRunner.Run(ctx)
								assertTaskAndDataSourceState(task.TaskStatePending)
								assertTaskAvailableAfterStandardDuration()
								assertTaskDeviceHash(transmitterID, secondDeviceHash)
								assertTaskRetryCountNotPresent()
								assertTaskResumeTimeNotPresent()
								assertTaskAndDataSourceErrorNotPresent()
								assertDataSourceLastImportTimePresent()
								assertProviderSessionRefreshedTimes(11)
							})

							It("is available soon if the runner duration maximum is exceeded and a segment remains", func() {
								runnerDurationMaximum = -time.Second
								firstDeviceHash := test.Must(firstDevice.Hash())

								// Only the first segment is expected, so fetching the second fails as an unexpected call
								segmentStartTime := startTime.AddDate(0, 0, dexcomFetch.DataRangeDaysMaximum)
								expectSegment(startTime, segmentStartTime, firstDevice)
								authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(5)

								dataSet := &data.DataSet{
									ID:       pointer.FromString("test-data-set-id"),
									UploadID: pointer.FromString("test-data-set-upload-id"),
								}
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(2)
								dataClient.EXPECT().CreateUserDataSet(matchContext(), "test-user-id", matchNotNil()).DoAndReturn(mockDataClientCreateUserDataSet(dataSet, nil))
								dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(nil))

								taskRunner.Run(ctx)
								assertTaskAndDataSourceState(task.TaskStatePending)
								assertTaskAvailableSoon()
								assertTaskDeviceHash(transmitterID, firstDeviceHash)
								assertTaskRetryCountNotPresent()
								assertTaskResumeTime(segmentStartTime)
								assertTaskAndDataSourceErrorNotPresent()
								assertProviderSessionRefreshedTimes(6)
							})

							// The data source data times only advance when data is stored, so without the resume data time
							// the next run would walk this same empty segment again, and never progress
							It("records the resume data time for a segment that yields no data", func() {
								runnerDurationMaximum = -time.Second

								// Only the first segment is expected, so fetching the second fails as an unexpected call
								segmentStartTime := startTime.AddDate(0, 0, dexcomFetch.DataRangeDaysMaximum)
								expectSegment(startTime, segmentStartTime)
								authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(5)

								taskRunner.Run(ctx)
								assertTaskAndDataSourceState(task.TaskStatePending)
								assertTaskAvailableSoon()
								assertTaskRetryCountNotPresent()
								assertTaskResumeTime(segmentStartTime)
								assertTaskAndDataSourceErrorNotPresent()
								Expect(dataSrc.LatestDataTime).To(BeNil())
								assertProviderSessionRefreshedTimes(6)
							})

							It("resumes from the resume data time recorded by an earlier run, then discards it", func() {
								secondDeviceHash := test.Must(secondDevice.Hash())

								// The remaining range is short enough for a single segment, and in UTC because that is
								// how the resume data time round trips
								resumeTime := startTime.AddDate(0, 0, dexcomFetch.DataRangeDaysMaximum).UTC()
								setTaskResumeTime(resumeTime, time.Now().Add(time.Minute))
								expectSegment(resumeTime, endTime, secondDevice)
								authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(5)

								dataSet := &data.DataSet{
									ID:       pointer.FromString("test-data-set-id"),
									UploadID: pointer.FromString("test-data-set-upload-id"),
								}
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(2)
								dataClient.EXPECT().CreateUserDataSet(matchContext(), "test-user-id", matchNotNil()).DoAndReturn(mockDataClientCreateUserDataSet(dataSet, nil))
								dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(nil))

								taskRunner.Run(ctx)
								logger.AssertDebug("Resuming fetch from resume data time", log.Fields{dexcom.DataKeyResumeDataTime: resumeTime})
								assertTaskAndDataSourceState(task.TaskStatePending)
								assertTaskAvailableAfterStandardDuration()
								assertTaskDeviceHash(transmitterID, secondDeviceHash)
								assertTaskRetryCountNotPresent()
								assertTaskResumeTimeNotPresent()
								assertTaskAndDataSourceErrorNotPresent()
								assertDataSourceLastImportTimePresent()
								assertProviderSessionRefreshedTimes(6)
							})

							// Reaching the end of the data range is the only thing that discards the resume data time
							// when the remaining segments store nothing, and a caught up task must carry none
							It("discards the resume data time on reaching the end of the data range with no data", func() {
								resumeTime := startTime.AddDate(0, 0, dexcomFetch.DataRangeDaysMaximum).UTC()
								setTaskResumeTime(resumeTime, time.Now().Add(time.Minute))
								expectSegment(resumeTime, endTime)
								authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(5)

								taskRunner.Run(ctx)
								assertTaskAndDataSourceState(task.TaskStatePending)
								assertTaskAvailableAfterStandardDuration()
								assertTaskRetryCountNotPresent()
								assertTaskResumeTimeNotPresent()
								assertTaskAndDataSourceErrorNotPresent()
								assertDataSourceLastImportTimePresent()
								Expect(dataSrc.LatestDataTime).To(BeNil())
								assertProviderSessionRefreshedTimes(6)
							})

							// Nothing deduplicates, so a stale resume data time must never pull the start back over data
							// already stored
							It("resumes from the latest data time when the resume data time is older", func() {
								secondDeviceHash := test.Must(secondDevice.Hash())

								// The remaining range is short enough for a single segment
								segmentStartTime := startTime.AddDate(0, 0, dexcomFetch.DataRangeDaysMaximum)
								dataSrc.LatestDataTime = pointer.FromTime(segmentStartTime)
								setTaskResumeTime(startTime, time.Now().Add(time.Minute))
								expectSegment(segmentStartTime, endTime, secondDevice)
								authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(5)

								dataSet := &data.DataSet{
									ID:       pointer.FromString("test-data-set-id"),
									UploadID: pointer.FromString("test-data-set-upload-id"),
								}
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(2)
								dataClient.EXPECT().CreateUserDataSet(matchContext(), "test-user-id", matchNotNil()).DoAndReturn(mockDataClientCreateUserDataSet(dataSet, nil))
								dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(nil))

								taskRunner.Run(ctx)
								assertTaskAndDataSourceState(task.TaskStatePending)
								assertTaskAvailableAfterStandardDuration()
								assertTaskDeviceHash(transmitterID, secondDeviceHash)
								assertTaskRetryCountNotPresent()
								assertTaskResumeTimeNotPresent()
								assertTaskAndDataSourceErrorNotPresent()
								assertDataSourceLastImportTimePresent()
								assertProviderSessionRefreshedTimes(6)
							})

							// A task is only rescheduled promptly, never guaranteed to run promptly, and Dexcom accepts
							// an upload for a segment already walked in the meantime, so a resume data time that cannot be
							// shown to be fresh is discarded rather than trusted
							Context("with an unusable resume data time", func() {
								var segmentStartTime time.Time
								var secondDeviceHash string

								BeforeEach(func() {
									segmentStartTime = startTime.AddDate(0, 0, dexcomFetch.DataRangeDaysMaximum)
									secondDeviceHash = test.Must(secondDevice.Hash())

									expectSegment(startTime, segmentStartTime, firstDevice)
									expectSegment(segmentStartTime, endTime, secondDevice)
									authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(10)

									dataSet := &data.DataSet{
										ID:       pointer.FromString("test-data-set-id"),
										UploadID: pointer.FromString("test-data-set-upload-id"),
									}
									dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(3)
									dataClient.EXPECT().CreateUserDataSet(matchContext(), "test-user-id", matchNotNil()).DoAndReturn(mockDataClientCreateUserDataSet(dataSet, nil))
									dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(nil)).Times(2)
								})

								assertWholeDataRangeWalked := func() {
									assertTaskAndDataSourceState(task.TaskStatePending)
									assertTaskAvailableAfterStandardDuration()
									assertTaskDeviceHash(transmitterID, secondDeviceHash)
									assertTaskRetryCountNotPresent()
									assertTaskResumeTimeNotPresent()
									assertTaskAndDataSourceErrorNotPresent()
									assertDataSourceLastImportTimePresent()
									assertProviderSessionRefreshedTimes(11)
								}

								It("walks the whole data range when it has expired", func() {
									setTaskResumeTime(segmentStartTime, time.Now().Add(-time.Second))
									taskRunner.Run(ctx)
									logger.AssertWarn("Ignoring expired resume data time", log.Fields{dexcom.DataKeyResumeDataTime: segmentStartTime.UTC()})
									assertWholeDataRangeWalked()
								})

								It("walks the whole data range when it has no expiration time", func() {
									setTaskResumeTime(segmentStartTime, time.Now().Add(time.Minute))
									delete(tsk.Data, dexcom.DataKeyResumeExpirationTime)
									taskRunner.Run(ctx)
									logger.AssertWarn("Ignoring expired resume data time", log.Fields{dexcom.DataKeyResumeDataTime: segmentStartTime.UTC()})
									assertWholeDataRangeWalked()
								})

								It("walks the whole data range when it cannot be parsed", func() {
									setTaskResumeTime(segmentStartTime, time.Now().Add(time.Minute))
									tsk.Data[dexcom.DataKeyResumeDataTime] = "test-invalid-resume-time"
									taskRunner.Run(ctx)
									logger.AssertWarn("Unable to parse data time", log.Fields{dexcom.DataKeyResumeDataTime: "test-invalid-resume-time"})
									assertWholeDataRangeWalked()
								})

								It("walks the whole data range when its expiration time cannot be parsed", func() {
									setTaskResumeTime(segmentStartTime, time.Now().Add(time.Minute))
									tsk.Data[dexcom.DataKeyResumeExpirationTime] = "test-invalid-resume-expiration-time"
									taskRunner.Run(ctx)
									logger.AssertWarn("Unable to parse data time", log.Fields{dexcom.DataKeyResumeExpirationTime: "test-invalid-resume-expiration-time"})
									assertWholeDataRangeWalked()
								})
							})

							It("retains the stored device hash when a later segment fails before storing its device", func() {
								testErr := errorsTest.RandomError()
								firstDeviceHash := test.Must(firstDevice.Hash())
								secondDeviceHash := test.Must(secondDevice.Hash())
								Expect(secondDeviceHash).ToNot(Equal(firstDeviceHash))

								segmentStartTime := startTime.AddDate(0, 0, dexcomFetch.DataRangeDaysMaximum)
								expectSegment(startTime, segmentStartTime, firstDevice)
								expectSegment(segmentStartTime, endTime, secondDevice)
								authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(10)

								dataSet := &data.DataSet{
									ID:       pointer.FromString("test-data-set-id"),
									UploadID: pointer.FromString("test-data-set-upload-id"),
								}
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(2)
								dataClient.EXPECT().CreateUserDataSet(matchContext(), "test-user-id", matchNotNil()).DoAndReturn(mockDataClientCreateUserDataSet(dataSet, nil))
								dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(nil))
								dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(testErr))

								taskRunner.Run(ctx)
								assertTaskAndDataSourceState(task.TaskStatePending)
								assertTaskDeviceHash(transmitterID, firstDeviceHash)
								assertTaskRetryCountNotPresent()
								assertTaskResumeTimeNotPresent()
								assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, testErr.Error())
								assertProviderSessionRefreshedTimes(11)
							})

							// A failed persist leaves the latest data time behind the data actually stored, so the resume
							// time is the only thing keeping the next run from walking that segment again and storing
							// every record in it a second time
							It("keeps the resume data time when persisting the latest data time fails", func() {
								testErr := errorsTest.RandomError()

								// Equal to the start of the data range, so the segment is unchanged by it
								setTaskResumeTime(startTime, time.Now().Add(dexcomFetch.ResumeExpirationDuration))

								// Only the first segment is expected, since storing its data fails the run
								segmentStartTime := startTime.AddDate(0, 0, dexcomFetch.DataRangeDaysMaximum)
								expectSegment(startTime, segmentStartTime, firstDevice)
								authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(5)

								dataSet := &data.DataSet{
									ID:       pointer.FromString("test-data-set-id"),
									UploadID: pointer.FromString("test-data-set-upload-id"),
								}
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).Return(nil, testErr)
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc))
								dataClient.EXPECT().CreateUserDataSet(matchContext(), "test-user-id", matchNotNil()).DoAndReturn(mockDataClientCreateUserDataSet(dataSet, nil))
								dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(nil))

								taskRunner.Run(ctx)
								assertTaskState(task.TaskStatePending)
								assertTaskRetryCountNotPresent()
								assertTaskResumeTime(startTime)
								assertTaskError(dexcomFetch.ErrorCodeResourceFailure, "unable to update data source")
								Expect(dataSrc.LatestDataTime).To(BeNil())
								assertProviderSessionRefreshedTimes(6)
							})

							// Stored data supersedes the resume data time, and a failure leaves neither a completed import
							// nor a new resume data time to discard it
							It("discards a superseded resume data time when a later segment fails", func() {
								testErr := errorsTest.RandomError()

								// Equal to the start of the data range, so the segments are unchanged by it
								setTaskResumeTime(startTime, time.Now().Add(time.Minute))

								segmentStartTime := startTime.AddDate(0, 0, dexcomFetch.DataRangeDaysMaximum)
								expectSegment(startTime, segmentStartTime, firstDevice)
								expectSegment(segmentStartTime, endTime, secondDevice)
								authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).Times(10)

								dataSet := &data.DataSet{
									ID:       pointer.FromString("test-data-set-id"),
									UploadID: pointer.FromString("test-data-set-upload-id"),
								}
								dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc)).Times(2)
								dataClient.EXPECT().CreateUserDataSet(matchContext(), "test-user-id", matchNotNil()).DoAndReturn(mockDataClientCreateUserDataSet(dataSet, nil))
								dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(nil))
								dataClient.EXPECT().CreateDataSetsData(matchContext(), "test-data-set-upload-id", matchNotNil()).DoAndReturn(mockDataClientCreateDataSetsData(testErr))

								taskRunner.Run(ctx)
								assertTaskAndDataSourceState(task.TaskStatePending)
								assertTaskRetryCountNotPresent()
								assertTaskResumeTimeNotPresent()
								Expect(dataSrc.LatestDataTime).ToNot(BeNil())
								assertTaskAndDataSourceError(dexcomFetch.ErrorCodeResourceFailure, testErr.Error())
								assertProviderSessionRefreshedTimes(11)
							})
						})
					})

					// ALTERNATES:
					// deviceHashes - not in data
					// dataSource.LatestDataTime - not nil (recent)
					// refresh token
				})

				Context("with provider session and a data range spanning multiple chunks", func() {
					var providerSession *auth.ProviderSession
					var firstChunkStartTime time.Time
					var firstChunkEndTime time.Time
					var secondChunkEndTime time.Time

					BeforeEach(func() {
						providerSession = &auth.ProviderSession{
							ID:     "test-provider-session-id",
							UserID: "test-user-id",
							OAuthToken: &auth.OAuthToken{
								AccessToken:    "test-access-token-1",
								TokenType:      "Bearer",
								RefreshToken:   "test-refresh-token-1",
								ExpirationTime: time.Now().Add(time.Minute),
							},
						}
						authClient.EXPECT().GetProviderSession(matchContext(), "test-provider-session-id").Return(providerSession, nil)
						authClient.EXPECT().UpdateProviderSession(matchContext(), "test-provider-session-id", matchNotNil()).DoAndReturn(mockAuthClientUpdateProviderSession(providerSession)).AnyTimes()
						firstChunkStartTime = time.Now().Add(-45 * Day)
						firstChunkEndTime = firstChunkStartTime.AddDate(0, 0, dexcomFetch.DataRangeDaysMaximum)
						secondChunkEndTime = time.Now().Add(-3 * Day)
						dataRangeResponse := &dexcom.DataRangesResponse{
							Calibrations: &dexcom.DataRange{
								Start: &dexcom.Moment{SystemTime: &dexcom.Time{Time: firstChunkStartTime}},
								End:   &dexcom.Moment{SystemTime: &dexcom.Time{Time: secondChunkEndTime}},
							},
						}
						dexcomClient.EXPECT().GetDataRange(matchContext(), nil, matchNotNil()).DoAndReturn(mockDexcomClientGetDataRange(nil, dataRangeResponse, nil))
					})

					// Expects the fetch of a single chunk, all responses empty, invoking onEvents, if any, during the
					// final fetch of the chunk
					expectFetchChunk := func(startTime time.Time, endTime time.Time, onEvents func()) {
						dexcomClient.EXPECT().GetAlerts(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, &dexcom.AlertsResponse{Records: &dexcom.Alerts{}}, nil))
						dexcomClient.EXPECT().GetCalibrations(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, &dexcom.CalibrationsResponse{Records: &dexcom.Calibrations{}}, nil))
						dexcomClient.EXPECT().GetDevices(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, &dexcom.DevicesResponse{Records: &dexcom.Devices{}}, nil))
						dexcomClient.EXPECT().GetEGVs(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData(nil, &dexcom.EGVsResponse{Records: &dexcom.EGVs{}}, nil))
						dexcomClient.EXPECT().GetEvents(matchContext(), startTime, endTime, matchNotNil()).DoAndReturn(func(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EventsResponse, error) {
							if onEvents != nil {
								onEvents()
							}
							return &dexcom.EventsResponse{Records: &dexcom.Events{}}, nil
						})
					}

					It("fetches every chunk of the data range", func() {
						expectFetchChunk(firstChunkStartTime, firstChunkEndTime, nil)
						expectFetchChunk(firstChunkEndTime, secondChunkEndTime, nil)
						dataSourceClient.EXPECT().Update(matchContext(), "test-data-source-id", matchNil(), matchNotNil()).DoAndReturn(mockDataSourceClientUpdate(dataSrc))
						taskRunner.Run(ctx)
						assertTaskAndDataSourceState(task.TaskStatePending)
						assertTaskAvailableAfterStandardDuration()
						assertTaskRetryCountNotPresent()
						assertTaskAndDataSourceErrorNotPresent()
						assertDataSourceLastImportTimePresent()
					})

					It("discards the run outcome if the task claim is lost mid-fetch", func() {
						claimContext, claimCancel := context.WithCancelCause(ctx)
						defer claimCancel(nil)
						expectFetchChunk(firstChunkStartTime, firstChunkEndTime, func() { claimCancel(task.ErrClaimLost) })
						// The canceled context fails the next chunk, ending the run
						dexcomClient.EXPECT().GetAlerts(matchContext(), firstChunkEndTime, secondChunkEndTime, matchNotNil()).DoAndReturn(mockDexcomClientGetData[dexcom.AlertsResponse](nil, nil, context.Canceled))
						taskRunner.Run(claimContext)
						assertTaskState(task.TaskStateRunning)
						Expect(dataSrc.State).To(Equal(dataSource.StateConnected))
						Expect(dataSrc.HasError()).To(BeFalse())
						logger.AssertWarn("Skipped updating data source and task because the task claim was lost")
					})
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
			localDataSrc.State = *update.State
		}
		if update.Error != nil {
			localDataSrc.Error = update.Error
		}
		if update.DataSetID != nil {
			localDataSrc.DataSetID = update.DataSetID
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
		tokenSource.UpdateToken(ctx)
		return response, err
	}
}

func mockDexcomClientGetData[T any](mockTokenSource *MockTokenSource, response *T, err error) func(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*T, error) {
	if mockTokenSource == nil {
		mockTokenSource = &MockTokenSource{}
	}
	return func(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*T, error) {
		tokenSource.HTTPClient(ctx, mockTokenSource)
		tokenSource.UpdateToken(ctx)
		return response, err
	}
}

type MockTokenSource struct {
	Refresh bool
	token   *auth.OAuthToken
}

func (m *MockTokenSource) TokenSource(ctx context.Context, token *auth.OAuthToken) (oauth2.TokenSource, error) {
	m.token = token
	return m, nil
}

func (m *MockTokenSource) Token() (*oauth2.Token, error) {
	if m.Refresh {
		m.token = &auth.OAuthToken{
			AccessToken:    fmt.Sprintf("%s*", m.token.AccessToken),
			TokenType:      m.token.TokenType,
			RefreshToken:   fmt.Sprintf("%s*", m.token.RefreshToken),
			ExpirationTime: time.Now().Add(time.Minute),
		}
	}
	return m.token.RawToken(), nil
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
