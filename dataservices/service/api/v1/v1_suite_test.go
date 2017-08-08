package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"net/url"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testDataDeduplicator "github.com/tidepool-org/platform/data/deduplicator/test"
	dataStore "github.com/tidepool-org/platform/data/store"
	testDataStore "github.com/tidepool-org/platform/data/store/test"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	"github.com/tidepool-org/platform/service"
	commonStore "github.com/tidepool-org/platform/store"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dataservices/service/api/v1")
}

type RecordMetricInput struct {
	context metricservicesClient.Context
	metric  string
	data    []map[string]string
}

type TestMetricServicesClient struct {
	RecordMetricInputs  []RecordMetricInput
	RecordMetricOutputs []error
}

func (t *TestMetricServicesClient) RecordMetric(context metricservicesClient.Context, metric string, data ...map[string]string) error {
	t.RecordMetricInputs = append(t.RecordMetricInputs, RecordMetricInput{context, metric, data})
	output := t.RecordMetricOutputs[0]
	t.RecordMetricOutputs = t.RecordMetricOutputs[1:]
	return output
}

func (t *TestMetricServicesClient) ValidateTest() bool {
	return len(t.RecordMetricOutputs) == 0
}

type GetUserPermissionsInput struct {
	context       service.Context
	requestUserID string
	targetUserID  string
}

type GetUserPermissionsOutput struct {
	permissions userservicesClient.Permissions
	err         error
}

type TestUserServicesClient struct {
	GetUserPermissionsInputs  []GetUserPermissionsInput
	GetUserPermissionsOutputs []GetUserPermissionsOutput
}

func (t *TestUserServicesClient) Start() error {
	panic("Unexpected invocation of Start on TestUserServicesClient")
}

func (t *TestUserServicesClient) Close() {
	panic("Unexpected invocation of Close on TestUserServicesClient")
}

func (t *TestUserServicesClient) ValidateAuthenticationToken(context service.Context, authenticationToken string) (userservicesClient.AuthenticationDetails, error) {
	panic("Unexpected invocation of ValidateAuthenticationToken on TestUserServicesClient")
}

func (t *TestUserServicesClient) GetUserPermissions(context service.Context, requestUserID string, targetUserID string) (userservicesClient.Permissions, error) {
	t.GetUserPermissionsInputs = append(t.GetUserPermissionsInputs, GetUserPermissionsInput{context, requestUserID, targetUserID})
	output := t.GetUserPermissionsOutputs[0]
	t.GetUserPermissionsOutputs = t.GetUserPermissionsOutputs[1:]
	return output.permissions, output.err
}

func (t *TestUserServicesClient) ServerToken() (string, error) {
	panic("Unexpected invocation of ServerToken on TestUserServicesClient")
}

func (t *TestUserServicesClient) ValidateTest() bool {
	return len(t.GetUserPermissionsOutputs) == 0
}

type RespondWithInternalServerFailureInput struct {
	message string
	failure []interface{}
}

type RespondWithStatusAndErrorsInput struct {
	statusCode int
	errors     []*service.Error
}

type RespondWithStatusAndDataInput struct {
	statusCode int
	data       interface{}
}

type TestSyncTaskStoreSession struct {
	DestroySyncTasksForUserByIDInputs  []string
	DestroySyncTasksForUserByIDOutputs []error
}

func (t *TestSyncTaskStoreSession) IsClosed() bool {
	panic("Unexpected invocation of IsClosed on TestSyncTaskStoreSession")
}

func (t *TestSyncTaskStoreSession) Close() {
	panic("Unexpected invocation of Close on TestSyncTaskStoreSession")
}

func (t *TestSyncTaskStoreSession) Logger() log.Logger {
	panic("Unexpected invocation of Logger on TestSyncTaskStoreSession")
}

func (t *TestSyncTaskStoreSession) SetAgent(agent commonStore.Agent) {
	panic("Unexpected invocation of SetAgent on TestSyncTaskStoreSession")
}

func (t *TestSyncTaskStoreSession) DestroySyncTasksForUserByID(userID string) error {
	t.DestroySyncTasksForUserByIDInputs = append(t.DestroySyncTasksForUserByIDInputs, userID)
	output := t.DestroySyncTasksForUserByIDOutputs[0]
	t.DestroySyncTasksForUserByIDOutputs = t.DestroySyncTasksForUserByIDOutputs[1:]
	return output
}

func (t *TestSyncTaskStoreSession) ValidateTest() bool {
	return len(t.DestroySyncTasksForUserByIDOutputs) == 0
}

type TestAuthenticationDetails struct {
	IsServerOutputs []bool
	UserIDOutputs   []string
}

func (t *TestAuthenticationDetails) Token() string {
	panic("Unexpected invocation of Token on TestAuthenticationDetails")
}

func (t *TestAuthenticationDetails) IsServer() bool {
	output := t.IsServerOutputs[0]
	t.IsServerOutputs = t.IsServerOutputs[1:]
	return output
}

func (t *TestAuthenticationDetails) UserID() string {
	output := t.UserIDOutputs[0]
	t.UserIDOutputs = t.UserIDOutputs[1:]
	return output
}

func (t *TestAuthenticationDetails) ValidateTest() bool {
	return len(t.IsServerOutputs) == 0 &&
		len(t.UserIDOutputs) == 0
}

type TestContext struct {
	LoggerImpl                             log.Logger
	RequestImpl                            *rest.Request
	RespondWithErrorInputs                 []*service.Error
	RespondWithInternalServerFailureInputs []RespondWithInternalServerFailureInput
	RespondWithStatusAndErrorsInputs       []RespondWithStatusAndErrorsInput
	RespondWithStatusAndDataInputs         []RespondWithStatusAndDataInput
	MetricServicesClientImpl               *TestMetricServicesClient
	UserServicesClientImpl                 *TestUserServicesClient
	DataDeduplicatorFactoryImpl            *testDataDeduplicator.Factory
	DataStoreSessionImpl                   *testDataStore.Session
	SyncTaskStoreSessionImpl               *TestSyncTaskStoreSession
	AuthenticationDetailsImpl              *TestAuthenticationDetails
}

func NewTestContext() *TestContext {
	return &TestContext{
		LoggerImpl: null.NewLogger(),
		RequestImpl: &rest.Request{
			Request: &http.Request{
				URL: &url.URL{},
			},
			PathParams: map[string]string{},
		},
		MetricServicesClientImpl:    &TestMetricServicesClient{},
		UserServicesClientImpl:      &TestUserServicesClient{},
		DataDeduplicatorFactoryImpl: testDataDeduplicator.NewFactory(),
		DataStoreSessionImpl:        testDataStore.NewSession(),
		SyncTaskStoreSessionImpl:    &TestSyncTaskStoreSession{},
		AuthenticationDetailsImpl:   &TestAuthenticationDetails{},
	}
}

func (t *TestContext) Logger() log.Logger {
	return t.LoggerImpl
}

func (t *TestContext) Request() *rest.Request {
	return t.RequestImpl
}

func (t *TestContext) Response() rest.ResponseWriter {
	panic("Unexpected invocation of Response on TestContext")
}

func (t *TestContext) RespondWithError(err *service.Error) {
	t.RespondWithErrorInputs = append(t.RespondWithErrorInputs, err)
}

func (t *TestContext) RespondWithInternalServerFailure(message string, failure ...interface{}) {
	t.RespondWithInternalServerFailureInputs = append(t.RespondWithInternalServerFailureInputs, RespondWithInternalServerFailureInput{message, failure})
}

func (t *TestContext) RespondWithStatusAndErrors(statusCode int, errors []*service.Error) {
	t.RespondWithStatusAndErrorsInputs = append(t.RespondWithStatusAndErrorsInputs, RespondWithStatusAndErrorsInput{statusCode, errors})
}

func (t *TestContext) RespondWithStatusAndData(statusCode int, data interface{}) {
	t.RespondWithStatusAndDataInputs = append(t.RespondWithStatusAndDataInputs, RespondWithStatusAndDataInput{statusCode, data})
}

func (t *TestContext) MetricServicesClient() metricservicesClient.Client {
	return t.MetricServicesClientImpl
}

func (t *TestContext) UserServicesClient() userservicesClient.Client {
	return t.UserServicesClientImpl
}

func (t *TestContext) DataFactory() data.Factory {
	panic("Unexpected invocation of DataFactory on TestContext")
}

func (t *TestContext) DataDeduplicatorFactory() deduplicator.Factory {
	return t.DataDeduplicatorFactoryImpl
}

func (t *TestContext) DataStoreSession() dataStore.Session {
	return t.DataStoreSessionImpl
}

func (t *TestContext) SyncTaskStoreSession() syncTaskStore.Session {
	return t.SyncTaskStoreSessionImpl
}

func (t *TestContext) AuthenticationDetails() userservicesClient.AuthenticationDetails {
	return t.AuthenticationDetailsImpl
}

func (t *TestContext) SetAuthenticationDetails(authenticationDetails userservicesClient.AuthenticationDetails) {
	panic("Unexpected invocation of SetAuthenticationDetails on TestContext")
}

func (t *TestContext) ValidateTest() bool {
	return (t.MetricServicesClientImpl == nil || t.MetricServicesClientImpl.ValidateTest()) &&
		(t.UserServicesClientImpl == nil || t.UserServicesClientImpl.ValidateTest()) &&
		(t.DataDeduplicatorFactoryImpl == nil || t.DataDeduplicatorFactoryImpl.UnusedOutputsCount() == 0) &&
		(t.DataStoreSessionImpl == nil || t.DataStoreSessionImpl.UnusedOutputsCount() == 0) &&
		(t.SyncTaskStoreSessionImpl == nil || t.SyncTaskStoreSessionImpl.ValidateTest()) &&
		(t.AuthenticationDetailsImpl == nil || t.AuthenticationDetailsImpl.ValidateTest())
}
