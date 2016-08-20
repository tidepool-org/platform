package truncate_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/deduplicator/truncate"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
)

type CreateDatasetDataInput struct {
	dataset     *upload.Upload
	datasetData []data.Datum
}

type TestDataStoreSession struct {
	UpdateDatasetInputs              []*upload.Upload
	UpdateDatasetOutputs             []error
	CreateDatasetDataInputs          []CreateDatasetDataInput
	CreateDatasetDataOutputs         []error
	ActivateAllDatasetDataInputs     []*upload.Upload
	ActivateAllDatasetDataOutputs    []error
	DeleteAllOtherDatasetDataInputs  []*upload.Upload
	DeleteAllOtherDatasetDataOutputs []error
}

func (t *TestDataStoreSession) IsClosed() bool {
	panic("Unexpected invocation of IsClosed on TestDataStoreSession")
}

func (t *TestDataStoreSession) Close() {
	panic("Unexpected invocation of Close on TestDataStoreSession")
}

func (t *TestDataStoreSession) GetDataset(datasetID string) (*upload.Upload, error) {
	panic("Unexpected invocation of GetDataset on TestDataStoreSession")
}

func (t *TestDataStoreSession) CreateDataset(dataset *upload.Upload) error {
	panic("Unexpected invocation of CreateDataset on TestDataStoreSession")
}

func (t *TestDataStoreSession) UpdateDataset(dataset *upload.Upload) error {
	t.UpdateDatasetInputs = append(t.UpdateDatasetInputs, dataset)
	output := t.UpdateDatasetOutputs[0]
	t.UpdateDatasetOutputs = t.UpdateDatasetOutputs[1:]
	return output
}

func (t *TestDataStoreSession) CreateDatasetData(dataset *upload.Upload, datasetData []data.Datum) error {
	t.CreateDatasetDataInputs = append(t.CreateDatasetDataInputs, CreateDatasetDataInput{dataset, datasetData})
	output := t.CreateDatasetDataOutputs[0]
	t.CreateDatasetDataOutputs = t.CreateDatasetDataOutputs[1:]
	return output
}

func (t *TestDataStoreSession) ActivateAllDatasetData(dataset *upload.Upload) error {
	t.ActivateAllDatasetDataInputs = append(t.ActivateAllDatasetDataInputs, dataset)
	output := t.ActivateAllDatasetDataOutputs[0]
	t.ActivateAllDatasetDataOutputs = t.ActivateAllDatasetDataOutputs[1:]
	return output
}

func (t *TestDataStoreSession) DeleteAllOtherDatasetData(dataset *upload.Upload) error {
	t.DeleteAllOtherDatasetDataInputs = append(t.DeleteAllOtherDatasetDataInputs, dataset)
	output := t.DeleteAllOtherDatasetDataOutputs[0]
	t.DeleteAllOtherDatasetDataOutputs = t.DeleteAllOtherDatasetDataOutputs[1:]
	return output
}

func StringAsPointer(sourceString string) *string { return &sourceString }

var _ = Describe("Truncate", func() {
	Context("NewFactory", func() {
		It("returns a new factory", func() {
			Expect(truncate.NewFactory()).ToNot(BeNil())
		})
	})

	Context("with a new factory", func() {
		var factory deduplicator.Factory
		var dataset *upload.Upload

		BeforeEach(func() {
			var err error
			factory, err = truncate.NewFactory()
			Expect(err).ToNot(HaveOccurred())
			Expect(factory).ToNot(BeNil())
			dataset = upload.Init()
			Expect(dataset).ToNot(BeNil())
			dataset.UserID = "user-id"
			dataset.GroupID = "group-id"
			dataset.DeviceID = StringAsPointer("device-id")
		})

		Context("CanDeduplicateDataset", func() {
			It("returns an error if the dataset is missing", func() {
				can, err := factory.CanDeduplicateDataset(nil)
				Expect(err).To(MatchError("truncate: dataset is missing"))
				Expect(can).To(Equal(false))
			})

			It("returns false if the dataset id is missing", func() {
				dataset.UploadID = ""
				Expect(factory.CanDeduplicateDataset(dataset)).To(Equal(false))
			})

			It("returns false if the user id is missing", func() {
				dataset.UserID = ""
				Expect(factory.CanDeduplicateDataset(dataset)).To(Equal(false))
			})

			It("returns false if the group id is missing", func() {
				dataset.GroupID = ""
				Expect(factory.CanDeduplicateDataset(dataset)).To(Equal(false))
			})

			It("returns false if the device id is missing", func() {
				dataset.DeviceID = nil
				Expect(factory.CanDeduplicateDataset(dataset)).To(Equal(false))
			})

			It("returns false if the device id is empty", func() {
				dataset.DeviceID = StringAsPointer("")
				Expect(factory.CanDeduplicateDataset(dataset)).To(Equal(false))
			})

			It("returns true if the device id is specified", func() {
				Expect(factory.CanDeduplicateDataset(dataset)).To(Equal(true))
			})

			Context("with deduplicator", func() {
				BeforeEach(func() {
					dataset.Deduplicator = &upload.Deduplicator{}
				})

				It("returns false if the deduplicator name is missing", func() {
					Expect(factory.CanDeduplicateDataset(dataset)).To(Equal(false))
				})

				It("returns true if the deduplicator name is not truncate", func() {
					dataset.Deduplicator.Name = "not-truncate"
					Expect(factory.CanDeduplicateDataset(dataset)).To(Equal(false))
				})

				It("returns true if the deduplicator name is truncate", func() {
					dataset.Deduplicator.Name = "truncate"
					Expect(factory.CanDeduplicateDataset(dataset)).To(Equal(true))
				})
			})
		})

		Context("NewDeduplicator", func() {
			It("returns an error if the logger is missing", func() {
				truncateDeduplicator, err := factory.NewDeduplicator(nil, &TestDataStoreSession{}, dataset)
				Expect(err).To(MatchError("truncate: logger is missing"))
				Expect(truncateDeduplicator).To(BeNil())
			})

			It("returns an error if the data store session is missing", func() {
				truncateDeduplicator, err := factory.NewDeduplicator(log.NewNullLogger(), nil, dataset)
				Expect(err).To(MatchError("truncate: data store session is missing"))
				Expect(truncateDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset is missing", func() {
				truncateDeduplicator, err := factory.NewDeduplicator(log.NewNullLogger(), &TestDataStoreSession{}, nil)
				Expect(err).To(MatchError("truncate: dataset is missing"))
				Expect(truncateDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset id is missing", func() {
				dataset.UploadID = ""
				truncateDeduplicator, err := factory.NewDeduplicator(log.NewNullLogger(), &TestDataStoreSession{}, dataset)
				Expect(err).To(MatchError("truncate: dataset id is missing"))
				Expect(truncateDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset user id is missing", func() {
				dataset.UserID = ""
				truncateDeduplicator, err := factory.NewDeduplicator(log.NewNullLogger(), &TestDataStoreSession{}, dataset)
				Expect(err).To(MatchError("truncate: dataset user id is missing"))
				Expect(truncateDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset group id is missing", func() {
				dataset.GroupID = ""
				truncateDeduplicator, err := factory.NewDeduplicator(log.NewNullLogger(), &TestDataStoreSession{}, dataset)
				Expect(err).To(MatchError("truncate: dataset group id is missing"))
				Expect(truncateDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset device id is missing", func() {
				dataset.DeviceID = nil
				truncateDeduplicator, err := factory.NewDeduplicator(log.NewNullLogger(), &TestDataStoreSession{}, dataset)
				Expect(err).To(MatchError("truncate: dataset device id is missing"))
				Expect(truncateDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset device id is empty", func() {
				dataset.DeviceID = StringAsPointer("")
				truncateDeduplicator, err := factory.NewDeduplicator(log.NewNullLogger(), &TestDataStoreSession{}, dataset)
				Expect(err).To(MatchError("truncate: dataset device id is missing"))
				Expect(truncateDeduplicator).To(BeNil())
			})

			It("returns a new deduplicator upon success", func() {
				Expect(factory.NewDeduplicator(log.NewNullLogger(), &TestDataStoreSession{}, dataset)).ToNot(BeNil())
			})
		})

		Context("with a new deduplicator", func() {
			var testDataStoreSession *TestDataStoreSession
			var truncateDeduplicator deduplicator.Deduplicator

			BeforeEach(func() {
				var err error
				testDataStoreSession = &TestDataStoreSession{}
				truncateDeduplicator, err = factory.NewDeduplicator(log.NewNullLogger(), testDataStoreSession, dataset)
				Expect(err).ToNot(HaveOccurred())
				Expect(truncateDeduplicator).ToNot(BeNil())
			})

			Context("InitializeDataset", func() {
				It("returns an error if there is an error", func() {
					testDataStoreSession.UpdateDatasetOutputs = []error{errors.New("test error")}
					err := truncateDeduplicator.InitializeDataset()
					Expect(err).To(MatchError("truncate: unable to initialize dataset; test error"))
					Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(dataset))
					Expect(testDataStoreSession.UpdateDatasetOutputs).To(BeEmpty())
				})

				It("returns successfully if there is no error", func() {
					testDataStoreSession.UpdateDatasetOutputs = []error{nil}
					Expect(truncateDeduplicator.InitializeDataset()).To(Succeed())
					Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(dataset))
					Expect(testDataStoreSession.UpdateDatasetOutputs).To(BeEmpty())
				})

				It("sets the dataset deduplicator if there is no error", func() {
					testDataStoreSession.UpdateDatasetOutputs = []error{nil}
					Expect(truncateDeduplicator.InitializeDataset()).To(Succeed())
					Expect(dataset.Deduplicator).To(Equal(&upload.Deduplicator{Name: "truncate"}))
					Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(dataset))
					Expect(testDataStoreSession.UpdateDatasetOutputs).To(BeEmpty())
				})
			})

			Context("AddDataToDataset", func() {
				It("returns an error if the dataset is missing", func() {
					err := truncateDeduplicator.AddDataToDataset(nil)
					Expect(err).To(MatchError("truncate: dataset data is missing"))
				})

				It("returns an error if there is an error", func() {
					testDataStoreSession.CreateDatasetDataOutputs = []error{errors.New("test error")}
					err := truncateDeduplicator.AddDataToDataset([]data.Datum{})
					Expect(err).To(MatchError("truncate: unable to add data to dataset; test error"))
					Expect(testDataStoreSession.CreateDatasetDataInputs).To(ConsistOf(CreateDatasetDataInput{dataset, []data.Datum{}}))
					Expect(testDataStoreSession.CreateDatasetDataOutputs).To(BeEmpty())
				})

				It("returns successfully if there is no error", func() {
					testDataStoreSession.CreateDatasetDataOutputs = []error{nil}
					Expect(truncateDeduplicator.AddDataToDataset([]data.Datum{})).To(Succeed())
					Expect(testDataStoreSession.CreateDatasetDataInputs).To(ConsistOf(CreateDatasetDataInput{dataset, []data.Datum{}}))
					Expect(testDataStoreSession.CreateDatasetDataOutputs).To(BeEmpty())
				})
			})

			Context("FinalizeDataset", func() {
				It("returns an error if there is an error on activate", func() {
					dataset.UploadID = "upload-id"
					testDataStoreSession.ActivateAllDatasetDataOutputs = []error{errors.New("test error")}
					err := truncateDeduplicator.FinalizeDataset()
					Expect(err).To(MatchError("truncate: unable to activate data in dataset with id \"upload-id\"; test error"))
					Expect(testDataStoreSession.ActivateAllDatasetDataInputs).To(ConsistOf(dataset))
					Expect(testDataStoreSession.ActivateAllDatasetDataOutputs).To(BeEmpty())
				})

				It("returns an error if there is an error on remove", func() {
					dataset.UploadID = "upload-id"
					testDataStoreSession.ActivateAllDatasetDataOutputs = []error{nil}
					testDataStoreSession.DeleteAllOtherDatasetDataOutputs = []error{errors.New("test error")}
					err := truncateDeduplicator.FinalizeDataset()
					Expect(err).To(MatchError("truncate: unable to remove all other data except dataset with id \"upload-id\"; test error"))
					Expect(testDataStoreSession.ActivateAllDatasetDataInputs).To(ConsistOf(dataset))
					Expect(testDataStoreSession.ActivateAllDatasetDataOutputs).To(BeEmpty())
					Expect(testDataStoreSession.DeleteAllOtherDatasetDataInputs).To(ConsistOf(dataset))
					Expect(testDataStoreSession.DeleteAllOtherDatasetDataOutputs).To(BeEmpty())
				})

				It("returns successfully if there is no error", func() {
					testDataStoreSession.ActivateAllDatasetDataOutputs = []error{nil}
					testDataStoreSession.DeleteAllOtherDatasetDataOutputs = []error{nil}
					Expect(truncateDeduplicator.FinalizeDataset()).To(Succeed())
					Expect(testDataStoreSession.ActivateAllDatasetDataInputs).To(ConsistOf(dataset))
					Expect(testDataStoreSession.ActivateAllDatasetDataOutputs).To(BeEmpty())
					Expect(testDataStoreSession.DeleteAllOtherDatasetDataInputs).To(ConsistOf(dataset))
					Expect(testDataStoreSession.DeleteAllOtherDatasetDataOutputs).To(BeEmpty())
				})
			})
		})
	})
})
