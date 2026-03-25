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
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("processor", func() {
	Context("Dependencies", func() {
		It("returns error when work client is missing", func() {
			dependencies := workBase.Dependencies{}
			Expect(dependencies.Validate()).To(MatchError("work client is missing"))
		})

		It("returns successfully", func() {
			dependencies := workBase.Dependencies{
				WorkClient: workTest.NewMockClient(gomock.NewController(GinkgoT())),
			}
			Expect(dependencies.Validate()).To(Succeed())
		})
	})

	Context("with dependencies and process result builder", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockDependencies workBase.Dependencies
		var mockProcessResultBuilder *workTest.MockProcessResultBuilder

		BeforeEach(func() {
			mockController, ctx = gomock.WithContext(log.NewContextWithLogger(context.Background(), logTest.NewLogger()), GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockDependencies = workBase.Dependencies{
				WorkClient: mockWorkClient,
			}
			mockProcessResultBuilder = workTest.NewMockProcessResultBuilder(mockController)
		})

		Context("NewProcessorWithoutMetadata", func() {
			It("returns error when dependencies is invalid", func() {
				mockDependencies.WorkClient = nil
				processor, err := workBase.NewProcessorWithoutMetadata(mockDependencies, mockProcessResultBuilder)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns error when process result builder is missing", func() {
				processor, err := workBase.NewProcessorWithoutMetadata(mockDependencies, nil)
				Expect(err).To(MatchError("process result builder is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns successfully", func() {
				processor, err := workBase.NewProcessorWithoutMetadata(mockDependencies, mockProcessResultBuilder)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})
		})

		Context("NewProcessor", func() {
			It("returns error when dependencies is invalid", func() {
				mockDependencies.WorkClient = nil
				processor, err := workBase.NewProcessor[workTest.MockMetadata](mockDependencies, mockProcessResultBuilder)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns error when process result builder is missing", func() {
				processor, err := workBase.NewProcessor[workTest.MockMetadata](mockDependencies, nil)
				Expect(err).To(MatchError("process result builder is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns successfully", func() {
				processor, err := workBase.NewProcessor[workTest.MockMetadata](mockDependencies, mockProcessResultBuilder)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})
		})

		Context("with Processor", func() {
			var processor *workBase.Processor[workTest.MockMetadata]
			var mockMetadata *workTest.MockMetadata
			var wrk *work.Work
			var mockProcessingUpdater *workTest.MockProcessingUpdater

			BeforeEach(func() {
				mockMetadata = workTest.RandomMockMetadata(test.AllowOptional())
				wrk = workTest.RandomWork()
				wrk.Metadata = mockMetadata.AsObject()
				mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)
			})

			JustBeforeEach(func() {
				var err error
				processor, err = workBase.NewProcessor[workTest.MockMetadata](mockDependencies, mockProcessResultBuilder)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})

			Context("ProcessPipeline", func() {
				It("returns failed ProcessResult when context is missing", func() {
					result := processor.ProcessPipeline(context.Context(nil), wrk, mockProcessingUpdater).Process(func() *work.ProcessResult { return nil })
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

				It("returns failed ProcessResult when unable to decode metadata", func() {
					wrk.Metadata["mock"] = true
					expectedProcessResult := workTest.RandomFailedProcessResult()
					mockProcessResultBuilder.EXPECT().
						Failed(gomock.Any(), wrk, gomock.Any()).
						DoAndReturn(func(_ context.Context, _ *work.Work, err error) *work.ProcessResult {
							Expect(err).To(MatchError(ContainSubstring("unable to decode metadata")))
							return expectedProcessResult
						})
					result := processor.ProcessPipeline(ctx, wrk, mockProcessingUpdater).Process(func() *work.ProcessResult { return nil })
					Expect(result).To(Equal(expectedProcessResult))
				})

				It("returns nil and sets context, work, and processing updater when all parameters are valid", func() {
					expectedResult := workTest.RandomSuccessProcessResult()
					result := processor.ProcessPipeline(ctx, wrk, mockProcessingUpdater).Process(func() *work.ProcessResult { return expectedResult })
					Expect(result).To(Equal(expectedResult))
					Expect(processor.Context()).ToNot(BeNil())
				})

			})

			Context("WorkClient", func() {
				It("returns the work client", func() {
					Expect(processor.WorkClient()).To(Equal(mockWorkClient))
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

				Context("ProcessingUpdate", func() {
					It("returns failed ProcessResult when encoding metadata fails", func() {
						processor.Metadata().Any = func() {}
						expectedProcessResult := workTest.RandomFailedProcessResult()
						mockProcessResultBuilder.EXPECT().
							Failed(gomock.Any(), wrk, gomock.Any()).
							DoAndReturn(func(_ context.Context, _ *work.Work, err error) *work.ProcessResult {
								Expect(err).To(MatchError(ContainSubstring("unable to encode object")))
								return expectedProcessResult
							})

						Expect(processor.ProcessingUpdate()).To(Equal(expectedProcessResult))
					})

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
						expectedMockMetadata := workTest.RandomMockMetadata(test.AllowOptional())
						expectedWork := workTest.RandomWork()
						expectedWork.Metadata = expectedMockMetadata.AsObject()
						mockProcessingUpdater.EXPECT().
							ProcessingUpdate(gomock.Any(), work.ProcessingUpdate{Metadata: wrk.Metadata}).
							Return(expectedWork, nil)

						Expect(processor.ProcessingUpdate()).To(BeNil())
						Expect(processor.Metadata()).To(Equal(expectedMockMetadata))
					})
				})

				Context("Metadata", func() {
					It("returns the work metadata", func() {
						Expect(processor.Metadata()).To(Equal(mockMetadata))
					})
				})

				Context("Pending", func() {
					It("returns failed ProcessResult when encoding metadata fails", func() {
						processor.Metadata().Any = func() {}
						expectedProcessResult := workTest.RandomFailedProcessResult()
						mockProcessResultBuilder.EXPECT().
							Failed(gomock.Any(), wrk, gomock.Any()).
							DoAndReturn(func(_ context.Context, _ *work.Work, err error) *work.ProcessResult {
								Expect(err).To(MatchError(ContainSubstring("unable to encode object")))
								return expectedProcessResult
							})

						Expect(processor.Pending()).To(Equal(expectedProcessResult))
					})

					It("calls process result builder Pending", func() {
						expectedProcessResult := workTest.RandomPendingProcessResult()
						mockProcessResultBuilder.EXPECT().
							Pending(gomock.Any(), wrk).
							Return(expectedProcessResult)

						Expect(processor.Pending()).To(Equal(expectedProcessResult))
					})
				})

				Context("Failing", func() {
					It("returns failed ProcessResult when encoding metadata fails", func() {
						processor.Metadata().Any = func() {}
						expectedProcessResult := workTest.RandomFailedProcessResult()
						mockProcessResultBuilder.EXPECT().
							Failed(gomock.Any(), wrk, gomock.Any()).
							DoAndReturn(func(_ context.Context, _ *work.Work, err error) *work.ProcessResult {
								Expect(err).To(MatchError(ContainSubstring("unable to encode object")))
								return expectedProcessResult
							})

						Expect(processor.Failing(errorsTest.RandomError())).To(Equal(expectedProcessResult))
					})

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
					It("returns failed ProcessResult when encoding metadata fails", func() {
						processor.Metadata().Any = func() {}
						expectedProcessResult := workTest.RandomFailedProcessResult()
						mockProcessResultBuilder.EXPECT().
							Failed(gomock.Any(), wrk, gomock.Any()).
							DoAndReturn(func(_ context.Context, _ *work.Work, err error) *work.ProcessResult {
								Expect(err).To(MatchError(ContainSubstring("unable to encode object")))
								return expectedProcessResult
							})

						Expect(processor.Failed(errorsTest.RandomError())).To(Equal(expectedProcessResult))
					})

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
					It("returns failed ProcessResult when encoding metadata fails", func() {
						processor.Metadata().Any = func() {}
						expectedProcessResult := workTest.RandomFailedProcessResult()
						mockProcessResultBuilder.EXPECT().
							Failed(gomock.Any(), wrk, gomock.Any()).
							DoAndReturn(func(_ context.Context, _ *work.Work, err error) *work.ProcessResult {
								Expect(err).To(MatchError(ContainSubstring("unable to encode object")))
								return expectedProcessResult
							})

						Expect(processor.Success()).To(Equal(expectedProcessResult))
					})

					It("calls process result builder Success", func() {
						expectedProcessResult := workTest.RandomSuccessProcessResult()
						mockProcessResultBuilder.EXPECT().
							Success(gomock.Any(), wrk).
							Return(expectedProcessResult)

						Expect(processor.Success()).To(Equal(expectedProcessResult))
					})
				})

				Context("Delete", func() {
					It("returns failed ProcessResult when encoding metadata fails", func() {
						processor.Metadata().Any = func() {}
						expectedProcessResult := workTest.RandomFailedProcessResult()
						mockProcessResultBuilder.EXPECT().
							Failed(gomock.Any(), wrk, gomock.Any()).
							DoAndReturn(func(_ context.Context, _ *work.Work, err error) *work.ProcessResult {
								Expect(err).To(MatchError(ContainSubstring("unable to encode object")))
								return expectedProcessResult
							})

						Expect(processor.Delete()).To(Equal(expectedProcessResult))
					})

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
})
