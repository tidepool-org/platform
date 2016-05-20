package environment

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import "github.com/tidepool-org/platform/app"

type Reporter interface {
	Name() string
	IsLocal() bool
	IsTest() bool
	IsDeployed() bool
}

func NewReporter(name string) (Reporter, error) {
	if name == "" {
		return nil, app.Error("environment", "name is missing")
	}

	return &reporter{
		name: name,
	}, nil
}

type reporter struct {
	name string
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
