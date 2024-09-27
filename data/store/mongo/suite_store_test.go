package mongo_test

import (
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/store/structured/mongo/test"
)

var suiteStore *mongo.Store
var suiteStoreOnce sync.Once

func GetSuiteStore() *mongo.Store {
	GinkgoHelper()
	suiteStoreOnce.Do(func() {
		base := test.GetSuiteStore()
		suiteStore = mongo.NewStoreFromBase(base)
		Expect(suiteStore.EnsureIndexes()).To(Succeed())
	})
	return suiteStore
}
