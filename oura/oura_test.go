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
	structureTest "github.com/tidepool-org/platform/structure/test"
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

	It("SubscriptionArrayLengthMaximum is expected", func() {
		Expect(oura.SubscriptionArrayLengthMaximum).To(Equal(100))
	})

	It("SubscriptionExpirationTimeFormat is expected", func() {
		Expect(oura.SubscriptionExpirationTimeFormat).To(Equal("2006-01-02T15:04:05.999999999"))
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
				test.ExpectSerializedObjectBSON(datum, ouraTest.NewObjectFromCreateSubscription(datum, test.ObjectFormatBSON))
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
						object["data_type"] = true
						object["event_type"] = true
						expectedDatum.CallbackURL = nil
						expectedDatum.VerificationToken = nil
						expectedDatum.DataType = nil
						expectedDatum.EventType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/callback_url"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/verification_token"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/data_type"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/event_type"),
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
				Entry("multiple errors",
					func(datum *oura.CreateSubscription) {
						datum.CallbackURL = nil
						datum.VerificationToken = nil
						datum.DataType = nil
						datum.EventType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/callback_url"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/verification_token"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
				),
			)
		})
	})

	Context("UpdateSubscription", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.UpdateSubscription)) {
				datum := ouraTest.RandomUpdateSubscription(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraTest.NewObjectFromUpdateSubscription(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraTest.NewObjectFromUpdateSubscription(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *oura.UpdateSubscription) {},
			),
			Entry("empty",
				func(datum *oura.UpdateSubscription) {
					*datum = oura.UpdateSubscription{}
				},
			),
			Entry("all",
				func(datum *oura.UpdateSubscription) {
					*datum = *ouraTest.RandomUpdateSubscription()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *oura.UpdateSubscription), expectedErrors ...error) {
					expectedDatum := ouraTest.RandomUpdateSubscription(test.AllowOptional())
					object := ouraTest.NewObjectFromUpdateSubscription(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &oura.UpdateSubscription{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *oura.UpdateSubscription) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *oura.UpdateSubscription) {
						clear(object)
						*expectedDatum = oura.UpdateSubscription{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *oura.UpdateSubscription) {
						object["callback_url"] = true
						object["verification_token"] = true
						object["data_type"] = true
						object["event_type"] = true
						expectedDatum.CallbackURL = nil
						expectedDatum.VerificationToken = nil
						expectedDatum.DataType = nil
						expectedDatum.EventType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/callback_url"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/verification_token"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/data_type"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/event_type"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *oura.UpdateSubscription), expectedErrors ...error) {
					datum := ouraTest.RandomUpdateSubscription(test.AllowOptional())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.UpdateSubscription) {},
				),
				Entry("callback_url",
					func(datum *oura.UpdateSubscription) {
						datum.CallbackURL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/callback_url"),
				),
				Entry("callback_url empty",
					func(datum *oura.UpdateSubscription) {
						datum.CallbackURL = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/callback_url"),
				),
				Entry("callback_url valid",
					func(datum *oura.UpdateSubscription) {
						datum.CallbackURL = pointer.FromString(ouraTest.RandomCallbackURL())
					},
				),
				Entry("verification_token",
					func(datum *oura.UpdateSubscription) {
						datum.VerificationToken = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/verification_token"),
				),
				Entry("verification_token empty",
					func(datum *oura.UpdateSubscription) {
						datum.VerificationToken = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/verification_token"),
				),
				Entry("verification_token valid",
					func(datum *oura.UpdateSubscription) {
						datum.VerificationToken = pointer.FromString(ouraTest.RandomVerificationToken())
					},
				),
				Entry("data_type",
					func(datum *oura.UpdateSubscription) {
						datum.DataType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
				),
				Entry("data_type invalid",
					func(datum *oura.UpdateSubscription) {
						datum.DataType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.DataTypes()), "/data_type"),
				),
				Entry("data_type valid",
					func(datum *oura.UpdateSubscription) {
						datum.DataType = pointer.FromString(ouraTest.RandomDataType())
					},
				),
				Entry("event_type",
					func(datum *oura.UpdateSubscription) {
						datum.EventType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
				),
				Entry("event_type invalid",
					func(datum *oura.UpdateSubscription) {
						datum.EventType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventTypes()), "/event_type"),
				),
				Entry("event_type valid",
					func(datum *oura.UpdateSubscription) {
						datum.EventType = pointer.FromString(ouraTest.RandomEventType())
					},
				),
				Entry("multiple errors",
					func(datum *oura.UpdateSubscription) {
						datum.CallbackURL = nil
						datum.VerificationToken = nil
						datum.DataType = nil
						datum.EventType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/callback_url"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/verification_token"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
				),
			)
		})
	})

	Context("Subscription", func() {
		Context("ParseSubscription", func() {
			It("returns nil if the object does not exist", func() {
				Expect(oura.ParseSubscription(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("parses the datum", func() {
				datum := ouraTest.RandomSubscription(test.AllowOptional())
				object := ouraTest.NewObjectFromSubscription(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(oura.ParseSubscription(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.Subscription)) {
				datum := ouraTest.RandomSubscription(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraTest.NewObjectFromSubscription(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraTest.NewObjectFromSubscription(datum, test.ObjectFormatBSON))
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
						object["data_type"] = true
						object["event_type"] = true
						object["expiration_time"] = true
						expectedDatum.ID = nil
						expectedDatum.CallbackURL = nil
						expectedDatum.DataType = nil
						expectedDatum.EventType = nil
						expectedDatum.ExpirationTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/id"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/callback_url"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/data_type"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/event_type"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/expiration_time"),
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
				Entry("expiration_time",
					func(datum *oura.Subscription) {
						datum.ExpirationTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/expiration_time"),
				),
				Entry("expiration_time zero",
					func(datum *oura.Subscription) {
						datum.ExpirationTime = pointer.FromString(time.Time{}.Format(oura.SubscriptionExpirationTimeFormat))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/expiration_time"),
				),
				Entry("expiration_time valid",
					func(datum *oura.Subscription) {
						datum.ExpirationTime = pointer.FromString(test.RandomTime().Format(oura.SubscriptionExpirationTimeFormat))
					},
				),
				Entry("multiple errors",
					func(datum *oura.Subscription) {
						datum.ID = nil
						datum.CallbackURL = nil
						datum.DataType = nil
						datum.EventType = nil
						datum.ExpirationTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/callback_url"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/expiration_time"),
				),
			)
		})
	})

	Context("Subscriptions", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.Subscriptions)) {
				datum := pointer.From(ouraTest.RandomSubscriptions(test.AllowOptional()))
				mutator(datum)
				array := test.AsAnyArray(*datum)
				test.ExpectSerializedArrayJSON(array, ouraTest.NewArrayFromSubscriptions(datum, test.ObjectFormatJSON))
				test.ExpectSerializedArrayBSON(array, ouraTest.NewArrayFromSubscriptions(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *oura.Subscriptions) {},
			),
			Entry("empty",
				func(datum *oura.Subscriptions) {
					*datum = oura.Subscriptions{}
				},
			),
			Entry("all",
				func(datum *oura.Subscriptions) {
					*datum = ouraTest.RandomSubscriptions()
				},
			),
		)

		Context("Parse", func() {
			It("successfully parses a nil array", func() {
				parser := structureParser.NewArray(logTest.NewLogger(), nil)
				datum := oura.Subscriptions{}
				datum.Parse(parser)
				Expect(datum).To(Equal(oura.Subscriptions{}))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})

			It("successfully parses an empty array", func() {
				parser := structureParser.NewArray(logTest.NewLogger(), &[]any{})
				datum := oura.Subscriptions{}
				datum.Parse(parser)
				Expect(datum).To(Equal(oura.Subscriptions{}))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})

			It("successfully parses a non-empty array", func() {
				expectedDatum := ouraTest.RandomSubscriptions()
				array := ouraTest.NewArrayFromSubscriptions(&expectedDatum, test.ObjectFormatJSON)
				parser := structureParser.NewArray(logTest.NewLogger(), &array)
				datum := oura.Subscriptions{}
				datum.Parse(parser)
				Expect(datum).To(Equal(expectedDatum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *oura.Subscriptions), expectedErrors ...error) {
					datum := pointer.From(ouraTest.RandomSubscriptions())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.Subscriptions) {},
				),
				Entry("empty",
					func(datum *oura.Subscriptions) { *datum = oura.Subscriptions{} },
				),
				Entry("nil",
					func(datum *oura.Subscriptions) {
						*datum = oura.Subscriptions{nil}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("single invalid",
					func(datum *oura.Subscriptions) {
						invalid := ouraTest.RandomSubscription()
						invalid.ID = nil
						*datum = oura.Subscriptions{invalid}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0/id"),
				),
				Entry("single valid",
					func(datum *oura.Subscriptions) {
						*datum = oura.Subscriptions{ouraTest.RandomSubscription()}
					},
				),
				Entry("multiple invalid",
					func(datum *oura.Subscriptions) {
						invalid := ouraTest.RandomSubscription()
						invalid.ID = nil
						*datum = oura.Subscriptions{ouraTest.RandomSubscription(), invalid, ouraTest.RandomSubscription()}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/id"),
				),
				Entry("multiple valid",
					func(datum *oura.Subscriptions) {
						*datum = ouraTest.RandomSubscriptions()
					},
				),
				Entry("multiple in range (upper)",
					func(datum *oura.Subscriptions) {
						*datum = oura.Subscriptions{}
						for count := oura.SubscriptionArrayLengthMaximum; count > 0; count-- {
							*datum = append(*datum, ouraTest.RandomSubscription())
						}
					},
				),
				Entry("multiple out of range range (upper)",
					func(datum *oura.Subscriptions) {
						*datum = oura.Subscriptions{}
						for count := oura.SubscriptionArrayLengthMaximum + 1; count > 0; count-- {
							*datum = append(*datum, ouraTest.RandomSubscription())
						}
					},
					structureValidator.ErrorLengthNotLessThanOrEqualTo(oura.SubscriptionArrayLengthMaximum+1, oura.SubscriptionArrayLengthMaximum),
				),
				Entry("multiple errors",
					func(datum *oura.Subscriptions) {
						invalid := ouraTest.RandomSubscription()
						invalid.ID = nil
						*datum = oura.Subscriptions{nil, invalid, ouraTest.RandomSubscription()}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/1/id"),
				),
			)
		})

		Context("with subscriptions", func() {
			var subscriptions oura.Subscriptions

			BeforeEach(func() {
				subscriptions = ouraTest.RandomSubscriptions()
			})

			It("returns nil if there is no subscription for the event type and data type", func() {
				Expect(subscriptions.Get("invalid", "invalid")).To(BeNil())
			})

			It("returns a subscription for the event type and data type", func() {
				for _, subscription := range subscriptions {
					Expect(subscriptions.Get(*subscription.DataType, *subscription.EventType)).To(Equal(subscription))
				}
			})
		})
	})

	Context("PersonalInfo", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.PersonalInfo)) {
				datum := ouraTest.RandomPersonalInfo(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraTest.NewObjectFromPersonalInfo(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraTest.NewObjectFromPersonalInfo(datum, test.ObjectFormatBSON))
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

	Context("IsValidDataID, DataIDValidator, and ValidateDataID", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(oura.IsValidDataID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				oura.DataIDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(oura.ValidateDataID(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("is valid", ouraTest.RandomID()),
		)
	})

	Context("IsValidDataType, DataTypeValidator, and ValidateDataType", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(oura.IsValidDataType(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				oura.DataTypeValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(oura.ValidateDataType(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("is invalid", "invalid", structureValidator.ErrorValueStringNotOneOf("invalid", oura.DataTypes())),
			Entry("is DataTypeDailyActivity", oura.DataTypeDailyActivity),
			Entry("is DataTypeDailyCardiovascularAge", oura.DataTypeDailyCardiovascularAge),
			Entry("is DataTypeDailyReadiness", oura.DataTypeDailyReadiness),
			Entry("is DataTypeDailyResilience", oura.DataTypeDailyResilience),
			Entry("is DataTypeDailySleep", oura.DataTypeDailySleep),
			Entry("is DataTypeDailySpO2", oura.DataTypeDailySpO2),
			Entry("is DataTypeDailyStress", oura.DataTypeDailyStress),
			Entry("is DataTypeEnhancedTag", oura.DataTypeEnhancedTag),
			Entry("is DataTypeHeartRate", oura.DataTypeHeartRate),
			Entry("is DataTypeRestModePeriod", oura.DataTypeRestModePeriod),
			Entry("is DataTypeRingConfiguration", oura.DataTypeRingConfiguration),
			Entry("is DataTypeSession", oura.DataTypeSession),
			Entry("is DataTypeSleep", oura.DataTypeSleep),
			Entry("is DataTypeSleepTime", oura.DataTypeSleepTime),
			Entry("is DataTypeVO2Max", oura.DataTypeVO2Max),
			Entry("is DataTypeWorkout", oura.DataTypeWorkout),
		)
	})

	Context("IsValidEventType, EventTypeValidator, and ValidateEventType", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(oura.IsValidEventType(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				oura.EventTypeValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(oura.ValidateEventType(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("is invalid", "invalid", structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventTypes())),
			Entry("is EventTypeCreate", oura.EventTypeCreate),
			Entry("is EventTypeUpdate", oura.EventTypeUpdate),
			Entry("is EventTypeDelete", oura.EventTypeDelete),
		)
	})
})
