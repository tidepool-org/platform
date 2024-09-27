package mongo_test

import (
	"context"
	"sync"

	"go.uber.org/fx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logNull "github.com/tidepool-org/platform/log/null"
	logtest "github.com/tidepool-org/platform/log/test"
	prescriptionStore "github.com/tidepool-org/platform/prescription/store"
	prescriptionStoreMongo "github.com/tidepool-org/platform/prescription/store/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Store", Label("mongodb", "slow", "integration"), func() {
	var store *prescriptionStoreMongo.PrescriptionStore

	BeforeEach(func() {
		store = GetSuiteStore()
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			prescrStr, err := prescriptionStoreMongo.NewStore(prescriptionStoreMongo.Params{
				Logger: nil,
			})

			Expect(err).To(HaveOccurred())
			Expect(prescrStr).To(BeNil())
		})

		It("returns successfully", func() {
			err := fx.New(
				fx.NopLogger,
				fx.Supply(store),
				fx.Provide(logNull.NewLogger),
			).Start(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			err := fx.New(
				fx.NopLogger,
				fx.Supply(store),
				fx.Provide(logNull.NewLogger),
			).Start(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})

		Context("With initialized store", func() {
			Context("GetPrescriptionRepository", func() {
				var repo prescriptionStore.PrescriptionRepository

				It("returns successfully", func() {
					repo = store.GetPrescriptionRepository()
					Expect(repo).ToNot(BeNil())
				})
			})
		})
	})
})

var suiteStore *prescriptionStoreMongo.PrescriptionStore
var suiteStoreOnce sync.Once

func GetSuiteStore() *prescriptionStoreMongo.PrescriptionStore {
	GinkgoHelper()
	suiteStoreOnce.Do(func() {
		base := storeStructuredMongoTest.GetSuiteStore()
		suiteStore = prescriptionStoreMongo.NewStoreFromBase(base, logtest.NewLogger())
		Expect(suiteStore.CreateIndexes(context.Background())).To(Succeed())
	})
	return suiteStore
}
