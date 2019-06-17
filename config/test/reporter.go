package test

import (
	"fmt"
	"os"

	"github.com/tidepool-org/platform/config"
)

type Reporter struct {
	Scopes []string
	Config map[string]interface{}
	Debug  bool
}

func NewReporter() *Reporter {
	return &Reporter{
		Scopes: []string{},
		Config: map[string]interface{}{},
	}
}

func (r *Reporter) Get(key string) (string, error) {
	raw, found := r.Config[key]
	if !found {
		r.debug("Reporter.Get(%q) returning error ErrorKeyNotFound", key)
		return "", config.ErrorKeyNotFound(key)
	}
	value, ok := raw.(string)
	if !ok {
		r.debug("Reporter.Get(%q) returning error ErrorKeyNotFound", key)
		return "", config.ErrorKeyNotFound(key)
	}
	r.debug("Reporter.Get(%q) returning value %q", key, value)
	return value, nil
}

func (r *Reporter) GetWithDefault(key string, defaultValue string) string {
	value, err := r.Get(key)
	if err != nil {
		r.debug("Reporter.GetWithDefault(%q, %q) returning default value", key, defaultValue)
		return defaultValue
	}
	r.debug("Reporter.GetWithDefault(%q, %q) returning value %q", key, defaultValue, value)
	return value
}

func (r *Reporter) Set(key string, value string) {
	r.debug("Reporter.Set(%q, %q)", key, value)
	r.Config[key] = value
}

func (r *Reporter) Delete(key string) {
	r.debug("Reporter.Delete(%q)", key)
	delete(r.Config, key)
}

func (r *Reporter) WithScopes(scopes ...string) config.Reporter {
	r.debug("Reporter.WithScopes(%#v) returning new Reporter", scopes)
	config := r.Config
	for _, scope := range scopes {
		raw, _ := config[scope]
		config, _ = raw.(map[string]interface{})
	}
	if config == nil {
		config = map[string]interface{}{}
	}
	return &Reporter{
		Scopes: append(r.Scopes, scopes...),
		Config: config,
	}
}

func (r *Reporter) debug(format string, a ...interface{}) {
	if r.Debug {
		fmt.Fprintf(os.Stderr, "DEBUG: %s with scopes %#v\n", fmt.Sprintf(format, a...), r.Scopes)
	}
}
