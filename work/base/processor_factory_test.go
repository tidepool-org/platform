package base_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gomock "go.uber.org/mock/gomock"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("processor_factory", func() {
	var typ string
	var quantity int
	var frequency time.Duration
	var newProcessorFunc workBase.NewProcessorFunc

	BeforeEach(func() {
		typ = test.RandomString()
		quantity = test.RandomIntFromRange(1, test.RandomIntMaximum())
		frequency = test.RandomDurationFromRange(1, test.RandomDurationFromRange(1, test.RandomDurationMaximum()))
		newProcessorFunc = func() (work.Processor, error) {
			return nil, nil
		}
	})

	Context("NewProcessorFactory", func() {
		It("returns an error when type is missing", func() {
			processorFactory, err := workBase.NewProcessorFactory("", quantity, frequency, newProcessorFunc)
			Expect(err).To(MatchError("type is missing"))
			Expect(processorFactory).To(BeNil())
		})

		It("returns an error when quantity is zero or less", func() {
			processorFactory, err := workBase.NewProcessorFactory(typ, 0, frequency, newProcessorFunc)
			Expect(err).To(MatchError("quantity is invalid"))
			Expect(processorFactory).To(BeNil())
		})

		It("returns an error when frequency is zero or less", func() {
			processorFactory, err := workBase.NewProcessorFactory(typ, quantity, 0, newProcessorFunc)
			Expect(err).To(MatchError("frequency is invalid"))
			Expect(processorFactory).To(BeNil())
		})

		It("returns an error when newProcessorFunc is nil", func() {
			processorFactory, err := workBase.NewProcessorFactory(typ, quantity, frequency, nil)
			Expect(err).To(MatchError("new processor func is missing"))
			Expect(processorFactory).To(BeNil())
		})

		It("returns ProcessorFactory when all parameters are valid", func() {
			processorFactory, err := workBase.NewProcessorFactory(typ, quantity, frequency, newProcessorFunc)
			Expect(err).ToNot(HaveOccurred())
			Expect(processorFactory).ToNot(BeNil())
		})
	})

	Context("with ProcessorFactory", func() {
		var processorFactory *workBase.ProcessorFactory

		JustBeforeEach(func() {
			var err error
			processorFactory, err = workBase.NewProcessorFactory(typ, quantity, frequency, newProcessorFunc)
			Expect(err).ToNot(HaveOccurred())
			Expect(processorFactory).ToNot(BeNil())
		})

		Context("Type", func() {
			It("returns the type", func() {
				Expect(processorFactory.Type()).To(Equal(typ))
			})
		})

		Context("Quantity", func() {
			It("returns the quantity", func() {
				Expect(processorFactory.Quantity()).To(Equal(quantity))
			})
		})

		Context("Frequency", func() {
			It("returns the frequency", func() {
				Expect(processorFactory.Frequency()).To(Equal(frequency))
			})
		})

		Context("New", func() {
			var mockController *gomock.Controller
			var expectedProcessor *workTest.MockProcessor
			var expectedErr error

			BeforeEach(func() {
				mockController = gomock.NewController(GinkgoT())
				expectedProcessor = workTest.NewMockProcessor(mockController)
				expectedErr = errorsTest.RandomError()
				newProcessorFunc = func() (work.Processor, error) { return expectedProcessor, expectedErr }
			})

			It("returns an error from newProcessorFunc", func() {
				processor, err := processorFactory.New()
				Expect(err).To(Equal(expectedErr))
				Expect(processor).To(Equal(expectedProcessor))
			})
		})
	})
})
