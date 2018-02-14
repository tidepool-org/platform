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
		return nil, errors.New("prefix is missing")
	}
	if !isValidPrefix(prefix) {
		return nil, errors.New("prefix is invalid")
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

func (r *reporter) Get(key string) (string, error) {
	limit := len(r.scopes) - 1
	if limit < 0 {
		limit = 0
	}

	for i := 0; i <= limit; i++ {
		if value, found := syscall.Getenv(r.getKey(r.scopes[i:], key)); found {
			return value, nil
		}
	}

	return "", config.ErrorKeyNotFound(r.getKey(r.scopes, key))
}

func (r *reporter) GetWithDefault(key string, defaultValue string) string {
	value, err := r.Get(key)
	if err != nil {
		return defaultValue
	}
	return value
}

func (r *reporter) Set(key string, value string) {
	syscall.Setenv(r.getKey(r.scopes, key), value) // Safely ignore error; cannot fail
}

func (r *reporter) Delete(key string) {
	syscall.Unsetenv(r.getKey(r.scopes, key)) // Safely ignore error; cannot fail
}

func (r *reporter) WithScopes(scopes ...string) config.Reporter {
	return &reporter{
		prefix: r.prefix,
		scopes: append(r.scopes, scopes...),
	}
}

func (r *reporter) getKey(scopes []string, key string) string {
	return GetKey(r.prefix, scopes, key)
}

func GetKey(prefix string, scopes []string, key string) string {
	parts := append(append([]string{prefix}, scopes...), key)
	return replaceInvalidCharacters(strings.ToUpper(strings.Join(parts, "_")), "_")
}
