package deduplicator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"fmt"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testDataStore "github.com/tidepool-org/platform/data/store/test"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("HashDeactivateOld", func() {
	Context("NewHashDeactivateOldFactory", func() {
		It("returns a new factory", func() {
			Expect(deduplicator.NewHashDeactivateOldFactory()).ToNot(BeNil())
		})
	})

	Context("with a new factory", func() {
		var testFactory deduplicator.Factory
		var testUploadID string
		var testUserID string
		var testDataset *upload.Upload

		BeforeEach(func() {
			var err error
			testFactory, err = deduplicator.NewHashDeactivateOldFactory()
			Expect(err).ToNot(HaveOccurred())
			Expect(testFactory).ToNot(BeNil())
			testUploadID = id.New()
			testUserID = id.New()
			testDataset = upload.Init()
			Expect(testDataset).ToNot(BeNil())
			testDataset.UploadID = testUploadID
			testDataset.UserID = testUserID
			testDataset.DeviceID = pointer.String(id.New())
			testDataset.DeviceManufacturers = pointer.StringArray([]string{"Medtronic"})
		})

		Context("CanDeduplicateDataset", func() {
			It("returns an error if the dataset is missing", func() {
				can, err := testFactory.CanDeduplicateDataset(nil)
				Expect(err).To(MatchError("deduplicator: dataset is missing"))
				Expect(can).To(BeFalse())
			})

			It("returns false if the dataset id is missing", func() {
				testDataset.UploadID = ""
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the dataset user id is missing", func() {
				testDataset.UserID = ""
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device id is missing", func() {
				testDataset.DeviceID = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device id is empty", func() {
				testDataset.DeviceID = pointer.String("")
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device manufacturers is missing", func() {
				testDataset.DeviceManufacturers = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device manufacturers is empty", func() {
				testDataset.DeviceManufacturers = pointer.StringArray([]string{})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device manufacturers does not contain expected device manufacturer", func() {
				testDataset.DeviceManufacturers = pointer.StringArray([]string{"Ant", "Zebra", "Cobra"})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns true if the device id and expected device manufacturer are specified", func() {
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
			})

			It("returns true if the device id and expected device manufacturer are specified with multiple device manufacturers", func() {
				testDataset.DeviceManufacturers = pointer.StringArray([]string{"Ant", "Zebra", "Medtronic", "Cobra"})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
			})
		})

		Context("with logger and data store session", func() {
			var testLogger log.Logger
			var testDataSession *testDataStore.DataSession

			BeforeEach(func() {
				testLogger = null.NewLogger()
				Expect(testLogger).ToNot(BeNil())
				testDataSession = testDataStore.NewDataSession()
				Expect(testDataSession).ToNot(BeNil())
			})

			AfterEach(func() {
				Expect(testDataSession.UnusedOutputsCount()).To(Equal(0))
			})

			Context("NewDeduplicatorForDataset", func() {
				It("returns an error if the logger is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(nil, testDataSession, testDataset)
					Expect(err).To(MatchError("deduplicator: logger is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data store session is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, nil, testDataset)
					Expect(err).To(MatchError("deduplicator: data store session is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, nil)
					Expect(err).To(MatchError("deduplicator: dataset is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset id is missing", func() {
					testDataset.UploadID = ""
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset user id is missing", func() {
					testDataset.UserID = ""
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset user id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset device id is missing", func() {
					testDataset.DeviceID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset device id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset device id is empty", func() {
					testDataset.DeviceID = pointer.String("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset device id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers is missing", func() {
					testDataset.DeviceManufacturers = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset device manufacturers does not contain expected device manufacturers"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers is empty", func() {
					testDataset.DeviceManufacturers = pointer.StringArray([]string{})
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset device manufacturers does not contain expected device manufacturers"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers does not contain expected device manufacturer", func() {
					testDataset.DeviceManufacturers = pointer.StringArray([]string{"Ant", "Zebra", "Cobra"})
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("deduplicator: dataset device manufacturers does not contain expected device manufacturers"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns a new deduplicator upon success", func() {
					Expect(testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)).ToNot(BeNil())
				})

				It("returns a new deduplicator upon success if the device id and expected device manufacturer are specified with multiple device manufacturers", func() {
					testDataset.DeviceManufacturers = pointer.StringArray([]string{"Ant", "Zebra", "Medtronic", "Cobra"})
					Expect(testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)).ToNot(BeNil())
				})
			})

			Context("with a new deduplicator", func() {
				var testDeduplicator data.Deduplicator
				var testDataData []*testData.Datum
				var testDatasetData []data.Datum

				BeforeEach(func() {
					var err error
					testDeduplicator, err = testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).ToNot(HaveOccurred())
					Expect(testDeduplicator).ToNot(BeNil())
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

				Context("AddDatasetData", func() {
					It("returns successfully if the data is nil", func() {
						Expect(testDeduplicator.AddDatasetData(nil)).To(Succeed())
					})

					It("returns successfully if there is no data", func() {
						Expect(testDeduplicator.AddDatasetData([]data.Datum{})).To(Succeed())
					})

					It("returns an error if any datum returns an error getting identity fields", func() {
						testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), id.New()}, Error: nil}}
						testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: nil, Error: errors.New("test error")}}
						err := testDeduplicator.AddDatasetData(testDatasetData)
						Expect(err).To(MatchError("deduplicator: unable to gather identity fields for datum; test error"))
					})

					It("returns an error if any datum returns no identity fields", func() {
						testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), id.New()}, Error: nil}}
						testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: nil, Error: nil}}
						err := testDeduplicator.AddDatasetData(testDatasetData)
						Expect(err).To(MatchError("deduplicator: unable to generate identity hash for datum; deduplicator: identity fields are missing"))
					})

					It("returns an error if any datum returns empty identity fields", func() {
						testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), id.New()}, Error: nil}}
						testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{}, Error: nil}}
						err := testDeduplicator.AddDatasetData(testDatasetData)
						Expect(err).To(MatchError("deduplicator: unable to generate identity hash for datum; deduplicator: identity fields are missing"))
					})

					It("returns an error if any datum returns any empty identity fields", func() {
						testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), id.New()}, Error: nil}}
						testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), ""}, Error: nil}}
						err := testDeduplicator.AddDatasetData(testDatasetData)
						Expect(err).To(MatchError("deduplicator: unable to generate identity hash for datum; deduplicator: identity field is empty"))
					})

					Context("with identity fields", func() {
						BeforeEach(func() {
							testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{"test", "0"}, Error: nil}}
							testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{"test", "1"}, Error: nil}}
							testDataData[2].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{"test", "2"}, Error: nil}}
						})

						AfterEach(func() {
							Expect(testDataData[0].DeduplicatorDescriptorValue).To(Equal(&data.DeduplicatorDescriptor{Hash: "GRp47M02cMlAzSn7oJTQ2LC9eb1Qd6mIPO1U8GeuoYg="}))
							Expect(testDataData[1].DeduplicatorDescriptorValue).To(Equal(&data.DeduplicatorDescriptor{Hash: "+cywqM0rcj9REPt87Vfx2U+j9m57cB0XW2kmNZm5Ao8="}))
							Expect(testDataData[2].DeduplicatorDescriptorValue).To(Equal(&data.DeduplicatorDescriptor{Hash: "dCPMoOxFVMbPvMkXMbyKeff8QmdBPu8hr/BVeHJhz78="}))
						})

						Context("with creating dataset data", func() {
							BeforeEach(func() {
								testDataSession.CreateDatasetDataOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(testDataSession.CreateDatasetDataInputs).To(ConsistOf(testDataStore.CreateDatasetDataInput{
									Dataset:     testDataset,
									DatasetData: testDatasetData,
								}))
							})

							It("returns an error if there is an error with CreateDatasetDataInput", func() {
								testDataSession.CreateDatasetDataOutputs = []error{errors.New("test error")}
								err := testDeduplicator.AddDatasetData(testDatasetData)
								Expect(err).To(MatchError(fmt.Sprintf(`deduplicator: unable to create dataset data with id "%s"; test error`, testDataset.UploadID)))
							})

							It("returns successfully if there is no error", func() {
								Expect(testDeduplicator.AddDatasetData(testDatasetData)).To(Succeed())
							})
						})
					})
				})

				Context("DeduplicateDataset", func() {
					Context("with archive device data using hashes from dataset", func() {
						BeforeEach(func() {
							testDataSession.ArchiveDeviceDataUsingHashesFromDatasetOutputs = []error{nil}
						})

						AfterEach(func() {
							Expect(testDataSession.ArchiveDeviceDataUsingHashesFromDatasetInputs).To(Equal([]*upload.Upload{testDataset}))
						})

						It("returns an error if there is an error with ArchiveDeviceDataUsingHashesFromDataset", func() {
							testDataSession.ArchiveDeviceDataUsingHashesFromDatasetOutputs = []error{errors.New("test error")}
							err := testDeduplicator.DeduplicateDataset()
							Expect(err).To(MatchError(fmt.Sprintf(`deduplicator: unable to archive device data using hashes from dataset with id "%s"; test error`, testUploadID)))
						})

						Context("with activating dataset data", func() {
							BeforeEach(func() {
								testDataSession.ActivateDatasetDataOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(testDataSession.ActivateDatasetDataInputs).To(ConsistOf(testDataset))
							})

							It("returns an error if there is an error with ActivateDatasetData", func() {
								testDataSession.ActivateDatasetDataOutputs = []error{errors.New("test error")}
								err := testDeduplicator.DeduplicateDataset()
								Expect(err).To(MatchError(fmt.Sprintf(`deduplicator: unable to activate dataset data with id "%s"; test error`, testUploadID)))
							})

							It("returns successfully if there is no error", func() {
								Expect(testDeduplicator.DeduplicateDataset()).To(Succeed())
							})
						})
					})
				})

				Context("DeleteDataset", func() {
					Context("with unarchive device data using hashes from dataset", func() {
						BeforeEach(func() {
							testDataSession.UnarchiveDeviceDataUsingHashesFromDatasetOutputs = []error{nil}
						})

						AfterEach(func() {
							Expect(testDataSession.UnarchiveDeviceDataUsingHashesFromDatasetInputs).To(Equal([]*upload.Upload{testDataset}))
						})

						It("returns an error if there is an error with UnarchiveDeviceDataUsingHashesFromDataset", func() {
							testDataSession.UnarchiveDeviceDataUsingHashesFromDatasetOutputs = []error{errors.New("test error")}
							err := testDeduplicator.DeleteDataset()
							Expect(err).To(MatchError(fmt.Sprintf(`deduplicator: unable to unarchive device data using hashes from dataset with id "%s"; test error`, testUploadID)))
						})

						Context("with deleting dataset", func() {
							BeforeEach(func() {
								testDataSession.DeleteDatasetOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(testDataSession.DeleteDatasetInputs).To(ConsistOf(testDataset))
							})

							It("returns an error if there is an error with DeleteDataset", func() {
								testDataSession.DeleteDatasetOutputs = []error{errors.New("test error")}
								err := testDeduplicator.DeleteDataset()
								Expect(err).To(MatchError(fmt.Sprintf(`deduplicator: unable to delete dataset with id "%s"; test error`, testUploadID)))
							})

							It("returns successfully if there is no error", func() {
								Expect(testDeduplicator.DeleteDataset()).To(Succeed())
							})
						})
					})
				})
			})
		})
	})
})
