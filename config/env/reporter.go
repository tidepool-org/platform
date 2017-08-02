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

	envPrefix := []string{r.prefix}
	for i := 0; i <= limit; i++ {
		envParts := append(append(envPrefix, r.scopes[i:]...), key)
		envKey := replaceInvalidCharacters(strings.ToUpper(strings.Join(envParts, "_")), "_")
		if value, found := syscall.Getenv(envKey); found {
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

func (r *reporter) WithScopes(scopes ...string) config.Reporter {
	return &reporter{
		prefix: r.prefix,
		scopes: append(r.scopes, scopes...),
	}
}
