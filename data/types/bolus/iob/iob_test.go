package iob_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/bolus/iob"
	dataTypesBolusIobTest "github.com/tidepool-org/platform/data/types/bolus/iob/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Iob", func() {

	It("InsulinOnBoardMaximum is expected", func() {
		Expect(iob.InsulinOnBoardMaximum).To(Equal(250.0))
	})

	It("InsulinOnBoardMinimum is expected", func() {
		Expect(iob.InsulinOnBoardMinimum).To(Equal(0.0))
	})

	Context("NewIob", func() {
		It("is successful", func() {
			Expect(iob.NewIob()).To(Equal(&iob.Iob{}))
		})
	})

	Context("Iob", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *iob.Iob), expectedErrors ...error) {
					datum := dataTypesBolusIobTest.NewIob()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *iob.Iob) {},
				),
				Entry("Valid value",
					func(datum *iob.Iob) {
						datum.InsulinOnBoard = pointer.FromFloat64(10)
					},
				),
				Entry("insulin on board out of range (lower)",
					func(datum *iob.Iob) {
						datum.InsulinOnBoard = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 250.0), "/insulinOnBoard"),
				),
				Entry("insulin on board out of range (upper)",
					func(datum *iob.Iob) {
						datum.InsulinOnBoard = pointer.FromFloat64(250.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(250.1, 0.0, 250.0), "/insulinOnBoard"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *iob.Iob)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBolusIobTest.NewIob()
						mutator(datum)
						expectedDatum := dataTypesBolusIobTest.CloneIob(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *iob.Iob) {},
				),
			)
		})
	})
})
