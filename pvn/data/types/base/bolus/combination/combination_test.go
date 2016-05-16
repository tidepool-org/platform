package combination_test

import (
	"github.com/tidepool-org/platform/pvn/data/types/base/testing"
	"github.com/tidepool-org/platform/pvn/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Combination Bolus", func() {

	var rawObject = testing.RawBaseObject()

	BeforeEach(func() {

		rawObject["type"] = "bolus"
		rawObject["subType"] = "dual/square"
		rawObject["duration"] = 0
		rawObject["extended"] = 25.5
		rawObject["normal"] = 55.5

	})

	Context("duration", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "duration", -1, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("greater than 86400000", rawObject, "duration", 86400001, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "duration", 2400),
		)

	})

	Context("extended", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "extended", -0.1, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("zero", rawObject, "extended", 0.0, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("greater than 100", rawObject, "extended", 100.1, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "extended", 5.5),
			Entry("also without decimal", rawObject, "extended", 5),
		)

	})

	Context("normal", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "normal", -0.1, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("zero", rawObject, "normal", 0.0, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("greater than 100", rawObject, "normal", 100.1, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "normal", 25.5),
			Entry("also without decimal", rawObject, "normal", 50),
		)

	})

})
