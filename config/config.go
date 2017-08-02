package config

type Reporter interface {
	Get(key string) (string, bool)
	GetWithDefault(key string, defaultValue string) string

	WithScopes(scopes ...string) Reporter
}
