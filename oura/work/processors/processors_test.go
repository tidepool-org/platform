package processors_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	dataRawTest "github.com/tidepool-org/platform/data/raw/test"
	dataSetTest "github.com/tidepool-org/platform/data/set/test"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	ouraWebhookWorkSubscribe "github.com/tidepool-org/platform/oura/webhook/work/subscribe"
	ouraWorkProcessors "github.com/tidepool-org/platform/oura/work/processors"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("processors", func() {
	Context("with dependencies", func() {
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockProviderSessionClient *providerSessionTest.MockClient
		var mockDataSourceClient *dataSourceTest.MockClient
		var mockDataRawClient *dataRawTest.MockClient
		var mockDataSetClient *dataSetTest.MockClient
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraWorkProcessors.Dependencies

		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockDataRawClient = dataRawTest.NewMockClient(mockController)
			mockDataSetClient = dataSetTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraWorkProcessors.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				ProviderSessionClient: mockProviderSessionClient,
				DataSourceClient:      mockDataSourceClient,
				DataRawClient:         mockDataRawClient,
				DataSetClient:         mockDataSetClient,
				OuraClient:            mockOuraClient,
			}
		})

		Context("Dependencies", func() {
			Context("Validate", func() {
				It("returns an error if work client is missing", func() {
					dependencies.WorkClient = nil
					Expect(dependencies.Validate()).To(MatchError("work client is missing"))
				})

				It("returns an error if provider session client is missing", func() {
					dependencies.ProviderSessionClient = nil
					Expect(dependencies.Validate()).To(MatchError("provider session client is missing"))
				})

				It("returns an error if data source client is missing", func() {
					dependencies.DataSourceClient = nil
					Expect(dependencies.Validate()).To(MatchError("data source client is missing"))
				})

				It("returns an error if data raw client is missing", func() {
					dependencies.DataRawClient = nil
					Expect(dependencies.Validate()).To(MatchError("data raw client is missing"))
				})

				It("returns an error if data set client is missing", func() {
					dependencies.DataSetClient = nil
					Expect(dependencies.Validate()).To(MatchError("data set client is missing"))
				})

				It("returns an error if oura client is missing", func() {
					dependencies.OuraClient = nil
					Expect(dependencies.Validate()).To(MatchError("oura client is missing"))
				})

				It("returns successfully", func() {
					Expect(dependencies.Validate()).To(Succeed())
				})
			})
		})

		Context("NewProcessorFactories", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.WorkClient = nil
				processorFactories, err := ouraWorkProcessors.NewProcessorFactories(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processorFactories).To(BeNil())
			})

			It("returns successfully", func() {
				processorFactories, err := ouraWorkProcessors.NewProcessorFactories(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processorFactories).To(HaveLen(5))
			})
		})
	})

	Context("EnsureWork", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
		})

		It("returns an error if context is missing", func() {
			Expect(ouraWorkProcessors.EnsureWork(context.Context(nil), mockWorkClient)).To(MatchError("context is missing"))
		})

		It("returns an error if context is missing", func() {
			Expect(ouraWorkProcessors.EnsureWork(ctx, nil)).To(MatchError("work client is missing"))
		})

		Context("with webhook subscribe work create", func() {
			var webhookSubscribeWorkCreate *work.Create
			var webhookSubscribeWork *work.Work

			BeforeEach(func() {
				var err error
				webhookSubscribeWorkCreate, err = ouraWebhookWorkSubscribe.NewWorkCreate()
				Expect(err).ToNot(HaveOccurred())
				Expect(webhookSubscribeWorkCreate).ToNot(BeNil())
				webhookSubscribeWork = workTest.NewWorkFromCreateWithState(webhookSubscribeWorkCreate, work.StatePending)
			})

			It("returns error if the work client returns an error", func() {
				testErr := errorsTest.RandomError()
				mockWorkClient.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, create *work.Create) {
						Expect(create).To(Equal(webhookSubscribeWorkCreate))
					}).
					Return(nil, testErr)
				Expect(ouraWorkProcessors.EnsureWork(ctx, mockWorkClient)).To(MatchError("unable to create webhook subscribe work; " + testErr.Error()))
			})

			It("returns successfully", func() {
				mockWorkClient.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, create *work.Create) {
						Expect(create).To(Equal(webhookSubscribeWorkCreate))
					}).
					Return(webhookSubscribeWork, nil)
				Expect(ouraWorkProcessors.EnsureWork(ctx, mockWorkClient)).To(Succeed())
			})
		})
	})
})
