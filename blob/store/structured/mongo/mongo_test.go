package mongo_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	blobStoreStructuredMongo "github.com/tidepool-org/platform/blob/store/structured/mongo"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Mongo", func() {
	var config *storeStructuredMongo.Config
	var store *blobStoreStructuredMongo.Store

	BeforeEach(func() {
		config = storeStructuredMongoTest.NewConfig()
	})

	AfterEach(func() {
		if store != nil {
			store.Terminate(context.Background())
		}
	})

	Context("NewStore", func() {
		It("returns an error when unsuccessful", func() {
			var err error
			store, err = blobStoreStructuredMongo.NewStore(nil)
			errorsTest.ExpectEqual(err, errors.New("database config is empty"))
			Expect(store).To(BeNil())
		})

		It("returns a new store and no error when successful", func() {
			var err error
			store, err = blobStoreStructuredMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

})
