package v1_test

import (
	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/tidepool-org/platform/auth"
	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testDataDeduplicator "github.com/tidepool-org/platform/data/deduplicator/test"
	dataStore "github.com/tidepool-org/platform/data/store"
	testDataStore "github.com/tidepool-org/platform/data/store/test"
	"github.com/tidepool-org/platform/log"
	metricClient "github.com/tidepool-org/platform/metric/client"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/store"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
	userClient "github.com/tidepool-org/platform/user/client"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/service/api/v1")
}

type RecordMetricInput struct {
	context auth.Context
	metric  string
	data    []map[string]string
}

type TestMetricClient struct {
	RecordMetricInputs  []RecordMetricInput
	RecordMetricOutputs []error
}

func (t *TestMetricClient) RecordMetric(context auth.Context, metric string, data ...map[string]string) error {
	t.RecordMetricInputs = append(t.RecordMetricInputs, RecordMetricInput{context, metric, data})
	output := t.RecordMetricOutputs[0]
	t.RecordMetricOutputs = t.RecordMetricOutputs[1:]
	return output
}

func (t *TestMetricClient) ValidateTest() bool {
	return len(t.RecordMetricOutputs) == 0
}

type GetUserPermissionsInput struct {
	context       auth.Context
	requestUserID string
	targetUserID  string
}

type GetUserPermissionsOutput struct {
	permissions userClient.Permissions
	err         error
}

type TestUserClient struct {
	GetUserPermissionsInputs  []GetUserPermissionsInput
	GetUserPermissionsOutputs []GetUserPermissionsOutput
}

func (t *TestUserClient) Start() error {
	panic("Unexpected invocation of Start on TestUserClient")
}

func (t *TestUserClient) Close() {
	panic("Unexpected invocation of Close on TestUserClient")
}

func (t *TestUserClient) GetUserPermissions(context auth.Context, requestUserID string, targetUserID string) (userClient.Permissions, error) {
	t.GetUserPermissionsInputs = append(t.GetUserPermissionsInputs, GetUserPermissionsInput{context, requestUserID, targetUserID})
	output := t.GetUserPermissionsOutputs[0]
	t.GetUserPermissionsOutputs = t.GetUserPermissionsOutputs[1:]
	return output.permissions, output.err
}

func (t *TestUserClient) ValidateTest() bool {
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

func (t *TestSyncTaskStoreSession) SetAgent(agent store.Agent) {
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

type TestContext struct {
	*testAuth.Context
	RespondWithErrorInputs                 []*service.Error
	RespondWithInternalServerFailureInputs []RespondWithInternalServerFailureInput
	RespondWithStatusAndErrorsInputs       []RespondWithStatusAndErrorsInput
	RespondWithStatusAndDataInputs         []RespondWithStatusAndDataInput
	MetricClientImpl                       *TestMetricClient
	UserClientImpl                         *TestUserClient
	DataDeduplicatorFactoryImpl            *testDataDeduplicator.Factory
	DataStoreSessionImpl                   *testDataStore.Session
	SyncTaskStoreSessionImpl               *TestSyncTaskStoreSession
}

func NewTestContext() *TestContext {
	return &TestContext{
		Context:                     testAuth.NewContext(),
		MetricClientImpl:            &TestMetricClient{},
		UserClientImpl:              &TestUserClient{},
		DataDeduplicatorFactoryImpl: testDataDeduplicator.NewFactory(),
		DataStoreSessionImpl:        testDataStore.NewSession(),
		SyncTaskStoreSessionImpl:    &TestSyncTaskStoreSession{},
	}
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

func (t *TestContext) MetricClient() metricClient.Client {
	return t.MetricClientImpl
}

func (t *TestContext) UserClient() userClient.Client {
	return t.UserClientImpl
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

func (t *TestContext) ValidateTest() bool {
	return (t.Context == nil || t.Context.UnusedOutputsCount() == 0) &&
		(t.MetricClientImpl == nil || t.MetricClientImpl.ValidateTest()) &&
		(t.UserClientImpl == nil || t.UserClientImpl.ValidateTest()) &&
		(t.DataDeduplicatorFactoryImpl == nil || t.DataDeduplicatorFactoryImpl.UnusedOutputsCount() == 0) &&
		(t.DataStoreSessionImpl == nil || t.DataStoreSessionImpl.UnusedOutputsCount() == 0) &&
		(t.SyncTaskStoreSessionImpl == nil || t.SyncTaskStoreSessionImpl.ValidateTest())
}
