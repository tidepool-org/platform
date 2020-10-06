package mongo_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/notification/store"
	"github.com/tidepool-org/platform/notification/store/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Mongo", func() {
	var cfg *storeStructuredMongo.Config
	var str *mongo.Store
	var repository store.NotificationsRepository

	BeforeEach(func() {
		cfg = storeStructuredMongoTest.NewConfig()
	})

	AfterEach(func() {
		if str != nil {
			str.Terminate(context.Background())
		}
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			str, err = mongo.NewStore(nil)
			Expect(err).To(HaveOccurred())
			Expect(str).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			str, err = mongo.NewStore(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			str, err = mongo.NewStore(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})

		Context("NewNotificationsRepository", func() {
			It("returns a new repository", func() {
				repository = str.NewNotificationsRepository()
				Expect(repository).ToNot(BeNil())
			})
		})
	})
})
