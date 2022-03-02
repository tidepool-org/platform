package controller_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesStatusController "github.com/tidepool-org/platform/data/types/status/controller"
	dataTypesStatusControllerTest "github.com/tidepool-org/platform/data/types/status/controller/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &dataTypes.Meta{
		Type: "controllerStatus",
	}
}

var _ = Describe("Controller", func() {
	It("Type is expected", func() {
		Expect(dataTypesStatusController.Type).To(Equal("controllerStatus"))
	})

	Context("Controller", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesStatusController.Controller)) {
				datum := dataTypesStatusControllerTest.RandomController()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesStatusControllerTest.NewObjectFromController(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesStatusControllerTest.NewObjectFromController(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesStatusController.Controller) {},
			),
			Entry("empty",
				func(datum *dataTypesStatusController.Controller) {
					*datum = *dataTypesStatusController.New()
				},
			),
			Entry("all",
				func(datum *dataTypesStatusController.Controller) {
					datum.Battery = dataTypesStatusControllerTest.RandomBattery()
				},
			),
		)

		Context("New", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesStatusController.New()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Type).To(Equal("controllerStatus"))
				Expect(datum.Battery).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesStatusController.Controller), expectedErrors ...error) {
					expectedDatum := dataTypesStatusControllerTest.RandomControllerForParser()
					object := dataTypesStatusControllerTest.NewObjectFromController(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesStatusController.New()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusController.Controller) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusController.Controller) {
						object["battery"] = true
						expectedDatum.Battery = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/battery", NewMeta()),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesStatusController.Controller), expectedErrors ...error) {
					datum := dataTypesStatusControllerTest.RandomController()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesStatusController.Controller) {},
				),
				Entry("type missing",
					func(datum *dataTypesStatusController.Controller) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypes.Meta{Type: ""}),
				),
				Entry("type invalid",
					func(datum *dataTypesStatusController.Controller) {
						datum.Type = "invalidType"
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "controllerStatus"), "/type", &dataTypes.Meta{Type: "invalidType"}),
				),
				Entry("type controllerStatus",
					func(datum *dataTypesStatusController.Controller) {
						datum.Type = "controllerStatus"
					},
				),
				Entry("battery missing",
					func(datum *dataTypesStatusController.Controller) { datum.Battery = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/battery", NewMeta()),
				),
				Entry("battery invalid",
					func(datum *dataTypesStatusController.Controller) {
						datum.Battery.State = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", dataTypesStatusController.BatteryStates()), "/battery/state", NewMeta()),
				),
				Entry("battery valid",
					func(datum *dataTypesStatusController.Controller) {
						datum.Battery = dataTypesStatusControllerTest.RandomBattery()
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesStatusController.Controller) {
						datum.Type = "invalidType"
						datum.Battery.State = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "controllerStatus"), "/type", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", dataTypesStatusController.BatteryStates()), "/battery/state", &dataTypes.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesStatusController.Controller), expectator func(datum *dataTypesStatusController.Controller, expectedDatum *dataTypesStatusController.Controller)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesStatusControllerTest.RandomController()
						mutator(datum)
						expectedDatum := dataTypesStatusControllerTest.CloneController(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesStatusController.Controller) {},
					nil,
				),
			)
		})
	})
})
