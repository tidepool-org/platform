// Code generated by MockGen. DO NOT EDIT.
// Source: runner.go
//
// Generated by this command:
//
//	mockgen -source=runner.go -destination=test/runner_mocks.go -package=test Provider
//

// Package test is a generated GoMock package.
package test

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "go.uber.org/mock/gomock"

	auth "github.com/tidepool-org/platform/auth"
	data "github.com/tidepool-org/platform/data"
	source "github.com/tidepool-org/platform/data/source"
	dexcom "github.com/tidepool-org/platform/dexcom"
	fetch "github.com/tidepool-org/platform/dexcom/fetch"
	oauth "github.com/tidepool-org/platform/oauth"
	request "github.com/tidepool-org/platform/request"
)

// MockAuthClient is a mock of AuthClient interface.
type MockAuthClient struct {
	ctrl     *gomock.Controller
	recorder *MockAuthClientMockRecorder
	isgomock struct{}
}

// MockAuthClientMockRecorder is the mock recorder for MockAuthClient.
type MockAuthClientMockRecorder struct {
	mock *MockAuthClient
}

// NewMockAuthClient creates a new mock instance.
func NewMockAuthClient(ctrl *gomock.Controller) *MockAuthClient {
	mock := &MockAuthClient{ctrl: ctrl}
	mock.recorder = &MockAuthClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAuthClient) EXPECT() *MockAuthClientMockRecorder {
	return m.recorder
}

// GetProviderSession mocks base method.
func (m *MockAuthClient) GetProviderSession(ctx context.Context, id string) (*auth.ProviderSession, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProviderSession", ctx, id)
	ret0, _ := ret[0].(*auth.ProviderSession)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProviderSession indicates an expected call of GetProviderSession.
func (mr *MockAuthClientMockRecorder) GetProviderSession(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProviderSession", reflect.TypeOf((*MockAuthClient)(nil).GetProviderSession), ctx, id)
}

// ServerSessionToken mocks base method.
func (m *MockAuthClient) ServerSessionToken() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ServerSessionToken")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ServerSessionToken indicates an expected call of ServerSessionToken.
func (mr *MockAuthClientMockRecorder) ServerSessionToken() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ServerSessionToken", reflect.TypeOf((*MockAuthClient)(nil).ServerSessionToken))
}

// UpdateProviderSession mocks base method.
func (m *MockAuthClient) UpdateProviderSession(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateProviderSession", ctx, id, update)
	ret0, _ := ret[0].(*auth.ProviderSession)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateProviderSession indicates an expected call of UpdateProviderSession.
func (mr *MockAuthClientMockRecorder) UpdateProviderSession(ctx, id, update any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateProviderSession", reflect.TypeOf((*MockAuthClient)(nil).UpdateProviderSession), ctx, id, update)
}

// MockDataClient is a mock of DataClient interface.
type MockDataClient struct {
	ctrl     *gomock.Controller
	recorder *MockDataClientMockRecorder
	isgomock struct{}
}

// MockDataClientMockRecorder is the mock recorder for MockDataClient.
type MockDataClientMockRecorder struct {
	mock *MockDataClient
}

// NewMockDataClient creates a new mock instance.
func NewMockDataClient(ctrl *gomock.Controller) *MockDataClient {
	mock := &MockDataClient{ctrl: ctrl}
	mock.recorder = &MockDataClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDataClient) EXPECT() *MockDataClientMockRecorder {
	return m.recorder
}

// CreateDataSetsData mocks base method.
func (m *MockDataClient) CreateDataSetsData(ctx context.Context, dataSetID string, datumArray []data.Datum) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateDataSetsData", ctx, dataSetID, datumArray)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateDataSetsData indicates an expected call of CreateDataSetsData.
func (mr *MockDataClientMockRecorder) CreateDataSetsData(ctx, dataSetID, datumArray any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateDataSetsData", reflect.TypeOf((*MockDataClient)(nil).CreateDataSetsData), ctx, dataSetID, datumArray)
}

// CreateUserDataSet mocks base method.
func (m *MockDataClient) CreateUserDataSet(ctx context.Context, userID string, create *data.DataSetCreate) (*data.DataSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUserDataSet", ctx, userID, create)
	ret0, _ := ret[0].(*data.DataSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUserDataSet indicates an expected call of CreateUserDataSet.
func (mr *MockDataClientMockRecorder) CreateUserDataSet(ctx, userID, create any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUserDataSet", reflect.TypeOf((*MockDataClient)(nil).CreateUserDataSet), ctx, userID, create)
}

// GetDataSet mocks base method.
func (m *MockDataClient) GetDataSet(ctx context.Context, id string) (*data.DataSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDataSet", ctx, id)
	ret0, _ := ret[0].(*data.DataSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDataSet indicates an expected call of GetDataSet.
func (mr *MockDataClientMockRecorder) GetDataSet(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDataSet", reflect.TypeOf((*MockDataClient)(nil).GetDataSet), ctx, id)
}

// UpdateDataSet mocks base method.
func (m *MockDataClient) UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*data.DataSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateDataSet", ctx, id, update)
	ret0, _ := ret[0].(*data.DataSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateDataSet indicates an expected call of UpdateDataSet.
func (mr *MockDataClientMockRecorder) UpdateDataSet(ctx, id, update any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateDataSet", reflect.TypeOf((*MockDataClient)(nil).UpdateDataSet), ctx, id, update)
}

// MockDataSourceClient is a mock of DataSourceClient interface.
type MockDataSourceClient struct {
	ctrl     *gomock.Controller
	recorder *MockDataSourceClientMockRecorder
	isgomock struct{}
}

// MockDataSourceClientMockRecorder is the mock recorder for MockDataSourceClient.
type MockDataSourceClientMockRecorder struct {
	mock *MockDataSourceClient
}

// NewMockDataSourceClient creates a new mock instance.
func NewMockDataSourceClient(ctrl *gomock.Controller) *MockDataSourceClient {
	mock := &MockDataSourceClient{ctrl: ctrl}
	mock.recorder = &MockDataSourceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDataSourceClient) EXPECT() *MockDataSourceClientMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockDataSourceClient) Get(ctx context.Context, id string) (*source.Source, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, id)
	ret0, _ := ret[0].(*source.Source)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockDataSourceClientMockRecorder) Get(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockDataSourceClient)(nil).Get), ctx, id)
}

// Update mocks base method.
func (m *MockDataSourceClient) Update(ctx context.Context, id string, condition *request.Condition, create *source.Update) (*source.Source, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, id, condition, create)
	ret0, _ := ret[0].(*source.Source)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockDataSourceClientMockRecorder) Update(ctx, id, condition, create any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockDataSourceClient)(nil).Update), ctx, id, condition, create)
}

// MockDexcomClient is a mock of DexcomClient interface.
type MockDexcomClient struct {
	ctrl     *gomock.Controller
	recorder *MockDexcomClientMockRecorder
	isgomock struct{}
}

// MockDexcomClientMockRecorder is the mock recorder for MockDexcomClient.
type MockDexcomClientMockRecorder struct {
	mock *MockDexcomClient
}

// NewMockDexcomClient creates a new mock instance.
func NewMockDexcomClient(ctrl *gomock.Controller) *MockDexcomClient {
	mock := &MockDexcomClient{ctrl: ctrl}
	mock.recorder = &MockDexcomClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDexcomClient) EXPECT() *MockDexcomClientMockRecorder {
	return m.recorder
}

// GetAlerts mocks base method.
func (m *MockDexcomClient) GetAlerts(ctx context.Context, startTime, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.AlertsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAlerts", ctx, startTime, endTime, tokenSource)
	ret0, _ := ret[0].(*dexcom.AlertsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAlerts indicates an expected call of GetAlerts.
func (mr *MockDexcomClientMockRecorder) GetAlerts(ctx, startTime, endTime, tokenSource any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAlerts", reflect.TypeOf((*MockDexcomClient)(nil).GetAlerts), ctx, startTime, endTime, tokenSource)
}

// GetCalibrations mocks base method.
func (m *MockDexcomClient) GetCalibrations(ctx context.Context, startTime, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.CalibrationsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCalibrations", ctx, startTime, endTime, tokenSource)
	ret0, _ := ret[0].(*dexcom.CalibrationsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCalibrations indicates an expected call of GetCalibrations.
func (mr *MockDexcomClientMockRecorder) GetCalibrations(ctx, startTime, endTime, tokenSource any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCalibrations", reflect.TypeOf((*MockDexcomClient)(nil).GetCalibrations), ctx, startTime, endTime, tokenSource)
}

// GetDataRange mocks base method.
func (m *MockDexcomClient) GetDataRange(ctx context.Context, lastSyncTime *time.Time, tokenSource oauth.TokenSource) (*dexcom.DataRangesResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDataRange", ctx, lastSyncTime, tokenSource)
	ret0, _ := ret[0].(*dexcom.DataRangesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDataRange indicates an expected call of GetDataRange.
func (mr *MockDexcomClientMockRecorder) GetDataRange(ctx, lastSyncTime, tokenSource any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDataRange", reflect.TypeOf((*MockDexcomClient)(nil).GetDataRange), ctx, lastSyncTime, tokenSource)
}

// GetDevices mocks base method.
func (m *MockDexcomClient) GetDevices(ctx context.Context, startTime, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.DevicesResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDevices", ctx, startTime, endTime, tokenSource)
	ret0, _ := ret[0].(*dexcom.DevicesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDevices indicates an expected call of GetDevices.
func (mr *MockDexcomClientMockRecorder) GetDevices(ctx, startTime, endTime, tokenSource any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDevices", reflect.TypeOf((*MockDexcomClient)(nil).GetDevices), ctx, startTime, endTime, tokenSource)
}

// GetEGVs mocks base method.
func (m *MockDexcomClient) GetEGVs(ctx context.Context, startTime, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EGVsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEGVs", ctx, startTime, endTime, tokenSource)
	ret0, _ := ret[0].(*dexcom.EGVsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEGVs indicates an expected call of GetEGVs.
func (mr *MockDexcomClientMockRecorder) GetEGVs(ctx, startTime, endTime, tokenSource any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEGVs", reflect.TypeOf((*MockDexcomClient)(nil).GetEGVs), ctx, startTime, endTime, tokenSource)
}

// GetEvents mocks base method.
func (m *MockDexcomClient) GetEvents(ctx context.Context, startTime, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EventsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEvents", ctx, startTime, endTime, tokenSource)
	ret0, _ := ret[0].(*dexcom.EventsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEvents indicates an expected call of GetEvents.
func (mr *MockDexcomClientMockRecorder) GetEvents(ctx, startTime, endTime, tokenSource any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEvents", reflect.TypeOf((*MockDexcomClient)(nil).GetEvents), ctx, startTime, endTime, tokenSource)
}

// MockProvider is a mock of Provider interface.
type MockProvider struct {
	ctrl     *gomock.Controller
	recorder *MockProviderMockRecorder
	isgomock struct{}
}

// MockProviderMockRecorder is the mock recorder for MockProvider.
type MockProviderMockRecorder struct {
	mock *MockProvider
}

// NewMockProvider creates a new mock instance.
func NewMockProvider(ctrl *gomock.Controller) *MockProvider {
	mock := &MockProvider{ctrl: ctrl}
	mock.recorder = &MockProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProvider) EXPECT() *MockProviderMockRecorder {
	return m.recorder
}

// AuthClient mocks base method.
func (m *MockProvider) AuthClient() fetch.AuthClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AuthClient")
	ret0, _ := ret[0].(fetch.AuthClient)
	return ret0
}

// AuthClient indicates an expected call of AuthClient.
func (mr *MockProviderMockRecorder) AuthClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthClient", reflect.TypeOf((*MockProvider)(nil).AuthClient))
}

// DataClient mocks base method.
func (m *MockProvider) DataClient() fetch.DataClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DataClient")
	ret0, _ := ret[0].(fetch.DataClient)
	return ret0
}

// DataClient indicates an expected call of DataClient.
func (mr *MockProviderMockRecorder) DataClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DataClient", reflect.TypeOf((*MockProvider)(nil).DataClient))
}

// DataSourceClient mocks base method.
func (m *MockProvider) DataSourceClient() fetch.DataSourceClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DataSourceClient")
	ret0, _ := ret[0].(fetch.DataSourceClient)
	return ret0
}

// DataSourceClient indicates an expected call of DataSourceClient.
func (mr *MockProviderMockRecorder) DataSourceClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DataSourceClient", reflect.TypeOf((*MockProvider)(nil).DataSourceClient))
}

// DexcomClient mocks base method.
func (m *MockProvider) DexcomClient() fetch.DexcomClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DexcomClient")
	ret0, _ := ret[0].(fetch.DexcomClient)
	return ret0
}

// DexcomClient indicates an expected call of DexcomClient.
func (mr *MockProviderMockRecorder) DexcomClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DexcomClient", reflect.TypeOf((*MockProvider)(nil).DexcomClient))
}

// GetRunnerDurationMaximum mocks base method.
func (m *MockProvider) GetRunnerDurationMaximum() time.Duration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRunnerDurationMaximum")
	ret0, _ := ret[0].(time.Duration)
	return ret0
}

// GetRunnerDurationMaximum indicates an expected call of GetRunnerDurationMaximum.
func (mr *MockProviderMockRecorder) GetRunnerDurationMaximum() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRunnerDurationMaximum", reflect.TypeOf((*MockProvider)(nil).GetRunnerDurationMaximum))
}
