package types_test

import (
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

func TestTypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "summary/types")
}

var suiteStore *dataStoreMongo.Store
var suiteStoreOnce sync.Once

func GetSuiteStore() *dataStoreMongo.Store {
	GinkgoHelper()
	suiteStoreOnce.Do(func() {
		base := storeStructuredMongoTest.GetSuiteStore()
		suiteStore = dataStoreMongo.NewStoreFromBase(base)
		Expect(suiteStore.EnsureIndexes()).To(Succeed())
	})
	return suiteStore
}
