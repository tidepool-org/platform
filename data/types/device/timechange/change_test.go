package timechange_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesDeviceTimechange "github.com/tidepool-org/platform/data/types/device/timechange"
	dataTypesDeviceTimechangeTest "github.com/tidepool-org/platform/data/types/device/timechange/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"

	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Change", func() {
	It("AgentAutomatic is expected", func() {
		Expect(dataTypesDeviceTimechange.AgentAutomatic).To(Equal("automatic"))
	})

	It("AgentManual is expected", func() {
		Expect(dataTypesDeviceTimechange.AgentManual).To(Equal("manual"))
	})

	It("FromTimeFormat is expected", func() {
		Expect(dataTypesDeviceTimechange.FromTimeFormat).To(Equal("2006-01-02T15:04:05"))
	})

	It("ToTimeFormat is expected", func() {
		Expect(dataTypesDeviceTimechange.ToTimeFormat).To(Equal("2006-01-02T15:04:05"))
	})

	It("Agents returns expected", func() {
		Expect(dataTypesDeviceTimechange.Agents()).To(Equal([]string{"automatic", "manual"}))
	})

	Context("ParseChange", func() {
		// TODO
	})

	Context("NewChange", func() {
		It("is successful", func() {
			Expect(dataTypesDeviceTimechange.NewChange()).To(Equal(&dataTypesDeviceTimechange.Change{}))
		})
	})

	Context("Change", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesDeviceTimechange.Change), expectedErrors ...error) {
					datum := dataTypesDeviceTimechangeTest.RandomChange()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesDeviceTimechange.Change) {},
				),
				Entry("agent missing",
					func(datum *dataTypesDeviceTimechange.Change) { datum.Agent = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/agent"),
				),
				Entry("agent invalid",
					func(datum *dataTypesDeviceTimechange.Change) { datum.Agent = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"automatic", "manual"}), "/agent"),
				),
				Entry("agent automatic",
					func(datum *dataTypesDeviceTimechange.Change) { datum.Agent = pointer.FromString("automatic") },
				),
				Entry("agent manual",
					func(datum *dataTypesDeviceTimechange.Change) { datum.Agent = pointer.FromString("manual") },
				),
				Entry("from missing",
					func(datum *dataTypesDeviceTimechange.Change) { datum.From = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/from"),
				),
				Entry("from invalid",
					func(datum *dataTypesDeviceTimechange.Change) { datum.From = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", "2006-01-02T15:04:05"), "/from"),
				),
				Entry("from valid",
					func(datum *dataTypesDeviceTimechange.Change) {
						datum.From = pointer.FromString(test.RandomTime().Format("2006-01-02T15:04:05"))
					},
				),
				Entry("to missing",
					func(datum *dataTypesDeviceTimechange.Change) { datum.To = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/to"),
				),
				Entry("to invalid",
					func(datum *dataTypesDeviceTimechange.Change) { datum.To = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", "2006-01-02T15:04:05"), "/to"),
				),
				Entry("to valid",
					func(datum *dataTypesDeviceTimechange.Change) {
						datum.To = pointer.FromString(test.RandomTime().Format("2006-01-02T15:04:05"))
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesDeviceTimechange.Change) {
						datum.Agent = nil
						datum.From = nil
						datum.To = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/agent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/from"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/to"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesDeviceTimechange.Change)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesDeviceTimechangeTest.RandomChange()
						mutator(datum)
						expectedDatum := dataTypesDeviceTimechangeTest.CloneChange(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesDeviceTimechange.Change) {},
				),
				Entry("does not modify the datum; agent missing",
					func(datum *dataTypesDeviceTimechange.Change) { datum.Agent = nil },
				),
				Entry("does not modify the datum; agent automatic",
					func(datum *dataTypesDeviceTimechange.Change) { datum.Agent = pointer.FromString("automatic") },
				),
				Entry("does not modify the datum; agent manual",
					func(datum *dataTypesDeviceTimechange.Change) { datum.Agent = pointer.FromString("manual") },
				),
				Entry("does not modify the datum; from missing",
					func(datum *dataTypesDeviceTimechange.Change) { datum.From = nil },
				),
				Entry("does not modify the datum; to missing",
					func(datum *dataTypesDeviceTimechange.Change) { datum.To = nil },
				),
			)
		})
	})
})
