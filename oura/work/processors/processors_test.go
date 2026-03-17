package processors_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	dataRawTest "github.com/tidepool-org/platform/data/raw/test"
	dataSetTest "github.com/tidepool-org/platform/data/set/test"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	ouraWorkProcessors "github.com/tidepool-org/platform/oura/work/processors"
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

		Context("EnsureWork", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.WorkClient = nil
				Expect(ouraWorkProcessors.EnsureWork(dependencies)).To(MatchError("dependencies is invalid; work client is missing"))
			})

			It("returns successfully", func() {
				Expect(ouraWorkProcessors.EnsureWork(dependencies)).To(Succeed())
			})
		})
	})
})
