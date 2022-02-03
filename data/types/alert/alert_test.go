package alert_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesAlert "github.com/tidepool-org/platform/data/types/alert"
	dataTypesAlertTest "github.com/tidepool-org/platform/data/types/alert/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/metadata"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &dataTypes.Meta{
		Type: "alert",
	}
}

var _ = Describe("Alert", func() {
	It("Type is expected", func() {
		Expect(dataTypesAlert.Type).To(Equal("alert"))
	})

	It("NameLengthMaximum is expected", func() {
		Expect(dataTypesAlert.NameLengthMaximum).To(Equal(100))
	})

	It("PriorityCritical is expected", func() {
		Expect(dataTypesAlert.PriorityCritical).To(Equal("critical"))
	})

	It("PriorityNormal is expected", func() {
		Expect(dataTypesAlert.PriorityNormal).To(Equal("normal"))
	})

	It("PriorityTimeSensitive is expected", func() {
		Expect(dataTypesAlert.PriorityTimeSensitive).To(Equal("timeSensitive"))
	})

	It("SoundName is expected", func() {
		Expect(dataTypesAlert.SoundName).To(Equal("name"))
	})

	It("SoundNameLengthMaximum is expected", func() {
		Expect(dataTypesAlert.SoundNameLengthMaximum).To(Equal(100))
	})

	It("SoundSilence is expected", func() {
		Expect(dataTypesAlert.SoundSilence).To(Equal("silence"))
	})

	It("SoundVibrate is expected", func() {
		Expect(dataTypesAlert.SoundVibrate).To(Equal("vibrate"))
	})

	It("TriggerDelayed is expected", func() {
		Expect(dataTypesAlert.TriggerDelayed).To(Equal("delayed"))
	})

	It("TriggerDelayMaximum is expected", func() {
		Expect(dataTypesAlert.TriggerDelayMaximum).To(Equal(86400))
	})

	It("TriggerDelayMinimum is expected", func() {
		Expect(dataTypesAlert.TriggerDelayMinimum).To(Equal(0))
	})

	It("TriggerImmediate is expected", func() {
		Expect(dataTypesAlert.TriggerImmediate).To(Equal("immediate"))
	})

	It("TriggerRepeating is expected", func() {
		Expect(dataTypesAlert.TriggerRepeating).To(Equal("repeating"))
	})

	It("Priorities returns expected", func() {
		Expect(dataTypesAlert.Priorities()).To(Equal([]string{"critical", "normal", "timeSensitive"}))
	})

	It("Sounds returns expected", func() {
		Expect(dataTypesAlert.Sounds()).To(Equal([]string{"name", "silence", "vibrate"}))
	})

	It("Triggers returns expected", func() {
		Expect(dataTypesAlert.Triggers()).To(Equal([]string{"delayed", "immediate", "repeating"}))
	})

	Context("Alert", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesAlert.Alert)) {
				datum := dataTypesAlertTest.RandomAlert()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesAlertTest.NewObjectFromAlert(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesAlertTest.NewObjectFromAlert(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesAlert.Alert) {},
			),
			Entry("empty",
				func(datum *dataTypesAlert.Alert) {
					*datum = *dataTypesAlert.New()
				},
			),
			Entry("all",
				func(datum *dataTypesAlert.Alert) {
					datum.Name = pointer.FromString(test.RandomStringFromRange(1, dataTypesAlert.NameLengthMaximum))
					datum.Priority = pointer.FromString(test.RandomStringFromArray(dataTypesAlert.Priorities()))
					datum.Trigger = pointer.FromString(test.RandomStringFromArray(dataTypesAlert.Triggers()))
					datum.TriggerDelay = pointer.FromInt(test.RandomIntFromRange(dataTypesAlert.TriggerDelayMinimum, dataTypesAlert.TriggerDelayMaximum))
					datum.Sound = pointer.FromString(test.RandomStringFromArray(dataTypesAlert.Sounds()))
					datum.SoundName = pointer.FromString(test.RandomStringFromRange(1, dataTypesAlert.SoundNameLengthMaximum))
					datum.Parameters = metadataTest.RandomMetadata()
					datum.IssuedTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()))
					datum.AcknowledgedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.IssuedTime, time.Now()))
					datum.RetractedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.IssuedTime, time.Now()))
				},
			),
		)

		Context("New", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesAlert.New()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Type).To(Equal("alert"))
				Expect(datum.Name).To(BeNil())
				Expect(datum.Priority).To(BeNil())
				Expect(datum.Trigger).To(BeNil())
				Expect(datum.TriggerDelay).To(BeNil())
				Expect(datum.Sound).To(BeNil())
				Expect(datum.SoundName).To(BeNil())
				Expect(datum.Parameters).To(BeNil())
				Expect(datum.IssuedTime).To(BeNil())
				Expect(datum.AcknowledgedTime).To(BeNil())
				Expect(datum.RetractedTime).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesAlert.Alert), expectedErrors ...error) {
					expectedDatum := dataTypesAlertTest.RandomAlertForParser()
					object := dataTypesAlertTest.NewObjectFromAlert(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesAlert.New()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesAlert.Alert) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesAlert.Alert) {
						object["name"] = true
						object["priority"] = true
						object["trigger"] = true
						object["triggerDelay"] = true
						object["sound"] = true
						object["soundName"] = true
						object["parameters"] = true
						object["issuedTime"] = true
						object["acknowledgedTime"] = true
						object["retractedTime"] = true
						expectedDatum.Name = nil
						expectedDatum.Priority = nil
						expectedDatum.Trigger = nil
						expectedDatum.TriggerDelay = nil
						expectedDatum.Sound = nil
						expectedDatum.SoundName = nil
						expectedDatum.Parameters = nil
						expectedDatum.IssuedTime = nil
						expectedDatum.AcknowledgedTime = nil
						expectedDatum.RetractedTime = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/name", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/priority", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/trigger", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotInt(true), "/triggerDelay", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/sound", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/soundName", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/parameters", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotTime(true), "/issuedTime", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotTime(true), "/acknowledgedTime", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotTime(true), "/retractedTime", NewMeta()),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesAlert.Alert), expectedErrors ...error) {
					datum := dataTypesAlertTest.RandomAlert()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesAlert.Alert) {},
				),
				Entry("type missing",
					func(datum *dataTypesAlert.Alert) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypes.Meta{}),
				),
				Entry("type invalid",
					func(datum *dataTypesAlert.Alert) {
						datum.Type = "invalidType"
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "alert"), "/type", &dataTypes.Meta{Type: "invalidType"}),
				),
				Entry("type alert",
					func(datum *dataTypesAlert.Alert) {
						datum.Type = "alert"
					},
				),
				Entry("name missing",
					func(datum *dataTypesAlert.Alert) {
						datum.Name = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/name", NewMeta()),
				),
				Entry("name empty",
					func(datum *dataTypesAlert.Alert) {
						datum.Name = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/name", NewMeta()),
				),
				Entry("name length in range (upper)",
					func(datum *dataTypesAlert.Alert) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("name length out of range (upper)",
					func(datum *dataTypesAlert.Alert) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name", NewMeta()),
				),
				Entry("priority missing",
					func(datum *dataTypesAlert.Alert) {
						datum.Priority = nil
					},
				),
				Entry("priority invalid",
					func(datum *dataTypesAlert.Alert) {
						datum.Priority = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"critical", "normal", "timeSensitive"}), "/priority", NewMeta()),
				),
				Entry("priority critical",
					func(datum *dataTypesAlert.Alert) {
						datum.Priority = pointer.FromString(dataTypesAlert.PriorityCritical)
					},
				),
				Entry("priority normal",
					func(datum *dataTypesAlert.Alert) {
						datum.Priority = pointer.FromString(dataTypesAlert.PriorityNormal)
					},
				),
				Entry("priority time sensitive",
					func(datum *dataTypesAlert.Alert) {
						datum.Priority = pointer.FromString(dataTypesAlert.PriorityTimeSensitive)
					},
				),
				Entry("trigger missing; trigger delay missing",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = nil
						datum.TriggerDelay = nil
					},
				),
				Entry("trigger missing; trigger delay exists",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = nil
						datum.TriggerDelay = pointer.FromInt(test.RandomIntFromRange(dataTypesAlert.TriggerDelayMinimum, dataTypesAlert.TriggerDelayMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/triggerDelay", NewMeta()),
				),
				Entry("trigger invalid; trigger delay missing",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = pointer.FromString("invalid")
						datum.TriggerDelay = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"delayed", "immediate", "repeating"}), "/trigger", NewMeta()),
				),
				Entry("trigger invalid; trigger delay exists",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = pointer.FromString("invalid")
						datum.TriggerDelay = pointer.FromInt(test.RandomIntFromRange(dataTypesAlert.TriggerDelayMinimum, dataTypesAlert.TriggerDelayMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"delayed", "immediate", "repeating"}), "/trigger", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/triggerDelay", NewMeta()),
				),
				Entry("trigger delayed; trigger delay missing",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = pointer.FromString(dataTypesAlert.TriggerDelayed)
						datum.TriggerDelay = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/triggerDelay", NewMeta()),
				),
				Entry("trigger delayed; trigger delay exists",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = pointer.FromString(dataTypesAlert.TriggerDelayed)
						datum.TriggerDelay = pointer.FromInt(test.RandomIntFromRange(dataTypesAlert.TriggerDelayMinimum, dataTypesAlert.TriggerDelayMaximum))
					},
				),
				Entry("trigger immediate; trigger delay missing",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = pointer.FromString(dataTypesAlert.TriggerImmediate)
						datum.TriggerDelay = nil
					},
				),
				Entry("trigger immediate; trigger delay exists",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = pointer.FromString(dataTypesAlert.TriggerImmediate)
						datum.TriggerDelay = pointer.FromInt(test.RandomIntFromRange(dataTypesAlert.TriggerDelayMinimum, dataTypesAlert.TriggerDelayMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/triggerDelay", NewMeta()),
				),
				Entry("trigger repeating; trigger delay missing",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = pointer.FromString(dataTypesAlert.TriggerRepeating)
						datum.TriggerDelay = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/triggerDelay", NewMeta()),
				),
				Entry("trigger repeating; trigger delay exists",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = pointer.FromString(dataTypesAlert.TriggerRepeating)
						datum.TriggerDelay = pointer.FromInt(test.RandomIntFromRange(dataTypesAlert.TriggerDelayMinimum, dataTypesAlert.TriggerDelayMaximum))
					},
				),
				Entry("trigger delay; out of range (lower)",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = pointer.FromString(dataTypesAlert.TriggerDelayed)
						datum.TriggerDelay = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400), "/triggerDelay", NewMeta()),
				),
				Entry("trigger delay; in range (lower)",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = pointer.FromString(dataTypesAlert.TriggerDelayed)
						datum.TriggerDelay = pointer.FromInt(0)
					},
				),
				Entry("trigger delay; in range (upper)",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = pointer.FromString(dataTypesAlert.TriggerDelayed)
						datum.TriggerDelay = pointer.FromInt(86400)
					},
				),
				Entry("trigger delay; out of range (upper)",
					func(datum *dataTypesAlert.Alert) {
						datum.Trigger = pointer.FromString(dataTypesAlert.TriggerDelayed)
						datum.TriggerDelay = pointer.FromInt(86401)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86401, 0, 86400), "/triggerDelay", NewMeta()),
				),
				Entry("sound missing; sound delay missing",
					func(datum *dataTypesAlert.Alert) {
						datum.Sound = nil
						datum.SoundName = nil
					},
				),
				Entry("sound missing; sound name exists",
					func(datum *dataTypesAlert.Alert) {
						datum.Sound = nil
						datum.SoundName = pointer.FromString(test.RandomStringFromRange(1, dataTypesAlert.SoundNameLengthMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/soundName", NewMeta()),
				),
				Entry("sound invalid; sound name missing",
					func(datum *dataTypesAlert.Alert) {
						datum.Sound = pointer.FromString("invalid")
						datum.SoundName = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"name", "silence", "vibrate"}), "/sound", NewMeta()),
				),
				Entry("sound invalid; sound name exists",
					func(datum *dataTypesAlert.Alert) {
						datum.Sound = pointer.FromString("invalid")
						datum.SoundName = pointer.FromString(test.RandomStringFromRange(1, dataTypesAlert.SoundNameLengthMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"name", "silence", "vibrate"}), "/sound", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/soundName", NewMeta()),
				),
				Entry("sound name; sound name missing",
					func(datum *dataTypesAlert.Alert) {
						datum.Sound = pointer.FromString(dataTypesAlert.SoundName)
						datum.SoundName = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/soundName", NewMeta()),
				),
				Entry("sound name; sound name exists",
					func(datum *dataTypesAlert.Alert) {
						datum.Sound = pointer.FromString(dataTypesAlert.SoundName)
						datum.SoundName = pointer.FromString(test.RandomStringFromRange(1, dataTypesAlert.SoundNameLengthMaximum))
					},
				),
				Entry("sound silence; sound name missing",
					func(datum *dataTypesAlert.Alert) {
						datum.Sound = pointer.FromString(dataTypesAlert.SoundSilence)
						datum.SoundName = nil
					},
				),
				Entry("sound silence; sound name exists",
					func(datum *dataTypesAlert.Alert) {
						datum.Sound = pointer.FromString(dataTypesAlert.SoundSilence)
						datum.SoundName = pointer.FromString(test.RandomStringFromRange(1, dataTypesAlert.SoundNameLengthMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/soundName", NewMeta()),
				),
				Entry("sound vibrate; sound name missing",
					func(datum *dataTypesAlert.Alert) {
						datum.Sound = pointer.FromString(dataTypesAlert.SoundVibrate)
						datum.SoundName = nil
					},
				),
				Entry("sound vibrate; sound name exists",
					func(datum *dataTypesAlert.Alert) {
						datum.Sound = pointer.FromString(dataTypesAlert.SoundVibrate)
						datum.SoundName = pointer.FromString(test.RandomStringFromRange(1, dataTypesAlert.SoundNameLengthMaximum))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/soundName", NewMeta()),
				),
				Entry("sound name empty",
					func(datum *dataTypesAlert.Alert) {
						datum.Sound = pointer.FromString(dataTypesAlert.SoundName)
						datum.SoundName = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/soundName", NewMeta()),
				),
				Entry("sound name length in range (upper)",
					func(datum *dataTypesAlert.Alert) {
						datum.Sound = pointer.FromString(dataTypesAlert.SoundName)
						datum.SoundName = pointer.FromString(test.RandomStringFromRange(1, 100))
					},
				),
				Entry("sound name length out of range (upper)",
					func(datum *dataTypesAlert.Alert) {
						datum.Sound = pointer.FromString(dataTypesAlert.SoundName)
						datum.SoundName = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/soundName", NewMeta()),
				),
				Entry("parameters missing",
					func(datum *dataTypesAlert.Alert) {
						datum.Parameters = nil
					},
				),
				Entry("parameters empty",
					func(datum *dataTypesAlert.Alert) {
						datum.Parameters = metadata.NewMetadata()
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/parameters", NewMeta()),
				),
				Entry("parameters valid",
					func(datum *dataTypesAlert.Alert) {
						datum.Parameters = metadataTest.RandomMetadata()
					},
				),
				Entry("issued time missing",
					func(datum *dataTypesAlert.Alert) {
						datum.IssuedTime = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/issuedTime", NewMeta()),
				),
				Entry("issued time zero",
					func(datum *dataTypesAlert.Alert) {
						datum.IssuedTime = pointer.FromTime(time.Time{})
						datum.AcknowledgedTime = nil
						datum.RetractedTime = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/issuedTime", NewMeta()),
				),
				Entry("issued time valid",
					func(datum *dataTypesAlert.Alert) {
						datum.IssuedTime = pointer.FromTime(time.Unix(1640955599, 0))
						datum.AcknowledgedTime = nil
						datum.RetractedTime = nil
					},
				),
				Entry("acknowledged time missing",
					func(datum *dataTypesAlert.Alert) {
						datum.IssuedTime = pointer.FromTime(time.Unix(1640955599, 0))
						datum.AcknowledgedTime = nil
						datum.RetractedTime = nil
					},
				),
				Entry("acknowledged time before issued time",
					func(datum *dataTypesAlert.Alert) {
						datum.IssuedTime = pointer.FromTime(time.Unix(1640955599, 0))
						datum.AcknowledgedTime = pointer.FromTime(time.Unix(1640955598, 0))
						datum.RetractedTime = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueTimeNotAfter(time.Unix(1640955598, 0), time.Unix(1640955599, 0)), "/acknowledgedTime", NewMeta()),
				),
				Entry("acknowledged time valid",
					func(datum *dataTypesAlert.Alert) {
						datum.IssuedTime = pointer.FromTime(time.Unix(1640955599, 0))
						datum.AcknowledgedTime = pointer.FromTime(time.Unix(1640955599, 0))
						datum.RetractedTime = nil
					},
				),
				Entry("retracted time missing",
					func(datum *dataTypesAlert.Alert) {
						datum.IssuedTime = pointer.FromTime(time.Unix(1640955599, 0))
						datum.AcknowledgedTime = nil
						datum.RetractedTime = nil
					},
				),
				Entry("retracted time before issued time",
					func(datum *dataTypesAlert.Alert) {
						datum.IssuedTime = pointer.FromTime(time.Unix(1640955599, 0))
						datum.AcknowledgedTime = nil
						datum.RetractedTime = pointer.FromTime(time.Unix(1640955598, 0))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueTimeNotAfter(time.Unix(1640955598, 0), time.Unix(1640955599, 0)), "/retractedTime", NewMeta()),
				),
				Entry("retracted time valid",
					func(datum *dataTypesAlert.Alert) {
						datum.IssuedTime = pointer.FromTime(time.Unix(1640955599, 0))
						datum.AcknowledgedTime = nil
						datum.RetractedTime = pointer.FromTime(time.Unix(1640955599, 0))
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesAlert.Alert) {
						datum.Type = "invalidType"
						datum.Name = nil
						datum.Priority = pointer.FromString("invalid")
						datum.Trigger = pointer.FromString("invalid")
						datum.TriggerDelay = pointer.FromInt(test.RandomIntFromRange(dataTypesAlert.TriggerDelayMinimum, dataTypesAlert.TriggerDelayMaximum))
						datum.Sound = pointer.FromString("invalid")
						datum.SoundName = pointer.FromString(test.RandomStringFromRange(1, dataTypesAlert.SoundNameLengthMaximum))
						datum.Parameters = metadata.NewMetadata()
						datum.IssuedTime = pointer.FromTime(time.Time{})
						datum.AcknowledgedTime = pointer.FromTime(time.Time{}.Add(-time.Second))
						datum.RetractedTime = pointer.FromTime(time.Time{}.Add(-time.Second))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "alert"), "/type", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/name", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"critical", "normal", "timeSensitive"}), "/priority", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"delayed", "immediate", "repeating"}), "/trigger", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/triggerDelay", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"name", "silence", "vibrate"}), "/sound", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/soundName", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/parameters", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/issuedTime", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueTimeNotAfter(time.Time{}.Add(-time.Second), time.Time{}), "/acknowledgedTime", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueTimeNotAfter(time.Time{}.Add(-time.Second), time.Time{}), "/retractedTime", &dataTypes.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesAlert.Alert), expectator func(datum *dataTypesAlert.Alert, expectedDatum *dataTypesAlert.Alert)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesAlertTest.RandomAlert()
						mutator(datum)
						expectedDatum := dataTypesAlertTest.CloneAlert(datum)
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
					func(datum *dataTypesAlert.Alert) {},
					nil,
				),
			)
		})
	})
})
