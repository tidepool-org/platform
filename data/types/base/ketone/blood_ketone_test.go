package ketone_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/bloodglucose"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/base"
	"github.com/tidepool-org/platform/data/types/base/ketone"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

func NewRawObjectMmolL() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "bloodKetone"
	rawObject["units"] = bloodglucose.MmolL
	rawObject["value"] = 5
	return rawObject
}

func NewRawObjectMgdL() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "bloodKetone"
	rawObject["units"] = bloodglucose.MgdL
	rawObject["value"] = 99
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
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("", []string{"mmol/l", "mmol/L", "mg/dl", "mg/dL"}), "/units", NewMeta())},
			),
			Entry("is not one of the predefined values", NewRawObjectMmolL(), "units", "wrong",
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("wrong", []string{"mmol/l", "mmol/L", "mg/dl", "mg/dL"}), "/units", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is mmol/l", NewRawObjectMmolL(), "units", "mmol/l"),
			Entry("is mmol/L", NewRawObjectMmolL(), "units", "mmol/L"),
			Entry("is mg/dl", NewRawObjectMgdL(), "units", "mg/dl"),
			Entry("is mg/dL", NewRawObjectMgdL(), "units", "mg/dL"),
		)
	})

	Context("value", func() {
		DescribeTable("value when", testing.ExpectFieldNotValid,
			Entry("is less than 0", NewRawObjectMgdL(), "value", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, bloodglucose.MgdLLowerLimit, bloodglucose.MgdLUpperLimit), "/value", NewMeta())},
			),
			Entry("is greater than 1000", NewRawObjectMgdL(), "value", 1000.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(1000.1, bloodglucose.MgdLLowerLimit, bloodglucose.MgdLUpperLimit), "/value", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is above 0", NewRawObjectMgdL(), "value", 0.1),
			Entry("is below 1000", NewRawObjectMgdL(), "value", bloodglucose.MgdLUpperLimit),
			Entry("is an integer", NewRawObjectMgdL(), "value", 4),
		)
	})

	Context("normalized when mmol/L", func() {
		DescribeTable("normalization", func(val, expected float64) {
			bloodKetone, err := ketone.New()
			units := bloodglucose.MmolL
			bloodKetone.Units = &units
			bloodKetone.Value = &val

			testContext := context.NewStandard()
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).To(BeNil())
			bloodKetone.Normalize(standardNormalizer)
			Expect(*bloodKetone.Units).To(Equal(bloodglucose.MmolL))
			Expect(*bloodKetone.Value).To(Equal(expected))
		},
			Entry("is expected lower bg value", 3.7, 3.7),
			Entry("is below max", 54.99, 54.99),
			Entry("is expected upper bg value", 23.0, 23.0),
		)
	})

	Context("normalized when mg/dL", func() {
		DescribeTable("normalization", func(val, expected float64) {
			bloodKetone, err := ketone.New()
			Expect(err).To(BeNil())
			units := bloodglucose.MgdL
			bloodKetone.Units = &units
			bloodKetone.Value = &val

			testContext := context.NewStandard()
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).To(BeNil())
			bloodKetone.Normalize(standardNormalizer)
			Expect(*bloodKetone.Units).To(Equal(bloodglucose.MmolL))
			Expect(*bloodKetone.Value).To(Equal(expected))
		},
			Entry("is expected lower bg value", 60.0, 3.33044879462732),
			Entry("is below max", bloodglucose.MgdLUpperLimit, 55.50747991045534),
			Entry("is expected upper bg value", 400.0, 22.202991964182132),
		)
	})
})
