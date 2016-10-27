package calibration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/calibration"
	"github.com/tidepool-org/platform/data/types/testing"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

func NewRawObjectMmolL() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "deviceEvent"
	rawObject["subType"] = "calibration"
	rawObject["units"] = glucose.MmolL
	rawObject["value"] = 5.5
	return rawObject
}

func NewRawObjectMgdL() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "deviceEvent"
	rawObject["subType"] = "calibration"
	rawObject["units"] = glucose.MgdL
	rawObject["value"] = 180
	return rawObject
}

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "calibration",
	}
}

var _ = Describe("Calibration", func() {
	Context("units", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty", NewRawObjectMmolL(), "units", "",
				[]*service.Error{testing.ComposeError(service.ErrorValueStringNotOneOf("", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta())},
			),
			Entry("is not one of the predefined values", NewRawObjectMmolL(), "units", "wrong",
				[]*service.Error{testing.ComposeError(service.ErrorValueStringNotOneOf("wrong", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta())},
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
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is less than 0", NewRawObjectMgdL(), "value", -0.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(-0.1, glucose.MgdLLowerLimit, glucose.MgdLUpperLimit), "/value", NewMeta())},
			),
			Entry("is greater than 1000", NewRawObjectMgdL(), "value", 1000.1,
				[]*service.Error{testing.ComposeError(service.ErrorValueNotInRange(1000.1, glucose.MgdLLowerLimit, glucose.MgdLUpperLimit), "/value", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is 0", NewRawObjectMgdL(), "value", 0.0),
			Entry("is above 0", NewRawObjectMgdL(), "value", 0.1),
			Entry("is below max", NewRawObjectMgdL(), "value", glucose.MgdLUpperLimit),
			Entry("is an integer", NewRawObjectMgdL(), "value", 4),
		)
	})

	Context("normalized when mmol/L", func() {
		DescribeTable("normalization", func(val, expected float64) {
			calibrationEvent := calibration.Init()
			units := glucose.MmolL
			calibrationEvent.Units = &units
			calibrationEvent.Value = &val

			testContext, err := context.NewStandard(log.NewNull())
			Expect(err).ToNot(HaveOccurred())
			Expect(testContext).ToNot(BeNil())
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).ToNot(HaveOccurred())
			Expect(standardNormalizer).ToNot(BeNil())
			calibrationEvent.Normalize(standardNormalizer)
			Expect(*calibrationEvent.Units).To(Equal(glucose.MmolL))
			Expect(*calibrationEvent.Value).To(Equal(expected))
		},
			Entry("is expected lower bg value", 3.7, 3.7),
			Entry("is below max", 54.99, 54.99),
			Entry("is expected upper bg value", 23.0, 23.0),
		)
	})

	Context("normalized when mg/dL", func() {
		DescribeTable("normalization", func(val, expected float64) {
			calibrationEvent := calibration.Init()
			units := glucose.MgdL
			calibrationEvent.Units = &units
			calibrationEvent.Value = &val

			testContext, err := context.NewStandard(log.NewNull())
			Expect(err).ToNot(HaveOccurred())
			Expect(testContext).ToNot(BeNil())
			standardNormalizer, err := normalizer.NewStandard(testContext)
			Expect(err).ToNot(HaveOccurred())
			Expect(standardNormalizer).ToNot(BeNil())
			calibrationEvent.Normalize(standardNormalizer)
			Expect(*calibrationEvent.Units).To(Equal(glucose.MmolL))
			Expect(*calibrationEvent.Value).To(Equal(expected))
		},
			Entry("is expected lower bg value", 60.0, 3.33045),
			Entry("is below max", glucose.MgdLUpperLimit, 55.50748),
			Entry("is expected upper bg value", 400.0, 22.20299),
		)
	})
})
