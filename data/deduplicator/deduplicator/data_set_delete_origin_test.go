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

var _ = Describe("DataSetDeleteOrigin", func() {
	It("DataSetDeleteOriginName is expected", func() {
		Expect(dataDeduplicatorDeduplicator.DataSetDeleteOriginName).To(Equal("org.tidepool.deduplicator.dataset.delete.origin"))
	})

	Context("NewDataSetDeleteOrigin", func() {
		It("returns succesfully", func() {
			Expect(dataDeduplicatorDeduplicator.NewDataSetDeleteOrigin()).ToNot(BeNil())
		})
	})

	Context("with new deduplicator", func() {
		var deduplicator *dataDeduplicatorDeduplicator.DataSetDeleteOrigin
		var dataSet *dataTypesUpload.Upload

		BeforeEach(func() {
			var err error
			deduplicator, err = dataDeduplicatorDeduplicator.NewDataSetDeleteOrigin()
			Expect(err).ToNot(HaveOccurred())
			Expect(deduplicator).ToNot(BeNil())
			dataSet = dataTypesUploadTest.RandomUpload()
			dataSet.Deduplicator = data.NewDeduplicatorDescriptor()
			dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.deduplicator.dataset.delete.origin")
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

			It("returns true when the deduplicator name matches deprecated", func() {
				dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.continuous.origin")
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

			It("returns true when the deduplicator name matches deprecated", func() {
				dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.continuous.origin")
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
						update.Deduplicator = &data.DeduplicatorDescriptor{
							Name:    pointer.FromString("org.tidepool.deduplicator.dataset.delete.origin"),
							Version: pointer.FromString("1.0.0"),
						}
					})

					AfterEach(func() {
						Expect(repository.UpdateDataSetInputs).To(Equal([]dataStoreTest.UpdateDataSetInput{{Context: ctx, ID: *dataSet.UploadID, Update: update}}))
					})

					updateAssertions := func() {
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
								update.Deduplicator.Version = dataSet.Deduplicator.Version
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
				var dataSetData data.Data
				var selectors *data.Selectors

				BeforeEach(func() {
					dataSetData = make(data.Data, test.RandomIntFromRange(1, 3))
					selectors = data.NewSelectors()
					for index := range dataSetData {
						base := dataTypesTest.RandomBase()
						dataSetData[index] = base
						*selectors = append(*selectors, &data.Selector{Origin: &data.SelectorOrigin{ID: pointer.CloneString(base.Origin.ID)}})
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

				dataSetTypeAssertions := func() {
					originAssertions := func() {
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
					}

					When("data set data does not have an origin", func() {
						BeforeEach(func() {
							for index := range dataSetData {
								base := dataTypesTest.RandomBase()
								base.Origin = nil
								dataSetData[index] = base
							}
						})

						originAssertions()
					})

					When("data set data does not have an origin id", func() {
						BeforeEach(func() {
							for index := range dataSetData {
								base := dataTypesTest.RandomBase()
								base.Origin.ID = nil
								dataSetData[index] = base
							}
						})

						originAssertions()
					})

					When("data set data has an origin id", func() {
						When("delete data set data using origin ids is invoked", func() {
							AfterEach(func() {
								Expect(repository.DeleteDataSetDataInputs).To(Equal([]dataStoreTest.DeleteDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
							})

							It("returns an error when delete data set data using origin id returns an error", func() {
								responseErr := errorsTest.RandomError()
								repository.DeleteDataSetDataOutputs = []error{responseErr}
								Expect(deduplicator.AddData(ctx, repository, dataSet, dataSetData)).To(Equal(responseErr))
							})

							When("create data set data is invoked", func() {
								BeforeEach(func() {
									repository.DeleteDataSetDataOutputs = []error{nil}
								})

								AfterEach(func() {
									Expect(repository.CreateDataSetDataInputs).To(Equal([]dataStoreTest.CreateDataSetDataInput{{Context: ctx, DataSet: dataSet, DataSetData: dataSetData}}))
								})

								It("returns an error when create data set data returns an error", func() {
									responseErr := errorsTest.RandomError()
									repository.CreateDataSetDataOutputs = []error{responseErr}
									Expect(deduplicator.AddData(ctx, repository, dataSet, dataSetData)).To(Equal(responseErr))
								})

								When("destroy deleted data set data is invoked", func() {
									BeforeEach(func() {
										repository.CreateDataSetDataOutputs = []error{nil}
									})

									AfterEach(func() {
										Expect(repository.DestroyDeletedDataSetDataInputs).To(Equal([]dataStoreTest.DestroyDeletedDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
									})

									It("returns an error when destroy deleted data set data returns an error", func() {
										responseErr := errorsTest.RandomError()
										repository.DestroyDeletedDataSetDataOutputs = []error{responseErr}
										Expect(deduplicator.AddData(ctx, repository, dataSet, dataSetData)).To(Equal(responseErr))
									})

									It("returns successfully when destroy deleted data set data returns successfully", func() {
										repository.DestroyDeletedDataSetDataOutputs = []error{nil}
										Expect(deduplicator.AddData(ctx, repository, dataSet, dataSetData)).To(Succeed())
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
						for _, datum := range dataSetData {
							base, ok := datum.(*dataTypes.Base)
							Expect(ok).To(BeTrue())
							Expect(base).ToNot(BeNil())
							Expect(base.Active).To(BeFalse())
						}
					})

					dataSetTypeAssertions()
				})

				When("data set type is continuous", func() {
					BeforeEach(func() {
						dataSet.DataSetType = pointer.FromString("continuous")
					})

					AfterEach(func() {
						for _, datum := range dataSetData {
							base, ok := datum.(*dataTypes.Base)
							Expect(ok).To(BeTrue())
							Expect(base).ToNot(BeNil())
							Expect(base.Active).To(BeTrue())
						}
					})

					dataSetTypeAssertions()
				})

				When("data set type is normal", func() {
					BeforeEach(func() {
						dataSet.DataSetType = pointer.FromString("normal")
					})

					AfterEach(func() {
						for _, datum := range dataSetData {
							base, ok := datum.(*dataTypes.Base)
							Expect(ok).To(BeTrue())
							Expect(base).ToNot(BeNil())
							Expect(base.Active).To(BeFalse())
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

				When("archive data set data is invoked", func() {
					AfterEach(func() {
						Expect(repository.ArchiveDataSetDataInputs).To(Equal([]dataStoreTest.ArchiveDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
					})

					It("returns an error when archive data set data returns an error", func() {
						responseErr := errorsTest.RandomError()
						repository.ArchiveDataSetDataOutputs = []error{responseErr}
						Expect(deduplicator.DeleteData(ctx, repository, dataSet, selectors)).To(Equal(responseErr))
					})

					It("returns successfully when archive data set data returns successfully", func() {
						repository.ArchiveDataSetDataOutputs = []error{nil}
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

				When("data set type is continuous", func() {
					BeforeEach(func() {
						dataSet.DataSetType = pointer.FromString("continuous")
					})

					It("returns successfully", func() {
						Expect(deduplicator.Close(ctx, repository, dataSet)).To(Succeed())
					})
				})

				When("UpdateDataSet is invoked", func() {
					AfterEach(func() {
						Expect(repository.UpdateDataSetInputs).To(Equal([]dataStoreTest.UpdateDataSetInput{{Context: ctx, ID: *dataSet.UploadID, Update: &data.DataSetUpdate{Active: pointer.FromBool(true)}}}))
					})

					updateAssertions := func() {
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
