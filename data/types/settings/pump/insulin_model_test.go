package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesSettingsPumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("InsulinModel", func() {
	It("InsulinModelActionDelayMaximum is expected", func() {
		Expect(dataTypesSettingsPump.InsulinModelActionDelayMaximum).To(Equal(86400))
	})

	It("InsulinModelActionDelayMinimum is expected", func() {
		Expect(dataTypesSettingsPump.InsulinModelActionDelayMinimum).To(Equal(0))
	})

	It("InsulinModelActionDurationMaximum is expected", func() {
		Expect(dataTypesSettingsPump.InsulinModelActionDurationMaximum).To(Equal(86400))
	})

	It("InsulinModelActionDurationMinimum is expected", func() {
		Expect(dataTypesSettingsPump.InsulinModelActionDurationMinimum).To(Equal(0))
	})

	It("InsulinModelActionPeakOffsetMaximum is expected", func() {
		Expect(dataTypesSettingsPump.InsulinModelActionPeakOffsetMaximum).To(Equal(86400))
	})

	It("InsulinModelActionPeakOffsetMinimum is expected", func() {
		Expect(dataTypesSettingsPump.InsulinModelActionPeakOffsetMinimum).To(Equal(0))
	})

	It("InsulinModelModelTypeFiasp is expected", func() {
		Expect(dataTypesSettingsPump.InsulinModelModelTypeFiasp).To(Equal("fiasp"))
	})

	It("InsulinModelModelTypeOther is expected", func() {
		Expect(dataTypesSettingsPump.InsulinModelModelTypeOther).To(Equal("other"))
	})

	It("InsulinModelModelTypeOtherLengthMaximum is expected", func() {
		Expect(dataTypesSettingsPump.InsulinModelModelTypeOtherLengthMaximum).To(Equal(100))
	})

	It("InsulinModelModelTypeRapidAdult is expected", func() {
		Expect(dataTypesSettingsPump.InsulinModelModelTypeRapidAdult).To(Equal("rapidAdult"))
	})

	It("InsulinModelModelTypeRapidChild is expected", func() {
		Expect(dataTypesSettingsPump.InsulinModelModelTypeRapidChild).To(Equal("rapidChild"))
	})

	It("InsulinModelModelTypeWalsh is expected", func() {
		Expect(dataTypesSettingsPump.InsulinModelModelTypeWalsh).To(Equal("walsh"))
	})

	It("InsulinModelModelTypes returns expected", func() {
		Expect(dataTypesSettingsPump.InsulinModelModelTypes()).To(Equal([]string{"fiasp", "other", "rapidAdult", "rapidChild", "walsh"}))
	})

	Context("ParseInsulinModel", func() {
		// TODO
	})

	Context("NewInsulinModel", func() {
		It("is successful", func() {
			Expect(dataTypesSettingsPump.NewInsulinModel()).To(Equal(&dataTypesSettingsPump.InsulinModel{}))
		})
	})

	Context("InsulinModel", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsPump.InsulinModel), expectedErrors ...error) {
					datum := dataTypesSettingsPumpTest.RandomInsulinModel()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsPump.InsulinModel) {},
				),
				Entry("model type missing",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ModelType = nil
						datum.ModelTypeOther = nil
					},
				),
				Entry("model type invalid",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ModelType = pointer.FromString("invalid")
						datum.ModelTypeOther = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"fiasp", "other", "rapidAdult", "rapidChild", "walsh"}), "/modelType"),
				),
				Entry("model type fiasp",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ModelType = pointer.FromString("fiasp")
						datum.ModelTypeOther = nil
					},
				),
				Entry("model type other",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ModelType = pointer.FromString("other")
						datum.ModelTypeOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("model type rapid adult",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ModelType = pointer.FromString("rapidAdult")
						datum.ModelTypeOther = nil
					},
				),
				Entry("model type rapid child",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ModelType = pointer.FromString("rapidChild")
						datum.ModelTypeOther = nil
					},
				),
				Entry("model type walsh",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ModelType = pointer.FromString("walsh")
						datum.ModelTypeOther = nil
					},
				),
				Entry("model type other exists",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ModelType = pointer.FromString("fiasp")
						datum.ModelTypeOther = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/modelTypeOther"),
				),
				Entry("model type other missing",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ModelType = pointer.FromString("other")
						datum.ModelTypeOther = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/modelTypeOther"),
				),
				Entry("model type other empty",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ModelType = pointer.FromString("other")
						datum.ModelTypeOther = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/modelTypeOther"),
				),
				Entry("model type other length in range (upper)",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ModelType = pointer.FromString("other")
						datum.ModelTypeOther = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("model type other length out of range (upper)",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ModelType = pointer.FromString("other")
						datum.ModelTypeOther = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/modelTypeOther"),
				),
				Entry("action delay out of range (lower)",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDelay = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400), "/actionDelay"),
				),
				Entry("action delay in range (lower)",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDelay = pointer.FromInt(0)
					},
				),
				Entry("action delay in range (upper)",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDelay = pointer.FromInt(86400)
					},
				),
				Entry("action delay out of range (upper)",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDelay = pointer.FromInt(86401)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(86401, 0, 86400), "/actionDelay"),
				),
				Entry("action duration out of range (lower)",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDuration = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400), "/actionDuration"),
				),
				Entry("action duration in range (lower)",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDuration = pointer.FromInt(0)
						datum.ActionPeakOffset = pointer.FromInt(0)
					},
				),
				Entry("action duration in range (upper)",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDuration = pointer.FromInt(86400)
					},
				),
				Entry("action duration out of range (upper)",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDuration = pointer.FromInt(86401)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(86401, 0, 86400), "/actionDuration"),
				),
				Entry("action peak offset out of range (lower); action duration missing",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDuration = nil
						datum.ActionPeakOffset = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400), "/actionPeakOffset"),
				),
				Entry("action peak offset in range (lower); action duration missing",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDuration = nil
						datum.ActionPeakOffset = pointer.FromInt(0)
					},
				),
				Entry("action peak offset in range (upper); action duration missing",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDuration = nil
						datum.ActionPeakOffset = pointer.FromInt(86400)
					},
				),
				Entry("action peak offset out of range (upper); action duration missing",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDuration = nil
						datum.ActionPeakOffset = pointer.FromInt(86401)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(86401, 0, 86400), "/actionPeakOffset"),
				),
				Entry("action peak offset out of range (lower); action duration exists",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDuration = pointer.FromInt(43200)
						datum.ActionPeakOffset = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 43200), "/actionPeakOffset"),
				),
				Entry("action peak offset in range (lower); action duration exists",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDuration = pointer.FromInt(43200)
						datum.ActionPeakOffset = pointer.FromInt(0)
					},
				),
				Entry("action peak offset in range (upper); action duration exists",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDuration = pointer.FromInt(43200)
						datum.ActionPeakOffset = pointer.FromInt(43200)
					},
				),
				Entry("action peak offset out of range (upper); action duration exists",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ActionDuration = pointer.FromInt(43200)
						datum.ActionPeakOffset = pointer.FromInt(43201)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(43201, 0, 43200), "/actionPeakOffset"),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsPump.InsulinModel) {
						datum.ModelType = pointer.FromString("invalid")
						datum.ModelTypeOther = pointer.FromString(test.RandomStringFromRange(1, 100))
						datum.ActionDelay = pointer.FromInt(-1)
						datum.ActionDuration = pointer.FromInt(-1)
						datum.ActionPeakOffset = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"fiasp", "other", "rapidAdult", "rapidChild", "walsh"}), "/modelType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/modelTypeOther"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400), "/actionDelay"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400), "/actionDuration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 86400), "/actionPeakOffset"),
				),
			)
		})
	})
})
