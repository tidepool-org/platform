package test

import (
	"time"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

//NewConfig creates a test Mongo configuration
func NewConfig() *storeStructuredMongo.Config {
	conf := &storeStructuredMongo.Config{
		Database:         Database(),
		CollectionPrefix: NewCollectionPrefix(),
		Timeout:          5 * time.Second,
	}
	conf.SetAddresses([]string{Address()})

	return conf
}
