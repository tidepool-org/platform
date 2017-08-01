package test

import "github.com/tidepool-org/platform/config"

type Reporter struct {
	Config map[string]string
}

func NewReporter() *Reporter {
	return &Reporter{
		Config: map[string]string{},
	}
}

func (r *Reporter) String(key string) (string, bool) {
	value, found := r.Config[key]
	return value, found
}

func (r *Reporter) StringOrDefault(key string, defaultValue string) string {
	if value, found := r.String(key); found {
		return value
	}

	return defaultValue
}

func (r *Reporter) WithScopes(scopes ...string) config.Reporter {
	panic("Unexpected invocation of WithScopes on Reporter")
}
