package mongo_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/fx/fxtest"

	"github.com/tidepool-org/platform/config/test"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/prescription/store"
	"github.com/tidepool-org/platform/prescription/store/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Store", func() {
	var mongoConfig *storeStructuredMongo.Config
	var str *mongo.PrescriptionStore
	var configReporter *test.Reporter

	BeforeEach(func() {
		mongoConfig = storeStructuredMongoTest.NewConfig()
		prescriptionStoreConfig := map[string]interface{}{
			"addresses":        strings.Join(mongoConfig.Addresses, ","),
			"collectionPrefix": mongoConfig.CollectionPrefix,
			"database":         mongoConfig.Database,
			"tls":              fmt.Sprintf("%v", mongoConfig.TLS),
			"timeout":          fmt.Sprintf("%v", int(mongoConfig.Timeout.Seconds())),
		}
		serviceConfig := map[string]interface{}{
			"prescription": map[string]interface{}{
				"store": prescriptionStoreConfig,
			},
		}

		configReporter = test.NewReporter()
		configReporter.Config = serviceConfig
	})

	AfterEach(func() {
		if str != nil && str.Store != nil {
			str.Store.Close()
		}
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			prescrStr, err := mongo.NewStore(mongo.Params{
				ConfigReporter: nil,
				Logger:         nil,
			})

			Expect(err).To(HaveOccurred())
			Expect(prescrStr).To(BeNil())
		})

		It("returns successfully", func() {
			prescrStr, err := mongo.NewStore(mongo.Params{
				ConfigReporter: configReporter,
				Logger:         logNull.NewLogger(),
				Lifestyle:      fxtest.NewLifecycle(GinkgoT()),
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(prescrStr).ToNot(BeNil())

			str = prescrStr.(*mongo.PrescriptionStore)
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			prescrStr, err := mongo.NewStore(mongo.Params{
				ConfigReporter: configReporter,
				Logger:         logNull.NewLogger(),
				Lifestyle:      fxtest.NewLifecycle(GinkgoT()),
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(prescrStr).ToNot(BeNil())

			str = prescrStr.(*mongo.PrescriptionStore)
		})

		Context("With initialized store", func() {
			BeforeEach(func() {
				err := str.Initialize()
				Expect(err).ToNot(HaveOccurred())
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
})
