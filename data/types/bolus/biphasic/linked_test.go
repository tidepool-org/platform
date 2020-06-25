package biphasic_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/bolus/biphasic"
	datatypesBolusBiphasicTest "github.com/tidepool-org/platform/data/types/bolus/biphasic/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/structure"
)

var _ = Describe("Linked Bolus", func() {
	It("Type is expected", func() {
		Expect(biphasic.NewLinkedBolus()).To(Equal(&biphasic.LinkedBolus{}))
	})

	Context("LinkedBolus", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *biphasic.LinkedBolus), expectedErrors ...error) {
					datum := datatypesBolusBiphasicTest.NewLinkedBolus()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *biphasic.LinkedBolus) {},
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *biphasic.LinkedBolus)) {
					for _, origin := range structure.Origins() {
						datum := datatypesBolusBiphasicTest.NewLinkedBolus()
						mutator(datum)
						expectedDatum := datatypesBolusBiphasicTest.CloneLinkedBolus(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *biphasic.LinkedBolus) {},
				),
			)
		})
	})
})
