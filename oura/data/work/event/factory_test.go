package event_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	providerSessionTest "github.com/tidepool-org/platform/auth/providersession/test"
	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	authTest "github.com/tidepool-org/platform/auth/test"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	"github.com/tidepool-org/platform/metadata"
	ouraDataWorkEvent "github.com/tidepool-org/platform/oura/data/work/event"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	ouraWebhookTest "github.com/tidepool-org/platform/oura/webhook/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("factory", func() {
	It("Type is expected", func() {
		Expect(ouraDataWorkEvent.Type).To(Equal("org.tidepool.oura.data.event"))
	})

	It("Quantity is expected", func() {
		Expect(ouraDataWorkEvent.Quantity).To(Equal(4))
	})

	It("Frequency is expected", func() {
		Expect(ouraDataWorkEvent.Frequency).To(Equal(5 * time.Second))
	})

	It("ProcessingTimeout is expected", func() {
		Expect(ouraDataWorkEvent.ProcessingTimeout).To(Equal(3 * time.Minute))
	})

	Context("with dependencies", func() {
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockProviderSessionClient *providerSessionTest.MockClient
		var mockDataSourceClient *dataSourceTest.MockClient
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraDataWorkEvent.Dependencies

		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraDataWorkEvent.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				ProviderSessionClient: mockProviderSessionClient,
				DataSourceClient:      mockDataSourceClient,
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
				processorFactory, err := ouraDataWorkEvent.NewProcessorFactory(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processorFactory).To(BeNil())
			})

			It("returns successfully", func() {
				processorFactory, err := ouraDataWorkEvent.NewProcessorFactory(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processorFactory).ToNot(BeNil())
			})

			Context("with processor factory", func() {
				var processorFactory *workBase.ProcessorFactory

				BeforeEach(func() {
					var err error
					processorFactory, err = ouraDataWorkEvent.NewProcessorFactory(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processorFactory).ToNot(BeNil())
				})

				Context("Type", func() {
					It("returns the expected type", func() {
						Expect(processorFactory.Type()).To(Equal(ouraDataWorkEvent.Type))
					})
				})

				Context("Quantity", func() {
					It("returns the expected quantity", func() {
						Expect(processorFactory.Quantity()).To(Equal(ouraDataWorkEvent.Quantity))
					})
				})

				Context("Frequency", func() {
					It("returns the expected frequency", func() {
						Expect(processorFactory.Frequency()).To(Equal(ouraDataWorkEvent.Frequency))
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
			workCreate, err := ouraDataWorkEvent.NewWorkCreate("", ouraWebhookTest.RandomEvent())
			Expect(err).To(MatchError("provider session id is missing"))
			Expect(workCreate).To(BeNil())
		})

		It("returns an error if event is missing", func() {
			workCreate, err := ouraDataWorkEvent.NewWorkCreate(authTest.RandomProviderSessionID(), nil)
			Expect(err).To(MatchError("event is missing"))
			Expect(workCreate).To(BeNil())
		})

		It("returns successfully", func() {
			providerSessionID := authTest.RandomProviderSessionID()
			event := ouraWebhookTest.RandomEvent()
			encodedEvent, err := metadata.Encode(event)
			Expect(err).ToNot(HaveOccurred())
			Expect(encodedEvent).ToNot(BeNil())
			workCreate, err := ouraDataWorkEvent.NewWorkCreate(providerSessionID, event)
			Expect(err).ToNot(HaveOccurred())
			Expect(workCreate).To(Equal(&work.Create{
				Type:              ouraDataWorkEvent.Type,
				GroupID:           pointer.FromString(fmt.Sprintf("org.tidepool.oura:%s", providerSessionID)),
				DeduplicationID:   pointer.FromString(fmt.Sprintf("%s:%s", providerSessionID, event.String())),
				SerialID:          pointer.FromString(fmt.Sprintf("org.tidepool.oura.data:%s", providerSessionID)),
				ProcessingTimeout: 180,
				Metadata: map[string]any{
					providerSessionWork.MetadataKeyProviderSessionID: providerSessionID,
					ouraWebhook.MetadataKeyEvent:                     encodedEvent,
				},
			}))
		})
	})
})
