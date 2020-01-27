package dosingdecision_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"

	"github.com/tidepool-org/platform/structure"

	"github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	InvalidType  = "invalidType"
	ValidVersion = "1.0"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: dosingdecision.Type,
	}
}

func NewDosingDecision() *dosingdecision.DosingDecision {
	datum := dosingdecision.NewDosingDecision()
	datum.Base = *dataTypesTest.NewBase()
	datum.DeviceType = pointer.FromString(test.RandomStringFromArray(dosingdecision.DeviceTypes()))
	datum.Type = dosingdecision.Type
	datum.Version = pointer.FromString(ValidVersion)
	return datum
}

func CloneDeviceStatus(datum *dosingdecision.DosingDecision) *dosingdecision.DosingDecision {
	if datum == nil {
		return nil
	}
	clone := dosingdecision.NewDosingDecision()
	return clone
}

var _ = Describe("DosingDecision", func() {

	Context("DosingDecision", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *dosingdecision.DosingDecision), expectedErrors ...error) {
					datum := NewDosingDecision()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dosingdecision.DosingDecision) {},
				),
				Entry("invalid Device Type",
					func(datum *dosingdecision.DosingDecision) {
						datum.DeviceType = pointer.FromString(InvalidType)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf(InvalidType, dosingdecision.DeviceTypes()), "/deviceType", NewMeta()),
				),
				Entry("version missing",
					func(datum *dosingdecision.DosingDecision) { datum.Version = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/version", NewMeta()),
				),
			)
		})
	})
})
