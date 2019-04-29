package structured

import (
	"io"

	dataSource "github.com/tidepool-org/platform/data/source"
)

type Store interface {
	NewSession() Session
}

type Session interface {
	io.Closer
	dataSource.Accessor
}
