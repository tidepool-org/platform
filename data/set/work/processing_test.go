package work_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	dataSetWork "github.com/tidepool-org/platform/data/set/work"
	dataSetWorkTest "github.com/tidepool-org/platform/data/set/work/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("Processor", func() {
	It("MetadataKeyID is expected", func() {
		Expect(dataSetWork.MetadataKeyID).To(Equal("dataSetId"))
	})

	Context("with base processing and client", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockClient *dataSetWorkTest.MockClient
		var baseProcessing *workBase.Processing

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockClient = dataSetWorkTest.NewMockClient(mockController)
			processResultBuilder := &workBase.ProcessResultBuilder{
				ProcessResultPendingBuilder: &workBase.ConstantProcessResultPendingBuilder{
					Duration: time.Minute,
				},
				ProcessResultFailingBuilder: &workBase.ConstantProcessResultFailingBuilder{
					Duration: time.Second,
				},
			}
			baseProcessing = workBase.NewProcessing(processResultBuilder)
		})

		Context("NewProcessing", func() {
			It("returns error if base processing is missing", func() {
				processing, err := dataSetWork.NewProcessing(nil, mockClient)
				Expect(err).To(MatchError("processing is missing"))
				Expect(processing).To(BeNil())
			})

			It("returns error if base processing is missing", func() {
				processing, err := dataSetWork.NewProcessing(baseProcessing, nil)
				Expect(err).To(MatchError("client is missing"))
				Expect(processing).To(BeNil())
			})

			It("returns processing success", func() {
				processing, err := dataSetWork.NewProcessing(baseProcessing, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(processing).ToNot(BeNil())
			})
		})

		Context("Processing", func() {
			var processing *dataSetWork.Processing
			var wrk *work.Work
			var mockProcessingUpdater *workTest.MockProcessingUpdater

			BeforeEach(func() {
				var err error
				processing, err = dataSetWork.NewProcessing(baseProcessing, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(processing).ToNot(BeNil())
				ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
				wrk = workTest.RandomWork()
				Expect(processing.Process(ctx, wrk, mockProcessingUpdater)()).To(BeNil())
			})

			Context("DataSetIDFromMetadata", func() {
				It("returns error if unable to parse", func() {
					wrk.Metadata[dataSetWork.MetadataKeyID] = true
					id, err := processing.DataSetIDFromMetadata()
					Expect(id).To(BeNil())
					Expect(err).To(MatchError("unable to parse data set id from metadata; type is not string, but bool"))
				})

				It("returns successfully", func() {
					expectedID := test.RandomString()
					wrk.Metadata[dataSetWork.MetadataKeyID] = expectedID
					id, err := processing.DataSetIDFromMetadata()
					Expect(err).ToNot(HaveOccurred())
					Expect(id).ToNot(BeNil())
					Expect(*id).To(Equal(expectedID))
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
						Get(gomock.Any(), id, nil).
						Return(nil, testErr).
						Times(1)
					processResult := processing.FetchDataSet(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to fetch data set; " + testErr.Error()))
					Expect(processing.DataSet).To(BeNil())
				})

				It("returns failed process result if client returns nil", func() {
					mockClient.EXPECT().
						Get(gomock.Any(), id, nil).
						Return(nil, nil).
						Times(1)
					processResult := processing.FetchDataSet(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data set is missing"))
					Expect(processing.DataSet).To(BeNil())
				})

				It("returns successfully", func() {
					expectedDataSet := dataTest.RandomDataSet()
					mockClient.EXPECT().
						Get(gomock.Any(), id, nil).
						Return(expectedDataSet, nil).
						Times(1)
					processResult := processing.FetchDataSet(id)
					Expect(processResult).To(BeNil())
					Expect(processing.DataSet).To(Equal(expectedDataSet))
				})
			})

			Context("UpdateDataSet", func() {
				It("returns failed process result if existing data set is missing", func() {
					dataSetUpdate := dataTest.RandomDataSetUpdate()
					processResult := processing.UpdateDataSet(*dataSetUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data set is missing"))
				})

				It("returns failing process result if client returns error", func() {
					existingDataSet := dataTest.RandomDataSet()
					processing.DataSet = existingDataSet
					dataSetUpdate := dataTest.RandomDataSetUpdate()
					testErr := errorsTest.RandomError()
					mockClient.EXPECT().
						Update(gomock.Any(), *existingDataSet.ID, nil, dataSetUpdate).
						Return(nil, testErr).
						Times(1)
					processResult := processing.UpdateDataSet(*dataSetUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to update data set; " + testErr.Error()))
					Expect(processing.DataSet).To(Equal(existingDataSet))
				})

				It("returns failed process result if client returns nil", func() {
					existingDataSet := dataTest.RandomDataSet()
					processing.DataSet = existingDataSet
					dataSetUpdate := dataTest.RandomDataSetUpdate()
					mockClient.EXPECT().
						Update(gomock.Any(), *existingDataSet.ID, nil, dataSetUpdate).
						Return(nil, nil).
						Times(1)
					processResult := processing.UpdateDataSet(*dataSetUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data set is missing"))
					Expect(processing.DataSet).To(Equal(existingDataSet))
				})

				It("returns successfully", func() {
					existingDataSet := dataTest.RandomDataSet()
					processing.DataSet = existingDataSet
					expectedDataSet := dataTest.RandomDataSet()
					dataSetUpdate := dataTest.RandomDataSetUpdate()
					mockClient.EXPECT().
						Update(gomock.Any(), *existingDataSet.ID, nil, dataSetUpdate).
						Return(expectedDataSet, nil).
						Times(1)
					processResult := processing.UpdateDataSet(*dataSetUpdate)
					Expect(processResult).To(BeNil())
					Expect(processing.DataSet).To(Equal(expectedDataSet))
				})
			})
		})
	})
})
