package biphasic_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/bolus/biphasic"
	dataTypesBolusBiphasicTest "github.com/tidepool-org/platform/data/types/bolus/biphasic/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &bolus.Meta{
		Type:    "bolus",
		SubType: "biphasic",
	}
}

var _ = Describe("Normal", func() {
	It("SubType is expected", func() {
		Expect(biphasic.SubType).To(Equal("biphasic"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := biphasic.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("bolus"))
			Expect(datum.SubType).To(Equal("biphasic"))
			Expect(datum.Normal).ToNot(BeNil())
			Expect(datum.NormalExpected).To(BeNil())
			Expect(datum.Part).To(BeNil())
			Expect(datum.EventID).To(BeNil())
			Expect(datum.LinkedBolus).To(BeNil())
		})
	})

	Context("Normal", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *biphasic.Biphasic), expectedErrors ...error) {
					datum := dataTypesBolusBiphasicTest.NewBiphasic()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *biphasic.Biphasic) {},
				),
				Entry("type missing",
					func(datum *biphasic.Biphasic) {
						datum.Type = ""
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &bolus.Meta{SubType: "biphasic"}),
				),
				Entry("type invalid",
					func(datum *biphasic.Biphasic) {
						datum.Type = "invalidType"
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "bolus"), "/type", &bolus.Meta{Type: "invalidType", SubType: "biphasic"}),
				),
				Entry("Part invalid",
					func(datum *biphasic.Biphasic) {
						datum.Part = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", biphasic.Parts()), "/part", NewMeta()),
				),
				Entry("linked bolus missing",
					func(datum *biphasic.Biphasic) {
						datum.LinkedBolus = nil
					},
				),
				Entry("Part missing",
					func(datum *biphasic.Biphasic) {
						datum.Part = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/part", NewMeta()),
				),
				Entry("EventID missing",
					func(datum *biphasic.Biphasic) {
						datum.EventID = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/eventId", NewMeta()),
				),
				Entry("Multiple errors",
					func(datum *biphasic.Biphasic) {
						datum.Part = nil
						datum.EventID = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/part", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/eventId", NewMeta()),
				),
			)
		})
	})
})
