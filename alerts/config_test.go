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

	"github.com/tidepool-org/platform/data/blood/glucose"
	logTest "github.com/tidepool-org/platform/log/test"
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
		err := request.DecodeObject(context.Background(), nil, buf, conf)
		Expect(err).ToNot(HaveOccurred())
		Expect(conf.UserID).To(Equal(mockUserID1))
		Expect(conf.FollowedUserID).To(Equal(mockUserID2))
		Expect(conf.UploadID).To(Equal(mockUploadID))
		Expect(conf.High.Enabled).To(Equal(false))
		Expect(conf.High.Repeat).To(Equal(DurationMinutes(30 * time.Minute)))
		Expect(conf.High.Delay).To(Equal(DurationMinutes(5 * time.Minute)))
		Expect(conf.High.Threshold.Value).To(Equal(10.0))
		Expect(conf.High.Threshold.Units).To(Equal(glucose.MmolL))
		Expect(conf.Low.Enabled).To(Equal(true))
		Expect(conf.Low.Repeat).To(Equal(DurationMinutes(30 * time.Minute)))
		Expect(conf.Low.Delay).To(Equal(DurationMinutes(10 * time.Minute)))
		Expect(conf.Low.Threshold.Value).To(Equal(80.0))
		Expect(conf.Low.Threshold.Units).To(Equal(glucose.MgdL))
		Expect(conf.UrgentLow.Enabled).To(Equal(false))
		Expect(conf.UrgentLow.Threshold.Value).To(Equal(47.5))
		Expect(conf.UrgentLow.Threshold.Units).To(Equal(glucose.MgdL))
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
			val := validator.New(logTest.NewLogger())
			c.Validate(val)
			Expect(val.Error()).To(MatchError(ContainSubstring("value is empty")))
		})

		It("requires an FollowedUserID", func() {
			c := testConfig()
			c.FollowedUserID = ""
			val := validator.New(logTest.NewLogger())
			c.Validate(val)
			Expect(val.Error()).To(MatchError(ContainSubstring("value is empty")))
		})

		It("requires an UserID", func() {
			c := testConfig()
			c.UserID = ""
			val := validator.New(logTest.NewLogger())
			c.Validate(val)
			Expect(val.Error()).To(MatchError(ContainSubstring("value is empty")))
		})
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
	})

	Context("repeat", func() {
		var defaultAlert = LowAlert{
			Threshold: Threshold{Value: 11, Units: glucose.MmolL},
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
}`, mockUserID1, mockUserID2, mockUploadID)
			conf := &Config{}
			err := request.DecodeObject(context.Background(), nil, buf, conf)
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
}`, mockUserID1, mockUserID2, mockUploadID, glucose.MgdL)
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
}`, mockUserID1, mockUserID2, mockUploadID, glucose.MgdL)
		cfg := &Config{}
		err := request.DecodeObject(context.Background(), nil, buf, cfg)
		Expect(err).To(MatchError("json is malformed"))
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
		buf := buff(`{"units":"%s","value":42}`, glucose.MgdL)
		threshold := &Threshold{}
		err := request.DecodeObject(context.Background(), nil, buf, threshold)
		Expect(err).To(BeNil())
		Expect(threshold.Value).To(Equal(42.0))
		Expect(threshold.Units).To(Equal(glucose.MgdL))
	})
	It("accepts mmol/L", func() {
		buf := buff(`{"units":"%s","value":42}`, glucose.MmolL)
		threshold := &Threshold{}
		err := request.DecodeObject(context.Background(), nil, buf, threshold)
		Expect(err).To(BeNil())
		Expect(threshold.Value).To(Equal(42.0))
		Expect(threshold.Units).To(Equal(glucose.MmolL))
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
		badUnits := strings.ToUpper(glucose.MmolL)
		buf := buff(`{"units":"%s","value":42}`, badUnits)
		err := request.DecodeObject(context.Background(), nil, buf, &Threshold{})
		Expect(err).Should(HaveOccurred())
	})

})

// buff is a helper for generating a JSON []byte representation.
func buff(format string, args ...interface{}) *bytes.Buffer {
	return bytes.NewBufferString(fmt.Sprintf(format, args...))
}
