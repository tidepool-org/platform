package basalsecurity_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types"
	basalsecurity "github.com/tidepool-org/platform/data/types/settings/basalsecurity"
	dataTypesSettingsPumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "basalSecurity",
	}
}

func NewBasalSecurity() *basalsecurity.BasalSecurity {
	datum := basalsecurity.New()
	datum.Base = *dataTypesTest.NewBase()
	datum.Type = "basalSecurity"
	datum.BasalRateSchedule = dataTypesSettingsPumpTest.NewBasalRateStartArray()
	return datum
}

var _ = Describe("BasalSecurity", func() {
	It("Type is expected", func() {
		Expect(basalsecurity.Type).To(Equal("basalSecurity"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := basalsecurity.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("basalSecurity"))
			Expect(datum.BasalRateSchedule).To(BeNil())
		})
	})

	Context("Basal Security", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *basalsecurity.BasalSecurity), expectedErrors ...error) {
					datum := NewBasalSecurity()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *basalsecurity.BasalSecurity) {},
				),
				Entry("type missing",
					func(datum *basalsecurity.BasalSecurity) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					func(datum *basalsecurity.BasalSecurity) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basalSecurity"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type basalSecurity",
					func(datum *basalsecurity.BasalSecurity) { datum.Type = "basalSecurity" },
				),
				Entry("basal rate schedule missing",
					func(datum *basalsecurity.BasalSecurity) {
						datum.BasalRateSchedule = nil
					},
				),
				Entry("basal rate schedule invalid",
					func(datum *basalsecurity.BasalSecurity) {
						invalidBasalRateSchedule := dataTypesSettingsPumpTest.NewBasalRateStartArray()
						(*invalidBasalRateSchedule)[0].Start = nil
						datum.BasalRateSchedule = invalidBasalRateSchedule
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalSchedule/0/start", NewMeta()),
				),
				Entry("basal rate schedule valid",
					func(datum *basalsecurity.BasalSecurity) {
						datum.BasalRateSchedule = dataTypesSettingsPumpTest.NewBasalRateStartArray()
					},
				),

				Entry("multiple errors",
					func(datum *basalsecurity.BasalSecurity) {
						datum.Type = "invalidType"
						invalidBasalRateSchedule := dataTypesSettingsPumpTest.NewBasalRateStartArray()
						(*invalidBasalRateSchedule)[0].Start = nil
						datum.BasalRateSchedule = invalidBasalRateSchedule
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basalSecurity"), "/type", &types.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalSchedule/0/start", &types.Meta{Type: "invalidType"}),
				),
			)
		})

	})
})
