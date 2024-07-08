package alerts

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	nontypesglucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/log"
	logtest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func TestSuite(t *testing.T) {
	test.Test(t)
}

const (
	mockUserID1  = "008c7f79-6545-4466-95fb-34e3ba728d38"
	mockUserID2  = "b1880201-30d5-4190-92bb-6afcf08ca15e"
	mockUploadID = "4d3b1abc280511ef9f41abf13a093b64"
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
}`, mockUserID1, mockUserID2, mockUploadID)
		conf := &Config{}
		err := request.DecodeObject(nil, buf, conf)
		Expect(err).ToNot(HaveOccurred())
		Expect(conf.UserID).To(Equal(mockUserID1))
		Expect(conf.FollowedUserID).To(Equal(mockUserID2))
		Expect(conf.UploadID).To(Equal(mockUploadID))
		Expect(conf.High.Enabled).To(Equal(false))
		Expect(conf.High.Repeat).To(Equal(DurationMinutes(30 * time.Minute)))
		Expect(conf.High.Delay).To(Equal(DurationMinutes(5 * time.Minute)))
		Expect(conf.High.Threshold.Value).To(Equal(10.0))
		Expect(conf.High.Threshold.Units).To(Equal(nontypesglucose.MmolL))
		Expect(conf.Low.Enabled).To(Equal(true))
		Expect(conf.Low.Repeat).To(Equal(DurationMinutes(30 * time.Minute)))
		Expect(conf.Low.Delay).To(Equal(DurationMinutes(10 * time.Minute)))
		Expect(conf.Low.Threshold.Value).To(Equal(80.0))
		Expect(conf.Low.Threshold.Units).To(Equal(nontypesglucose.MgdL))
		Expect(conf.UrgentLow.Enabled).To(Equal(false))
		Expect(conf.UrgentLow.Threshold.Value).To(Equal(47.5))
		Expect(conf.UrgentLow.Threshold.Units).To(Equal(nontypesglucose.MgdL))
		Expect(conf.NotLooping.Enabled).To(Equal(true))
		Expect(conf.NotLooping.Delay).To(Equal(DurationMinutes(4 * time.Minute)))
		Expect(conf.NoCommunication.Enabled).To(Equal(true))
		Expect(conf.NoCommunication.Delay).To(Equal(DurationMinutes(6 * time.Minute)))
	})

	Context("validations", func() {
		testConfig := func() Config {
			return Config{
				UserID:         mockUserID1,
				FollowedUserID: mockUserID2,
				UploadID:       mockUploadID,
			}
		}

		It("requires an UploadID", func() {
			c := testConfig()
			c.UploadID = ""
			val := validator.New()
			c.Validate(val)
			Expect(val.Error()).To(MatchError(ContainSubstring("value is empty")))
		})

		It("requires an FollowedUserID", func() {
			c := testConfig()
			c.FollowedUserID = ""
			val := validator.New()
			c.Validate(val)
			Expect(val.Error()).To(MatchError(ContainSubstring("value is empty")))
		})

		It("requires an UserID", func() {
			c := testConfig()
			c.UserID = ""
			val := validator.New()
			c.Validate(val)
			Expect(val.Error()).To(MatchError(ContainSubstring("value is empty")))
		})
	})

	Describe("Evaluate", func() {
		Context("when a note is returned", func() {
			It("injects the userIDs", func() {
				ctx := contextWithTestLogger()
				mockGlucoseData := []*glucose.Glucose{
					{
						Blood: blood.Blood{
							Base: types.Base{
								Time: pointer.FromAny(time.Now()),
							},
							Units: pointer.FromAny(nontypesglucose.MmolL),
							Value: pointer.FromAny(0.0),
						},
					},
				}
				conf := Config{
					UserID:         mockUserID1,
					FollowedUserID: mockUserID2,
					Alerts: Alerts{
						UrgentLow: &UrgentLowAlert{
							Base: Base{Enabled: true},
							Threshold: Threshold{
								Value: 10,
								Units: nontypesglucose.MmolL,
							},
						},
					},
				}

				note := conf.Evaluate(ctx, mockGlucoseData, nil)

				Expect(note).ToNot(BeNil())
				Expect(note.RecipientUserID).To(Equal(mockUserID1))
				Expect(note.FollowedUserID).To(Equal(mockUserID2))
			})
		})
	})

	Context("Base", func() {
		Context("Activity", func() {
			Context("IsActive()", func() {
				It("is true", func() {
					triggered := time.Now()
					resolved := triggered.Add(-time.Nanosecond)
					a := Activity{
						Triggered: triggered,
						Resolved:  resolved,
					}
					Expect(a.IsActive()).To(BeTrue())
				})

				It("is false", func() {
					triggered := time.Now()
					resolved := triggered.Add(time.Nanosecond)
					a := Activity{
						Triggered: triggered,
						Resolved:  resolved,
					}
					Expect(a.IsActive()).To(BeFalse())
				})
			})

			Context("IsSent()", func() {
				It("is true", func() {
					triggered := time.Now()
					sent := triggered.Add(time.Nanosecond)
					a := Activity{
						Triggered: triggered,
						Sent:      sent,
					}
					Expect(a.IsSent()).To(BeTrue())
				})

				It("is false", func() {
					triggered := time.Now()
					notified := triggered.Add(-time.Nanosecond)
					a := Activity{
						Triggered: triggered,
						Sent:      notified,
					}
					Expect(a.IsSent()).To(BeFalse())
				})
			})
		})
	})

	var testGlucoseDatum = func(v float64) *glucose.Glucose {
		return &glucose.Glucose{
			Blood: blood.Blood{
				Base: types.Base{
					Time: pointer.FromAny(time.Now()),
				},
				Units: pointer.FromAny(nontypesglucose.MmolL),
				Value: pointer.FromAny(v),
			},
		}
	}

	Context("UrgentLowAlert", func() {
		Context("Threshold", func() {
			It("accepts values between 0 and 1000 mg/dL", func() {
				val := validator.New()
				b := UrgentLowAlert{Threshold: Threshold{Value: 0, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = UrgentLowAlert{Threshold: Threshold{Value: 1000, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = UrgentLowAlert{Threshold: Threshold{Value: 1001, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 1001 is not between 0 and 1000"))

				val = validator.New()
				b = UrgentLowAlert{Threshold: Threshold{Value: -1, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1 is not between 0 and 1000"))
			})
		})

		Context("Evaluate", func() {
			testUrgentLow := func() *UrgentLowAlert {
				return &UrgentLowAlert{
					Threshold: Threshold{
						Value: 4.0,
						Units: nontypesglucose.MmolL,
					},
				}
			}

			It("handles being passed empty data", func() {
				ctx := contextWithTestLogger()
				var note *Note

				alert := testUrgentLow()

				Expect(func() {
					note = alert.Evaluate(ctx, []*glucose.Glucose{})
				}).ToNot(Panic())
				Expect(note).To(BeNil())
				Expect(func() {
					note = alert.Evaluate(ctx, nil)
				}).ToNot(Panic())
				Expect(note).To(BeNil())
			})

			It("logs evaluation results", func() {
				ctx := contextWithTestLogger()
				data := []*glucose.Glucose{testGlucoseDatum(1.1)}

				alert := testUrgentLow()

				Expect(func() {
					alert.Evaluate(ctx, data)
				}).ToNot(Panic())
				Expect(func() {
					lgr := log.LoggerFromContext(ctx).(*logtest.Logger)
					lgr.AssertLog(log.InfoLevel, "urgent low", log.Fields{
						"threshold":   4.0,
						"value":       1.1,
						"isAlerting?": true,
					})
				}).ToNot(Panic())
			})

			Context("when currently active", func() {
				It("marks itself resolved", func() {
					ctx := contextWithTestLogger()

					alert := testUrgentLow()

					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(1.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).To(BeZero())
					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(5.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).ToNot(BeZero())
				})
			})

			Context("when currently INactive", func() {
				It("doesn't re-mark itself resolved", func() {
					ctx := contextWithTestLogger()

					alert := testUrgentLow()

					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(1.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).To(BeZero())
					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(5.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).ToNot(BeZero())
					was := alert.Resolved
					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(5.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).To(Equal(was))
				})
			})

			It("marks itself triggered", func() {
				ctx := contextWithTestLogger()

				alert := testUrgentLow()

				Expect(func() {
					alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(5.0)})
				}).ToNot(Panic())
				Expect(alert.Triggered).To(BeZero())
				Expect(func() {
					alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(1.0)})
				}).ToNot(Panic())
				Expect(alert.Triggered).ToNot(BeZero())
			})

			It("validates glucose data", func() {
				ctx := contextWithTestLogger()
				var note *Note

				Expect(func() {
					note = testUrgentLow().Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(1)})
				}).ToNot(Panic())
				Expect(note).ToNot(BeNil())

				badUnits := testGlucoseDatum(1)
				badUnits.Units = nil
				Expect(func() {
					note = testUrgentLow().Evaluate(ctx, []*glucose.Glucose{badUnits})
				}).ToNot(Panic())
				Expect(note).To(BeNil())

				badValue := testGlucoseDatum(1)
				badValue.Value = nil
				Expect(func() {
					note = testUrgentLow().Evaluate(ctx, []*glucose.Glucose{badValue})
				}).ToNot(Panic())
				Expect(note).To(BeNil())

				badTime := testGlucoseDatum(1)
				badTime.Time = nil
				Expect(func() {
					note = testUrgentLow().Evaluate(ctx, []*glucose.Glucose{badTime})
				}).ToNot(Panic())
				Expect(note).To(BeNil())

			})
		})
	})

	Context("LowAlert", func() {
		Context("Threshold", func() {
			It("accepts values in mmol/L", func() {
				val := validator.New()
				b := LowAlert{Threshold: Threshold{Value: 4.2735, Units: "mmol/L"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())
			})

			It("accepts values between 0 and 1000 mg/dL", func() {
				val := validator.New()
				b := LowAlert{Threshold: Threshold{Value: 0, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = LowAlert{Threshold: Threshold{Value: 1000, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = LowAlert{Threshold: Threshold{Value: 1001, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 1001 is not between 0 and 1000"))

				val = validator.New()
				b = LowAlert{Threshold: Threshold{Value: -1, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1 is not between 0 and 1000"))
			})
		})

		Context("Delay", func() {
			It("accepts values between 0 and 6 hours (inclusive)", func() {
				okThresh := Threshold{Units: "mg/dL", Value: 123}

				val := validator.New()
				b := HighAlert{Delay: 0, Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = HighAlert{Delay: DurationMinutes(time.Hour * 6 / time.Minute), Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = HighAlert{Delay: -1, Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1ns is not between 0s and 6h0m0s"))

				val = validator.New()
				b = HighAlert{Delay: DurationMinutes(time.Hour*6 + time.Minute), Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 6h1m0s is not between 0s and 6h0m0s"))
			})
		})

		Context("Evaluate", func() {
			testLow := func() *LowAlert {
				return &LowAlert{
					Threshold: Threshold{
						Value: 4.0,
						Units: nontypesglucose.MmolL,
					},
				}
			}

			It("handles being passed empty data", func() {
				ctx := contextWithTestLogger()
				var note *Note

				alert := testLow()

				Expect(func() {
					note = alert.Evaluate(ctx, []*glucose.Glucose{})
				}).ToNot(Panic())
				Expect(note).To(BeNil())
				Expect(func() {
					note = alert.Evaluate(ctx, nil)
				}).ToNot(Panic())
				Expect(note).To(BeNil())
			})

			It("logs evaluation results", func() {
				ctx := contextWithTestLogger()
				data := []*glucose.Glucose{testGlucoseDatum(1.1)}

				alert := testLow()

				Expect(func() {
					alert.Evaluate(ctx, data)
				}).ToNot(Panic())
				Expect(func() {
					lgr := log.LoggerFromContext(ctx).(*logtest.Logger)
					lgr.AssertLog(log.InfoLevel, "low", log.Fields{
						"threshold":   4.0,
						"value":       1.1,
						"isAlerting?": true,
					})
				}).ToNot(Panic())
			})

			Context("when currently active", func() {
				It("marks itself resolved", func() {
					ctx := contextWithTestLogger()

					alert := testLow()

					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(1.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).To(BeZero())
					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(5.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).ToNot(BeZero())
				})
			})

			Context("when currently INactive", func() {
				It("doesn't re-mark itself resolved", func() {
					ctx := contextWithTestLogger()

					alert := testLow()

					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(1.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).To(BeZero())
					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(5.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).ToNot(BeZero())
					was := alert.Resolved
					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(5.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).To(Equal(was))
				})
			})

			It("marks itself triggered", func() {
				ctx := contextWithTestLogger()

				alert := testLow()

				Expect(func() {
					alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(5.0)})
				}).ToNot(Panic())
				Expect(alert.Triggered).To(BeZero())
				Expect(func() {
					alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(1.0)})
				}).ToNot(Panic())
				Expect(alert.Triggered).ToNot(BeZero())
			})

			It("validates glucose data", func() {
				ctx := contextWithTestLogger()
				var note *Note

				Expect(func() {
					note = testLow().Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(1)})
				}).ToNot(Panic())
				Expect(note).ToNot(BeNil())

				badUnits := testGlucoseDatum(1)
				badUnits.Units = nil
				Expect(func() {
					note = testLow().Evaluate(ctx, []*glucose.Glucose{badUnits})
				}).ToNot(Panic())
				Expect(note).To(BeNil())

				badValue := testGlucoseDatum(1)
				badValue.Value = nil
				Expect(func() {
					note = testLow().Evaluate(ctx, []*glucose.Glucose{badValue})
				}).ToNot(Panic())
				Expect(note).To(BeNil())

				badTime := testGlucoseDatum(1)
				badTime.Time = nil
				Expect(func() {
					note = testLow().Evaluate(ctx, []*glucose.Glucose{badTime})
				}).ToNot(Panic())
				Expect(note).To(BeNil())
			})
		})
	})

	Context("HighAlert", func() {
		Context("Threshold", func() {
			It("accepts values between 0 and 1000 mg/dL", func() {
				val := validator.New()
				b := HighAlert{Threshold: Threshold{Value: 0, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = HighAlert{Threshold: Threshold{Value: 1000, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = HighAlert{Threshold: Threshold{Value: 1001, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 1001 is not between 0 and 1000"))

				val = validator.New()
				b = HighAlert{Threshold: Threshold{Value: -1, Units: "mg/dL"}}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1 is not between 0 and 1000"))
			})
		})

		Context("Delay", func() {
			It("accepts values between 0 and 6 hours (inclusive)", func() {
				okThresh := Threshold{Units: "mg/dL", Value: 123}

				val := validator.New()
				b := HighAlert{Delay: 0, Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = HighAlert{Delay: DurationMinutes(time.Hour * 6 / time.Minute), Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = HighAlert{Delay: -1, Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1ns is not between 0s and 6h0m0s"))

				val = validator.New()
				b = HighAlert{Delay: DurationMinutes(time.Hour*6 + time.Minute), Threshold: okThresh}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 6h1m0s is not between 0s and 6h0m0s"))
			})
		})

		Context("Evaluate", func() {
			testHigh := func() *HighAlert {
				return &HighAlert{
					Threshold: Threshold{
						Value: 20.0,
						Units: nontypesglucose.MmolL,
					},
				}
			}

			It("handles being passed empty data", func() {
				ctx := contextWithTestLogger()
				var note *Note

				alert := testHigh()

				Expect(func() {
					note = alert.Evaluate(ctx, []*glucose.Glucose{})
				}).ToNot(Panic())
				Expect(note).To(BeNil())
				Expect(func() {
					note = alert.Evaluate(ctx, nil)
				}).ToNot(Panic())
				Expect(note).To(BeNil())
			})

			It("logs evaluation results", func() {
				ctx := contextWithTestLogger()
				data := []*glucose.Glucose{testGlucoseDatum(21.1)}

				alert := testHigh()

				Expect(func() {
					alert.Evaluate(ctx, data)
				}).ToNot(Panic())
				Expect(func() {
					lgr := log.LoggerFromContext(ctx).(*logtest.Logger)
					lgr.AssertLog(log.InfoLevel, "high", log.Fields{
						"threshold":   20.0,
						"value":       21.1,
						"isAlerting?": true,
					})
				}).ToNot(Panic())
			})

			Context("when currently active", func() {
				It("marks itself resolved", func() {
					ctx := contextWithTestLogger()

					alert := testHigh()

					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(21.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).To(BeZero())
					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(5.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).ToNot(BeZero())
				})
			})

			Context("when currently INactive", func() {
				It("doesn't re-mark itself resolved", func() {
					ctx := contextWithTestLogger()

					alert := testHigh()

					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(21.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).To(BeZero())
					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(5.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).ToNot(BeZero())
					was := alert.Resolved
					Expect(func() {
						alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(5.0)})
					}).ToNot(Panic())
					Expect(alert.Resolved).To(Equal(was))
				})
			})

			It("marks itself triggered", func() {
				ctx := contextWithTestLogger()

				alert := testHigh()

				Expect(func() {
					alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(5.0)})
				}).ToNot(Panic())
				Expect(alert.Triggered).To(BeZero())
				Expect(func() {
					alert.Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(21.0)})
				}).ToNot(Panic())
				Expect(alert.Triggered).ToNot(BeZero())
			})

			It("validates glucose data", func() {
				ctx := contextWithTestLogger()
				var note *Note

				Expect(func() {
					note = testHigh().Evaluate(ctx, []*glucose.Glucose{testGlucoseDatum(21)})
				}).ToNot(Panic())
				Expect(note).ToNot(BeNil())

				badUnits := testGlucoseDatum(1)
				badUnits.Units = nil
				Expect(func() {
					note = testHigh().Evaluate(ctx, []*glucose.Glucose{badUnits})
				}).ToNot(Panic())
				Expect(note).To(BeNil())

				badValue := testGlucoseDatum(1)
				badValue.Value = nil
				Expect(func() {
					note = testHigh().Evaluate(ctx, []*glucose.Glucose{badValue})
				}).ToNot(Panic())
				Expect(note).To(BeNil())

				badTime := testGlucoseDatum(1)
				badTime.Time = nil
				Expect(func() {
					note = testHigh().Evaluate(ctx, []*glucose.Glucose{badTime})
				}).ToNot(Panic())
				Expect(note).To(BeNil())
			})
		})
	})

	Context("NoCommunicationAlert", func() {
		Context("Delay", func() {
			It("accepts values between 0 and 6 hours (inclusive)", func() {
				val := validator.New()
				b := NoCommunicationAlert{Delay: 0}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = NoCommunicationAlert{Delay: DurationMinutes(time.Hour * 6)}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = NoCommunicationAlert{Delay: -1}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1ns is not between 0s and 6h0m0s"))

				val = validator.New()
				b = NoCommunicationAlert{Delay: DurationMinutes(time.Hour*6 + time.Second)}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 6h0m1s is not between 0s and 6h0m0s"))
			})
		})
	})

	Context("NotLoopingAlert", func() {
		Context("Delay", func() {
			It("accepts values between 0 and 2 hours (inclusive)", func() {
				val := validator.New()
				b := NotLoopingAlert{Delay: 0}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = NotLoopingAlert{Delay: DurationMinutes(2 * time.Hour)}
				b.Validate(val)
				Expect(val.Error()).To(Succeed())

				val = validator.New()
				b = NotLoopingAlert{Delay: -1}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value -1ns is not between 0s and 2h0m0s"))

				val = validator.New()
				b = NotLoopingAlert{Delay: DurationMinutes(2*time.Hour + time.Second)}
				b.Validate(val)
				Expect(val.Error()).To(MatchError("value 2h0m1s is not between 0s and 2h0m0s"))
			})

		})
	})

	Context("repeat", func() {
		var defaultAlert = LowAlert{
			Threshold: Threshold{Value: 11, Units: nontypesglucose.MmolL},
		}

		It("accepts values of 0 (indicating disabled)", func() {
			val := validator.New()
			l := defaultAlert
			l.Repeat = 0
			l.Validate(val)
			Expect(val.Error()).To(Succeed())
		})

		It("accepts values of 15 minutes to 4 hours (inclusive)", func() {
			val := validator.New()
			l := defaultAlert
			l.Repeat = DurationMinutes(15 * time.Minute)
			l.Validate(val)
			Expect(val.Error()).To(Succeed())

			val = validator.New()
			l = defaultAlert
			l.Repeat = DurationMinutes(4 * time.Hour)
			l.Validate(val)
			Expect(val.Error()).To(Succeed())

			val = validator.New()
			l = defaultAlert
			l.Repeat = DurationMinutes(4*time.Hour + 1)
			l.Validate(val)
			Expect(val.Error()).NotTo(Succeed())

			val = validator.New()
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
			err := request.DecodeObject(nil, buf, threshold)
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
}`, mockUserID1, mockUserID2, mockUploadID)
			conf := &Config{}
			err := request.DecodeObject(nil, buf, conf)
			Expect(err).To(Succeed())
			Expect(conf.Low.Repeat).To(Equal(DurationMinutes(0)))
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
}`, mockUserID1, mockUserID2, mockUploadID, nontypesglucose.MgdL)
		cfg := &Config{}
		err := request.DecodeObject(nil, buf, cfg)
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
}`, mockUserID1, mockUserID2, mockUploadID, nontypesglucose.MgdL)
		cfg := &Config{}
		err := request.DecodeObject(nil, buf, cfg)
		Expect(err).To(MatchError("json is malformed"))
	})
})

var (
	testNoCommunicationAlert = func() *NoCommunicationAlert {
		return &NoCommunicationAlert{
			Base: Base{Enabled: true},
		}
	}
	testLowAlert = func() *LowAlert {
		return &LowAlert{
			Base: Base{Enabled: true},
			Threshold: Threshold{
				Value: 4,
				Units: nontypesglucose.MmolL,
			},
		}
	}
	testHighAlert = func() *HighAlert {
		return &HighAlert{
			Base: Base{Enabled: true},
			Threshold: Threshold{
				Value: 10,
				Units: nontypesglucose.MmolL,
			},
		}
	}
	testUrgentLowAlert = func() *UrgentLowAlert {
		return &UrgentLowAlert{
			Base: Base{Enabled: true},
			Threshold: Threshold{
				Value: 3,
				Units: nontypesglucose.MmolL,
			},
		}
	}
	testNotLoopingAlert = func() *NotLoopingAlert {
		return &NotLoopingAlert{
			Base: Base{Enabled: true},
		}
	}
	testNoCommunicationDatum = &glucose.Glucose{
		Blood: blood.Blood{
			Base: types.Base{
				Time: pointer.FromAny(time.Now()),
			},
			Units: pointer.FromAny(nontypesglucose.MmolL),
			Value: pointer.FromAny(11.0),
		},
	}
	testHighDatum = &glucose.Glucose{
		Blood: blood.Blood{
			Base: types.Base{
				Time: pointer.FromAny(time.Now()),
			},
			Units: pointer.FromAny(nontypesglucose.MmolL),
			Value: pointer.FromAny(11.0),
		},
	}
	testLowDatum = &glucose.Glucose{
		Blood: blood.Blood{
			Base: types.Base{
				Time: pointer.FromAny(time.Now()),
			},
			Units: pointer.FromAny(nontypesglucose.MmolL),
			Value: pointer.FromAny(3.9),
		},
	}
	testUrgentLowDatum = &glucose.Glucose{
		Blood: blood.Blood{
			Base: types.Base{
				Time: pointer.FromAny(time.Now()),
			},
			Units: pointer.FromAny(nontypesglucose.MmolL),
			Value: pointer.FromAny(2.9),
		},
	}
)

var _ = Describe("Alerts", func() {
	Describe("LongestDelay", func() {
		It("does what it says", func() {
			noComm := testNoCommunicationAlert()
			noComm.Delay = DurationMinutes(10 * time.Minute)
			low := testLowAlert()
			low.Delay = DurationMinutes(5 * time.Minute)
			high := testHighAlert()
			high.Delay = DurationMinutes(5 * time.Minute)
			notLooping := testNotLoopingAlert()
			notLooping.Delay = DurationMinutes(5 * time.Minute)

			a := Alerts{
				NoCommunication: noComm,
				Low:             low,
				High:            high,
				NotLooping:      notLooping,
			}

			delay := a.LongestDelay()

			Expect(delay).To(Equal(10 * time.Minute))
		})

		It("ignores disabled alerts", func() {
			noComm := testNoCommunicationAlert()
			noComm.Delay = DurationMinutes(10 * time.Minute)
			noComm.Enabled = false
			low := testLowAlert()
			low.Delay = DurationMinutes(7 * time.Minute)
			high := testHighAlert()
			high.Delay = DurationMinutes(5 * time.Minute)
			notLooping := testNotLoopingAlert()
			notLooping.Delay = DurationMinutes(5 * time.Minute)

			a := Alerts{
				NoCommunication: noComm,
				Low:             low,
				High:            high,
				NotLooping:      notLooping,
			}

			delay := a.LongestDelay()

			Expect(delay).To(Equal(7 * time.Minute))
		})

		It("returns a Zero Duration when no alerts are set", func() {
			a := Alerts{
				NoCommunication: nil,
				Low:             nil,
				High:            nil,
				NotLooping:      nil,
			}

			delay := a.LongestDelay()

			Expect(delay).To(Equal(time.Duration(0)))
		})
	})

	Describe("Evaluate", func() {
		Context("when not communicating", func() {
			It("returns only NoCommunication alerts", func() {
				ctx := contextWithTestLogger()
				data := []*glucose.Glucose{testNoCommunicationDatum}
				data[0].Value = pointer.FromAny(0.0)
				a := Alerts{
					NoCommunication: testNoCommunicationAlert(),
					UrgentLow:       testUrgentLowAlert(),
					Low:             testLowAlert(),
					High:            testHighAlert(),
				}

				note := a.Evaluate(ctx, data, nil)

				Expect(note).To(HaveField("Message", ContainSubstring(NoCommunicationMessage)))
			})
		})

		It("logs decisions", func() {
			Skip("TODO logAlertEvaluation")
		})

		It("detects low data", func() {
			ctx := contextWithTestLogger()
			data := []*glucose.Glucose{testLowDatum}
			a := Alerts{
				Low: testLowAlert(),
			}

			note := a.Evaluate(ctx, data, nil)

			Expect(note).ToNot(BeNil())
			Expect(note.Message).To(ContainSubstring("below low threshold"))
		})

		It("detects high data", func() {
			ctx := contextWithTestLogger()
			data := []*glucose.Glucose{testHighDatum}
			a := Alerts{
				High: testHighAlert(),
			}

			note := a.Evaluate(ctx, data, nil)

			Expect(note).ToNot(BeNil())
			Expect(note.Message).To(ContainSubstring("above high threshold"))
		})

		Context("with both low and urgent low alerts detected", func() {
			It("prefers urgent low", func() {
				ctx := contextWithTestLogger()
				data := []*glucose.Glucose{testUrgentLowDatum}
				a := Alerts{
					Low:       testLowAlert(),
					UrgentLow: testUrgentLowAlert(),
				}

				note := a.Evaluate(ctx, data, nil)

				Expect(note).ToNot(BeNil())
				Expect(note.Message).To(ContainSubstring("below urgent low threshold"))
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
		err := request.DecodeObject(nil, buf, threshold)
		Expect(err).To(BeNil())
		Expect(threshold.Value).To(Equal(42.0))
		Expect(threshold.Units).To(Equal(nontypesglucose.MgdL))
	})
	It("accepts mmol/L", func() {
		buf := buff(`{"units":"%s","value":42}`, nontypesglucose.MmolL)
		threshold := &Threshold{}
		err := request.DecodeObject(nil, buf, threshold)
		Expect(err).To(BeNil())
		Expect(threshold.Value).To(Equal(42.0))
		Expect(threshold.Units).To(Equal(nontypesglucose.MmolL))
	})
	It("rejects lb/gal", func() {
		buf := buff(`{"units":"%s","value":42}`, "lb/gal")
		err := request.DecodeObject(nil, buf, &Threshold{})
		Expect(err).Should(HaveOccurred())
	})
	It("rejects blank units", func() {
		buf := buff(`{"units":"","value":42}`)
		err := request.DecodeObject(nil, buf, &Threshold{})
		Expect(err).Should(HaveOccurred())
	})
	It("is case-sensitive with respect to Units", func() {
		badUnits := strings.ToUpper(nontypesglucose.MmolL)
		buf := buff(`{"units":"%s","value":42}`, badUnits)
		err := request.DecodeObject(nil, buf, &Threshold{})
		Expect(err).Should(HaveOccurred())
	})

})

// buff is a helper for generating a JSON []byte representation.
func buff(format string, args ...interface{}) *bytes.Buffer {
	return bytes.NewBufferString(fmt.Sprintf(format, args...))
}

func contextWithTestLogger() context.Context {
	lgr := logtest.NewLogger()
	return log.NewContextWithLogger(context.Background(), lgr)
}
