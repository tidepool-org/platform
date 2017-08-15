package test

import (
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/test"
)

type Reporter struct {
	*test.Mock
	Config map[string]string
}

func NewReporter() *Reporter {
	return &Reporter{
		Mock:   test.NewMock(),
		Config: map[string]string{},
	}
}

func (r *Reporter) Get(key string) (string, bool) {
	value, found := r.Config[key]
	return value, found
}

func (r *Reporter) GetWithDefault(key string, defaultValue string) string {
	if value, found := r.Get(key); found {
		return value
	}

	return defaultValue
}

func (r *Reporter) Set(key string, value string) {
	r.Config[key] = value
}

func (r *Reporter) Delete(key string) {
	delete(r.Config, key)
}

func (r *Reporter) WithScopes(scopes ...string) config.Reporter {
	panic("Unexpected invocation of WithScopes on Reporter")
}
