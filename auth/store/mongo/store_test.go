package mongo_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/auth/store/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Store", func() {
	var params storeStructuredMongo.Params
	var str *mongo.Store

	BeforeEach(func() {
		params = storeStructuredMongo.Params{
			DatabaseConfig: storeStructuredMongoTest.NewConfig(),
		}
	})

	AfterEach(func() {
		if str != nil {
			str.Terminate(context.Background())
		}
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			str, err = mongo.NewStore(storeStructuredMongo.Params{})
			Expect(err).To(HaveOccurred())
			Expect(str).To(BeNil())
		})

		It("returns successfully", func() {
			var err error
			str, err = mongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			str, err = mongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})

		// TODO: EnsureIndexes

		Context("NewProviderSessionRepository", func() {
			var repository store.ProviderSessionRepository

			It("returns successfully", func() {
				repository = str.NewProviderSessionRepository()
				Expect(repository).ToNot(BeNil())
			})
		})

		Context("NewRestrictedTokenRepository", func() {
			var repository store.RestrictedTokenRepository

			It("returns successfully", func() {
				repository = str.NewRestrictedTokenRepository()
				Expect(repository).ToNot(BeNil())
			})
		})
	})
})
