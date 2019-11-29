package continuous_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewTestRate() *continuous.Rate {
	datum := continuous.NewRate()
	return datum
}

func CloneRate(datum *continuous.Rate) *continuous.Rate {
	if datum == nil {
		return nil
	}
	clone := continuous.NewRate()
	return clone
}

var _ = Describe("Rate", func() {

	Context("New", func() {
		It("returns the expected datum", func() {
			datum := NewTestRate()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Units).To(BeNil())
			Expect(datum.Value).To(BeNil())
		})
	})
	Context("Validate", func() {
		DescribeTable("validates the datum",
			func(units *string, mutator func(datum *continuous.Rate, units *string), expectedErrors ...error) {
				datum := NewTestRate()
				mutator(datum, units)
				dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
			},
			Entry("rate missing",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Rate, units *string) { datum.Value = nil },
			),
			Entry("rate out of range (lower)",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Rate, units *string) { datum.Value = pointer.FromFloat64(-100.1) },
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-100.1, -100.0, 100.0), "/value"),
			),
			Entry("rate in range (lower)",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Rate, units *string) { datum.Value = pointer.FromFloat64(-100.0) },
			),
			Entry("rate in range (upper)",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Rate, units *string) { datum.Value = pointer.FromFloat64(100.0) },
			),
			Entry("rate out of range (upper)",
				pointer.FromString("mg/dl"),
				func(datum *continuous.Rate, units *string) { datum.Value = pointer.FromFloat64(100.1) },
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, -100.0, 100.0), "/value"),
			),
		)
	})

})
