// Code generated by MockGen. DO NOT EDIT.
// Source: client.go
//
// Generated by this command:
//
//	mockgen -source=client.go -destination=test/client.go -package test Client
//

// Package test is a generated GoMock package.
package test

import (
	context "context"
	io "io"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"

	raw "github.com/tidepool-org/platform/data/raw"
	page "github.com/tidepool-org/platform/page"
	request "github.com/tidepool-org/platform/request"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
	isgomock struct{}
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockClient) Create(ctx context.Context, userID, dataSetID string, create *raw.Create, data io.Reader) (*raw.Raw, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, userID, dataSetID, create, data)
	ret0, _ := ret[0].(*raw.Raw)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockClientMockRecorder) Create(ctx, userID, dataSetID, create, data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockClient)(nil).Create), ctx, userID, dataSetID, create, data)
}

// Delete mocks base method.
func (m *MockClient) Delete(ctx context.Context, id string, condition *request.Condition) (*raw.Raw, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, id, condition)
	ret0, _ := ret[0].(*raw.Raw)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Delete indicates an expected call of Delete.
func (mr *MockClientMockRecorder) Delete(ctx, id, condition any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockClient)(nil).Delete), ctx, id, condition)
}

// DeleteAllByDataSetID mocks base method.
func (m *MockClient) DeleteAllByDataSetID(ctx context.Context, dataSetID string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAllByDataSetID", ctx, dataSetID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteAllByDataSetID indicates an expected call of DeleteAllByDataSetID.
func (mr *MockClientMockRecorder) DeleteAllByDataSetID(ctx, dataSetID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAllByDataSetID", reflect.TypeOf((*MockClient)(nil).DeleteAllByDataSetID), ctx, dataSetID)
}

// DeleteAllByUserID mocks base method.
func (m *MockClient) DeleteAllByUserID(ctx context.Context, userID string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAllByUserID", ctx, userID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteAllByUserID indicates an expected call of DeleteAllByUserID.
func (mr *MockClientMockRecorder) DeleteAllByUserID(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAllByUserID", reflect.TypeOf((*MockClient)(nil).DeleteAllByUserID), ctx, userID)
}

// DeleteMultiple mocks base method.
func (m *MockClient) DeleteMultiple(ctx context.Context, ids []string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteMultiple", ctx, ids)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteMultiple indicates an expected call of DeleteMultiple.
func (mr *MockClientMockRecorder) DeleteMultiple(ctx, ids any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMultiple", reflect.TypeOf((*MockClient)(nil).DeleteMultiple), ctx, ids)
}

// Get mocks base method.
func (m *MockClient) Get(ctx context.Context, id string, condition *request.Condition) (*raw.Raw, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, id, condition)
	ret0, _ := ret[0].(*raw.Raw)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockClientMockRecorder) Get(ctx, id, condition any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockClient)(nil).Get), ctx, id, condition)
}

// GetContent mocks base method.
func (m *MockClient) GetContent(ctx context.Context, id string, condition *request.Condition) (*raw.Content, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContent", ctx, id, condition)
	ret0, _ := ret[0].(*raw.Content)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContent indicates an expected call of GetContent.
func (mr *MockClientMockRecorder) GetContent(ctx, id, condition any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContent", reflect.TypeOf((*MockClient)(nil).GetContent), ctx, id, condition)
}

// List mocks base method.
func (m *MockClient) List(ctx context.Context, userID string, filter *raw.Filter, pagination *page.Pagination) ([]*raw.Raw, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, userID, filter, pagination)
	ret0, _ := ret[0].([]*raw.Raw)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockClientMockRecorder) List(ctx, userID, filter, pagination any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockClient)(nil).List), ctx, userID, filter, pagination)
}