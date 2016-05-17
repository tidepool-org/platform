package calculator_test

import (
	"github.com/tidepool-org/platform/pvn/data/types/base/testing"
	"github.com/tidepool-org/platform/pvn/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Calculator Bolus", func() {

	var rawObject = testing.RawBaseObject()

	BeforeEach(func() {

		rawObject["type"] = "wizard"
		rawObject["units"] = "mg/dl"
		rawObject["bgInput"] = 100
		rawObject["carbInput"] = 120
		rawObject["insulinSensitivity"] = 90
		rawObject["insulinCarbRatio"] = 50
		rawObject["insulinOnBoard"] = 70

	})

	Context("insulinOnBoard", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "insulinOnBoard", -1, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("greater than 250", rawObject, "insulinOnBoard", 251, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "insulinOnBoard", 99),
		)

	})

	Context("insulinCarbRatio", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "insulinCarbRatio", -1, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("greater than 250", rawObject, "insulinCarbRatio", 251, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "insulinCarbRatio", 99),
		)

	})

	Context("units", func() {

		DescribeTable("units when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "units", "", []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("not one of the predefined values", rawObject, "units", "wrong", []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("mmol/l", rawObject, "units", "mmol/l"),
			Entry("mmol/L", rawObject, "units", "mmol/L"),
			Entry("mg/dl", rawObject, "units", "mg/dl"),
			Entry("mg/dL", rawObject, "units", "mg/dL"),
		)

	})

	Context("bgInput", func() {

		DescribeTable("value when", testing.ExpectFieldNotValid,
			Entry("less than 0", rawObject, "bgInput", -0.1, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("greater than 1000", rawObject, "bgInput", 1000.1, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("0", rawObject, "bgInput", 0.0),
			Entry("above 0", rawObject, "bgInput", 0.1),
			Entry("below 1000", rawObject, "bgInput", 999.99),
			Entry("as integer", rawObject, "bgInput", 4),
		)

	})

	Context("insulinSensitivity", func() {

		DescribeTable("value when", testing.ExpectFieldNotValid,
			Entry("less than 0", rawObject, "insulinSensitivity", -0.1, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("greater than 1000", rawObject, "insulinSensitivity", 1000.1, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("0", rawObject, "insulinSensitivity", 0.0),
			Entry("above 0", rawObject, "insulinSensitivity", 0.1),
			Entry("below 1000", rawObject, "insulinSensitivity", 999.99),
			Entry("as integer", rawObject, "insulinSensitivity", 4),
		)

	})

	Context("carbInput", func() {

		DescribeTable("value when", testing.ExpectFieldNotValid,
			Entry("less than 0", rawObject, "carbInput", -1, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("greater than 1000", rawObject, "carbInput", 1001, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("0", rawObject, "carbInput", 0),
			Entry("in range", rawObject, "carbInput", 250),
			Entry("below 1000", rawObject, "carbInput", 999),
		)

	})

})
