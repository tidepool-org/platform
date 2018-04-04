package request

import (
	"net/http"

	"github.com/tidepool-org/platform/errors"
)

type Mutator interface {
	Mutate(req *http.Request) error
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

func (h *HeaderMutator) Mutate(req *http.Request) error {
	if req == nil {
		return errors.New("request is missing")
	}
	if h.Key == "" {
		return errors.New("key is missing")
	}

	req.Header.Add(h.Key, h.Value)
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

func (p *ParameterMutator) Mutate(req *http.Request) error {
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

func (p *ParametersMutator) Mutate(req *http.Request) error {
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
