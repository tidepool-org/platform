package recalculate_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	dataWorkSummaryRecalculate "github.com/tidepool-org/platform/data/work/summary/recalculate"
	dataWorkSummaryRecalculateTest "github.com/tidepool-org/platform/data/work/summary/recalculate/test"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("factory", func() {
	It("Type is expected", func() {
		Expect(dataWorkSummaryRecalculate.Type).To(Equal("org.tidepool.data.summary.recalculate"))
	})

	It("Quantity is expected", func() {
		Expect(dataWorkSummaryRecalculate.Quantity).To(Equal(1))
	})

	It("Frequency is expected", func() {
		Expect(dataWorkSummaryRecalculate.Frequency).To(Equal(30 * time.Second))
	})

	It("ProcessingTimeout is expected", func() {
		Expect(dataWorkSummaryRecalculate.ProcessingTimeout).To(Equal(5 * time.Minute))
	})

	It("ID is expected", func() {
		Expect(dataWorkSummaryRecalculate.ID).To(Equal("BACK-4158-gap-based-cut-points"))
	})

	Context("with dependencies", func() {
		var dependencies dataWorkSummaryRecalculate.Dependencies

		BeforeEach(func() {
			mockController := gomock.NewController(GinkgoT())
			dependencies = dataWorkSummaryRecalculate.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: workTest.NewMockClient(mockController),
				},
				SummaryLister: dataWorkSummaryRecalculateTest.NewMockSummaryLister(mockController),
			}
		})

		Context("Dependencies", func() {
			Context("Validate", func() {
				It("returns an error if work client is missing", func() {
					dependencies.WorkClient = nil
					Expect(dependencies.Validate()).To(MatchError("work client is missing"))
				})

				It("returns an error if summary lister is missing", func() {
					dependencies.SummaryLister = nil
					Expect(dependencies.Validate()).To(MatchError("summary lister is missing"))
				})

				It("returns successfully", func() {
					Expect(dependencies.Validate()).To(Succeed())
				})
			})
		})

		Context("NewProcessorFactory", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.SummaryLister = nil
				processorFactory, err := dataWorkSummaryRecalculate.NewProcessorFactory(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; summary lister is missing"))
				Expect(processorFactory).To(BeNil())
			})

			It("returns successfully", func() {
				processorFactory, err := dataWorkSummaryRecalculate.NewProcessorFactory(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processorFactory).ToNot(BeNil())
				Expect(processorFactory.Type()).To(Equal(dataWorkSummaryRecalculate.Type))
				Expect(processorFactory.Quantity()).To(Equal(dataWorkSummaryRecalculate.Quantity))
				Expect(processorFactory.Frequency()).To(Equal(dataWorkSummaryRecalculate.Frequency))

				processor, err := processorFactory.New()
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})
		})
	})

	Context("NewWorkCreate", func() {
		It("returns an error if the id is missing", func() {
			workCreate, err := dataWorkSummaryRecalculate.NewWorkCreate("")
			Expect(err).To(MatchError("id is missing"))
			Expect(workCreate).To(BeNil())
		})

		It("returns work deduplicated by the id and available now", func() {
			workCreate, err := dataWorkSummaryRecalculate.NewWorkCreate(dataWorkSummaryRecalculate.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(workCreate).To(PointTo(MatchAllFields(Fields{
				"Type":                    Equal(dataWorkSummaryRecalculate.Type),
				"GroupID":                 BeNil(),
				"DeduplicationID":         PointTo(Equal(dataWorkSummaryRecalculate.ID)),
				"SerialID":                BeNil(),
				"ProcessingAvailableTime": BeTemporally("~", time.Now(), time.Second),
				"ProcessingPriority":      Equal(0),
				"ProcessingTimeout":       Equal(300),
				"Metadata":                BeNil(),
			})))
		})
	})
})
