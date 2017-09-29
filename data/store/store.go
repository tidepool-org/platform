package store

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/store"
)

type Store interface {
	store.Store

	NewDataSourceSession() DataSourceSession
}

type DataSourceSession interface {
	store.Session
	data.DataSourceAccessor
}
