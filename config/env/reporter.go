package env

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"go.uber.org/fx"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
)

// NOTE: To debug Reporter issues, set the environment value TIDEPOOL_CONFIG_REPORTER_DEBUG to "true".
// In Kubernetes, the environment variable can be manually set on the Deployment (not Pod). Restart the
// deployment after setting the environment variable to have the related Pods pickup the change. For example:
//
// $ kubectl set env deployment/auth TIDEPOOL_CONFIG_REPORTER_DEBUG=true
// $ kubectl rollout restart deployment/auth

var isValidPrefix = regexp.MustCompile("^[A-Z][A-Z0-9_]*$").MatchString
var replaceInvalidCharacters = regexp.MustCompile("[^A-Z0-9_]").ReplaceAllString

var Module = fx.Provide(NewDefaultReporter)

func NewDefaultReporter() (config.Reporter, error) {
	return NewReporter("TIDEPOOL")
}

func NewReporter(prefix string) (config.Reporter, error) {
	if prefix == "" {
		return nil, errors.New("prefix is missing")
	}
	if !isValidPrefix(prefix) {
		return nil, errors.New("prefix is invalid")
	}

	reporter := &reporter{
		prefix: prefix,
		scopes: []string{},
	}

	if debug, err := strconv.ParseBool(reporter.WithScopes("config_reporter").GetWithDefault("debug", "false")); err == nil {
		reporter.debug = debug
	}

	return reporter, nil
}

type reporter struct {
	prefix string
	scopes []string
	debug  bool
}

func (r *reporter) Get(key string) (string, error) {
	value, err := r.get(key)
	if err != nil {
		r.debugMessage("NOT FOUND", r.getKey(r.scopes, key), "", nil)
		return "", err
	}
	return value, nil
}

func (r *reporter) GetWithDefault(key string, defaultValue string) string {
	value, err := r.get(key)
	if err != nil {
		r.debugMessage("DEFAULT", r.getKey(r.scopes, key), "", &defaultValue)
		return defaultValue
	}
	return value
}

func (r *reporter) get(key string) (string, error) {
	limit := len(r.scopes) - 1
	if limit < 0 {
		limit = 0
	}

	fullKey := r.getKey(r.scopes, key)
	for i := 0; i <= limit; i++ {
		matchKey := r.getKey(r.scopes[i:], key)
		if value, found := syscall.Getenv(matchKey); found {
			r.debugMessage("FOUND", fullKey, matchKey, &value)
			return value, nil
		}
	}

	return "", config.ErrorKeyNotFound(fullKey)
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
		debug:  r.debug,
	}
}

func (r *reporter) getKey(scopes []string, key string) string {
	return GetKey(r.prefix, scopes, key)
}

func (r *reporter) debugMessage(state string, fullKey string, matchKey string, value *string) {
	if r.debug {
		if value != nil {
			value = pointer.FromString(fmt.Sprintf("%q", *value))
		}
		fmt.Printf("[ConfigReporter] | %10s | %-80s | %-80s | %s\n", state, fullKey, matchKey, *pointer.DefaultString(value, ""))
	}
}

func GetKey(prefix string, scopes []string, key string) string {
	parts := append(append([]string{prefix}, scopes...), key)
	return replaceInvalidCharacters(strings.ToUpper(strings.Join(parts, "_")), "_")
}
