package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
)

// MakeSimpleRequest returns a http.Request. The returned request object can be
// further prepared by adding headers and query string parmaters, for instance.
func MakeSimpleRequest(method string, urlStr string, payload interface{}) *http.Request {
	var s string

	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			panic(err)
		}
		s = fmt.Sprintf("%s", b)
	}

	r, err := http.NewRequest(method, urlStr, strings.NewReader(s))
	if err != nil {
		panic(err)
	}
	r.Header.Set("Accept-Encoding", "gzip")
	if payload != nil {
		r.Header.Set("Content-Type", "application/json")
	}

	return r
}

// CodeIs compares the rescorded status code
func CodeIs(r *httptest.ResponseRecorder, expectedCode int) bool {
	return r.Code == expectedCode
}

// HeaderIs tests the first value for the given headerKey
func HeaderIs(r *httptest.ResponseRecorder, headerKey, expectedValue string) bool {
	value := r.HeaderMap.Get(headerKey)
	return value == expectedValue
}

// ContentTypeIsJSON tests that application/json is set
func ContentTypeIsJSON(r *httptest.ResponseRecorder) bool {
	return HeaderIs(r, "Content-Type", "application/json")
}

// ContentEncodingIsGzip tests that gzip is set
func ContentEncodingIsGzip(r *httptest.ResponseRecorder) bool {
	return HeaderIs(r, "Content-Encoding", "gzip")
}

// BodyIs compares the rescorded body
func BodyIs(r *httptest.ResponseRecorder, expectedBody string) bool {
	body := r.Body.String()
	return strings.Trim(body, "\"") == expectedBody
}

// DecodeJSONPayload decodes the recorded payload to JSON
func DecodeJSONPayload(r *httptest.ResponseRecorder, v interface{}) error {
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, v)
	if err != nil {
		return err
	}
	return nil
}

// Recorded type
type Recorded struct {
	Recorder *httptest.ResponseRecorder
}

// Private responseWriter intantiated by the resource handler.
// It implements the following interfaces:
// ResponseWriter
// http.ResponseWriter
type responseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *responseWriter) WriteHeader(code int) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.ResponseWriter.WriteHeader(code)
	w.wroteHeader = true
}

func (w *responseWriter) EncodeJson(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Encode the object in JSON and call Write.
func (w *responseWriter) WriteJson(v interface{}) error {
	b, err := w.EncodeJson(v)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// Provided in order to implement the http.ResponseWriter interface.
func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// Handle the transition between net/http and go-json-rest objects.
// It intanciates the rest.Request and rest.ResponseWriter, ...
func adapterFunc(handler rest.HandlerFunc) http.HandlerFunc {

	return func(origWriter http.ResponseWriter, origRequest *http.Request) {

		// instantiate the rest objects
		request := &rest.Request{
			origRequest,
			nil,
			map[string]interface{}{},
		}

		writer := &responseWriter{
			origWriter,
			false,
		}

		// call the wrapped handler
		handler(writer, request)
	}
}

// RunRequest runs a HTTP request through the given handler
func RunRequest(restHandler rest.HandlerFunc, request *http.Request) *Recorded {
	handler := adapterFunc(restHandler)
	recorder := httptest.NewRecorder()
	handler(recorder, request)
	return &Recorded{recorder}
}

// CodeIs for Recorded
func (rd *Recorded) CodeIs(expectedCode int) bool {
	return CodeIs(rd.Recorder, expectedCode)
}

// HeaderIs for Recorded
func (rd *Recorded) HeaderIs(headerKey, expectedValue string) bool {
	return HeaderIs(rd.Recorder, headerKey, expectedValue)
}

// ContentTypeIsJSON for Recorded
func (rd *Recorded) ContentTypeIsJSON() bool {
	return rd.HeaderIs("Content-Type", "application/json")
}

// ContentEncodingIsGzip for Recorded
func (rd *Recorded) ContentEncodingIsGzip() bool {
	return rd.HeaderIs("Content-Encoding", "gzip")
}

// BodyIs for Recorded
func (rd *Recorded) BodyIs(expectedBody string) bool {
	return BodyIs(rd.Recorder, expectedBody)
}

// DecodeJSONPayload for Recorded
func (rd *Recorded) DecodeJSONPayload(v interface{}) error {
	return DecodeJSONPayload(rd.Recorder, v)
}
