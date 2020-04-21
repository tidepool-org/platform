package status

import (
	"github.com/tidepool-org/platform/prescription/store/mongo"
	"github.com/tidepool-org/platform/version"
)

type Reporter interface {
	Status() *Status
}

type defaultReporter struct {
	versionReporter   version.Reporter
	prescriptionStore *mongo.Store
}

func NewReporter(versionReporter version.Reporter, prescriptionStore *mongo.Store) Reporter {
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
