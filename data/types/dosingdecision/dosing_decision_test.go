package dosingdecision_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types/settings/pump"

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
	datum.Type = dosingdecision.Type
	datum.Units = pump.NewUnits()
	datum.Units.BloodGlucose = pointer.FromString("mmol/L")
	datum.Units.Carbohydrate = pointer.FromString(test.RandomStringFromArray(pump.Carbohydrates()))
	return datum
}

func CloneDosingDecision(datum *dosingdecision.DosingDecision) *dosingdecision.DosingDecision {
	if datum == nil {
		return nil
	}
	clone := dosingdecision.NewDosingDecision()

	clone.Units = pump.CloneUnits(datum.Units)
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

				Entry("blood glucose target schedule invalid",
					func(datum *dosingdecision.DosingDecision) {
						invalidBloodGlucoseTargetSchedule := pump.NewBloodGlucoseTargetStartArrayTest(pointer.FromString("mmol/L"))
						(*invalidBloodGlucoseTargetSchedule)[0].Start = nil
						datum.GlucoseTargetRangeSchedule = invalidBloodGlucoseTargetSchedule
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/pumpManagerStatus/0/start", NewMeta()),
				),
				Entry("blood glucose target schedule valid",
					func(datum *dosingdecision.DosingDecision) {
						datum.GlucoseTargetRangeSchedule = pump.NewBloodGlucoseTargetStartArrayTest(pointer.FromString("mmol/L"))
					},
				),

				Entry("units missing",
					func(datum *dosingdecision.DosingDecision) { datum.Units = nil },
				),
				Entry("units invalid",
					func(datum *dosingdecision.DosingDecision) { datum.Units.BloodGlucose = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units/bg", NewMeta()),
				),
				Entry("units valid",
					func(datum *dosingdecision.DosingDecision) {
						datum.Units = pump.NewUnits()
						datum.Units.BloodGlucose = pointer.FromString("mmol/L")
						datum.Units.Carbohydrate = pointer.FromString(test.RandomStringFromArray(pump.Carbohydrates()))
					},
				),
			)
		})
	})
})
