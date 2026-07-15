package structured

import (
	dataSource "github.com/tidepool-org/platform/data/source"
)

//go:generate mockgen -source=structured.go -destination=test/structured_mocks.go -package=test -typed

type Store interface {
	NewDataSourcesRepository() DataSourcesRepository
}

type DataSourcesRepository interface {
	dataSource.Client
}
