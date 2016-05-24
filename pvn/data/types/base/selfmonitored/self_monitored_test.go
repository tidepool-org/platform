package selfmonitored_test

import (
	"github.com/tidepool-org/platform/pvn/data/context"
	"github.com/tidepool-org/platform/pvn/data/normalizer"
	"github.com/tidepool-org/platform/pvn/data/types/base/selfmonitored"
	"github.com/tidepool-org/platform/pvn/data/types/base/testing"
	"github.com/tidepool-org/platform/pvn/data/types/common/bloodglucose"
	"github.com/tidepool-org/platform/pvn/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Selfmonitored BloodGlucose", func() {

	var rawObject = testing.RawBaseObject()

	rawObject["type"] = "smbg"
	rawObject["units"] = "mmol/L"
	rawObject["value"] = 5

	Context("units", func() {

		DescribeTable("units when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "units", "",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("", []string{common.Mmoll, common.MmolL, common.Mgdl, common.MgdL}), "/units")},
			),
			Entry("not one of the predefined values", rawObject, "units", "wrong",
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorStringNotOneOf("wrong", []string{common.Mmoll, common.MmolL, common.Mgdl, common.MgdL}), "/units")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("mmol/l", rawObject, "units", "mmol/l"),
			Entry("mmol/L", rawObject, "units", "mmol/L"),
			Entry("mg/dl", rawObject, "units", "mg/dl"),
			Entry("mg/dL", rawObject, "units", "mg/dL"),
		)

	})

	Context("value", func() {

		DescribeTable("value when", testing.ExpectFieldNotValid,
			Entry("less than 0", rawObject, "value", -0.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(-0.1, common.MgdLFromValue, common.MgdLToValue), "/value")},
			),
			Entry("greater than 1000", rawObject, "value", 1000.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(1000.1, common.MgdLFromValue, common.MgdLToValue), "/value")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("above 0", rawObject, "value", 0.1),
			Entry("below 1000", rawObject, "value", 990.85745),
			Entry("as integer", rawObject, "value", 12),
		)

	})

	Context("normalized when mmol/L", func() {

		DescribeTable("normalization", func(val, expected float64) {
			selfMonitoredBg := selfmonitored.New()
			selfMonitoredBg.Value = &val
			selfMonitoredBg.Units = &common.Mmoll

			testContext := context.NewStandard()
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).To(BeNil())
			selfMonitoredBg.Normalize(standardNormalizer)
			Expect(selfMonitoredBg.Units).To(Equal(&common.MmolL))
			Expect(selfMonitoredBg.Value).To(Equal(&expected))
		},
			Entry("expected lower bg value", 3.7, 3.7),
			Entry("below max", 54.99, 54.99),
			Entry("expected upper bg value", 23.0, 23.0),
		)
	})

	Context("normalized when mg/dL", func() {

		DescribeTable("normalization", func(val, expected float64) {
			selfMonitoredBg := selfmonitored.New()
			selfMonitoredBg.Value = &val
			selfMonitoredBg.Units = &common.Mgdl

			testContext := context.NewStandard()
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).To(BeNil())
			selfMonitoredBg.Normalize(standardNormalizer)
			Expect(selfMonitoredBg.Units).To(Equal(&common.MmolL))
			Expect(selfMonitoredBg.Value).To(Equal(&expected))
		},
			Entry("expected lower bg value", 60.0, 3.33044879462732),
			Entry("below max", 990.85745, 55.0),
			Entry("expected upper bg value", 400.0, 22.202991964182132),
		)
	})

})
