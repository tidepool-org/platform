package store_test

import (
	. "github.com/tidepool-org/platform/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {

	Context("Can be created", func() {
		It("and should match the interface", func() {
			var testStore Store
			testStore = NewMongoStore()
			Expect(testStore).To(Not(BeNil()))
		})
	})
})
