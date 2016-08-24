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
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dataservices/service/api/v1")
}

type GetUserPermissionsInput struct {
	context       service.Context
	requestUserID string
	targetUserID  string
}

type GetUserPermissionsOutput struct {
	permissions client.Permissions
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

func (t *TestUserServicesClient) ValidateUserSession(context service.Context, sessionToken string) (string, error) {
	panic("Unexpected invocation of ValidateUserSession on TestUserServicesClient")
}

func (t *TestUserServicesClient) GetUserPermissions(context service.Context, requestUserID string, targetUserID string) (client.Permissions, error) {
	t.GetUserPermissionsInputs = append(t.GetUserPermissionsInputs, GetUserPermissionsInput{context, requestUserID, targetUserID})
	output := t.GetUserPermissionsOutputs[0]
	t.GetUserPermissionsOutputs = t.GetUserPermissionsOutputs[1:]
	return output.permissions, output.err
}

func (t *TestUserServicesClient) GetUserGroupID(context service.Context, userID string) (string, error) {
	panic("Unexpected invocation of GetUserGroupID on TestUserServicesClient")
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

type TestContext struct {
	RequestImpl                            *rest.Request
	RespondWithErrorInputs                 []*service.Error
	RespondWithInternalServerFailureInputs []RespondWithInternalServerFailureInput
	RespondWithStatusAndDataInputs         []RespondWithStatusAndDataInput
	DataStoreSessionImpl                   *TestDataStoreSession
	UserServicesClientImpl                 *TestUserServicesClient
	requestUserID                          string
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

func (t *TestContext) DataFactory() data.Factory {
	panic("Unexpected invocation of DataFactory on TestContext")
}

func (t *TestContext) DataStoreSession() store.Session {
	return t.DataStoreSessionImpl
}

func (t *TestContext) DataDeduplicatorFactory() deduplicator.Factory {
	panic("Unexpected invocation of DataDeduplicatorFactory on TestContext")
}

func (t *TestContext) UserServicesClient() client.Client {
	return t.UserServicesClientImpl
}

func (t *TestContext) RequestUserID() string {
	return t.requestUserID
}

func (t *TestContext) SetRequestUserID(requestUserID string) {
	t.requestUserID = requestUserID
}

func NewTestContext() *TestContext {
	return &TestContext{
		RequestImpl: &rest.Request{
			PathParams: map[string]string{},
		},
		UserServicesClientImpl: &TestUserServicesClient{},
		DataStoreSessionImpl:   &TestDataStoreSession{},
	}
}
