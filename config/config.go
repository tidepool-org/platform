package config

type Reporter interface {
	String(key string) (string, bool)
	StringOrDefault(key string, defaultValue string) string

	WithScopes(scopes ...string) Reporter
}
