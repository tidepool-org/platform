package request

import (
	"net/http"

	"github.com/tidepool-org/platform/errors"
)

type RequestMutator interface {
	MutateRequest(req *http.Request) error
}

type ResponseMutator interface {
	MutateResponse(res http.ResponseWriter) error
}

type HeaderMutator struct {
	Key   string
	Value string
}

func NewHeaderMutator(key string, value string) *HeaderMutator {
	return &HeaderMutator{
		Key:   key,
		Value: value,
	}
}

func (h *HeaderMutator) MutateRequest(req *http.Request) error {
	if req == nil {
		return errors.New("request is missing")
	}
	if h.Key == "" {
		return errors.New("key is missing")
	}

	req.Header.Add(h.Key, h.Value)
	return nil
}

func (h *HeaderMutator) MutateResponse(res http.ResponseWriter) error {
	if res == nil {
		return errors.New("response is missing")
	}
	if h.Key == "" {
		return errors.New("key is missing")
	}

	res.Header().Add(h.Key, h.Value)
	return nil
}

type ParameterMutator struct {
	Key   string
	Value string
}

func NewParameterMutator(key string, value string) *ParameterMutator {
	return &ParameterMutator{
		Key:   key,
		Value: value,
	}
}

func (p *ParameterMutator) MutateRequest(req *http.Request) error {
	if req == nil {
		return errors.New("request is missing")
	}
	if p.Key == "" {
		return errors.New("key is missing")
	}

	query := req.URL.Query()
	query.Add(p.Key, p.Value)
	req.URL.RawQuery = query.Encode()
	return nil
}

type ParametersMutator struct {
	Parameters map[string]string
}

func NewParametersMutator(parameters map[string]string) *ParametersMutator {
	return &ParametersMutator{
		Parameters: parameters,
	}
}

func (p *ParametersMutator) MutateRequest(req *http.Request) error {
	if req == nil {
		return errors.New("request is missing")
	}

	query := req.URL.Query()
	for key, value := range p.Parameters {
		if key == "" {
			return errors.New("key is missing")
		}
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()

	return nil
}

// ArrayParametersMutator mutates a request by modifying its query string to contain Parameters.
type ArrayParametersMutator struct {
	Parameters map[string][]string
}

func NewArrayParametersMutator(parameters map[string][]string) *ArrayParametersMutator {
	return &ArrayParametersMutator{
		Parameters: parameters,
	}
}

func (p *ArrayParametersMutator) MutateRequest(req *http.Request) error {
	if req == nil {
		return errors.New("request is missing")
	}

	query := req.URL.Query()
	for key, values := range p.Parameters {
		if key == "" {
			return errors.New("key is missing")
		}
		for _, value := range values {
			query.Add(key, value)
		}
	}
	req.URL.RawQuery = query.Encode()

	return nil
}
