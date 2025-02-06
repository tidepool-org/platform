package test

import (
	"encoding/json"
	"net/http/httptest"

	"github.com/tidepool-org/platform/request"
)

var (
	TestUserID1 = "62a372fa-7096-4d33-ab3a-1f26d7701f76"
	TestUserID2 = "89d13ccb-32fb-47ef-9a8c-9d45f5d1c145"
	TestToken1  = "token1"
	TestToken2  = "token2"
)

// MockRestResponseWriter implements rest.ResponseWriter
type MockRestResponseWriter struct {
	*httptest.ResponseRecorder
}

func NewMockRestResponseWriter() *MockRestResponseWriter {
	return &MockRestResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
	}
}

func (w *MockRestResponseWriter) WriteJson(v interface{}) error {
	data, err := w.EncodeJson(v)
	if err != nil {
		return err
	}
	if _, err := w.ResponseRecorder.Write(data); err != nil {
		return err
	}
	return nil
}

func (w *MockRestResponseWriter) EncodeJson(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MockAuthDetails implements request.MockAuthDetails with test helpers.
type MockAuthDetails struct {
	request.AuthDetails
}

func NewMockAuthDetailsDefault() *MockAuthDetails {
	return NewMockAuthDetails(request.MethodSessionToken, TestUserID1, TestToken1)
}

func NewMockAuthDetails(authMethod request.Method, userID, token string) *MockAuthDetails {
	return &MockAuthDetails{
		AuthDetails: request.NewAuthDetails(authMethod, userID, token),
	}
}
