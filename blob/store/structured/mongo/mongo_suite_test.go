package mongo_test

import (
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	blobStoreStructuredMongo "github.com/tidepool-org/platform/blob/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
)

func TestSuite(t *testing.T) {
	test.Test(t)
}

var suiteStore *blobStoreStructuredMongo.Store
var suiteStoreOnce sync.Once

func GetSuiteStore() *blobStoreStructuredMongo.Store {
	GinkgoHelper()
	suiteStoreOnce.Do(func() {
		base := storeStructuredMongoTest.GetSuiteStore()
		suiteStore = blobStoreStructuredMongo.NewStoreFromBase(base)
		Expect(suiteStore.EnsureIndexes()).To(Succeed())
	})
	return suiteStore
}
