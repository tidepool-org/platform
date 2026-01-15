package work_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	dataSetTest "github.com/tidepool-org/platform/data/set/test"
	dataSetWork "github.com/tidepool-org/platform/data/set/work"
	dataTest "github.com/tidepool-org/platform/data/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("Mixin", func() {
	It("MetadataKeyID is expected", func() {
		Expect(dataSetWork.MetadataKeyID).To(Equal("dataSetId"))
	})

	Context("with base processor and client", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockClient *dataSetTest.MockClient
		var processor *workBase.Processor

		BeforeEach(func() {
			var err error
			ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockClient = dataSetTest.NewMockClient(mockController)
			processResultBuilder := &workBase.ProcessResultBuilder{
				ProcessResultPendingBuilder: &workBase.ConstantProcessResultPendingBuilder{
					Duration: time.Minute,
				},
				ProcessResultFailingBuilder: &workBase.ConstantProcessResultFailingBuilder{
					Duration: time.Second,
				},
			}
			processor, err = workBase.NewProcessor(processResultBuilder)
			Expect(err).ToNot(HaveOccurred())
			Expect(processor).ToNot(BeNil())
		})

		Context("NewMixin", func() {
			It("returns error if processor is missing", func() {
				mixin, err := dataSetWork.NewMixin(nil, mockClient)
				Expect(err).To(MatchError("processor is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns error if client is missing", func() {
				mixin, err := dataSetWork.NewMixin(processor, nil)
				Expect(err).To(MatchError("client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns success", func() {
				mixin, err := dataSetWork.NewMixin(processor, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("Mixin", func() {
			var mixin *dataSetWork.Mixin
			var wrk *work.Work
			var mockProcessingUpdater *workTest.MockProcessingUpdater

			BeforeEach(func() {
				var err error
				mixin, err = dataSetWork.NewMixin(processor, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
				ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
				wrk = workTest.RandomWork()
				Expect(mixin.Process(ctx, wrk, mockProcessingUpdater)).To(BeNil())
			})

			Context("DataSetIDFromMetadata", func() {
				It("returns error if unable to parse", func() {
					wrk.Metadata[dataSetWork.MetadataKeyID] = true
					id, err := mixin.DataSetIDFromMetadata()
					Expect(id).To(BeNil())
					Expect(err).To(MatchError("unable to parse data set id from metadata; type is not string, but bool"))
				})

				It("returns successfully", func() {
					expectedID := test.RandomString()
					wrk.Metadata[dataSetWork.MetadataKeyID] = expectedID
					id, err := mixin.DataSetIDFromMetadata()
					Expect(err).ToNot(HaveOccurred())
					Expect(id).To(PointTo(Equal(expectedID)))
				})
			})

			Context("FetchDataSetFromMetadata", func() {
				It("returns failed process result if unable to parse id", func() {
					wrk.Metadata[dataSetWork.MetadataKeyID] = true
					processResult := mixin.FetchDataSetFromMetadata()
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("unable to get data set id from metadata; unable to parse data set id from metadata; type is not string, but bool"))
					Expect(mixin.DataSet).To(BeNil())
				})

				It("returns failed process result if id is missing", func() {
					wrk.Metadata[dataSetWork.MetadataKeyID] = nil
					processResult := mixin.FetchDataSetFromMetadata()
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("unable to get data set id from metadata"))
					Expect(mixin.DataSet).To(BeNil())
				})

				When("id is valid", func() {
					var id string

					BeforeEach(func() {
						id = test.RandomString()
						wrk.Metadata[dataSetWork.MetadataKeyID] = id
					})

					It("returns failing process result if client returns error", func() {
						testErr := errorsTest.RandomError()
						mockClient.EXPECT().
							GetDataSet(gomock.Any(), id).
							Return(nil, testErr).
							Times(1)
						processResult := mixin.FetchDataSetFromMetadata()
						Expect(processResult).ToNot(BeNil())
						Expect(processResult.Result).To(Equal(work.ResultFailing))
						Expect(processResult.FailingUpdate).ToNot(BeNil())
						Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to fetch data set; " + testErr.Error()))
						Expect(mixin.DataSet).To(BeNil())
					})

					It("returns failed process result if client returns nil", func() {
						mockClient.EXPECT().
							GetDataSet(gomock.Any(), id).
							Return(nil, nil).
							Times(1)
						processResult := mixin.FetchDataSetFromMetadata()
						Expect(processResult).ToNot(BeNil())
						Expect(processResult.Result).To(Equal(work.ResultFailed))
						Expect(processResult.FailedUpdate).ToNot(BeNil())
						Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data set is missing"))
						Expect(mixin.DataSet).To(BeNil())
					})

					It("returns successfully", func() {
						expectedDataSet := dataTest.RandomDataSet()
						mockClient.EXPECT().
							GetDataSet(gomock.Any(), id).
							Return(expectedDataSet, nil).
							Times(1)
						processResult := mixin.FetchDataSetFromMetadata()
						Expect(processResult).To(BeNil())
						Expect(mixin.DataSet).To(Equal(expectedDataSet))
					})
				})
			})

			Context("FetchDataSet", func() {
				var id string

				BeforeEach(func() {
					id = test.RandomString()
				})

				It("returns failing process result if client returns error", func() {
					testErr := errorsTest.RandomError()
					mockClient.EXPECT().
						GetDataSet(gomock.Any(), id).
						Return(nil, testErr).
						Times(1)
					processResult := mixin.FetchDataSet(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to fetch data set; " + testErr.Error()))
					Expect(mixin.DataSet).To(BeNil())
				})

				It("returns failed process result if client returns nil", func() {
					mockClient.EXPECT().
						GetDataSet(gomock.Any(), id).
						Return(nil, nil).
						Times(1)
					processResult := mixin.FetchDataSet(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data set is missing"))
					Expect(mixin.DataSet).To(BeNil())
				})

				It("returns successfully", func() {
					expectedDataSet := dataTest.RandomDataSet()
					mockClient.EXPECT().
						GetDataSet(gomock.Any(), id).
						Return(expectedDataSet, nil).
						Times(1)
					processResult := mixin.FetchDataSet(id)
					Expect(processResult).To(BeNil())
					Expect(mixin.DataSet).To(Equal(expectedDataSet))
				})
			})

			Context("UpdateDataSet", func() {
				It("returns failed process result if existing data set is missing", func() {
					dataSetUpdate := dataTest.RandomDataSetUpdate()
					processResult := mixin.UpdateDataSet(*dataSetUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data set is missing"))
				})

				It("returns failing process result if client returns error", func() {
					existingDataSet := dataTest.RandomDataSet()
					mixin.DataSet = existingDataSet
					dataSetUpdate := dataTest.RandomDataSetUpdate()
					testErr := errorsTest.RandomError()
					mockClient.EXPECT().
						UpdateDataSet(gomock.Any(), *existingDataSet.ID, dataSetUpdate).
						Return(nil, testErr).
						Times(1)
					processResult := mixin.UpdateDataSet(*dataSetUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to update data set; " + testErr.Error()))
					Expect(mixin.DataSet).To(Equal(existingDataSet))
				})

				It("returns failed process result if client returns nil", func() {
					existingDataSet := dataTest.RandomDataSet()
					mixin.DataSet = existingDataSet
					dataSetUpdate := dataTest.RandomDataSetUpdate()
					mockClient.EXPECT().
						UpdateDataSet(gomock.Any(), *existingDataSet.ID, dataSetUpdate).
						Return(nil, nil).
						Times(1)
					processResult := mixin.UpdateDataSet(*dataSetUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data set is missing"))
					Expect(mixin.DataSet).To(Equal(existingDataSet))
				})

				It("returns successfully", func() {
					existingDataSet := dataTest.RandomDataSet()
					mixin.DataSet = existingDataSet
					expectedDataSet := dataTest.RandomDataSet()
					dataSetUpdate := dataTest.RandomDataSetUpdate()
					mockClient.EXPECT().
						UpdateDataSet(gomock.Any(), *existingDataSet.ID, dataSetUpdate).
						Return(expectedDataSet, nil).
						Times(1)
					processResult := mixin.UpdateDataSet(*dataSetUpdate)
					Expect(processResult).To(BeNil())
					Expect(mixin.DataSet).To(Equal(expectedDataSet))
				})
			})
		})
	})
})
