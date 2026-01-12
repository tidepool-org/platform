package work_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	authProviderSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	authProviderSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	authTest "github.com/tidepool-org/platform/auth/test"
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
		Expect(authProviderSessionWork.MetadataKeyID).To(Equal("providerSessionId"))
	})

	Context("with base processing and client", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockClient *authProviderSessionWorkTest.MockClient
		var baseProcessing *workBase.Processing

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockClient = authProviderSessionWorkTest.NewMockClient(mockController)
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
				processing, err := authProviderSessionWork.NewProcessing(nil, mockClient)
				Expect(err).To(MatchError("processing is missing"))
				Expect(processing).To(BeNil())
			})

			It("returns error if base processing is missing", func() {
				processing, err := authProviderSessionWork.NewProcessing(baseProcessing, nil)
				Expect(err).To(MatchError("client is missing"))
				Expect(processing).To(BeNil())
			})

			It("returns processing success", func() {
				processing, err := authProviderSessionWork.NewProcessing(baseProcessing, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(processing).ToNot(BeNil())
			})
		})

		Context("Processing", func() {
			var processing *authProviderSessionWork.Processing
			var wrk *work.Work
			var mockProcessingUpdater *workTest.MockProcessingUpdater

			BeforeEach(func() {
				var err error
				processing, err = authProviderSessionWork.NewProcessing(baseProcessing, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(processing).ToNot(BeNil())
				ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
				wrk = workTest.RandomWork()
				Expect(processing.Process(ctx, wrk, mockProcessingUpdater)()).To(BeNil())
			})

			Context("ProviderSessionIDFromMetadata", func() {
				It("returns error if unable to parse", func() {
					wrk.Metadata[authProviderSessionWork.MetadataKeyID] = true
					id, err := processing.ProviderSessionIDFromMetadata()
					Expect(id).To(BeNil())
					Expect(err).To(MatchError("unable to parse provider session id from metadata; type is not string, but bool"))
				})

				It("returns successfully", func() {
					expectedID := test.RandomString()
					wrk.Metadata[authProviderSessionWork.MetadataKeyID] = expectedID
					id, err := processing.ProviderSessionIDFromMetadata()
					Expect(err).ToNot(HaveOccurred())
					Expect(id).ToNot(BeNil())
					Expect(*id).To(Equal(expectedID))
				})
			})

			Context("FetchProviderSession", func() {
				var id string

				BeforeEach(func() {
					id = test.RandomString()
				})

				It("returns failing process result if client returns error", func() {
					testErr := errorsTest.RandomError()
					mockClient.EXPECT().
						GetProviderSession(gomock.Any(), id).
						Return(nil, testErr).
						Times(1)
					processResult := processing.FetchProviderSession(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to fetch provider session; " + testErr.Error()))
					Expect(processing.ProviderSession).To(BeNil())
				})

				It("returns failed process result if client returns nil", func() {
					mockClient.EXPECT().
						GetProviderSession(gomock.Any(), id).
						Return(nil, nil).
						Times(1)
					processResult := processing.FetchProviderSession(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("provider session is missing"))
					Expect(processing.ProviderSession).To(BeNil())
				})

				It("returns successfully", func() {
					expectedProviderSession := authTest.RandomProviderSession()
					mockClient.EXPECT().
						GetProviderSession(gomock.Any(), id).
						Return(expectedProviderSession, nil).
						Times(1)
					processResult := processing.FetchProviderSession(id)
					Expect(processResult).To(BeNil())
					Expect(processing.ProviderSession).To(Equal(expectedProviderSession))
				})
			})

			Context("UpdateProviderSession", func() {
				It("returns failed process result if existing provider session is missing", func() {
					providerSessionUpdate := authTest.RandomProviderSessionUpdate()
					processResult := processing.UpdateProviderSession(*providerSessionUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("provider session is missing"))
				})

				It("returns failing process result if client returns error", func() {
					existingProviderSession := authTest.RandomProviderSession()
					processing.ProviderSession = existingProviderSession
					providerSessionUpdate := authTest.RandomProviderSessionUpdate()
					testErr := errorsTest.RandomError()
					mockClient.EXPECT().
						UpdateProviderSession(gomock.Any(), existingProviderSession.ID, providerSessionUpdate).
						Return(nil, testErr).
						Times(1)
					processResult := processing.UpdateProviderSession(*providerSessionUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to update provider session; " + testErr.Error()))
					Expect(processing.ProviderSession).To(Equal(existingProviderSession))
				})

				It("returns failed process result if client returns nil", func() {
					existingProviderSession := authTest.RandomProviderSession()
					processing.ProviderSession = existingProviderSession
					providerSessionUpdate := authTest.RandomProviderSessionUpdate()
					mockClient.EXPECT().
						UpdateProviderSession(gomock.Any(), existingProviderSession.ID, providerSessionUpdate).
						Return(nil, nil).
						Times(1)
					processResult := processing.UpdateProviderSession(*providerSessionUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("provider session is missing"))
					Expect(processing.ProviderSession).To(Equal(existingProviderSession))
				})

				It("returns successfully", func() {
					existingProviderSession := authTest.RandomProviderSession()
					processing.ProviderSession = existingProviderSession
					expectedProviderSession := authTest.RandomProviderSession()
					providerSessionUpdate := authTest.RandomProviderSessionUpdate()
					mockClient.EXPECT().
						UpdateProviderSession(gomock.Any(), existingProviderSession.ID, providerSessionUpdate).
						Return(expectedProviderSession, nil).
						Times(1)
					processResult := processing.UpdateProviderSession(*providerSessionUpdate)
					Expect(processResult).To(BeNil())
					Expect(processing.ProviderSession).To(Equal(expectedProviderSession))
				})
			})
		})
	})
})
