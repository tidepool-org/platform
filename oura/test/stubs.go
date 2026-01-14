package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
)

func NewStubServer(resp *StubResponses) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, stubbed := range resp.responses {
			if matches := stubbed.matchers.MatchesAll(r); matches {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(stubbed.response.StatusCode)
				w.Write([]byte(stubbed.response.Body))
				return
			}
		}

		resp.unmatchedResponses++

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
		return
	}))
}

type RequestMatchers []RequestMatcher

func (rm RequestMatchers) MatchesAll(r *http.Request) bool {
	for _, m := range rm {
		if !m(r) {
			return false
		}
	}
	return true
}

type RequestMatcher func(r *http.Request) bool

func NewRequestMethodAndPathMatcher(method, path string) RequestMatcher {
	return func(r *http.Request) bool {
		return r.Method == method && r.URL.Path == path
	}
}

func NewRequestJSONBodyMatcher(expected string) RequestMatcher {
	return func(r *http.Request) bool {
		expectedJSON := map[string]interface{}{}
		err := json.Unmarshal([]byte(expected), &expectedJSON)
		if err != nil {
			panic(err)
		}

		actualJSON := map[string]interface{}{}
		if err := json.NewDecoder(r.Body).Decode(&actualJSON); err != nil {
			panic(err)
		}

		return reflect.DeepEqual(expectedJSON, actualJSON)
	}
}

type StubResponses struct {
	responses          []StubResponse
	unmatchedResponses int
}

func NewStubResponses() *StubResponses {
	return &StubResponses{}
}

func (s *StubResponses) AddResponse(matchers []RequestMatcher, response Response) {
	s.responses = append(s.responses, StubResponse{matchers: matchers, response: response})
}

func (s *StubResponses) UnmatchedResponses() int {
	return s.unmatchedResponses
}

type StubResponse struct {
	matchers RequestMatchers
	response Response
}

type Response struct {
	StatusCode int
	Body       string
}
