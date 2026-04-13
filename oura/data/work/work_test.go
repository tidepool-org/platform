package work_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authTest "github.com/tidepool-org/platform/auth/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
	ouraDataWorkTest "github.com/tidepool-org/platform/oura/data/work/test"
	ouraWebhookTest "github.com/tidepool-org/platform/oura/webhook/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	timesTest "github.com/tidepool-org/platform/times/test"
)

var _ = Describe("work", func() {
	It("Domain is expected", func() {
		Expect(ouraDataWork.Domain).To(Equal("org.tidepool.oura.data"))
	})

	Context("Metadata", func() {
		It("MetadataKeyScope is expected", func() {
			Expect(ouraDataWork.MetadataKeyScope).To(Equal("scope"))
		})

		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *ouraDataWork.Metadata)) {
				datum := ouraDataWorkTest.RandomMetadata(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraDataWorkTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraDataWorkTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *ouraDataWork.Metadata) {},
			),
			Entry("empty",
				func(datum *ouraDataWork.Metadata) {
					*datum = ouraDataWork.Metadata{}
				},
			),
			Entry("all",
				func(datum *ouraDataWork.Metadata) {
					*datum = *ouraDataWorkTest.RandomMetadata()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *ouraDataWork.Metadata), expectedErrors ...error) {
					expectedDatum := ouraDataWorkTest.RandomMetadata(test.AllowOptional())
					object := ouraDataWorkTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &ouraDataWork.Metadata{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *ouraDataWork.Metadata) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *ouraDataWork.Metadata) {
						clear(object)
						*expectedDatum = ouraDataWork.Metadata{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *ouraDataWork.Metadata) {
						object["scope"] = true
						object["timeRange"] = true
						object["event"] = true
						expectedDatum.Scope = nil
						expectedDatum.TimeRange = nil
						expectedDatum.Event = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/scope"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/timeRange"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/event"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *ouraDataWork.Metadata), expectedErrors ...error) {
					datum := ouraDataWorkTest.RandomMetadata(test.AllowOptional())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *ouraDataWork.Metadata) {},
				),
				Entry("scope missing",
					func(datum *ouraDataWork.Metadata) {
						datum.Scope = nil
					},
				),
				Entry("scope empty",
					func(datum *ouraDataWork.Metadata) {
						datum.Scope = pointer.From([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/scope"),
				),
				Entry("scope valid",
					func(datum *ouraDataWork.Metadata) {
						datum.Scope = pointer.From(authTest.RandomScope())
					},
				),
				Entry("time range missing",
					func(datum *ouraDataWork.Metadata) {
						datum.TimeRange = nil
						datum.Event = ouraWebhookTest.RandomEvent(test.AllowOptional())
					},
				),
				Entry("time range invalid",
					func(datum *ouraDataWork.Metadata) {
						datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptional())
						datum.TimeRange.From = pointer.From(time.Time{})
						datum.Event = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timeRange/from"),
				),
				Entry("time range valid",
					func(datum *ouraDataWork.Metadata) {
						datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptional())
						datum.Event = nil
					},
				),
				Entry("event missing",
					func(datum *ouraDataWork.Metadata) {
						datum.Event = nil
						datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptional())
					},
				),
				Entry("event invalid",
					func(datum *ouraDataWork.Metadata) {
						datum.TimeRange = nil
						datum.Event = ouraWebhookTest.RandomEvent(test.AllowOptional())
						datum.Event.EventTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event/event_time"),
				),
				Entry("event valid",
					func(datum *ouraDataWork.Metadata) {
						datum.TimeRange = nil
						datum.Event = ouraWebhookTest.RandomEvent(test.AllowOptional())
					},
				),
				Entry("neither time range nor event",
					func(datum *ouraDataWork.Metadata) {
						datum.TimeRange = nil
						datum.Event = nil
					},
					structureValidator.ErrorValuesNotExistForOne("event", "timeRange"),
				),
				Entry("both time range and event",
					func(datum *ouraDataWork.Metadata) {
						datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptional())
						datum.Event = ouraWebhookTest.RandomEvent(test.AllowOptional())
					},
					structureValidator.ErrorValuesNotExistForOne("event", "timeRange"),
				),
				Entry("multiple errors",
					func(datum *ouraDataWork.Metadata) {
						datum.Scope = pointer.From([]string{})
						datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptional())
						datum.TimeRange.From = pointer.From(time.Time{})
						datum.Event = ouraWebhookTest.RandomEvent(test.AllowOptional())
						datum.Event.EventTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/scope"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timeRange/from"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event/event_time"),
					structureValidator.ErrorValuesNotExistForOne("event", "timeRange"),
				),
			)
		})
	})

	Context("SerialIDFromProviderSessionID", func() {
		It("returns expected", func() {
			providerSessionID := authTest.RandomProviderSessionID()
			Expect(ouraDataWork.SerialIDFromProviderSessionID(providerSessionID)).To(Equal(ouraDataWork.Domain + ":" + providerSessionID))
		})
	})
})
