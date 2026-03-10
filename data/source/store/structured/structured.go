package structured

import (
	dataSource "github.com/tidepool-org/platform/data/source"
)

//go:generate mockgen -source=structured.go -destination=test/structured_mocks.go -package=test Store
type Store interface {
	NewDataSourcesRepository() DataSourcesRepository
}

//go:generate mockgen -source=structured.go -destination=test/structured_mocks.go -package=test DataSourcesRepository
type DataSourcesRepository interface {
	dataSource.Client
}
