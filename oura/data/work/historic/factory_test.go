package historic_test

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
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/oura"
	ouraDataWorkHistoric "github.com/tidepool-org/platform/oura/data/work/historic"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/times"
	timesTest "github.com/tidepool-org/platform/times/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("factory", func() {
	It("Type is expected", func() {
		Expect(ouraDataWorkHistoric.Type).To(Equal("org.tidepool.oura.data.historic"))
	})

	It("Quantity is expected", func() {
		Expect(ouraDataWorkHistoric.Quantity).To(Equal(4))
	})

	It("Frequency is expected", func() {
		Expect(ouraDataWorkHistoric.Frequency).To(Equal(5 * time.Second))
	})

	It("ProcessingTimeout is expected", func() {
		Expect(ouraDataWorkHistoric.ProcessingTimeout).To(Equal(15 * time.Minute))
	})

	Context("with dependencies", func() {
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockProviderSessionClient *providerSessionTest.MockClient
		var mockDataSourceClient *dataSourceTest.MockClient
		var mockDataRawClient *dataRawTest.MockClient
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraDataWorkHistoric.Dependencies

		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockProviderSessionClient = providerSessionTest.NewMockClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockDataRawClient = dataRawTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraDataWorkHistoric.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				ProviderSessionClient: mockProviderSessionClient,
				DataSourceClient:      mockDataSourceClient,
				DataRawClient:         mockDataRawClient,
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
				processorFactory, err := ouraDataWorkHistoric.NewProcessorFactory(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processorFactory).To(BeNil())
			})

			It("returns successfully", func() {
				processorFactory, err := ouraDataWorkHistoric.NewProcessorFactory(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processorFactory).ToNot(BeNil())
			})

			Context("with processor factory", func() {
				var processorFactory *workBase.ProcessorFactory

				BeforeEach(func() {
					var err error
					processorFactory, err = ouraDataWorkHistoric.NewProcessorFactory(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processorFactory).ToNot(BeNil())
				})

				Context("Type", func() {
					It("returns the expected type", func() {
						Expect(processorFactory.Type()).To(Equal(ouraDataWorkHistoric.Type))
					})
				})

				Context("Quantity", func() {
					It("returns the expected quantity", func() {
						Expect(processorFactory.Quantity()).To(Equal(ouraDataWorkHistoric.Quantity))
					})
				})

				Context("Frequency", func() {
					It("returns the expected frequency", func() {
						Expect(processorFactory.Frequency()).To(Equal(ouraDataWorkHistoric.Frequency))
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
		var timeRange times.TimeRange
		var timeRangeTruncated times.TimeRange
		var encodedTimeRangeTruncated map[string]any

		BeforeEach(func() {
			var err error
			timeRange = *timesTest.RandomTimeRange(test.AllowOptionals())
			timeRangeTruncated = timeRange.InLocation(time.UTC).Date()
			encodedTimeRangeTruncated, err = metadata.Encode(&timeRangeTruncated)
			Expect(err).ToNot(HaveOccurred())
			Expect(encodedTimeRangeTruncated).ToNot(BeNil())
		})

		It("returns an error if provider session id is missing", func() {
			workCreate, err := ouraDataWorkHistoric.NewWorkCreate("", timeRange)
			Expect(err).To(MatchError("provider session id is missing"))
			Expect(workCreate).To(BeNil())
		})

		It("returns successfully", func() {
			providerSessionID := authTest.RandomProviderSessionID()
			workCreate, err := ouraDataWorkHistoric.NewWorkCreate(providerSessionID, timeRange)
			Expect(err).ToNot(HaveOccurred())
			Expect(workCreate).To(Equal(&work.Create{
				Type:              ouraDataWorkHistoric.Type,
				GroupID:           pointer.From(fmt.Sprintf("org.tidepool.oura:%s", providerSessionID)),
				DeduplicationID:   pointer.From(fmt.Sprintf("%s:%s", providerSessionID, timeRangeTruncated.String(oura.TimeRangeFormat))),
				SerialID:          pointer.From(fmt.Sprintf("org.tidepool.oura.data:%s", providerSessionID)),
				ProcessingTimeout: 900,
				Metadata: map[string]any{
					providerSessionWork.MetadataKeyProviderSessionID: providerSessionID,
					times.MetadataKeyTimeRange:                       encodedTimeRangeTruncated,
				},
			}))
		})
	})
})
