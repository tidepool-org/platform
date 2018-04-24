package timechange_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/device/timechange"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewChange() *timechange.Change {
	datum := timechange.NewChange()
	datum.Agent = pointer.String(test.RandomStringFromStringArray(timechange.Agents()))
	datum.From = pointer.String(test.NewTime().Format("2006-01-02T15:04:05"))
	datum.To = pointer.String(test.NewTime().Format("2006-01-02T15:04:05"))
	return datum
}

func CloneChange(datum *timechange.Change) *timechange.Change {
	if datum == nil {
		return nil
	}
	clone := timechange.NewChange()
	clone.Agent = test.CloneString(datum.Agent)
	clone.From = test.CloneString(datum.From)
	clone.To = test.CloneString(datum.To)
	return clone
}

func NewTestChange(agent interface{}, from interface{}, to interface{}) *timechange.Change {
	datum := timechange.NewChange()
	if val, ok := agent.(string); ok {
		datum.Agent = &val
	}
	if val, ok := from.(string); ok {
		datum.From = &val
	}
	if val, ok := to.(string); ok {
		datum.To = &val
	}
	return datum
}

var _ = Describe("Change", func() {
	It("AgentAutomatic is expected", func() {
		Expect(timechange.AgentAutomatic).To(Equal("automatic"))
	})

	It("AgentManual is expected", func() {
		Expect(timechange.AgentManual).To(Equal("manual"))
	})

	It("FromTimeFormat is expected", func() {
		Expect(timechange.FromTimeFormat).To(Equal("2006-01-02T15:04:05"))
	})

	It("ToTimeFormat is expected", func() {
		Expect(timechange.ToTimeFormat).To(Equal("2006-01-02T15:04:05"))
	})

	It("Agents returns expected", func() {
		Expect(timechange.Agents()).To(Equal([]string{"automatic", "manual"}))
	})

	Context("ParseChange", func() {
		// TODO
	})

	Context("NewChange", func() {
		It("is successful", func() {
			Expect(timechange.NewChange()).To(Equal(&timechange.Change{}))
		})
	})

	Context("Change", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *timechange.Change), expectedErrors ...error) {
					datum := NewChange()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *timechange.Change) {},
				),
				Entry("agent missing",
					func(datum *timechange.Change) { datum.Agent = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/agent"),
				),
				Entry("agent invalid",
					func(datum *timechange.Change) { datum.Agent = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"automatic", "manual"}), "/agent"),
				),
				Entry("agent automatic",
					func(datum *timechange.Change) { datum.Agent = pointer.String("automatic") },
				),
				Entry("agent manual",
					func(datum *timechange.Change) { datum.Agent = pointer.String("manual") },
				),
				Entry("from missing",
					func(datum *timechange.Change) { datum.From = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/from"),
				),
				Entry("from invalid",
					func(datum *timechange.Change) { datum.From = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", "2006-01-02T15:04:05"), "/from"),
				),
				Entry("from valid",
					func(datum *timechange.Change) {
						datum.From = pointer.String(test.NewTime().Format("2006-01-02T15:04:05"))
					},
				),
				Entry("to missing",
					func(datum *timechange.Change) { datum.To = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/to"),
				),
				Entry("to invalid",
					func(datum *timechange.Change) { datum.To = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", "2006-01-02T15:04:05"), "/to"),
				),
				Entry("to valid",
					func(datum *timechange.Change) {
						datum.To = pointer.String(test.NewTime().Format("2006-01-02T15:04:05"))
					},
				),
				Entry("multiple errors",
					func(datum *timechange.Change) {
						datum.Agent = nil
						datum.From = nil
						datum.To = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/agent"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/from"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/to"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *timechange.Change)) {
					for _, origin := range structure.Origins() {
						datum := NewChange()
						mutator(datum)
						expectedDatum := CloneChange(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *timechange.Change) {},
				),
				Entry("does not modify the datum; agent missing",
					func(datum *timechange.Change) { datum.Agent = nil },
				),
				Entry("does not modify the datum; agent automatic",
					func(datum *timechange.Change) { datum.Agent = pointer.String("automatic") },
				),
				Entry("does not modify the datum; agent manual",
					func(datum *timechange.Change) { datum.Agent = pointer.String("manual") },
				),
				Entry("does not modify the datum; from missing",
					func(datum *timechange.Change) { datum.From = nil },
				),
				Entry("does not modify the datum; to missing",
					func(datum *timechange.Change) { datum.To = nil },
				),
			)
		})
	})
})
