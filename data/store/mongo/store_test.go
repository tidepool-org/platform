package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/store/mongo"
	nullLog "github.com/tidepool-org/platform/log/null"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Store", func() {
	var cfg *storeStructuredMongo.Config
	var str *mongo.Store
	var ssn store.DataSourceSession

	BeforeEach(func() {
		cfg = storeStructuredMongoTest.NewConfig()
	})

	AfterEach(func() {
		if ssn != nil {
			ssn.Close()
		}
		if str != nil {
			str.Close()
		}
	})

	Context("NewStore", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			str, err = mongo.NewStore(nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(str).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			str, err = mongo.NewStore(cfg, nullLog.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			str, err = mongo.NewStore(cfg, nullLog.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})

		Context("NewDataSourceSession", func() {
			It("returns a new session", func() {
				ssn = str.NewDataSourceSession()
				Expect(ssn).ToNot(BeNil())
			})
		})
	})
})
