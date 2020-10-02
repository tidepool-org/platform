package deduplicator_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataStoreTest "github.com/tidepool-org/platform/data/store/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	dataTypesUploadTest "github.com/tidepool-org/platform/data/types/upload/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("DeviceTruncateDataSet", func() {
	It("DeviceTruncateDataSetName is expected", func() {
		Expect(dataDeduplicatorDeduplicator.DeviceTruncateDataSetName).To(Equal("org.tidepool.deduplicator.device.truncate.dataset"))
	})

	Context("NewDeviceTruncateDataSet", func() {
		It("returns succesfully", func() {
			Expect(dataDeduplicatorDeduplicator.NewDeviceTruncateDataSet()).ToNot(BeNil())
		})
	})

	Context("with new deduplicator", func() {
		var deduplicator *dataDeduplicatorDeduplicator.DeviceTruncateDataSet
		var dataSet *dataTypesUpload.Upload

		BeforeEach(func() {
			var err error
			deduplicator, err = dataDeduplicatorDeduplicator.NewDeviceTruncateDataSet()
			Expect(err).ToNot(HaveOccurred())
			Expect(deduplicator).ToNot(BeNil())
			dataSet = dataTypesUploadTest.RandomUpload()
			dataSet.DataSetType = pointer.FromString("normal")
			dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.deduplicator.device.truncate.dataset")
			dataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Animas"})
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

			dataSetTypeAssertions := func() {
				It("returns false when the deduplicator name does not match", func() {
					dataSet.Deduplicator.Name = pointer.FromString(netTest.RandomReverseDomain())
					Expect(deduplicator.New(dataSet)).To(BeFalse())
				})

				It("returns true when the deduplicator name matches", func() {
					Expect(deduplicator.New(dataSet)).To(BeTrue())
				})

				It("returns true when the deduplicator name matches deprecated", func() {
					dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.truncate")
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

					It("returns true when the device manufacturers matches", func() {
						Expect(deduplicator.New(dataSet)).To(BeTrue())
					})

					It("returns true when the device manufacturers matches with multiple", func() {
						dataSet.DeviceManufacturers = pointer.FromStringArray([]string{"Alpha", "Animas", "Bravo"})
						Expect(deduplicator.New(dataSet)).To(BeTrue())
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
				dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.truncate")
				Expect(deduplicator.Get(dataSet)).To(BeTrue())
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

				When("update data set is invoked", func() {
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
					dataSetData = make(data.Data, test.RandomIntFromRange(0, 3))
					for index := range dataSetData {
						dataSetData[index] = dataTypesTest.NewBase()
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

				When("delete other data set data is invoked", func() {
					AfterEach(func() {
						Expect(repository.DeleteOtherDataSetDataInputs).To(Equal([]dataStoreTest.DeleteOtherDataSetDataInput{{Context: ctx, DataSet: dataSet}}))
					})

					It("returns an error when delete other data set data returns an error", func() {
						responseErr := errorsTest.RandomError()
						repository.DeleteOtherDataSetDataOutputs = []error{responseErr}
						Expect(deduplicator.Close(ctx, repository, dataSet)).To(Equal(responseErr))
					})

					When("update data set is invoked", func() {
						BeforeEach(func() {
							repository.DeleteOtherDataSetDataOutputs = []error{nil}
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

				When("delete data set is invoked", func() {
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
