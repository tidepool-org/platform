package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	nullLog "github.com/tidepool-org/platform/log/null"
	storeMongo "github.com/tidepool-org/platform/store/mongo"
	"github.com/tidepool-org/platform/task/store"
	"github.com/tidepool-org/platform/task/store/mongo"
	testMongo "github.com/tidepool-org/platform/test/mongo"
)

var _ = Describe("Mongo", func() {
	var cfg *storeMongo.Config
	var str *mongo.Store
	var ssn store.StoreSession

	BeforeEach(func() {
		cfg = &storeMongo.Config{
			Addresses:  []string{testMongo.Address()},
			Database:   testMongo.Database(),
			Collection: testMongo.NewCollectionName(),
			Timeout:    5 * time.Second,
		}
	})

	AfterEach(func() {
		if ssn != nil {
			ssn.Close()
		}
		if str != nil {
			str.Close()
		}
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			str, err = mongo.New(nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(str).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			str, err = mongo.New(nullLog.NewLogger(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			str, err = mongo.New(nullLog.NewLogger(), cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeNil())
		})

		Context("NewSession", func() {
			It("returns a new session if no logger specified", func() {
				ssn = str.NewSession(nil)
				Expect(ssn).ToNot(BeNil())
				Expect(ssn.Logger()).ToNot(BeNil())
			})

			It("returns a new session if logger specified", func() {
				ssn = str.NewSession(nullLog.NewLogger())
				Expect(ssn).ToNot(BeNil())
				Expect(ssn.Logger()).ToNot(BeNil())
			})
		})
	})
})
