package deduplicator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"
	"fmt"

	"github.com/tidepool-org/platform/data"
	dataDeduplicator "github.com/tidepool-org/platform/data/deduplicator"
	dataStoreDEPRECATEDTest "github.com/tidepool-org/platform/data/storeDEPRECATED/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("ContinuousOrigin", func() {
	Context("NewContinuousOriginFactory", func() {
		It("returns a new factory", func() {
			Expect(dataDeduplicator.NewContinuousOriginFactory()).ToNot(BeNil())
		})
	})

	Context("with a new factory", func() {
		var factory dataDeduplicator.Factory
		var dataSetID string
		var userID string
		var dataSet *dataTypesUpload.Upload

		BeforeEach(func() {
			var err error
			factory, err = dataDeduplicator.NewContinuousOriginFactory()
			Expect(err).ToNot(HaveOccurred())
			Expect(factory).ToNot(BeNil())
			dataSetID = dataTest.RandomSetID()
			userID = userTest.RandomID()
			dataSet = dataTypesUpload.New()
			Expect(dataSet).ToNot(BeNil())
			dataSet.DataSetType = pointer.FromString(dataTypesUpload.DataSetTypeContinuous)
			dataSet.Deduplicator = &data.DeduplicatorDescriptor{
				Name: "org.tidepool.continuous.origin",
			}
			dataSet.UploadID = pointer.FromString(dataSetID)
			dataSet.UserID = pointer.FromString(userID)
		})

		Context("CanDeduplicateDataSet", func() {
			It("returns an error if the data set is missing", func() {
				can, err := factory.CanDeduplicateDataSet(nil)
				Expect(err).To(MatchError("data set is missing"))
				Expect(can).To(BeFalse())
			})

			It("returns false if the data set id is missing", func() {
				dataSet.UploadID = nil
				Expect(factory.CanDeduplicateDataSet(dataSet)).To(BeFalse())
			})

			It("returns false if the data set id is empty", func() {
				dataSet.UploadID = pointer.FromString("")
				Expect(factory.CanDeduplicateDataSet(dataSet)).To(BeFalse())
			})

			It("returns false if the data set user id is missing", func() {
				dataSet.UserID = nil
				Expect(factory.CanDeduplicateDataSet(dataSet)).To(BeFalse())
			})

			It("returns false if the data set user id is empty", func() {
				dataSet.UserID = pointer.FromString("")
				Expect(factory.CanDeduplicateDataSet(dataSet)).To(BeFalse())
			})

			It("returns false if the deduplicator is missing", func() {
				dataSet.Deduplicator = nil
				Expect(factory.CanDeduplicateDataSet(dataSet)).To(BeFalse())
			})

			It("returns false if the deduplicator name is not matching", func() {
				dataSet.Deduplicator.Name = "not-matching"
				Expect(factory.CanDeduplicateDataSet(dataSet)).To(BeFalse())
			})

			It("returns false if the data set type is missing", func() {
				dataSet.DataSetType = nil
				Expect(factory.CanDeduplicateDataSet(dataSet)).To(BeFalse())
			})

			It("returns false if the data set type is not continuous", func() {
				dataSet.DataSetType = pointer.FromString(dataTypesUpload.DataSetTypeNormal)
				Expect(factory.CanDeduplicateDataSet(dataSet)).To(BeFalse())
			})

			It("returns true if successful", func() {
				Expect(factory.CanDeduplicateDataSet(dataSet)).To(BeTrue())
			})
		})

		Context("with logger and data store session", func() {
			var logger log.Logger
			var dataSession *dataStoreDEPRECATEDTest.DataSession

			BeforeEach(func() {
				logger = logTest.NewLogger()
				dataSession = dataStoreDEPRECATEDTest.NewDataSession()
				Expect(dataSession).ToNot(BeNil())
			})

			AfterEach(func() {
				dataSession.Expectations()
			})

			Context("NewDeduplicatorForDataSet", func() {
				It("returns an error if the logger is missing", func() {
					deduplicator, err := factory.NewDeduplicatorForDataSet(nil, dataSession, dataSet)
					Expect(err).To(MatchError("logger is missing"))
					Expect(deduplicator).To(BeNil())
				})

				It("returns an error if the data store session is missing", func() {
					deduplicator, err := factory.NewDeduplicatorForDataSet(logger, nil, dataSet)
					Expect(err).To(MatchError("data store session is missing"))
					Expect(deduplicator).To(BeNil())
				})

				It("returns an error if the data set is missing", func() {
					deduplicator, err := factory.NewDeduplicatorForDataSet(logger, dataSession, nil)
					Expect(err).To(MatchError("data set is missing"))
					Expect(deduplicator).To(BeNil())
				})

				It("returns an error if the data set id is missing", func() {
					dataSet.UploadID = nil
					deduplicator, err := factory.NewDeduplicatorForDataSet(logger, dataSession, dataSet)
					Expect(err).To(MatchError("data set id is missing"))
					Expect(deduplicator).To(BeNil())
				})

				It("returns an error if the data set id is empty", func() {
					dataSet.UploadID = pointer.FromString("")
					deduplicator, err := factory.NewDeduplicatorForDataSet(logger, dataSession, dataSet)
					Expect(err).To(MatchError("data set id is empty"))
					Expect(deduplicator).To(BeNil())
				})

				It("returns an error if the data set user id is missing", func() {
					dataSet.UserID = nil
					deduplicator, err := factory.NewDeduplicatorForDataSet(logger, dataSession, dataSet)
					Expect(err).To(MatchError("data set user id is missing"))
					Expect(deduplicator).To(BeNil())
				})

				It("returns an error if the data set user id is empty", func() {
					dataSet.UserID = pointer.FromString("")
					deduplicator, err := factory.NewDeduplicatorForDataSet(logger, dataSession, dataSet)
					Expect(err).To(MatchError("data set user id is empty"))
					Expect(deduplicator).To(BeNil())
				})

				It("returns an error if the deduplicator is missing", func() {
					dataSet.Deduplicator = nil
					deduplicator, err := factory.NewDeduplicatorForDataSet(logger, dataSession, dataSet)
					Expect(err).To(MatchError("data set deduplicator is missing"))
					Expect(deduplicator).To(BeNil())
				})

				It("returns an error if the deduplicator name is not matching", func() {
					dataSet.Deduplicator.Name = "not-matching"
					deduplicator, err := factory.NewDeduplicatorForDataSet(logger, dataSession, dataSet)
					Expect(err).To(MatchError("data set is not registered with deduplicator"))
					Expect(deduplicator).To(BeNil())
				})

				It("returns an error if the data set type is missing", func() {
					dataSet.DataSetType = nil
					deduplicator, err := factory.NewDeduplicatorForDataSet(logger, dataSession, dataSet)
					Expect(err).To(MatchError("data set type is missing"))
					Expect(deduplicator).To(BeNil())
				})

				It("returns an error if the data set type is not continuous", func() {
					dataSet.DataSetType = pointer.FromString(dataTypesUpload.DataSetTypeNormal)
					deduplicator, err := factory.NewDeduplicatorForDataSet(logger, dataSession, dataSet)
					Expect(err).To(MatchError("data set type is not continuous"))
					Expect(deduplicator).To(BeNil())
				})

				It("returns successfully", func() {
					Expect(factory.NewDeduplicatorForDataSet(logger, dataSession, dataSet)).ToNot(BeNil())
				})
			})

			Context("with a context and new deduplicator", func() {
				var ctx context.Context
				var deduplicator data.Deduplicator
				var originIDs []string
				var dataSetData []data.Datum

				BeforeEach(func() {
					ctx = context.Background()
					var err error
					deduplicator, err = factory.NewDeduplicatorForDataSet(logger, dataSession, dataSet)
					Expect(err).ToNot(HaveOccurred())
					Expect(deduplicator).ToNot(BeNil())
					originIDs = []string{}
					dataSetData = []data.Datum{}
					for i := 0; i < 3; i++ {
						baseDatum := dataTypesTest.NewBase()
						baseDatum.Active = false
						originIDs = append(originIDs, *baseDatum.Origin.ID)
						dataSetData = append(dataSetData, baseDatum)
					}
				})

				Context("AddDataSetData", func() {
					It("returns successfully if the data is nil", func() {
						Expect(deduplicator.AddDataSetData(ctx, nil)).To(Succeed())
					})

					It("returns successfully if there is no data", func() {
						Expect(deduplicator.AddDataSetData(ctx, []data.Datum{})).To(Succeed())
					})

					Context("when archive data set data is invoked", func() {
						AfterEach(func() {
							Expect(dataSession.ArchiveDataSetDataUsingOriginIDsInputs).To(ConsistOf(dataStoreDEPRECATEDTest.ArchiveDataSetDataUsingOriginIDsInput{Context: ctx, DataSet: dataSet, OriginIDs: originIDs}))
						})

						It("returns an error if archive data set data using origin ids returns an error", func() {
							err := errorsTest.RandomError()
							dataSession.ArchiveDataSetDataUsingOriginIDsOutputs = []error{err}
							Expect(deduplicator.AddDataSetData(ctx, dataSetData)).To(MatchError(fmt.Sprintf("unable to archive device data using origin from data set with id %q; %s", *dataSet.UploadID, err)))

						})

						Context("archive data set data returns successfully and create data set data is invoked", func() {
							BeforeEach(func() {
								dataSession.ArchiveDataSetDataUsingOriginIDsOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(dataSession.CreateDataSetDataInputs).To(ConsistOf(dataStoreDEPRECATEDTest.CreateDataSetDataInput{Context: ctx, DataSet: dataSet, DataSetData: dataSetData}))
							})

							It("returns an error if create data set data returns an error", func() {
								err := errorsTest.RandomError()
								dataSession.CreateDataSetDataOutputs = []error{err}
								Expect(deduplicator.AddDataSetData(ctx, dataSetData)).To(MatchError(fmt.Sprintf("unable to create data set data with id %q; %s", *dataSet.UploadID, err)))
							})

							Context("create data set data returns successfully and delete archived data set data is invoked", func() {
								BeforeEach(func() {
									dataSession.CreateDataSetDataOutputs = []error{nil}
								})

								AfterEach(func() {
									Expect(dataSession.DeleteArchivedDataSetDataInputs).To(ConsistOf(dataStoreDEPRECATEDTest.DeleteArchivedDataSetDataInput{Context: ctx, DataSet: dataSet}))
								})

								It("returns an error if delete archive data set data returns an error", func() {
									err := errorsTest.RandomError()
									dataSession.DeleteArchivedDataSetDataOutputs = []error{err}
									Expect(deduplicator.AddDataSetData(ctx, dataSetData)).To(MatchError(fmt.Sprintf("unable to delete archived device data from data set with id %q; %s", *dataSet.UploadID, err)))
								})

								Context("delete archived data set data returns successfully", func() {
									BeforeEach(func() {
										dataSession.DeleteArchivedDataSetDataOutputs = []error{nil}
									})

									It("returns successfully", func() {
										Expect(deduplicator.AddDataSetData(ctx, dataSetData)).To(Succeed())
									})

									It("marks the data set data as active", func() {
										Expect(deduplicator.AddDataSetData(ctx, dataSetData)).To(Succeed())
										for _, datum := range dataSetData {
											baseDatum, _ := datum.(*dataTypes.Base)
											Expect(baseDatum.Active).To(BeTrue())
										}
									})
								})
							})
						})
					})
				})
			})
		})
	})
})
