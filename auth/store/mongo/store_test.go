package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/auth/store/mongo"
	logNull "github.com/tidepool-org/platform/log/null"
	storeMongo "github.com/tidepool-org/platform/store/mongo"
	testMongo "github.com/tidepool-org/platform/test/mongo"
)

var _ = Describe("Mongo", func() {
	var cfg *storeMongo.Config
	var str *mongo.Store

	BeforeEach(func() {
		cfg = testMongo.NewConfig()
	})

	AfterEach(func() {
		if str != nil {
			str.Close()
		}
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			str, err = mongo.NewStore(nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(str).To(BeNil())
		})

		It("returns successfully", func() {
			var err error
			str, err = mongo.NewStore(cfg, logNull.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			str, err = mongo.NewStore(cfg, logNull.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})

		// TODO: EnsureIndexes

		Context("NewProviderSessionSession", func() {
			var ssn store.ProviderSessionSession

			AfterEach(func() {
				if ssn != nil {
					ssn.Close()
				}
			})

			It("returns successfully", func() {
				ssn = str.NewProviderSessionSession()
				Expect(ssn).ToNot(BeNil())
			})
		})

		Context("NewRestrictedTokenSession", func() {
			var ssn store.RestrictedTokenSession

			AfterEach(func() {
				if ssn != nil {
					ssn.Close()
				}
			})

			It("returns successfully", func() {
				ssn = str.NewRestrictedTokenSession()
				Expect(ssn).ToNot(BeNil())
			})
		})
	})
})
