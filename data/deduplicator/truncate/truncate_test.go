package truncate_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	"github.com/tidepool-org/platform/data/deduplicator/truncate"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log/test"
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
	RemoveAllOtherDatasetDataInputs  []*upload.Upload
	RemoveAllOtherDatasetDataOutputs []error
}

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

func (t *TestDataStoreSession) RemoveAllOtherDatasetData(dataset *upload.Upload) error {
	t.RemoveAllOtherDatasetDataInputs = append(t.RemoveAllOtherDatasetDataInputs, dataset)
	output := t.RemoveAllOtherDatasetDataOutputs[0]
	t.RemoveAllOtherDatasetDataOutputs = t.RemoveAllOtherDatasetDataOutputs[1:]
	return output
}

func StringPtr(str string) *string { return &str }

var _ = Describe("Truncate", func() {
	Context("NewFactory", func() {
		It("returns a new factory", func() {
			factory, err := truncate.NewFactory()
			Expect(factory).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("with a new factory", func() {
		var factory deduplicator.Factory
		var dataset *upload.Upload

		BeforeEach(func() {
			var err error
			factory, err = truncate.NewFactory()
			Expect(factory).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
			dataset, err = upload.New()
			Expect(dataset).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
			dataset.UserID = "user-id"
			dataset.GroupID = "group-id"
			dataset.DeviceID = StringPtr("device-id")
		})

		Context("CanDeduplicateDataset", func() {
			It("returns an error if the dataset is missing", func() {
				can, err := factory.CanDeduplicateDataset(nil)
				Expect(can).To(Equal(false))
				Expect(err).To(MatchError("truncate: dataset is missing"))
			})

			It("returns false if the dataset id is missing", func() {
				dataset.UploadID = ""
				can, err := factory.CanDeduplicateDataset(dataset)
				Expect(can).To(Equal(false))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns false if the user id is missing", func() {
				dataset.UserID = ""
				can, err := factory.CanDeduplicateDataset(dataset)
				Expect(can).To(Equal(false))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns false if the group id is missing", func() {
				dataset.GroupID = ""
				can, err := factory.CanDeduplicateDataset(dataset)
				Expect(can).To(Equal(false))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns false if the device id is missing", func() {
				dataset.DeviceID = nil
				can, err := factory.CanDeduplicateDataset(dataset)
				Expect(can).To(Equal(false))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns false if the device id is empty", func() {
				dataset.DeviceID = StringPtr("")
				can, err := factory.CanDeduplicateDataset(dataset)
				Expect(can).To(Equal(false))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns true if the device id is specified", func() {
				can, err := factory.CanDeduplicateDataset(dataset)
				Expect(can).To(Equal(true))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("with deduplicator", func() {
				BeforeEach(func() {
					dataset.Deduplicator = &upload.Deduplicator{}
				})

				It("returns false if the deduplicator name is missing", func() {
					can, err := factory.CanDeduplicateDataset(dataset)
					Expect(can).To(Equal(false))
					Expect(err).ToNot(HaveOccurred())
				})

				It("returns true if the deduplicator name is not truncate", func() {
					dataset.Deduplicator.Name = "not-truncate"
					can, err := factory.CanDeduplicateDataset(dataset)
					Expect(can).To(Equal(false))
					Expect(err).ToNot(HaveOccurred())
				})

				It("returns true if the deduplicator name is truncate", func() {
					dataset.Deduplicator.Name = "truncate"
					can, err := factory.CanDeduplicateDataset(dataset)
					Expect(can).To(Equal(true))
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("NewDeduplicator", func() {
			It("returns an error if the logger is missing", func() {
				truncateDeduplicator, err := factory.NewDeduplicator(nil, &TestDataStoreSession{}, dataset)
				Expect(truncateDeduplicator).To(BeNil())
				Expect(err).To(MatchError("truncate: logger is missing"))
			})

			It("returns an error if the data store session is missing", func() {
				truncateDeduplicator, err := factory.NewDeduplicator(test.NewLogger(), nil, dataset)
				Expect(truncateDeduplicator).To(BeNil())
				Expect(err).To(MatchError("truncate: data store session is missing"))
			})

			It("returns an error if the dataset is missing", func() {
				truncateDeduplicator, err := factory.NewDeduplicator(test.NewLogger(), &TestDataStoreSession{}, nil)
				Expect(truncateDeduplicator).To(BeNil())
				Expect(err).To(MatchError("truncate: dataset is missing"))
			})

			It("returns an error if the dataset id is missing", func() {
				dataset.UploadID = ""
				truncateDeduplicator, err := factory.NewDeduplicator(test.NewLogger(), &TestDataStoreSession{}, dataset)
				Expect(truncateDeduplicator).To(BeNil())
				Expect(err).To(MatchError("truncate: dataset id is missing"))
			})

			It("returns an error if the dataset user id is missing", func() {
				dataset.UserID = ""
				truncateDeduplicator, err := factory.NewDeduplicator(test.NewLogger(), &TestDataStoreSession{}, dataset)
				Expect(truncateDeduplicator).To(BeNil())
				Expect(err).To(MatchError("truncate: dataset user id is missing"))
			})

			It("returns an error if the dataset group id is missing", func() {
				dataset.GroupID = ""
				truncateDeduplicator, err := factory.NewDeduplicator(test.NewLogger(), &TestDataStoreSession{}, dataset)
				Expect(truncateDeduplicator).To(BeNil())
				Expect(err).To(MatchError("truncate: dataset group id is missing"))
			})

			It("returns an error if the dataset device id is missing", func() {
				dataset.DeviceID = nil
				truncateDeduplicator, err := factory.NewDeduplicator(test.NewLogger(), &TestDataStoreSession{}, dataset)
				Expect(truncateDeduplicator).To(BeNil())
				Expect(err).To(MatchError("truncate: dataset device id is missing"))
			})

			It("returns an error if the dataset device id is empty", func() {
				dataset.DeviceID = StringPtr("")
				truncateDeduplicator, err := factory.NewDeduplicator(test.NewLogger(), &TestDataStoreSession{}, dataset)
				Expect(truncateDeduplicator).To(BeNil())
				Expect(err).To(MatchError("truncate: dataset device id is missing"))
			})

			It("returns a new deduplicator upon success", func() {
				truncateDeduplicator, err := factory.NewDeduplicator(test.NewLogger(), &TestDataStoreSession{}, dataset)
				Expect(truncateDeduplicator).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with a new deduplicator", func() {
			var testDataStoreSession *TestDataStoreSession
			var truncateDeduplicator deduplicator.Deduplicator

			BeforeEach(func() {
				var err error
				testDataStoreSession = &TestDataStoreSession{}
				truncateDeduplicator, err = factory.NewDeduplicator(test.NewLogger(), testDataStoreSession, dataset)
				Expect(truncateDeduplicator).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
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
					err := truncateDeduplicator.InitializeDataset()
					Expect(err).ToNot(HaveOccurred())
					Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(dataset))
					Expect(testDataStoreSession.UpdateDatasetOutputs).To(BeEmpty())
				})

				It("sets the dataset deduplicator if there is no error", func() {
					testDataStoreSession.UpdateDatasetOutputs = []error{nil}
					err := truncateDeduplicator.InitializeDataset()
					Expect(err).ToNot(HaveOccurred())
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
					err := truncateDeduplicator.AddDataToDataset([]data.Datum{})
					Expect(err).ToNot(HaveOccurred())
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
					testDataStoreSession.RemoveAllOtherDatasetDataOutputs = []error{errors.New("test error")}
					err := truncateDeduplicator.FinalizeDataset()
					Expect(err).To(MatchError("truncate: unable to remove all other data except dataset with id \"upload-id\"; test error"))
					Expect(testDataStoreSession.ActivateAllDatasetDataInputs).To(ConsistOf(dataset))
					Expect(testDataStoreSession.ActivateAllDatasetDataOutputs).To(BeEmpty())
					Expect(testDataStoreSession.RemoveAllOtherDatasetDataInputs).To(ConsistOf(dataset))
					Expect(testDataStoreSession.RemoveAllOtherDatasetDataOutputs).To(BeEmpty())
				})

				It("returns successfully if there is no error", func() {
					testDataStoreSession.ActivateAllDatasetDataOutputs = []error{nil}
					testDataStoreSession.RemoveAllOtherDatasetDataOutputs = []error{nil}
					err := truncateDeduplicator.FinalizeDataset()
					Expect(err).ToNot(HaveOccurred())
					Expect(testDataStoreSession.ActivateAllDatasetDataInputs).To(ConsistOf(dataset))
					Expect(testDataStoreSession.ActivateAllDatasetDataOutputs).To(BeEmpty())
					Expect(testDataStoreSession.RemoveAllOtherDatasetDataInputs).To(ConsistOf(dataset))
					Expect(testDataStoreSession.RemoveAllOtherDatasetDataOutputs).To(BeEmpty())
				})
			})
		})
	})
})
