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
	structureParser "github.com/tidepool-org/platform/structure/parser"
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
			Expect(datum.GUID).To(BeNil())
			Expect(datum.LinkedBolus).To(BeNil())
		})
	})

	Context("Normal", func() {
		Context("Parse", func() {
			var parsedBase *biphasic.Biphasic
			It("parses eventId when biphasicId is missing", func() {
				parsedBase = biphasic.New()
				object := map[string]interface{}{"eventId": "1234"}
				parser := structureParser.NewObject(&object)
				parsedBase.Parse(parser)
				Expect(parsedBase.GUID).To(BeNil())
				Expect(*parsedBase.BiphasicID).To(Equal("1234"))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
			It("parses eventId when biphasicId is empty", func() {
				parsedBase = biphasic.New()
				object := map[string]interface{}{"eventId": "1234", "biphasicId": ""}
				parser := structureParser.NewObject(&object)
				parsedBase.Parse(parser)
				Expect(parsedBase.GUID).To(BeNil())
				Expect(*parsedBase.BiphasicID).To(Equal("1234"))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
			It("doesn't parses eventId when biphasicId is not empty", func() {
				parsedBase = biphasic.New()
				object := map[string]interface{}{"eventId": "1234", "biphasicId": "4567"}
				parser := structureParser.NewObject(&object)
				parsedBase.Parse(parser)
				Expect(*parsedBase.GUID).To(Equal("1234"))
				Expect(*parsedBase.BiphasicID).To(Equal("4567"))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
			It("parses biphasicId", func() {
				parsedBase = biphasic.New()
				object := map[string]interface{}{"biphasicId": "4567"}
				parser := structureParser.NewObject(&object)
				parsedBase.Parse(parser)
				Expect(parsedBase.GUID).To(BeNil())
				Expect(*parsedBase.BiphasicID).To(Equal("4567"))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
			It("parses without error with empty biphasicId and eventId", func() {
				parsedBase = biphasic.New()
				object := map[string]interface{}{}
				parser := structureParser.NewObject(&object)
				parsedBase.Parse(parser)
				Expect(parsedBase.GUID).To(BeNil())
				Expect(parsedBase.BiphasicID).To(BeNil())
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
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
				Entry("normal missing",
					func(datum *biphasic.Biphasic) {
						datum.Normal.Normal = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
				),
				Entry("normal missing; normal expected missing",
					func(datum *biphasic.Biphasic) {
						datum.Normal.Normal = nil
						datum.Normal.NormalExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/normal", NewMeta()),
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
				Entry("GUID missing",
					func(datum *biphasic.Biphasic) {
						datum.BiphasicID = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/biphasicId", NewMeta()),
				),
				Entry("Multiple errors",
					func(datum *biphasic.Biphasic) {
						datum.Part = nil
						datum.BiphasicID = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/part", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/biphasicId", NewMeta()),
				),
			)
		})
	})
})
