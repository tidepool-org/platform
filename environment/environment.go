package environment

import (
	"fmt"
	"os"

	"github.com/tidepool-org/platform/errors"
)

type Reporter interface {
	Name() string
	IsLocal() bool
	IsTest() bool
	IsDeployed() bool

	Prefix() string
	GetKey(key string) string
	GetValue(key string) string
}

func NewReporter(name string, prefix string) (Reporter, error) {
	if name == "" {
		return nil, errors.New("environment", "name is missing")
	}

	return &reporter{
		name:   name,
		prefix: prefix,
	}, nil
}

type reporter struct {
	name   string
	prefix string
}

func (r *reporter) Name() string {
	return r.name
}

func (r *reporter) IsLocal() bool {
	return r.Name() == "local"
}

func (r *reporter) IsTest() bool {
	return r.Name() == "test"
}

func (r *reporter) IsDeployed() bool {
	return !r.IsLocal() && !r.IsTest()
}

func (r *reporter) Prefix() string {
	return r.prefix
}

func (r *reporter) GetKey(key string) string {
	return GetKey(key, r.prefix)
}

func (r *reporter) GetValue(key string) string {
	return GetValue(key, r.prefix)
}

func GetKey(key string, prefix string) string {
	if prefix == "" {
		return key
	}
	return fmt.Sprintf("%s_%s", prefix, key)
}

func GetValue(key string, prefix string) string {
	return os.Getenv(GetKey(key, prefix))
}
