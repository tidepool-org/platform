package deduplicator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testDataDeduplicator "github.com/tidepool-org/platform/data/deduplicator/test"
	testDataStore "github.com/tidepool-org/platform/data/store/test"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
)

var _ = Describe("Delegate", func() {
	Context("NewDelegateFactory", func() {
		It("returns an error if factories is nil", func() {
			testFactory, err := deduplicator.NewDelegateFactory(nil)
			Expect(err).To(MatchError("deduplicator: factories is missing"))
			Expect(testFactory).To(BeNil())
		})

		It("returns an error if there are no factories", func() {
			testFactory, err := deduplicator.NewDelegateFactory([]deduplicator.Factory{})
			Expect(err).To(MatchError("deduplicator: factories is missing"))
			Expect(testFactory).To(BeNil())
		})

		It("returns success with one factory", func() {
			Expect(deduplicator.NewDelegateFactory([]deduplicator.Factory{testDataDeduplicator.NewFactory()})).ToNot(BeNil())
		})

		It("returns success with multiple factories", func() {
			Expect(deduplicator.NewDelegateFactory([]deduplicator.Factory{testDataDeduplicator.NewFactory(), testDataDeduplicator.NewFactory(), testDataDeduplicator.NewFactory(), testDataDeduplicator.NewFactory()})).ToNot(BeNil())
		})
	})

	Context("with a new delegate factory", func() {
		var testFirstFactory *testDataDeduplicator.Factory
		var testSecondFactory *testDataDeduplicator.Factory
		var testDelegateFactory deduplicator.Factory
		var testDataset *upload.Upload

		BeforeEach(func() {
			var err error
			testFirstFactory = testDataDeduplicator.NewFactory()
			testSecondFactory = testDataDeduplicator.NewFactory()
			testDelegateFactory, err = deduplicator.NewDelegateFactory([]deduplicator.Factory{testFirstFactory, testSecondFactory})
			Expect(err).ToNot(HaveOccurred())
			Expect(testDelegateFactory).ToNot(BeNil())
			testDataset = upload.Init()
			Expect(testDataset).ToNot(BeNil())
		})

		AfterEach(func() {
			Expect(testSecondFactory.UnusedOutputsCount()).To(Equal(0))
			Expect(testFirstFactory.UnusedOutputsCount()).To(Equal(0))
		})

		Context("with unregistered dataset", func() {
			BeforeEach(func() {
				testFirstFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: false, Error: nil}}
				testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: false, Error: nil}}
			})

			Context("CanDeduplicateDataset", func() {
				It("returns an error if the dataset is missing", func() {
					testFirstFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
					testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
					can, err := testDelegateFactory.CanDeduplicateDataset(nil)
					Expect(err).To(MatchError("deduplicator: dataset is missing"))
					Expect(can).To(BeFalse())
				})

				It("returns an error if any contained factory returns an error", func() {
					testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: false, Error: errors.New("test error")}}
					can, err := testDelegateFactory.CanDeduplicateDataset(testDataset)
					Expect(err).To(MatchError("test error"))
					Expect(can).To(BeFalse())
					Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
				})

				It("return false if no factory can deduplicate the dataset", func() {
					Expect(testDelegateFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
					Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
				})

				It("returns true if any contained factory can deduplicate the dataset", func() {
					testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: true, Error: nil}}
					Expect(testDelegateFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
					Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
				})

				It("returns true if any contained factory can deduplicate the dataset even if a later factory returns an error", func() {
					testFirstFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: true, Error: nil}}
					testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
					Expect(testDelegateFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
					Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.CanDeduplicateDatasetInputs).To(BeEmpty())
				})
			})

			Context("NewDeduplicatorForDataset", func() {
				var testLogger log.Logger
				var testDataStoreSession *testDataStore.Session

				BeforeEach(func() {
					testLogger = null.NewLogger()
					Expect(testLogger).ToNot(BeNil())
					testDataStoreSession = testDataStore.NewSession()
					Expect(testDataStoreSession).ToNot(BeNil())
				})

				AfterEach(func() {
					Expect(testDataStoreSession.UnusedOutputsCount()).To(Equal(0))
				})

				It("returns an error if the logger is missing", func() {
					testFirstFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
					testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
					testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataset(nil, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("deduplicator: logger is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data store session is missing", func() {
					testFirstFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
					testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
					testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataset(testLogger, nil, testDataset)
					Expect(err).To(MatchError("deduplicator: data store session is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset is missing", func() {
					testFirstFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
					testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
					testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, nil)
					Expect(err).To(MatchError("deduplicator: dataset is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if any contained factory returns an error", func() {
					testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: false, Error: errors.New("test error")}}
					testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("test error"))
					Expect(testDeduplicator).To(BeNil())
					Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
				})

				It("returns an error if no factory can deduplicate the dataset", func() {
					testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("deduplicator: deduplicator not found"))
					Expect(testDeduplicator).To(BeNil())
					Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
				})

				It("returns a deduplicator if any contained factory can deduplicate the dataset", func() {
					secondDeduplicator := testData.NewDeduplicator()
					testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: true, Error: nil}}
					testSecondFactory.NewDeduplicatorForDatasetOutputs = []testDataDeduplicator.NewDeduplicatorForDatasetOutput{{Deduplicator: secondDeduplicator, Error: nil}}
					Expect(testDelegateFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)).To(Equal(secondDeduplicator))
					Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
				})

				It("returns a deduplicator if any contained factory can deduplicate the dataset even if a later factory returns an error", func() {
					firstDeduplicator := testData.NewDeduplicator()
					testFirstFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: true, Error: nil}}
					testFirstFactory.NewDeduplicatorForDatasetOutputs = []testDataDeduplicator.NewDeduplicatorForDatasetOutput{{Deduplicator: firstDeduplicator, Error: nil}}
					testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
					Expect(testDelegateFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)).To(Equal(firstDeduplicator))
					Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
					Expect(testFirstFactory.NewDeduplicatorForDatasetInputs).To(ConsistOf(testDataDeduplicator.NewDeduplicatorForDatasetInput{Logger: testLogger, DataStoreSession: testDataStoreSession, Dataset: testDataset}))
				})

				It("returns an error if any contained factory can deduplicate the dataset, but returns an error when creating", func() {
					testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: true, Error: nil}}
					testSecondFactory.NewDeduplicatorForDatasetOutputs = []testDataDeduplicator.NewDeduplicatorForDatasetOutput{{Deduplicator: nil, Error: errors.New("test error")}}
					testDeduplicator, err := testDelegateFactory.NewDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("test error"))
					Expect(testDeduplicator).To(BeNil())
					Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.NewDeduplicatorForDatasetInputs).To(ConsistOf(testDataDeduplicator.NewDeduplicatorForDatasetInput{Logger: testLogger, DataStoreSession: testDataStoreSession, Dataset: testDataset}))
				})
			})
		})

		Context("with registered dataset", func() {
			BeforeEach(func() {
				testFirstFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{{Is: false, Error: nil}}
				testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{{Is: false, Error: nil}}
				testDataset.SetDeduplicatorDescriptor(&data.DeduplicatorDescriptor{Name: "test"})
			})

			Context("IsRegisteredWithDataset", func() {
				It("returns an error if the dataset is missing", func() {
					testFirstFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					can, err := testDelegateFactory.IsRegisteredWithDataset(nil)
					Expect(err).To(MatchError("deduplicator: dataset is missing"))
					Expect(can).To(BeFalse())
				})

				It("returns an error if any contained factory returns an error", func() {
					testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{{Is: false, Error: errors.New("test error")}}
					can, err := testDelegateFactory.IsRegisteredWithDataset(testDataset)
					Expect(err).To(MatchError("test error"))
					Expect(can).To(BeFalse())
					Expect(testFirstFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
				})

				It("return false if no factory is registered with the dataset", func() {
					Expect(testDelegateFactory.IsRegisteredWithDataset(testDataset)).To(BeFalse())
					Expect(testFirstFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
				})

				It("returns true if any contained factory is registered with the dataset", func() {
					testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{{Is: true, Error: nil}}
					Expect(testDelegateFactory.IsRegisteredWithDataset(testDataset)).To(BeTrue())
					Expect(testFirstFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
				})

				It("returns true if any contained factory is registered with the dataset even if a later factory returns an error", func() {
					testFirstFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{{Is: true, Error: nil}}
					testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					Expect(testDelegateFactory.IsRegisteredWithDataset(testDataset)).To(BeTrue())
					Expect(testFirstFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.IsRegisteredWithDatasetInputs).To(BeEmpty())
				})
			})

			Context("NewRegisteredDeduplicatorForDataset", func() {
				var testLogger log.Logger
				var testDataStoreSession *testDataStore.Session

				BeforeEach(func() {
					testLogger = null.NewLogger()
					testDataStoreSession = testDataStore.NewSession()
					Expect(testDataStoreSession).ToNot(BeNil())
				})

				AfterEach(func() {
					Expect(testDataStoreSession.UnusedOutputsCount()).To(Equal(0))
				})

				It("returns an error if the logger is missing", func() {
					testFirstFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataset(nil, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("deduplicator: logger is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data store session is missing", func() {
					testFirstFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataset(testLogger, nil, testDataset)
					Expect(err).To(MatchError("deduplicator: data store session is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset is missing", func() {
					testFirstFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, nil)
					Expect(err).To(MatchError("deduplicator: dataset is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset is not registered with a deduplicator", func() {
					testFirstFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					testDataset.SetDeduplicatorDescriptor(nil)
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset not registered with deduplicator"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset is not registered with a deduplicator with a name", func() {
					testFirstFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					testDataset.SetDeduplicatorDescriptor(&data.DeduplicatorDescriptor{})
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset not registered with deduplicator"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if any contained factory returns an error", func() {
					testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{{Is: false, Error: errors.New("test error")}}
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("test error"))
					Expect(testDeduplicator).To(BeNil())
					Expect(testFirstFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
				})

				It("returns an error if no factory is registered with the dataset", func() {
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("deduplicator: deduplicator not found"))
					Expect(testDeduplicator).To(BeNil())
					Expect(testFirstFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
				})

				It("returns a deduplicator if any contained factory is registered with the dataset", func() {
					secondDeduplicator := testData.NewDeduplicator()
					testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{{Is: true, Error: nil}}
					testSecondFactory.NewRegisteredDeduplicatorForDatasetOutputs = []testDataDeduplicator.NewRegisteredDeduplicatorForDatasetOutput{{Deduplicator: secondDeduplicator, Error: nil}}
					Expect(testDelegateFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)).To(Equal(secondDeduplicator))
					Expect(testFirstFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
				})

				It("returns a deduplicator if any contained factory is registered with the dataset even if a later factory returns an error", func() {
					firstDeduplicator := testData.NewDeduplicator()
					testFirstFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{{Is: true, Error: nil}}
					testFirstFactory.NewRegisteredDeduplicatorForDatasetOutputs = []testDataDeduplicator.NewRegisteredDeduplicatorForDatasetOutput{{Deduplicator: firstDeduplicator, Error: nil}}
					testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
					Expect(testDelegateFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)).To(Equal(firstDeduplicator))
					Expect(testFirstFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
					Expect(testFirstFactory.NewRegisteredDeduplicatorForDatasetInputs).To(ConsistOf(testDataDeduplicator.NewRegisteredDeduplicatorForDatasetInput{Logger: testLogger, DataStoreSession: testDataStoreSession, Dataset: testDataset}))
				})

				It("returns an error if any contained factory is registered with the dataset, but returns an error when creating", func() {
					testSecondFactory.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{{Is: true, Error: nil}}
					testSecondFactory.NewRegisteredDeduplicatorForDatasetOutputs = []testDataDeduplicator.NewRegisteredDeduplicatorForDatasetOutput{{Deduplicator: nil, Error: errors.New("test error")}}
					testDeduplicator, err := testDelegateFactory.NewRegisteredDeduplicatorForDataset(testLogger, testDataStoreSession, testDataset)
					Expect(err).To(MatchError("test error"))
					Expect(testDeduplicator).To(BeNil())
					Expect(testFirstFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.IsRegisteredWithDatasetInputs).To(ConsistOf(testDataset))
					Expect(testSecondFactory.NewRegisteredDeduplicatorForDatasetInputs).To(ConsistOf(testDataDeduplicator.NewRegisteredDeduplicatorForDatasetInput{Logger: testLogger, DataStoreSession: testDataStoreSession, Dataset: testDataset}))
				})
			})
		})
	})
})
