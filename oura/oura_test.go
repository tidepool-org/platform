package oura_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oura"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("oura", func() {
	It("DataTypeDailyActivity is expected", func() {
		Expect(oura.DataTypeDailyActivity).To(Equal("daily_activity"))
	})

	It("DataTypeDailyCardiovascularAge is expected", func() {
		Expect(oura.DataTypeDailyCardiovascularAge).To(Equal("daily_cardiovascular_age"))
	})

	It("DataTypeDailyReadiness is expected", func() {
		Expect(oura.DataTypeDailyReadiness).To(Equal("daily_readiness"))
	})

	It("DataTypeDailyResilience is expected", func() {
		Expect(oura.DataTypeDailyResilience).To(Equal("daily_resilience"))
	})

	It("DataTypeDailySleep is expected", func() {
		Expect(oura.DataTypeDailySleep).To(Equal("daily_sleep"))
	})

	It("DataTypeDailySpO2 is expected", func() {
		Expect(oura.DataTypeDailySpO2).To(Equal("daily_spo2"))
	})

	It("DataTypeDailyStress is expected", func() {
		Expect(oura.DataTypeDailyStress).To(Equal("daily_stress"))
	})

	It("DataTypeEnhancedTag is expected", func() {
		Expect(oura.DataTypeEnhancedTag).To(Equal("enhanced_tag"))
	})

	It("DataTypeHeartRate is expected", func() {
		Expect(oura.DataTypeHeartRate).To(Equal("heartrate"))
	})

	It("DataTypeRestModePeriod is expected", func() {
		Expect(oura.DataTypeRestModePeriod).To(Equal("rest_mode_period"))
	})

	It("DataTypeRingConfiguration is expected", func() {
		Expect(oura.DataTypeRingConfiguration).To(Equal("ring_configuration"))
	})

	It("DataTypeSession is expected", func() {
		Expect(oura.DataTypeSession).To(Equal("session"))
	})

	It("DataTypeSleep is expected", func() {
		Expect(oura.DataTypeSleep).To(Equal("sleep"))
	})

	It("DataTypeSleepTime is expected", func() {
		Expect(oura.DataTypeSleepTime).To(Equal("sleep_time"))
	})

	It("DataTypeVO2Max is expected", func() {
		Expect(oura.DataTypeVO2Max).To(Equal("vo2_max"))
	})

	It("DataTypeWorkout is expected", func() {
		Expect(oura.DataTypeWorkout).To(Equal("workout"))
	})

	It("EventTypeCreate is expected", func() {
		Expect(oura.EventTypeCreate).To(Equal("create"))
	})

	It("EventTypeUpdate is expected", func() {
		Expect(oura.EventTypeUpdate).To(Equal("update"))
	})

	It("EventTypeDelete is expected", func() {
		Expect(oura.EventTypeDelete).To(Equal("delete"))
	})

	It("ProviderName is expected", func() {
		Expect(oura.ProviderName).To(Equal("oura"))
	})

	It("PartnerName is expected", func() {
		Expect(oura.PartnerName).To(Equal("oura"))
	})

	It("PartnerPathPrefix is expected", func() {
		Expect(oura.PartnerPathPrefix).To(Equal("/v1/partners/oura"))
	})

	It("TimeRangeFormat is expected", func() {
		Expect(oura.TimeRangeFormat).To(Equal(time.RFC3339))
	})

	It("TimeRangeTruncatedDuration is expected", func() {
		Expect(oura.TimeRangeTruncatedDuration).To(Equal(time.Second))
	})

	It("TimeRangeMaximumYears is expected", func() {
		Expect(oura.TimeRangeMaximumYears).To(Equal(10))
	})

	Context("DataTypes", func() {
		It("returns expected data types", func() {
			Expect(oura.DataTypes()).To(Equal([]string{
				oura.DataTypeDailyActivity,
				oura.DataTypeDailyCardiovascularAge,
				oura.DataTypeDailyReadiness,
				oura.DataTypeDailyResilience,
				oura.DataTypeDailySleep,
				oura.DataTypeDailySpO2,
				oura.DataTypeDailyStress,
				oura.DataTypeEnhancedTag,
				oura.DataTypeHeartRate,
				oura.DataTypeRestModePeriod,
				oura.DataTypeRingConfiguration,
				oura.DataTypeSession,
				oura.DataTypeSleep,
				oura.DataTypeSleepTime,
				oura.DataTypeVO2Max,
				oura.DataTypeWorkout,
			}))
		})
	})

	Context("EventTypes", func() {
		It("returns expected event types", func() {
			Expect(oura.EventTypes()).To(Equal([]string{
				oura.EventTypeCreate,
				oura.EventTypeUpdate,
				oura.EventTypeDelete,
			}))
		})
	})

	Context("CreateSubscription", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.CreateSubscription)) {
				datum := ouraTest.RandomCreateSubscription(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraTest.NewObjectFromCreateSubscription(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *oura.CreateSubscription) {},
			),
			Entry("empty",
				func(datum *oura.CreateSubscription) {
					*datum = oura.CreateSubscription{}
				},
			),
			Entry("all",
				func(datum *oura.CreateSubscription) {
					*datum = *ouraTest.RandomCreateSubscription()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *oura.CreateSubscription), expectedErrors ...error) {
					expectedDatum := ouraTest.RandomCreateSubscription(test.AllowOptional())
					object := ouraTest.NewObjectFromCreateSubscription(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &oura.CreateSubscription{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *oura.CreateSubscription) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *oura.CreateSubscription) {
						clear(object)
						*expectedDatum = oura.CreateSubscription{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *oura.CreateSubscription) {
						object["callback_url"] = true
						object["verification_token"] = true
						object["event_type"] = true
						object["data_type"] = true
						expectedDatum.CallbackURL = nil
						expectedDatum.VerificationToken = nil
						expectedDatum.EventType = nil
						expectedDatum.DataType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/callback_url"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/verification_token"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/event_type"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/data_type"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *oura.CreateSubscription), expectedErrors ...error) {
					datum := ouraTest.RandomCreateSubscription(test.AllowOptional())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.CreateSubscription) {},
				),
				Entry("callback_url",
					func(datum *oura.CreateSubscription) {
						datum.CallbackURL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/callback_url"),
				),
				Entry("callback_url empty",
					func(datum *oura.CreateSubscription) {
						datum.CallbackURL = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/callback_url"),
				),
				Entry("callback_url valid",
					func(datum *oura.CreateSubscription) {
						datum.CallbackURL = pointer.FromString(ouraTest.RandomCallbackURL())
					},
				),
				Entry("verification_token",
					func(datum *oura.CreateSubscription) {
						datum.VerificationToken = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/verification_token"),
				),
				Entry("verification_token empty",
					func(datum *oura.CreateSubscription) {
						datum.VerificationToken = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/verification_token"),
				),
				Entry("verification_token valid",
					func(datum *oura.CreateSubscription) {
						datum.VerificationToken = pointer.FromString(ouraTest.RandomVerificationToken())
					},
				),
				Entry("event_type",
					func(datum *oura.CreateSubscription) {
						datum.EventType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
				),
				Entry("event_type invalid",
					func(datum *oura.CreateSubscription) {
						datum.EventType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventTypes()), "/event_type"),
				),
				Entry("event_type valid",
					func(datum *oura.CreateSubscription) {
						datum.EventType = pointer.FromString(ouraTest.RandomEventType())
					},
				),
				Entry("data_type",
					func(datum *oura.CreateSubscription) {
						datum.DataType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
				),
				Entry("data_type invalid",
					func(datum *oura.CreateSubscription) {
						datum.DataType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.DataTypes()), "/data_type"),
				),
				Entry("data_type valid",
					func(datum *oura.CreateSubscription) {
						datum.DataType = pointer.FromString(ouraTest.RandomDataType())
					},
				),
				Entry("multiple errors",
					func(datum *oura.CreateSubscription) {
						datum.CallbackURL = nil
						datum.VerificationToken = nil
						datum.EventType = nil
						datum.DataType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/callback_url"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/verification_token"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
				),
			)
		})
	})

	Context("Subscription", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.Subscription)) {
				datum := ouraTest.RandomSubscription(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraTest.NewObjectFromSubscription(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *oura.Subscription) {},
			),
			Entry("empty",
				func(datum *oura.Subscription) {
					*datum = oura.Subscription{}
				},
			),
			Entry("all",
				func(datum *oura.Subscription) {
					*datum = *ouraTest.RandomSubscription()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *oura.Subscription), expectedErrors ...error) {
					expectedDatum := ouraTest.RandomSubscription(test.AllowOptional())
					object := ouraTest.NewObjectFromSubscription(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &oura.Subscription{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *oura.Subscription) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *oura.Subscription) {
						clear(object)
						*expectedDatum = oura.Subscription{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *oura.Subscription) {
						object["id"] = true
						object["callback_url"] = true
						object["event_type"] = true
						object["data_type"] = true
						object["expiration_time"] = true
						expectedDatum.ID = nil
						expectedDatum.CallbackURL = nil
						expectedDatum.EventType = nil
						expectedDatum.DataType = nil
						expectedDatum.ExpirationTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/id"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/callback_url"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/event_type"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/data_type"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/expiration_time"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *oura.Subscription), expectedErrors ...error) {
					datum := ouraTest.RandomSubscription(test.AllowOptional())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.Subscription) {},
				),
				Entry("id",
					func(datum *oura.Subscription) {
						datum.ID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
				),
				Entry("id empty",
					func(datum *oura.Subscription) {
						datum.ID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id valid",
					func(datum *oura.Subscription) {
						datum.ID = pointer.FromString(ouraTest.RandomID())
					},
				),
				Entry("callback_url",
					func(datum *oura.Subscription) {
						datum.CallbackURL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/callback_url"),
				),
				Entry("callback_url empty",
					func(datum *oura.Subscription) {
						datum.CallbackURL = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/callback_url"),
				),
				Entry("callback_url valid",
					func(datum *oura.Subscription) {
						datum.CallbackURL = pointer.FromString(ouraTest.RandomCallbackURL())
					},
				),
				Entry("event_type",
					func(datum *oura.Subscription) {
						datum.EventType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
				),
				Entry("event_type invalid",
					func(datum *oura.Subscription) {
						datum.EventType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventTypes()), "/event_type"),
				),
				Entry("event_type valid",
					func(datum *oura.Subscription) {
						datum.EventType = pointer.FromString(ouraTest.RandomEventType())
					},
				),
				Entry("data_type",
					func(datum *oura.Subscription) {
						datum.DataType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
				),
				Entry("data_type invalid",
					func(datum *oura.Subscription) {
						datum.DataType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.DataTypes()), "/data_type"),
				),
				Entry("data_type valid",
					func(datum *oura.Subscription) {
						datum.DataType = pointer.FromString(ouraTest.RandomDataType())
					},
				),
				Entry("expiration_time",
					func(datum *oura.Subscription) {
						datum.ExpirationTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/expiration_time"),
				),
				Entry("expiration_time zero",
					func(datum *oura.Subscription) {
						datum.ExpirationTime = pointer.FromTime(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/expiration_time"),
				),
				Entry("expiration_time valid",
					func(datum *oura.Subscription) {
						datum.ExpirationTime = pointer.FromTime(test.RandomTime())
					},
				),
				Entry("multiple errors",
					func(datum *oura.Subscription) {
						datum.ID = nil
						datum.CallbackURL = nil
						datum.EventType = nil
						datum.DataType = nil
						datum.ExpirationTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/callback_url"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/expiration_time"),
				),
			)
		})
	})

	Context("Subscriptions", func() {
	})

	Context("PersonalInfo", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.PersonalInfo)) {
				datum := ouraTest.RandomPersonalInfo(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraTest.NewObjectFromPersonalInfo(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *oura.PersonalInfo) {},
			),
			Entry("empty",
				func(datum *oura.PersonalInfo) {
					*datum = oura.PersonalInfo{}
				},
			),
			Entry("all",
				func(datum *oura.PersonalInfo) {
					*datum = *ouraTest.RandomPersonalInfo()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *oura.PersonalInfo), expectedErrors ...error) {
					expectedDatum := ouraTest.RandomPersonalInfo(test.AllowOptional())
					object := ouraTest.NewObjectFromPersonalInfo(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &oura.PersonalInfo{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *oura.PersonalInfo) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *oura.PersonalInfo) {
						clear(object)
						*expectedDatum = oura.PersonalInfo{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *oura.PersonalInfo) {
						object["id"] = true
						object["age"] = true
						object["weight"] = true
						object["height"] = true
						object["biological_sex"] = true
						object["email"] = true
						expectedDatum.ID = nil
						expectedDatum.Age = nil
						expectedDatum.Weight = nil
						expectedDatum.Height = nil
						expectedDatum.BiologicalSex = nil
						expectedDatum.Email = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/id"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/age"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/weight"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/height"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/biological_sex"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/email"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *oura.PersonalInfo), expectedErrors ...error) {
					datum := ouraTest.RandomPersonalInfo(test.AllowOptional())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.PersonalInfo) {},
				),
				Entry("id",
					func(datum *oura.PersonalInfo) {
						datum.ID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
				),
				Entry("id empty",
					func(datum *oura.PersonalInfo) {
						datum.ID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id valid",
					func(datum *oura.PersonalInfo) {
						datum.ID = pointer.FromString(ouraTest.RandomID())
					},
				),
			)
		})
	})
})
