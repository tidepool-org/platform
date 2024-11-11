package deduplicator_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataStoreTest "github.com/tidepool-org/platform/data/store/test"
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

	It("DeviceDeactivateHashVersionCurrent is current-hash", func() {
		Expect(dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionCurrent).To(Equal("current-hash"))
	})

	It("DeviceDeactivateHashVersionLegacy is legacy-hash", func() {
		Expect(dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionLegacy).To(Equal("legacy-hash"))
	})

	It("DeviceDeactivateHashVersionUnknown is unkown-hash", func() {
		Expect(dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionUnknown).To(Equal("unknown-hash"))
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
				found, err := deduplicator.New(context.Background(), nil)
				Expect(err).To(MatchError("data set is missing"))
				Expect(found).To(BeFalse())
			})

			It("returns false when the data set type is not normal", func() {
				dataSet.DataSetType = pointer.FromString("continuous")
				Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
			})

			It("returns false when the device id is missing", func() {
				dataSet.DeviceID = nil
				Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
			})

			It("returns false when the deduplicator name does not match", func() {
				dataSet.Deduplicator.Name = pointer.FromString(netTest.RandomReverseDomain())
				Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
			})

			DescribeTable("returns true when",
				func(deviceManufacturer string, deviceModel string) {
					dataSet.DeviceManufacturers = pointer.FromStringArray([]string{deviceManufacturer})
					dataSet.DeviceModel = pointer.FromString(deviceModel)
					Expect(deduplicator.New(context.Background(), dataSet)).To(BeTrue())
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

			dataSetTypeAssertions := func() {
				It("returns false when the deduplicator name does not match", func() {
					dataSet.Deduplicator.Name = pointer.FromString(netTest.RandomReverseDomain())
					Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
				})

				It("returns true when the deduplicator name matches", func() {
					Expect(deduplicator.New(context.Background(), dataSet)).To(BeTrue())
				})

				It("returns true when the deduplicator name matches deprecated", func() {
					dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.hash-deactivate-old")
					Expect(deduplicator.New(context.Background(), dataSet)).To(BeTrue())
				})

				When("the deduplicator is missing", func() {
					BeforeEach(func() {
						dataSet.Deduplicator = nil
					})

					It("returns false when the device manufacturers is missing", func() {
						dataSet.DeviceManufacturers = nil
						Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
					})

					It("returns false when the device manufacturers is empty", func() {
						dataSet.DeviceManufacturers = pointer.FromStringArray([]string{})
						Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
					})

					It("returns false when the device manufacturers does not match", func() {
						dataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Alpha", "Bravo"})
						Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
					})

					It("returns false when the device model is missing", func() {
						dataSet.DeviceModel = nil
						Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
					})

					It("returns false when the device model is empty", func() {
						dataSet.DeviceModel = pointer.FromString("")
						Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
					})

					It("returns false when the device model does not match", func() {
						dataSet.DeviceModel = pointer.FromString("Alpha")
						Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
					})

					It("returns true when the device manufacturers and device model matches", func() {
						Expect(deduplicator.New(context.Background(), dataSet)).To(BeTrue())
					})

					It("returns true when the device manufacturers and device model matches with multiple", func() {
						dataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Alpha", "Abbott", "Bravo"})
						Expect(deduplicator.New(context.Background(), dataSet)).To(BeTrue())
					})
				})

				DescribeTable("returns true when",
					func(deviceManufacturer string, deviceModel string) {
						dataSet.DeviceManufacturers = pointer.FromStringArray([]string{deviceManufacturer})
						dataSet.DeviceModel = pointer.FromString(deviceModel)
						Expect(deduplicator.New(context.Background(), dataSet)).To(BeTrue())
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

				dataSetTypeAssertions()
			})

			When("the data set type is normal", func() {
				BeforeEach(func() {
					dataSet.DataSetType = pointer.FromString("normal")
				})

				dataSetTypeAssertions()
			})
		})

		Context("Get", func() {
			It("returns an error when the data set is missing", func() {
				found, err := deduplicator.Get(context.Background(), nil)
				Expect(err).To(MatchError("data set is missing"))
				Expect(found).To(BeFalse())
			})

			It("returns false when the deduplicator is missing", func() {
				dataSet.Deduplicator = nil
				Expect(deduplicator.Get(context.Background(), dataSet)).To(BeFalse())
			})

			It("returns false when the deduplicator name is missing", func() {
				dataSet.Deduplicator.Name = nil
				Expect(deduplicator.Get(context.Background(), dataSet)).To(BeFalse())
			})

			It("returns false when the deduplicator name does not match", func() {
				dataSet.Deduplicator.Name = pointer.FromString(netTest.RandomReverseDomain())
				Expect(deduplicator.Get(context.Background(), dataSet)).To(BeFalse())
			})

			It("returns true when the deduplicator name matches", func() {
				Expect(deduplicator.Get(context.Background(), dataSet)).To(BeTrue())
			})

			It("returns true when the deduplicator name matches deprecated", func() {
				dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.hash-deactivate-old")
				Expect(deduplicator.Get(context.Background(), dataSet)).To(BeTrue())
			})

			It("returns true when the deduplicator name matches and legacy id device and model", func() {
				dataSet.Deduplicator.Name = pointer.FromString(dataDeduplicatorDeduplicator.DeviceDeactivateHashName)
				dataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Tandem"})
				dataSet.DeviceModel = pointer.FromString("1002717")
				Expect(deduplicator.Get(context.Background(), dataSet)).To(BeTrue())
			})
		})

		Context("with context and repository", func() {
			var ctx context.Context
			var repository *dataStoreTest.DataRepository

			BeforeEach(func() {
				ctx = context.Background()
				repository = dataStoreTest.NewDataRepository()
			})

			AfterEach(func() {
				repository.AssertOutputsEmpty()
			})

			Context("Open", func() {
				It("returns an error when the context is missing", func() {
					result, err := deduplicator.Open(nil, repository, dataSet)
					Expect(err).To(MatchError("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the repository is missing", func() {
					result, err := deduplicator.Open(ctx, nil, dataSet)
					Expect(err).To(MatchError("repository is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the data set is missing", func() {
					result, err := deduplicator.Open(ctx, repository, nil)
					Expect(err).To(MatchError("data set is missing"))
					Expect(result).To(BeNil())
				})

				When("UpdateDataSet is invoked", func() {
					var update *data.DataSetUpdate

					BeforeEach(func() {
						update = data.NewDataSetUpdate()
						update.Active = pointer.FromBool(false)
						update.Deduplicator = &data.DeduplicatorDescriptor{
							Name:    pointer.FromString("org.tidepool.deduplicator.device.deactivate.hash"),
							Version: pointer.FromString(dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionCurrent),
						}
					})

					AfterEach(func() {
						Expect(repository.UpdateDataSetInputs).To(Equal([]dataStoreTest.UpdateDataSetInput{{Context: ctx, ID: *dataSet.UploadID, Update: update}}))
					})

					When("the data set does not have a deduplicator", func() {
						BeforeEach(func() {
							dataSet.Deduplicator = nil
						})

						It("returns an error when update data set returns an error", func() {
							responseErr := errorsTest.RandomError()
							repository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: nil, Error: responseErr}}
							result, err := deduplicator.Open(ctx, repository, dataSet)
							Expect(err).To(Equal(responseErr))
							Expect(result).To(BeNil())
						})

						It("returns successfully when update data set returns successfully", func() {
							responseDataSet := dataTypesUploadTest.RandomUpload()
							repository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: responseDataSet, Error: nil}}
							Expect(deduplicator.Open(ctx, repository, dataSet)).To(Equal(responseDataSet))
						})
					})

					When("the data set has a deduplicator with matching name and version does not exist", func() {
						BeforeEach(func() {
							dataSet.Deduplicator.Version = nil
						})

						It("returns an error when update data set returns an error", func() {
							responseErr := errorsTest.RandomError()
							repository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: nil, Error: responseErr}}
							result, err := deduplicator.Open(ctx, repository, dataSet)
							Expect(err).To(Equal(responseErr))
							Expect(result).To(BeNil())
						})

						It("returns successfully when update data set returns successfully", func() {
							responseDataSet := dataTypesUploadTest.RandomUpload()
							repository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: responseDataSet, Error: nil}}
							Expect(deduplicator.Open(ctx, repository, dataSet)).To(Equal(responseDataSet))
						})
					})

					When("the data set has a deduplicator with matching name and version exists", func() {
						BeforeEach(func() {
							dataSet.Deduplicator.Version = pointer.FromString(netTest.RandomSemanticVersion())
						})

						It("returns an error when update data set returns an error", func() {
							responseErr := errorsTest.RandomError()
							repository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: nil, Error: responseErr}}
							result, err := deduplicator.Open(ctx, repository, dataSet)
							Expect(err).To(Equal(responseErr))
							Expect(result).To(BeNil())
						})

						It("returns successfully when update data set returns successfully", func() {
							responseDataSet := dataTypesUploadTest.RandomUpload()
							repository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: responseDataSet, Error: nil}}
							Expect(deduplicator.Open(ctx, repository, dataSet)).To(Equal(responseDataSet))
						})
					})
				})
			})

			Context("AddData", func() {
				var dataSetData data.Data

				BeforeEach(func() {
					dataSetData = make(data.Data, test.RandomIntFromRange(1, 3))
					for index := range dataSetData {
						base := dataTypesTest.RandomBase()
						base.Deduplicator.Hash = nil
						dataSetData[index] = base
					}
				})

				It("returns an error when the context is missing", func() {
					Expect(deduplicator.AddData(nil, repository, dataSet, dataSetData)).To(MatchError("context is missing"))
				})

				It("returns an error when the repository is missing", func() {
					Expect(deduplicator.AddData(ctx, nil, dataSet, dataSetData)).To(MatchError("repository is missing"))
				})

				It("returns an error when the data set is missing", func() {
					Expect(deduplicator.AddData(ctx, repository, nil, dataSetData)).To(MatchError("data set is missing"))
				})

				It("returns an error when the data set data is missing", func() {
					Expect(deduplicator.AddData(ctx, repository, dataSet, nil)).To(MatchError("data set data is missing"))
				})

				When("create data set data is invoked", func() {
					AfterEach(func() {
						Expect(repository.CreateDataSetDataInputs).To(Equal([]dataStoreTest.CreateDataSetDataInput{{Context: ctx, DataSet: dataSet, DataSetData: dataSetData}}))
						for _, datum := range dataSetData {
							base, ok := datum.(*dataTypes.Base)
							Expect(ok).To(BeTrue())
							Expect(base).ToNot(BeNil())
							Expect(base.Deduplicator.Hash).ToNot(BeNil())
						}
					})

					It("returns an error when create data set data returns an error", func() {
						responseErr := errorsTest.RandomError()
						repository.CreateDataSetDataOutputs = []error{responseErr}
						Expect(deduplicator.AddData(ctx, repository, dataSet, dataSetData)).To(Equal(responseErr))
					})

					It("returns successfully when create data set data returns successfully", func() {
						repository.CreateDataSetDataOutputs = []error{nil}
						Expect(deduplicator.AddData(ctx, repository, dataSet, dataSetData)).To(Succeed())
					})
				})
			})

			Context("DeleteData", func() {
				var selectors *data.Selectors

				BeforeEach(func() {
					selectors = dataTest.RandomSelectors()
				})

				It("returns an error when the context is missing", func() {
					Expect(deduplicator.DeleteData(nil, repository, dataSet, selectors)).To(MatchError("context is missing"))
				})

				It("returns an error when the repository is missing", func() {
					Expect(deduplicator.DeleteData(ctx, nil, dataSet, selectors)).To(MatchError("repository is missing"))
				})

				It("returns an error when the data set is missing", func() {
					Expect(deduplicator.DeleteData(ctx, repository, nil, selectors)).To(MatchError("data set is missing"))
				})

				It("returns an error when the selectors is missing", func() {
					Expect(deduplicator.DeleteData(ctx, repository, dataSet, nil)).To(MatchError("selectors is missing"))
				})

				When("destroy data set data is invoked", func() {
					AfterEach(func() {
						Expect(repository.DestroyDataSetDataInputs).To(Equal([]dataStoreTest.DestroyDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
					})

					It("returns an error when destroy data set data returns an error", func() {
						responseErr := errorsTest.RandomError()
						repository.DestroyDataSetDataOutputs = []error{responseErr}
						Expect(deduplicator.DeleteData(ctx, repository, dataSet, selectors)).To(Equal(responseErr))
					})

					It("returns successfully when destroy data set data returns successfully", func() {
						repository.DestroyDataSetDataOutputs = []error{nil}
						Expect(deduplicator.DeleteData(ctx, repository, dataSet, selectors)).To(Succeed())
					})
				})
			})

			Context("Close", func() {
				It("returns an error when the context is missing", func() {
					Expect(deduplicator.Close(nil, repository, dataSet)).To(MatchError("context is missing"))
				})

				It("returns an error when the repository is missing", func() {
					Expect(deduplicator.Close(ctx, nil, dataSet)).To(MatchError("repository is missing"))
				})

				It("returns an error when the data set is missing", func() {
					Expect(deduplicator.Close(ctx, repository, nil)).To(MatchError("data set is missing"))
				})

				When("archive device data using hashes from data sets is invoked", func() {
					AfterEach(func() {
						Expect(repository.ArchiveDeviceDataUsingHashesFromDataSetInputs).To(Equal([]dataStoreTest.ArchiveDeviceDataUsingHashesFromDataSetInput{{Context: ctx, DataSet: dataSet}}))
					})

					It("returns an error when archive device data using hashes from data sets returns an error", func() {
						responseErr := errorsTest.RandomError()
						repository.ArchiveDeviceDataUsingHashesFromDataSetOutputs = []error{responseErr}
						Expect(deduplicator.Close(ctx, repository, dataSet)).To(Equal(responseErr))
					})

					When("UpdateDataSet is invoked", func() {
						BeforeEach(func() {
							repository.ArchiveDeviceDataUsingHashesFromDataSetOutputs = []error{nil}
						})

						AfterEach(func() {
							Expect(repository.UpdateDataSetInputs).To(Equal([]dataStoreTest.UpdateDataSetInput{{Context: ctx, ID: *dataSet.UploadID, Update: &data.DataSetUpdate{Active: pointer.FromBool(true)}}}))
						})

						It("returns an error when update data set data returns an error", func() {
							responseErr := errorsTest.RandomError()
							repository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: nil, Error: responseErr}}
							Expect(deduplicator.Close(ctx, repository, dataSet)).To(Equal(responseErr))
						})

						When("activate data set data is invoked", func() {
							BeforeEach(func() {
								repository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: dataSet, Error: nil}}
							})

							AfterEach(func() {
								Expect(repository.ActivateDataSetDataInputs).To(Equal([]dataStoreTest.ActivateDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: nil}}))
							})

							It("returns an error when active data set data returns an error", func() {
								responseErr := errorsTest.RandomError()
								repository.ActivateDataSetDataOutputs = []error{responseErr}
								Expect(deduplicator.Close(ctx, repository, dataSet)).To(Equal(responseErr))
							})

							It("returns successfully when active data set data returns successfully", func() {
								repository.ActivateDataSetDataOutputs = []error{nil}
								Expect(deduplicator.Close(ctx, repository, dataSet)).To(Succeed())
							})
						})
					})
				})
			})

			Context("Delete", func() {
				It("returns an error when the context is missing", func() {
					Expect(deduplicator.Delete(nil, repository, dataSet)).To(MatchError("context is missing"))
				})

				It("returns an error when the repository is missing", func() {
					Expect(deduplicator.Delete(ctx, nil, dataSet)).To(MatchError("repository is missing"))
				})

				It("returns an error when the data set is missing", func() {
					Expect(deduplicator.Delete(ctx, repository, nil)).To(MatchError("data set is missing"))
				})

				When("unarchive device data using hashes from data sets is invoked", func() {
					AfterEach(func() {
						Expect(repository.UnarchiveDeviceDataUsingHashesFromDataSetInputs).To(Equal([]dataStoreTest.UnarchiveDeviceDataUsingHashesFromDataSetInput{{Context: ctx, DataSet: dataSet}}))
					})

					It("returns an error when unarchive device data using hashes from data sets returns an error", func() {
						responseErr := errorsTest.RandomError()
						repository.UnarchiveDeviceDataUsingHashesFromDataSetOutputs = []error{responseErr}
						Expect(deduplicator.Delete(ctx, repository, dataSet)).To(Equal(responseErr))
					})

					When("delete data set is invoked", func() {
						BeforeEach(func() {
							repository.UnarchiveDeviceDataUsingHashesFromDataSetOutputs = []error{nil}
						})

						AfterEach(func() {
							Expect(repository.DeleteDataSetInputs).To(Equal([]dataStoreTest.DeleteDataSetInput{{Context: ctx, DataSet: dataSet}}))
						})

						It("returns an error when delete data set returns an error", func() {
							responseErr := errorsTest.RandomError()
							repository.DeleteDataSetOutputs = []error{responseErr}
							Expect(deduplicator.Delete(ctx, repository, dataSet)).To(Equal(responseErr))
						})

						It("returns successfully when delete data set returns successfully", func() {
							repository.DeleteDataSetOutputs = []error{nil}
							Expect(deduplicator.Delete(ctx, repository, dataSet)).To(Succeed())
						})
					})
				})
			})
		})
	})
})
