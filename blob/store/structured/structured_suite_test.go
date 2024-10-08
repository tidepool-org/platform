package structured_test

import (
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/store/mongo"
	mongotest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
)

func TestSuite(t *testing.T) {
	test.Test(t)
}

var suiteStore *mongo.Store
var suiteStoreOnce sync.Once

func GetSuiteStore() *mongo.Store {
	GinkgoHelper()
	suiteStoreOnce.Do(func() {
		base := mongotest.GetSuiteStore()
		suiteStore = mongo.NewStoreFromBase(base)
		Expect(suiteStore.EnsureIndexes()).To(Succeed())
	})
	return suiteStore
}
