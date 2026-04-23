package webhook_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oura"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	ouraWebhookTest "github.com/tidepool-org/platform/oura/webhook/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Webhook", func() {
	It("EventPath is expected", func() {
		Expect(ouraWebhook.EventPath).To(Equal("/event"))
	})

	Context("DataTypes", func() {
		It("returns expected data types", func() {
			Expect(ouraWebhook.DataTypes()).To(Equal([]string{
				oura.DataTypeDailyActivity,
				oura.DataTypeDailyCyclePhases,
				oura.DataTypeDailyReadiness,
				oura.DataTypeDailyResilience,
				oura.DataTypeDailySleep,
				oura.DataTypeDailySpO2,
				oura.DataTypeDailyStress,
				oura.DataTypeEnhancedTag,
				oura.DataTypeRestModePeriod,
				oura.DataTypeRingConfiguration,
				oura.DataTypeSession,
				oura.DataTypeSleep,
				oura.DataTypeSleepTime,
				oura.DataTypeWorkout,
			}))
		})
	})

	Context("Event", func() {
		Context("ParseEvent", func() {
			It("returns nil if the object does not exist", func() {
				Expect(ouraWebhook.ParseEvent(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("parses the datum", func() {
				datum := ouraWebhookTest.RandomEvent(test.AllowOptionals())
				object := ouraWebhookTest.NewObjectFromEvent(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(ouraWebhook.ParseEvent(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *ouraWebhook.Event)) {
				datum := ouraWebhookTest.RandomEvent(test.AllowOptionals())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraWebhookTest.NewObjectFromEvent(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraWebhookTest.NewObjectFromEvent(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *ouraWebhook.Event) {},
			),
			Entry("empty",
				func(datum *ouraWebhook.Event) {
					*datum = ouraWebhook.Event{}
				},
			),
			Entry("all",
				func(datum *ouraWebhook.Event) {
					*datum = *ouraWebhookTest.RandomEvent()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *ouraWebhook.Event), expectedErrors ...error) {
					expectedDatum := ouraWebhookTest.RandomEvent(test.AllowOptionals())
					object := ouraWebhookTest.NewObjectFromEvent(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &ouraWebhook.Event{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *ouraWebhook.Event) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *ouraWebhook.Event) {
						clear(object)
						*expectedDatum = ouraWebhook.Event{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *ouraWebhook.Event) {
						object["event_time"] = true
						object["event_type"] = true
						object["user_id"] = true
						object["object_id"] = true
						object["data_type"] = true
						expectedDatum.EventTime = nil
						expectedDatum.EventType = nil
						expectedDatum.UserID = nil
						expectedDatum.ObjectID = nil
						expectedDatum.DataType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/event_time"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/event_type"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/user_id"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/object_id"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/data_type"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *ouraWebhook.Event), expectedErrors ...error) {
					datum := ouraWebhookTest.RandomEvent(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *ouraWebhook.Event) {},
				),
				Entry("event_time",
					func(datum *ouraWebhook.Event) {
						datum.EventTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_time"),
				),
				Entry("event_time zero",
					func(datum *ouraWebhook.Event) {
						datum.EventTime = pointer.From(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/event_time"),
				),
				Entry("event_time valid",
					func(datum *ouraWebhook.Event) {
						datum.EventTime = pointer.From(test.RandomTime())
					},
				),
				Entry("event_type",
					func(datum *ouraWebhook.Event) {
						datum.EventType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
				),
				Entry("event_type invalid",
					func(datum *ouraWebhook.Event) {
						datum.EventType = pointer.From("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventTypes()), "/event_type"),
				),
				Entry("event_type valid",
					func(datum *ouraWebhook.Event) {
						datum.EventType = pointer.From(test.RandomStringFromArray(oura.EventTypes()))
					},
				),
				Entry("user_id",
					func(datum *ouraWebhook.Event) {
						datum.UserID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/user_id"),
				),
				Entry("user_id empty",
					func(datum *ouraWebhook.Event) {
						datum.UserID = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/user_id"),
				),
				Entry("user_id valid",
					func(datum *ouraWebhook.Event) {
						datum.UserID = pointer.From(test.RandomString())
					},
				),
				Entry("object_id",
					func(datum *ouraWebhook.Event) {
						datum.ObjectID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/object_id"),
				),
				Entry("object_id zero",
					func(datum *ouraWebhook.Event) {
						datum.ObjectID = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/object_id"),
				),
				Entry("object_id valid",
					func(datum *ouraWebhook.Event) {
						datum.ObjectID = pointer.From(test.RandomString())
					},
				),
				Entry("data_type",
					func(datum *ouraWebhook.Event) {
						datum.DataType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
				),
				Entry("data_type invalid",
					func(datum *ouraWebhook.Event) {
						datum.DataType = pointer.From("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", ouraWebhook.DataTypes()), "/data_type"),
				),
				Entry("data_type valid",
					func(datum *ouraWebhook.Event) {
						datum.DataType = pointer.From(test.RandomStringFromArray(ouraWebhook.DataTypes()))
					},
				),
				Entry("multiple errors",
					func(datum *ouraWebhook.Event) {
						datum.EventTime = nil
						datum.EventType = nil
						datum.UserID = nil
						datum.ObjectID = nil
						datum.DataType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_time"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/user_id"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/object_id"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
				),
			)
		})

		Context("with event", func() {
			var eventTime, _ = time.ParseInLocation(time.RFC3339Nano, "2026-01-15T20:15:42.123Z", time.UTC)

			DescribeTable("String returns expected string",
				func(eventTime *time.Time, eventType *string, userID *string, objectID *string, dataType *string, expectedString string) {
					datum := &ouraWebhook.Event{
						EventTime: eventTime,
						EventType: eventType,
						UserID:    userID,
						ObjectID:  objectID,
						DataType:  dataType,
					}
					Expect(datum.String()).To(Equal(expectedString))
				},
				Entry("all", pointer.From(eventTime), pointer.From("alpha"), pointer.From("beta"), pointer.From("charlie"), pointer.From("delta"), "2026-01-15T20:15:42.123Z:alpha:beta:charlie:delta"),
				Entry("event_time missing", nil, pointer.From("alpha"), pointer.From("beta"), pointer.From("charlie"), pointer.From("delta"), ":alpha:beta:charlie:delta"),
				Entry("event_type missing", pointer.From(eventTime), nil, pointer.From("beta"), pointer.From("charlie"), pointer.From("delta"), "2026-01-15T20:15:42.123Z::beta:charlie:delta"),
				Entry("user_id missing", pointer.From(eventTime), pointer.From("alpha"), nil, pointer.From("charlie"), pointer.From("delta"), "2026-01-15T20:15:42.123Z:alpha::charlie:delta"),
				Entry("object_id missing", pointer.From(eventTime), pointer.From("alpha"), pointer.From("beta"), nil, pointer.From("delta"), "2026-01-15T20:15:42.123Z:alpha:beta::delta"),
				Entry("data_type missing", pointer.From(eventTime), pointer.From("alpha"), pointer.From("beta"), pointer.From("charlie"), nil, "2026-01-15T20:15:42.123Z:alpha:beta:charlie:"),
				Entry("all missing", nil, nil, nil, nil, nil, "::::"),
			)
		})
	})

	Context("EventMetadata", func() {
		It("MetadataKeyEvent is expected", func() {
			Expect(ouraWebhook.MetadataKeyEvent).To(Equal("event"))
		})

		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *ouraWebhook.EventMetadata)) {
				datum := ouraWebhookTest.RandomEventMetadata(test.AllowOptionals())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraWebhookTest.NewObjectFromEventMetadata(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraWebhookTest.NewObjectFromEventMetadata(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *ouraWebhook.EventMetadata) {},
			),
			Entry("empty",
				func(datum *ouraWebhook.EventMetadata) {
					*datum = ouraWebhook.EventMetadata{}
				},
			),
			Entry("all",
				func(datum *ouraWebhook.EventMetadata) {
					datum.Event = ouraWebhookTest.RandomEvent()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *ouraWebhook.EventMetadata), expectedErrors ...error) {
					expectedDatum := ouraWebhookTest.RandomEventMetadata(test.AllowOptionals())
					object := ouraWebhookTest.NewObjectFromEventMetadata(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &ouraWebhook.EventMetadata{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *ouraWebhook.EventMetadata) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *ouraWebhook.EventMetadata) {
						clear(object)
						*expectedDatum = ouraWebhook.EventMetadata{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *ouraWebhook.EventMetadata) {
						object["event"] = true
						expectedDatum.Event = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/event"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *ouraWebhook.EventMetadata), expectedErrors ...error) {
					datum := ouraWebhookTest.RandomEventMetadata(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *ouraWebhook.EventMetadata) {},
				),
				Entry("event missing",
					func(datum *ouraWebhook.EventMetadata) {
						datum.Event = nil
					},
				),
				Entry("event invalid",
					func(datum *ouraWebhook.EventMetadata) {
						datum.Event = ouraWebhookTest.RandomEvent(test.AllowOptionals())
						datum.Event.EventTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event/event_time"),
				),
				Entry("multiple errors",
					func(datum *ouraWebhook.EventMetadata) {
						datum.Event = ouraWebhookTest.RandomEvent(test.AllowOptionals())
						datum.Event.EventTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event/event_time"),
				),
			)
		})
	})
})
