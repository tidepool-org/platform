package common_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types/common"
	dataTypesCommonTest "github.com/tidepool-org/platform/data/types/common/test"
	"github.com/tidepool-org/platform/test"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("InputTime", func() {

	Context("NewInputTime", func() {
		It("is successful", func() {
			Expect(common.NewInputTime()).To(Equal(&common.InputTime{}))
		})
	})

	Context("InputTime", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *common.InputTime), expectedErrors ...error) {
					datum := dataTypesCommonTest.NewInputTime()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *common.InputTime) {},
				),
				Entry("Valid inputTime",
					func(datum *common.InputTime) {
						datum.InputTime = pointer.FromString(test.RandomTime().Format(time.RFC3339Nano))
					},
				),
				Entry("invalid inputTime",
					func(datum *common.InputTime) {
						datum.InputTime = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339Nano), "/inputTime"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *common.InputTime)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesCommonTest.NewInputTime()
						mutator(datum)
						expectedDatum := dataTypesCommonTest.CloneInputTime(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *common.InputTime) {},
				),
			)
		})
	})
})
