package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
	metricservicesClient "github.com/tidepool-org/platform/metricservices/client"
	"github.com/tidepool-org/platform/service"
	userservicesClient "github.com/tidepool-org/platform/userservices/client"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dataservices/service/api/v1")
}

type TestMetricServicesClient struct {
}

func (t *TestMetricServicesClient) RecordMetric(context metricservicesClient.Context, metric string, data map[string]string) error {
	panic("Unexpected invocation of RecordMetric on TestMetricServicesClient")
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

func (t *TestUserServicesClient) GetUserGroupID(context service.Context, userID string) (string, error) {
	panic("Unexpected invocation of GetUserGroupID on TestUserServicesClient")
}

func (t *TestUserServicesClient) ValidateTest() bool {
	return len(t.GetUserPermissionsOutputs) == 0
}

type RespondWithInternalServerFailureInput struct {
	message string
	failure []interface{}
}

type RespondWithStatusAndDataInput struct {
	statusCode int
	data       interface{}
}

type GetDatasetsForUserOutput struct {
	datasets []*upload.Upload
	err      error
}

type TestDataStoreSession struct {
	GetDatasetsForUserInputs  []string
	GetDatasetsForUserOutputs []GetDatasetsForUserOutput
}

func (t *TestDataStoreSession) IsClosed() bool {
	panic("Unexpected invocation of IsClosed on TestDataStoreSession")
}

func (t *TestDataStoreSession) Close() {
	panic("Unexpected invocation of Close on TestDataStoreSession")
}

func (t *TestDataStoreSession) GetDatasetsForUser(userID string) ([]*upload.Upload, error) {
	t.GetDatasetsForUserInputs = append(t.GetDatasetsForUserInputs, userID)
	output := t.GetDatasetsForUserOutputs[0]
	t.GetDatasetsForUserOutputs = t.GetDatasetsForUserOutputs[1:]
	return output.datasets, output.err
}

func (t *TestDataStoreSession) GetDataset(datasetID string) (*upload.Upload, error) {
	panic("Unexpected invocation of GetDataset on TestDataStoreSession")
}

func (t *TestDataStoreSession) CreateDataset(dataset *upload.Upload) error {
	panic("Unexpected invocation of CreateDataset on TestDataStoreSession")
}

func (t *TestDataStoreSession) UpdateDataset(dataset *upload.Upload) error {
	panic("Unexpected invocation of UpdateDataset on TestDataStoreSession")
}

func (t *TestDataStoreSession) DeleteDataset(datasetID string) error {
	panic("Unexpected invocation of DeleteDataset on TestDataStoreSession")
}

func (t *TestDataStoreSession) CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error {
	panic("Unexpected invocation of CreateDatasetData on TestDataStoreSession")
}

func (t *TestDataStoreSession) ActivateAllDatasetData(dataset *upload.Upload) error {
	panic("Unexpected invocation of ActivateAllDatasetData on TestDataStoreSession")
}

func (t *TestDataStoreSession) DeleteAllOtherDatasetData(dataset *upload.Upload) error {
	panic("Unexpected invocation of DeleteAllOtherDatasetData on TestDataStoreSession")
}

func (t *TestDataStoreSession) ValidateTest() bool {
	return len(t.GetDatasetsForUserOutputs) == 0
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
	RequestImpl                            *rest.Request
	RespondWithErrorInputs                 []*service.Error
	RespondWithInternalServerFailureInputs []RespondWithInternalServerFailureInput
	RespondWithStatusAndDataInputs         []RespondWithStatusAndDataInput
	MetricServicesClientImpl               *TestMetricServicesClient
	UserServicesClientImpl                 *TestUserServicesClient
	DataStoreSessionImpl                   *TestDataStoreSession
	AuthenticationDetailsImpl              *TestAuthenticationDetails
}

func NewTestContext() *TestContext {
	return &TestContext{
		RequestImpl: &rest.Request{
			PathParams: map[string]string{},
		},
		MetricServicesClientImpl:  &TestMetricServicesClient{},
		UserServicesClientImpl:    &TestUserServicesClient{},
		DataStoreSessionImpl:      &TestDataStoreSession{},
		AuthenticationDetailsImpl: &TestAuthenticationDetails{},
	}
}

func (t *TestContext) Logger() log.Logger {
	panic("Unexpected invocation of Logger on TestContext")
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
	panic("Unexpected invocation of RespondWithStatusAndErrors on TestContext")
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

func (t *TestContext) DataStoreSession() store.Session {
	return t.DataStoreSessionImpl
}

func (t *TestContext) DataDeduplicatorFactory() deduplicator.Factory {
	panic("Unexpected invocation of DataDeduplicatorFactory on TestContext")
}

func (t *TestContext) AuthenticationDetails() userservicesClient.AuthenticationDetails {
	return t.AuthenticationDetailsImpl
}

func (t *TestContext) SetAuthenticationDetails(authenticationDetails userservicesClient.AuthenticationDetails) {
	panic("Unexpected invocation of SetAuthenticationDetails on TestContext")
}

func (t *TestContext) ValidateTest() bool {
	return (t.UserServicesClientImpl == nil || t.UserServicesClientImpl.ValidateTest()) &&
		(t.DataStoreSessionImpl == nil || t.DataStoreSessionImpl.ValidateTest()) &&
		(t.AuthenticationDetailsImpl == nil || t.AuthenticationDetailsImpl.ValidateTest())
}
