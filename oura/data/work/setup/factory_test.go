package setup_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	authTest "github.com/tidepool-org/platform/auth/test"
	dataSetTest "github.com/tidepool-org/platform/data/set/test"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	ouraDataWorkSetup "github.com/tidepool-org/platform/oura/data/work/setup"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("factory", func() {
	It("Type is expected", func() {
		Expect(ouraDataWorkSetup.Type).To(Equal("org.tidepool.oura.data.setup"))
	})

	It("Quantity is expected", func() {
		Expect(ouraDataWorkSetup.Quantity).To(Equal(1))
	})

	It("Frequency is expected", func() {
		Expect(ouraDataWorkSetup.Frequency).To(Equal(5 * time.Second))
	})

	It("ProcessingTimeout is expected", func() {
		Expect(ouraDataWorkSetup.ProcessingTimeout).To(Equal(3 * time.Minute))
	})

	Context("with dependencies", func() {
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockProviderSessionClient *providerSessionTest.MockClient
		var mockDataSourceClient *dataSourceTest.MockClient
		var mockDataSetClient *dataSetTest.MockClient
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraDataWorkSetup.Dependencies

		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockDataSetClient = dataSetTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraDataWorkSetup.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				ProviderSessionClient: mockProviderSessionClient,
				DataSourceClient:      mockDataSourceClient,
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

		Context("NewProcessorFactory", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.WorkClient = nil
				processorFactory, err := ouraDataWorkSetup.NewProcessorFactory(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processorFactory).To(BeNil())
			})

			It("returns successfully", func() {
				processorFactory, err := ouraDataWorkSetup.NewProcessorFactory(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processorFactory).ToNot(BeNil())
			})

			Context("with processor factory", func() {
				var processorFactory *workBase.ProcessorFactory

				BeforeEach(func() {
					var err error
					processorFactory, err = ouraDataWorkSetup.NewProcessorFactory(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processorFactory).ToNot(BeNil())
				})

				Context("Type", func() {
					It("returns the expected type", func() {
						Expect(processorFactory.Type()).To(Equal(ouraDataWorkSetup.Type))
					})
				})

				Context("Quantity", func() {
					It("returns the expected quantity", func() {
						Expect(processorFactory.Quantity()).To(Equal(ouraDataWorkSetup.Quantity))
					})
				})

				Context("Frequency", func() {
					It("returns the expected frequency", func() {
						Expect(processorFactory.Frequency()).To(Equal(ouraDataWorkSetup.Frequency))
					})
				})

				Context("New", func() {
					It("returns a new processor", func() {
						processor, err := processorFactory.New()
						Expect(err).ToNot(HaveOccurred())
						Expect(processor).ToNot(BeNil())
					})
				})
			})
		})
	})

	Context("NewWorkCreate", func() {
		It("returns an error if provider session id is missing", func() {
			workCreate, err := ouraDataWorkSetup.NewWorkCreate("")
			Expect(err).To(MatchError("provider session id is missing"))
			Expect(workCreate).To(BeNil())
		})

		It("returns successfully", func() {
			providerSessionID := authTest.RandomProviderSessionID()
			workCreate, err := ouraDataWorkSetup.NewWorkCreate(providerSessionID)
			Expect(err).ToNot(HaveOccurred())
			Expect(workCreate).To(Equal(&work.Create{
				Type:              ouraDataWorkSetup.Type,
				GroupID:           pointer.From(fmt.Sprintf("org.tidepool.oura:%s", providerSessionID)),
				DeduplicationID:   pointer.From(providerSessionID),
				SerialID:          pointer.From(fmt.Sprintf("org.tidepool.oura.data:%s", providerSessionID)),
				ProcessingTimeout: 180,
				Metadata: map[string]any{
					providerSessionWork.MetadataKeyProviderSessionID: providerSessionID,
				},
			}))
		})
	})
})
