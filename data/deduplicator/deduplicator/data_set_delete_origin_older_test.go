package deduplicator_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataStoreTest "github.com/tidepool-org/platform/data/store/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("DataSetDeleteOriginOlder", func() {
	It("DataSetDeleteOriginOlderName is expected", func() {
		Expect(dataDeduplicatorDeduplicator.DataSetDeleteOriginOlderName).To(Equal("org.tidepool.deduplicator.dataset.delete.origin.older"))
	})

	It("DataSetDeleteOriginOlderVersion is expected", func() {
		Expect(dataDeduplicatorDeduplicator.DataSetDeleteOriginOlderVersion).To(Equal("1.0.0"))
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

		Context("NewDataSetDeleteOriginOlder", func() {
			It("returns successfully", func() {
				Expect(dataDeduplicatorDeduplicator.NewDataSetDeleteOriginOlder(dependencies)).ToNot(BeNil())
			})
		})

		Context("with new deduplicator", func() {
			var deduplicator *dataDeduplicatorDeduplicator.DataSetDeleteOriginOlder
			var dataSet *data.DataSet

			BeforeEach(func() {
				var err error
				deduplicator, err = dataDeduplicatorDeduplicator.NewDataSetDeleteOriginOlder(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(deduplicator).ToNot(BeNil())
				dataSet = dataTest.RandomDataSet()
				dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.deduplicator.dataset.delete.origin.older")
			})

			Context("New", func() {
				It("returns an error when the data set is missing", func() {
					found, err := deduplicator.New(context.Background(), nil)
					Expect(err).To(MatchError("data set is missing"))
					Expect(found).To(BeFalse())
				})

				It("returns false when the deduplicator is missing", func() {
					dataSet.Deduplicator = nil
					Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
				})

				It("returns false when the deduplicator name is missing", func() {
					dataSet.Deduplicator.Name = nil
					Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
				})

				It("returns false when the deduplicator name does not match", func() {
					dataSet.Deduplicator.Name = pointer.FromString(netTest.RandomReverseDomain())
					Expect(deduplicator.New(context.Background(), dataSet)).To(BeFalse())
				})

				It("returns true when the deduplicator name matches", func() {
					Expect(deduplicator.New(context.Background(), dataSet)).To(BeTrue())
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
							update.Deduplicator = &data.DeduplicatorDescriptor{
								Name:    pointer.FromString("org.tidepool.deduplicator.dataset.delete.origin.older"),
								Version: pointer.FromString("1.0.0"),
							}
						})

						AfterEach(func() {
							Expect(dataSetRepository.UpdateDataSetInputs).To(Equal([]dataStoreTest.UpdateDataSetInput{{Context: ctx, ID: *dataSet.UploadID, Update: update}}))
						})

						updateAssertions := func() {
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
						}

						When("data set type is not specified", func() {
							BeforeEach(func() {
								dataSet.DataSetType = nil
								update.Active = pointer.FromBool(false)
							})

							AfterEach(func() {
								Expect(dataSet.Active).To(BeFalse())
							})

							updateAssertions()
						})

						When("data set type is continuous", func() {
							BeforeEach(func() {
								dataSet.DataSetType = pointer.FromString("continuous")
								update.Active = pointer.FromBool(true)
							})

							AfterEach(func() {
								Expect(dataSet.Active).To(BeTrue())
							})

							updateAssertions()
						})

						When("data set type is normal", func() {
							BeforeEach(func() {
								dataSet.DataSetType = pointer.FromString("normal")
								update.Active = pointer.FromBool(false)
							})

							AfterEach(func() {
								Expect(dataSet.Active).To(BeFalse())
							})

							updateAssertions()
						})
					})
				})

				Context("AddData", func() {
					var allData data.Data
					var createdData data.Data
					var filteredSelectors *data.Selectors
					var deletedSelectors *data.Selectors
					var databaseSelectors *data.Selectors
					var expectedActive bool

					selectorFromBase := func(datum *dataTypes.Base) *data.Selector {
						return &data.Selector{Origin: &data.SelectorOrigin{ID: pointer.CloneString(datum.Origin.ID), Time: pointer.CloneString(datum.Origin.Time)}}
					}

					BeforeEach(func() {
						allData = data.Data{}
						createdData = data.Data{}
						filteredSelectors = data.NewSelectors()
						deletedSelectors = data.NewSelectors()
						databaseSelectors = data.NewSelectors()
						expectedActive = false

						// Not filtered
						for range test.RandomIntFromRange(1, 3) {
							base := dataTypesTest.RandomBase()

							allData = append(allData, base)
							createdData = append(createdData, base)

							selector := selectorFromBase(base)
							*deletedSelectors = append(*deletedSelectors, selector)
						}

						// Filtered, not in database
						for range test.RandomIntFromRange(1, 3) {
							base := dataTypesTest.RandomBase()
							base.Type = test.RandomStringFromArray([]string{"bolus", "food"})

							allData = append(allData, base)
							createdData = append(createdData, base)

							selector := selectorFromBase(base)
							*filteredSelectors = append(*filteredSelectors, selector)
							*deletedSelectors = append(*deletedSelectors, selector)
						}

						// Filtered, older in database
						for range test.RandomIntFromRange(1, 3) {
							base := dataTypesTest.RandomBase()
							base.Type = test.RandomStringFromArray([]string{"bolus", "food"})

							allData = append(allData, base)
							createdData = append(createdData, base)

							selector := selectorFromBase(base)
							*filteredSelectors = append(*filteredSelectors, selector)
							*deletedSelectors = append(*deletedSelectors, selector)

							originTime, err := time.Parse(time.RFC3339, *base.Origin.Time)
							Expect(err).ToNot(HaveOccurred())

							databaseSelector := dataTest.CloneSelector(selector)
							databaseSelector.Origin.Time = pointer.FromString(test.RandomTimeBefore(originTime).Format(time.RFC3339))
							*databaseSelectors = append(*databaseSelectors, databaseSelector)
						}

						// Filtered, newer in database
						for range test.RandomIntFromRange(1, 3) {
							base := dataTypesTest.RandomBase()
							base.Type = test.RandomStringFromArray([]string{"bolus", "food"})

							allData = append(allData, base)

							selector := selectorFromBase(base)
							*filteredSelectors = append(*filteredSelectors, selector)

							originTime, err := time.Parse(time.RFC3339, *base.Origin.Time)
							Expect(err).ToNot(HaveOccurred())

							databaseSelector := dataTest.CloneSelector(selector)
							databaseSelector.Origin.Time = pointer.FromString(test.RandomTimeAfter(originTime).Format(time.RFC3339))
							*databaseSelectors = append(*databaseSelectors, databaseSelector)
						}
					})

					It("returns an error when the context is missing", func() {
						Expect(deduplicator.AddData(nil, dataSet, allData)).To(MatchError("context is missing"))
					})

					It("returns an error when the data set is missing", func() {
						Expect(deduplicator.AddData(ctx, nil, allData)).To(MatchError("data set is missing"))
					})

					It("returns an error when the data set data is missing", func() {
						Expect(deduplicator.AddData(ctx, dataSet, nil)).To(MatchError("data set data is missing"))
					})

					dataSetTypeAssertions := func() {
						originAssertions := func() {
							When("create data set data is invoked", func() {
								AfterEach(func() {
									Expect(dataRepository.CreateDataSetDataInputs).To(Equal([]dataStoreTest.CreateDataSetDataInput{{Context: ctx, DataSet: dataSet, DataSetData: createdData}}))
								})

								It("returns an error when create data set data returns an error", func() {
									responseErr := errorsTest.RandomError()
									dataRepository.CreateDataSetDataOutputs = []error{responseErr}
									Expect(deduplicator.AddData(ctx, dataSet, allData)).To(Equal(responseErr))
								})

								It("returns successfully when create data set data returns successfully", func() {
									dataRepository.CreateDataSetDataOutputs = []error{nil}
									Expect(deduplicator.AddData(ctx, dataSet, allData)).To(Succeed())
								})
							})
						}

						When("data set data does not have an origin", func() {
							BeforeEach(func() {
								for index := range allData {
									base := dataTypesTest.RandomBase()
									base.Origin = nil
									allData[index] = base
								}
								createdData = allData
							})

							originAssertions()
						})

						When("data set data does not have an origin id", func() {
							BeforeEach(func() {
								for index := range allData {
									base := dataTypesTest.RandomBase()
									base.Origin.ID = nil
									allData[index] = base
								}
								createdData = allData
							})

							originAssertions()
						})

						When("data set data has an origin id", func() {
							When("existing data set data using origin ids is invoked", func() {
								AfterEach(func() {
									Expect(dataRepository.ExistingDataSetDataInputs).To(Equal([]dataStoreTest.ExistingDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: filteredSelectors}}))
								})

								It("returns an error when existing data set data using origin id returns an error", func() {
									expectedActive = false
									responseErr := errorsTest.RandomError()
									dataRepository.ExistingDataSetDataOutputs = []dataStoreTest.ExistingDataSetDataOutput{{Selectors: nil, Error: responseErr}}
									Expect(deduplicator.AddData(ctx, dataSet, allData)).To(Equal(responseErr))
								})

								When("delete data set data using origin ids is invoked", func() {
									BeforeEach(func() {
										dataRepository.ExistingDataSetDataOutputs = []dataStoreTest.ExistingDataSetDataOutput{{Selectors: databaseSelectors, Error: nil}}
									})

									AfterEach(func() {
										Expect(dataRepository.DeleteDataSetDataInputs).To(Equal([]dataStoreTest.DeleteDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: deletedSelectors}}))
									})

									It("returns an error when delete data set data using origin id returns an error", func() {
										expectedActive = false
										responseErr := errorsTest.RandomError()
										dataRepository.DeleteDataSetDataOutputs = []error{responseErr}
										Expect(deduplicator.AddData(ctx, dataSet, allData)).To(Equal(responseErr))
									})

									When("create data set data is invoked", func() {
										BeforeEach(func() {
											dataRepository.DeleteDataSetDataOutputs = []error{nil}
										})

										AfterEach(func() {
											Expect(dataRepository.CreateDataSetDataInputs).To(Equal([]dataStoreTest.CreateDataSetDataInput{{Context: ctx, DataSet: dataSet, DataSetData: createdData}}))
										})

										It("returns an error when create data set data returns an error", func() {
											responseErr := errorsTest.RandomError()
											dataRepository.CreateDataSetDataOutputs = []error{responseErr}
											Expect(deduplicator.AddData(ctx, dataSet, allData)).To(Equal(responseErr))
										})

										When("destroy deleted data set data is invoked", func() {
											BeforeEach(func() {
												dataRepository.CreateDataSetDataOutputs = []error{nil}
											})

											AfterEach(func() {
												Expect(dataRepository.DestroyDeletedDataSetDataInputs).To(Equal([]dataStoreTest.DestroyDeletedDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: deletedSelectors}}))
											})

											It("returns an error when destroy deleted data set data returns an error", func() {
												responseErr := errorsTest.RandomError()
												dataRepository.DestroyDeletedDataSetDataOutputs = []error{responseErr}
												Expect(deduplicator.AddData(ctx, dataSet, allData)).To(Equal(responseErr))
											})

											It("returns successfully when destroy deleted data set data returns successfully", func() {
												dataRepository.DestroyDeletedDataSetDataOutputs = []error{nil}
												Expect(deduplicator.AddData(ctx, dataSet, allData)).To(Succeed())
											})
										})
									})
								})
							})
						})
					}

					When("data set type is not specified", func() {
						BeforeEach(func() {
							dataSet.DataSetType = nil
						})

						AfterEach(func() {
							for _, datum := range createdData {
								base, ok := datum.(*dataTypes.Base)
								Expect(ok).To(BeTrue())
								Expect(base).ToNot(BeNil())
								Expect(base.Active).To(Equal(expectedActive))
							}
						})

						dataSetTypeAssertions()
					})

					When("data set type is continuous", func() {
						BeforeEach(func() {
							dataSet.DataSetType = pointer.FromString("continuous")
							expectedActive = true
						})

						AfterEach(func() {
							for _, datum := range createdData {
								base, ok := datum.(*dataTypes.Base)
								Expect(ok).To(BeTrue())
								Expect(base).ToNot(BeNil())
								Expect(base.Active).To(Equal(expectedActive))
							}
						})

						dataSetTypeAssertions()
					})

					When("data set type is normal", func() {
						BeforeEach(func() {
							dataSet.DataSetType = pointer.FromString("normal")
						})

						AfterEach(func() {
							for _, datum := range createdData {
								base, ok := datum.(*dataTypes.Base)
								Expect(ok).To(BeTrue())
								Expect(base).ToNot(BeNil())
								Expect(base.Active).To(Equal(expectedActive))
							}
						})

						dataSetTypeAssertions()
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

					When("archive data set data is invoked", func() {
						AfterEach(func() {
							Expect(dataRepository.ArchiveDataSetDataInputs).To(Equal([]dataStoreTest.ArchiveDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
						})

						It("returns an error when archive data set data returns an error", func() {
							responseErr := errorsTest.RandomError()
							dataRepository.ArchiveDataSetDataOutputs = []error{responseErr}
							Expect(deduplicator.DeleteData(ctx, dataSet, selectors)).To(Equal(responseErr))
						})

						It("returns successfully when archive data set data returns successfully", func() {
							dataRepository.ArchiveDataSetDataOutputs = []error{nil}
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

					When("data set type is continuous", func() {
						BeforeEach(func() {
							dataSet.DataSetType = pointer.FromString("continuous")
						})

						It("returns successfully", func() {
							Expect(deduplicator.Close(ctx, dataSet)).To(Succeed())
						})
					})

					When("UpdateDataSet is invoked", func() {
						AfterEach(func() {
							Expect(dataSetRepository.UpdateDataSetInputs).To(Equal([]dataStoreTest.UpdateDataSetInput{{Context: ctx, ID: *dataSet.UploadID, Update: &data.DataSetUpdate{Active: pointer.FromBool(true)}}}))
						})

						updateAssertions := func() {
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
						}

						When("data set type is not specified", func() {
							BeforeEach(func() {
								dataSet.DataSetType = nil
							})

							updateAssertions()
						})

						When("data set type is normal", func() {
							BeforeEach(func() {
								dataSet.DataSetType = pointer.FromString("normal")
							})

							updateAssertions()
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
