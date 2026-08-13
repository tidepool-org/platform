package postprocess_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	dataWorkPostprocessTest "github.com/tidepool-org/platform/data/work/postprocess/test"
	userTest "github.com/tidepool-org/platform/user/test"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("Factory", func() {
	It("Quantity is expected", func() {
		Expect(dataWorkPostprocess.Quantity).To(Equal(15))
	})

	It("Frequency is expected", func() {
		Expect(dataWorkPostprocess.Frequency).To(Equal(30 * time.Second))
	})

	Context("with dependencies", func() {
		var controller *gomock.Controller
		var dependencies dataWorkPostprocess.Dependencies

		BeforeEach(func() {
			controller = gomock.NewController(GinkgoT())
			dependencies = dataWorkPostprocess.Dependencies{
				Dependencies:  workBase.Dependencies{WorkClient: workTest.NewMockClient(controller)},
				Summarizers:   dataWorkPostprocessTest.NewMockSummarizers(controller),
				ClinicsClient: clinicsTest.NewMockClient(controller),
				UserClient:    userTest.NewMockClient(controller),
			}
		})

		AfterEach(func() {
			controller.Finish()
		})

		DescribeTable("Validate",
			func(mutator func(dependencies *dataWorkPostprocess.Dependencies), expected string) {
				mutator(&dependencies)
				if expected == "" {
					Expect(dependencies.Validate()).To(Succeed())
				} else {
					Expect(dependencies.Validate()).To(MatchError(expected))
				}
			},
			Entry("succeeds", func(dependencies *dataWorkPostprocess.Dependencies) {}, ""),
			Entry("reports the work client is missing", func(dependencies *dataWorkPostprocess.Dependencies) {
				dependencies.WorkClient = nil
			}, "work client is missing"),
			Entry("reports the summarizers are missing", func(dependencies *dataWorkPostprocess.Dependencies) {
				dependencies.Summarizers = nil
			}, "summarizers is missing"),
			Entry("reports the clinics client is missing", func(dependencies *dataWorkPostprocess.Dependencies) {
				dependencies.ClinicsClient = nil
			}, "clinics client is missing"),
			Entry("reports the user client is missing", func(dependencies *dataWorkPostprocess.Dependencies) {
				dependencies.UserClient = nil
			}, "user client is missing"),
		)

		It("NewProcessorFactory reports the type, quantity and frequency of the work", func() {
			processorFactory, err := dataWorkPostprocess.NewProcessorFactory(dependencies)
			Expect(err).ToNot(HaveOccurred())
			Expect(processorFactory.Type()).To(Equal(dataWorkPostprocess.Type))
			Expect(processorFactory.Quantity()).To(Equal(dataWorkPostprocess.Quantity))
			Expect(processorFactory.Frequency()).To(Equal(dataWorkPostprocess.Frequency))

			processor, err := processorFactory.New()
			Expect(err).ToNot(HaveOccurred())
			Expect(processor).ToNot(BeNil())
		})

		It("NewProcessorFactory returns an error when the dependencies are invalid", func() {
			dependencies.ClinicsClient = nil
			processorFactory, err := dataWorkPostprocess.NewProcessorFactory(dependencies)
			Expect(err).To(MatchError(ContainSubstring("clinics client is missing")))
			Expect(processorFactory).To(BeNil())
		})
	})
})
