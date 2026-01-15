package work_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataSourceWork "github.com/tidepool-org/platform/data/source/work"
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
		Expect(dataSourceWork.MetadataKeyID).To(Equal("dataSourceId"))
	})

	Context("with base processor and client", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		// var mockProviderSessionClient *providerSessionTest.MockClient
		var mockClient *dataSourceTest.MockClient
		var processor *workBase.Processor
		// var providerSessionMixin *providerSessionWork.Mixin

		BeforeEach(func() {
			var err error
			ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			// mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockClient = dataSourceTest.NewMockClient(mockController)
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
			// providerSessionMixin, err = providerSessionWork.NewMixin(processor, mockProviderSessionClient)
			// Expect(err).ToNot(HaveOccurred())
		})

		Context("NewMixin", func() {
			It("returns error if processor is missing", func() {
				mixin, err := dataSourceWork.NewMixin(nil, mockClient)
				Expect(err).To(MatchError("processor is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns error if client is missing", func() {
				mixin, err := dataSourceWork.NewMixin(processor, nil)
				Expect(err).To(MatchError("client is missing"))
				Expect(mixin).To(BeNil())
			})

			It("returns processor success", func() {
				mixin, err := dataSourceWork.NewMixin(processor, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
			})
		})

		Context("Mixin", func() {
			var mixin *dataSourceWork.Mixin
			var wrk *work.Work
			var mockProcessingUpdater *workTest.MockProcessingUpdater

			BeforeEach(func() {
				var err error
				mixin, err = dataSourceWork.NewMixin(processor, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(mixin).ToNot(BeNil())
				ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
				wrk = workTest.RandomWork()
				Expect(mixin.Process(ctx, wrk, mockProcessingUpdater)).To(BeNil())
			})

			// 			Context("ProviderSessionIDFromDataSource", func() {
			// 				It("returns error if data source is missing", func() {
			// 					id, err := mixin.ProviderSessionIDFromDataSource()
			// 					Expect(err).To(MatchError("data source is missing"))
			// 					Expect(id).To(BeNil())
			// 				})
			//
			// 				It("returns successfully", func() {
			// 					expectedID := test.RandomString()
			// 					mixin.DataSource = dataSourceTest.RandomSource()
			// 					mixin.DataSource.ProviderSessionID = pointer.FromString(expectedID)
			// 					id, err := mixin.ProviderSessionIDFromDataSource()
			// 					Expect(err).ToNot(HaveOccurred())
			// 					Expect(id).To(PointTo(Equal(expectedID)))
			// 				})
			// 			})
			//
			// 			Context("FetchProviderSessionFromDataSource", func() {
			// 				It("returns failed process result if data source is missing", func() {
			// 					processResult := mixin.FetchProviderSessionFromDataSource()
			// 					Expect(processResult).ToNot(BeNil())
			// 					Expect(processResult.Result).To(Equal(work.ResultFailed))
			// 					Expect(processResult.FailedUpdate).ToNot(BeNil())
			// 					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("unable to get provider session id from data source; data source is missing"))
			// 					Expect(mixin.ProviderSession).To(BeNil())
			// 				})
			//
			// 				It("returns failed process result if data source provider session id is missing", func() {
			// 					mixin.DataSource = dataSourceTest.RandomSource()
			// 					mixin.DataSource.ProviderSessionID = nil
			// 					processResult := mixin.FetchProviderSessionFromDataSource()
			// 					Expect(processResult).ToNot(BeNil())
			// 					Expect(processResult.Result).To(Equal(work.ResultFailed))
			// 					Expect(processResult.FailedUpdate).ToNot(BeNil())
			// 					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("unable to get provider session id from data source"))
			// 					Expect(mixin.ProviderSession).To(BeNil())
			// 				})
			//
			// 				When("id is valid", func() {
			// 					var id string
			//
			// 					BeforeEach(func() {
			// 						id = test.RandomString()
			// 						mixin.DataSource = dataSourceTest.RandomSource()
			// 						mixin.DataSource.ProviderSessionID = pointer.FromString(id)
			// 					})
			//
			// 					It("returns failing process result if client returns error", func() {
			// 						testErr := errorsTest.RandomError()
			// 						mockProviderSessionClient.EXPECT().
			// 							GetProviderSession(gomock.Any(), id).
			// 							Return(nil, testErr).
			// 							Times(1)
			// 						processResult := mixin.FetchProviderSessionFromDataSource()
			// 						Expect(processResult).ToNot(BeNil())
			// 						Expect(processResult.Result).To(Equal(work.ResultFailing))
			// 						Expect(processResult.FailingUpdate).ToNot(BeNil())
			// 						Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to fetch provider session; " + testErr.Error()))
			// 						Expect(mixin.ProviderSession).To(BeNil())
			// 					})
			//
			// 					It("returns failed process result if client returns nil", func() {
			// 						mockProviderSessionClient.EXPECT().
			// 							GetProviderSession(gomock.Any(), id).
			// 							Return(nil, nil).
			// 							Times(1)
			// 						processResult := mixin.FetchProviderSessionFromDataSource()
			// 						Expect(processResult).ToNot(BeNil())
			// 						Expect(processResult.Result).To(Equal(work.ResultFailed))
			// 						Expect(processResult.FailedUpdate).ToNot(BeNil())
			// 						Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("provider session is missing"))
			// 						Expect(mixin.ProviderSession).To(BeNil())
			// 					})
			//
			// 					It("returns successfully", func() {
			// 						expectedProviderSession := authTest.RandomProviderSession()
			// 						mockProviderSessionClient.EXPECT().
			// 							GetProviderSession(gomock.Any(), id).
			// 							Return(expectedProviderSession, nil).
			// 							Times(1)
			// 						processResult := mixin.FetchProviderSessionFromDataSource()
			// 						Expect(processResult).To(BeNil())
			// 						Expect(mixin.ProviderSession).To(Equal(expectedProviderSession))
			// 					})
			// 				})
			// 			})

			Context("DataSourceIDFromMetadata", func() {
				It("returns error if unable to parse", func() {
					wrk.Metadata[dataSourceWork.MetadataKeyID] = true
					id, err := mixin.DataSourceIDFromMetadata()
					Expect(id).To(BeNil())
					Expect(err).To(MatchError("unable to parse data source id from metadata; type is not string, but bool"))
				})

				It("returns successfully", func() {
					expectedID := test.RandomString()
					wrk.Metadata[dataSourceWork.MetadataKeyID] = expectedID
					id, err := mixin.DataSourceIDFromMetadata()
					Expect(err).ToNot(HaveOccurred())
					Expect(id).To(PointTo(Equal(expectedID)))
				})
			})

			Context("FetchDataSourceFromMetadata", func() {
				It("returns failed process result if unable to parse id", func() {
					wrk.Metadata[dataSourceWork.MetadataKeyID] = true
					processResult := mixin.FetchDataSourceFromMetadata()
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("unable to get data source id from metadata; unable to parse data source id from metadata; type is not string, but bool"))
					Expect(mixin.DataSource).To(BeNil())
				})

				It("returns failed process result if id is missing", func() {
					wrk.Metadata[dataSourceWork.MetadataKeyID] = nil
					processResult := mixin.FetchDataSourceFromMetadata()
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("unable to get data source id from metadata"))
					Expect(mixin.DataSource).To(BeNil())
				})

				When("id is valid", func() {
					var id string

					BeforeEach(func() {
						id = test.RandomString()
						wrk.Metadata[dataSourceWork.MetadataKeyID] = id
					})

					It("returns failing process result if client returns error", func() {
						testErr := errorsTest.RandomError()
						mockClient.EXPECT().
							Get(gomock.Any(), id).
							Return(nil, testErr).
							Times(1)
						processResult := mixin.FetchDataSourceFromMetadata()
						Expect(processResult).ToNot(BeNil())
						Expect(processResult.Result).To(Equal(work.ResultFailing))
						Expect(processResult.FailingUpdate).ToNot(BeNil())
						Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to fetch data source; " + testErr.Error()))
						Expect(mixin.DataSource).To(BeNil())
					})

					It("returns failed process result if client returns nil", func() {
						mockClient.EXPECT().
							Get(gomock.Any(), id).
							Return(nil, nil).
							Times(1)
						processResult := mixin.FetchDataSourceFromMetadata()
						Expect(processResult).ToNot(BeNil())
						Expect(processResult.Result).To(Equal(work.ResultFailed))
						Expect(processResult.FailedUpdate).ToNot(BeNil())
						Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data source is missing"))
						Expect(mixin.DataSource).To(BeNil())
					})

					It("returns successfully", func() {
						expectedDataSource := dataSourceTest.RandomSource()
						mockClient.EXPECT().
							Get(gomock.Any(), id).
							Return(expectedDataSource, nil).
							Times(1)
						processResult := mixin.FetchDataSourceFromMetadata()
						Expect(processResult).To(BeNil())
						Expect(mixin.DataSource).To(Equal(expectedDataSource))
					})
				})
			})

			Context("FetchDataSource", func() {
				var id string

				BeforeEach(func() {
					id = test.RandomString()
				})

				It("returns failing process result if client returns error", func() {
					testErr := errorsTest.RandomError()
					mockClient.EXPECT().
						Get(gomock.Any(), id).
						Return(nil, testErr).
						Times(1)
					processResult := mixin.FetchDataSource(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to fetch data source; " + testErr.Error()))
					Expect(mixin.DataSource).To(BeNil())
				})

				It("returns failed process result if client returns nil", func() {
					mockClient.EXPECT().
						Get(gomock.Any(), id).
						Return(nil, nil).
						Times(1)
					processResult := mixin.FetchDataSource(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data source is missing"))
					Expect(mixin.DataSource).To(BeNil())
				})

				It("returns successfully", func() {
					expectedDataSource := dataSourceTest.RandomSource()
					mockClient.EXPECT().
						Get(gomock.Any(), id).
						Return(expectedDataSource, nil).
						Times(1)
					processResult := mixin.FetchDataSource(id)
					Expect(processResult).To(BeNil())
					Expect(mixin.DataSource).To(Equal(expectedDataSource))
				})
			})

			Context("UpdateDataSource", func() {
				It("returns failed process result if existing data source is missing", func() {
					dataSourceUpdate := dataSourceTest.RandomUpdate()
					processResult := mixin.UpdateDataSource(*dataSourceUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data source is missing"))
				})

				It("returns failing process result if client returns error", func() {
					existingDataSource := dataSourceTest.RandomSource()
					mixin.DataSource = existingDataSource
					dataSourceUpdate := dataSourceTest.RandomUpdate()
					testErr := errorsTest.RandomError()
					mockClient.EXPECT().
						Update(gomock.Any(), *existingDataSource.ID, nil, dataSourceUpdate).
						Return(nil, testErr).
						Times(1)
					processResult := mixin.UpdateDataSource(*dataSourceUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to update data source; " + testErr.Error()))
					Expect(mixin.DataSource).To(Equal(existingDataSource))
				})

				It("returns failed process result if client returns nil", func() {
					existingDataSource := dataSourceTest.RandomSource()
					mixin.DataSource = existingDataSource
					dataSourceUpdate := dataSourceTest.RandomUpdate()
					mockClient.EXPECT().
						Update(gomock.Any(), *existingDataSource.ID, nil, dataSourceUpdate).
						Return(nil, nil).
						Times(1)
					processResult := mixin.UpdateDataSource(*dataSourceUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data source is missing"))
					Expect(mixin.DataSource).To(Equal(existingDataSource))
				})

				It("returns successfully", func() {
					existingDataSource := dataSourceTest.RandomSource()
					mixin.DataSource = existingDataSource
					expectedDataSource := dataSourceTest.RandomSource()
					dataSourceUpdate := dataSourceTest.RandomUpdate()
					mockClient.EXPECT().
						Update(gomock.Any(), *existingDataSource.ID, nil, dataSourceUpdate).
						Return(expectedDataSource, nil).
						Times(1)
					processResult := mixin.UpdateDataSource(*dataSourceUpdate)
					Expect(processResult).To(BeNil())
					Expect(mixin.DataSource).To(Equal(expectedDataSource))
				})
			})
		})
	})
})
