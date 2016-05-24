package prime_test

import (
	"github.com/tidepool-org/platform/pvn/data/types/base/testing"
	"github.com/tidepool-org/platform/pvn/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Prime Event", func() {

	var rawObject = testing.RawBaseObject()

	BeforeEach(func() {

		rawObject["type"] = "deviceEvent"
		rawObject["subType"] = "prime"
		rawObject["primeTarget"] = "cannula"
		rawObject["volume"] = 0.0

	})

	Context("primeTarget", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "primeTarget", "",
				[]*service.Error{testing.SetExpectedErrorSource(
					validator.ErrorStringNotOneOf("", []string{"cannula", "tubing"}), "/primeTarget",
				)},
			),
			Entry("not one of the predefined types", rawObject, "primeTarget", "bad",
				[]*service.Error{testing.SetExpectedErrorSource(
					validator.ErrorStringNotOneOf("bad", []string{"cannula", "tubing"}), "/primeTarget",
				)},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("cannula type", rawObject, "primeTarget", "cannula"),
			Entry("tubing type", rawObject, "primeTarget", "tubing"),
		)

	})

	Context("cannula volume", func() {

		BeforeEach(func() {
			rawObject["type"] = "deviceEvent"
			rawObject["subType"] = "prime"
			rawObject["primeTarget"] = "cannula"
		})

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("less than 0", rawObject, "volume", -0.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(-0.1, 0.0, 3.0), "/volume")},
			),
			Entry("more than 3", rawObject, "volume", 3.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(3.1, 0.0, 3.0), "/volume")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("0", rawObject, "volume", 0.0),
			Entry("3.0", rawObject, "volume", 3.0),
			Entry("no decimal", rawObject, "volume", 2),
		)

	})

	Context("tubing volume", func() {

		BeforeEach(func() {
			rawObject["type"] = "deviceEvent"
			rawObject["subType"] = "prime"
			rawObject["primeTarget"] = "tubing"
		})

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("less than 0", rawObject, "volume", -0.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(-0.1, 0.0, 100.0), "/volume")},
			),
			Entry("more than 100", rawObject, "volume", 100.1,
				[]*service.Error{testing.SetExpectedErrorSource(validator.ErrorFloatNotInRange(100.1, 0.0, 100.0), "/volume")},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("0", rawObject, "volume", 0.0),
			Entry("100.0", rawObject, "volume", 100.0),
			Entry("no decimal", rawObject, "volume", 55),
		)

	})

})
