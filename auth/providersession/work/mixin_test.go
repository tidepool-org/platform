package work_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	authTest "github.com/tidepool-org/platform/auth/test"
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
		Expect(providerSessionWork.MetadataKeyID).To(Equal("providerSessionId"))
	})

	Context("with base processor and client", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockClient *providerSessionTest.MockClient
		var processor *workBase.Processor

		BeforeEach(func() {
			var err error
			ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockClient = providerSessionTest.NewMockClient(mockController)
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
				mixin, err := providerSessionWork.NewMixin(nil, mockClient)
				Expect(err).To(MatchError("processor is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns error if client is missing", func() {
				mixin, err := providerSessionWork.NewMixin(processor, nil)
				Expect(err).To(MatchError("client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns success", func() {
				mixin, err := providerSessionWork.NewMixin(processor, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("Mixin", func() {
			var mixin *providerSessionWork.Mixin
			var wrk *work.Work
			var mockProcessingUpdater *workTest.MockProcessingUpdater

			BeforeEach(func() {
				var err error
				mixin, err = providerSessionWork.NewMixin(processor, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
				ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
				wrk = workTest.RandomWork()
				Expect(mixin.Process(ctx, wrk, mockProcessingUpdater)).To(BeNil())
			})

			Context("ProviderSessionIDFromMetadata", func() {
				It("returns error if unable to parse", func() {
					wrk.Metadata[providerSessionWork.MetadataKeyID] = true
					id, err := mixin.ProviderSessionIDFromMetadata()
					Expect(id).To(BeNil())
					Expect(err).To(MatchError("unable to parse provider session id from metadata; type is not string, but bool"))
				})

				It("returns successfully", func() {
					expectedID := test.RandomString()
					wrk.Metadata[providerSessionWork.MetadataKeyID] = expectedID
					id, err := mixin.ProviderSessionIDFromMetadata()
					Expect(err).ToNot(HaveOccurred())
					Expect(id).To(PointTo(Equal(expectedID)))
				})
			})

			Context("FetchProviderSessionFromMetadata", func() {
				It("returns failed process result if unable to parse id", func() {
					wrk.Metadata[providerSessionWork.MetadataKeyID] = true
					processResult := mixin.FetchProviderSessionFromMetadata()
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("unable to get provider session id from metadata; unable to parse provider session id from metadata; type is not string, but bool"))
					Expect(mixin.ProviderSession).To(BeNil())
				})

				It("returns failed process result if id is missing", func() {
					wrk.Metadata[providerSessionWork.MetadataKeyID] = nil
					processResult := mixin.FetchProviderSessionFromMetadata()
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("unable to get provider session id from metadata"))
					Expect(mixin.ProviderSession).To(BeNil())
				})

				When("id is valid", func() {
					var id string

					BeforeEach(func() {
						id = test.RandomString()
						wrk.Metadata[providerSessionWork.MetadataKeyID] = id
					})

					It("returns failing process result if client returns error", func() {
						testErr := errorsTest.RandomError()
						mockClient.EXPECT().
							GetProviderSession(gomock.Any(), id).
							Return(nil, testErr).
							Times(1)
						processResult := mixin.FetchProviderSessionFromMetadata()
						Expect(processResult).ToNot(BeNil())
						Expect(processResult.Result).To(Equal(work.ResultFailing))
						Expect(processResult.FailingUpdate).ToNot(BeNil())
						Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to fetch provider session; " + testErr.Error()))
						Expect(mixin.ProviderSession).To(BeNil())
					})

					It("returns failed process result if client returns nil", func() {
						mockClient.EXPECT().
							GetProviderSession(gomock.Any(), id).
							Return(nil, nil).
							Times(1)
						processResult := mixin.FetchProviderSessionFromMetadata()
						Expect(processResult).ToNot(BeNil())
						Expect(processResult.Result).To(Equal(work.ResultFailed))
						Expect(processResult.FailedUpdate).ToNot(BeNil())
						Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("provider session is missing"))
						Expect(mixin.ProviderSession).To(BeNil())
					})

					It("returns successfully", func() {
						expectedProviderSession := authTest.RandomProviderSession()
						mockClient.EXPECT().
							GetProviderSession(gomock.Any(), id).
							Return(expectedProviderSession, nil).
							Times(1)
						processResult := mixin.FetchProviderSessionFromMetadata()
						Expect(processResult).To(BeNil())
						Expect(mixin.ProviderSession).To(Equal(expectedProviderSession))
					})
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
					processResult := mixin.FetchProviderSession(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to fetch provider session; " + testErr.Error()))
					Expect(mixin.ProviderSession).To(BeNil())
				})

				It("returns failed process result if client returns nil", func() {
					mockClient.EXPECT().
						GetProviderSession(gomock.Any(), id).
						Return(nil, nil).
						Times(1)
					processResult := mixin.FetchProviderSession(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("provider session is missing"))
					Expect(mixin.ProviderSession).To(BeNil())
				})

				It("returns successfully", func() {
					expectedProviderSession := authTest.RandomProviderSession()
					mockClient.EXPECT().
						GetProviderSession(gomock.Any(), id).
						Return(expectedProviderSession, nil).
						Times(1)
					processResult := mixin.FetchProviderSession(id)
					Expect(processResult).To(BeNil())
					Expect(mixin.ProviderSession).To(Equal(expectedProviderSession))
				})
			})

			Context("UpdateProviderSession", func() {
				It("returns failed process result if existing provider session is missing", func() {
					providerSessionUpdate := authTest.RandomProviderSessionUpdate()
					processResult := mixin.UpdateProviderSession(*providerSessionUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("provider session is missing"))
				})

				It("returns failing process result if client returns error", func() {
					existingProviderSession := authTest.RandomProviderSession()
					mixin.ProviderSession = existingProviderSession
					providerSessionUpdate := authTest.RandomProviderSessionUpdate()
					testErr := errorsTest.RandomError()
					mockClient.EXPECT().
						UpdateProviderSession(gomock.Any(), existingProviderSession.ID, providerSessionUpdate).
						Return(nil, testErr).
						Times(1)
					processResult := mixin.UpdateProviderSession(*providerSessionUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to update provider session; " + testErr.Error()))
					Expect(mixin.ProviderSession).To(Equal(existingProviderSession))
				})

				It("returns failed process result if client returns nil", func() {
					existingProviderSession := authTest.RandomProviderSession()
					mixin.ProviderSession = existingProviderSession
					providerSessionUpdate := authTest.RandomProviderSessionUpdate()
					mockClient.EXPECT().
						UpdateProviderSession(gomock.Any(), existingProviderSession.ID, providerSessionUpdate).
						Return(nil, nil).
						Times(1)
					processResult := mixin.UpdateProviderSession(*providerSessionUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("provider session is missing"))
					Expect(mixin.ProviderSession).To(Equal(existingProviderSession))
				})

				It("returns successfully", func() {
					existingProviderSession := authTest.RandomProviderSession()
					mixin.ProviderSession = existingProviderSession
					expectedProviderSession := authTest.RandomProviderSession()
					providerSessionUpdate := authTest.RandomProviderSessionUpdate()
					mockClient.EXPECT().
						UpdateProviderSession(gomock.Any(), existingProviderSession.ID, providerSessionUpdate).
						Return(expectedProviderSession, nil).
						Times(1)
					processResult := mixin.UpdateProviderSession(*providerSessionUpdate)
					Expect(processResult).To(BeNil())
					Expect(mixin.ProviderSession).To(Equal(expectedProviderSession))
				})
			})
		})
	})
})
