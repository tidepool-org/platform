package normal_test

import (
	"github.com/tidepool-org/platform/pvn/data/types/base/testing"
	"github.com/tidepool-org/platform/pvn/data/validator"
	"github.com/tidepool-org/platform/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Normal Bolus", func() {

	var rawObject = testing.RawBaseObject()

	BeforeEach(func() {

		rawObject["type"] = "bolus"
		rawObject["subType"] = "normal"
		rawObject["normal"] = 52.1

	})

	Context("normal", func() {

		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("negative", rawObject, "normal", -0.1, []*service.Error{validator.ErrorValueNotTrue()}),
			Entry("greater than 20", rawObject, "normal", 100.1, []*service.Error{validator.ErrorValueNotTrue()}),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("within bounds", rawObject, "normal", 25.5),
			Entry("also without decimal", rawObject, "normal", 50),
		)

	})

})
