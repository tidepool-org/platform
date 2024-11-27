package mongo_test

import (
	"context"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logtest "github.com/tidepool-org/platform/log/test"
	prescriptionStoreMongo "github.com/tidepool-org/platform/prescription/store/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
)

func TestSuite(t *testing.T) {
	test.Test(t)
}

var suiteStore *prescriptionStoreMongo.PrescriptionStore
var suiteStoreOnce sync.Once

func GetSuiteStore() *prescriptionStoreMongo.PrescriptionStore {
	GinkgoHelper()
	suiteStoreOnce.Do(func() {
		base := storeStructuredMongoTest.GetSuiteStore()
		suiteStore = prescriptionStoreMongo.NewStoreFromBase(base, logtest.NewLogger())
		Expect(suiteStore.CreateIndexes(context.Background())).To(Succeed())
	})
	return suiteStore
}
