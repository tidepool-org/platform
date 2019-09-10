package deduplicator_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataStoreDEPRECATEDTest "github.com/tidepool-org/platform/data/storeDEPRECATED/test"
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
					dataSetData = make(data.Data, test.RandomIntFromRange(0, 3))
					for index := range dataSetData {
						dataSetData[index] = dataTypesTest.NewBase()
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

				When("destroy data set data is invoked", func() {
					AfterEach(func() {
						Expect(session.DestroyDataSetDataInputs).To(Equal([]dataStoreDEPRECATEDTest.DestroyDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
					})

					It("returns an error when destroy data set data returns an error", func() {
						responseErr := errorsTest.RandomError()
						session.DestroyDataSetDataOutputs = []error{responseErr}
						Expect(deduplicator.DeleteData(ctx, session, dataSet, selectors)).To(Equal(responseErr))
					})

					It("returns successfully when destroy data set data returns successfully", func() {
						session.DestroyDataSetDataOutputs = []error{nil}
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

				When("delete other data set data is invoked", func() {
					AfterEach(func() {
						Expect(session.DeleteOtherDataSetDataInputs).To(Equal([]dataStoreDEPRECATEDTest.DeleteOtherDataSetDataInput{{Context: ctx, DataSet: dataSet}}))
					})

					It("returns an error when delete other data set data returns an error", func() {
						responseErr := errorsTest.RandomError()
						session.DeleteOtherDataSetDataOutputs = []error{responseErr}
						Expect(deduplicator.Close(ctx, session, dataSet)).To(Equal(responseErr))
					})

					When("update data set is invoked", func() {
						BeforeEach(func() {
							session.DeleteOtherDataSetDataOutputs = []error{nil}
						})

						AfterEach(func() {
							Expect(session.UpdateDataSetInputs).To(Equal([]dataStoreDEPRECATEDTest.UpdateDataSetInput{{Context: ctx, ID: *dataSet.UploadID, Update: &data.DataSetUpdate{Active: pointer.FromBool(true)}}}))
						})

						It("returns an error when update data set data returns an error", func() {
							responseErr := errorsTest.RandomError()
							session.UpdateDataSetOutputs = []dataStoreDEPRECATEDTest.UpdateDataSetOutput{{DataSet: nil, Error: responseErr}}
							Expect(deduplicator.Close(ctx, session, dataSet)).To(Equal(responseErr))
						})

						When("activate data set data is invoked", func() {
							BeforeEach(func() {
								session.UpdateDataSetOutputs = []dataStoreDEPRECATEDTest.UpdateDataSetOutput{{DataSet: dataSet, Error: nil}}
							})

							AfterEach(func() {
								Expect(session.ActivateDataSetDataInputs).To(Equal([]dataStoreDEPRECATEDTest.ActivateDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: nil}}))
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
			})

			Context("Delete", func() {
				It("returns an error when the context is missing", func() {
					Expect(deduplicator.Delete(nil, session, dataSet, false)).To(MatchError("context is missing"))
				})

				It("returns an error when the session is missing", func() {
					Expect(deduplicator.Delete(ctx, nil, dataSet, false)).To(MatchError("session is missing"))
				})

				It("returns an error when the data set is missing", func() {
					Expect(deduplicator.Delete(ctx, session, nil, false)).To(MatchError("data set is missing"))
				})

				When("delete data set is invoked", func() {
					AfterEach(func() {
						Expect(session.DeleteDataSetInputs).To(Equal([]dataStoreDEPRECATEDTest.DeleteDataSetInput{{Context: ctx, DataSet: dataSet}}))
					})

					It("returns an error when delete data set returns an error", func() {
						responseErr := errorsTest.RandomError()
						session.DeleteDataSetOutputs = []error{responseErr}
						Expect(deduplicator.Delete(ctx, session, dataSet, false)).To(Equal(responseErr))
					})

					It("returns successfully when delete data set returns successfully", func() {
						session.DeleteDataSetOutputs = []error{nil}
						Expect(deduplicator.Delete(ctx, session, dataSet, false)).To(Succeed())
					})
				})
			})
		})
	})
})
