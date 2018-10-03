package store

import (
	"io"

	"github.com/tidepool-org/platform/data"
)

type Store interface {
	NewDataSourceSession() DataSourceSession
}

type DataSourceSession interface {
	io.Closer
	data.DataSourceAccessor
}
