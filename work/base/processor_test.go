package base_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	gomock "go.uber.org/mock/gomock"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("Processor", func() {
	var mockController *gomock.Controller
	var mockWorkClient *workTest.MockClient
	var mockProcessResultBuilder *workTest.MockProcessResultBuilder
	var dependencies workBase.Dependencies

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockWorkClient = workTest.NewMockClient(mockController)
		mockProcessResultBuilder = workTest.NewMockProcessResultBuilder(mockController)
		dependencies = workBase.Dependencies{
			WorkClient: mockWorkClient,
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("NewProcessor", func() {
		It("returns error when dependencies is invalid", func() {
			dependencies.WorkClient = nil
			processor, err := workBase.NewProcessor(dependencies, mockProcessResultBuilder)
			Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
			Expect(processor).To(BeNil())
		})

		It("returns error when process result builder is missing", func() {
			processor, err := workBase.NewProcessor(dependencies, nil)
			Expect(err).To(MatchError("process result builder is missing"))
			Expect(processor).To(BeNil())
		})

		It("returns Processor when process result builder is valid", func() {
			processor, err := workBase.NewProcessor(dependencies, mockProcessResultBuilder)
			Expect(err).ToNot(HaveOccurred())
			Expect(processor).ToNot(BeNil())
		})
	})

	Context("with Processor", func() {
		var processor *workBase.Processor
		var ctx context.Context
		var wrk *work.Work
		var mockProcessingUpdater *workTest.MockProcessingUpdater

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			wrk = workTest.RandomWork()
			mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)
		})

		JustBeforeEach(func() {
			var err error
			processor, err = workBase.NewProcessor(dependencies, mockProcessResultBuilder)
			Expect(err).ToNot(HaveOccurred())
			Expect(processor).ToNot(BeNil())
		})

		Context("Process", func() {
			It("returns failed ProcessResult when context is missing", func() {
				result := processor.ProcessPipeline(nil, wrk, mockProcessingUpdater).Process(func() *work.ProcessResult { return nil })
				Expect(result).To(workTest.MatchFailedProcessResult(MatchAllFields(Fields{
					"FailedError": errorsTest.MatchSerialized(MatchError("context is missing")),
					"Metadata":    BeNil(),
				})))
			})

			It("returns failed ProcessResult when work is missing", func() {
				result := processor.ProcessPipeline(ctx, nil, mockProcessingUpdater).Process(func() *work.ProcessResult { return nil })
				Expect(result).To(workTest.MatchFailedProcessResult(MatchAllFields(Fields{
					"FailedError": errorsTest.MatchSerialized(MatchError("work is missing")),
					"Metadata":    BeNil(),
				})))
			})

			It("returns failed ProcessResult when processing updater is missing", func() {
				result := processor.ProcessPipeline(ctx, wrk, nil).Process(func() *work.ProcessResult { return nil })
				Expect(result).To(workTest.MatchFailedProcessResult(MatchAllFields(Fields{
					"FailedError": errorsTest.MatchSerialized(MatchError("processing updater is missing")),
					"Metadata":    BeNil(),
				})))
			})

			It("returns nil and sets context, work, and processing updater when all parameters are valid", func() {
				result := processor.ProcessPipeline(ctx, wrk, mockProcessingUpdater).Process(func() *work.ProcessResult { return nil })
				Expect(result).To(BeNil())
				Expect(processor.Context()).ToNot(BeNil())
				Expect(processor.Work()).To(Equal(wrk))
			})

		})

		Context("ProcessPipeline", func() {
			It("returns a function that calls Process", func() {
				processPipeline := processor.ProcessPipeline(ctx, wrk, mockProcessingUpdater)
				Expect(processPipeline).ToNot(BeNil())
				Expect(processPipeline.Process(func() *work.ProcessResult { return nil })).To(BeNil())
				Expect(processor.Work()).To(Equal(wrk))
			})
		})

		Context("after Process is called with valid parameters", func() {
			JustBeforeEach(func() {
				result := processor.ProcessPipeline(ctx, wrk, mockProcessingUpdater).Process(func() *work.ProcessResult { return nil })
				Expect(result).To(BeNil())
			})

			Context("Context", func() {
				It("returns the context", func() {
					Expect(processor.Context()).ToNot(BeNil())
				})
			})

			Context("AddFieldToContext", func() {
				It("adds field to context", func() {
					processor.AddFieldToContext("key1", "key2")
					Expect(processor.Context()).ToNot(BeNil())
				})
			})

			Context("AddFieldsToContext", func() {
				It("adds fields to context", func() {
					processor.AddFieldsToContext(log.Fields{"key1": "value1", "key2": "value2"})
					Expect(processor.Context()).ToNot(BeNil())
				})
			})

			Context("Logger", func() {
				It("returns logger from context", func() {
					Expect(processor.Logger()).ToNot(BeNil())
				})
			})

			Context("Work", func() {
				It("returns the work", func() {
					Expect(processor.Work()).To(Equal(wrk))
				})
			})

			Context("Metadata", func() {
				It("returns the work metadata", func() {
					Expect(processor.Metadata()).To(Equal(wrk.Metadata))
				})
			})

			Context("MetadataParser", func() {
				Context("without work metadata", func() {
					BeforeEach(func() {
						wrk.Metadata = nil
					})

					It("returns an object parser for for nil metadata", func() {
						parser := processor.MetadataParser()
						Expect(parser).ToNot(BeNil())
						Expect(parser.Exists()).To(BeFalse())
					})
				})
				It("returns an object parser for the metadata", func() {
					parser := processor.MetadataParser()
					Expect(parser).ToNot(BeNil())
					Expect(parser.Exists()).To(BeTrue())
				})
			})

			Context("ProcessingUpdate", func() {
				It("returns failing ProcessResult when processing update fails", func() {
					expectedErr := errorsTest.RandomError()
					mockProcessingUpdater.EXPECT().
						ProcessingUpdate(gomock.Any(), work.ProcessingUpdate{Metadata: wrk.Metadata}).
						Return(nil, expectedErr)

					expectedProcessResult := workTest.RandomFailingProcessResult()
					mockProcessResultBuilder.EXPECT().
						Failing(gomock.Any(), wrk, errorsTest.Matcher(fmt.Sprintf("unable to update work; %s", expectedErr.Error()))).
						Return(expectedProcessResult)

					Expect(processor.ProcessingUpdate()).To(Equal(expectedProcessResult))
				})

				It("returns failed ProcessResult when processing update returns nil work", func() {
					mockProcessingUpdater.EXPECT().
						ProcessingUpdate(gomock.Any(), work.ProcessingUpdate{Metadata: wrk.Metadata}).
						Return(nil, nil)

					expectedProcessResult := workTest.RandomFailedProcessResult()
					mockProcessResultBuilder.EXPECT().
						Failed(gomock.Any(), wrk, errorsTest.Matcher("work is missing")).
						Return(expectedProcessResult)

					Expect(processor.ProcessingUpdate()).To(Equal(expectedProcessResult))
				})

				It("returns nil and updates work when processing update succeeds", func() {
					expectedWork := workTest.RandomWork()
					mockProcessingUpdater.EXPECT().
						ProcessingUpdate(gomock.Any(), work.ProcessingUpdate{Metadata: wrk.Metadata}).
						Return(expectedWork, nil)

					Expect(processor.ProcessingUpdate()).To(BeNil())
					Expect(processor.Work()).To(Equal(expectedWork))
				})
			})

			Context("Pending", func() {
				It("calls process result builder Pending", func() {
					expectedProcessResult := workTest.RandomPendingProcessResult()
					mockProcessResultBuilder.EXPECT().
						Pending(gomock.Any(), wrk).
						Return(expectedProcessResult)

					Expect(processor.Pending()).To(Equal(expectedProcessResult))
				})
			})

			Context("Failing", func() {
				It("calls process result builder Failing", func() {
					expectedErr := errorsTest.RandomError()
					expectedProcessResult := workTest.RandomFailingProcessResult()
					mockProcessResultBuilder.EXPECT().
						Failing(gomock.Any(), wrk, expectedErr).
						Return(expectedProcessResult)

					Expect(processor.Failing(expectedErr)).To(Equal(expectedProcessResult))
				})
			})

			Context("Failed", func() {
				It("calls process result builder Failed", func() {
					expectedErr := errorsTest.RandomError()
					expectedProcessResult := workTest.RandomFailedProcessResult()
					mockProcessResultBuilder.EXPECT().
						Failed(gomock.Any(), wrk, expectedErr).
						Return(expectedProcessResult)

					Expect(processor.Failed(expectedErr)).To(Equal(expectedProcessResult))
				})
			})

			Context("Success", func() {
				It("calls process result builder Success", func() {
					expectedProcessResult := workTest.RandomSuccessProcessResult()
					mockProcessResultBuilder.EXPECT().
						Success(gomock.Any(), wrk).
						Return(expectedProcessResult)

					Expect(processor.Success()).To(Equal(expectedProcessResult))
				})
			})

			Context("Delete", func() {
				It("calls process result builder Delete", func() {
					expectedProcessResult := work.NewProcessResultDelete()
					mockProcessResultBuilder.EXPECT().
						Delete(gomock.Any(), wrk).
						Return(expectedProcessResult)

					Expect(processor.Delete()).To(Equal(expectedProcessResult))
				})
			})
		})
	})
})
