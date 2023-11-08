package pump_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataTypesCommon "github.com/tidepool-org/platform/data/types/common"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesSettingsPumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("SleepSchedule", func() {

	Context("NewSleepSchedules", func() {
		It("returns successfully with default values", func() {
			datum := dataTypesSettingsPump.NewSleepScheduleMap()
			Expect(datum).ToNot(BeNil())
			Expect(*datum).To(BeEmpty())
		})
	})

	Context("SleepSchedules", func() {

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsPump.SleepScheduleMap), expectedErrors ...error) {
					datum := dataTypesSettingsPumpTest.RandomSleepSchedules(3)
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsPump.SleepScheduleMap) {},
				),
				Entry("empty",
					func(datum *dataTypesSettingsPump.SleepScheduleMap) {
						*datum = *dataTypesSettingsPump.NewSleepScheduleMap()
					},
				),
				Entry("has one",
					func(datum *dataTypesSettingsPump.SleepScheduleMap) {
						*datum = *dataTypesSettingsPumpTest.RandomSleepSchedules(1)
					},
				),
				Entry("has many",
					func(datum *dataTypesSettingsPump.SleepScheduleMap) {
						*datum = *dataTypesSettingsPumpTest.RandomSleepSchedules(19)
					},
				),
				Entry("entry missing",
					func(datum *dataTypesSettingsPump.SleepScheduleMap) {
						(*datum)[dataTypesSettingsPumpTest.SleepScheduleName(0)] = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), fmt.Sprintf("/%s", dataTypesSettingsPumpTest.SleepScheduleName(0))),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsPump.SleepScheduleMap) {
						*datum = *dataTypesSettingsPumpTest.RandomSleepSchedules(3)
						(*datum)[dataTypesSettingsPumpTest.SleepScheduleName(1)] = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), fmt.Sprintf("/%s", dataTypesSettingsPumpTest.SleepScheduleName(1))),
				),
			)
		})
	})

	Context("NewSleepSchedule", func() {
		It("returns successfully with default values", func() {
			datum := dataTypesSettingsPump.NewSleepSchedule()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Enabled).To(BeNil())
			Expect(datum.Days).To(BeNil())
			Expect(datum.Start).To(BeNil())
			Expect(datum.End).To(BeNil())
		})
	})

	Context("SleepSchedule", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesSettingsPump.SleepSchedule)) {
				datum := dataTypesSettingsPumpTest.RandomSleepSchedule()
				mutator(datum)
				test.ExpectSerializedObjectBSON(datum, dataTypesSettingsPumpTest.NewObjectFromSleepSchedule(datum, test.ObjectFormatBSON))
				test.ExpectSerializedObjectJSON(datum, dataTypesSettingsPumpTest.NewObjectFromSleepSchedule(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataTypesSettingsPump.SleepSchedule) {},
			),
			Entry("empty",
				func(datum *dataTypesSettingsPump.SleepSchedule) { *datum = dataTypesSettingsPump.SleepSchedule{} },
			),
		)

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesSettingsPump.SleepSchedule), expectedErrors ...error) {
					datum := dataTypesSettingsPumpTest.RandomSleepSchedule()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesSettingsPump.SleepSchedule) {},
				),
				Entry("enabled empty",
					func(datum *dataTypesSettingsPump.SleepSchedule) { datum.Enabled = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/enabled"),
				),

				Entry("days missing",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.Days = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/days"),
				),
				Entry("days contains invalid",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.Days = pointer.FromStringArray(append([]string{"invalid"}, test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(0, len(dataTypesCommon.DaysOfWeek())-1, dataTypesCommon.DaysOfWeek())...))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}), "/days/0"),
				),
				Entry("days contains duplicate",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						duplicate := test.RandomStringFromArray(dataTypesCommon.DaysOfWeek())
						datum.Days = pointer.FromStringArray([]string{duplicate, duplicate})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/days/1"),
				),
				Entry("days valid",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.Days = pointer.FromStringArray(test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(dataTypesCommon.DaysOfWeek()), dataTypesCommon.DaysOfWeek()))
					},
				),
				Entry("start missing",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.Start = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
				Entry("start out of range (lower)",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.Start = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1,
						dataTypesSettingsPump.SleepSchedulesMidnightOffsetMinimum,
						dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum), "/start"),
				),
				Entry("start in range (lower)",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.Start = pointer.FromInt(dataTypesSettingsPump.SleepSchedulesMidnightOffsetMinimum)
					},
				),
				Entry("start in range (upper)",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.Start = pointer.FromInt(dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum)
					},
				),
				Entry("start out of range (upper)",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.Start = pointer.FromInt(dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(
						dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum+1,
						dataTypesSettingsPump.SleepSchedulesMidnightOffsetMinimum,
						dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum), "/start"),
				),
				Entry("end missing",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.End = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/end"),
				),
				Entry("end out of range (lower)",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.End = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-1,
						dataTypesSettingsPump.SleepSchedulesMidnightOffsetMinimum,
						dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum), "/end"),
				),
				Entry("end in range (lower)",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.End = pointer.FromInt(dataTypesSettingsPump.SleepSchedulesMidnightOffsetMinimum)
					},
				),
				Entry("end in range (upper)",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.End = pointer.FromInt(dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum)
					},
				),
				Entry("end out of range (upper)",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.End = pointer.FromInt(dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum + 1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(
						dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum+1,
						dataTypesSettingsPump.SleepSchedulesMidnightOffsetMinimum,
						dataTypesSettingsPump.SleepSchedulesMidnightOffsetMaximum), "/end"),
				),
				Entry("multiple errors",
					func(datum *dataTypesSettingsPump.SleepSchedule) {
						datum.Days = nil
						datum.Start = nil
						datum.End = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/days"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/end"),
				),
			)
		})
	})
})
