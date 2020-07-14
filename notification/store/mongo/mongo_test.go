package mongo_test

import (
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
	var ssn store.NotificationsRepository

	BeforeEach(func() {
		cfg = storeStructuredMongoTest.NewConfig()
	})

	AfterEach(func() {
		if str != nil {
			str.Terminate(nil)
		}
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: nil}
			str, err = mongo.NewStore(params)
			Expect(err).To(HaveOccurred())
			Expect(str).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: cfg}
			str, err = mongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: cfg}
			str, err = mongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})

		Context("NewNotificationsRepository", func() {
			It("returns a new session", func() {
				ssn = str.NewNotificationsRepository()
				Expect(ssn).ToNot(BeNil())
			})
		})
	})
})
