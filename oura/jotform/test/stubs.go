package test

import (
	"net/http"
	"net/http/httptest"
)

type Response struct {
	StatusCode int
	Body       string
}

type StubResponses struct {
	responses map[string]Response
}

func NewStubResponses() *StubResponses {
	return &StubResponses{responses: make(map[string]Response)}
}

func (s *StubResponses) AddResponse(method, path string, Response Response) {
	s.responses[method+" "+path] = Response
}

func NewStubServer(resp *StubResponses) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response, ok := resp.responses[r.Method+" "+r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.StatusCode)
		w.Write([]byte(response.Body))
	}))
}
