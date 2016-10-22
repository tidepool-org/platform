package deduplicator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testDataStore "github.com/tidepool-org/platform/data/store/test"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/log"
)

var _ = Describe("Truncate", func() {
	Context("NewFactory", func() {
		It("returns a new factory", func() {
			Expect(deduplicator.NewTruncateFactory()).ToNot(BeNil())
		})
	})

	Context("with a new factory", func() {
		var testFactory deduplicator.Factory
		var testDataset *upload.Upload

		BeforeEach(func() {
			var err error
			testFactory, err = deduplicator.NewTruncateFactory()
			Expect(err).ToNot(HaveOccurred())
			Expect(testFactory).ToNot(BeNil())
			testDataset = upload.Init()
			Expect(testDataset).ToNot(BeNil())
			testDataset.UserID = "user-id"
			testDataset.GroupID = "group-id"
			testDataset.DeviceID = app.StringAsPointer("device-id")
			testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Animas"})
		})

		Context("CanDeduplicateDataset", func() {
			It("returns an error if the dataset is missing", func() {
				can, err := testFactory.CanDeduplicateDataset(nil)
				Expect(err).To(MatchError("deduplicator: dataset is missing"))
				Expect(can).To(Equal(false))
			})

			Context("with deduplicator", func() {
				BeforeEach(func() {
					testDataset.Deduplicator = &upload.Deduplicator{}
				})

				It("returns false if the deduplicator name is missing", func() {
					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(false))
				})

				It("returns true if the deduplicator name is not truncate", func() {
					testDataset.Deduplicator.Name = "not-truncate"
					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(false))
				})

				It("returns true if the deduplicator name is truncate", func() {
					testDataset.Deduplicator.Name = "truncate"
					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(true))
				})
			})

			It("returns false if the dataset id is missing", func() {
				testDataset.UploadID = ""
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(false))
			})

			It("returns false if the user id is missing", func() {
				testDataset.UserID = ""
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(false))
			})

			It("returns false if the group id is missing", func() {
				testDataset.GroupID = ""
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(false))
			})

			It("returns false if the device id is missing", func() {
				testDataset.DeviceID = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(false))
			})

			It("returns false if the device id is empty", func() {
				testDataset.DeviceID = app.StringAsPointer("")
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(false))
			})

			It("returns false if the device manufacturers is missing", func() {
				testDataset.DeviceManufacturers = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(false))
			})

			It("returns false if the device manufacturers is empty", func() {
				testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(false))
			})

			It("returns false if the device manufacturers does not contain expected device manufacturer", func() {
				testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Ant", "Zebra", "Cobra"})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(false))
			})

			It("returns true if the device id and expected device manufacturer is specified", func() {
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(true))
			})

			It("returns true if the device id and expected device manufacturer is specified with multiple device manufacturers", func() {
				testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Ant", "Zebra", "Animas", "Cobra"})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(true))
			})
		})

		Context("NewDeduplicator", func() {
			It("returns an error if the logger is missing", func() {
				testTruncateDeduplicator, err := testFactory.NewDeduplicator(nil, &testDataStore.Session{}, testDataset)
				Expect(err).To(MatchError("deduplicator: logger is missing"))
				Expect(testTruncateDeduplicator).To(BeNil())
			})

			It("returns an error if the data store session is missing", func() {
				testTruncateDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), nil, testDataset)
				Expect(err).To(MatchError("deduplicator: data store session is missing"))
				Expect(testTruncateDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset is missing", func() {
				testTruncateDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), &testDataStore.Session{}, nil)
				Expect(err).To(MatchError("deduplicator: dataset is missing"))
				Expect(testTruncateDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset id is missing", func() {
				testDataset.UploadID = ""
				testTruncateDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), &testDataStore.Session{}, testDataset)
				Expect(err).To(MatchError("deduplicator: dataset id is missing"))
				Expect(testTruncateDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset user id is missing", func() {
				testDataset.UserID = ""
				testTruncateDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), &testDataStore.Session{}, testDataset)
				Expect(err).To(MatchError("deduplicator: dataset user id is missing"))
				Expect(testTruncateDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset group id is missing", func() {
				testDataset.GroupID = ""
				testTruncateDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), &testDataStore.Session{}, testDataset)
				Expect(err).To(MatchError("deduplicator: dataset group id is missing"))
				Expect(testTruncateDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset device id is missing", func() {
				testDataset.DeviceID = nil
				testTruncateDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), &testDataStore.Session{}, testDataset)
				Expect(err).To(MatchError("deduplicator: dataset device id is missing"))
				Expect(testTruncateDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset device id is empty", func() {
				testDataset.DeviceID = app.StringAsPointer("")
				testTruncateDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), &testDataStore.Session{}, testDataset)
				Expect(err).To(MatchError("deduplicator: dataset device id is empty"))
				Expect(testTruncateDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset device manufacturers is missing", func() {
				testDataset.DeviceManufacturers = nil
				testTruncateDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), &testDataStore.Session{}, testDataset)
				Expect(err).To(MatchError("deduplicator: dataset device manufacturers is missing"))
				Expect(testTruncateDeduplicator).To(BeNil())
			})

			It("returns false if the device manufacturers is empty", func() {
				testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{})
				testTruncateDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), &testDataStore.Session{}, testDataset)
				Expect(err).To(MatchError("deduplicator: dataset device manufacturers does not contain expected device manufacturer"))
				Expect(testTruncateDeduplicator).To(BeNil())
			})

			It("returns false if the device manufacturers does not contain expected device manufacturer", func() {
				testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Ant", "Zebra", "Cobra"})
				testTruncateDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), &testDataStore.Session{}, testDataset)
				Expect(err).To(MatchError("deduplicator: dataset device manufacturers does not contain expected device manufacturer"))
				Expect(testTruncateDeduplicator).To(BeNil())
			})

			It("returns a new deduplicator upon success", func() {
				Expect(testFactory.NewDeduplicator(log.NewNull(), &testDataStore.Session{}, testDataset)).ToNot(BeNil())
			})

			It("returns a new deduplicator upon success with multiple device manufacturers", func() {
				testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Ant", "Zebra", "Animas", "Cobra"})
				Expect(testFactory.NewDeduplicator(log.NewNull(), &testDataStore.Session{}, testDataset)).ToNot(BeNil())
			})
		})

		Context("with a new deduplicator", func() {
			var testDataStoreSession *testDataStore.Session
			var testTruncateDeduplicator deduplicator.Deduplicator

			BeforeEach(func() {
				var err error
				testDataStoreSession = &testDataStore.Session{}
				testTruncateDeduplicator, err = testFactory.NewDeduplicator(log.NewNull(), testDataStoreSession, testDataset)
				Expect(err).ToNot(HaveOccurred())
				Expect(testTruncateDeduplicator).ToNot(BeNil())
			})

			AfterEach(func() {
				Expect(testDataStoreSession.UnusedOutputsCount()).To(Equal(0))
			})

			Context("InitializeDataset", func() {
				It("returns an error if there is an error", func() {
					testDataStoreSession.UpdateDatasetOutputs = []error{errors.New("test error")}
					err := testTruncateDeduplicator.InitializeDataset()
					Expect(err).To(MatchError("deduplicator: unable to initialize dataset; test error"))
					Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(testDataset))
				})

				It("returns successfully if there is no error", func() {
					testDataStoreSession.UpdateDatasetOutputs = []error{nil}
					Expect(testTruncateDeduplicator.InitializeDataset()).To(Succeed())
					Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(testDataset))
				})

				It("sets the dataset deduplicator if there is no error", func() {
					testDataStoreSession.UpdateDatasetOutputs = []error{nil}
					Expect(testTruncateDeduplicator.InitializeDataset()).To(Succeed())
					Expect(testDataset.Deduplicator).To(Equal(&upload.Deduplicator{Name: "truncate"}))
					Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(testDataset))
				})
			})

			Context("AddDataToDataset", func() {
				It("returns an error if the dataset is missing", func() {
					err := testTruncateDeduplicator.AddDataToDataset(nil)
					Expect(err).To(MatchError("deduplicator: dataset data is missing"))
				})

				It("returns an error if there is an error", func() {
					testDataStoreSession.CreateDatasetDataOutputs = []error{errors.New("test error")}
					err := testTruncateDeduplicator.AddDataToDataset([]data.Datum{})
					Expect(err).To(MatchError("deduplicator: unable to add data to dataset; test error"))
					Expect(testDataStoreSession.CreateDatasetDataInputs).To(ConsistOf(testDataStore.CreateDatasetDataInput{Dataset: testDataset, DatasetData: []data.Datum{}}))
				})

				It("returns successfully if there is no error", func() {
					testDataStoreSession.CreateDatasetDataOutputs = []error{nil}
					Expect(testTruncateDeduplicator.AddDataToDataset([]data.Datum{})).To(Succeed())
					Expect(testDataStoreSession.CreateDatasetDataInputs).To(ConsistOf(testDataStore.CreateDatasetDataInput{Dataset: testDataset, DatasetData: []data.Datum{}}))
				})
			})

			Context("FinalizeDataset", func() {
				It("returns an error if there is an error on activate", func() {
					testDataset.UploadID = "upload-id"
					testDataStoreSession.ActivateDatasetDataOutputs = []error{errors.New("test error")}
					err := testTruncateDeduplicator.FinalizeDataset()
					Expect(err).To(MatchError(`deduplicator: unable to activate data in dataset with id "upload-id"; test error`))
					Expect(testDataStoreSession.ActivateDatasetDataInputs).To(ConsistOf(testDataset))
				})

				It("returns an error if there is an error on remove", func() {
					testDataset.UploadID = "upload-id"
					testDataStoreSession.ActivateDatasetDataOutputs = []error{nil}
					testDataStoreSession.DeleteOtherDatasetDataOutputs = []error{errors.New("test error")}
					err := testTruncateDeduplicator.FinalizeDataset()
					Expect(err).To(MatchError(`deduplicator: unable to remove all other data except dataset with id "upload-id"; test error`))
					Expect(testDataStoreSession.ActivateDatasetDataInputs).To(ConsistOf(testDataset))
					Expect(testDataStoreSession.DeleteOtherDatasetDataInputs).To(ConsistOf(testDataset))
				})

				It("returns successfully if there is no error", func() {
					testDataStoreSession.ActivateDatasetDataOutputs = []error{nil}
					testDataStoreSession.DeleteOtherDatasetDataOutputs = []error{nil}
					Expect(testTruncateDeduplicator.FinalizeDataset()).To(Succeed())
					Expect(testDataStoreSession.ActivateDatasetDataInputs).To(ConsistOf(testDataset))
					Expect(testDataStoreSession.DeleteOtherDatasetDataInputs).To(ConsistOf(testDataset))
				})
			})
		})
	})
})
