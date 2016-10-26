package deduplicator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testDataStore "github.com/tidepool-org/platform/data/store/test"
	testData "github.com/tidepool-org/platform/data/test"
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
			testDataset.UserID = app.NewID()
			testDataset.GroupID = app.NewID()
			testDataset.DeviceID = app.StringAsPointer(app.NewID())
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
					testDataset.Deduplicator = &data.DeduplicatorDescriptor{Name: "truncate"}
				})

				It("returns false if the deduplicator name is empty", func() {
					testDataset.Deduplicator.Name = ""
					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(false))
				})

				It("returns true if the deduplicator name does not match", func() {
					testDataset.Deduplicator.Name = "not-truncate"
					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(false))
				})

				It("returns true if the deduplicator name does match", func() {
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

			It("returns true if the device id and expected device manufacturer are specified", func() {
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(true))
			})

			It("returns true if the device id and expected device manufacturer are specified with multiple device manufacturers", func() {
				testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Ant", "Zebra", "Animas", "Cobra"})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(Equal(true))
			})
		})

		Context("NewDeduplicator", func() {
			It("returns an error if the logger is missing", func() {
				testDeduplicator, err := testFactory.NewDeduplicator(nil, testDataStore.NewSession(), testDataset)
				Expect(err).To(MatchError("deduplicator: logger is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the data store session is missing", func() {
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), nil, testDataset)
				Expect(err).To(MatchError("deduplicator: data store session is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset is missing", func() {
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), nil)
				Expect(err).To(MatchError("deduplicator: dataset is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset id is missing", func() {
				testDataset.UploadID = ""
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)
				Expect(err).To(MatchError("deduplicator: dataset id is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset user id is missing", func() {
				testDataset.UserID = ""
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)
				Expect(err).To(MatchError("deduplicator: dataset user id is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset group id is missing", func() {
				testDataset.GroupID = ""
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)
				Expect(err).To(MatchError("deduplicator: dataset group id is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset device id is missing", func() {
				testDataset.DeviceID = nil
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)
				Expect(err).To(MatchError("deduplicator: dataset device id is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset device id is empty", func() {
				testDataset.DeviceID = app.StringAsPointer("")
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)
				Expect(err).To(MatchError("deduplicator: dataset device id is empty"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset device manufacturers is missing", func() {
				testDataset.DeviceManufacturers = nil
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)
				Expect(err).To(MatchError("deduplicator: dataset device manufacturers is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns false if the device manufacturers is empty", func() {
				testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{})
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)
				Expect(err).To(MatchError("deduplicator: dataset device manufacturers does not contain expected device manufacturer"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns false if the device manufacturers does not contain expected device manufacturer", func() {
				testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Ant", "Zebra", "Cobra"})
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)
				Expect(err).To(MatchError("deduplicator: dataset device manufacturers does not contain expected device manufacturer"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns a new deduplicator upon success", func() {
				Expect(testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)).ToNot(BeNil())
			})

			It("returns a new deduplicator upon success with multiple device manufacturers", func() {
				testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Ant", "Zebra", "Animas", "Cobra"})
				Expect(testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)).ToNot(BeNil())
			})
		})

		Context("with a new deduplicator", func() {
			var testDataStoreSession *testDataStore.Session
			var testDeduplicator data.Deduplicator

			BeforeEach(func() {
				var err error
				testDataStoreSession = testDataStore.NewSession()
				testDeduplicator, err = testFactory.NewDeduplicator(log.NewNull(), testDataStoreSession, testDataset)
				Expect(err).ToNot(HaveOccurred())
				Expect(testDeduplicator).ToNot(BeNil())
			})

			AfterEach(func() {
				Expect(testDataStoreSession.UnusedOutputsCount()).To(Equal(0))
			})

			Context("InitializeDataset", func() {
				It("returns an error if there is an error", func() {
					testDataStoreSession.UpdateDatasetOutputs = []error{errors.New("test error")}
					err := testDeduplicator.InitializeDataset()
					Expect(err).To(MatchError("deduplicator: unable to initialize dataset; test error"))
					Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(testDataset))
				})

				It("returns successfully if there is no error", func() {
					testDataStoreSession.UpdateDatasetOutputs = []error{nil}
					Expect(testDeduplicator.InitializeDataset()).To(Succeed())
					Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(testDataset))
				})

				It("sets the dataset deduplicator if there is no error", func() {
					testDataStoreSession.UpdateDatasetOutputs = []error{nil}
					Expect(testDeduplicator.InitializeDataset()).To(Succeed())
					Expect(testDataset.DeduplicatorDescriptor()).To(Equal(&data.DeduplicatorDescriptor{Name: "truncate"}))
					Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(testDataset))
				})
			})

			Context("AddDataToDataset", func() {
				var testDataData []*testData.Datum
				var testDatasetData []data.Datum

				BeforeEach(func() {
					testDataData = []*testData.Datum{}
					testDatasetData = []data.Datum{}
					for i := 0; i < 3; i++ {
						testDatum := testData.NewDatum()
						testDataData = append(testDataData, testDatum)
						testDatasetData = append(testDatasetData, testDatum)
					}
				})

				AfterEach(func() {
					for _, testDataDatum := range testDataData {
						Expect(testDataDatum.UnusedOutputsCount()).To(Equal(0))
					}
				})

				It("returns an error if the dataset is missing", func() {
					err := testDeduplicator.AddDataToDataset(nil)
					Expect(err).To(MatchError("deduplicator: dataset data is missing"))
				})

				It("returns successfully if there is no data", func() {
					Expect(testDeduplicator.AddDataToDataset([]data.Datum{})).To(Succeed())
				})

				It("returns an error if there is an error with CreateDatasetDataInput", func() {
					testDataStoreSession.CreateDatasetDataOutputs = []error{errors.New("test error")}
					err := testDeduplicator.AddDataToDataset(testDatasetData)
					Expect(err).To(MatchError("deduplicator: unable to add data to dataset; test error"))
					Expect(testDataStoreSession.CreateDatasetDataInputs).To(ConsistOf(testDataStore.CreateDatasetDataInput{Dataset: testDataset, DatasetData: testDatasetData}))
				})

				It("returns successfully if there is no error", func() {
					testDataStoreSession.CreateDatasetDataOutputs = []error{nil}
					Expect(testDeduplicator.AddDataToDataset(testDatasetData)).To(Succeed())
					Expect(testDataStoreSession.CreateDatasetDataInputs).To(ConsistOf(testDataStore.CreateDatasetDataInput{Dataset: testDataset, DatasetData: testDatasetData}))
				})
			})

			Context("FinalizeDataset", func() {
				It("returns an error if there is an error on activate", func() {
					testDataset.UploadID = "upload-id"
					testDataStoreSession.ActivateDatasetDataOutputs = []error{errors.New("test error")}
					err := testDeduplicator.FinalizeDataset()
					Expect(err).To(MatchError(`deduplicator: unable to activate data in dataset with id "upload-id"; test error`))
					Expect(testDataStoreSession.ActivateDatasetDataInputs).To(ConsistOf(testDataset))
				})

				It("returns an error if there is an error on remove", func() {
					testDataset.UploadID = "upload-id"
					testDataStoreSession.ActivateDatasetDataOutputs = []error{nil}
					testDataStoreSession.DeleteOtherDatasetDataOutputs = []error{errors.New("test error")}
					err := testDeduplicator.FinalizeDataset()
					Expect(err).To(MatchError(`deduplicator: unable to remove all other data except dataset with id "upload-id"; test error`))
					Expect(testDataStoreSession.ActivateDatasetDataInputs).To(ConsistOf(testDataset))
					Expect(testDataStoreSession.DeleteOtherDatasetDataInputs).To(ConsistOf(testDataset))
				})

				It("returns successfully if there is no error", func() {
					testDataStoreSession.ActivateDatasetDataOutputs = []error{nil}
					testDataStoreSession.DeleteOtherDatasetDataOutputs = []error{nil}
					Expect(testDeduplicator.FinalizeDataset()).To(Succeed())
					Expect(testDataStoreSession.ActivateDatasetDataInputs).To(ConsistOf(testDataset))
					Expect(testDataStoreSession.DeleteOtherDatasetDataInputs).To(ConsistOf(testDataset))
				})
			})
		})
	})
})
