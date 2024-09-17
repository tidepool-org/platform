package mongo_test

import (
	"context"

	"go.uber.org/fx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logNull "github.com/tidepool-org/platform/log/null"
	prescriptionStore "github.com/tidepool-org/platform/prescription/store"
	prescriptionStoreMongo "github.com/tidepool-org/platform/prescription/store/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

var _ = Describe("Store", Label("mongodb", "slow", "integration"), func() {
	var mongoConfig *storeStructuredMongo.Config
	var store *prescriptionStoreMongo.PrescriptionStore

	BeforeEach(func() {
		mongoConfig = storeStructuredMongoTest.NewConfig()
	})

	AfterEach(func() {
		if store != nil && store.Store != nil {
			store.Store.Terminate(context.Background())
		}
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
				fx.Supply(mongoConfig),
				fx.Provide(logNull.NewLogger),
				fx.Provide(storeStructuredMongo.NewStore),
				fx.Provide(prescriptionStoreMongo.NewStore),
				fx.Invoke(func(str prescriptionStore.Store) {
					store = str.(*prescriptionStoreMongo.PrescriptionStore)
				}),
			).Start(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			err := fx.New(
				fx.NopLogger,
				fx.Supply(mongoConfig),
				fx.Provide(logNull.NewLogger),
				fx.Provide(storeStructuredMongo.NewStore),
				fx.Provide(prescriptionStoreMongo.NewStore),
				fx.Invoke(func(str prescriptionStore.Store) {
					store = str.(*prescriptionStoreMongo.PrescriptionStore)
				}),
			).Start(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})

		Context("With initialized store", func() {
			BeforeEach(func() {
				err := store.CreateIndexes(context.Background())
				Expect(err).ToNot(HaveOccurred())
			})

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
