package test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/test"
)

var database string
var address string

var cleanupHandlers []func()
var cleanupHandlersMu sync.Mutex

var _ = ginkgo.BeforeSuite(func() {
	database = generateUniqueName("database")
	address = os.Getenv("TIDEPOOL_STORE_ADDRESSES")
})

var _ = ginkgo.AfterSuite(func() {
	cfg := NewConfig()
	clientOptions := options.Client().ApplyURI(cfg.AsConnectionString())
	client, err := mongo.Connect(context.Background(), clientOptions)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	err = client.Database(database).Drop(context.Background())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	runDeferredCleanups()
})

// DeferredCleanup provides a way to have something run during AfterSuite.
//
// Ginkgo only allows a single AfterSuite to be instantiated, and it's here in this file, so
// give other tests a chance to make use of it.
func DeferredCleanup(f func()) {
	cleanupHandlersMu.Lock()
	defer cleanupHandlersMu.Unlock()
	if cleanupHandlers == nil {
		cleanupHandlers = []func(){}
	}
	cleanupHandlers = append(cleanupHandlers, f)
}

func runDeferredCleanups() {
	cleanupHandlersMu.Lock()
	defer cleanupHandlersMu.Unlock()
	for _, handler := range cleanupHandlers {
		if handler != nil {
			handler()
		}
	}
}

func Address() string {
	return address
}

func Database() string {
	return database
}

func NewCollectionPrefix() string {
	return generateUniqueName("collection_")
}

func generateUniqueName(base string) string {
	return fmt.Sprintf("test_%s_%s_%s", time.Now().Format("20060102150405"), test.RandomStringFromRangeAndCharset(4, 4, test.CharsetNumeric), base)
}

// MongoIndex models the output of the mongo driver Index().List() function
type MongoIndex struct {
	Key                     bson.D
	Name                    string
	Background              bool
	Unique                  bool
	Sparse                  bool
	PartialFilterExpression bson.D
}

// MakeKeySlice is a convenience function to convert an `mgo`-style key list into a bson.D
// This is important, because a bson.D is ordered, whereas a bson.M is not.
func MakeKeySlice(keyList ...string) bson.D {
	keySlice := bson.D{}
	for _, key := range keyList {
		order := int32(1)
		if key[0] == '-' {
			order = int32(-1)
			key = key[1:]
		}
		keySlice = append(keySlice, bson.E{Key: key, Value: order})
	}
	return keySlice
}
