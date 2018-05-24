package deduplicator_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"errors"
	"fmt"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testDataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED/test"
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
			testDataset = upload.New()
			Expect(testDataset).ToNot(BeNil())
			testDataset.UploadID = &testUploadID
			testDataset.UserID = &testUserID
			testDataset.DeviceID = pointer.FromString(id.New())
			testDataset.DeviceManufacturers = &[]string{"Medtronic"}
			testDataset.DeviceModel = pointer.FromString("523")
		})

		Context("CanDeduplicateDataset", func() {
			It("returns an error if the dataset is missing", func() {
				can, err := testFactory.CanDeduplicateDataset(nil)
				Expect(err).To(MatchError("dataset is missing"))
				Expect(can).To(BeFalse())
			})

			It("returns false if the dataset id is missing", func() {
				testDataset.UploadID = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the dataset id is empty", func() {
				testDataset.UploadID = pointer.FromString("")
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the dataset user id is missing", func() {
				testDataset.UserID = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the dataset user id is empty", func() {
				testDataset.UserID = pointer.FromString("")
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device id is missing", func() {
				testDataset.DeviceID = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device id is empty", func() {
				testDataset.DeviceID = pointer.FromString("")
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device manufacturers is missing", func() {
				testDataset.DeviceManufacturers = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device manufacturers is empty", func() {
				testDataset.DeviceManufacturers = pointer.FromStringArray([]string{})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device model is missing", func() {
				testDataset.DeviceModel = nil
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device model is empty", func() {
				testDataset.DeviceModel = pointer.FromString("")
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device manufacturers does not contain expected device manufacturer", func() {
				testDataset.DeviceManufacturers = pointer.FromStringArray([]string{"Ant", "Zebra", "Cobra"})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the device model does not contain expected device model", func() {
				testDataset.DeviceModel = pointer.FromString("123")
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns true if the device id and expected device manufacturer are specified", func() {
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
			})

			It("returns true if the device id and expected device manufacturer are specified with multiple device manufacturers", func() {
				testDataset.DeviceManufacturers = pointer.FromStringArray([]string{"Ant", "Zebra", "Medtronic", "Cobra"})
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
			})

			DescribeTable("returns true when",
				func(deviceManufacturer string, deviceModel string) {
					testDataset.DeviceManufacturers = pointer.FromStringArray([]string{deviceManufacturer})
					testDataset.DeviceModel = pointer.FromString(deviceModel)
					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
				},
				Entry("is Abbott FreeStyle Libre", "Abbott", "FreeStyle Libre"),
				Entry("is LifeScan OneTouch Ultra 2", "LifeScan", "OneTouch Ultra 2"),
				Entry("is LifeScan OneTouch UltraMini", "LifeScan", "OneTouch UltraMini"),
				Entry("is LifeScan Verio", "LifeScan", "Verio"),
				Entry("is LifeScan Verio Flex", "LifeScan", "Verio Flex"),
				Entry("is Medtronic 523", "Medtronic", "523"),
				Entry("is Medtronic 523K", "Medtronic", "523K"),
				Entry("is Medtronic 551", "Medtronic", "551"),
				Entry("is Medtronic 554", "Medtronic", "554"),
				Entry("is Medtronic 723", "Medtronic", "723"),
				Entry("is Medtronic 723K", "Medtronic", "723K"),
				Entry("is Medtronic 751", "Medtronic", "751"),
				Entry("is Medtronic 754", "Medtronic", "754"),
				Entry("is Medtronic 1510", "Medtronic", "1510"),
				Entry("is Medtronic 1510K", "Medtronic", "1510K"),
				Entry("is Medtronic 1511", "Medtronic", "1511"),
				Entry("is Medtronic 1512", "Medtronic", "1512"),
				Entry("is Medtronic 1580", "Medtronic", "1580"),
				Entry("is Medtronic 1581", "Medtronic", "1581"),
				Entry("is Medtronic 1582", "Medtronic", "1582"),
				Entry("is Medtronic 1710", "Medtronic", "1710"),
				Entry("is Medtronic 1710K", "Medtronic", "1710K"),
				Entry("is Medtronic 1711", "Medtronic", "1711"),
				Entry("is Medtronic 1712", "Medtronic", "1712"),
				Entry("is Medtronic 1714", "Medtronic", "1714"),
				Entry("is Medtronic 1714K", "Medtronic", "1714K"),
				Entry("is Medtronic 1715", "Medtronic", "1715"),
				Entry("is Medtronic 1780", "Medtronic", "1780"),
				Entry("is Medtronic 1781", "Medtronic", "1781"),
				Entry("is Medtronic 1782", "Medtronic", "1782"),
				Entry("is Trividia Health TRUE METRIX", "Trividia Health", "TRUE METRIX"),
				Entry("is Trividia Health TRUE METRIX AIR", "Trividia Health", "TRUE METRIX AIR"),
				Entry("is Trividia Health TRUE METRIX GO", "Trividia Health", "TRUE METRIX GO"),
			)
		})

		Context("with logger and data store session", func() {
			var testLogger log.Logger
			var testDataSession *testDataStoreDEPRECATED.DataSession

			BeforeEach(func() {
				testLogger = null.NewLogger()
				Expect(testLogger).ToNot(BeNil())
				testDataSession = testDataStoreDEPRECATED.NewDataSession()
				Expect(testDataSession).ToNot(BeNil())
			})

			AfterEach(func() {
				testDataSession.Expectations()
			})

			Context("NewDeduplicatorForDataset", func() {
				It("returns an error if the logger is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(nil, testDataSession, testDataset)
					Expect(err).To(MatchError("logger is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data store session is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, nil, testDataset)
					Expect(err).To(MatchError("data store session is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, nil)
					Expect(err).To(MatchError("dataset is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset id is missing", func() {
					testDataset.UploadID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset id is empty", func() {
					testDataset.UploadID = pointer.FromString("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset user id is missing", func() {
					testDataset.UserID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset user id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset user id is empty", func() {
					testDataset.UserID = pointer.FromString("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset user id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset device id is missing", func() {
					testDataset.DeviceID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset device id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the dataset device id is empty", func() {
					testDataset.DeviceID = pointer.FromString("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset device id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers is missing", func() {
					testDataset.DeviceManufacturers = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset device manufacturers is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers is empty", func() {
					testDataset.DeviceManufacturers = pointer.FromStringArray([]string{})
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset device manufacturer and model does not contain expected device manufacturers and models"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device model is missing", func() {
					testDataset.DeviceModel = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset device model is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device model is empty", func() {
					testDataset.DeviceModel = pointer.FromString("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset device manufacturer and model does not contain expected device manufacturers and models"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers does not contain expected device manufacturer", func() {
					testDataset.DeviceManufacturers = pointer.FromStringArray([]string{"Ant", "Zebra", "Cobra"})
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset device manufacturer and model does not contain expected device manufacturers and models"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device model does not contain expected device model", func() {
					testDataset.DeviceModel = pointer.FromString("123")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)
					Expect(err).To(MatchError("dataset device manufacturer and model does not contain expected device manufacturers and models"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns a new deduplicator upon success", func() {
					Expect(testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)).ToNot(BeNil())
				})

				It("returns a new deduplicator upon success if the device id and expected device manufacturer are specified with multiple device manufacturers", func() {
					testDataset.DeviceManufacturers = pointer.FromStringArray([]string{"Ant", "Zebra", "Medtronic", "Cobra"})
					Expect(testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)).ToNot(BeNil())
				})

				DescribeTable("returns a new deduplicator when",
					func(deviceManufacturer string, deviceModel string) {
						testDataset.DeviceManufacturers = pointer.FromStringArray([]string{deviceManufacturer})
						testDataset.DeviceModel = pointer.FromString(deviceModel)
						Expect(testFactory.NewDeduplicatorForDataset(testLogger, testDataSession, testDataset)).ToNot(BeNil())
					},
					Entry("is Abbott FreeStyle Libre", "Abbott", "FreeStyle Libre"),
					Entry("is LifeScan OneTouch Ultra 2", "LifeScan", "OneTouch Ultra 2"),
					Entry("is LifeScan OneTouch UltraMini", "LifeScan", "OneTouch UltraMini"),
					Entry("is LifeScan Verio", "LifeScan", "Verio"),
					Entry("is LifeScan Verio Flex", "LifeScan", "Verio Flex"),
					Entry("is Medtronic 523", "Medtronic", "523"),
					Entry("is Medtronic 523K", "Medtronic", "523K"),
					Entry("is Medtronic 551", "Medtronic", "551"),
					Entry("is Medtronic 554", "Medtronic", "554"),
					Entry("is Medtronic 723", "Medtronic", "723"),
					Entry("is Medtronic 723K", "Medtronic", "723K"),
					Entry("is Medtronic 751", "Medtronic", "751"),
					Entry("is Medtronic 754", "Medtronic", "754"),
					Entry("is Medtronic 1510", "Medtronic", "1510"),
					Entry("is Medtronic 1510K", "Medtronic", "1510K"),
					Entry("is Medtronic 1511", "Medtronic", "1511"),
					Entry("is Medtronic 1512", "Medtronic", "1512"),
					Entry("is Medtronic 1580", "Medtronic", "1580"),
					Entry("is Medtronic 1581", "Medtronic", "1581"),
					Entry("is Medtronic 1582", "Medtronic", "1582"),
					Entry("is Medtronic 1710", "Medtronic", "1710"),
					Entry("is Medtronic 1710K", "Medtronic", "1710K"),
					Entry("is Medtronic 1711", "Medtronic", "1711"),
					Entry("is Medtronic 1712", "Medtronic", "1712"),
					Entry("is Medtronic 1714", "Medtronic", "1714"),
					Entry("is Medtronic 1714K", "Medtronic", "1714K"),
					Entry("is Medtronic 1715", "Medtronic", "1715"),
					Entry("is Medtronic 1780", "Medtronic", "1780"),
					Entry("is Medtronic 1781", "Medtronic", "1781"),
					Entry("is Medtronic 1782", "Medtronic", "1782"),
					Entry("is Trividia Health TRUE METRIX", "Trividia Health", "TRUE METRIX"),
					Entry("is Trividia Health TRUE METRIX AIR", "Trividia Health", "TRUE METRIX AIR"),
					Entry("is Trividia Health TRUE METRIX GO", "Trividia Health", "TRUE METRIX GO"),
				)
			})

			Context("with a context and new deduplicator", func() {
				var ctx context.Context
				var testDeduplicator data.Deduplicator
				var testDataData []*testData.Datum
				var testDatasetData []data.Datum

				BeforeEach(func() {
					ctx = context.Background()
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
						testDataDatum.Expectations()
					}
				})

				Context("AddDatasetData", func() {
					It("returns successfully if the data is nil", func() {
						Expect(testDeduplicator.AddDatasetData(ctx, nil)).To(Succeed())
					})

					It("returns successfully if there is no data", func() {
						Expect(testDeduplicator.AddDatasetData(ctx, []data.Datum{})).To(Succeed())
					})

					It("returns an error if any datum returns an error getting identity fields", func() {
						testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), id.New()}, Error: nil}}
						testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: nil, Error: errors.New("test error")}}
						err := testDeduplicator.AddDatasetData(ctx, testDatasetData)
						Expect(err).To(MatchError("unable to gather identity fields for datum; test error"))
					})

					It("returns an error if any datum returns no identity fields", func() {
						testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), id.New()}, Error: nil}}
						testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: nil, Error: nil}}
						err := testDeduplicator.AddDatasetData(ctx, testDatasetData)
						Expect(err).To(MatchError("unable to generate identity hash for datum; identity fields are missing"))
					})

					It("returns an error if any datum returns empty identity fields", func() {
						testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), id.New()}, Error: nil}}
						testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{}, Error: nil}}
						err := testDeduplicator.AddDatasetData(ctx, testDatasetData)
						Expect(err).To(MatchError("unable to generate identity hash for datum; identity fields are missing"))
					})

					It("returns an error if any datum returns any empty identity fields", func() {
						testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), id.New()}, Error: nil}}
						testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), ""}, Error: nil}}
						err := testDeduplicator.AddDatasetData(ctx, testDatasetData)
						Expect(err).To(MatchError("unable to generate identity hash for datum; identity field is empty"))
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
								Expect(testDataSession.CreateDatasetDataInputs).To(ConsistOf(testDataStoreDEPRECATED.CreateDatasetDataInput{
									Context:     ctx,
									Dataset:     testDataset,
									DatasetData: testDatasetData,
								}))
							})

							It("returns an error if there is an error with CreateDatasetDataInput", func() {
								testDataSession.CreateDatasetDataOutputs = []error{errors.New("test error")}
								err := testDeduplicator.AddDatasetData(ctx, testDatasetData)
								Expect(err).To(MatchError(fmt.Sprintf("unable to create dataset data with id %q; test error", testUploadID)))
							})

							It("returns successfully if there is no error", func() {
								Expect(testDeduplicator.AddDatasetData(ctx, testDatasetData)).To(Succeed())
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
							Expect(testDataSession.ArchiveDeviceDataUsingHashesFromDatasetInputs).To(ConsistOf(testDataStoreDEPRECATED.ArchiveDeviceDataUsingHashesFromDatasetInput{Context: ctx, Dataset: testDataset}))
						})

						It("returns an error if there is an error with ArchiveDeviceDataUsingHashesFromDataset", func() {
							testDataSession.ArchiveDeviceDataUsingHashesFromDatasetOutputs = []error{errors.New("test error")}
							err := testDeduplicator.DeduplicateDataset(ctx)
							Expect(err).To(MatchError(fmt.Sprintf("unable to archive device data using hashes from dataset with id %q; test error", testUploadID)))
						})

						Context("with activating dataset data", func() {
							BeforeEach(func() {
								testDataSession.ActivateDatasetDataOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(testDataSession.ActivateDatasetDataInputs).To(ConsistOf(testDataStoreDEPRECATED.ActivateDatasetDataInput{Context: ctx, Dataset: testDataset}))
							})

							It("returns an error if there is an error with ActivateDatasetData", func() {
								testDataSession.ActivateDatasetDataOutputs = []error{errors.New("test error")}
								err := testDeduplicator.DeduplicateDataset(ctx)
								Expect(err).To(MatchError(fmt.Sprintf("unable to activate dataset data with id %q; test error", testUploadID)))
							})

							It("returns successfully if there is no error", func() {
								Expect(testDeduplicator.DeduplicateDataset(ctx)).To(Succeed())
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
							Expect(testDataSession.UnarchiveDeviceDataUsingHashesFromDatasetInputs).To(ConsistOf(testDataStoreDEPRECATED.UnarchiveDeviceDataUsingHashesFromDatasetInput{Context: ctx, Dataset: testDataset}))
						})

						It("returns an error if there is an error with UnarchiveDeviceDataUsingHashesFromDataset", func() {
							testDataSession.UnarchiveDeviceDataUsingHashesFromDatasetOutputs = []error{errors.New("test error")}
							err := testDeduplicator.DeleteDataset(ctx)
							Expect(err).To(MatchError(fmt.Sprintf("unable to unarchive device data using hashes from dataset with id %q; test error", testUploadID)))
						})

						Context("with deleting dataset", func() {
							BeforeEach(func() {
								testDataSession.DeleteDatasetOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(testDataSession.DeleteDatasetInputs).To(ConsistOf(testDataStoreDEPRECATED.DeleteDatasetInput{Context: ctx, Dataset: testDataset}))
							})

							It("returns an error if there is an error with DeleteDataset", func() {
								testDataSession.DeleteDatasetOutputs = []error{errors.New("test error")}
								err := testDeduplicator.DeleteDataset(ctx)
								Expect(err).To(MatchError(fmt.Sprintf("unable to delete dataset with id %q; test error", testUploadID)))
							})

							It("returns successfully if there is no error", func() {
								Expect(testDeduplicator.DeleteDataset(ctx)).To(Succeed())
							})
						})
					})
				})
			})
		})
	})
})
