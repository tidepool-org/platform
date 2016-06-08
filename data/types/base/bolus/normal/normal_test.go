package normal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/base/bolus"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

func NewRawObject() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "bolus"
	rawObject["subType"] = "normal"
	return rawObject
}

func NewNormalRawObject() map[string]interface{} {
	rawObject := NewRawObject()
	rawObject["normal"] = 52.1
	return rawObject
}

func NewExpectedNormalRawObject() map[string]interface{} {
	rawObject := NewRawObject()
	rawObject["normal"] = 0.0
	rawObject["expectedNormal"] = 52.1
	return rawObject
}

func NewMeta() interface{} {
	return &bolus.Meta{
		Type:    "bolus",
		SubType: "normal",
	}
}

var _ = Describe("Normal", func() {
	Context("normal", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("does not exist", NewNormalRawObject(), "normal", nil,
				[]*service.Error{testing.ComposeError(validator.ErrorValueNotExists(), "/normal", NewMeta())},
			),
			Entry("is less than lower limit", NewNormalRawObject(), "normal", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta())},
			),
			Entry("is zero without expectedNormal", NewNormalRawObject(), "normal", 0.0,
				[]*service.Error{testing.ComposeError(validator.ErrorValueNotExists(), "/expectedNormal", NewMeta())},
			),
			Entry("is greater than upper limit", NewNormalRawObject(), "normal", 100.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is approaching lower limit", NewNormalRawObject(), "normal", 0.01),
			Entry("is within lower and upper limit", NewNormalRawObject(), "normal", 25.5),
			Entry("is at upper limit", NewNormalRawObject(), "normal", 100.0),
			Entry("is without decimal", NewNormalRawObject(), "normal", 50),
		)
	})

	Context("expectedNormal", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("does not exist", NewExpectedNormalRawObject(), "expectedNormal", nil,
				[]*service.Error{testing.ComposeError(validator.ErrorValueNotExists(), "/expectedNormal", NewMeta())},
			),
			Entry("is less than lower limit", NewExpectedNormalRawObject(), "expectedNormal", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorValueNotGreaterThan(-0.1, 0.0), "/expectedNormal", NewMeta())},
			),
			Entry("is zero", NewExpectedNormalRawObject(), "expectedNormal", 0.0,
				[]*service.Error{testing.ComposeError(validator.ErrorValueNotGreaterThan(0.0, 0.0), "/expectedNormal", NewMeta())},
			),
			Entry("is greater than upper limit", NewExpectedNormalRawObject(), "expectedNormal", 100.1,
				[]*service.Error{testing.ComposeError(validator.ErrorValueNotLessThanOrEqualTo(100.1, 100.0), "/expectedNormal", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is approaching lower limit", NewExpectedNormalRawObject(), "expectedNormal", 0.01),
			Entry("is within lower and upper limit", NewExpectedNormalRawObject(), "expectedNormal", 25.5),
			Entry("is at upper limit", NewExpectedNormalRawObject(), "expectedNormal", 100.0),
			Entry("is without decimal", NewExpectedNormalRawObject(), "expectedNormal", 50),
		)
	})
})
