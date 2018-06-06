package test

import (
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/test"
)

type Reporter struct {
	*test.Mock
	Config map[string]interface{}
}

func NewReporter() *Reporter {
	return &Reporter{
		Mock:   test.NewMock(),
		Config: map[string]interface{}{},
	}
}

func (r *Reporter) Get(key string) (string, error) {
	raw, found := r.Config[key]
	if !found {
		return "", config.ErrorKeyNotFound(key)
	}
	value, ok := raw.(string)
	if !ok {
		return "", config.ErrorKeyNotFound(key)
	}
	return value, nil
}

func (r *Reporter) GetWithDefault(key string, defaultValue string) string {
	value, err := r.Get(key)
	if err != nil {
		return defaultValue
	}
	return value
}

func (r *Reporter) Set(key string, value string) {
	r.Config[key] = value
}

func (r *Reporter) Delete(key string) {
	delete(r.Config, key)
}

func (r *Reporter) WithScopes(scopes ...string) config.Reporter {
	config := r.Config
	for _, scope := range scopes {
		raw, _ := config[scope]
		config, _ = raw.(map[string]interface{})
	}
	if config == nil {
		config = map[string]interface{}{}
	}
	return &Reporter{
		Mock:   test.NewMock(),
		Config: config,
	}
}
