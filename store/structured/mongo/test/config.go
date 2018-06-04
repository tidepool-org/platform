package test

import (
	"time"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func NewConfig() *storeStructuredMongo.Config {
	return &storeStructuredMongo.Config{
		Addresses:        []string{Address()},
		Database:         Database(),
		CollectionPrefix: NewCollectionPrefix(),
		Timeout:          5 * time.Second,
	}
}
