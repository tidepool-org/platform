package work_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	authProviderSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	authProviderSessionWorkTest "github.com/tidepool-org/platform/auth/providersession/work/test"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataSourceWork "github.com/tidepool-org/platform/data/source/work"
	dataSourceWorkTest "github.com/tidepool-org/platform/data/source/work/test"
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
		Expect(dataSourceWork.MetadataKeyID).To(Equal("dataSourceId"))
	})

	Context("with base processing and client", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockClient *dataSourceWorkTest.MockClient
		var baseProcessing *workBase.Processing
		var providerSessionProcessing *authProviderSessionWork.Processing

		BeforeEach(func() {
			var err error

			ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockProviderSessionClient := authProviderSessionWorkTest.NewMockClient(mockController)
			mockClient = dataSourceWorkTest.NewMockClient(mockController)
			processResultBuilder := &workBase.ProcessResultBuilder{
				ProcessResultPendingBuilder: &workBase.ConstantProcessResultPendingBuilder{
					Duration: time.Minute,
				},
				ProcessResultFailingBuilder: &workBase.ConstantProcessResultFailingBuilder{
					Duration: time.Second,
				},
			}
			baseProcessing = workBase.NewProcessing(processResultBuilder)
			providerSessionProcessing, err = authProviderSessionWork.NewProcessing(baseProcessing, mockProviderSessionClient)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("NewProcessing", func() {
			It("returns error if base processing is missing", func() {
				processing, err := dataSourceWork.NewProcessing(nil, mockClient)
				Expect(err).To(MatchError("processing is missing"))
				Expect(processing).To(BeNil())
			})

			It("returns error if base processing is missing", func() {
				processing, err := dataSourceWork.NewProcessing(providerSessionProcessing, nil)
				Expect(err).To(MatchError("client is missing"))
				Expect(processing).To(BeNil())
			})

			It("returns processing success", func() {
				processing, err := dataSourceWork.NewProcessing(providerSessionProcessing, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(processing).ToNot(BeNil())
			})
		})

		Context("Processing", func() {
			var processing *dataSourceWork.Processing
			var wrk *work.Work
			var mockProcessingUpdater *workTest.MockProcessingUpdater

			BeforeEach(func() {
				var err error
				processing, err = dataSourceWork.NewProcessing(providerSessionProcessing, mockClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(processing).ToNot(BeNil())
				ctx = log.NewContextWithLogger(context.Background(), logNull.NewLogger())
				wrk = workTest.RandomWork()
				Expect(processing.Process(ctx, wrk, mockProcessingUpdater)()).To(BeNil())
			})

			Context("DataSourceIDFromMetadata", func() {
				It("returns error if unable to parse", func() {
					wrk.Metadata[dataSourceWork.MetadataKeyID] = true
					id, err := processing.DataSourceIDFromMetadata()
					Expect(id).To(BeNil())
					Expect(err).To(MatchError("unable to parse data source id from metadata; type is not string, but bool"))
				})

				It("returns successfully", func() {
					expectedID := test.RandomString()
					wrk.Metadata[dataSourceWork.MetadataKeyID] = expectedID
					id, err := processing.DataSourceIDFromMetadata()
					Expect(err).ToNot(HaveOccurred())
					Expect(id).ToNot(BeNil())
					Expect(*id).To(Equal(expectedID))
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
					processResult := processing.FetchDataSource(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to fetch data source; " + testErr.Error()))
					Expect(processing.DataSource).To(BeNil())
				})

				It("returns failed process result if client returns nil", func() {
					mockClient.EXPECT().
						Get(gomock.Any(), id).
						Return(nil, nil).
						Times(1)
					processResult := processing.FetchDataSource(id)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data source is missing"))
					Expect(processing.DataSource).To(BeNil())
				})

				It("returns successfully", func() {
					expectedDataSource := dataSourceTest.RandomSource()
					mockClient.EXPECT().
						Get(gomock.Any(), id).
						Return(expectedDataSource, nil).
						Times(1)
					processResult := processing.FetchDataSource(id)
					Expect(processResult).To(BeNil())
					Expect(processing.DataSource).To(Equal(expectedDataSource))
				})
			})

			Context("UpdateDataSource", func() {
				It("returns failed process result if existing data source is missing", func() {
					dataSourceUpdate := dataSourceTest.RandomUpdate()
					processResult := processing.UpdateDataSource(*dataSourceUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data source is missing"))
				})

				It("returns failing process result if client returns error", func() {
					existingDataSource := dataSourceTest.RandomSource()
					processing.DataSource = existingDataSource
					dataSourceUpdate := dataSourceTest.RandomUpdate()
					testErr := errorsTest.RandomError()
					mockClient.EXPECT().
						Update(gomock.Any(), *existingDataSource.ID, nil, dataSourceUpdate).
						Return(nil, testErr).
						Times(1)
					processResult := processing.UpdateDataSource(*dataSourceUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailing))
					Expect(processResult.FailingUpdate).ToNot(BeNil())
					Expect(processResult.FailingUpdate.FailingError.Error).To(MatchError("unable to update data source; " + testErr.Error()))
					Expect(processing.DataSource).To(Equal(existingDataSource))
				})

				It("returns failed process result if client returns nil", func() {
					existingDataSource := dataSourceTest.RandomSource()
					processing.DataSource = existingDataSource
					dataSourceUpdate := dataSourceTest.RandomUpdate()
					mockClient.EXPECT().
						Update(gomock.Any(), *existingDataSource.ID, nil, dataSourceUpdate).
						Return(nil, nil).
						Times(1)
					processResult := processing.UpdateDataSource(*dataSourceUpdate)
					Expect(processResult).ToNot(BeNil())
					Expect(processResult.Result).To(Equal(work.ResultFailed))
					Expect(processResult.FailedUpdate).ToNot(BeNil())
					Expect(processResult.FailedUpdate.FailedError.Error).To(MatchError("data source is missing"))
					Expect(processing.DataSource).To(Equal(existingDataSource))
				})

				It("returns successfully", func() {
					existingDataSource := dataSourceTest.RandomSource()
					processing.DataSource = existingDataSource
					expectedDataSource := dataSourceTest.RandomSource()
					dataSourceUpdate := dataSourceTest.RandomUpdate()
					mockClient.EXPECT().
						Update(gomock.Any(), *existingDataSource.ID, nil, dataSourceUpdate).
						Return(expectedDataSource, nil).
						Times(1)
					processResult := processing.UpdateDataSource(*dataSourceUpdate)
					Expect(processResult).To(BeNil())
					Expect(processing.DataSource).To(Equal(expectedDataSource))
				})
			})
		})
	})
})
