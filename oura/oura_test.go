package oura_test

import (
	"slices"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/data"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/oura"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("oura", func() {
	It("DataSetClientName is expected", func() {
		Expect(oura.DataSetClientName).To(Equal("org.tidepool.oura.api"))
	})

	It("DataSetClientVersion is expected", func() {
		Expect(oura.DataSetClientVersion).To(Equal("1.0.0"))
	})

	It("DataTypeDailyActivity is expected", func() {
		Expect(oura.DataTypeDailyActivity).To(Equal("daily_activity"))
	})

	It("DataTypeDailyCardiovascularAge is expected", func() {
		Expect(oura.DataTypeDailyCardiovascularAge).To(Equal("daily_cardiovascular_age"))
	})

	It("DataTypeDailyCyclePhases is expected", func() {
		Expect(oura.DataTypeDailyCyclePhases).To(Equal("daily_cycle_phases"))
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

	It("DataTypePersonalInfo is expected", func() {
		Expect(oura.DataTypePersonalInfo).To(Equal("personal_info"))
	})

	It("DataTypeRestModePeriod is expected", func() {
		Expect(oura.DataTypeRestModePeriod).To(Equal("rest_mode_period"))
	})

	It("DataTypeRingBatteryLevel is expected", func() {
		Expect(oura.DataTypeRingBatteryLevel).To(Equal("ring_battery_level"))
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

	It("DeviceManufacturer is expected", func() {
		Expect(oura.DeviceManufacturer).To(Equal("Oura"))
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

	It("ScopeDaily is expected", func() {
		Expect(oura.ScopeDaily).To(Equal("extapi:daily"))
	})

	It("ScopeEmail is expected", func() {
		Expect(oura.ScopeEmail).To(Equal("extapi:email"))
	})

	It("ScopeHeartHealth is expected", func() {
		Expect(oura.ScopeHeartHealth).To(Equal("extapi:heart_health"))
	})

	It("ScopeHeartRate is expected", func() {
		Expect(oura.ScopeHeartRate).To(Equal("extapi:heartrate"))
	})

	It("ScopePersonal is expected", func() {
		Expect(oura.ScopePersonal).To(Equal("extapi:personal"))
	})

	It("ScopeReproductiveCycle is expected", func() {
		Expect(oura.ScopeReproductiveCycle).To(Equal("extapi:reproductive_cycle"))
	})

	It("ScopeRingConfiguration is expected", func() {
		Expect(oura.ScopeRingConfiguration).To(Equal("extapi:ring_configuration"))
	})

	It("ScopeSession is expected", func() {
		Expect(oura.ScopeSession).To(Equal("extapi:session"))
	})

	It("ScopeSpo2 is expected", func() {
		Expect(oura.ScopeSpo2).To(Equal("extapi:spo2"))
	})

	It("ScopeStress is expected", func() {
		Expect(oura.ScopeStress).To(Equal("extapi:stress"))
	})

	It("ScopeTag is expected", func() {
		Expect(oura.ScopeTag).To(Equal("extapi:tag"))
	})

	It("ScopeWorkout is expected", func() {
		Expect(oura.ScopeWorkout).To(Equal("extapi:workout"))
	})

	It("SubscriptionArrayLengthMaximum is expected", func() {
		Expect(oura.SubscriptionArrayLengthMaximum).To(Equal(100))
	})

	It("SubscriptionExpirationTimeFormat is expected", func() {
		Expect(oura.SubscriptionExpirationTimeFormat).To(Equal("2006-01-02T15:04:05.999999999"))
	})

	It("TimeFormat is expected", func() {
		Expect(oura.TimeFormat).To(Equal(time.RFC3339))
	})

	It("TimeWithoutTimezoneFormat is expected", func() {
		Expect(oura.TimeWithoutTimezoneFormat).To(Equal("2006-01-02T15:04:05"))
	})

	It("DeviceManufacturers is expected", func() {
		Expect(oura.DeviceManufacturers).To(Equal([]string{oura.DeviceManufacturer}))
	})

	It("DeviceTags is expected", func() {
		Expect(oura.DeviceTags).To(Equal([]string{data.DeviceTagActivityMonitor}))
	})

	Context("DataTypes", func() {
		It("returns expected data types", func() {
			Expect(oura.DataTypes()).To(Equal([]string{
				oura.DataTypeDailyActivity,
				oura.DataTypeDailyCardiovascularAge,
				oura.DataTypeDailyCyclePhases,
				oura.DataTypeDailyReadiness,
				oura.DataTypeDailyResilience,
				oura.DataTypeDailySleep,
				oura.DataTypeDailySpO2,
				oura.DataTypeDailyStress,
				oura.DataTypeEnhancedTag,
				oura.DataTypeHeartRate,
				oura.DataTypePersonalInfo,
				oura.DataTypeRestModePeriod,
				oura.DataTypeRingBatteryLevel,
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

	Context("EventDataTypes", func() {
		It("returns expected data types", func() {
			Expect(oura.EventDataTypes()).To(Equal([]string{
				oura.DataTypeDailyActivity,
				oura.DataTypeDailyCardiovascularAge,
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
				oura.DataTypeVO2Max,
				oura.DataTypeWorkout,
			}))
		})
	})

	Context("Scopes", func() {
		It("returns expected data types", func() {
			Expect(oura.Scopes()).To(Equal([]string{
				oura.ScopeDaily,
				oura.ScopeEmail,
				oura.ScopeHeartHealth,
				oura.ScopeHeartRate,
				oura.ScopePersonal,
				oura.ScopeReproductiveCycle,
				oura.ScopeRingConfiguration,
				oura.ScopeSession,
				oura.ScopeSpo2,
				oura.ScopeStress,
				oura.ScopeTag,
				oura.ScopeWorkout,
			}))
		})
	})

	Context("ScopesForDataTypes", func() {
		It("returns empty scopes if data types is nil", func() {
			Expect(oura.ScopesForDataTypes(nil)).To(BeNil())
		})

		It("returns empty scopes if data types is empty", func() {
			Expect(oura.ScopesForDataTypes([]string{})).To(BeNil())
		})

		It("returns empty scopes if data types is unknown", func() {
			Expect(oura.ScopesForDataTypes([]string{"unknown"})).To(BeNil())
		})

		It("returns all scopes for all data types", func() {
			Expect(oura.ScopesForDataTypes(oura.DataTypes())).To(Equal(oura.Scopes()))
		})
	})

	Context("ScopesForDataType", func() {
		DescribeTable("returns the expected scope for the data type",
			func(dataType string, expectedScopes []string) {
				Expect(oura.ScopesForDataType(dataType)).To(Equal(expectedScopes))
			},
			Entry("DataTypeDailyActivity", oura.DataTypeDailyActivity, []string{oura.ScopeDaily}),
			Entry("DataTypeDailyCardiovascularAge", oura.DataTypeDailyCardiovascularAge, []string{oura.ScopeHeartHealth}),
			Entry("DataTypeDailyCyclePhases", oura.DataTypeDailyCyclePhases, []string{oura.ScopeReproductiveCycle}),
			Entry("DataTypeDailyReadiness", oura.DataTypeDailyReadiness, []string{oura.ScopeDaily}),
			Entry("DataTypeDailyResilience", oura.DataTypeDailyResilience, []string{oura.ScopeStress}),
			Entry("DataTypeDailySleep", oura.DataTypeDailySleep, []string{oura.ScopeDaily}),
			Entry("DataTypeDailySpO2", oura.DataTypeDailySpO2, []string{oura.ScopeSpo2}),
			Entry("DataTypeDailyStress", oura.DataTypeDailyStress, []string{oura.ScopeDaily}),
			Entry("DataTypeEnhancedTag", oura.DataTypeEnhancedTag, []string{oura.ScopeTag}),
			Entry("DataTypeHeartRate", oura.DataTypeHeartRate, []string{oura.ScopeHeartRate}),
			Entry("DataTypePersonalInfo", oura.DataTypePersonalInfo, []string{oura.ScopeEmail, oura.ScopePersonal}),
			Entry("DataTypeRestModePeriod", oura.DataTypeRestModePeriod, []string{oura.ScopeDaily}),
			Entry("DataTypeRingBatteryLevel", oura.DataTypeRingBatteryLevel, []string{oura.ScopeRingConfiguration}),
			Entry("DataTypeRingConfiguration", oura.DataTypeRingConfiguration, []string{oura.ScopeRingConfiguration}),
			Entry("DataTypeSession", oura.DataTypeSession, []string{oura.ScopeSession}),
			Entry("DataTypeSleep", oura.DataTypeSleep, []string{oura.ScopeDaily}),
			Entry("DataTypeSleepTime", oura.DataTypeSleepTime, []string{oura.ScopeDaily}),
			Entry("DataTypeVO2Max", oura.DataTypeVO2Max, []string{oura.ScopeHeartHealth}),
			Entry("DataTypeWorkout", oura.DataTypeWorkout, []string{oura.ScopeWorkout}),
		)
	})

	Context("DataTypeInScopes", func() {
		It("returns true if scopes is nil", func() {
			Expect(oura.DataTypeInScopes(oura.DataTypeDailyActivity, nil)).To(BeTrue())
		})

		It("returns true if scopes is empty", func() {
			Expect(oura.DataTypeInScopes(oura.DataTypeDailyActivity, &[]string{})).To(BeTrue())
		})

		It("returns false if scopes does not include data type", func() {
			dataType := ouraTest.RandomDataType()
			scopesForDataType := oura.ScopesForDataType(dataType)
			scopes := slices.DeleteFunc(oura.Scopes(), func(scope string) bool {
				return slices.Contains(scopesForDataType, scope)
			})
			Expect(oura.DataTypeInScopes(dataType, &scopes)).To(BeFalse())
		})

		It("returns true if scopes includes data type", func() {
			dataType := ouraTest.RandomDataType()
			scopes := oura.ScopesForDataType(dataType)
			Expect(oura.DataTypeInScopes(dataType, &scopes)).To(BeTrue())
		})

		It("returns true if scopes is all scopes", func() {
			Expect(oura.DataTypeInScopes(ouraTest.RandomDataType(), pointer.From(oura.Scopes()))).To(BeTrue())
		})
	})

	Context("DataTypeInScope", func() {
		It("returns false if scope does not include data type", func() {
			dataType := ouraTest.RandomDataType()
			scopesForDataType := oura.ScopesForDataType(dataType)
			scopes := slices.DeleteFunc(oura.Scopes(), func(scope string) bool {
				return slices.Contains(scopesForDataType, scope)
			})
			Expect(oura.DataTypeInScope(dataType, test.RandomStringFromArray(scopes))).To(BeFalse())
		})

		It("returns true if scope includes data type", func() {
			dataType := ouraTest.RandomDataType()
			scopes := oura.ScopesForDataType(dataType)
			Expect(oura.DataTypeInScope(dataType, test.RandomStringFromArray(scopes))).To(BeTrue())
		})
	})

	Context("CreateSubscription", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.CreateSubscription)) {
				datum := ouraTest.RandomCreateSubscription(test.AllowOptionals())
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
					datum.CallbackURL = pointer.From(ouraTest.RandomCallbackURL())
					datum.VerificationToken = pointer.From(ouraTest.RandomVerificationToken())
					datum.DataType = pointer.From(ouraTest.RandomEventDataType())
					datum.EventType = pointer.From(ouraTest.RandomEventType())
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *oura.CreateSubscription), expectedErrors ...error) {
					expectedDatum := ouraTest.RandomCreateSubscription(test.AllowOptionals())
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
					datum := ouraTest.RandomCreateSubscription(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.CreateSubscription) {},
				),
				Entry("callback_url missing",
					func(datum *oura.CreateSubscription) {
						datum.CallbackURL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/callback_url"),
				),
				Entry("callback_url empty",
					func(datum *oura.CreateSubscription) {
						datum.CallbackURL = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/callback_url"),
				),
				Entry("callback_url valid",
					func(datum *oura.CreateSubscription) {
						datum.CallbackURL = pointer.From(ouraTest.RandomCallbackURL())
					},
				),
				Entry("verification_token missing",
					func(datum *oura.CreateSubscription) {
						datum.VerificationToken = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/verification_token"),
				),
				Entry("verification_token empty",
					func(datum *oura.CreateSubscription) {
						datum.VerificationToken = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/verification_token"),
				),
				Entry("verification_token valid",
					func(datum *oura.CreateSubscription) {
						datum.VerificationToken = pointer.From(ouraTest.RandomVerificationToken())
					},
				),
				Entry("data_type missing",
					func(datum *oura.CreateSubscription) {
						datum.DataType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
				),
				Entry("data_type invalid",
					func(datum *oura.CreateSubscription) {
						datum.DataType = pointer.From("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventDataTypes()), "/data_type"),
				),
				Entry("data_type valid",
					func(datum *oura.CreateSubscription) {
						datum.DataType = pointer.From(ouraTest.RandomEventDataType())
					},
				),
				Entry("event_type missing",
					func(datum *oura.CreateSubscription) {
						datum.EventType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
				),
				Entry("event_type invalid",
					func(datum *oura.CreateSubscription) {
						datum.EventType = pointer.From("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventTypes()), "/event_type"),
				),
				Entry("event_type valid",
					func(datum *oura.CreateSubscription) {
						datum.EventType = pointer.From(ouraTest.RandomEventType())
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
				datum := ouraTest.RandomUpdateSubscription(test.AllowOptionals())
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
					datum.CallbackURL = pointer.From(ouraTest.RandomCallbackURL())
					datum.VerificationToken = pointer.From(ouraTest.RandomVerificationToken())
					datum.DataType = pointer.From(ouraTest.RandomEventDataType())
					datum.EventType = pointer.From(ouraTest.RandomEventType())
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *oura.UpdateSubscription), expectedErrors ...error) {
					expectedDatum := ouraTest.RandomUpdateSubscription(test.AllowOptionals())
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
					datum := ouraTest.RandomUpdateSubscription(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.UpdateSubscription) {},
				),
				Entry("callback_url missing",
					func(datum *oura.UpdateSubscription) {
						datum.CallbackURL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/callback_url"),
				),
				Entry("callback_url empty",
					func(datum *oura.UpdateSubscription) {
						datum.CallbackURL = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/callback_url"),
				),
				Entry("callback_url valid",
					func(datum *oura.UpdateSubscription) {
						datum.CallbackURL = pointer.From(ouraTest.RandomCallbackURL())
					},
				),
				Entry("verification_token missing",
					func(datum *oura.UpdateSubscription) {
						datum.VerificationToken = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/verification_token"),
				),
				Entry("verification_token empty",
					func(datum *oura.UpdateSubscription) {
						datum.VerificationToken = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/verification_token"),
				),
				Entry("verification_token valid",
					func(datum *oura.UpdateSubscription) {
						datum.VerificationToken = pointer.From(ouraTest.RandomVerificationToken())
					},
				),
				Entry("data_type missing",
					func(datum *oura.UpdateSubscription) {
						datum.DataType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
				),
				Entry("data_type invalid",
					func(datum *oura.UpdateSubscription) {
						datum.DataType = pointer.From("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventDataTypes()), "/data_type"),
				),
				Entry("data_type valid",
					func(datum *oura.UpdateSubscription) {
						datum.DataType = pointer.From(ouraTest.RandomEventDataType())
					},
				),
				Entry("event_type missing",
					func(datum *oura.UpdateSubscription) {
						datum.EventType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
				),
				Entry("event_type invalid",
					func(datum *oura.UpdateSubscription) {
						datum.EventType = pointer.From("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventTypes()), "/event_type"),
				),
				Entry("event_type valid",
					func(datum *oura.UpdateSubscription) {
						datum.EventType = pointer.From(ouraTest.RandomEventType())
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
				datum := ouraTest.RandomSubscription(test.AllowOptionals())
				object := ouraTest.NewObjectFromSubscription(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(oura.ParseSubscription(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.Subscription)) {
				datum := ouraTest.RandomSubscription(test.AllowOptionals())
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
					datum.ID = pointer.From(ouraTest.RandomID())
					datum.CallbackURL = pointer.From(ouraTest.RandomCallbackURL())
					datum.DataType = pointer.From(ouraTest.RandomEventDataType())
					datum.EventType = pointer.From(ouraTest.RandomEventType())
					datum.ExpirationTime = pointer.From(test.RandomTimeAfterNow().UTC().Format(oura.SubscriptionExpirationTimeFormat))
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *oura.Subscription), expectedErrors ...error) {
					expectedDatum := ouraTest.RandomSubscription(test.AllowOptionals())
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
					datum := ouraTest.RandomSubscription(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.Subscription) {},
				),
				Entry("id missing",
					func(datum *oura.Subscription) {
						datum.ID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
				),
				Entry("id empty",
					func(datum *oura.Subscription) {
						datum.ID = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id valid",
					func(datum *oura.Subscription) {
						datum.ID = pointer.From(ouraTest.RandomID())
					},
				),
				Entry("callback_url missing",
					func(datum *oura.Subscription) {
						datum.CallbackURL = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/callback_url"),
				),
				Entry("callback_url empty",
					func(datum *oura.Subscription) {
						datum.CallbackURL = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/callback_url"),
				),
				Entry("callback_url valid",
					func(datum *oura.Subscription) {
						datum.CallbackURL = pointer.From(ouraTest.RandomCallbackURL())
					},
				),
				Entry("data_type missing",
					func(datum *oura.Subscription) {
						datum.DataType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
				),
				Entry("data_type invalid",
					func(datum *oura.Subscription) {
						datum.DataType = pointer.From("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventDataTypes()), "/data_type"),
				),
				Entry("data_type valid",
					func(datum *oura.Subscription) {
						datum.DataType = pointer.From(ouraTest.RandomEventDataType())
					},
				),
				Entry("event_type missing",
					func(datum *oura.Subscription) {
						datum.EventType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
				),
				Entry("event_type invalid",
					func(datum *oura.Subscription) {
						datum.EventType = pointer.From("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventTypes()), "/event_type"),
				),
				Entry("event_type valid",
					func(datum *oura.Subscription) {
						datum.EventType = pointer.From(ouraTest.RandomEventType())
					},
				),
				Entry("expiration_time missing",
					func(datum *oura.Subscription) {
						datum.ExpirationTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/expiration_time"),
				),
				Entry("expiration_time zero",
					func(datum *oura.Subscription) {
						datum.ExpirationTime = pointer.From(time.Time{}.Format(oura.SubscriptionExpirationTimeFormat))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/expiration_time"),
				),
				Entry("expiration_time valid",
					func(datum *oura.Subscription) {
						datum.ExpirationTime = pointer.From(test.RandomTime().Format(oura.SubscriptionExpirationTimeFormat))
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
				datum := pointer.From(ouraTest.RandomSubscriptions(test.AllowOptionals()))
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
					*datum = oura.Subscriptions{ouraTest.RandomSubscription()}
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
				Entry("multiple out of range (upper)",
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

	Context("Event", func() {
		Context("ParseEvent", func() {
			It("returns nil if the object does not exist", func() {
				Expect(oura.ParseEvent(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("parses the datum", func() {
				datum := ouraTest.RandomEvent(test.AllowOptionals())
				object := ouraTest.NewObjectFromEvent(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(oura.ParseEvent(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.Event)) {
				datum := ouraTest.RandomEvent(test.AllowOptionals())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraTest.NewObjectFromEvent(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraTest.NewObjectFromEvent(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *oura.Event) {},
			),
			Entry("empty",
				func(datum *oura.Event) {
					*datum = oura.Event{}
				},
			),
			Entry("all",
				func(datum *oura.Event) {
					datum.EventTime = pointer.From(test.RandomTime())
					datum.EventType = pointer.From(test.RandomStringFromArray(oura.EventTypes()))
					datum.UserID = pointer.From(ouraTest.RandomUserID())
					datum.ObjectID = pointer.From(ouraTest.RandomObjectID())
					datum.DataType = pointer.From(test.RandomStringFromArray(oura.EventDataTypes()))
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *oura.Event), expectedErrors ...error) {
					expectedDatum := ouraTest.RandomEvent(test.AllowOptionals())
					object := ouraTest.NewObjectFromEvent(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &oura.Event{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *oura.Event) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *oura.Event) {
						clear(object)
						*expectedDatum = oura.Event{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *oura.Event) {
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
				func(mutator func(datum *oura.Event), expectedErrors ...error) {
					datum := ouraTest.RandomEvent(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.Event) {},
				),
				Entry("event_time",
					func(datum *oura.Event) {
						datum.EventTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_time"),
				),
				Entry("event_time zero",
					func(datum *oura.Event) {
						datum.EventTime = pointer.From(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/event_time"),
				),
				Entry("event_time valid",
					func(datum *oura.Event) {
						datum.EventTime = pointer.From(test.RandomTime())
					},
				),
				Entry("event_type",
					func(datum *oura.Event) {
						datum.EventType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event_type"),
				),
				Entry("event_type invalid",
					func(datum *oura.Event) {
						datum.EventType = pointer.From("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventTypes()), "/event_type"),
				),
				Entry("event_type valid",
					func(datum *oura.Event) {
						datum.EventType = pointer.From(test.RandomStringFromArray(oura.EventTypes()))
					},
				),
				Entry("user_id",
					func(datum *oura.Event) {
						datum.UserID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/user_id"),
				),
				Entry("user_id empty",
					func(datum *oura.Event) {
						datum.UserID = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/user_id"),
				),
				Entry("user_id valid",
					func(datum *oura.Event) {
						datum.UserID = pointer.From(test.RandomString())
					},
				),
				Entry("object_id",
					func(datum *oura.Event) {
						datum.ObjectID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/object_id"),
				),
				Entry("object_id zero",
					func(datum *oura.Event) {
						datum.ObjectID = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/object_id"),
				),
				Entry("object_id valid",
					func(datum *oura.Event) {
						datum.ObjectID = pointer.From(test.RandomString())
					},
				),
				Entry("data_type",
					func(datum *oura.Event) {
						datum.DataType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/data_type"),
				),
				Entry("data_type invalid",
					func(datum *oura.Event) {
						datum.DataType = pointer.From("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventDataTypes()), "/data_type"),
				),
				Entry("data_type valid",
					func(datum *oura.Event) {
						datum.DataType = pointer.From(test.RandomStringFromArray(oura.EventDataTypes()))
					},
				),
				Entry("multiple errors",
					func(datum *oura.Event) {
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

			DescribeTable("Hash returns expected string",
				func(eventTime *time.Time, eventType *string, userID *string, objectID *string, dataType *string, expectedHash string) {
					datum := &oura.Event{
						EventTime: eventTime,
						EventType: eventType,
						UserID:    userID,
						ObjectID:  objectID,
						DataType:  dataType,
					}
					Expect(datum.Hash()).To(Equal(expectedHash))
				},
				Entry("all", pointer.From(eventTime), pointer.From("alpha"), pointer.From("beta"), pointer.From("charlie"), pointer.From("delta"), "r6tSfxM7nHw3VapPUw9I/LtnjFBzd+5NOfCBsUd9yM0="),
				Entry("event_time missing", nil, pointer.From("alpha"), pointer.From("beta"), pointer.From("charlie"), pointer.From("delta"), "pfx/tnsxL/JWx1T/WZC6NIqTWwlIMJoLiAvRtK+B4vY="),
				Entry("event_type missing", pointer.From(eventTime), nil, pointer.From("beta"), pointer.From("charlie"), pointer.From("delta"), "sUkvKDFwLh28pjcVnt8VJY9CTjTwjoxqsRme4GX/PQE="),
				Entry("user_id missing", pointer.From(eventTime), pointer.From("alpha"), nil, pointer.From("charlie"), pointer.From("delta"), "9lke8eMWiEZQ8ayKVlEqoeSdACnjR8PMpC8iWNQF7Os="),
				Entry("object_id missing", pointer.From(eventTime), pointer.From("alpha"), pointer.From("beta"), nil, pointer.From("delta"), "QJ7dEqDastXbeO3LW1ZMVEygeU9tm+wOCLqVqWdAA44="),
				Entry("data_type missing", pointer.From(eventTime), pointer.From("alpha"), pointer.From("beta"), pointer.From("charlie"), nil, "ANvEmvnT7mSQuT+UoVdOIvSz2cEXg7geGU/lFq08vjc="),
				Entry("all missing", nil, nil, nil, nil, nil, "RBNvo1WzZ4oRRq0W9+hknpT7T8If536DEMBg9hyq/4o="),
			)
		})
	})

	It("MetadataKeyEvent is expected", func() {
		Expect(oura.MetadataKeyEvent).To(Equal("event"))
	})

	Context("ParseEventMetadata", func() {
		It("returns nil when the object is missing", func() {
			Expect(oura.ParseEventMetadata(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
		})

		It("returns new datum when the object is valid", func() {
			datum := ouraTest.RandomEventMetadata()
			object := ouraTest.NewObjectFromEventMetadata(datum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(oura.ParseEventMetadata(parser)).To(Equal(datum))
			Expect(parser.Error()).ToNot(HaveOccurred())
		})
	})

	Context("EventMetadata", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.EventMetadata)) {
				datum := ouraTest.RandomEventMetadata(test.AllowOptionals())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraTest.NewObjectFromEventMetadata(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraTest.NewObjectFromEventMetadata(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *oura.EventMetadata) {},
			),
			Entry("empty",
				func(datum *oura.EventMetadata) {
					*datum = oura.EventMetadata{}
				},
			),
			Entry("all",
				func(datum *oura.EventMetadata) {
					datum.Event = ouraTest.RandomEvent()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *oura.EventMetadata), expectedErrors ...error) {
					expectedDatum := ouraTest.RandomEventMetadata(test.AllowOptionals())
					object := ouraTest.NewObjectFromEventMetadata(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &oura.EventMetadata{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *oura.EventMetadata) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *oura.EventMetadata) {
						clear(object)
						*expectedDatum = oura.EventMetadata{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *oura.EventMetadata) {
						object["event"] = true
						expectedDatum.Event = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/event"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *oura.EventMetadata), expectedErrors ...error) {
					datum := ouraTest.RandomEventMetadata(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.EventMetadata) {},
				),
				Entry("event missing",
					func(datum *oura.EventMetadata) {
						datum.Event = nil
					},
				),
				Entry("event invalid",
					func(datum *oura.EventMetadata) {
						datum.Event.EventTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event/event_time"),
				),
				Entry("multiple errors",
					func(datum *oura.EventMetadata) {
						datum.Event.EventTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/event/event_time"),
				),
			)
		})
	})

	Context("PersonalInfo", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.PersonalInfo)) {
				datum := ouraTest.RandomPersonalInfo(test.AllowOptionals())
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
					datum.ID = pointer.From(ouraTest.RandomUserID())
					datum.Age = pointer.From(test.RandomInt())
					datum.Weight = pointer.From(test.RandomFloat64())
					datum.Height = pointer.From(test.RandomFloat64())
					datum.BiologicalSex = pointer.From(test.RandomString())
					datum.Email = pointer.From(netTest.RandomEmail())
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *oura.PersonalInfo), expectedErrors ...error) {
					expectedDatum := ouraTest.RandomPersonalInfo(test.AllowOptionals())
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
					datum := ouraTest.RandomPersonalInfo(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.PersonalInfo) {},
				),
				Entry("id missing",
					func(datum *oura.PersonalInfo) {
						datum.ID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
				),
				Entry("id empty",
					func(datum *oura.PersonalInfo) {
						datum.ID = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id valid",
					func(datum *oura.PersonalInfo) {
						datum.ID = pointer.From(ouraTest.RandomID())
					},
				),
			)
		})

		Context("with personal info", func() {
			DescribeTable("Hash returns expected string",
				func(id *string, age *int, weight *float64, height *float64, biologicalSex *string, email *string, expectedHash string) {
					datum := &oura.PersonalInfo{
						ID:            id,
						Age:           age,
						Weight:        weight,
						Height:        height,
						BiologicalSex: biologicalSex,
						Email:         email,
					}
					Expect(datum.Hash()).To(Equal(expectedHash))
				},
				Entry("all", pointer.From("da349d70cc124d57bada133ce04f0513"), pointer.From(52), pointer.From(156.4), pointer.From(72.4), pointer.From("male"), pointer.From("bob@tidepool.io"), "yzFQ5RxZLgIzTnZxiAv3laIjSSn6+ktgVjfJDa8d/lQ="),
				Entry("id missing", nil, pointer.From(41), pointer.From(152.7), pointer.From(64.0), pointer.From("female"), pointer.From("ann@tidepool.io"), "oTafvI7fAFUYapPCU6BI+gphl7634FLs6l+4tFYOvAs="),
				Entry("age missing", pointer.From("e556b8b19d034cfc9f9d5b7dae9ed3c1"), nil, pointer.From(114.5), pointer.From(68.2), pointer.From(""), pointer.From("amy@tidepool.io"), "n37K1jw/lNmWF+IwmtJZgz2NrkrsGzLlmuRtbO/bOG4="),
				Entry("weight missing", pointer.From("66249a46d97a4d409b1a7dee4be85069"), pointer.From(34), nil, pointer.From(69.2), pointer.From("other"), pointer.From("jim@tidepool.io"), "ojh/ltPFb6qyWpKgfUVKZEv32AoYuvCnLIl2bjFD5o4="),
				Entry("height missing", pointer.From("a125b96ff4fb4c45aad383d0094d34e4"), pointer.From(45), pointer.From(179.4), nil, pointer.From("non-binary"), pointer.From("jane@tidepool.io"), "QpInH6ifkBDeCnj6fXXK0RHeET824jSupJ/P7HCONBk="),
				Entry("biologicalSex missing", pointer.From("a9f95c54cda64041be1fbb2bfb8ce442"), pointer.From(26), pointer.From(213.2), pointer.From(56.4), nil, pointer.From("jon@tidepool.io"), "o/76T/S37K186nvqG7ZVheKU6CkC4IjQfnklETlpduI="),
				Entry("email missing", pointer.From("a9f95c54cda64041be1fbb2bfb8ce442"), pointer.From(26), pointer.From(213.2), pointer.From(56.4), pointer.From("unspecified"), nil, "lUCGa4kv/0h02IEyM4cRAdQvE1rOGwj8OirKO/tLTQU="),
				Entry("all missing", nil, nil, nil, nil, nil, nil, "RBNvo1WzZ4oRRq0W9+hknpT7T8If536DEMBg9hyq/4o="),
			)
		})
	})

	Context("Pagination", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.Pagination)) {
				datum := ouraTest.RandomPagination(test.AllowOptionals())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraTest.NewObjectFromPagination(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraTest.NewObjectFromPagination(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *oura.Pagination) {},
			),
			Entry("empty",
				func(datum *oura.Pagination) {
					*datum = oura.Pagination{}
				},
			),
			Entry("all",
				func(datum *oura.Pagination) {
					datum.NextToken = pointer.From(ouraTest.RandomNextToken())
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *oura.Pagination), expectedErrors ...error) {
					expectedDatum := ouraTest.RandomPagination(test.AllowOptionals())
					object := ouraTest.NewObjectFromPagination(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &oura.Pagination{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *oura.Pagination) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *oura.Pagination) {
						clear(object)
						*expectedDatum = oura.Pagination{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *oura.Pagination) {
						object["next_token"] = true
						expectedDatum.NextToken = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/next_token"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *oura.Pagination), expectedErrors ...error) {
					datum := ouraTest.RandomPagination(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.Pagination) {},
				),
				Entry("next_token missing",
					func(datum *oura.Pagination) {
						datum.NextToken = nil
					},
				),
				Entry("next_token empty",
					func(datum *oura.Pagination) {
						datum.NextToken = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/next_token"),
				),
				Entry("next_token valid",
					func(datum *oura.Pagination) {
						datum.NextToken = pointer.From(ouraTest.RandomNextToken())
					},
				),
			)
		})

		Context("HasNext", func() {
			It("returns false if the next token is nil", func() {
				datum := &oura.Pagination{}
				Expect(datum.HasNext()).To(BeFalse())
			})

			It("returns true if the next token is not nil", func() {
				datum := &oura.Pagination{NextToken: pointer.From(ouraTest.RandomNextToken())}
				Expect(datum.HasNext()).To(BeTrue())
			})
		})
	})

	Context("DataResponse", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.DataResponse)) {
				datum := ouraTest.RandomDataResponse(test.AllowOptionals())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, ouraTest.NewObjectFromDataResponse(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, ouraTest.NewObjectFromDataResponse(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *oura.DataResponse) {},
			),
			Entry("empty",
				func(datum *oura.DataResponse) {
					*datum = oura.DataResponse{}
				},
			),
			Entry("all",
				func(datum *oura.DataResponse) {
					datum.Data = ouraTest.RandomData()
					datum.Pagination = *ouraTest.RandomPagination()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *oura.DataResponse), expectedErrors ...error) {
					expectedDatum := ouraTest.RandomDataResponse(test.AllowOptionals())
					object := ouraTest.NewObjectFromDataResponse(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &oura.DataResponse{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *oura.DataResponse) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *oura.DataResponse) {
						clear(object)
						*expectedDatum = oura.DataResponse{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *oura.DataResponse) {
						object["data"] = true
						object["next_token"] = true
						expectedDatum.Data = nil
						expectedDatum.NextToken = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/data"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/next_token"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *oura.DataResponse), expectedErrors ...error) {
					datum := ouraTest.RandomDataResponse(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.DataResponse) {},
				),
				Entry("data missing",
					func(datum *oura.DataResponse) {
						datum.Data = nil
					},
				),
				Entry("data empty",
					func(datum *oura.DataResponse) {
						datum.Data = oura.Data{}
					},
				),
				Entry("data valid",
					func(datum *oura.DataResponse) {
						datum.Data = ouraTest.RandomData(test.AllowOptionals())
					},
				),
				Entry("next_token missing",
					func(datum *oura.DataResponse) {
						datum.NextToken = nil
					},
				),
				Entry("next_token empty",
					func(datum *oura.DataResponse) {
						datum.NextToken = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/next_token"),
				),
				Entry("next_token valid",
					func(datum *oura.DataResponse) {
						datum.NextToken = pointer.From(ouraTest.RandomNextToken())
					},
				),
				Entry("multiple errors",
					func(datum *oura.DataResponse) {
						datum.NextToken = pointer.From("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/next_token"),
				),
			)
		})
	})

	Context("Data", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *oura.Data)) {
				datum := ouraTest.RandomData(test.AllowOptionals())
				mutator(&datum)
				test.ExpectSerializedArrayJSON(test.AsAnyArray(datum), ouraTest.NewArrayFromData(datum, test.ObjectFormatJSON))
				test.ExpectSerializedArrayBSON(test.AsAnyArray(datum), ouraTest.NewArrayFromData(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *oura.Data) {},
			),
			Entry("empty",
				func(datum *oura.Data) {
					*datum = oura.Data{}
				},
			),
			Entry("all",
				func(datum *oura.Data) {
					*datum = oura.Data{ouraTest.RandomDatum(test.AllowOptionals())}
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(array []any, expectedDatum *oura.Data), expectedErrors ...error) {
					expectedDatum := ouraTest.RandomData(test.AllowOptionals())
					array := ouraTest.NewArrayFromData(expectedDatum, test.ObjectFormatJSON)
					mutator(array, &expectedDatum)
					result := &oura.Data{}
					errorsTest.ExpectEqual(structureParser.NewArray(logTest.NewLogger(), &array).Parse(result), expectedErrors...)
					Expect(result).To(Equal(&expectedDatum))
				},
				Entry("succeeds",
					func(array []any, expectedDatum *oura.Data) {},
				),
				Entry("empty",
					func(array []any, expectedDatum *oura.Data) {
						clear(array)
						*expectedDatum = oura.Data{}
					},
				),
				Entry("multiple errors",
					func(array []any, expectedDatum *oura.Data) {
						clear(array)
						array[0] = true
						*expectedDatum = oura.Data{}
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/0"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *oura.Data), expectedErrors ...error) {
					datum := ouraTest.RandomData(test.AllowOptionals())
					mutator(&datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(&datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *oura.Data) {},
				),
				Entry("datum missing",
					func(datum *oura.Data) {
						(*datum)[0] = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
				Entry("multiple errors",
					func(datum *oura.Data) {
						(*datum)[0] = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/0"),
				),
			)
		})

		Context("TimeMaximum", func() {
			It("returns nil when data is empty", func() {
				datum := oura.Data{}
				Expect(datum.TimeMaximum()).To(BeNil())
			})

			It("returns nil when no datum has a timestamp", func() {
				datum := oura.Data{oura.Datum{}, oura.Datum{}}
				Expect(datum.TimeMaximum()).To(BeNil())
			})

			It("returns the timestamp when there is a single datum with a timestamp", func() {
				tm := test.RandomTime().Truncate(time.Second)
				datum := oura.Data{oura.Datum{"timestamp": tm.UTC().Format(time.RFC3339)}}
				Expect(datum.TimeMaximum()).To(PointTo(BeTemporally("==", tm)))
			})

			It("returns the maximum timestamp across multiple data", func() {
				tm := test.RandomTime().Truncate(time.Second)
				tmAfter := tm.Add(time.Hour)
				tmBefore := tm.Add(-time.Hour)
				datum := oura.Data{
					oura.Datum{"timestamp": tm.UTC().Format(time.RFC3339)},
					oura.Datum{"timestamp": tmAfter.UTC().Format(time.RFC3339)},
					oura.Datum{"timestamp": tmBefore.UTC().Format(time.RFC3339)},
				}
				Expect(datum.TimeMaximum()).To(PointTo(BeTemporally("==", tmAfter)))
			})

			It("ignores data without a timestamp", func() {
				tm := test.RandomTime().Truncate(time.Second)
				datum := oura.Data{
					oura.Datum{},
					oura.Datum{"timestamp": tm.UTC().Format(time.RFC3339)},
					oura.Datum{},
				}
				Expect(datum.TimeMaximum()).To(PointTo(BeTemporally("==", tm)))
			})

			It("ignores data with an invalid timestamp", func() {
				tm := test.RandomTime().Truncate(time.Second)
				datum := oura.Data{
					oura.Datum{"timestamp": "not-a-timestamp"},
					oura.Datum{"timestamp": tm.UTC().Format(time.RFC3339)},
				}
				Expect(datum.TimeMaximum()).To(PointTo(BeTemporally("==", tm)))
			})
		})
	})

	Context("Datum", func() {
		Context("Time", func() {
			It("returns nil when timestamp key is absent", func() {
				datum := oura.Datum{}
				Expect(datum.Time()).To(BeNil())
			})

			It("returns nil when timestamp value is not a string", func() {
				datum := oura.Datum{"timestamp": 12345}
				Expect(datum.Time()).To(BeNil())
			})

			It("returns nil when timestamp string is not a valid RFC3339 time", func() {
				datum := oura.Datum{"timestamp": "not-a-timestamp"}
				Expect(datum.Time()).To(BeNil())
			})

			It("returns the parsed time when timestamp is a valid RFC3339 string", func() {
				tm := test.RandomTime().Truncate(time.Second)
				datum := oura.Datum{"timestamp": tm.UTC().Format(time.RFC3339)}
				Expect(datum.Time()).To(PointTo(BeTemporally("==", tm)))
			})

			It("returns the time in UTC when timestamp has a non-UTC offset", func() {
				datum := oura.Datum{"timestamp": "2024-03-15T10:00:00+05:00"}
				result := datum.Time()
				Expect(result).ToNot(BeNil())
				Expect(*result).To(BeTemporally("==", time.Date(2024, 3, 15, 5, 0, 0, 0, time.UTC)))
			})

			It("returns the time in UTC when timestamp has no timezone offset", func() {
				datum := oura.Datum{"timestamp": "2024-03-15T10:00:00"}
				result := datum.Time()
				Expect(result).ToNot(BeNil())
				Expect(result.Location()).To(Equal(time.UTC))
				Expect(*result).To(BeTemporally("==", time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)))
			})
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
			Entry("is DataTypeDailyCyclePhases", oura.DataTypeDailyCyclePhases),
			Entry("is DataTypeDailyReadiness", oura.DataTypeDailyReadiness),
			Entry("is DataTypeDailyResilience", oura.DataTypeDailyResilience),
			Entry("is DataTypeDailySleep", oura.DataTypeDailySleep),
			Entry("is DataTypeDailySpO2", oura.DataTypeDailySpO2),
			Entry("is DataTypeDailyStress", oura.DataTypeDailyStress),
			Entry("is DataTypeEnhancedTag", oura.DataTypeEnhancedTag),
			Entry("is DataTypeHeartRate", oura.DataTypeHeartRate),
			Entry("is DataTypePersonalInfo", oura.DataTypePersonalInfo),
			Entry("is DataTypeRestModePeriod", oura.DataTypeRestModePeriod),
			Entry("is DataTypeRingBatteryLevel", oura.DataTypeRingBatteryLevel),
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

	Context("IsValidEventDataType, EventDataTypeValidator, and ValidateEventDataType", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(oura.IsValidEventDataType(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				oura.EventDataTypeValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(oura.ValidateEventDataType(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("is invalid", "invalid", structureValidator.ErrorValueStringNotOneOf("invalid", oura.EventDataTypes())),
			Entry("is DataTypeDailyActivity", oura.DataTypeDailyActivity),
			Entry("is DataTypeDailyCardiovascularAge", oura.DataTypeDailyCardiovascularAge),
			Entry("is DataTypeDailyCyclePhases", oura.DataTypeDailyCyclePhases),
			Entry("is DataTypeDailyReadiness", oura.DataTypeDailyReadiness),
			Entry("is DataTypeDailyResilience", oura.DataTypeDailyResilience),
			Entry("is DataTypeDailySleep", oura.DataTypeDailySleep),
			Entry("is DataTypeDailySpO2", oura.DataTypeDailySpO2),
			Entry("is DataTypeDailyStress", oura.DataTypeDailyStress),
			Entry("is DataTypeEnhancedTag", oura.DataTypeEnhancedTag),
			Entry("is DataTypeRestModePeriod", oura.DataTypeRestModePeriod),
			Entry("is DataTypeRingConfiguration", oura.DataTypeRingConfiguration),
			Entry("is DataTypeSession", oura.DataTypeSession),
			Entry("is DataTypeSleep", oura.DataTypeSleep),
			Entry("is DataTypeSleepTime", oura.DataTypeSleepTime),
			Entry("is DataTypeVO2Max", oura.DataTypeVO2Max),
			Entry("is DataTypeWorkout", oura.DataTypeWorkout),
		)
	})
})
