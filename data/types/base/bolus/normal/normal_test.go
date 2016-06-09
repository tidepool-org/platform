package normal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/data/types/base/bolus"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/service"
)

func NewRawObject() map[string]interface{} {
	rawObject := testing.RawBaseObject()
	rawObject["type"] = "bolus"
	rawObject["subType"] = "normal"
	rawObject["normal"] = 7.6
	return rawObject
}

func NewExpectedRawObject() map[string]interface{} {
	rawObject := NewRawObject()
	rawObject["expectedNormal"] = 8.9
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
			Entry("does not exist", NewRawObject(), "normal", nil,
				[]*service.Error{testing.ComposeError(validator.ErrorValueNotExists(), "/normal", NewMeta())},
			),
			Entry("is not a number", NewRawObject(), "normal", "not-a-number",
				[]*service.Error{
					testing.ComposeError(parser.ErrorTypeNotFloat("not-a-number"), "/normal", NewMeta()),
					testing.ComposeError(validator.ErrorValueNotExists(), "/normal", NewMeta()),
				},
			),
			Entry("is less than lower limit", NewRawObject(), "normal", -0.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(-0.1, 0.0, 100.0), "/normal", NewMeta())},
			),
			Entry("is greater than upper limit", NewRawObject(), "normal", 100.1,
				[]*service.Error{testing.ComposeError(validator.ErrorFloatNotInRange(100.1, 0.0, 100.0), "/normal", NewMeta())},
			),
			Entry("is zero without expectedNormal", NewRawObject(), "normal", 0.0,
				[]*service.Error{testing.ComposeError(validator.ErrorValueNotExists(), "/expectedNormal", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is approaching lower limit", NewRawObject(), "normal", 0.01),
			Entry("is within lower and upper limit", NewRawObject(), "normal", 14.5),
			Entry("is at upper limit", NewRawObject(), "normal", 100.0),
			Entry("is without decimal", NewRawObject(), "normal", 14),
		)
	})

	Context("expectedNormal", func() {
		DescribeTable("invalid when", testing.ExpectFieldNotValid,
			Entry("is not a number", NewExpectedRawObject(), "expectedNormal", "not-a-number",
				[]*service.Error{testing.ComposeError(parser.ErrorTypeNotFloat("not-a-number"), "/expectedNormal", NewMeta())},
			),
			Entry("is less than or equal to normal", NewExpectedRawObject(), "expectedNormal", 7.6,
				[]*service.Error{testing.ComposeError(validator.ErrorValueNotGreaterThan(7.6, 7.6), "/expectedNormal", NewMeta())},
			),
			Entry("is greater than upper limit", NewExpectedRawObject(), "expectedNormal", 100.1,
				[]*service.Error{testing.ComposeError(validator.ErrorValueNotLessThanOrEqualTo(100.1, 100.0), "/expectedNormal", NewMeta())},
			),
		)

		DescribeTable("valid when", testing.ExpectFieldIsValid,
			Entry("is approaching normal", NewExpectedRawObject(), "expectedNormal", 7.61),
			Entry("is at upper limit", NewExpectedRawObject(), "expectedNormal", 100.0),
			Entry("is without decimal", NewExpectedRawObject(), "expectedNormal", 14),
		)
	})
})
