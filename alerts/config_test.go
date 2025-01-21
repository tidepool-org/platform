package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	nontypesglucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func TestSuite(t *testing.T) {
	test.Test(t)
}

const (
	mockUserID1   = "11111111-7357-7357-7357-111111111111"
	mockUserID2   = "22222222-7357-7357-7357-222222222222"
	mockUserID3   = "33333333-7357-7357-7357-333333333333"
	mockDataSetID = "73577357735773577357735773577357"
)

var _ = Describe("Config", func() {
	It("parses all the things", func() {
		buf := buff(`{
  "userId": "%s",
  "followedUserId": "%s",
  "uploadId": "%s",
  "low": {
    "enabled": true,
    "repeat": 30,
    "delay": 10,
    "threshold": {
      "units": "mg/dL",
      "value": 80
    }
  },
  "urgentLow": {
    "enabled": false,
    "threshold": {
      "units": "mg/dL",
      "value": 47.5
    }
  },
  "high": {
    "enabled": false,
    "repeat": 30,
    "delay": 5,
    "threshold": {
      "units": "mmol/L",
      "value": 10
    }
  },
  "notLooping": {
    "enabled": true,
    "delay": 4
  },
  "noCommunication": {
    "enabled": true,
    "delay": 6
  }
}`, mockUserID1, mockUserID2, mockDataSetID)
		cfg := &Config{}
		err := request.DecodeObject(context.Background(), nil, buf, cfg)
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.UserID).To(Equal(mockUserID1))
		Expect(cfg.FollowedUserID).To(Equal(mockUserID2))
		Expect(cfg.UploadID).To(Equal(mockDataSetID))
		Expect(cfg.Alerts.High.Enabled).To(Equal(false))
		Expect(cfg.Alerts.High.Repeat).To(Equal(DurationMinutes(30 * time.Minute)))
		Expect(cfg.Alerts.High.Delay).To(Equal(DurationMinutes(5 * time.Minute)))
		Expect(cfg.Alerts.High.Threshold.Value).To(Equal(10.0))
		Expect(cfg.Alerts.High.Threshold.Units).To(Equal(nontypesglucose.MmolL))
		Expect(cfg.Alerts.Low.Enabled).To(Equal(true))
		Expect(cfg.Alerts.Low.Repeat).To(Equal(DurationMinutes(30 * time.Minute)))
		Expect(cfg.Alerts.Low.Delay).To(Equal(DurationMinutes(10 * time.Minute)))
		Expect(cfg.Alerts.Low.Threshold.Value).To(Equal(80.0))
		Expect(cfg.Alerts.Low.Threshold.Units).To(Equal(nontypesglucose.MgdL))
		Expect(cfg.Alerts.UrgentLow.Enabled).To(Equal(false))
		Expect(cfg.Alerts.UrgentLow.Threshold.Value).To(Equal(47.5))
		Expect(cfg.Alerts.UrgentLow.Threshold.Units).To(Equal(nontypesglucose.MgdL))
		Expect(cfg.Alerts.NotLooping.Enabled).To(Equal(true))
		Expect(cfg.Alerts.NotLooping.Delay).To(Equal(DurationMinutes(4 * time.Minute)))
		// Expect(conf.Alerts.NoCommunication.Enabled).To(Equal(true))
		// Expect(conf.Alerts.NoCommunication.Delay).To(Equal(DurationMinutes(6 * time.Minute)))
	})

	Context("validations", func() {
		It("requires an UploadID", func() {
			cfg := testConfig()
			cfg.UploadID = ""
			val := validator.New(logTest.NewLogger())
			cfg.Validate(val)
			Expect(val.Error()).To(MatchError(ContainSubstring("value is empty")))
		})

		It("requires an FollowedUserID", func() {
			cfg := testConfig()
			cfg.FollowedUserID = ""
			val := validator.New(logTest.NewLogger())
			cfg.Validate(val)
			Expect(val.Error()).To(MatchError(ContainSubstring("value is empty")))
		})

		It("requires an UserID", func() {
			cfg := testConfig()
			cfg.UserID = ""
			val := validator.New(logTest.NewLogger())
			cfg.Validate(val)
			Expect(val.Error()).To(MatchError(ContainSubstring("value is empty")))
		})
	})

	Context("when a notification is returned", func() {
		Describe("EvaluateNoCommunication", func() {
			It("injects user ids", func() {
				ctx, _, cfg := newConfigTest()
				cfg.Alerts.NoCommunication.Enabled = true

				when := time.Now().Add(-(DefaultNoCommunicationDelay + time.Second))
				n, _ := cfg.EvaluateNoCommunication(ctx, when)

				Expect(n).ToNot(BeNil())
				Expect(n.RecipientUserID).To(Equal(mockUserID1))
				Expect(n.FollowedUserID).To(Equal(mockUserID2))
			})
		})
	})

	Describe("EvaluateData", func() {
		var okGlucose = []*Glucose{testInRangeDatum()}
		var okDosing = []*DosingDecision{testDosingDecision(time.Second)}

		type evalTest struct {
			Name     string
			Activity func(*Config) *AlertActivity
			Glucose  []*Glucose
			Dosing   []*DosingDecision
		}

		tests := []evalTest{
			{"UrgentLow", func(c *Config) *AlertActivity { return &c.Activity.UrgentLow },
				[]*Glucose{testUrgentLowDatum()}, nil},
			{"Low", func(c *Config) *AlertActivity { return &c.Activity.Low },
				[]*Glucose{testLowDatum()}, nil},
			{"High", func(c *Config) *AlertActivity { return &c.Activity.High },
				[]*Glucose{testHighDatum()}, nil},
			{"NotLooping", func(c *Config) *AlertActivity { return &c.Activity.NotLooping },
				nil, []*DosingDecision{testDosingDecision(-30 * time.Hour)}},
		}
		for _, test := range tests {
			Context(test.Name, func() {
				It("is triggered", func() {
					ctx, _, cfg := newConfigTest()
					cfg.Alerts.NotLooping = testNotLooping()
					cfg.EvaluateData(ctx, okGlucose, okDosing)
					n, _ := cfg.EvaluateData(ctx, test.Glucose, test.Dosing)
					Expect(n).ToNot(BeNil())
					Expect(test.Activity(cfg).Triggered).ToNot(BeZero())
				})

				It("doesn't update its triggered time", func() {
					ctx, _, cfg := newConfigTest()
					cfg.Alerts.NotLooping = testNotLooping()
					cfg.EvaluateData(ctx, okGlucose, okDosing)
					n, _ := cfg.EvaluateData(ctx, test.Glucose, test.Dosing)
					Expect(n).ToNot(BeNil())
					Expect(test.Activity(cfg).Triggered).ToNot(BeZero())
					prev := test.Activity(cfg).Triggered
					n, _ = cfg.EvaluateData(ctx, test.Glucose, test.Dosing)
					Expect(n).ToNot(BeNil())
					Expect(test.Activity(cfg).Triggered).To(Equal(prev))
				})

				It("is resolved", func() {
					ctx, _, cfg := newConfigTest()
					cfg.Alerts.NotLooping = testNotLooping()
					n, _ := cfg.EvaluateData(ctx, test.Glucose, test.Dosing)
					Expect(n).ToNot(BeNil())
					Expect(test.Activity(cfg).Resolved).To(BeZero())
					n, _ = cfg.EvaluateData(ctx, okGlucose, okDosing)
					Expect(n).To(BeNil())
					Expect(test.Activity(cfg).Resolved).To(BeTemporally("~", time.Now()))
				})

				It("doesn't update its resolved time", func() {
					ctx, _, cfg := newConfigTest()
					cfg.Alerts.NotLooping = testNotLooping()
					n, _ := cfg.EvaluateData(ctx, test.Glucose, test.Dosing)
					Expect(n).ToNot(BeNil())
					n, _ = cfg.EvaluateData(ctx, okGlucose, okDosing)
					Expect(n).To(BeNil())
					prev := test.Activity(cfg).Resolved
					n, _ = cfg.EvaluateData(ctx, okGlucose, okDosing)
					Expect(n).To(BeNil())
					Expect(test.Activity(cfg).Resolved).To(Equal(prev))
				})
			})
		}

		type logTest struct {
			Name   string
			Msg    string
			Fields log.Fields
		}

		logTests := []logTest{
			{"UrgentLow", "urgent low", log.Fields{
				"isAlerting?": false, "value": 6.0, "threshold": 3.0}},
			{"Low", "low", log.Fields{
				"isAlerting?": false, "value": 6.0, "threshold": 4.0}},
			{"High", "high", log.Fields{
				"isAlerting?": false, "value": 6.0, "threshold": 10.0}},
			{"NotLooping", "not looping", log.Fields{
				"isAlerting?": false,
				// "value" is time-dependent, and would require a lot of work to mock. This
				// should be close enough.
				"threshold": DefaultNotLoopingDelay,
			}},
		}
		for _, test := range logTests {
			It(test.Name+" logs evaluations", func() {
				ctx, lgr, cfg := newConfigTest()
				cfg.Alerts.NotLooping.Base.Enabled = true
				glucose := []*Glucose{testInRangeDatum()}
				dosing := []*DosingDecision{testDosingDecision(-1)}
				cfg.EvaluateData(ctx, glucose, dosing)

				Expect(func() {
					lgr.AssertLog(log.InfoLevel, test.Msg, test.Fields)
				}).ToNot(Panic(), quickJSON(map[string]any{
					"got":      lgr.SerializedFields,
					"expected": map[string]any{"message": test.Msg, "fields": test.Fields},
				}))
			})
		}

		It("injects user IDs into the returned Notification", func() {
			ctx, _, cfg := newConfigTest()
			mockGlucoseData := []*Glucose{testUrgentLowDatum()}

			n, _ := cfg.EvaluateData(ctx, mockGlucoseData, nil)

			Expect(n).ToNot(BeNil())
			Expect(n.RecipientUserID).To(Equal(mockUserID1))
			Expect(n.FollowedUserID).To(Equal(mockUserID2))
		})

		It("ripples the needs upsert value (from urgent low)", func() {
			ctx, _, cfg := newConfigTest()

			// Generate an urgent low notification.
			n, _ := cfg.EvaluateData(ctx, []*Glucose{testUrgentLowDatum()}, nil)
			Expect(n).ToNot(Equal(nil))
			// Now resolve the alert, resulting in changed being true, but without a
			// notification.
			n, needsUpsert := cfg.EvaluateData(ctx, []*Glucose{testInRangeDatum()}, nil)
			Expect(n).To(BeNil())
			Expect(needsUpsert).To(Equal(true))
		})

		It("ripples the needs upsert value (from low)", func() {
			ctx, _, cfg := newConfigTest()

			// Generate a low notification.
			n, needsUpsert := cfg.EvaluateData(ctx, []*Glucose{testLowDatum()}, nil)
			Expect(n).ToNot(BeNil())
			Expect(needsUpsert).To(Equal(true))
			// Now resolve the alert, resulting in changed being true, but without a
			// notification.
			n, needsUpsert = cfg.EvaluateData(ctx, []*Glucose{testInRangeDatum()}, nil)
			Expect(n).To(BeNil())
			Expect(needsUpsert).To(Equal(true))
		})

		It("ripples the needs upsert value (form high)", func() {
			ctx, _, cfg := newConfigTest()

			// Generate a high notification.
			n, needsUpsert := cfg.EvaluateData(ctx, []*Glucose{testHighDatum()}, nil)
			Expect(n).ToNot(BeNil())
			Expect(needsUpsert).To(Equal(true))
			// Now resolve the alert, resulting in changed being true, but without a
			// notification.
			n, needsUpsert = cfg.EvaluateData(ctx, []*Glucose{testInRangeDatum()}, nil)
			Expect(n).To(BeNil())
			Expect(needsUpsert).To(Equal(true))
		})

		Describe("Repeat", func() {
			It("Low is respected", func() {
				ctx, _, cfg := newConfigTest()
				cfg.Alerts.Low.Repeat = DurationMinutes(10 * time.Minute)
				cfg.Alerts.Low.Delay = DurationMinutes(1 * time.Nanosecond)
				cfg.Activity.Low.Triggered = time.Now().Add(-time.Hour)
				cfg.Activity.Low.Sent = time.Now().Add((-10 * time.Minute) + time.Second)
				testData := []*Glucose{testLowDatum()}

				n, _ := cfg.EvaluateData(ctx, testData, nil)
				Expect(n).To(BeNil())

				cfg.Activity.Low.Sent = time.Now().Add((-10 * time.Minute) - time.Second)

				n, _ = cfg.EvaluateData(ctx, testData, nil)
				Expect(n).ToNot(BeNil())
			})

			It("High is respected", func() {
				ctx, _, cfg := newConfigTest()
				cfg.Alerts.High.Repeat = DurationMinutes(10 * time.Minute)
				cfg.Alerts.High.Delay = DurationMinutes(1 * time.Nanosecond)
				cfg.Activity.High.Triggered = time.Now().Add(-time.Hour)
				cfg.Activity.High.Sent = time.Now().Add((-10 * time.Minute) + time.Second)
				delayed := []*Glucose{testHighDatum()}

				n, _ := cfg.EvaluateData(ctx, delayed, nil)
				Expect(n).To(BeNil())

				cfg.Activity.High.Sent = time.Now().Add((-10 * time.Minute) - time.Second)

				n, _ = cfg.EvaluateData(ctx, delayed, nil)
				Expect(n).ToNot(BeNil())
			})
		})

		Describe("Delay", func() {
			It("Low is respected", func() {
				ctx, _, cfg := newConfigTest()
				cfg.Alerts.Low.Delay = DurationMinutes(5 * time.Minute)
				cfg.Alerts.Low.Repeat = DurationMinutes(1 * time.Nanosecond)
				delayed := []*Glucose{testLowDatum()}

				n, _ := cfg.EvaluateData(ctx, delayed, nil)
				Expect(n).To(BeNil())

				delayed[0].Time = pointer.FromAny(time.Now().Add(-5 * time.Minute))

				n, _ = cfg.EvaluateData(ctx, delayed, nil)
				Expect(n).ToNot(BeNil())
			})

			It("High is respected", func() {
				ctx, _, cfg := newConfigTest()
				cfg.Alerts.High.Delay = DurationMinutes(5 * time.Minute)
				cfg.Alerts.High.Repeat = DurationMinutes(1 * time.Nanosecond)
				delayed := []*Glucose{testHighDatum()}

				n, _ := cfg.EvaluateData(ctx, delayed, nil)
				Expect(n).To(BeNil())

				delayed[0].Time = pointer.FromAny(time.Now().Add(-5 * time.Minute))

				n, _ = cfg.EvaluateData(ctx, delayed, nil)
				Expect(n).ToNot(BeNil())
			})

			It("NotLooping is respected", func() {
				ctx, _, cfg := newConfigTest()
				cfg.Alerts.NotLooping.Enabled = true
				delay := 10 * time.Minute
				lessThanDelay := delay - time.Second
				cfg.Alerts.NotLooping.Delay = DurationMinutes(delay)
				delayed := []*DosingDecision{testDosingDecision(-lessThanDelay)}

				n, _ := cfg.EvaluateData(ctx, nil, delayed)
				Expect(n).To(BeNil())

				moreThanDelay := delay + time.Second
				delayed[0].Time = pointer.FromAny(time.Now().Add(-moreThanDelay))

				n, _ = cfg.EvaluateData(ctx, nil, delayed)
				Expect(n).ToNot(BeNil())
			})

			It("NotLooping uses its default", func() {
				ctx, _, cfg := newConfigTest()
				cfg.Alerts.NotLooping.Enabled = true
				cfg.Alerts.NotLooping.Delay = 0
				lessThanDelay := DefaultNotLoopingDelay - time.Second
				delayed := []*DosingDecision{testDosingDecision(-lessThanDelay)}

				n, _ := cfg.EvaluateData(ctx, nil, delayed)
				Expect(n).To(BeNil())

				moreThanDelay := DefaultNotLoopingDelay + time.Second
				delayed[0].Time = pointer.FromAny(time.Now().Add(-moreThanDelay))

				n, _ = cfg.EvaluateData(ctx, nil, delayed)
				Expect(n).ToNot(BeNil())
			})
		})
	})

	It("observes NotLoopingRepeat between notifications", func() {
		ctx, _, cfg := newConfigTest()
		cfg.Alerts.NotLooping = testNotLooping()
		yesterday := []*DosingDecision{testDosingDecision(-24 * time.Hour)}

		cfg.Activity.NotLooping.Sent = time.Now()
		n, _ := cfg.EvaluateData(ctx, nil, yesterday)
		Expect(n).To(BeNil())

		cfg.Activity.NotLooping.Sent = time.Now().Add(-(1 + NotLoopingRepeat))
		n, _ = cfg.EvaluateData(ctx, nil, yesterday)
		Expect(n).ToNot(BeNil())
	})

	Context("UrgentLowAlert", func() {
		Context("Threshold", func() {
			It("accepts values between 0 and 1000 mg/dL", func() {
				val := validator.New(logTest.NewLogger())
				b := UrgentLowAlert{Threshold: Threshold{Value: 0, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = UrgentLowAlert{Threshold: Threshold{Value: 1000, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = UrgentLowAlert{Threshold: Threshold{Value: 1001, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 1001 is not between 0 and 1000"))

				val = validator.New(logTest.NewLogger())
				b = UrgentLowAlert{Threshold: Threshold{Value: -1, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1 is not between 0 and 1000"))
			})
		})

		Context("Evaluate", func() {
			It("handles being passed empty data", func() {
				ctx, _, cfg := newConfigTest()
				ul := cfg.Alerts.UrgentLow

				er := EvalResult{}
				Expect(func() {
					er = ul.Evaluate(ctx, []*Glucose{})
				}).ToNot(Panic())
				Expect(func() {
					er = ul.Evaluate(ctx, nil)
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(false))
			})

			It("validates glucose data", func() {
				ctx, _, cfg := newConfigTest()
				ul := cfg.Alerts.UrgentLow

				er := EvalResult{}
				Expect(func() {
					er = ul.Evaluate(ctx, []*Glucose{testUrgentLowDatum()})
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(true))

				badUnits := testInRangeDatum()
				badUnits.Units = nil
				Expect(func() {
					er = ul.Evaluate(ctx, []*Glucose{badUnits})
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(false))

				badValue := testInRangeDatum()
				badValue.Value = nil
				Expect(func() {
					er = ul.Evaluate(ctx, []*Glucose{badValue})
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(false))

				// TODO is this still useful?
				//
				// badTime := testGlucoseDatum(1)
				// badTime.Time = nil
				// Expect(func() {
				// 	notification, _ = testUrgentLow().Evaluate(ctx, []*Glucose{badTime})
				// }).ToNot(Panic())
				// Expect(notification).To(BeNil())

			})
		})
	})

	Context("NoCommunicationAlert", func() {
		Context("Evaluate", func() {

			It("handles being passed a Zero time.Time value", func() {
				ctx, _, cfg := newConfigTest()
				nc := cfg.Alerts.NoCommunication

				Expect(func() {
					nc.Evaluate(ctx, time.Time{})
				}).ToNot(Panic())
			})

			It("logs evaluation results", func() {
				ctx, lgr, cfg := newConfigTest()
				nc := cfg.Alerts.NoCommunication

				Expect(func() {
					nc.Evaluate(ctx, time.Now().Add(-12*time.Hour))
				}).ToNot(Panic())
				Expect(func() {
					lgr.AssertLog(log.InfoLevel, "no communication", log.Fields{
						"isAlerting?": true,
					})
				}).ToNot(Panic())
			})

			It("honors non-Zero Delay values", func() {
				ctx, _, cfg := newConfigTest()
				nc := cfg.Alerts.NoCommunication
				nc.Enabled = true
				nc.Delay = DurationMinutes(10 * time.Minute)

				wontTrigger := time.Now().Add(-(nc.Delay.Duration() - time.Second))
				er := nc.Evaluate(ctx, wontTrigger)
				Expect(er.OutOfRange).To(Equal(false))

				willTrigger := time.Now().Add(-(nc.Delay.Duration() + time.Second))
				er = nc.Evaluate(ctx, willTrigger)
				Expect(er.OutOfRange).To(Equal(true))
			})

			It("validates the time at which data was last received", func() {
				ctx, _, cfg := newConfigTest()
				validLastReceived := time.Now().Add(-10*time.Minute + -DefaultNoCommunicationDelay)
				invalidLastReceived := time.Time{}
				er := EvalResult{}
				nc := cfg.Alerts.NoCommunication

				er = nc.Evaluate(ctx, validLastReceived)
				Expect(er.OutOfRange).To(Equal(true))

				er = nc.Evaluate(ctx, invalidLastReceived)
				Expect(er.OutOfRange).To(Equal(false))
			})
		})
	})

	Context("LowAlert", func() {
		Context("Threshold", func() {
			It("accepts values in mmol/L", func() {
				val := validator.New(logTest.NewLogger())
				b := LowAlert{Threshold: Threshold{Value: 4.2735, Units: "mmol/L"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())
			})

			It("accepts values between 0 and 1000 mg/dL", func() {
				val := validator.New(logTest.NewLogger())
				b := LowAlert{Threshold: Threshold{Value: 0, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = LowAlert{Threshold: Threshold{Value: 1000, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = LowAlert{Threshold: Threshold{Value: 1001, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 1001 is not between 0 and 1000"))

				val = validator.New(logTest.NewLogger())
				b = LowAlert{Threshold: Threshold{Value: -1, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1 is not between 0 and 1000"))
			})
		})

		Context("Delay", func() {
			It("accepts values between 0 and 6 hours (inclusive)", func() {
				okThresh := Threshold{Units: "mg/dL", Value: 123}

				val := validator.New(logTest.NewLogger())
				b := HighAlert{Delay: 0, Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = HighAlert{Delay: DurationMinutes(time.Hour * 6 / time.Minute), Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = HighAlert{Delay: -1, Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1ns is not between 0s and 6h0m0s"))

				val = validator.New(logTest.NewLogger())
				b = HighAlert{Delay: DurationMinutes(time.Hour*6 + time.Minute), Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 6h1m0s is not between 0s and 6h0m0s"))
			})
		})

		Context("Evaluate", func() {
			It("handles being passed empty data", func() {
				ctx, _, cfg := newConfigTest()
				er := EvalResult{}
				low := cfg.Alerts.Low

				Expect(func() {
					er = low.Evaluate(ctx, []*Glucose{})
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(false))
				Expect(func() {
					er = low.Evaluate(ctx, nil)
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(false))
			})

			It("validates glucose data", func() {
				ctx, _, cfg := newConfigTest()
				er := EvalResult{}
				low := cfg.Alerts.Low

				Expect(func() {
					er = low.Evaluate(ctx, []*Glucose{testUrgentLowDatum()})
				}).ToNot(Panic())
				Expect(er.OutOfRange).ToNot(Equal(false))

				badUnits := testUrgentLowDatum()
				badUnits.Units = nil
				Expect(func() {
					er = low.Evaluate(ctx, []*Glucose{badUnits})
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(false))

				badValue := testUrgentLowDatum()
				badValue.Value = nil
				Expect(func() {
					er = low.Evaluate(ctx, []*Glucose{badValue})
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(false))

				// TODO is this useful?
				//
				// badTime := testGlucoseDatum(1)
				// badTime.Time = nil
				// Expect(func() {
				// 	notification, _ = low.Evaluate(ctx, []*Glucose{badTime})
				// }).ToNot(Panic())
				// Expect(notification).To(BeNil())
			})
		})
	})

	Context("HighAlert", func() {
		Context("Threshold", func() {
			It("accepts values between 0 and 1000 mg/dL", func() {
				val := validator.New(logTest.NewLogger())
				b := HighAlert{Threshold: Threshold{Value: 0, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = HighAlert{Threshold: Threshold{Value: 1000, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = HighAlert{Threshold: Threshold{Value: 1001, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 1001 is not between 0 and 1000"))

				val = validator.New(logTest.NewLogger())
				b = HighAlert{Threshold: Threshold{Value: -1, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1 is not between 0 and 1000"))
			})
		})

		Context("Delay", func() {
			It("accepts values between 0 and 6 hours (inclusive)", func() {
				okThresh := Threshold{Units: "mg/dL", Value: 123}

				val := validator.New(logTest.NewLogger())
				b := HighAlert{Delay: 0, Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = HighAlert{Delay: DurationMinutes(time.Hour * 6 / time.Minute), Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = HighAlert{Delay: -1, Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1ns is not between 0s and 6h0m0s"))

				val = validator.New(logTest.NewLogger())
				b = HighAlert{Delay: DurationMinutes(time.Hour*6 + time.Minute), Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 6h1m0s is not between 0s and 6h0m0s"))
			})
		})

		Context("Evaluate", func() {

			It("handles being passed empty data", func() {
				ctx, _, cfg := newConfigTest()
				er := EvalResult{}
				high := cfg.Alerts.High

				Expect(func() {
					er = high.Evaluate(ctx, []*Glucose{})
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(false))
				Expect(func() {
					er = high.Evaluate(ctx, nil)
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(false))
			})

			It("validates glucose data", func() {
				ctx, _, cfg := newConfigTest()
				er := EvalResult{}
				high := cfg.Alerts.High

				Expect(func() {
					er = high.Evaluate(ctx, []*Glucose{testHighDatum()})
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(true))

				badUnits := testInRangeDatum()
				badUnits.Units = nil
				Expect(func() {
					er = high.Evaluate(ctx, []*Glucose{badUnits})
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(false))

				badValue := testInRangeDatum()
				badValue.Value = nil
				Expect(func() {
					er = high.Evaluate(ctx, []*Glucose{badValue})
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(false))

				// TODO is this still useful?
				badTime := testInRangeDatum()
				badTime.Time = nil
				Expect(func() {
					er = high.Evaluate(ctx, []*Glucose{badTime})
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(false))
			})
		})
	})

	Context("NoCommunicationAlert", func() {
		Context("Delay", func() {
			It("accepts values between 0 and 6 hours (inclusive)", func() {
				val := validator.New(logTest.NewLogger())
				b := NoCommunicationAlert{Delay: 0}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = NoCommunicationAlert{Delay: DurationMinutes(time.Hour * 6)}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = NoCommunicationAlert{Delay: -1}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1ns is not between 0s and 6h0m0s"))

				val = validator.New(logTest.NewLogger())
				b = NoCommunicationAlert{Delay: DurationMinutes(time.Hour*6 + time.Second)}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 6h0m1s is not between 0s and 6h0m0s"))
			})
		})
	})

	Context("NotLoopingAlert", func() {

		Context("Delay", func() {
			It("accepts values between 0 and 2 hours (inclusive)", func() {
				val := validator.New(logTest.NewLogger())
				b := NotLoopingAlert{Delay: 0}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = NotLoopingAlert{Delay: DurationMinutes(2 * time.Hour)}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New(logTest.NewLogger())
				b = NotLoopingAlert{Delay: -1}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1ns is not between 0s and 2h0m0s"))

				val = validator.New(logTest.NewLogger())
				b = NotLoopingAlert{Delay: DurationMinutes(2*time.Hour + time.Second)}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 2h0m1s is not between 0s and 2h0m0s"))
			})
		})

		Context("Evaluate", func() {

			It("uses a default delay of 30 minutes", func() {
				ctx, _, cfg := newConfigTest()
				decisionsNoAlert := []*DosingDecision{
					testDosingDecision(-29 * time.Minute),
				}
				decisionsWithAlert := []*DosingDecision{
					testDosingDecision(-30 * time.Minute),
				}
				nl := cfg.Alerts.NotLooping

				er := nl.Evaluate(ctx, decisionsNoAlert)
				Expect(er.OutOfRange).To(Equal(false), er.String())
				er = nl.Evaluate(ctx, decisionsWithAlert)
				Expect(er.OutOfRange).To(Equal(true))
			})

			It("respects custom delays", func() {
				ctx, _, cfg := newConfigTest()
				decisionsNoAlert := []*DosingDecision{
					testDosingDecision(-14 * time.Minute),
				}
				decisionsWithAlert := []*DosingDecision{
					testDosingDecision(-15 * time.Minute),
				}
				nl := cfg.Alerts.NotLooping
				nl.Delay = DurationMinutes(15 * time.Minute)

				er := nl.Evaluate(ctx, decisionsNoAlert)
				Expect(er.OutOfRange).To(Equal(false))
				er = nl.Evaluate(ctx, decisionsWithAlert)
				Expect(er.OutOfRange).To(Equal(true))
			})

			It("handles being passed empty data", func() {
				ctx, _, cfg := newConfigTest()
				er := EvalResult{}

				nl := cfg.Alerts.NotLooping

				Expect(func() {
					er = nl.Evaluate(ctx, []*DosingDecision{})
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(true))
				Expect(func() {
					er = nl.Evaluate(ctx, nil)
				}).ToNot(Panic())
				Expect(er.OutOfRange).To(Equal(true))
			})

			It("ignores decisions without a reason", func() {
				ctx, _, cfg := newConfigTest()
				nl := cfg.Alerts.NotLooping
				noReason := testDosingDecision(time.Second)
				noReason.Reason = nil
				decisions := []*DosingDecision{
					testDosingDecision(-time.Hour),
					noReason,
				}

				er := nl.Evaluate(ctx, decisions)
				Expect(er.OutOfRange).To(Equal(true))
			})

			It("ignores decisions without a time", func() {
				ctx, _, cfg := newConfigTest()

				nl := cfg.Alerts.NotLooping
				noTime := testDosingDecision(time.Second)
				noTime.Time = nil
				decisions := []*DosingDecision{
					testDosingDecision(-time.Hour),
					noTime,
				}

				er := nl.Evaluate(ctx, decisions)
				Expect(er.OutOfRange).To(Equal(true))
			})
		})
	})

	Context("repeat", func() {
		var defaultAlert = LowAlert{
			Threshold: Threshold{Value: 11, Units: nontypesglucose.MmolL},
		}

		It("accepts values of 0 (indicating disabled)", func() {
			val := validator.New(logTest.NewLogger())
			l := defaultAlert
			l.Repeat = 0
			l.Validate(val)
			Expect(val.Error()).To(Succeed())
		})

		It("accepts values of 15 minutes to 4 hours (inclusive)", func() {
			val := validator.New(logTest.NewLogger())
			l := defaultAlert
			l.Repeat = DurationMinutes(15 * time.Minute)
			l.Validate(val)
			Expect(val.Error()).To(Succeed())

			val = validator.New(logTest.NewLogger())
			l = defaultAlert
			l.Repeat = DurationMinutes(4 * time.Hour)
			l.Validate(val)
			Expect(val.Error()).To(Succeed())

			val = validator.New(logTest.NewLogger())
			l = defaultAlert
			l.Repeat = DurationMinutes(4*time.Hour + 1)
			l.Validate(val)
			Expect(val.Error()).NotTo(Succeed())

			val = validator.New(logTest.NewLogger())
			l = defaultAlert
			l.Repeat = DurationMinutes(15*time.Minute - 1)
			l.Validate(val)
			Expect(val.Error()).NotTo(Succeed())
		})
	})

	Context("urgentLow", func() {
		It("validates threshold units", func() {
			buf := buff(`{"urgentLow": {"threshold": {"units":"%s","value":42}}`, "garbage")
			threshold := &Threshold{}
			err := request.DecodeObject(context.Background(), nil, buf, threshold)
			Expect(err).To(MatchError("json is malformed"))
		})
	})

	Context("low", func() {
		It("accepts a blank repeat", func() {
			buf := buff(`{
  "userId": "%s",
  "followedUserId": "%s",
  "uploadId": "%s",
  "low": {
    "enabled": true,
    "delay": 10,
    "threshold": {
      "units": "mg/dL",
      "value": 80
    }
  }
}`, mockUserID1, mockUserID2, mockDataSetID)
			conf := &Config{}
			err := request.DecodeObject(context.Background(), nil, buf, conf)
			Expect(err).To(Succeed())
			Expect(conf.Alerts.Low.Repeat).To(Equal(DurationMinutes(0)))
		})
	})
	It("validates repeat minutes (negative)", func() {
		buf := buff(`{
  "userId": "%s",
  "followedUserId": "%s",
  "uploadId": "%s",
  "low": {
    "enabled": false,
    "repeat": -11,
    "threshold": {
      "units": "%s",
      "value": 47.5
    }
  }
}`, mockUserID1, mockUserID2, mockDataSetID, nontypesglucose.MgdL)
		cfg := &Config{}
		err := request.DecodeObject(context.Background(), nil, buf, cfg)
		Expect(err).To(MatchError("value -11m0s is not greater than or equal to 15m0s"))
	})
	It("validates repeat minutes (string)", func() {
		buf := buff(`{
  "userId": "%s",
  "followedUserId": "%s",
  "uploadId": "%s",
  "low": {
    "enabled": false,
    "repeat": "a",
    "threshold": {
      "units": "%s",
      "value": 1
    }
  }
}`, mockUserID1, mockUserID2, mockDataSetID, nontypesglucose.MgdL)
		cfg := &Config{}
		err := request.DecodeObject(context.Background(), nil, buf, cfg)
		Expect(err).To(MatchError("json is malformed"))
	})
})

var _ = Describe("Alerts", func() {
	Describe("LongestDelay", func() {
		It("does what it says", func() {
			low := testLowAlert()
			low.Delay = DurationMinutes(10 * time.Minute)
			high := testHighAlert()
			high.Delay = DurationMinutes(5 * time.Minute)
			notLooping := testNotLoopingAlert()
			notLooping.Delay = DurationMinutes(5 * time.Minute)

			a := Alerts{
				Low:        low,
				High:       high,
				NotLooping: notLooping,
			}

			delay := a.LongestDelay()

			Expect(delay).To(Equal(10 * time.Minute))
		})

		It("ignores disabled alerts", func() {
			low := testLowAlert()
			low.Delay = DurationMinutes(7 * time.Minute)
			high := testHighAlert()
			high.Delay = DurationMinutes(5 * time.Minute)
			notLooping := testNotLoopingAlert()
			notLooping.Delay = DurationMinutes(5 * time.Minute)

			a := Alerts{
				Low:        low,
				High:       high,
				NotLooping: notLooping,
			}

			delay := a.LongestDelay()

			Expect(delay).To(Equal(7 * time.Minute))
		})

		It("returns a Zero Duration when no alerts are set", func() {
			a := Alerts{
				Low:        nil,
				High:       nil,
				NotLooping: nil,
			}

			delay := a.LongestDelay()

			Expect(delay).To(Equal(time.Duration(0)))
		})
	})

	Describe("Evaluate", func() {

		It("detects urgent low data", func() {
			ctx, _, cfg := newConfigTest()
			data := []*Glucose{testUrgentLowDatum()}
			n, _ := cfg.EvaluateData(ctx, data, nil)

			Expect(n).ToNot(BeNil())
			Expect(n.Message).To(ContainSubstring("below urgent low threshold"))
		})

		It("detects low data", func() {
			ctx, _, cfg := newConfigTest()
			data := []*Glucose{testLowDatum()}
			n, _ := cfg.EvaluateData(ctx, data, nil)

			Expect(n).ToNot(BeNil())
			Expect(n.Message).To(ContainSubstring("below low threshold"))
		})

		It("detects high data", func() {
			ctx, _, cfg := newConfigTest()
			data := []*Glucose{testHighDatum()}
			n, _ := cfg.EvaluateData(ctx, data, nil)

			Expect(n).ToNot(BeNil())
			Expect(n.Message).To(ContainSubstring("above high threshold"))
		})

		Context("with both low and urgent low alerts detected", func() {
			It("prefers urgent low", func() {
				ctx, _, cfg := newConfigTest()
				data := []*Glucose{testUrgentLowDatum()}
				n, _ := cfg.EvaluateData(ctx, data, nil)

				Expect(n).ToNot(BeNil())
				Expect(n.Message).To(ContainSubstring("below urgent low threshold"))
			})
		})
	})
})

var _ = Describe("DurationMinutes", func() {
	It("parses 42", func() {
		d := DurationMinutes(0)
		err := d.UnmarshalJSON([]byte(`42`))
		Expect(err).To(BeNil())
		Expect(d.Duration()).To(Equal(42 * time.Minute))
	})
	It("parses 0", func() {
		d := DurationMinutes(time.Minute)
		err := d.UnmarshalJSON([]byte(`0`))
		Expect(err).To(BeNil())
		Expect(d.Duration()).To(Equal(time.Duration(0)))
	})
	It("parses null as 0 minutes", func() {
		d := DurationMinutes(time.Minute)
		err := d.UnmarshalJSON([]byte(`null`))
		Expect(err).To(BeNil())
		Expect(d.Duration()).To(Equal(time.Duration(0)))
	})
	It("parses an empty value as 0 minutes", func() {
		d := DurationMinutes(time.Minute)
		err := d.UnmarshalJSON([]byte(``))
		Expect(err).To(BeNil())
		Expect(d.Duration()).To(Equal(time.Duration(0)))
	})
	It("marshals to 5", func() {
		d := DurationMinutes(5 * time.Minute)
		out, err := d.MarshalJSON()
		Expect(err).To(Succeed())
		Expect(out).To(Equal([]byte("5")))
	})
})

var _ = Describe("Threshold", func() {
	It("accepts mg/dL", func() {
		buf := buff(`{"units":"%s","value":42}`, nontypesglucose.MgdL)
		threshold := &Threshold{}
		err := request.DecodeObject(context.Background(), nil, buf, threshold)
		Expect(err).To(BeNil())
		Expect(threshold.Value).To(Equal(42.0))
		Expect(threshold.Units).To(Equal(nontypesglucose.MgdL))
	})
	It("accepts mmol/L", func() {
		buf := buff(`{"units":"%s","value":42}`, nontypesglucose.MmolL)
		threshold := &Threshold{}
		err := request.DecodeObject(context.Background(), nil, buf, threshold)
		Expect(err).To(BeNil())
		Expect(threshold.Value).To(Equal(42.0))
		Expect(threshold.Units).To(Equal(nontypesglucose.MmolL))
	})
	It("rejects lb/gal", func() {
		buf := buff(`{"units":"%s","value":42}`, "lb/gal")
		err := request.DecodeObject(context.Background(), nil, buf, &Threshold{})
		Expect(err).Should(HaveOccurred())
	})
	It("rejects blank units", func() {
		buf := buff(`{"units":"","value":42}`)
		err := request.DecodeObject(context.Background(), nil, buf, &Threshold{})
		Expect(err).Should(HaveOccurred())
	})
	It("is case-sensitive with respect to Units", func() {
		badUnits := strings.ToUpper(nontypesglucose.MmolL)
		buf := buff(`{"units":"%s","value":42}`, badUnits)
		err := request.DecodeObject(context.Background(), nil, buf, &Threshold{})
		Expect(err).Should(HaveOccurred())
	})

})

var _ = Describe("AlertActivity", func() {
	Describe("IsActive()", func() {
		It("is true", func() {
			triggered := time.Now()
			resolved := triggered.Add(-time.Nanosecond)
			a := AlertActivity{
				Triggered: triggered,
				Resolved:  resolved,
			}
			Expect(a.IsActive()).To(BeTrue())
		})

		It("is false", func() {
			triggered := time.Now()
			resolved := triggered.Add(time.Nanosecond)
			a := AlertActivity{
				Triggered: triggered,
				Resolved:  resolved,
			}
			Expect(a.IsActive()).To(BeFalse())
		})
	})

	Describe("IsSent()", func() {
		It("is true", func() {
			triggered := time.Now()
			sent := triggered.Add(time.Nanosecond)
			a := AlertActivity{
				Triggered: triggered,
				Sent:      sent,
			}
			Expect(a.IsSent()).To(BeTrue())
		})

		It("is false", func() {
			triggered := time.Now()
			notified := triggered.Add(-time.Nanosecond)
			a := AlertActivity{
				Triggered: triggered,
				Sent:      notified,
			}
			Expect(a.IsSent()).To(BeFalse())
		})
	})

	Describe("normalizeUnits", func() {
		Context("given the same units", func() {
			It("doesn't alter them at all", func() {
				d := testUrgentLowDatum()
				t := Threshold{
					Value: 5.0,
					Units: nontypesglucose.MmolL,
				}
				dv, tv, err := normalizeUnits(d, t)
				Expect(err).To(Succeed())
				Expect(tv).To(Equal(5.0))
				Expect(dv).To(Equal(2.9))

				d = testUrgentLowDatum()
				d.Blood.Units = pointer.FromAny(nontypesglucose.MgdL)
				t = Threshold{
					Value: 5.0,
					Units: nontypesglucose.MgdL,
				}
				dv, tv, err = normalizeUnits(d, t)
				Expect(err).To(Succeed())
				Expect(tv).To(Equal(5.0))
				Expect(dv).To(Equal(2.9))
			})
		})

		Context("value in Mmol/L & threshold in mg/dL", func() {
			It("normalizes to Mmol/L", func() {
				d := testUrgentLowDatum()
				d.Blood.Units = pointer.FromAny(nontypesglucose.MmolL)
				t := Threshold{
					Value: 90.0,
					Units: nontypesglucose.MgdL,
				}
				dv, tv, err := normalizeUnits(d, t)
				Expect(err).To(Succeed())
				Expect(tv).To(Equal(4.99567))
				Expect(dv).To(Equal(2.9))
			})
		})

		Context("value in mg/dL & threshold in Mmol/L", func() {
			It("normalizes to Mmol/L", func() {
				d := testUrgentLowDatum()
				d.Blood.Value = pointer.FromAny(90.0)
				d.Blood.Units = pointer.FromAny(nontypesglucose.MgdL)
				t := Threshold{
					Value: 5.0,
					Units: nontypesglucose.MmolL,
				}
				dv, tv, err := normalizeUnits(d, t)
				Expect(err).To(Succeed())
				Expect(tv).To(Equal(5.0))
				Expect(dv).To(Equal(4.99567))
			})
		})
	})
})

// buff is a helper for generating a JSON []byte representation.
func buff(format string, args ...interface{}) *bytes.Buffer {
	return bytes.NewBufferString(fmt.Sprintf(format, args...))
}

func testDosingDecision(d time.Duration) *DosingDecision {
	return &DosingDecision{
		Base: types.Base{
			Time: pointer.FromAny(time.Now().Add(d)),
		},
		Reason: pointer.FromAny(DosingDecisionReasonLoop),
	}
}

func testConfig() Config {
	return Config{
		UserID:         mockUserID1,
		FollowedUserID: mockUserID2,
		UploadID:       mockDataSetID,
	}
}

func testUrgentLowDatum() *Glucose {
	return &Glucose{
		Blood: blood.Blood{
			Base: types.Base{
				Time: pointer.FromAny(time.Now()),
			},
			Units: pointer.FromAny(nontypesglucose.MmolL),
			Value: pointer.FromAny(2.9),
		},
	}
}

func testHighDatum() *Glucose {
	return &Glucose{
		Blood: blood.Blood{
			Base: types.Base{
				Time: pointer.FromAny(time.Now()),
			},
			Units: pointer.FromAny(nontypesglucose.MmolL),
			Value: pointer.FromAny(11.0),
		},
	}
}

func testLowDatum() *Glucose {
	return &Glucose{
		Blood: blood.Blood{
			Base: types.Base{
				Time: pointer.FromAny(time.Now()),
			},
			Units: pointer.FromAny(nontypesglucose.MmolL),
			Value: pointer.FromAny(3.9),
		},
	}
}

func testInRangeDatum() *Glucose {
	return &Glucose{
		Blood: blood.Blood{
			Base: types.Base{
				Time: pointer.FromAny(time.Now()),
			},
			Units: pointer.FromAny(nontypesglucose.MmolL),
			Value: pointer.FromAny(6.0),
		},
	}
}

func testNoCommunication() *NoCommunicationAlert {
	return &NoCommunicationAlert{
		Base: Base{Enabled: true},
	}
}

func testNoCommunicationDisabled() *NoCommunicationAlert {
	nc := testNoCommunication()
	nc.Enabled = false
	return nc
}

func testNotLoopingDisabled() *NotLoopingAlert {
	nl := testNotLooping()
	nl.Enabled = false
	return nl
}

func testNotLooping() *NotLoopingAlert {
	return &NotLoopingAlert{
		Base:  Base{Enabled: true},
		Delay: 0,
	}
}

func testAlertsActivity() Activity {
	return Activity{}
}

func testLowAlert() *LowAlert {
	return &LowAlert{
		Base: Base{Enabled: true},
		Threshold: Threshold{
			Value: 4,
			Units: nontypesglucose.MmolL,
		},
	}
}
func testHighAlert() *HighAlert {
	return &HighAlert{
		Base: Base{Enabled: true},
		Threshold: Threshold{
			Value: 10,
			Units: nontypesglucose.MmolL,
		},
	}
}
func testUrgentLowAlert() *UrgentLowAlert {
	return &UrgentLowAlert{
		Base: Base{Enabled: true},
		Threshold: Threshold{
			Value: 3,
			Units: nontypesglucose.MmolL,
		},
	}
}
func testNotLoopingAlert() *NotLoopingAlert {
	return &NotLoopingAlert{
		Base: Base{Enabled: true},
	}
}

func newConfigTest() (context.Context, *logTest.Logger, *Config) {
	lgr := logTest.NewLogger()
	ctx := log.NewContextWithLogger(context.Background(), lgr)
	cfg := &Config{
		UserID:         mockUserID1,
		FollowedUserID: mockUserID2,
		UploadID:       mockDataSetID,
		Alerts: Alerts{
			UrgentLow:       testUrgentLowAlert(),
			Low:             testLowAlert(),
			High:            testHighAlert(),
			NotLooping:      testNotLoopingDisabled(),      // NOTE: disabled
			NoCommunication: testNoCommunicationDisabled(), // NOTE: disabled
		},
		Activity: testAlertsActivity(),
	}
	return ctx, lgr, cfg
}

func quickJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("<error marshaling %T>", v)
	}
	return string(b)
}
