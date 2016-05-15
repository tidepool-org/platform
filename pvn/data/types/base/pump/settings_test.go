package pump_test

import (
	"github.com/tidepool-org/platform/pvn/data/types/base/testing"
	"github.com/tidepool-org/platform/pvn/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Pump Settings", func() {

	var rawObject = testing.RawBaseObject()

	BeforeEach(func() {

		rawObject["type"] = "pumpSettings"
		rawObject["activeSchedule"] = "standard"

		rawObject["units"] = map[string]interface{}{
			"carb": "grams",
			"bg":   "mmol/L",
		}

		rawObject["carbRatio"] = []interface{}{
			map[string]interface{}{"amount": 12, "start": 0},
			map[string]interface{}{"amount": 10, "start": 21600000},
		}

		rawObject["bgTarget"] = []interface{}{
			map[string]interface{}{"low": 5.5, "high": 6.7, "start": 0},
			map[string]interface{}{"low": 5.0, "high": 6.1, "start": 18000000},
		}

		rawObject["insulinSensitivity"] = []interface{}{
			map[string]interface{}{"amount": 3.6, "start": 0},
			map[string]interface{}{"amount": 2.5, "start": 18000000},
		}

		/*rawObject["basalSchedules"] = map[string][]map[string]interface{}{
			"standard":  {{"rate": 0.8, "start": 0}, {"rate": 0.75, "start": 3600000}},
			"pattern a": {{"rate": 0.95, "start": 0}, {"rate": 0.9, "start": 3600000}},
		}*/

	})

	Context("activeSchedule", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("empty", rawObject, "activeSchedule", "", []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("more than 1 characters", rawObject, "activeSchedule", "A"),
			Entry("freetext", rawObject, "activeSchedule", "standard"),
		)

	})

	Context("units", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("bg empty", rawObject, "units", map[string]interface{}{"carb": "grams", "bg": ""}, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("bg not predefined type", rawObject, "units", map[string]interface{}{"carb": "grams", "bg": "na"}, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("carb empty", rawObject, "units", map[string]interface{}{"carb": "", "bg": "mmol/L"}, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("carbs set and bg set as mmol/L", rawObject, "units", map[string]interface{}{"carb": "grams", "bg": "mmol/L"}),
			Entry("carbs set and bg set as mg/dl", rawObject, "units", map[string]interface{}{"carb": "grams", "bg": "mg/dl"}),
		)

	})

	Context("carbRatio", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("start negative", rawObject, "carbRatio",
				[]interface{}{map[string]interface{}{"amount": 12, "start": -1}},
				[]*service.Error{validator.ErrorValueNotTrue()},
			),
			Entry("start greater then 86400000", rawObject, "carbRatio",
				[]interface{}{map[string]interface{}{"amount": 12, "start": 86400001}},
				[]*service.Error{validator.ErrorValueNotTrue()},
			),
			Entry("amount negative", rawObject, "carbRatio",
				[]interface{}{map[string]interface{}{"amount": -1, "start": 21600000}},
				[]*service.Error{validator.ErrorValueNotTrue()},
			),
			Entry("amount greater then 250", rawObject, "carbRatio",
				[]interface{}{map[string]interface{}{"amount": 251, "start": 21600000}},
				[]*service.Error{validator.ErrorValueNotTrue()},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("start and amount within bounds", rawObject, "carbRatio",
				[]interface{}{map[string]interface{}{"amount": 12.0, "start": 0}},
			),
		)

	})

})
