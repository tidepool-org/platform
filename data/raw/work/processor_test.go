package work_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	dataRawTest "github.com/tidepool-org/platform/data/raw/test"
	dataRawWork "github.com/tidepool-org/platform/data/raw/work"
	dataRawWorkTest "github.com/tidepool-org/platform/data/raw/work/test"
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
		Expect(dataRawWork.MetadataKeyID).To(Equal("dataRawId"))
	})

	Context("with base processor and client", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockClient *dataRawWorkTest.MockClient
		var baseProcessor *workBase.Processor

		BeforeEach(func() {
			var err error
			ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockClient = dataRawWorkTest.NewMockClient(mockController)
			processResultBuilder := &workBase.ProcessResultBuilder{
				ProcessResultPendingBuilder: &workBase.ConstantProcessResultPendingBuilder{
					Duration: time.Minute,
				},
				ProcessResultFailingBuilder: &workBase.ConstantProcessResultFailingBuilder{
					Duration: time.Second,
				},
			}
			baseProcessor, err = workBase.NewProcessor(processResultBuilder)
			Expect(err).ToNot(HaveOccurred())
			Expect(baseProcessor).ToNot(BeNil())
		})

		Context("NewProcessor", func() {
			It("returns error if processor is missing", func() {
				processor, err := dataRawWork.NewProcessor(nil, mockClient)
				Expect(err).To(MatchError("processor is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns error if client is missing", func() {
				processor, err := dataRawWork.NewProcessor(baseProcessor, nil)
				Expect(err).To(MatchError("client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns processor success", func() {
				processor, err := dataRawWork.NewProcessor(baseProcessor, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})
		})

		Context("Processor", func() {
			var processor *dataRawWork.Processor
			var wrk *work.Work
			var mockProcessingUpdater *workTest.MockProcessingUpdater

			BeforeEach(func() {
				var err error
				processor, err = dataRawWork.NewProcessor(baseProcessor, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
				ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
				wrk = workTest.RandomWork()
				Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(BeNil())
			})

			Context("DataRawIDFromMetadata", func() {
				It("returns error if unable to parse", func() {
					wrk.Metadata[dataRawWork.MetadataKeyID] = true
					id, err := processor.DataRawIDFromMetadata()
					Expect(id).To(BeNil())
					Expect(err).To(MatchError("unable to parse data raw id from metadata; type is not string, but bool"))
				})

				It("returns successfully", func() {
					expectedID := test.RandomString()
					wrk.Metadata[dataRawWork.MetadataKeyID] = expectedID
					id, err := processor.DataRawIDFromMetadata()
					Expect(err).ToNot(HaveOccurred())
					Expect(id).ToNot(BeNil())
					Expect(*id).To(Equal(expectedID))
				})
			})

			Context("FetchDataRaw", func() {
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
					processResult := processor.FetchDataRaw(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to fetch data raw; " + testErr.Error()))
					Expect(processor.DataRaw).To(BeNil())
				})

				It("returns failed process result if client returns nil", func() {
					mockClient.EXPECT().
						Get(gomock.Any(), id, nil).
						Return(nil, nil).
						Times(1)
					processResult := processor.FetchDataRaw(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data raw is missing"))
					Expect(processor.DataRaw).To(BeNil())
				})

				It("returns successfully", func() {
					expectedDataRaw := dataRawTest.RandomRaw()
					mockClient.EXPECT().
						Get(gomock.Any(), id, nil).
						Return(expectedDataRaw, nil).
						Times(1)
					processResult := processor.FetchDataRaw(id)
					Expect(processResult).To(BeNil())
					Expect(processor.DataRaw).To(Equal(expectedDataRaw))
				})
			})

			Context("UpdateDataRaw", func() {
				It("returns failed process result if existing data raw is missing", func() {
					dataRawUpdate := dataRawTest.RandomUpdate()
					processResult := processor.UpdateDataRaw(*dataRawUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data raw is missing"))
				})

				It("returns failing process result if client returns error", func() {
					existingDataRaw := dataRawTest.RandomRaw()
					processor.DataRaw = existingDataRaw
					dataRawUpdate := dataRawTest.RandomUpdate()
					testErr := errorsTest.RandomError()
					mockClient.EXPECT().
						Update(gomock.Any(), existingDataRaw.ID, nil, dataRawUpdate).
						Return(nil, testErr).
						Times(1)
					processResult := processor.UpdateDataRaw(*dataRawUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to update data raw; " + testErr.Error()))
					Expect(processor.DataRaw).To(Equal(existingDataRaw))
				})

				It("returns failed process result if client returns nil", func() {
					existingDataRaw := dataRawTest.RandomRaw()
					processor.DataRaw = existingDataRaw
					dataRawUpdate := dataRawTest.RandomUpdate()
					mockClient.EXPECT().
						Update(gomock.Any(), existingDataRaw.ID, nil, dataRawUpdate).
						Return(nil, nil).
						Times(1)
					processResult := processor.UpdateDataRaw(*dataRawUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data raw is missing"))
					Expect(processor.DataRaw).To(Equal(existingDataRaw))
				})

				It("returns successfully", func() {
					existingDataRaw := dataRawTest.RandomRaw()
					processor.DataRaw = existingDataRaw
					expectedDataRaw := dataRawTest.RandomRaw()
					dataRawUpdate := dataRawTest.RandomUpdate()
					mockClient.EXPECT().
						Update(gomock.Any(), existingDataRaw.ID, nil, dataRawUpdate).
						Return(expectedDataRaw, nil).
						Times(1)
					processResult := processor.UpdateDataRaw(*dataRawUpdate)
					Expect(processResult).To(BeNil())
					Expect(processor.DataRaw).To(Equal(expectedDataRaw))
				})
			})
		})
	})
})
