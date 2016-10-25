package continuous_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types/base/blood/glucose/continuous"
)

var _ = Describe("Continuous", func() {
	Context("Type", func() {
		It("returns the expected type", func() {
			Expect(continuous.Type()).To(Equal("cbg"))
		})
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(continuous.NewDatum()).To(Equal(&continuous.Continuous{}))
		})
	})

	Context("New", func() {
		It("returns the expected continuous", func() {
			Expect(continuous.New()).To(Equal(&continuous.Continuous{}))
		})
	})

	Context("Init", func() {
		It("returns the expected continuous", func() {
			testContinuous := continuous.Init()
			Expect(testContinuous).ToNot(BeNil())
			Expect(testContinuous.ID).ToNot(BeEmpty())
			Expect(testContinuous.Type).To(Equal("cbg"))
		})
	})

	Context("with new continuous", func() {
		var testContinuous *continuous.Continuous

		BeforeEach(func() {
			testContinuous = continuous.New()
			Expect(testContinuous).ToNot(BeNil())
		})

		Context("Init", func() {
			It("initializes the continuous", func() {
				testContinuous.Init()
				Expect(testContinuous.ID).ToNot(BeEmpty())
				Expect(testContinuous.Type).To(Equal("cbg"))
			})
		})
	})
})
