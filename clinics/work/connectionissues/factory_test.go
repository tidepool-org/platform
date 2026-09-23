package connectionissues_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	clinicsWorkConnectionIssues "github.com/tidepool-org/platform/clinics/work/connectionissues"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("factory", func() {
	It("Type is expected", func() {
		Expect(clinicsWorkConnectionIssues.Type).
			To(Equal("org.tidepool.clinic.patient.connection.issues.update"))
	})

	It("Quantity is expected", func() {
		Expect(clinicsWorkConnectionIssues.Quantity).To(Equal(1))
	})

	It("Frequency is expected", func() {
		Expect(clinicsWorkConnectionIssues.Frequency).To(Equal(1 * time.Minute))
	})

	It("ProcessingTimeout is expected", func() {
		Expect(clinicsWorkConnectionIssues.ProcessingTimeout).To(Equal(5 * time.Minute))
	})

	Context("with dependencies", func() {
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockClinicClient *clinicsTest.MockClient
		var dependencies clinicsWorkConnectionIssues.Dependencies

		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockClinicClient = clinicsTest.NewMockClient(mockController)
			dependencies = clinicsWorkConnectionIssues.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				ClinicClient: mockClinicClient,
			}
		})

		Context("Dependencies", func() {
			Context("Validate", func() {
				It("returns an error if work client is missing", func() {
					dependencies.WorkClient = nil
					Expect(dependencies.Validate()).To(MatchError("work client is missing"))
				})

				It("returns an error if clinic client is missing", func() {
					dependencies.ClinicClient = nil
					Expect(dependencies.Validate()).
						To(MatchError("clinic client is missing"))
				})

				It("returns successfully", func() {
					Expect(dependencies.Validate()).To(Succeed())
				})
			})
		})

		Context("NewProcessorFactory", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.ClinicClient = nil
				processorFactory, err := clinicsWorkConnectionIssues.
					NewProcessorFactory(dependencies)
				Expect(err).
					To(MatchError("dependencies is invalid; clinic client is missing"))
				Expect(processorFactory).To(BeNil())
			})

			Context("with processor factory", func() {
				var processorFactory *workBase.ProcessorFactory

				BeforeEach(func() {
					var err error
					processorFactory, err = clinicsWorkConnectionIssues.
						NewProcessorFactory(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processorFactory).ToNot(BeNil())
				})

				It("has the expected type, quantity and frequency", func() {
					Expect(processorFactory.Type()).
						To(Equal(clinicsWorkConnectionIssues.Type))
					Expect(processorFactory.Quantity()).
						To(Equal(clinicsWorkConnectionIssues.Quantity))
					Expect(processorFactory.Frequency()).
						To(Equal(clinicsWorkConnectionIssues.Frequency))
				})

				It("creates a processor", func() {
					processor, err := processorFactory.New()
					Expect(err).ToNot(HaveOccurred())
					Expect(processor).ToNot(BeNil())
				})
			})
		})
	})

	Context("NewWorkCreate", func() {
		It("returns the singleton work create", func() {
			workCreate, err := clinicsWorkConnectionIssues.NewWorkCreate()
			Expect(err).ToNot(HaveOccurred())
			Expect(workCreate).To(Equal(&work.Create{
				Type:            clinicsWorkConnectionIssues.Type,
				DeduplicationID: pointer.From(work.DeduplicationIDSingleton),
				ProcessingTimeout: int(
					clinicsWorkConnectionIssues.ProcessingTimeout.Seconds()),
			}))
		})
	})
})
