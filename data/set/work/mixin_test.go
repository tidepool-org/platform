package work_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/data"
	dataSetTest "github.com/tidepool-org/platform/data/set/test"
	dataSetWork "github.com/tidepool-org/platform/data/set/work"
	dataSetWorkTest "github.com/tidepool-org/platform/data/set/work/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("mixin", func() {
	Context("Metadata", func() {
		Context("MetadataKeyDataSetID", func() {
			It("returns expected value", func() {
				Expect(dataSetWork.MetadataKeyDataSetID).To(Equal("dataSetId"))
			})
		})

		Context("Metadata", func() {
			DescribeTable("serializes the datum as expected",
				func(mutator func(datum *dataSetWork.Metadata)) {
					datum := dataSetWorkTest.RandomMetadata(test.AllowOptionals())
					mutator(datum)
					test.ExpectSerializedObjectJSON(datum, dataSetWorkTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
					test.ExpectSerializedObjectBSON(datum, dataSetWorkTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
				},
				Entry("succeeds",
					func(datum *dataSetWork.Metadata) {},
				),
				Entry("empty",
					func(datum *dataSetWork.Metadata) {
						*datum = dataSetWork.Metadata{}
					},
				),
				Entry("all",
					func(datum *dataSetWork.Metadata) {
						datum.DataSetID = pointer.From(dataTest.RandomDataSetID())
					},
				),
			)

			Context("Parse", func() {
				DescribeTable("parses the datum",
					func(mutator func(object map[string]any, expectedDatum *dataSetWork.Metadata), expectedErrors ...error) {
						expectedDatum := dataSetWorkTest.RandomMetadata(test.AllowOptionals())
						object := dataSetWorkTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
						mutator(object, expectedDatum)
						result := &dataSetWork.Metadata{}
						errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
						Expect(result).To(Equal(expectedDatum))
					},
					Entry("succeeds",
						func(object map[string]any, expectedDatum *dataSetWork.Metadata) {},
					),
					Entry("empty",
						func(object map[string]any, expectedDatum *dataSetWork.Metadata) {
							clear(object)
							*expectedDatum = dataSetWork.Metadata{}
						},
					),
					Entry("multiple errors",
						func(object map[string]any, expectedDatum *dataSetWork.Metadata) {
							object["dataSetId"] = true
							expectedDatum.DataSetID = nil
						},
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/dataSetId"),
					),
				)
			})

			Context("Validate", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *dataSetWork.Metadata), expectedErrors ...error) {
						datum := dataSetWorkTest.RandomMetadata(test.AllowOptionals())
						mutator(datum)
						errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *dataSetWork.Metadata) {},
					),
					Entry("data set id missing",
						func(datum *dataSetWork.Metadata) {
							datum.DataSetID = nil
						},
					),
					Entry("data set id empty",
						func(datum *dataSetWork.Metadata) {
							datum.DataSetID = pointer.From("")
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetId"),
					),
					Entry("data set id invalid",
						func(datum *dataSetWork.Metadata) {
							datum.DataSetID = pointer.From("invalid")
						},
						errorsTest.WithPointerSource(data.ErrorValueStringAsSetIDNotValid("invalid"), "/dataSetId"),
					),
					Entry("data set id valid",
						func(datum *dataSetWork.Metadata) {
							datum.DataSetID = pointer.From(dataTest.RandomDataSetID())
						},
					),
					Entry("multiple errors",
						func(datum *dataSetWork.Metadata) {
							datum.DataSetID = pointer.From("")
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetId"),
					),
				)
			})
		})
	})

	Context("with context and mocks", func() {
		var mockLogger *logTest.Logger
		var mockController *gomock.Controller
		var mockWorkProvider *workTest.Provider
		var mockDataSetClient *dataSetTest.MockClient

		BeforeEach(func() {
			mockLogger = logTest.NewLogger()
			ctx := log.NewContextWithLogger(context.Background(), mockLogger)
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkProvider = workTest.NewProvider(ctx)
			mockDataSetClient = dataSetTest.NewMockClient(mockController)
		})

		Context("NewMixin", func() {
			It("returns an error when provider is missing", func() {
				mixin, err := dataSetWork.NewMixin(nil, mockDataSetClient)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when data set client is missing", func() {
				mixin, err := dataSetWork.NewMixin(mockWorkProvider, nil)
				Expect(err).To(MatchError("data set client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := dataSetWork.NewMixin(mockWorkProvider, mockDataSetClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("NewMixinFromWork", func() {
			var workMetadata *dataSetWork.Metadata

			BeforeEach(func() {
				workMetadata = dataSetWorkTest.RandomMetadata(test.AllowOptionals())
			})

			It("returns an error when provider is missing", func() {
				mixin, err := dataSetWork.NewMixinFromWork(nil, mockDataSetClient, workMetadata)
				Expect(err).To(MatchError("provider is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when data set client is missing", func() {
				mixin, err := dataSetWork.NewMixinFromWork(mockWorkProvider, nil, workMetadata)
				Expect(err).To(MatchError("data set client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns an error when work metadata is missing", func() {
				mixin, err := dataSetWork.NewMixinFromWork(mockWorkProvider, mockDataSetClient, nil)
				Expect(err).To(MatchError("work metadata is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns successfully", func() {
				mixin, err := dataSetWork.NewMixinFromWork(mockWorkProvider, mockDataSetClient, workMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("mixin", func() {
			var workMetadata *dataSetWork.Metadata
			var mixin dataSetWork.MixinFromWork

			BeforeEach(func() {
				var err error
				workMetadata = dataSetWorkTest.RandomMetadata(test.AllowOptionals())
				mixin, err = dataSetWork.NewMixinFromWork(mockWorkProvider, mockDataSetClient, workMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})

			Context("DataSetClient", func() {
				It("returns the data set client", func() {
					Expect(mixin.DataSetClient()).To(Equal(mockDataSetClient))
				})
			})

			Context("HasDataSet", func() {
				It("returns false initially", func() {
					Expect(mixin.HasDataSet()).To(BeFalse())
				})

				It("returns true after SetDataSet is called with a data set", func() {
					Expect(mixin.SetDataSet(dataTest.RandomDataSet(test.AllowOptionals()))).To(BeNil())
					Expect(mixin.HasDataSet()).To(BeTrue())
				})

				It("returns false after SetDataSet is called with nil", func() {
					Expect(mixin.SetDataSet(dataTest.RandomDataSet(test.AllowOptionals()))).To(BeNil())
					Expect(mixin.HasDataSet()).To(BeTrue())
					Expect(mixin.SetDataSet(nil)).To(BeNil())
					Expect(mixin.HasDataSet()).To(BeFalse())
				})
			})

			Context("DataSet", func() {
				It("returns nil initially", func() {
					Expect(mixin.DataSet()).To(BeNil())
				})

				It("returns the data set after SetDataSet is called with a data set", func() {
					dataSt := dataTest.RandomDataSet(test.AllowOptionals())
					Expect(mixin.SetDataSet(dataSt)).To(BeNil())
					Expect(mixin.DataSet()).To(Equal(dataSt))
				})

				It("returns nil after SetDataSet is called with nil", func() {
					Expect(mixin.SetDataSet(dataTest.RandomDataSet(test.AllowOptionals()))).To(BeNil())
					Expect(mixin.SetDataSet(nil)).To(BeNil())
					Expect(mixin.DataSet()).To(BeNil())
				})
			})

			Context("SetDataSet", func() {
				It("decodes metadata from data set and returns nil", func() {
					dataSt := dataTest.RandomDataSet(test.AllowOptionals())
					Expect(mixin.SetDataSet(dataSt)).To(BeNil())
					Expect(mixin.DataSet()).To(Equal(dataSt))
				})

				It("clears metadata when data set is nil and returns nil", func() {
					Expect(mixin.SetDataSet(dataTest.RandomDataSet(test.AllowOptionals()))).To(BeNil())
					Expect(mixin.SetDataSet(nil)).To(BeNil())
					Expect(mixin.DataSet()).To(BeNil())
				})
			})

			Context("FetchDataSet", func() {
				var dataStID string

				BeforeEach(func() {
					dataStID = dataTest.RandomDataSetID()
				})

				It("returns failing process result when data set client returns an error", func() {
					testErr := errorsTest.RandomError()
					mockDataSetClient.EXPECT().GetDataSet(gomock.Not(gomock.Nil()), dataStID).Return(nil, testErr)
					Expect(mixin.FetchDataSet(dataStID)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to get data set").Error())))
				})

				It("returns failed process result when data set is nil", func() {
					mockDataSetClient.EXPECT().GetDataSet(gomock.Not(gomock.Nil()), dataStID).Return(nil, nil)
					Expect(mixin.FetchDataSet(dataStID)).To(workTest.MatchFailedProcessResultError(MatchError("data set is missing")))
				})

				It("sets the data set and returns nil on success", func() {
					dataSt := dataTest.RandomDataSet(test.AllowOptionals())
					mockDataSetClient.EXPECT().GetDataSet(gomock.Not(gomock.Nil()), dataStID).Return(dataSt, nil)
					Expect(mixin.FetchDataSet(dataStID)).To(BeNil())
					Expect(mixin.DataSet()).To(Equal(dataSt))
				})
			})

			Context("CreateDataSet", func() {
				var userID string
				var dataSetCreate *data.DataSetCreate

				BeforeEach(func() {
					userID = userTest.RandomUserID()
					dataSetCreate = dataTest.RandomDataSetCreate(test.AllowOptionals())
				})

				It("returns failed process result when data set is missing", func() {
					Expect(mixin.SetDataSet(dataTest.RandomDataSet(test.AllowOptionals()))).To(BeNil())
					Expect(mixin.CreateDataSet(userID, dataSetCreate)).To(workTest.MatchFailedProcessResultError(MatchError("data set already exists")))
				})

				Context("with an existing data set", func() {
					It("returns failing process result when the client returns an error", func() {
						testErr := errorsTest.RandomError()
						mockDataSetClient.EXPECT().CreateUserDataSet(gomock.Not(gomock.Nil()), userID, dataSetCreate).Return(nil, testErr)
						Expect(mixin.CreateDataSet(userID, dataSetCreate)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to create data set").Error())))
					})

					It("returns failed process result when the client returns a nil data set", func() {
						mockDataSetClient.EXPECT().CreateUserDataSet(gomock.Not(gomock.Nil()), userID, dataSetCreate).Return(nil, nil)
						Expect(mixin.CreateDataSet(userID, dataSetCreate)).To(workTest.MatchFailedProcessResultError(MatchError("data set is missing")))
					})

					It("updates the data set and returns nil on success", func() {
						createdDataSet := dataTest.RandomDataSet(test.AllowOptionals())
						mockDataSetClient.EXPECT().CreateUserDataSet(gomock.Not(gomock.Nil()), userID, dataSetCreate).Return(createdDataSet, nil)
						Expect(mixin.CreateDataSet(userID, dataSetCreate)).To(BeNil())
						Expect(mixin.DataSet()).To(Equal(createdDataSet))
					})
				})
			})

			Context("UpdateDataSet", func() {
				It("returns failed process result when data set is missing", func() {
					Expect(mixin.UpdateDataSet(&data.DataSetUpdate{})).To(workTest.MatchFailedProcessResultError(MatchError("data set is missing")))
				})

				Context("with an existing data set", func() {
					var dataSt *data.DataSet
					var dataStUpdate *data.DataSetUpdate

					BeforeEach(func() {
						dataSt = dataTest.RandomDataSet(test.AllowOptionals())
						Expect(mixin.SetDataSet(dataSt)).To(BeNil())
						dataStUpdate = dataTest.RandomDataSetUpdate(test.AllowOptionals())
					})

					It("returns failed process result when data set id is missing", func() {
						dataSt.ID = nil
						Expect(mixin.UpdateDataSet(dataStUpdate)).To(workTest.MatchFailedProcessResultError(MatchError("data set id is missing")))
					})

					It("returns failing process result when the client returns an error", func() {
						testErr := errorsTest.RandomError()
						mockDataSetClient.EXPECT().UpdateDataSet(gomock.Not(gomock.Nil()), *dataSt.ID, dataStUpdate).Return(nil, testErr)
						Expect(mixin.UpdateDataSet(dataStUpdate)).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to update data set").Error())))
					})

					It("returns failed process result when the client returns a nil data set", func() {
						mockDataSetClient.EXPECT().UpdateDataSet(gomock.Not(gomock.Nil()), *dataSt.ID, dataStUpdate).Return(nil, nil)
						Expect(mixin.UpdateDataSet(dataStUpdate)).To(workTest.MatchFailedProcessResultError(MatchError("data set is missing")))
					})

					It("updates the data set and returns nil on success", func() {
						updatedDataSet := dataTest.RandomDataSet(test.AllowOptionals())
						mockDataSetClient.EXPECT().UpdateDataSet(gomock.Not(gomock.Nil()), *dataSt.ID, dataStUpdate).Return(updatedDataSet, nil)
						Expect(mixin.UpdateDataSet(dataStUpdate)).To(BeNil())
						Expect(mixin.DataSet()).To(Equal(updatedDataSet))
					})
				})
			})

			Context("HasWorkMetadata", func() {
				It("returns false when work metadata is missing", func() {
					mixinWithoutMetadata, err := dataSetWork.NewMixin(mockWorkProvider, mockDataSetClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					var ok bool
					mixin, ok = mixinWithoutMetadata.(dataSetWork.MixinFromWork)
					Expect(ok).To(BeTrue())
					Expect(mixin.HasWorkMetadata()).To(BeFalse())
				})

				It("returns true when work metadata is present", func() {
					Expect(mixin.HasWorkMetadata()).To(BeTrue())
				})
			})

			Context("FetchDataSetFromWorkMetadata", func() {
				It("returns failed process result when work metadata is missing", func() {
					mixinWithoutMetadata, err := dataSetWork.NewMixin(mockWorkProvider, mockDataSetClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					var ok bool
					mixin, ok = mixinWithoutMetadata.(dataSetWork.MixinFromWork)
					Expect(ok).To(BeTrue())
					Expect(mixin.FetchDataSetFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata is missing")))
				})

				It("returns failed process result when work metadata data set id is missing", func() {
					workMetadata.DataSetID = nil
					Expect(mixin.FetchDataSetFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata data set id is missing")))
				})

				It("returns failing process result when the client returns an error", func() {
					workMetadata.DataSetID = pointer.From(dataTest.RandomDataSetID())
					testErr := errorsTest.RandomError()
					mockDataSetClient.EXPECT().GetDataSet(gomock.Not(gomock.Nil()), *workMetadata.DataSetID).Return(nil, testErr)
					Expect(mixin.FetchDataSetFromWorkMetadata()).To(workTest.MatchFailingProcessResultError(MatchError(errors.Wrap(testErr, "unable to get data set").Error())))
				})

				It("returns failed process result when the data set is nil", func() {
					workMetadata.DataSetID = pointer.From(dataTest.RandomDataSetID())
					mockDataSetClient.EXPECT().GetDataSet(gomock.Not(gomock.Nil()), *workMetadata.DataSetID).Return(nil, nil)
					Expect(mixin.FetchDataSetFromWorkMetadata()).To(workTest.MatchFailedProcessResultError(MatchError("data set is missing")))
				})

				It("sets the data set and returns nil on success", func() {
					workMetadata.DataSetID = pointer.From(dataTest.RandomDataSetID())
					dataSt := dataTest.RandomDataSet(test.AllowOptionals())
					mockDataSetClient.EXPECT().GetDataSet(gomock.Not(gomock.Nil()), *workMetadata.DataSetID).Return(dataSt, nil)
					Expect(mixin.FetchDataSetFromWorkMetadata()).To(BeNil())
					Expect(mixin.DataSet()).To(Equal(dataSt))
				})
			})

			Context("UpdateWorkMetadataFromDataSet", func() {
				It("returns failed process result when data set is missing", func() {
					Expect(mixin.UpdateWorkMetadataFromDataSet()).To(workTest.MatchFailedProcessResultError(MatchError("data set is missing")))
				})

				It("returns failed process result when data set id is missing", func() {
					dataSt := dataTest.RandomDataSet(test.AllowOptionals())
					dataSt.ID = nil
					Expect(mixin.SetDataSet(dataSt)).To(BeNil())
					Expect(mixin.UpdateWorkMetadataFromDataSet()).To(workTest.MatchFailedProcessResultError(MatchError("data set id is missing")))
				})

				It("returns failed process result when work metadata is missing", func() {
					mixinWithoutMetadata, err := dataSetWork.NewMixin(mockWorkProvider, mockDataSetClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(mixinWithoutMetadata).ToNot(BeNil())
					var ok bool
					mixin, ok = mixinWithoutMetadata.(dataSetWork.MixinFromWork)
					Expect(ok).To(BeTrue())
					dataSet := dataTest.RandomDataSet(test.AllowOptionals())
					Expect(mixin.SetDataSet(dataSet)).To(BeNil())
					Expect(mixin.UpdateWorkMetadataFromDataSet()).To(workTest.MatchFailedProcessResultError(MatchError("work metadata is missing")))
				})

				It("updates work metadata with the data set id and returns nil", func() {
					dataSt := dataTest.RandomDataSet(test.AllowOptionals())
					Expect(mixin.SetDataSet(dataSt)).To(BeNil())
					Expect(mixin.UpdateWorkMetadataFromDataSet()).To(BeNil())
					Expect(workMetadata.DataSetID).To(Equal(dataSt.ID))
				})
			})

			Context("AddDataSetToContext", func() {
				It("adds nil fields to context", func() {
					mixin.AddDataSetToContext()
					Expect(mockWorkProvider.Fields).To(Equal(log.Fields{"dataSet": log.Fields(nil)}))
				})

				It("adds non-nil fields to context", func() {
					dataSt := dataTest.RandomDataSet(test.AllowOptionals())
					Expect(mixin.SetDataSet(dataSt)).To(BeNil())
					Expect(mockWorkProvider.Fields).To(Equal(log.Fields{
						"dataSet": log.Fields{
							"id":     dataSt.ID,
							"userId": dataSt.UserID,
						},
					}))
				})
			})
		})
	})
})
