package ketone_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	commonKetone "github.com/tidepool-org/platform/data/blood/ketone"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/base"
	"github.com/tidepool-org/platform/data/types/base/blood/ketone"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

func NewRawObjectMmolL() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "bloodKetone"
	rawObject["units"] = commonKetone.MmolL
	rawObject["value"] = 5
	return rawObject
}

func NewMeta() interface{} {
	return &base.Meta{
		Type: "bloodKetone",
	}
}

var _ = Describe("BloodKetone", func() {
	Context("units", func() {
		DescribeTable("units when", testing.ExpectFieldNotValid,
			Entry("is empty", NewRawObjectMmolL(), "units", "",
				[]*service.Error{testing.ComposeError(service.ErrorValueStringNotOneOf("", []string{"mmol/L", "mmol/l"}), "/units", NewMeta())},
			),
			Entry("is not one of the predefined values", NewRawObjectMmolL(), "units", "wrong",
				[]*service.Error{testing.ComposeError(service.ErrorValueStringNotOneOf("wrong", []string{"mmol/L", "mmol/l"}), "/units", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is mmol/l", NewRawObjectMmolL(), "units", "mmol/l"),
			Entry("is mmol/L", NewRawObjectMmolL(), "units", "mmol/L"),
		)
	})

	Context("value", func() {
		DescribeTable("value when", testing.ExpectFieldNotValid,
			Entry("is less than 0.0", NewRawObjectMmolL(), "value", -0.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-0.1, commonKetone.MmolLLowerLimit, commonKetone.MmolLUpperLimit), "/value", NewMeta())},
			),
			Entry("is greater than 10.0", NewRawObjectMmolL(), "value", 10.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(10.1, commonKetone.MmolLLowerLimit, commonKetone.MmolLUpperLimit), "/value", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is above 0.0", NewRawObjectMmolL(), "value", 0.0),
			Entry("is below 10.0", NewRawObjectMmolL(), "value", commonKetone.MmolLUpperLimit),
			Entry("is an integer", NewRawObjectMmolL(), "value", 4),
		)
	})

	Context("normalized when mmol/L", func() {
		DescribeTable("normalization", func(val, expected float64) {
			bloodKetone := ketone.Init()
			units := commonKetone.MmolL
			bloodKetone.Units = &units
			bloodKetone.Value = &val

			testContext, err := context.NewStandard(log.NewNull())
			Expect(err).ToNot(HaveOccurred())
			Expect(testContext).ToNot(BeNil())
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).ToNot(HaveOccurred())
			Expect(standardNormalizer).ToNot(BeNil())
			bloodKetone.Normalize(standardNormalizer)
			Expect(*bloodKetone.Units).To(Equal(commonKetone.MmolL))
			Expect(*bloodKetone.Value).To(Equal(expected))
		},
			Entry("is expected lower bg value", 3.7, 3.7),
			Entry("is below max", 9.99, 9.99),
			Entry("is expected upper bg value", 7.0, 7.0),
		)
	})
})
