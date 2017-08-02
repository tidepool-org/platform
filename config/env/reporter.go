package env

import (
	"regexp"
	"strings"
	"syscall"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

var isValidPrefix = regexp.MustCompile("^[A-Z][A-Z0-9_]*$").MatchString
var replaceInvalidCharacters = regexp.MustCompile("[^A-Z0-9_]").ReplaceAllString

func NewReporter(prefix string) (config.Reporter, error) {
	if prefix == "" {
		return nil, errors.New("env", "prefix is missing")
	}
	if !isValidPrefix(prefix) {
		return nil, errors.New("env", "prefix is invalid")
	}

	return &reporter{
		prefix: prefix,
		scopes: []string{},
	}, nil
}

type reporter struct {
	prefix string
	scopes []string
}

func (r *reporter) Get(key string) (string, bool) {
	limit := len(r.scopes) - 1
	if limit < 0 {
		limit = 0
	}

	for i := 0; i <= limit; i++ {
		if value, found := syscall.Getenv(r.getEnvKey(r.scopes[i:], key)); found {
			return value, true
		}
	}

	return "", false
}

func (r *reporter) GetWithDefault(key string, defaultValue string) string {
	if value, found := r.Get(key); found {
		return value
	}

	return defaultValue
}

func (r *reporter) Set(key string, value string) {
	syscall.Setenv(r.getEnvKey(r.scopes, key), value) // Safely ignore error; cannot fail
}

func (r *reporter) Delete(key string) {
	syscall.Unsetenv(r.getEnvKey(r.scopes, key)) // Safely ignore error; cannot fail
}

func (r *reporter) WithScopes(scopes ...string) config.Reporter {
	return &reporter{
		prefix: r.prefix,
		scopes: append(r.scopes, scopes...),
	}
}

func (r *reporter) getEnvKey(scopes []string, key string) string {
	parts := append(append([]string{r.prefix}, scopes...), key)
	return replaceInvalidCharacters(strings.ToUpper(strings.Join(parts, "_")), "_")
}
