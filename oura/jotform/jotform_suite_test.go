package jotform_test

import (
	"sync"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/oura/jotform/store"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
)

func TestSuite(t *testing.T) {
	test.Test(t)
}

var suiteStore store.Store
var suiteStoreOnce sync.Once

func GetSuiteStore() store.Store {
	GinkgoHelper()
	suiteStoreOnce.Do(func() {
		var err error
		base := storeStructuredMongoTest.GetSuiteStore()
		suiteStore, err = store.NewStore(base)
		Expect(err).ToNot(HaveOccurred())
	})
	return suiteStore
}
