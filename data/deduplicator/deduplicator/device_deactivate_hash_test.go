package deduplicator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"context"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataStoreDEPRECATEDTest "github.com/tidepool-org/platform/data/storeDEPRECATED/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	dataTypesUploadTest "github.com/tidepool-org/platform/data/types/upload/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("DeviceDeactivateHash", func() {
	It("DeviceDeactivateHashName is expected", func() {
		Expect(dataDeduplicatorDeduplicator.DeviceDeactivateHashName).To(Equal("org.tidepool.deduplicator.device.deactivate.hash"))
	})

	Context("NewDeviceDeactivateHash", func() {
		It("returns succesfully", func() {
			Expect(dataDeduplicatorDeduplicator.NewDeviceDeactivateHash()).ToNot(BeNil())
		})
	})

	Context("with new deduplicator", func() {
		var deduplicator *dataDeduplicatorDeduplicator.DeviceDeactivateHash
		var dataSet *dataTypesUpload.Upload

		BeforeEach(func() {
			var err error
			deduplicator, err = dataDeduplicatorDeduplicator.NewDeviceDeactivateHash()
			Expect(err).ToNot(HaveOccurred())
			Expect(deduplicator).ToNot(BeNil())
			dataSet = dataTypesUploadTest.RandomUpload()
			dataSet.DataSetType = pointer.FromString("normal")
			dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.deduplicator.device.deactivate.hash")
			dataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Abbott"})
			dataSet.DeviceModel = pointer.FromString("FreeStyle Libre")
		})

		Context("New", func() {
			It("returns an error when the data set is missing", func() {
				found, err := deduplicator.New(nil)
				Expect(err).To(MatchError("data set is missing"))
				Expect(found).To(BeFalse())
			})

			It("returns false when the data set type is not normal", func() {
				dataSet.DataSetType = pointer.FromString("continuous")
				Expect(deduplicator.New(dataSet)).To(BeFalse())
			})

			It("returns false when the device id is missing", func() {
				dataSet.DeviceID = nil
				Expect(deduplicator.New(dataSet)).To(BeFalse())
			})

			dataSetTypeValidations := func() {
				It("returns false when the deduplicator name does not match", func() {
					dataSet.Deduplicator.Name = pointer.FromString(netTest.RandomReverseDomain())
					Expect(deduplicator.New(dataSet)).To(BeFalse())
				})

				It("returns true when the deduplicator name matches", func() {
					Expect(deduplicator.New(dataSet)).To(BeTrue())
				})

				It("returns true when the deduplicator name matches deprecated", func() {
					dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.hash-deactivate-old")
					Expect(deduplicator.New(dataSet)).To(BeTrue())
				})

				When("the deduplicator is missing", func() {
					BeforeEach(func() {
						dataSet.Deduplicator = nil
					})

					It("returns false when the device manufacturers is missing", func() {
						dataSet.DeviceManufacturers = nil
						Expect(deduplicator.New(dataSet)).To(BeFalse())
					})

					It("returns false when the device manufacturers is empty", func() {
						dataSet.DeviceManufacturers = pointer.FromStringArray([]string{})
						Expect(deduplicator.New(dataSet)).To(BeFalse())
					})

					It("returns false when the device manufacturers does not match", func() {
						dataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Alpha", "Bravo"})
						Expect(deduplicator.New(dataSet)).To(BeFalse())
					})

					It("returns false when the device model is missing", func() {
						dataSet.DeviceModel = nil
						Expect(deduplicator.New(dataSet)).To(BeFalse())
					})

					It("returns false when the device model is empty", func() {
						dataSet.DeviceModel = pointer.FromString("")
						Expect(deduplicator.New(dataSet)).To(BeFalse())
					})

					It("returns false when the device model does not match", func() {
						dataSet.DeviceModel = pointer.FromString("Alpha")
						Expect(deduplicator.New(dataSet)).To(BeFalse())
					})

					It("returns true when the device manufacturers and device model matches", func() {
						Expect(deduplicator.New(dataSet)).To(BeTrue())
					})

					It("returns true when the device manufacturers and device model matches with multiple", func() {
						dataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Alpha", "Abbott", "Bravo"})
						Expect(deduplicator.New(dataSet)).To(BeTrue())
					})
				})

				DescribeTable("returns true when",
					func(deviceManufacturer string, deviceModel string) {
						dataSet.DeviceManufacturers = pointer.FromStringArray([]string{deviceManufacturer})
						dataSet.DeviceModel = pointer.FromString(deviceModel)
						Expect(deduplicator.New(dataSet)).To(BeTrue())
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
			}

			When("the data set type is missing", func() {
				BeforeEach(func() {
					dataSet.DataSetType = nil
				})

				dataSetTypeValidations()
			})

			When("the data set type is normal", func() {
				BeforeEach(func() {
					dataSet.DataSetType = pointer.FromString("normal")
				})

				dataSetTypeValidations()
			})
		})

		Context("Get", func() {
			It("returns an error when the data set is missing", func() {
				found, err := deduplicator.Get(nil)
				Expect(err).To(MatchError("data set is missing"))
				Expect(found).To(BeFalse())
			})

			It("returns false when the deduplicator is missing", func() {
				dataSet.Deduplicator = nil
				Expect(deduplicator.Get(dataSet)).To(BeFalse())
			})

			It("returns false when the deduplicator name is missing", func() {
				dataSet.Deduplicator.Name = nil
				Expect(deduplicator.Get(dataSet)).To(BeFalse())
			})

			It("returns false when the deduplicator name does not match", func() {
				dataSet.Deduplicator.Name = pointer.FromString(netTest.RandomReverseDomain())
				Expect(deduplicator.Get(dataSet)).To(BeFalse())
			})

			It("returns true when the deduplicator name matches", func() {
				Expect(deduplicator.Get(dataSet)).To(BeTrue())
			})

			It("returns true when the deduplicator name matches deprecated", func() {
				dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.hash-deactivate-old")
				Expect(deduplicator.Get(dataSet)).To(BeTrue())
			})
		})

		Context("with context and session", func() {
			var ctx context.Context
			var session *dataStoreDEPRECATEDTest.DataSession

			BeforeEach(func() {
				ctx = context.Background()
				session = dataStoreDEPRECATEDTest.NewDataSession()
			})

			AfterEach(func() {
				session.AssertOutputsEmpty()
			})

			Context("Open", func() {
				It("returns an error when the context is missing", func() {
					result, err := deduplicator.Open(nil, session, dataSet)
					Expect(err).To(MatchError("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the session is missing", func() {
					result, err := deduplicator.Open(ctx, nil, dataSet)
					Expect(err).To(MatchError("session is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the data set is missing", func() {
					result, err := deduplicator.Open(ctx, session, nil)
					Expect(err).To(MatchError("data set is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the data set deduplicator name does not match", func() {
					dataSet.Deduplicator.Name = pointer.FromString(netTest.RandomReverseDomain())
					result, err := deduplicator.Open(ctx, session, dataSet)
					Expect(err).To(MatchError("data set uses different deduplicator"))
					Expect(result).To(BeNil())
				})

				When("update data set is invoked", func() {
					var update *data.DataSetUpdate

					BeforeEach(func() {
						update = data.NewDataSetUpdate()
						update.Active = pointer.FromBool(false)
						update.Deduplicator = &data.DeduplicatorDescriptor{
							Name:    pointer.FromString("org.tidepool.deduplicator.device.deactivate.hash"),
							Version: pointer.FromString("1.1.0"),
						}
					})

					AfterEach(func() {
						Expect(session.UpdateDataSetInputs).To(Equal([]dataStoreDEPRECATEDTest.UpdateDataSetInput{{Context: ctx, ID: *dataSet.UploadID, Update: update}}))
					})

					When("the data set does not have a deduplicator", func() {
						BeforeEach(func() {
							dataSet.Deduplicator = nil
						})

						It("returns an error when update data set returns an error", func() {
							responseErr := errorsTest.RandomError()
							session.UpdateDataSetOutputs = []dataStoreDEPRECATEDTest.UpdateDataSetOutput{{DataSet: nil, Error: responseErr}}
							result, err := deduplicator.Open(ctx, session, dataSet)
							Expect(err).To(Equal(responseErr))
							Expect(result).To(BeNil())
						})

						It("returns successfully when update data set returns successfully", func() {
							responseDataSet := dataTypesUploadTest.RandomUpload()
							session.UpdateDataSetOutputs = []dataStoreDEPRECATEDTest.UpdateDataSetOutput{{DataSet: responseDataSet, Error: nil}}
							Expect(deduplicator.Open(ctx, session, dataSet)).To(Equal(responseDataSet))
						})
					})

					When("the data set has a deduplicator with matching name and version does not exist", func() {
						BeforeEach(func() {
							dataSet.Deduplicator.Version = nil
						})

						It("returns an error when update data set returns an error", func() {
							responseErr := errorsTest.RandomError()
							session.UpdateDataSetOutputs = []dataStoreDEPRECATEDTest.UpdateDataSetOutput{{DataSet: nil, Error: responseErr}}
							result, err := deduplicator.Open(ctx, session, dataSet)
							Expect(err).To(Equal(responseErr))
							Expect(result).To(BeNil())
						})

						It("returns successfully when update data set returns successfully", func() {
							responseDataSet := dataTypesUploadTest.RandomUpload()
							session.UpdateDataSetOutputs = []dataStoreDEPRECATEDTest.UpdateDataSetOutput{{DataSet: responseDataSet, Error: nil}}
							Expect(deduplicator.Open(ctx, session, dataSet)).To(Equal(responseDataSet))
						})
					})

					When("the data set has a deduplicator with matching name and version exists", func() {
						BeforeEach(func() {
							dataSet.Deduplicator.Version = pointer.FromString(netTest.RandomSemanticVersion())
						})

						It("returns an error when update data set returns an error", func() {
							responseErr := errorsTest.RandomError()
							session.UpdateDataSetOutputs = []dataStoreDEPRECATEDTest.UpdateDataSetOutput{{DataSet: nil, Error: responseErr}}
							result, err := deduplicator.Open(ctx, session, dataSet)
							Expect(err).To(Equal(responseErr))
							Expect(result).To(BeNil())
						})

						It("returns successfully when update data set returns successfully", func() {
							responseDataSet := dataTypesUploadTest.RandomUpload()
							session.UpdateDataSetOutputs = []dataStoreDEPRECATEDTest.UpdateDataSetOutput{{DataSet: responseDataSet, Error: nil}}
							Expect(deduplicator.Open(ctx, session, dataSet)).To(Equal(responseDataSet))
						})
					})
				})
			})

			Context("AddData", func() {
				var dataSetData data.Data

				BeforeEach(func() {
					dataSetData = make(data.Data, test.RandomIntFromRange(1, 3))
					for index := range dataSetData {
						base := dataTypesTest.NewBase()
						base.Deduplicator.Hash = nil
						dataSetData[index] = base
					}
				})

				It("returns an error when the context is missing", func() {
					Expect(deduplicator.AddData(nil, session, dataSet, dataSetData)).To(MatchError("context is missing"))
				})

				It("returns an error when the session is missing", func() {
					Expect(deduplicator.AddData(ctx, nil, dataSet, dataSetData)).To(MatchError("session is missing"))
				})

				It("returns an error when the data set is missing", func() {
					Expect(deduplicator.AddData(ctx, session, nil, dataSetData)).To(MatchError("data set is missing"))
				})

				It("returns an error when the data set data is missing", func() {
					Expect(deduplicator.AddData(ctx, session, dataSet, nil)).To(MatchError("data set data is missing"))
				})

				When("create data set data is invoked", func() {
					AfterEach(func() {
						Expect(session.CreateDataSetDataInputs).To(Equal([]dataStoreDEPRECATEDTest.CreateDataSetDataInput{{Context: ctx, DataSet: dataSet, DataSetData: dataSetData}}))
						for _, datum := range dataSetData {
							base, ok := datum.(*dataTypes.Base)
							Expect(ok).To(BeTrue())
							Expect(base).ToNot(BeNil())
							Expect(base.Deduplicator.Hash).ToNot(BeNil())
						}
					})

					It("returns an error when create data set data returns an error", func() {
						responseErr := errorsTest.RandomError()
						session.CreateDataSetDataOutputs = []error{responseErr}
						Expect(deduplicator.AddData(ctx, session, dataSet, dataSetData)).To(Equal(responseErr))
					})

					It("returns successfully when create data set data returns successfully", func() {
						session.CreateDataSetDataOutputs = []error{nil}
						Expect(deduplicator.AddData(ctx, session, dataSet, dataSetData)).To(Succeed())
					})
				})
			})

			Context("DeleteData", func() {
				var selectors *data.Selectors

				BeforeEach(func() {
					selectors = dataTest.RandomSelectors()
				})

				It("returns an error when the context is missing", func() {
					Expect(deduplicator.DeleteData(nil, session, dataSet, selectors)).To(MatchError("context is missing"))
				})

				It("returns an error when the session is missing", func() {
					Expect(deduplicator.DeleteData(ctx, nil, dataSet, selectors)).To(MatchError("session is missing"))
				})

				It("returns an error when the data set is missing", func() {
					Expect(deduplicator.DeleteData(ctx, session, nil, selectors)).To(MatchError("data set is missing"))
				})

				It("returns an error when the selectors is missing", func() {
					Expect(deduplicator.DeleteData(ctx, session, dataSet, nil)).To(MatchError("selectors is missing"))
				})

				When("delete data set data is invoked", func() {
					AfterEach(func() {
						Expect(session.DeleteDataSetDataInputs).To(Equal([]dataStoreDEPRECATEDTest.DeleteDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
					})

					It("returns an error when delete data set data returns an error", func() {
						responseErr := errorsTest.RandomError()
						session.DeleteDataSetDataOutputs = []error{responseErr}
						Expect(deduplicator.DeleteData(ctx, session, dataSet, selectors)).To(Equal(responseErr))
					})

					It("returns successfully when delete data set data returns successfully", func() {
						session.DeleteDataSetDataOutputs = []error{nil}
						Expect(deduplicator.DeleteData(ctx, session, dataSet, selectors)).To(Succeed())
					})
				})
			})

			Context("Close", func() {
				It("returns an error when the context is missing", func() {
					Expect(deduplicator.Close(nil, session, dataSet)).To(MatchError("context is missing"))
				})

				It("returns an error when the session is missing", func() {
					Expect(deduplicator.Close(ctx, nil, dataSet)).To(MatchError("session is missing"))
				})

				It("returns an error when the data set is missing", func() {
					Expect(deduplicator.Close(ctx, session, nil)).To(MatchError("data set is missing"))
				})

				When("archive device data using hashes from data sets is invoked", func() {
					AfterEach(func() {
						Expect(session.ArchiveDeviceDataUsingHashesFromDataSetInputs).To(Equal([]dataStoreDEPRECATEDTest.ArchiveDeviceDataUsingHashesFromDataSetInput{{Context: ctx, DataSet: dataSet}}))
					})

					It("returns an error when archive device data using hashes from data sets returns an error", func() {
						responseErr := errorsTest.RandomError()
						session.ArchiveDeviceDataUsingHashesFromDataSetOutputs = []error{responseErr}
						Expect(deduplicator.Close(ctx, session, dataSet)).To(Equal(responseErr))
					})

					When("activate data set data is invoked", func() {
						BeforeEach(func() {
							session.ArchiveDeviceDataUsingHashesFromDataSetOutputs = []error{nil}
						})

						AfterEach(func() {
							Expect(session.ActivateDataSetDataInputs).To(Equal([]dataStoreDEPRECATEDTest.ActivateDataSetDataInput{{Context: ctx, DataSet: dataSet}}))
						})

						It("returns an error when active data set data returns an error", func() {
							responseErr := errorsTest.RandomError()
							session.ActivateDataSetDataOutputs = []error{responseErr}
							Expect(deduplicator.Close(ctx, session, dataSet)).To(Equal(responseErr))
						})

						It("returns successfully when active data set data returns successfully", func() {
							session.ActivateDataSetDataOutputs = []error{nil}
							Expect(deduplicator.Close(ctx, session, dataSet)).To(Succeed())
						})
					})
				})
			})

			Context("Delete", func() {
				It("returns an error when the context is missing", func() {
					Expect(deduplicator.Delete(nil, session, dataSet)).To(MatchError("context is missing"))
				})

				It("returns an error when the session is missing", func() {
					Expect(deduplicator.Delete(ctx, nil, dataSet)).To(MatchError("session is missing"))
				})

				It("returns an error when the data set is missing", func() {
					Expect(deduplicator.Delete(ctx, session, nil)).To(MatchError("data set is missing"))
				})

				When("unarchive device data using hashes from data sets is invoked", func() {
					AfterEach(func() {
						Expect(session.UnarchiveDeviceDataUsingHashesFromDataSetInputs).To(Equal([]dataStoreDEPRECATEDTest.UnarchiveDeviceDataUsingHashesFromDataSetInput{{Context: ctx, DataSet: dataSet}}))
					})

					It("returns an error when unarchive device data using hashes from data sets returns an error", func() {
						responseErr := errorsTest.RandomError()
						session.UnarchiveDeviceDataUsingHashesFromDataSetOutputs = []error{responseErr}
						Expect(deduplicator.Delete(ctx, session, dataSet)).To(Equal(responseErr))
					})

					When("delete data set is invoked", func() {
						BeforeEach(func() {
							session.UnarchiveDeviceDataUsingHashesFromDataSetOutputs = []error{nil}
						})

						AfterEach(func() {
							Expect(session.DeleteDataSetInputs).To(Equal([]dataStoreDEPRECATEDTest.DeleteDataSetInput{{Context: ctx, DataSet: dataSet}}))
						})

						It("returns an error when delete data set returns an error", func() {
							responseErr := errorsTest.RandomError()
							session.DeleteDataSetOutputs = []error{responseErr}
							Expect(deduplicator.Delete(ctx, session, dataSet)).To(Equal(responseErr))
						})

						It("returns successfully when delete data set returns successfully", func() {
							session.DeleteDataSetOutputs = []error{nil}
							Expect(deduplicator.Delete(ctx, session, dataSet)).To(Succeed())
						})
					})
				})
			})
		})
	})
})
