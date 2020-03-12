package store

import (
	"io"

	"github.com/tidepool-org/platform/prescription"
)

type Store interface {
	NewPrescriptionSession() PrescriptionSession
}

type PrescriptionSession interface {
	io.Closer
	prescription.Accessor
}
