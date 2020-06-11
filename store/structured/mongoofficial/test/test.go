package test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/test"
)

var database string

var _ = ginkgo.BeforeSuite(func() {
	database = generateUniqueName("database")
})

var _ = ginkgo.AfterSuite(func() {
	cfg := NewConfig()
	clientOptions := options.Client().ApplyURI(cfg.AsConnectionString())
	client, err := mongo.Connect(context.Background(), clientOptions)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	err = client.Database(database).Drop(context.Background())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
})

func Address() string {
	return os.Getenv("TIDEPOOL_STORE_ADDRESSES")
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
