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
	commonStore "github.com/tidepool-org/platform/store"
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
	panic("Unexpected invocation of InitializeDataset on TestDeduplicator")
}

func (t *TestDeduplicator) AddDataToDataset(datasetData []data.Datum) error {
	panic("Unexpected invocation of AddDataToDataset on TestDeduplicator")
}

func (t *TestDeduplicator) FinalizeDataset() error {
	panic("Unexpected invocation of FinalizeDataset on TestDeduplicator")
}

type TestDataStoreSession struct{}

func (t *TestDataStoreSession) IsClosed() bool {
	panic("Unexpected invocation of IsClosed on TestDataStoreSession")
}

func (t *TestDataStoreSession) Close() {
	panic("Unexpected invocation of Close on TestDataStoreSession")
}

func (t *TestDataStoreSession) SetAgent(agent commonStore.Agent) {
	panic("Unexpected invocation of SetAgent on TestDataStoreSession")
}

func (t *TestDataStoreSession) GetDatasetsForUserByID(userID string, filter *store.Filter, pagination *store.Pagination) ([]*upload.Upload, error) {
	panic("Unexpected invocation of GetDatasetsForUserByID on TestDataStoreSession")
}

func (t *TestDataStoreSession) GetDatasetByID(datasetID string) (*upload.Upload, error) {
	panic("Unexpected invocation of GetDatasetByID on TestDataStoreSession")
}

func (t *TestDataStoreSession) CreateDataset(dataset *upload.Upload) error {
	panic("Unexpected invocation of CreateDataset on TestDataStoreSession")
}

func (t *TestDataStoreSession) UpdateDataset(dataset *upload.Upload) error {
	panic("Unexpected invocation of UpdateDataset on TestDataStoreSession")
}

func (t *TestDataStoreSession) DeleteDataset(dataset *upload.Upload) error {
	panic("Unexpected invocation of DeleteDataset on TestDataStoreSession")
}

func (t *TestDataStoreSession) CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error {
	panic("Unexpected invocation of CreateDatasetData on TestDataStoreSession")
}

func (t *TestDataStoreSession) ActivateDatasetData(dataset *upload.Upload) error {
	panic("Unexpected invocation of ActivateDatasetData on TestDataStoreSession")
}

func (t *TestDataStoreSession) DeleteOtherDatasetData(dataset *upload.Upload) error {
	panic("Unexpected invocation of DeleteOtherDatasetData on TestDataStoreSession")
}

func (t *TestDataStoreSession) DestroyDataForUserByID(userID string) error {
	panic("Unexpected invocation of DestroyDataForUserByID on TestDataStoreSession")
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
			Expect(delegate.NewFactory([]deduplicator.Factory{&TestFactory{}})).ToNot(BeNil())
		})

		It("returns success with multiple factories", func() {
			Expect(delegate.NewFactory([]deduplicator.Factory{&TestFactory{}, &TestFactory{}, &TestFactory{}, &TestFactory{}})).ToNot(BeNil())
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
				Expect(delegateFactory.CanDeduplicateDataset(dataset)).To(BeFalse())
				Expect(firstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(firstFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
				Expect(secondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(secondFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
			})

			It("returns true if any contained factory can deduplicate the dataset", func() {
				secondFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{{true, nil}}
				Expect(delegateFactory.CanDeduplicateDataset(dataset)).To(BeTrue())
				Expect(firstFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(firstFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
				Expect(secondFactory.CanDeduplicateDatasetInputs).To(ConsistOf(dataset))
				Expect(secondFactory.CanDeduplicateDatasetOutputs).To(BeEmpty())
			})

			It("returns true if any contained factory can deduplicate the dataset even if a later factory returns an error", func() {
				firstFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{{true, nil}}
				secondFactory.CanDeduplicateDatasetOutputs = []CanDeduplicateDatasetOutput{{false, errors.New("test error")}}
				Expect(delegateFactory.CanDeduplicateDataset(dataset)).To(BeTrue())
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
				logger = log.NewNull()
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
				Expect(delegateFactory.NewDeduplicator(logger, dataStoreSession, dataset)).To(Equal(secondDeduplicator))
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
				Expect(delegateFactory.NewDeduplicator(logger, dataStoreSession, dataset)).To(Equal(firstDeduplicator))
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
