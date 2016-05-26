package prime_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/device"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Prime", func() {
	var rawObject = testing.RawBaseObject()
	var meta = &device.Meta{
		Type:    "deviceEvent",
		SubType: "prime",
	}

	BeforeEach(func() {
		rawObject["type"] = "deviceEvent"
		rawObject["subType"] = "prime"
		rawObject["primeTarget"] = "cannula"
		rawObject["volume"] = 0.0
	})

	Context("primeTarget", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is empty", rawObject, "primeTarget", "",
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("", []string{"cannula", "tubing"}), "/primeTarget", meta)},
			),
			Entry("is not one of the predefined types", rawObject, "primeTarget", "bad",
				[]*service.Error{testing.ComposeError(validator.ErrorStringNotOneOf("bad", []string{"cannula", "tubing"}), "/primeTarget", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is cannula type", rawObject, "primeTarget", "cannula"),
			Entry("is tubing type", rawObject, "primeTarget", "tubing"),
		)
	})

	Context("cannula volume", func() {
		BeforeEach(func() {
			rawObject["type"] = "deviceEvent"
			rawObject["subType"] = "prime"
			rawObject["primeTarget"] = "cannula"
		})

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is less than 0", rawObject, "volume", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 3.0), "/volume", meta)},
			),
			Entry("is more than 3", rawObject, "volume", 3.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(3.1, 0.0, 3.0), "/volume", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is 0", rawObject, "volume", 0.0),
			Entry("is 3.0", rawObject, "volume", 3.0),
			Entry("is no decimal", rawObject, "volume", 2),
		)
	})

	Context("tubing volume", func() {
		BeforeEach(func() {
			rawObject["type"] = "deviceEvent"
			rawObject["subType"] = "prime"
			rawObject["primeTarget"] = "tubing"
		})

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is less than 0", rawObject, "volume", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 100.0), "/volume", meta)},
			),
			Entry("is more than 100", rawObject, "volume", 100.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(100.1, 0.0, 100.0), "/volume", meta)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is 0", rawObject, "volume", 0.0),
			Entry("is 100.0", rawObject, "volume", 100.0),
			Entry("is no decimal", rawObject, "volume", 55),
		)
	})
})
