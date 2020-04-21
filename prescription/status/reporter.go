package status

import (
	"github.com/tidepool-org/platform/prescription/store"
	"github.com/tidepool-org/platform/version"
	"go.uber.org/fx"
)

type Reporter interface {
	Status() *Status
}

var Module = fx.Provide(NewReporter)

type defaultReporter struct {
	versionReporter   version.Reporter
	prescriptionStore store.Store
}

func NewReporter(versionReporter version.Reporter, prescriptionStore store.Store) Reporter {
	return &defaultReporter{
		versionReporter:   versionReporter,
		prescriptionStore: prescriptionStore,
	}
}

func (r *defaultReporter) Status() *Status {
	return &Status{
		Version: r.versionReporter.Long(),
		Store:   r.prescriptionStore.Status(),
	}
}

type Status struct {
	Version string
	Store   interface{}
}
