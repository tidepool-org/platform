package common_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/common"
	dataTypeCommonTest "github.com/tidepool-org/platform/data/types/common/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Prescriptor", func() {
	Context("Object Creation", func() {

		It("Manual Prescriptor is expected", func() {
			Expect(common.ManualPrescriptor).To(Equal("manual"))
		})

		It("Auto Prescriptor is expected", func() {
			Expect(common.AutoPrescriptor).To(Equal("auto"))
		})

		It("Hybrid Prescriptor is expected", func() {
			Expect(common.HybridPrescriptor).To(Equal("hybrid"))
		})

		Context("NewPrescriptor", func() {
			It("is successful", func() {
				Expect(common.NewPrescriptor()).To(Equal(&common.Prescriptor{}))
			})
		})

	})
	Context("Prescriptor", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *common.Prescriptor), expectedErrors ...error) {
					datum := dataTypeCommonTest.NewPrescriptor()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *common.Prescriptor) {},
				),
				Entry("Prescriptor is nil",
					func(datum *common.Prescriptor) {
						datum.Prescriptor = nil
					},
				),
				Entry("Invalid Prescriptor value",
					func(datum *common.Prescriptor) {
						datum.Prescriptor = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"auto", "manual", "hybrid"}), "/prescriptor"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *common.Prescriptor)) {
					for _, origin := range structure.Origins() {
						datum := dataTypeCommonTest.NewPrescriptor()
						mutator(datum)
						expectedDatum := dataTypeCommonTest.ClonePrescriptor(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *common.Prescriptor) {},
				),
			)
		})
	})
})
