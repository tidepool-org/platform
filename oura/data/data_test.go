package data_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oura"
	ouraData "github.com/tidepool-org/platform/oura/data"
	ouraDataTest "github.com/tidepool-org/platform/oura/data/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	timesTest "github.com/tidepool-org/platform/times/test"
)

var _ = Describe("data", func() {
	It("MetadataKeyDataType is expected", func() {
		Expect(ouraData.MetadataKeyDataType).To(Equal("dataType"))
	})

	Context("Metadata", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *ouraData.Metadata)) {
				datum := ouraDataTest.RandomMetadata(test.AllowOptionals())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraDataTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraDataTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *ouraData.Metadata) {},
			),
			Entry("empty",
				func(datum *ouraData.Metadata) {
					*datum = ouraData.Metadata{}
				},
			),
			Entry("all",
				func(datum *ouraData.Metadata) {
					datum.DataType = ouraTest.RandomDataType()
					datum.Event = ouraTest.RandomEvent()
					datum.TimeRange = timesTest.RandomTimeRange()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *ouraData.Metadata), expectedErrors ...error) {
					expectedDatum := ouraDataTest.RandomMetadata(test.AllowOptionals())
					object := ouraDataTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &ouraData.Metadata{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *ouraData.Metadata) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *ouraData.Metadata) {
						clear(object)
						*expectedDatum = ouraData.Metadata{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *ouraData.Metadata) {
						object["dataType"] = true
						object["event"] = true
						object["timeRange"] = true
						expectedDatum.DataType = ""
						expectedDatum.Event = nil
						expectedDatum.TimeRange = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/dataType"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/event"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/timeRange"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *ouraData.Metadata), expectedErrors ...error) {
					datum := ouraDataTest.RandomMetadata(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *ouraData.Metadata) {},
				),
				Entry("dataType missing",
					func(datum *ouraData.Metadata) {
						datum.DataType = ""
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", oura.DataTypes()), "/dataType"),
				),
				Entry("dataType invalid",
					func(datum *ouraData.Metadata) {
						datum.DataType = "invalid"
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.DataTypes()), "/dataType"),
				),
				Entry("dataType valid",
					func(datum *ouraData.Metadata) {
						datum.DataType = ouraTest.RandomDataType()
					},
				),
				Entry("event missing",
					func(datum *ouraData.Metadata) {
						datum.Event = nil
						datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptionals())
					},
				),
				Entry("event invalid",
					func(datum *ouraData.Metadata) {
						datum.Event = ouraTest.RandomEvent(test.AllowOptionals())
						datum.Event.EventTime = nil
						datum.TimeRange = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event/event_time"),
				),
				Entry("event valid",
					func(datum *ouraData.Metadata) {
						datum.Event = ouraTest.RandomEvent(test.AllowOptionals())
						datum.TimeRange = nil
					},
				),
				Entry("timeRange missing",
					func(datum *ouraData.Metadata) {
						datum.Event = ouraTest.RandomEvent(test.AllowOptionals())
						datum.TimeRange = nil
					},
				),
				Entry("timeRange invalid",
					func(datum *ouraData.Metadata) {
						datum.Event = nil
						datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptionals())
						datum.TimeRange.From = pointer.From(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timeRange/from"),
				),
				Entry("timeRange valid",
					func(datum *ouraData.Metadata) {
						datum.Event = nil
						datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptionals())
					},
				),
				Entry("neither event nor timeRange",
					func(datum *ouraData.Metadata) {
						datum.Event = nil
						datum.TimeRange = nil
					},
					structureValidator.ErrorValuesNotExistForOne("event", "timeRange"),
				),
				Entry("both event and timeRange",
					func(datum *ouraData.Metadata) {
						datum.Event = ouraTest.RandomEvent(test.AllowOptionals())
						datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptionals())
					},
					structureValidator.ErrorValuesNotExistForOne("event", "timeRange"),
				),
				Entry("multiple errors",
					func(datum *ouraData.Metadata) {
						datum.DataType = "invalid"
						datum.Event = ouraTest.RandomEvent(test.AllowOptionals())
						datum.Event.EventTime = nil
						datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptionals())
						datum.TimeRange.From = pointer.From(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.DataTypes()), "/dataType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event/event_time"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timeRange/from"),
					structureValidator.ErrorValuesNotExistForOne("event", "timeRange"),
				),
			)
		})
	})
})
