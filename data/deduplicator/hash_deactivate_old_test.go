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
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/user"
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
		var testDataSet *upload.Upload

		BeforeEach(func() {
			var err error
			testFactory, err = deduplicator.NewHashDeactivateOldFactory()
			Expect(err).ToNot(HaveOccurred())
			Expect(testFactory).ToNot(BeNil())
			testUploadID = data.NewSetID()
			testUserID = user.NewID()
			testDataSet = upload.New()
			Expect(testDataSet).ToNot(BeNil())
			testDataSet.UploadID = &testUploadID
			testDataSet.UserID = &testUserID
			testDataSet.DeviceID = pointer.FromString(testData.NewDeviceID())
			testDataSet.DeviceManufacturers = &[]string{"Medtronic"}
			testDataSet.DeviceModel = pointer.FromString("523")
		})

		Context("CanDeduplicateDataSet", func() {
			It("returns an error if the data set is missing", func() {
				can, err := testFactory.CanDeduplicateDataSet(nil)
				Expect(err).To(MatchError("data set is missing"))
				Expect(can).To(BeFalse())
			})

			It("returns false if the data set id is missing", func() {
				testDataSet.UploadID = nil
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the data set id is empty", func() {
				testDataSet.UploadID = pointer.FromString("")
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the data set user id is missing", func() {
				testDataSet.UserID = nil
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the data set user id is empty", func() {
				testDataSet.UserID = pointer.FromString("")
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the device id is missing", func() {
				testDataSet.DeviceID = nil
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the device id is empty", func() {
				testDataSet.DeviceID = pointer.FromString("")
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the device manufacturers is missing", func() {
				testDataSet.DeviceManufacturers = nil
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the device manufacturers is empty", func() {
				testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{})
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the device model is missing", func() {
				testDataSet.DeviceModel = nil
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the device model is empty", func() {
				testDataSet.DeviceModel = pointer.FromString("")
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the device manufacturers does not contain expected device manufacturer", func() {
				testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Ant", "Zebra", "Cobra"})
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns false if the device model does not contain expected device model", func() {
				testDataSet.DeviceModel = pointer.FromString("123")
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeFalse())
			})

			It("returns true if the device id and expected device manufacturer are specified", func() {
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeTrue())
			})

			It("returns true if the device id and expected device manufacturer are specified with multiple device manufacturers", func() {
				testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Ant", "Zebra", "Medtronic", "Cobra"})
				Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeTrue())
			})

			DescribeTable("returns true when",
				func(deviceManufacturer string, deviceModel string) {
					testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{deviceManufacturer})
					testDataSet.DeviceModel = pointer.FromString(deviceModel)
					Expect(testFactory.CanDeduplicateDataSet(testDataSet)).To(BeTrue())
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
				Entry("is Abbott FreeStyle Libre", "Abbott", "FreeStyle Libre"),
				Entry("is Diabeloop DBLG1", "Diabeloop", "DBLG1"),
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

			Context("NewDeduplicatorForDataSet", func() {
				It("returns an error if the logger is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(nil, testDataSession, testDataSet)
					Expect(err).To(MatchError("logger is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data store session is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, nil, testDataSet)
					Expect(err).To(MatchError("data store session is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set is missing", func() {
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, nil)
					Expect(err).To(MatchError("data set is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set id is missing", func() {
					testDataSet.UploadID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set id is empty", func() {
					testDataSet.UploadID = pointer.FromString("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set user id is missing", func() {
					testDataSet.UserID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set user id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set user id is empty", func() {
					testDataSet.UserID = pointer.FromString("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set user id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set device id is missing", func() {
					testDataSet.DeviceID = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set device id is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the data set device id is empty", func() {
					testDataSet.DeviceID = pointer.FromString("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set device id is empty"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers is missing", func() {
					testDataSet.DeviceManufacturers = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set device manufacturers is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers is empty", func() {
					testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{})
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set device manufacturer and model does not contain expected device manufacturers and models"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device model is missing", func() {
					testDataSet.DeviceModel = nil
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set device model is missing"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device model is empty", func() {
					testDataSet.DeviceModel = pointer.FromString("")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set device manufacturer and model does not contain expected device manufacturers and models"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device manufacturers does not contain expected device manufacturer", func() {
					testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Ant", "Zebra", "Cobra"})
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set device manufacturer and model does not contain expected device manufacturers and models"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns an error if the device model does not contain expected device model", func() {
					testDataSet.DeviceModel = pointer.FromString("123")
					testDeduplicator, err := testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).To(MatchError("data set device manufacturer and model does not contain expected device manufacturers and models"))
					Expect(testDeduplicator).To(BeNil())
				})

				It("returns a new deduplicator upon success", func() {
					Expect(testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)).ToNot(BeNil())
				})

				It("returns a new deduplicator upon success if the device id and expected device manufacturer are specified with multiple device manufacturers", func() {
					testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Ant", "Zebra", "Medtronic", "Cobra"})
					Expect(testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)).ToNot(BeNil())
				})

				DescribeTable("returns a new deduplicator when",
					func(deviceManufacturer string, deviceModel string) {
						testDataSet.DeviceManufacturers = pointer.FromStringArray([]string{deviceManufacturer})
						testDataSet.DeviceModel = pointer.FromString(deviceModel)
						Expect(testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)).ToNot(BeNil())
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
				var testDataSetData []data.Datum

				BeforeEach(func() {
					ctx = context.Background()
					var err error
					testDeduplicator, err = testFactory.NewDeduplicatorForDataSet(testLogger, testDataSession, testDataSet)
					Expect(err).ToNot(HaveOccurred())
					Expect(testDeduplicator).ToNot(BeNil())
					testDataData = []*testData.Datum{}
					testDataSetData = []data.Datum{}
					for i := 0; i < 3; i++ {
						testDatum := testData.NewDatum()
						testDataData = append(testDataData, testDatum)
						testDataSetData = append(testDataSetData, testDatum)
					}
				})

				AfterEach(func() {
					for _, testDataDatum := range testDataData {
						testDataDatum.Expectations()
					}
				})

				Context("AddDataSetData", func() {
					It("returns successfully if the data is nil", func() {
						Expect(testDeduplicator.AddDataSetData(ctx, nil)).To(Succeed())
					})

					It("returns successfully if there is no data", func() {
						Expect(testDeduplicator.AddDataSetData(ctx, []data.Datum{})).To(Succeed())
					})

					It("returns an error if any datum returns an error getting identity fields", func() {
						testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{user.NewID(), testData.NewDeviceID()}, Error: nil}}
						testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: nil, Error: errors.New("test error")}}
						err := testDeduplicator.AddDataSetData(ctx, testDataSetData)
						Expect(err).To(MatchError("unable to gather identity fields for datum; test error"))
					})

					It("returns an error if any datum returns no identity fields", func() {
						testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{user.NewID(), testData.NewDeviceID()}, Error: nil}}
						testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: nil, Error: nil}}
						err := testDeduplicator.AddDataSetData(ctx, testDataSetData)
						Expect(err).To(MatchError("unable to generate identity hash for datum; identity fields are missing"))
					})

					It("returns an error if any datum returns empty identity fields", func() {
						testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{user.NewID(), testData.NewDeviceID()}, Error: nil}}
						testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{}, Error: nil}}
						err := testDeduplicator.AddDataSetData(ctx, testDataSetData)
						Expect(err).To(MatchError("unable to generate identity hash for datum; identity fields are missing"))
					})

					It("returns an error if any datum returns any empty identity fields", func() {
						testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{user.NewID(), testData.NewDeviceID()}, Error: nil}}
						testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{user.NewID(), ""}, Error: nil}}
						err := testDeduplicator.AddDataSetData(ctx, testDataSetData)
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

						Context("with creating data set data", func() {
							BeforeEach(func() {
								testDataSession.CreateDataSetDataOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(testDataSession.CreateDataSetDataInputs).To(ConsistOf(testDataStoreDEPRECATED.CreateDataSetDataInput{
									Context:     ctx,
									DataSet:     testDataSet,
									DataSetData: testDataSetData,
								}))
							})

							It("returns an error if there is an error with CreateDataSetDataInput", func() {
								testDataSession.CreateDataSetDataOutputs = []error{errors.New("test error")}
								err := testDeduplicator.AddDataSetData(ctx, testDataSetData)
								Expect(err).To(MatchError(fmt.Sprintf("unable to create data set data with id %q; test error", testUploadID)))
							})

							It("returns successfully if there is no error", func() {
								Expect(testDeduplicator.AddDataSetData(ctx, testDataSetData)).To(Succeed())
							})
						})
					})
				})

				Context("DeduplicateDataSet", func() {
					Context("with archive device data using hashes from data set", func() {
						BeforeEach(func() {
							testDataSession.ArchiveDeviceDataUsingHashesFromDataSetOutputs = []error{nil}
						})

						AfterEach(func() {
							Expect(testDataSession.ArchiveDeviceDataUsingHashesFromDataSetInputs).To(ConsistOf(testDataStoreDEPRECATED.ArchiveDeviceDataUsingHashesFromDataSetInput{Context: ctx, DataSet: testDataSet}))
						})

						It("returns an error if there is an error with ArchiveDeviceDataUsingHashesFromDataSet", func() {
							testDataSession.ArchiveDeviceDataUsingHashesFromDataSetOutputs = []error{errors.New("test error")}
							err := testDeduplicator.DeduplicateDataSet(ctx)
							Expect(err).To(MatchError(fmt.Sprintf("unable to archive device data using hashes from data set with id %q; test error", testUploadID)))
						})

						Context("with activating data set data", func() {
							BeforeEach(func() {
								testDataSession.ActivateDataSetDataOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(testDataSession.ActivateDataSetDataInputs).To(ConsistOf(testDataStoreDEPRECATED.ActivateDataSetDataInput{Context: ctx, DataSet: testDataSet}))
							})

							It("returns an error if there is an error with ActivateDataSetData", func() {
								testDataSession.ActivateDataSetDataOutputs = []error{errors.New("test error")}
								err := testDeduplicator.DeduplicateDataSet(ctx)
								Expect(err).To(MatchError(fmt.Sprintf("unable to activate data set data with id %q; test error", testUploadID)))
							})

							It("returns successfully if there is no error", func() {
								Expect(testDeduplicator.DeduplicateDataSet(ctx)).To(Succeed())
							})
						})
					})
				})

				Context("DeleteDataSet", func() {
					Context("with unarchive device data using hashes from data set", func() {
						BeforeEach(func() {
							testDataSession.UnarchiveDeviceDataUsingHashesFromDataSetOutputs = []error{nil}
						})

						AfterEach(func() {
							Expect(testDataSession.UnarchiveDeviceDataUsingHashesFromDataSetInputs).To(ConsistOf(testDataStoreDEPRECATED.UnarchiveDeviceDataUsingHashesFromDataSetInput{Context: ctx, DataSet: testDataSet}))
						})

						It("returns an error if there is an error with UnarchiveDeviceDataUsingHashesFromDataSet", func() {
							testDataSession.UnarchiveDeviceDataUsingHashesFromDataSetOutputs = []error{errors.New("test error")}
							err := testDeduplicator.DeleteDataSet(ctx)
							Expect(err).To(MatchError(fmt.Sprintf("unable to unarchive device data using hashes from data set with id %q; test error", testUploadID)))
						})

						Context("with deleting data set", func() {
							BeforeEach(func() {
								testDataSession.DeleteDataSetOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(testDataSession.DeleteDataSetInputs).To(ConsistOf(testDataStoreDEPRECATED.DeleteDataSetInput{Context: ctx, DataSet: testDataSet}))
							})

							It("returns an error if there is an error with DeleteDataSet", func() {
								testDataSession.DeleteDataSetOutputs = []error{errors.New("test error")}
								err := testDeduplicator.DeleteDataSet(ctx)
								Expect(err).To(MatchError(fmt.Sprintf("unable to delete data set with id %q; test error", testUploadID)))
							})

							It("returns successfully if there is no error", func() {
								Expect(testDeduplicator.DeleteDataSet(ctx)).To(Succeed())
							})
						})
					})
				})
			})
		})
	})
})
