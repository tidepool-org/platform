package config

type Reporter interface {
	Get(key string) (string, bool)
	GetWithDefault(key string, defaultValue string) string

	Set(key string, value string)

	Delete(key string)

	WithScopes(scopes ...string) Reporter
}
