package deduplicator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/data/deduplicator"
	testDataDeduplicator "github.com/tidepool-org/platform/data/deduplicator/test"
	testDataStore "github.com/tidepool-org/platform/data/store/test"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
)

var _ = Describe("Delegate", func() {
	Context("NewFactory", func() {
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

	Context("with a new factory", func() {
		var testFirstFactory *testDataDeduplicator.Factory
		var testSecondFactory *testDataDeduplicator.Factory
		var testDelegateFactory deduplicator.Factory
		var testDataset *upload.Upload

		BeforeEach(func() {
			var err error
			testFirstFactory = testDataDeduplicator.NewFactory()
			testFirstFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: false, Error: nil}}
			testSecondFactory = testDataDeduplicator.NewFactory()
			testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: false, Error: nil}}
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

		Context("NewDeduplicator", func() {
			var testLogger log.Logger
			var testDataStoreSession *testDataStore.Session

			BeforeEach(func() {
				testLogger = log.NewNull()
				testDataStoreSession = testDataStore.NewSession()
			})

			AfterEach(func() {
				Expect(testDataStoreSession.UnusedOutputsCount()).To(Equal(0))
			})

			It("returns an error if the logger is missing", func() {
				testFirstFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
				testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
				deduplicator, err := testDelegateFactory.NewDeduplicator(nil, testDataStoreSession, testDataset)
				Expect(err).To(MatchError("deduplicator: logger is missing"))
				Expect(deduplicator).To(BeNil())
			})

			It("returns an error if the data store session is missing", func() {
				testFirstFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
				testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
				deduplicator, err := testDelegateFactory.NewDeduplicator(testLogger, nil, testDataset)
				Expect(err).To(MatchError("deduplicator: data store session is missing"))
				Expect(deduplicator).To(BeNil())
			})

			It("returns an error if the dataset is missing", func() {
				testFirstFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
				testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
				deduplicator, err := testDelegateFactory.NewDeduplicator(testLogger, testDataStoreSession, nil)
				Expect(err).To(MatchError("deduplicator: dataset is missing"))
				Expect(deduplicator).To(BeNil())
			})

			It("returns an error if any contained factory returns an error", func() {
				testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: false, Error: errors.New("test error")}}
				deduplicator, err := testDelegateFactory.NewDeduplicator(testLogger, testDataStoreSession, testDataset)
				Expect(err).To(MatchError("test error"))
				Expect(deduplicator).To(BeNil())
				Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
				Expect(testSecondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
			})

			It("returns an error if no factory can deduplicate the dataset", func() {
				deduplicator, err := testDelegateFactory.NewDeduplicator(testLogger, testDataStoreSession, testDataset)
				Expect(err).To(MatchError("deduplicator: deduplicator not found"))
				Expect(deduplicator).To(BeNil())
				Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
				Expect(testSecondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
			})

			It("returns a deduplicator if any contained factory can deduplicate the dataset", func() {
				secondDeduplicator := testData.NewDeduplicator()
				testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: true, Error: nil}}
				testSecondFactory.NewDeduplicatorOutputs = []testDataDeduplicator.NewDeduplicatorOutput{{Deduplicator: secondDeduplicator, Error: nil}}
				Expect(testDelegateFactory.NewDeduplicator(testLogger, testDataStoreSession, testDataset)).To(Equal(secondDeduplicator))
				Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
				Expect(testSecondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
			})

			It("returns a deduplicator if any contained factory can deduplicate the dataset even if a later factory returns an error", func() {
				firstDeduplicator := testData.NewDeduplicator()
				testFirstFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: true, Error: nil}}
				testFirstFactory.NewDeduplicatorOutputs = []testDataDeduplicator.NewDeduplicatorOutput{{Deduplicator: firstDeduplicator, Error: nil}}
				testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{}
				Expect(testDelegateFactory.NewDeduplicator(testLogger, testDataStoreSession, testDataset)).To(Equal(firstDeduplicator))
				Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
				Expect(testFirstFactory.NewDeduplicatorInputs).To(ConsistOf(testDataDeduplicator.NewDeduplicatorInput{Logger: testLogger, DataStoreSession: testDataStoreSession, Dataset: testDataset}))
			})

			It("returns an error if any contained factory can deduplicate the dataset, but returns an error when creating", func() {
				testSecondFactory.CanDeduplicateDatasetOutputs = []testDataDeduplicator.CanDeduplicateDatasetOutput{{Can: true, Error: nil}}
				testSecondFactory.NewDeduplicatorOutputs = []testDataDeduplicator.NewDeduplicatorOutput{{Deduplicator: nil, Error: errors.New("test error")}}
				deduplicator, err := testDelegateFactory.NewDeduplicator(testLogger, testDataStoreSession, testDataset)
				Expect(err).To(MatchError("test error"))
				Expect(deduplicator).To(BeNil())
				Expect(testFirstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
				Expect(testSecondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(testDataset))
				Expect(testSecondFactory.NewDeduplicatorInputs).To(ConsistOf(testDataDeduplicator.NewDeduplicatorInput{Logger: testLogger, DataStoreSession: testDataStoreSession, Dataset: testDataset}))
			})
		})
	})
})
