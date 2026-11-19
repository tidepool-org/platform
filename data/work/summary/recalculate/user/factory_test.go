package user_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	dataWorkSummaryRecalculateUser "github.com/tidepool-org/platform/data/work/summary/recalculate/user"
	dataWorkSummaryRecalculateUserTest "github.com/tidepool-org/platform/data/work/summary/recalculate/user/test"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("factory", func() {
	It("Quantity is expected", func() {
		Expect(dataWorkSummaryRecalculateUser.Quantity).To(Equal(10))
	})

	It("Frequency is expected", func() {
		Expect(dataWorkSummaryRecalculateUser.Frequency).To(Equal(5 * time.Second))
	})

	Context("NewSummarizers", func() {
		It("returns an error if the summarizer registry is missing", func() {
			summarizers, err := dataWorkSummaryRecalculateUser.NewSummarizers(nil)
			Expect(err).To(MatchError("summarizer registry is missing"))
			Expect(summarizers).To(BeNil())
		})
	})

	Context("with dependencies", func() {
		var dependencies dataWorkSummaryRecalculateUser.Dependencies

		BeforeEach(func() {
			mockController := gomock.NewController(GinkgoT())
			dependencies = dataWorkSummaryRecalculateUser.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: workTest.NewMockClient(mockController),
				},
				Summarizers: dataWorkSummaryRecalculateUserTest.
					NewMockSummarizers(mockController),
				ClinicsClient: clinicsTest.NewMockClient(mockController),
			}
		})

		Context("Dependencies", func() {
			Context("Validate", func() {
				It("returns an error if work client is missing", func() {
					dependencies.WorkClient = nil
					Expect(dependencies.Validate()).To(MatchError("work client is missing"))
				})

				It("returns an error if summarizers is missing", func() {
					dependencies.Summarizers = nil
					Expect(dependencies.Validate()).To(MatchError("summarizers is missing"))
				})

				It("returns an error if clinics client is missing", func() {
					dependencies.ClinicsClient = nil
					Expect(dependencies.Validate()).
						To(MatchError("clinics client is missing"))
				})

				It("returns successfully", func() {
					Expect(dependencies.Validate()).To(Succeed())
				})
			})
		})

		Context("NewProcessorFactory", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.Summarizers = nil
				processorFactory, err :=
					dataWorkSummaryRecalculateUser.NewProcessorFactory(dependencies)
				Expect(err).
					To(MatchError("dependencies is invalid; summarizers is missing"))
				Expect(processorFactory).To(BeNil())
			})

			It("returns successfully", func() {
				processorFactory, err :=
					dataWorkSummaryRecalculateUser.NewProcessorFactory(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processorFactory).ToNot(BeNil())
				Expect(processorFactory.Type()).
					To(Equal(dataWorkSummaryRecalculateUser.Type))
				Expect(processorFactory.Quantity()).
					To(Equal(dataWorkSummaryRecalculateUser.Quantity))
				Expect(processorFactory.Frequency()).
					To(Equal(dataWorkSummaryRecalculateUser.Frequency))

				processor, err := processorFactory.New()
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})
		})
	})
})
