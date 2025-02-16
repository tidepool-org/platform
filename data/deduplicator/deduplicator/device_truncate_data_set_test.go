package deduplicator_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataStoreTest "github.com/tidepool-org/platform/data/store/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("DeviceTruncateDataSet", func() {
	It("DeviceTruncateDataSetName is expected", func() {
		Expect(dataDeduplicatorDeduplicator.DeviceTruncateDataSetName).To(Equal("org.tidepool.deduplicator.device.truncate.dataset"))
	})

	Context("with dependencies", func() {
		var dataSetRepository *dataStoreTest.DataRepository
		var dataRepository *dataStoreTest.DataRepository
		var dependencies dataDeduplicatorDeduplicator.Dependencies

		BeforeEach(func() {
			dataSetRepository = dataStoreTest.NewDataRepository()
			dataRepository = dataStoreTest.NewDataRepository()
			dependencies = dataDeduplicatorDeduplicator.Dependencies{
				DataSetStore: dataSetRepository,
				DataStore:    dataRepository,
			}
		})

		AfterEach(func() {
			dataRepository.AssertOutputsEmpty()
			dataSetRepository.AssertOutputsEmpty()
		})

		Context("NewDeviceTruncateDataSet", func() {
			It("returns succesfully", func() {
				Expect(dataDeduplicatorDeduplicator.NewDeviceTruncateDataSet(dependencies)).ToNot(BeNil())
			})
		})

		Context("with new deduplicator", func() {
			var deduplicator *dataDeduplicatorDeduplicator.DeviceTruncateDataSet
			var dataSet *data.DataSet

			BeforeEach(func() {
				var err error
				deduplicator, err = dataDeduplicatorDeduplicator.NewDeviceTruncateDataSet(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(deduplicator).ToNot(BeNil())
				dataSet = dataTest.RandomDataSet()
				dataSet.DataSetType = pointer.FromString("normal")
				dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.deduplicator.device.truncate.dataset")
				dataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Animas"})
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

				dataSetTypeAssertions := func() {
					It("returns false when the deduplicator name does not match", func() {
						dataSet.Deduplicator.Name = pointer.FromString(netTest.RandomReverseDomain())
						Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
					})

					It("returns true when the deduplicator name matches", func() {
						Expect(deduplicator.New(context.Background(), dataSet)).To(BeTrue())
					})

					It("returns true when the deduplicator name matches deprecated", func() {
						dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.truncate")
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

						It("returns true when the device manufacturers matches", func() {
							Expect(deduplicator.New(context.Background(), dataSet)).To(BeTrue())
						})

						It("returns true when the device manufacturers matches with multiple", func() {
							dataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Alpha", "Animas", "Bravo"})
							Expect(deduplicator.New(context.Background(), dataSet)).To(BeTrue())
						})
					})
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
					dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.truncate")
					Expect(deduplicator.Get(context.Background(), dataSet)).To(BeTrue())
				})
			})

			Context("with context", func() {
				var ctx context.Context

				BeforeEach(func() {
					ctx = context.Background()
				})

				Context("Open", func() {
					It("returns an error when the context is missing", func() {
						result, err := deduplicator.Open(nil, dataSet)
						Expect(err).To(MatchError("context is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the data set is missing", func() {
						result, err := deduplicator.Open(ctx, nil)
						Expect(err).To(MatchError("data set is missing"))
						Expect(result).To(BeNil())
					})

					When("UpdateDataSet is invoked", func() {
						var update *data.DataSetUpdate

						BeforeEach(func() {
							update = data.NewDataSetUpdate()
							update.Active = pointer.FromBool(false)
							update.Deduplicator = &data.DeduplicatorDescriptor{
								Name:    pointer.FromString("org.tidepool.deduplicator.device.truncate.dataset"),
								Version: pointer.FromString("1.1.0"),
							}
						})

						AfterEach(func() {
							Expect(dataSetRepository.UpdateDataSetInputs).To(Equal([]dataStoreTest.UpdateDataSetInput{{Context: ctx, ID: *dataSet.UploadID, Update: update}}))
						})

						When("the data set does not have a deduplicator", func() {
							BeforeEach(func() {
								dataSet.Deduplicator = nil
							})

							It("returns an error when update data set returns an error", func() {
								responseErr := errorsTest.RandomError()
								dataSetRepository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: nil, Error: responseErr}}
								result, err := deduplicator.Open(ctx, dataSet)
								Expect(err).To(Equal(responseErr))
								Expect(result).To(BeNil())
							})

							It("returns successfully when update data set returns successfully", func() {
								responseDataSet := dataTest.RandomDataSet()
								dataSetRepository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: responseDataSet, Error: nil}}
								Expect(deduplicator.Open(ctx, dataSet)).To(Equal(responseDataSet))
							})
						})

						When("the data set has a deduplicator with matching name and version does not exist", func() {
							BeforeEach(func() {
								dataSet.Deduplicator.Version = nil
							})

							It("returns an error when update data set returns an error", func() {
								responseErr := errorsTest.RandomError()
								dataSetRepository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: nil, Error: responseErr}}
								result, err := deduplicator.Open(ctx, dataSet)
								Expect(err).To(Equal(responseErr))
								Expect(result).To(BeNil())
							})

							It("returns successfully when update data set returns successfully", func() {
								responseDataSet := dataTest.RandomDataSet()
								dataSetRepository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: responseDataSet, Error: nil}}
								Expect(deduplicator.Open(ctx, dataSet)).To(Equal(responseDataSet))
							})
						})

						When("the data set has a deduplicator with matching name and version exists", func() {
							BeforeEach(func() {
								dataSet.Deduplicator.Version = pointer.FromString(netTest.RandomSemanticVersion())
							})

							It("returns an error when update data set returns an error", func() {
								responseErr := errorsTest.RandomError()
								dataSetRepository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: nil, Error: responseErr}}
								result, err := deduplicator.Open(ctx, dataSet)
								Expect(err).To(Equal(responseErr))
								Expect(result).To(BeNil())
							})

							It("returns successfully when update data set returns successfully", func() {
								responseDataSet := dataTest.RandomDataSet()
								dataSetRepository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: responseDataSet, Error: nil}}
								Expect(deduplicator.Open(ctx, dataSet)).To(Equal(responseDataSet))
							})
						})
					})
				})

				Context("AddData", func() {
					var dataSetData data.Data

					BeforeEach(func() {
						dataSetData = make(data.Data, test.RandomIntFromRange(0, 3))
						for index := range dataSetData {
							dataSetData[index] = dataTypesTest.RandomBase()
						}
					})

					It("returns an error when the context is missing", func() {
						Expect(deduplicator.AddData(nil, dataSet, dataSetData)).To(MatchError("context is missing"))
					})

					It("returns an error when the data set is missing", func() {
						Expect(deduplicator.AddData(ctx, nil, dataSetData)).To(MatchError("data set is missing"))
					})

					It("returns an error when the data set data is missing", func() {
						Expect(deduplicator.AddData(ctx, dataSet, nil)).To(MatchError("data set data is missing"))
					})

					When("create data set data is invoked", func() {
						AfterEach(func() {
							Expect(dataRepository.CreateDataSetDataInputs).To(Equal([]dataStoreTest.CreateDataSetDataInput{{Context: ctx, DataSet: dataSet, DataSetData: dataSetData}}))
						})

						It("returns an error when create data set data returns an error", func() {
							responseErr := errorsTest.RandomError()
							dataRepository.CreateDataSetDataOutputs = []error{responseErr}
							Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(Equal(responseErr))
						})

						It("returns successfully when create data set data returns successfully", func() {
							dataRepository.CreateDataSetDataOutputs = []error{nil}
							Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(Succeed())
						})
					})
				})

				Context("DeleteData", func() {
					var selectors *data.Selectors

					BeforeEach(func() {
						selectors = dataTest.RandomSelectors()
					})

					It("returns an error when the context is missing", func() {
						Expect(deduplicator.DeleteData(nil, dataSet, selectors)).To(MatchError("context is missing"))
					})

					It("returns an error when the data set is missing", func() {
						Expect(deduplicator.DeleteData(ctx, nil, selectors)).To(MatchError("data set is missing"))
					})

					It("returns an error when the selectors is missing", func() {
						Expect(deduplicator.DeleteData(ctx, dataSet, nil)).To(MatchError("selectors is missing"))
					})

					When("destroy data set data is invoked", func() {
						AfterEach(func() {
							Expect(dataRepository.DestroyDataSetDataInputs).To(Equal([]dataStoreTest.DestroyDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
						})

						It("returns an error when destroy data set data returns an error", func() {
							responseErr := errorsTest.RandomError()
							dataRepository.DestroyDataSetDataOutputs = []error{responseErr}
							Expect(deduplicator.DeleteData(ctx, dataSet, selectors)).To(Equal(responseErr))
						})

						It("returns successfully when destroy data set data returns successfully", func() {
							dataRepository.DestroyDataSetDataOutputs = []error{nil}
							Expect(deduplicator.DeleteData(ctx, dataSet, selectors)).To(Succeed())
						})
					})
				})

				Context("Close", func() {
					It("returns an error when the context is missing", func() {
						Expect(deduplicator.Close(nil, dataSet)).To(MatchError("context is missing"))
					})

					It("returns an error when the data set is missing", func() {
						Expect(deduplicator.Close(ctx, nil)).To(MatchError("data set is missing"))
					})

					When("delete other data set data is invoked", func() {
						AfterEach(func() {
							Expect(dataRepository.DeleteOtherDataSetDataInputs).To(Equal([]dataStoreTest.DeleteOtherDataSetDataInput{{Context: ctx, DataSet: dataSet}}))
						})

						It("returns an error when delete other data set data returns an error", func() {
							responseErr := errorsTest.RandomError()
							dataRepository.DeleteOtherDataSetDataOutputs = []error{responseErr}
							Expect(deduplicator.Close(ctx, dataSet)).To(Equal(responseErr))
						})

						When("UpdateDataSet is invoked", func() {
							BeforeEach(func() {
								dataRepository.DeleteOtherDataSetDataOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(dataSetRepository.UpdateDataSetInputs).To(Equal([]dataStoreTest.UpdateDataSetInput{{Context: ctx, ID: *dataSet.UploadID, Update: &data.DataSetUpdate{Active: pointer.FromBool(true)}}}))
							})

							It("returns an error when update data set data returns an error", func() {
								responseErr := errorsTest.RandomError()
								dataSetRepository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: nil, Error: responseErr}}
								Expect(deduplicator.Close(ctx, dataSet)).To(Equal(responseErr))
							})

							When("activate data set data is invoked", func() {
								BeforeEach(func() {
									dataSetRepository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: dataSet, Error: nil}}
								})

								AfterEach(func() {
									Expect(dataRepository.ActivateDataSetDataInputs).To(Equal([]dataStoreTest.ActivateDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: nil}}))
								})

								It("returns an error when active data set data returns an error", func() {
									responseErr := errorsTest.RandomError()
									dataRepository.ActivateDataSetDataOutputs = []error{responseErr}
									Expect(deduplicator.Close(ctx, dataSet)).To(Equal(responseErr))
								})

								It("returns successfully when active data set data returns successfully", func() {
									dataRepository.ActivateDataSetDataOutputs = []error{nil}
									Expect(deduplicator.Close(ctx, dataSet)).To(Succeed())
								})
							})
						})
					})
				})

				Context("Delete", func() {
					It("returns an error when the context is missing", func() {
						Expect(deduplicator.Delete(nil, dataSet)).To(MatchError("context is missing"))
					})

					It("returns an error when the data set is missing", func() {
						Expect(deduplicator.Delete(ctx, nil)).To(MatchError("data set is missing"))
					})

					When("delete data set is invoked", func() {
						AfterEach(func() {
							Expect(dataSetRepository.DeleteDataSetInputs).To(Equal([]dataStoreTest.DeleteDataSetInput{{Context: ctx, DataSet: dataSet}}))
						})

						It("returns an error when delete data set returns an error", func() {
							responseErr := errorsTest.RandomError()
							dataSetRepository.DeleteDataSetOutputs = []error{responseErr}
							Expect(deduplicator.Delete(ctx, dataSet)).To(Equal(responseErr))
						})

						It("returns successfully when delete data set returns successfully", func() {
							dataSetRepository.DeleteDataSetOutputs = []error{nil}
							Expect(deduplicator.Delete(ctx, dataSet)).To(Succeed())
						})
					})
				})
			})
		})
	})
})
