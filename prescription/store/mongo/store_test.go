package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/prescription/store"
	"github.com/tidepool-org/platform/prescription/store/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Store", func() {
	var cfg *storeStructuredMongo.Config
	var str *mongo.Store

	BeforeEach(func() {
		cfg = storeStructuredMongoTest.NewConfig()
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

		Context("NewPrescriptionSession", func() {
			var ssn store.PrescriptionSession

			AfterEach(func() {
				if ssn != nil {
					ssn.Close()
				}
			})

			It("returns successfully", func() {
				ssn = str.NewPrescriptionSession()
				Expect(ssn).ToNot(BeNil())
			})
		})
	})
})
