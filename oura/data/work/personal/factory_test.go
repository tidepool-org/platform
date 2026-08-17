package personal_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	authTest "github.com/tidepool-org/platform/auth/test"
	dataRawTest "github.com/tidepool-org/platform/data/raw/test"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
	ouraDataWorkPersonal "github.com/tidepool-org/platform/oura/data/work/personal"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("factory", func() {
	It("Type is expected", func() {
		Expect(ouraDataWorkPersonal.Type).To(Equal("org.tidepool.oura.data.personal"))
	})

	It("Quantity is expected", func() {
		Expect(ouraDataWorkPersonal.Quantity).To(Equal(4))
	})

	It("Frequency is expected", func() {
		Expect(ouraDataWorkPersonal.Frequency).To(Equal(5 * time.Second))
	})

	It("ProcessingTimeout is expected", func() {
		Expect(ouraDataWorkPersonal.ProcessingTimeout).To(Equal(3 * time.Minute))
	})

	Context("with dependencies", func() {
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockProviderSessionClient *providerSessionTest.MockClient
		var mockDataSourceClient *dataSourceTest.MockClient
		var mockDataRawClient *dataRawTest.MockClient
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraDataWork.Dependencies

		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockDataRawClient = dataRawTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraDataWork.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				ProviderSessionClient: mockProviderSessionClient,
				DataSourceClient:      mockDataSourceClient,
				DataRawClient:         mockDataRawClient,
				OuraClient:            mockOuraClient,
			}
		})

		Context("NewProcessorFactory", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.WorkClient = nil
				processorFactory, err := ouraDataWorkPersonal.NewProcessorFactory(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processorFactory).To(BeNil())
			})

			It("returns successfully", func() {
				processorFactory, err := ouraDataWorkPersonal.NewProcessorFactory(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processorFactory).ToNot(BeNil())
			})

			Context("with processor factory", func() {
				var processorFactory *workBase.ProcessorFactory

				BeforeEach(func() {
					var err error
					processorFactory, err = ouraDataWorkPersonal.NewProcessorFactory(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processorFactory).ToNot(BeNil())
				})

				Context("Type", func() {
					It("returns the expected type", func() {
						Expect(processorFactory.Type()).To(Equal(ouraDataWorkPersonal.Type))
					})
				})

				Context("Quantity", func() {
					It("returns the expected quantity", func() {
						Expect(processorFactory.Quantity()).To(Equal(ouraDataWorkPersonal.Quantity))
					})
				})

				Context("Frequency", func() {
					It("returns the expected frequency", func() {
						Expect(processorFactory.Frequency()).To(Equal(ouraDataWorkPersonal.Frequency))
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
			workCreate, err := ouraDataWorkPersonal.NewWorkCreate("")
			Expect(err).To(MatchError("provider session id is missing"))
			Expect(workCreate).To(BeNil())
		})

		It("returns successfully", func() {
			providerSessionID := authTest.RandomProviderSessionID()
			workCreate, err := ouraDataWorkPersonal.NewWorkCreate(providerSessionID)
			Expect(err).ToNot(HaveOccurred())
			Expect(workCreate).To(Equal(&work.Create{
				Type:              ouraDataWorkPersonal.Type,
				GroupID:           pointer.From(fmt.Sprintf("org.tidepool.oura:%s", providerSessionID)),
				DeduplicationID:   pointer.From(providerSessionID),
				SerialID:          pointer.From(fmt.Sprintf("org.tidepool.oura:%s", providerSessionID)),
				ProcessingTimeout: 180,
				Metadata: map[string]any{
					providerSessionWork.MetadataKeyProviderSessionID: providerSessionID,
				},
			}))
		})
	})
})
