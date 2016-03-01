package service

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
)

// MakeSimpleRequest returns a http.Request. The returned request object can be
// further prepared by adding headers and query string parmaters, for instance.
func MakeSimpleRequest(method string, urlStr string, body io.Reader) *http.Request {

	r, err := http.NewRequest(method, urlStr, body)

	if err != nil {
		panic(err)
	}
	r.Header.Set("Accept-Encoding", "gzip")
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}

	return r
}

// MakeBlobRequest returns a http.Request. The returned request object can be
// further prepared by adding headers and query string parmaters, for instance.
func MakeBlobRequest(method string, urlStr string, body io.Reader) *http.Request {

	r, err := http.NewRequest(method, urlStr, body)

	if err != nil {
		panic(err)
	}
	r.Header.Set("Accept-Encoding", "gzip")
	if body != nil {
		r.Header.Set("Content-Type", "mime/multipart")
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

// BodyContains compares the rescorded body
func BodyContains(r *httptest.ResponseRecorder, expectedToContain string) bool {
	return strings.Contains(r.Body.String(), expectedToContain)
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
func adapterFunc(handler rest.HandlerFunc, env map[string]interface{}, pathParams map[string]string) http.HandlerFunc {

	return func(origWriter http.ResponseWriter, origRequest *http.Request) {

		// instantiate the rest objects
		request := &rest.Request{
			Request:    origRequest,
			PathParams: pathParams,
			Env:        env,
		}

		writer := &responseWriter{
			ResponseWriter: origWriter,
			wroteHeader:    false,
		}

		// call the wrapped handler
		handler(writer, request)
	}
}

// RunRequest runs a HTTP request through the given handler
func RunRequest(restHandler rest.HandlerFunc, request *http.Request, pathParams map[string]string, env map[string]interface{}) *Recorded {
	handler := adapterFunc(restHandler, env, pathParams)
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
	return ContentTypeIsJSON(rd.Recorder)
}

// ContentEncodingIsGzip for Recorded
func (rd *Recorded) ContentEncodingIsGzip() bool {
	return ContentEncodingIsGzip(rd.Recorder)
}

// BodyIs for Recorded
func (rd *Recorded) BodyIs(expectedToContain string) bool {
	return BodyIs(rd.Recorder, expectedToContain)
}

// BodyContains for Recorded
func (rd *Recorded) BodyContains(expectedBody string) bool {
	return BodyIs(rd.Recorder, expectedBody)
}

// DecodeJSONPayload for Recorded
func (rd *Recorded) DecodeJSONPayload(v interface{}) error {
	return DecodeJSONPayload(rd.Recorder, v)
}
