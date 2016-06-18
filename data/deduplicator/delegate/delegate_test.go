package delegate_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/deduplicator/delegate"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/test"
)

type CanDeduplicateDatasetOutput struct {
	Bool  bool
	Error error
}

type NewDeduplicatorInput struct {
	Logger           log.Logger
	DataStoreSession store.Session
	Dataset          *upload.Upload
}

type NewDeduplicatorOutput struct {
	Deduplicator deduplicator.Deduplicator
	Error        error
}

type TestFactory struct {
	CanDeduplicateDatasetInputs  []*upload.Upload
	CanDeduplicateDatasetOutputs []CanDeduplicateDatasetOutput
	NewDeduplicatorInputs        []NewDeduplicatorInput
	NewDeduplicatorOutputs       []NewDeduplicatorOutput
}

func (t *TestFactory) CanDeduplicateDataset(dataset *upload.Upload) (bool, error) {
	t.CanDeduplicateDatasetInputs = append(t.CanDeduplicateDatasetInputs, dataset)
	output := t.CanDeduplicateDatasetOutputs[0]
	t.CanDeduplicateDatasetOutputs = t.CanDeduplicateDatasetOutputs[1:]
	return output.Bool, output.Error
}

func (t *TestFactory) NewDeduplicator(logger log.Logger, dataStoreSession store.Session, dataset *upload.Upload) (deduplicator.Deduplicator, error) {
	t.NewDeduplicatorInputs = append(t.NewDeduplicatorInputs, NewDeduplicatorInput{logger, dataStoreSession, dataset})
	output := t.NewDeduplicatorOutputs[0]
	t.NewDeduplicatorOutputs = t.NewDeduplicatorOutputs[1:]
	return output.Deduplicator, output.Error
}

type TestDeduplicator struct{}

func (t *TestDeduplicator) InitializeDataset() error {
	panic("unexpected")
}

func (t *TestDeduplicator) AddDataToDataset(datasetData []data.Datum) error {
	panic("unexpected")
}

func (t *TestDeduplicator) FinalizeDataset() error {
	panic("unexpected")
}

type TestDataStoreSession struct{}

func (t *TestDataStoreSession) IsClosed() bool {
	panic("unexpected")
}

func (t *TestDataStoreSession) Close() {
	panic("unexpected")
}

func (t *TestDataStoreSession) GetDataset(datasetID string) (*upload.Upload, error) {
	panic("unexpected")
}

func (t *TestDataStoreSession) CreateDataset(dataset *upload.Upload) error {
	panic("unexpected")
}

func (t *TestDataStoreSession) UpdateDataset(dataset *upload.Upload) error {
	panic("unexpected")
}

func (t *TestDataStoreSession) CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error {
	panic("unexpected")
}

func (t *TestDataStoreSession) ActivateAllDatasetData(dataset *upload.Upload) error {
	panic("unexpected")
}

func (t *TestDataStoreSession) RemoveAllOtherDatasetData(dataset *upload.Upload) error {
	panic("unexpected")
}

var _ = Describe("Delegate", func() {
	Context("NewFactory", func() {
		It("returns an error if factories is nil", func() {
			factory, err := delegate.NewFactory(nil)
			Expect(err).To(MatchError("delegate: factories is missing"))
			Expect(factory).To(BeNil())
		})

		It("returns an error if there are no factories", func() {
			factory, err := delegate.NewFactory([]deduplicator.Factory{})
			Expect(err).To(MatchError("delegate: factories is missing"))
			Expect(factory).To(BeNil())
		})

		It("returns success with one factory", func() {
			factory, err := delegate.NewFactory([]deduplicator.Factory{&TestFactory{}})
			Expect(err).ToNot(HaveOccurred())
			Expect(factory).ToNot(BeNil())
		})

		It("returns success with multiple factories", func() {
			factory, err := delegate.NewFactory([]deduplicator.Factory{&TestFactory{}, &TestFactory{}, &TestFactory{}, &TestFactory{}})
			Expect(err).ToNot(HaveOccurred())
			Expect(factory).ToNot(BeNil())
		})
	})

	Context("with a new factory", func() {
		var firstFactory *TestFactory
		var secondFactory *TestFactory
		var delegateFactory deduplicator.Factory
		var dataset *upload.Upload

		BeforeEach(func() {
			var err error
			firstFactory = &TestFactory{CanDeduplicateDatasetOutputs: []CanDeduplicateDatasetOutput{{false, nil}}}
			secondFactory = &TestFactory{CanDeduplicateDatasetOutputs: []CanDeduplicateDatasetOutput{{false, nil}}}
			delegateFactory, err = delegate.NewFactory([]deduplicator.Factory{firstFactory, secondFactory})
			Expect(err).ToNot(HaveOccurred())
			Expect(delegateFactory).ToNot(BeNil())
			dataset = upload.Init()
			Expect(dataset).ToNot(BeNil())
		})

		Context("CanDeduplicateDataset", func() {
			It("returns an error if the dataset is missing", func() {
				firstFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{}
				secondFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{}
				can, err := delegateFactory.CanDeduplicateDataset(nil)
				Expect(err).To(MatchError("delegate: dataset is missing"))
				Expect(can).To(BeFalse())
			})

			It("returns an error if any contained factory returns an error", func() {
				secondFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{{false, errors.New("test error")}}
				can, err := delegateFactory.CanDeduplicateDataset(dataset)
				Expect(err).To(MatchError("test error"))
				Expect(can).To(BeFalse())
				Expect(firstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(firstFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
				Expect(secondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(secondFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
			})

			It("return false if no factory can deduplicate the dataset", func() {
				can, err := delegateFactory.CanDeduplicateDataset(dataset)
				Expect(err).ToNot(HaveOccurred())
				Expect(can).To(BeFalse())
				Expect(firstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(firstFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
				Expect(secondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(secondFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
			})

			It("returns true if any contained factory can deduplicate the dataset", func() {
				secondFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{{true, nil}}
				can, err := delegateFactory.CanDeduplicateDataset(dataset)
				Expect(err).ToNot(HaveOccurred())
				Expect(can).To(BeTrue())
				Expect(firstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(firstFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
				Expect(secondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(secondFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
			})

			It("returns true if any contained factory can deduplicate the dataset even if a later factory returns an error", func() {
				firstFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{{true, nil}}
				secondFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{{false, errors.New("test error")}}
				can, err := delegateFactory.CanDeduplicateDataset(dataset)
				Expect(err).ToNot(HaveOccurred())
				Expect(can).To(BeTrue())
				Expect(firstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(firstFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
				Expect(secondFactory.CanDeduplicateDatasetInputs).To(BeEmpty())
				Expect(secondFactory.CanDeduplicateDatasetOutputs).To(ConsistOf(CanDeduplicateDatasetOutput{false, errors.New("test error")}))
			})
		})

		Context("NewDeduplicator", func() {
			var logger log.Logger
			var dataStoreSession store.Session

			BeforeEach(func() {
				logger = test.NewLogger()
				dataStoreSession = &TestDataStoreSession{}
			})

			It("returns an error if the logger is missing", func() {
				firstFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{}
				secondFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{}
				deduplicator, err := delegateFactory.NewDeduplicator(nil, dataStoreSession, dataset)
				Expect(err).To(MatchError("delegate: logger is missing"))
				Expect(deduplicator).To(BeNil())
			})

			It("returns an error if the data store session is missing", func() {
				firstFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{}
				secondFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{}
				deduplicator, err := delegateFactory.NewDeduplicator(logger, nil, dataset)
				Expect(err).To(MatchError("delegate: data store session is missing"))
				Expect(deduplicator).To(BeNil())
			})

			It("returns an error if the dataset is missing", func() {
				firstFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{}
				secondFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{}
				deduplicator, err := delegateFactory.NewDeduplicator(logger, dataStoreSession, nil)
				Expect(err).To(MatchError("delegate: dataset is missing"))
				Expect(deduplicator).To(BeNil())
			})

			It("returns an error if any contained factory returns an error", func() {
				secondFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{{false, errors.New("test error")}}
				deduplicator, err := delegateFactory.NewDeduplicator(logger, dataStoreSession, dataset)
				Expect(err).To(MatchError("test error"))
				Expect(deduplicator).To(BeNil())
				Expect(firstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(firstFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
				Expect(secondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(secondFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
			})

			It("returns an error if no factory can deduplicate the dataset", func() {
				deduplicator, err := delegateFactory.NewDeduplicator(logger, dataStoreSession, dataset)
				Expect(err).To(MatchError("delegate: deduplicator not found"))
				Expect(deduplicator).To(BeNil())
				Expect(firstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(firstFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
				Expect(secondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(secondFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
			})

			It("returns a deduplicator if any contained factory can deduplicate the dataset", func() {
				secondDeduplicator := &TestDeduplicator{}
				secondFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{{true, nil}}
				secondFactory.NewDeduplicatorOutputs = []NewDeduplicatorOutput{{secondDeduplicator, nil}}
				deduplicator, err := delegateFactory.NewDeduplicator(logger, dataStoreSession, dataset)
				Expect(err).ToNot(HaveOccurred())
				Expect(deduplicator).To(Equal(secondDeduplicator))
				Expect(firstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(firstFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
				Expect(secondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(secondFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
				Expect(secondFactory.NewDeduplicatorOutputs).To(BeEmpty())
			})

			It("returns a deduplicator if any contained factory can deduplicate the dataset even if a later factory returns an error", func() {
				firstDeduplicator := &TestDeduplicator{}
				firstFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{{true, nil}}
				firstFactory.NewDeduplicatorOutputs = []NewDeduplicatorOutput{{firstDeduplicator, nil}}
				secondFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{{false, errors.New("test error")}}
				deduplicator, err := delegateFactory.NewDeduplicator(logger, dataStoreSession, dataset)
				Expect(err).ToNot(HaveOccurred())
				Expect(deduplicator).To(Equal(firstDeduplicator))
				Expect(firstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(firstFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
				Expect(firstFactory.NewDeduplicatorInputs).To(ConsistOf(NewDeduplicatorInput{logger, dataStoreSession, dataset}))
				Expect(firstFactory.NewDeduplicatorOutputs).To(BeEmpty())
				Expect(secondFactory.CanDeduplicateDatasetOutputs).To(ConsistOf(CanDeduplicateDatasetOutput{false, errors.New("test error")}))
			})

			It("returns an error if any contained factory can deduplicate the dataset, but returns an error when creating", func() {
				secondFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{{true, nil}}
				secondFactory.NewDeduplicatorOutputs = []NewDeduplicatorOutput{{nil, errors.New("test error")}}
				deduplicator, err := delegateFactory.NewDeduplicator(logger, dataStoreSession, dataset)
				Expect(err).To(MatchError("test error"))
				Expect(deduplicator).To(BeNil())
				Expect(firstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(firstFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
				Expect(secondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(secondFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
				Expect(secondFactory.NewDeduplicatorInputs).To(ConsistOf(NewDeduplicatorInput{logger, dataStoreSession, dataset}))
				Expect(secondFactory.NewDeduplicatorOutputs).To(BeEmpty())
			})
		})
	})
})
