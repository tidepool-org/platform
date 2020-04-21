package status

import (
	"github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/version"
	"go.uber.org/fx"
)

type Reporter interface {
	Status() *Status
}

var ReporterModule = fx.Provide(NewReporter)

type defaultReporter struct {
	versionReporter version.Reporter
	store           mongo.Store
}

func NewReporter(versionReporter version.Reporter, store mongo.Store) Reporter {
	return &defaultReporter{
		versionReporter: versionReporter,
		store:           store,
	}
}

func (r *defaultReporter) Status() *Status {
	return &Status{
		Version: r.versionReporter.Long(),
		Store:   r.Status(),
	}
}

type Status struct {
	Version string
	Store   interface{}
}
