package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/dexcom"
	dexcomTest "github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var futureTime = time.Unix(4102444800, 0)

var _ = Describe("Device", func() {
	Context("AlertSetting", func() {
		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.AlertSetting), expectedErrors ...error) {
					datum := dexcomTest.RandomAlertSetting()
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.AlertSetting) {},
				),
				Entry("system time is missing",
					func(datum *dexcom.AlertSetting) { datum.SystemTime = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/systemTime"),
				),
				Entry("system time is zero",
					func(datum *dexcom.AlertSetting) { datum.SystemTime = pointer.FromTime(time.Time{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/systemTime"),
				),
				Entry("system time is after now",
					func(datum *dexcom.AlertSetting) { datum.SystemTime = pointer.FromTime(futureTime) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/systemTime"),
				),
				Entry("system time is valid",
					func(datum *dexcom.AlertSetting) {
						datum.SystemTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()))
					},
				),
				Entry("display time is missing",
					func(datum *dexcom.AlertSetting) { datum.DisplayTime = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayTime"),
				),
				Entry("display time is zero",
					func(datum *dexcom.AlertSetting) { datum.DisplayTime = pointer.FromTime(time.Time{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/displayTime"),
				),
				Entry("display time is valid",
					func(datum *dexcom.AlertSetting) { datum.DisplayTime = pointer.FromTime(test.RandomTime()) },
				),
				Entry("alert name is missing",
					func(datum *dexcom.AlertSetting) { datum.AlertName = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertName"),
				),
				Entry("alert name is empty",
					func(datum *dexcom.AlertSetting) { datum.AlertName = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", []string{"fixedLow", "low", "high", "rise", "fall", "outOfRange"}), "/alertName"),
				),
				Entry("alert name is invalid",
					func(datum *dexcom.AlertSetting) { datum.AlertName = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"fixedLow", "low", "high", "rise", "fall", "outOfRange"}), "/alertName"),
				),
				Entry("alert name is fixedLow",
					func(datum *dexcom.AlertSetting) { *datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow") },
				),
				Entry("alert name is fixedLow; unit is empty",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow")
						datum.Unit = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dexcom.AlertSettingFixedLowUnits()), "/unit"),
				),
				Entry("alert name is fixedLow; unit is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow")
						datum.Unit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.AlertSettingFixedLowUnits()), "/unit"),
				),
				Entry("alert name is fixedLow; unit is mg/dL",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow")
						datum.Unit = pointer.FromString("mg/dL")
					},
				),
				Entry("alert name is fixedLow; unit is mg/dL; value is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow")
						datum.Unit = pointer.FromString("mg/dL")
						datum.Value = pointer.FromFloat64(1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(1, dexcom.AlertSettingFixedLowValuesForUnits(pointer.FromString("mg/dL"))), "/value"),
				),
				Entry("alert name is fixedLow; unit is mg/dL; value is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow")
						datum.Unit = pointer.FromString("mg/dL")
						datum.Value = pointer.FromFloat64(test.RandomFloat64FromArray(dexcom.AlertSettingFixedLowValuesForUnits(pointer.FromString("mg/dL"))))
					},
				),
				Entry("alert name is fixedLow; delay is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow")
						datum.Delay = pointer.FromInt(60)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(60, []int{0}), "/delay"),
				),
				Entry("alert name is fixedLow; delay is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow")
						datum.Delay = pointer.FromInt(0)
					},
				),
				Entry("alert name is fixedLow; snooze is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow")
						datum.Snooze = pointer.FromInt(5)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(5, dexcom.AlertSettingFixedLowSnoozes()), "/snooze"),
				),
				Entry("alert name is fixedLow; snooze is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow")
						datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dexcom.AlertSettingFixedLowSnoozes()))
					},
				),
				Entry("alert name is fixedLow; snooze is valid, but 28 (HACK: Dexcom)",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow")
						datum.Snooze = pointer.FromInt(28)
					},
				),
				Entry("alert name is fixedLow; enabled is false",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow")
						datum.Enabled = pointer.FromBool(false)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueBoolNotTrue(), "/enabled"),
				),
				Entry("alert name is fixedLow; enabled is true",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow")
						datum.Enabled = pointer.FromBool(true)
					},
				),
				Entry("alert name is fixedLow; multiple errors",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fixedLow")
						datum.SystemTime = pointer.FromTime(time.Time{})
						datum.DisplayTime = pointer.FromTime(time.Time{})
						datum.Unit = pointer.FromString("mg/dL")
						datum.Value = pointer.FromFloat64(1)
						datum.Delay = pointer.FromInt(60)
						datum.Snooze = pointer.FromInt(5)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/systemTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/displayTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(1, dexcom.AlertSettingFixedLowValuesForUnits(pointer.FromString("mg/dL"))), "/value"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(60, []int{0}), "/delay"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(5, dexcom.AlertSettingFixedLowSnoozes()), "/snooze"),
				),
				Entry("alert name is low",
					func(datum *dexcom.AlertSetting) { *datum = *dexcomTest.RandomAlertSettingWithAlertName("low") },
				),
				Entry("alert name is low; unit is empty",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("low")
						datum.Unit = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dexcom.AlertSettingLowUnits()), "/unit"),
				),
				Entry("alert name is low; unit is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("low")
						datum.Unit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.AlertSettingLowUnits()), "/unit"),
				),
				Entry("alert name is low; unit is mg/dL",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("low")
						datum.Unit = pointer.FromString("mg/dL")
					},
				),
				Entry("alert name is low; unit is mg/dL; value is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("low")
						datum.Unit = pointer.FromString("mg/dL")
						datum.Value = pointer.FromFloat64(1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(1, dexcom.AlertSettingLowValuesForUnits(pointer.FromString("mg/dL"))), "/value"),
				),
				Entry("alert name is low; unit is mg/dL; value is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("low")
						datum.Unit = pointer.FromString("mg/dL")
						datum.Value = pointer.FromFloat64(test.RandomFloat64FromArray(dexcom.AlertSettingLowValuesForUnits(pointer.FromString("mg/dL"))))
					},
				),
				Entry("alert name is low; delay is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("low")
						datum.Delay = pointer.FromInt(60)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(60, []int{0}), "/delay"),
				),
				Entry("alert name is low; delay is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("low")
						datum.Delay = pointer.FromInt(0)
					},
				),
				Entry("alert name is low; snooze is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("low")
						datum.Snooze = pointer.FromInt(5)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(5, dexcom.AlertSettingLowSnoozes()), "/snooze"),
				),
				Entry("alert name is low; snooze is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("low")
						datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dexcom.AlertSettingLowSnoozes()))
					},
				),
				Entry("alert name is low; enabled is false",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("low")
						datum.Enabled = pointer.FromBool(false)
					},
				),
				Entry("alert name is low; enabled is true",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("low")
						datum.Enabled = pointer.FromBool(true)
					},
				),
				Entry("alert name is low; multiple errors",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("low")
						datum.SystemTime = pointer.FromTime(time.Time{})
						datum.DisplayTime = pointer.FromTime(time.Time{})
						datum.Unit = pointer.FromString("mg/dL")
						datum.Value = pointer.FromFloat64(1)
						datum.Delay = pointer.FromInt(60)
						datum.Snooze = pointer.FromInt(5)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/systemTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/displayTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(1, dexcom.AlertSettingLowValuesForUnits(pointer.FromString("mg/dL"))), "/value"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(60, []int{0}), "/delay"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(5, dexcom.AlertSettingLowSnoozes()), "/snooze"),
				),
				Entry("alert name is high",
					func(datum *dexcom.AlertSetting) { *datum = *dexcomTest.RandomAlertSettingWithAlertName("high") },
				),
				Entry("alert name is high; unit is empty",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("high")
						datum.Unit = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dexcom.AlertSettingHighUnits()), "/unit"),
				),
				Entry("alert name is high; unit is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("high")
						datum.Unit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.AlertSettingHighUnits()), "/unit"),
				),
				Entry("alert name is high; unit is mg/dL",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("high")
						datum.Unit = pointer.FromString("mg/dL")
					},
				),
				Entry("alert name is high; unit is mg/dL; value is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("high")
						datum.Unit = pointer.FromString("mg/dL")
						datum.Value = pointer.FromFloat64(1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(1, dexcom.AlertSettingHighValuesForUnits(pointer.FromString("mg/dL"))), "/value"),
				),
				Entry("alert name is high; unit is mg/dL; value is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("high")
						datum.Unit = pointer.FromString("mg/dL")
						datum.Value = pointer.FromFloat64(test.RandomFloat64FromArray(dexcom.AlertSettingHighValuesForUnits(pointer.FromString("mg/dL"))))
					},
				),
				Entry("alert name is high; delay is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("high")
						datum.Delay = pointer.FromInt(60)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(60, []int{0}), "/delay"),
				),
				Entry("alert name is high; delay is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("high")
						datum.Delay = pointer.FromInt(0)
					},
				),
				Entry("alert name is high; snooze is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("high")
						datum.Snooze = pointer.FromInt(5)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(5, dexcom.AlertSettingHighSnoozes()), "/snooze"),
				),
				Entry("alert name is high; snooze is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("high")
						datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dexcom.AlertSettingHighSnoozes()))
					},
				),
				Entry("alert name is high; enabled is false",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("high")
						datum.Enabled = pointer.FromBool(false)
					},
				),
				Entry("alert name is high; enabled is true",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("high")
						datum.Enabled = pointer.FromBool(true)
					},
				),
				Entry("alert name is high; multiple errors",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("high")
						datum.SystemTime = pointer.FromTime(time.Time{})
						datum.DisplayTime = pointer.FromTime(time.Time{})
						datum.Unit = pointer.FromString("mg/dL")
						datum.Value = pointer.FromFloat64(1)
						datum.Delay = pointer.FromInt(60)
						datum.Snooze = pointer.FromInt(5)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/systemTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/displayTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(1, dexcom.AlertSettingHighValuesForUnits(pointer.FromString("mg/dL"))), "/value"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(60, []int{0}), "/delay"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(5, dexcom.AlertSettingHighSnoozes()), "/snooze"),
				),
				Entry("alert name is rise",
					func(datum *dexcom.AlertSetting) { *datum = *dexcomTest.RandomAlertSettingWithAlertName("rise") },
				),
				Entry("alert name is rise; unit is empty",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("rise")
						datum.Unit = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dexcom.AlertSettingRiseUnits()), "/unit"),
				),
				Entry("alert name is rise; unit is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("rise")
						datum.Unit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.AlertSettingRiseUnits()), "/unit"),
				),
				Entry("alert name is rise; unit is mg/dL/min",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("rise")
						datum.Unit = pointer.FromString("mg/dL/min")
					},
				),
				Entry("alert name is rise; unit is mg/dL/min; value is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("rise")
						datum.Unit = pointer.FromString("mg/dL/min")
						datum.Value = pointer.FromFloat64(1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(1, dexcom.AlertSettingRiseValuesForUnits(pointer.FromString("mg/dL/min"))), "/value"),
				),
				Entry("alert name is rise; unit is mg/dL/min; value is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("rise")
						datum.Unit = pointer.FromString("mg/dL/min")
						datum.Value = pointer.FromFloat64(test.RandomFloat64FromArray(dexcom.AlertSettingRiseValuesForUnits(pointer.FromString("mg/dL/min"))))
					},
				),
				Entry("alert name is rise; delay is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("rise")
						datum.Delay = pointer.FromInt(60)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(60, []int{0}), "/delay"),
				),
				Entry("alert name is rise; delay is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("rise")
						datum.Delay = pointer.FromInt(0)
					},
				),
				Entry("alert name is rise; snooze is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("rise")
						datum.Snooze = pointer.FromInt(5)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(5, dexcom.AlertSettingRiseSnoozes()), "/snooze"),
				),
				Entry("alert name is rise; snooze is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("rise")
						datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dexcom.AlertSettingRiseSnoozes()))
					},
				),
				Entry("alert name is rise; enabled is false",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("rise")
						datum.Enabled = pointer.FromBool(false)
					},
				),
				Entry("alert name is rise; enabled is true",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("rise")
						datum.Enabled = pointer.FromBool(true)
					},
				),
				Entry("alert name is rise; multiple errors",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("rise")
						datum.SystemTime = pointer.FromTime(time.Time{})
						datum.DisplayTime = pointer.FromTime(time.Time{})
						datum.Unit = pointer.FromString("mg/dL/min")
						datum.Value = pointer.FromFloat64(1)
						datum.Delay = pointer.FromInt(60)
						datum.Snooze = pointer.FromInt(5)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/systemTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/displayTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(1, dexcom.AlertSettingRiseValuesForUnits(pointer.FromString("mg/dL/min"))), "/value"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(60, []int{0}), "/delay"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(5, dexcom.AlertSettingRiseSnoozes()), "/snooze"),
				),
				Entry("alert name is fall",
					func(datum *dexcom.AlertSetting) { *datum = *dexcomTest.RandomAlertSettingWithAlertName("fall") },
				),
				Entry("alert name is fall; unit is empty",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fall")
						datum.Unit = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dexcom.AlertSettingFallUnits()), "/unit"),
				),
				Entry("alert name is fall; unit is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fall")
						datum.Unit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.AlertSettingFallUnits()), "/unit"),
				),
				Entry("alert name is fall; unit is mg/dL/min",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fall")
						datum.Unit = pointer.FromString("mg/dL/min")
					},
				),
				Entry("alert name is fall; unit is mg/dL/min; value is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fall")
						datum.Unit = pointer.FromString("mg/dL/min")
						datum.Value = pointer.FromFloat64(1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(1, dexcom.AlertSettingFallValuesForUnits(pointer.FromString("mg/dL/min"))), "/value"),
				),
				Entry("alert name is fall; unit is mg/dL/min; value is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fall")
						datum.Unit = pointer.FromString("mg/dL/min")
						datum.Value = pointer.FromFloat64(test.RandomFloat64FromArray(dexcom.AlertSettingFallValuesForUnits(pointer.FromString("mg/dL/min"))))
					},
				),
				Entry("alert name is fall; unit is mg/dL/min; value is valid, but negative (HACK: Dexcom)",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fall")
						datum.Unit = pointer.FromString("mg/dL/min")
						datum.Value = pointer.FromFloat64(-test.RandomFloat64FromArray(dexcom.AlertSettingFallValuesForUnits(pointer.FromString("mg/dL/min"))))
					},
				),
				Entry("alert name is fall; delay is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fall")
						datum.Delay = pointer.FromInt(60)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(60, []int{0}), "/delay"),
				),
				Entry("alert name is fall; delay is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fall")
						datum.Delay = pointer.FromInt(0)
					},
				),
				Entry("alert name is fall; snooze is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fall")
						datum.Snooze = pointer.FromInt(5)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(5, dexcom.AlertSettingFallSnoozes()), "/snooze"),
				),
				Entry("alert name is fall; snooze is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fall")
						datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dexcom.AlertSettingFallSnoozes()))
					},
				),
				Entry("alert name is fall; enabled is false",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fall")
						datum.Enabled = pointer.FromBool(false)
					},
				),
				Entry("alert name is fall; enabled is true",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fall")
						datum.Enabled = pointer.FromBool(true)
					},
				),
				Entry("alert name is fall; multiple errors",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("fall")
						datum.SystemTime = pointer.FromTime(time.Time{})
						datum.DisplayTime = pointer.FromTime(time.Time{})
						datum.Unit = pointer.FromString("mg/dL/min")
						datum.Value = pointer.FromFloat64(1)
						datum.Delay = pointer.FromInt(60)
						datum.Snooze = pointer.FromInt(5)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/systemTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/displayTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(1, dexcom.AlertSettingFallValuesForUnits(pointer.FromString("mg/dL/min"))), "/value"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(60, []int{0}), "/delay"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(5, dexcom.AlertSettingFallSnoozes()), "/snooze"),
				),
				Entry("alert name is outOfRange",
					func(datum *dexcom.AlertSetting) { *datum = *dexcomTest.RandomAlertSettingWithAlertName("outOfRange") },
				),
				Entry("alert name is outOfRange; unit is empty",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("outOfRange")
						datum.Unit = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dexcom.AlertSettingOutOfRangeUnits()), "/unit"),
				),
				Entry("alert name is outOfRange; unit is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("outOfRange")
						datum.Unit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.AlertSettingOutOfRangeUnits()), "/unit"),
				),
				Entry("alert name is outOfRange; unit is minutes",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("outOfRange")
						datum.Unit = pointer.FromString("minutes")
					},
				),
				Entry("alert name is outOfRange; unit is minutes; value is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("outOfRange")
						datum.Unit = pointer.FromString("minutes")
						datum.Value = pointer.FromFloat64(1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(1, dexcom.AlertSettingOutOfRangeValuesForUnits(pointer.FromString("minutes"))), "/value"),
				),
				Entry("alert name is outOfRange; unit is minutes; value is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("outOfRange")
						datum.Unit = pointer.FromString("minutes")
						datum.Value = pointer.FromFloat64(test.RandomFloat64FromArray(dexcom.AlertSettingOutOfRangeValuesForUnits(pointer.FromString("minutes"))))
					},
				),
				Entry("alert name is outOfRange; delay is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("outOfRange")
						datum.Delay = pointer.FromInt(60)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(60, []int{0}), "/delay"),
				),
				Entry("alert name is outOfRange; delay is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("outOfRange")
						datum.Delay = pointer.FromInt(0)
					},
				),
				Entry("alert name is outOfRange; snooze is invalid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("outOfRange")
						datum.Snooze = pointer.FromInt(5)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(5, dexcom.AlertSettingOutOfRangeSnoozes()), "/snooze"),
				),
				Entry("alert name is outOfRange; snooze is valid",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("outOfRange")
						datum.Snooze = pointer.FromInt(test.RandomIntFromArray(dexcom.AlertSettingOutOfRangeSnoozes()))
					},
				),
				Entry("alert name is outOfRange; enabled is false",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("outOfRange")
						datum.Enabled = pointer.FromBool(false)
					},
				),
				Entry("alert name is outOfRange; enabled is true",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("outOfRange")
						datum.Enabled = pointer.FromBool(true)
					},
				),
				Entry("alert name is outOfRange; multiple errors",
					func(datum *dexcom.AlertSetting) {
						*datum = *dexcomTest.RandomAlertSettingWithAlertName("outOfRange")
						datum.SystemTime = pointer.FromTime(time.Time{})
						datum.DisplayTime = pointer.FromTime(time.Time{})
						datum.Unit = pointer.FromString("minutes")
						datum.Value = pointer.FromFloat64(1)
						datum.Delay = pointer.FromInt(60)
						datum.Snooze = pointer.FromInt(5)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/systemTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/displayTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueFloat64NotOneOf(1, dexcom.AlertSettingOutOfRangeValuesForUnits(pointer.FromString("minutes"))), "/value"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(60, []int{0}), "/delay"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueIntNotOneOf(5, dexcom.AlertSettingOutOfRangeSnoozes()), "/snooze"),
				),
			)
		})

		Context("with new alert setting", func() {
			var alertSetting *dexcom.AlertSetting

			BeforeEach(func() {
				alertSetting = dexcomTest.RandomAlertSetting()
			})

			Context("IsNewerMatchThan", func() {
				var testAlertSetting *dexcom.AlertSetting

				BeforeEach(func() {
					testAlertSetting = dexcomTest.CloneAlertSetting(alertSetting)
					testAlertSetting.SystemTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), *alertSetting.SystemTime))
				})

				It("returns false if the name does not match", func() {
					alertSetting.AlertName = pointer.FromString("fixedLow")
					testAlertSetting.AlertName = pointer.FromString("low")
					Expect(alertSetting.IsNewerMatchThan(testAlertSetting)).To(BeFalse())
				})

				It("returns false if the system time is not newer", func() {
					testAlertSetting.SystemTime = pointer.FromTime(test.RandomTimeFromRange(*alertSetting.SystemTime, test.RandomTimeMaximum()))
					Expect(alertSetting.IsNewerMatchThan(testAlertSetting)).To(BeFalse())
				})

				It("returns true if the system time is newer", func() {
					Expect(alertSetting.IsNewerMatchThan(testAlertSetting)).To(BeTrue())
				})
			})
		})
	})

	Context("AlertSettings", func() {
		Context("with new alert setting", func() {
			var alertSettings *dexcom.AlertSettings

			BeforeEach(func() {
				alertSettings = dexcomTest.RandomAlertSettings()
			})

			Context("Deduplicate", func() {
				It("returns alert settings that have no duplicates as-is", func() {
					duplicateAlertSettings := dexcomTest.CloneAlertSettings(alertSettings)
					duplicateAlertSettings.Deduplicate()
					Expect(duplicateAlertSettings).To(Equal(alertSettings))
				})

				It("returns alert settings that have duplicates without duplicates in the same order", func() {
					duplicateAlertSettings := append(append(*alertSettings, *dexcomTest.CloneAlertSettings(alertSettings)...), *dexcomTest.CloneAlertSettings(alertSettings)...)
					duplicateAlertSettings.Deduplicate()
					Expect(&duplicateAlertSettings).To(Equal(alertSettings))
				})
			})
		})
	})

	It("AlertSettingFixedLowUnits is expected", func() {
		Expect(dexcom.AlertSettingFixedLowUnits()).To(Equal([]string{"mg/dL"}))
	})

	Context("AlertSettingFixedLowValuesForUnits is expected", func() {
		It("returns expected values for mg/dL units", func() {
			Expect(dexcom.AlertSettingFixedLowValuesForUnits(pointer.FromString("mg/dL"))).To(Equal([]float64{55}))
		})

		It("returns expected values for mmol/L units", func() {
			Expect(dexcom.AlertSettingFixedLowValuesForUnits(pointer.FromString("mmol/L"))).To(BeNil())
		})

		It("returns expected values for unknown units", func() {
			Expect(dexcom.AlertSettingFixedLowValuesForUnits(pointer.FromString("unknown"))).To(BeNil())
		})
	})

	It("AlertSettingFixedLowSnoozes is expected", func() {
		Expect(dexcom.AlertSettingFixedLowSnoozes()).To(Equal([]int{0, 30}))
	})

	It("AlertSettingLowUnits is expected", func() {
		Expect(dexcom.AlertSettingLowUnits()).To(Equal([]string{"mg/dL"}))
	})

	Context("AlertSettingLowValuesForUnits is expected", func() {
		It("returns expected values for mg/dL units", func() {
			Expect(dexcom.AlertSettingLowValuesForUnits(pointer.FromString("mg/dL"))).To(Equal([]float64{60, 65, 70, 75, 80, 85, 90, 95, 100}))
		})

		It("returns expected values for mmol/L units", func() {
			Expect(dexcom.AlertSettingLowValuesForUnits(pointer.FromString("mmol/L"))).To(BeNil())
		})

		It("returns expected values for unknown units", func() {
			Expect(dexcom.AlertSettingLowValuesForUnits(pointer.FromString("unknown"))).To(BeNil())
		})
	})

	It("AlertSettingLowSnoozes is expected", func() {
		Expect(dexcom.AlertSettingLowSnoozes()).To(Equal([]int{0, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85,
			90, 95, 100, 105, 110, 115, 120, 125, 130, 135, 140, 145, 150, 155, 160, 165, 170, 175, 180, 185, 190, 195, 200,
			205, 210, 215, 220, 225, 230, 235, 240, 255, 270, 285, 300}))
	})

	It("AlertSettingHighUnits is expected", func() {
		Expect(dexcom.AlertSettingHighUnits()).To(Equal([]string{"mg/dL"}))
	})

	Context("AlertSettingHighValuesForUnits is expected", func() {
		It("returns expected values for mg/dL units", func() {
			Expect(dexcom.AlertSettingHighValuesForUnits(pointer.FromString("mg/dL"))).To(Equal([]float64{120, 130, 140, 150, 160, 170, 180, 190,
				200, 210, 220, 230, 240, 250, 260, 270, 280, 290, 300, 310, 320, 330, 340, 350, 360, 370, 380, 390, 400}))
		})

		It("returns expected values for mmol/L units", func() {
			Expect(dexcom.AlertSettingHighValuesForUnits(pointer.FromString("mmol/L"))).To(BeNil())
		})

		It("returns expected values for unknown units", func() {
			Expect(dexcom.AlertSettingHighValuesForUnits(pointer.FromString("unknown"))).To(BeNil())
		})
	})

	It("AlertSettingHighSnoozes is expected", func() {
		Expect(dexcom.AlertSettingHighSnoozes()).To(Equal([]int{0, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85,
			90, 95, 100, 105, 110, 115, 120, 125, 130, 135, 140, 145, 150, 155, 160, 165, 170, 175, 180, 185, 190, 195, 200,
			205, 210, 215, 220, 225, 230, 235, 240, 255, 270, 285, 300}))
	})

	It("AlertSettingRiseUnits is expected", func() {
		Expect(dexcom.AlertSettingRiseUnits()).To(Equal([]string{"mg/dL/min"}))
	})

	Context("AlertSettingRiseValuesForUnits is expected", func() {
		It("returns expected values for mg/dL/min units", func() {
			Expect(dexcom.AlertSettingRiseValuesForUnits(pointer.FromString("mg/dL/min"))).To(Equal([]float64{2, 3}))
		})

		It("returns expected values for mmol/L/min units", func() {
			Expect(dexcom.AlertSettingRiseValuesForUnits(pointer.FromString("mmol/L/min"))).To(BeNil())
		})

		It("returns expected values for unknown units", func() {
			Expect(dexcom.AlertSettingRiseValuesForUnits(pointer.FromString("unknown"))).To(BeNil())
		})
	})

	It("AlertSettingRiseSnoozes is expected", func() {
		Expect(dexcom.AlertSettingRiseSnoozes()).To(Equal([]int{0, 30}))
	})

	It("AlertSettingFallUnits is expected", func() {
		Expect(dexcom.AlertSettingFallUnits()).To(Equal([]string{"mg/dL/min"}))
	})

	Context("AlertSettingFallValuesForUnits is expected", func() {
		It("returns expected values for mg/dL/min units", func() {
			Expect(dexcom.AlertSettingFallValuesForUnits(pointer.FromString("mg/dL/min"))).To(Equal([]float64{2, 3}))
		})

		It("returns expected values for mmol/L/min units", func() {
			Expect(dexcom.AlertSettingFallValuesForUnits(pointer.FromString("mmol/L/min"))).To(BeNil())
		})

		It("returns expected values for unknown units", func() {
			Expect(dexcom.AlertSettingFallValuesForUnits(pointer.FromString("unknown"))).To(BeNil())
		})
	})

	It("AlertSettingFallSnoozes is expected", func() {
		Expect(dexcom.AlertSettingFallSnoozes()).To(Equal([]int{0, 30}))
	})

	It("AlertSettingOutOfRangeUnits is expected", func() {
		Expect(dexcom.AlertSettingOutOfRangeUnits()).To(Equal([]string{"minutes"}))
	})

	Context("AlertSettingOutOfRangeValuesForUnits is expected", func() {
		It("returns expected values for minutes units", func() {
			Expect(dexcom.AlertSettingOutOfRangeValuesForUnits(pointer.FromString("minutes"))).To(Equal([]float64{20, 25, 30, 35, 40, 45, 50, 55,
				60, 65, 70, 75, 80, 85, 90, 95, 100, 105, 110, 115, 120, 125, 130, 135, 140, 145, 150, 155, 160, 165, 170,
				175, 180, 185, 190, 195, 200, 205, 210, 215, 220, 225, 230, 235, 240}))
		})

		It("returns expected values for unknown units", func() {
			Expect(dexcom.AlertSettingOutOfRangeValuesForUnits(pointer.FromString("unknown"))).To(BeNil())
		})
	})

	It("AlertSettingOutOfRangeSnoozes is expected", func() {
		Expect(dexcom.AlertSettingOutOfRangeSnoozes()).To(Equal([]int{0, 20, 25, 30}))
	})
})
