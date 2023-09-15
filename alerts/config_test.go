package alerts

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
)

func TestSuite(t *testing.T) {
	test.Test(t)
}

var _ = Describe("Config", func() {
	It("parses all the things", func() {
		buf := buff(`{
  "invitorID": "1",
  "ownerID": "2",
  "low": {
    "enabled": true,
    "repeat": 30,
    "delay": 10,
    "threshold": {
      "units": "mg/dL",
      "value": 123.4
    }
  },
  "urgentLow": {
    "enabled": false,
    "repeat": 30,
    "threshold": {
      "units": "mg/dL",
      "value": 456.7
    }
  },
  "high": {
    "enabled": false,
    "repeat": 30,
    "delay": 5,
    "threshold": {
      "units": "mmol/L",
      "value": 456.7
    }
  },
  "notLooping": {
    "enabled": true,
    "repeat": 32,
    "delay": 4
  },
  "noCommunication": {
    "enabled": true,
    "repeat": 33,
    "delay": 6
  }
}`)
		conf := &Config{}
		err := request.DecodeObject(nil, buf, conf)
		Expect(err).ToNot(HaveOccurred())
		Expect(conf.InvitorID).To(Equal("1"))
		Expect(conf.OwnerID).To(Equal("2"))
		Expect(conf.High.Enabled).To(Equal(false))
		Expect(conf.High.Repeat).To(Equal(DurationMinutes(30 * time.Minute)))
		Expect(conf.High.Delay).To(Equal(DurationMinutes(5 * time.Minute)))
		Expect(conf.High.Threshold.Value).To(Equal(456.7))
		Expect(conf.High.Threshold.Units).To(Equal(glucose.MmolL))
		Expect(conf.Low.Enabled).To(Equal(true))
		Expect(conf.Low.Repeat).To(Equal(DurationMinutes(30 * time.Minute)))
		Expect(conf.Low.Delay).To(Equal(DurationMinutes(10 * time.Minute)))
		Expect(conf.Low.Threshold.Value).To(Equal(123.4))
		Expect(conf.Low.Threshold.Units).To(Equal(glucose.MgdL))
		Expect(conf.UrgentLow.Enabled).To(Equal(false))
		Expect(conf.UrgentLow.Repeat).To(Equal(DurationMinutes(30 * time.Minute)))
		Expect(conf.UrgentLow.Threshold.Value).To(Equal(456.7))
		Expect(conf.UrgentLow.Threshold.Units).To(Equal(glucose.MgdL))
		Expect(conf.NotLooping.Enabled).To(Equal(true))
		Expect(conf.NotLooping.Repeat).To(Equal(DurationMinutes(32 * time.Minute)))
		Expect(conf.NotLooping.Delay).To(Equal(DurationMinutes(4 * time.Minute)))
		Expect(conf.NoCommunication.Enabled).To(Equal(true))
		Expect(conf.NoCommunication.Repeat).To(Equal(DurationMinutes(33 * time.Minute)))
		Expect(conf.NoCommunication.Delay).To(Equal(DurationMinutes(6 * time.Minute)))
	})

	Context("urgentLow", func() {
		It("validates threshold units", func() {
			buf := buff(`{"urgentLow": {"threshold": {"units":%q,"value":42}}`, "garbage")
			threshold := &Threshold{}
			err := request.DecodeObject(nil, buf, threshold)
			Expect(err).To(MatchError("json is malformed"))
		})
		It("validates repeat minutes (negative)", func() {
			buf := buff(`{
  "invitorID": "1", "ownerID": "2",
  "urgentLow": {
    "enabled": false,
    "repeat": -11,
    "threshold": {
      "units": "%s",
      "value": 1
    }
  }
}`, glucose.MgdL)
			cfg := &Config{}
			err := request.DecodeObject(nil, buf, cfg)
			Expect(err).To(MatchError("value -11m0s is not greater than 0s"))
		})
		It("validates repeat minutes (string)", func() {
			buf := buff(`{
  "invitorID": "1", "ownerID": "2",
  "urgentLow": {
    "enabled": false,
    "repeat": "a",
    "threshold": {
      "units": "%s",
      "value": 1
    }
  }
}`, glucose.MgdL)
			cfg := &Config{}
			err := request.DecodeObject(nil, buf, cfg)
			Expect(err).To(MatchError("json is malformed"))
		})
	})

	Context("low", func() {
		It("rejects a blank repeat", func() {
			buf := buff(`{
  "invitorID":"1", "ownerID":"2",
  "low": {
    "enabled": true,
    "delay": 10,
    "threshold": {
      "units": "mg/dL",
      "value": 123.4
    }
  }
}`)
			conf := &Config{}
			err := request.DecodeObject(nil, buf, conf)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("Duration", func() {
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
})

var _ = Describe("Threshold", func() {
	It("accepts mg/dL", func() {
		buf := buff(`{"units":%q,"value":42}`, glucose.MgdL)
		threshold := &Threshold{}
		err := request.DecodeObject(nil, buf, threshold)
		Expect(err).To(BeNil())
		Expect(threshold.Value).To(Equal(42.0))
		Expect(threshold.Units).To(Equal(glucose.MgdL))
	})
	It("accepts mmol/L", func() {
		buf := buff(`{"units":%q,"value":42}`, glucose.MmolL)
		threshold := &Threshold{}
		err := request.DecodeObject(nil, buf, threshold)
		Expect(err).To(BeNil())
		Expect(threshold.Value).To(Equal(42.0))
		Expect(threshold.Units).To(Equal(glucose.MmolL))
	})
	It("rejects lb/gal", func() {
		buf := buff(`{"units":%q,"value":42}`, "lb/gal")
		err := request.DecodeObject(nil, buf, &Threshold{})
		Expect(err).Should(HaveOccurred())
	})
	It("rejects blank units", func() {
		buf := buff(`{"units":"","value":42}`)
		err := request.DecodeObject(nil, buf, &Threshold{})
		Expect(err).Should(HaveOccurred())
	})
	It("is case-sensitive with respect to Units", func() {
		badUnits := strings.ToUpper(glucose.MmolL)
		buf := buff(`{"units":%q,"value":42}`, badUnits)
		err := request.DecodeObject(nil, buf, &Threshold{})
		Expect(err).Should(HaveOccurred())
	})

})

// buff is a helper for generating a JSON []byte representation.
func buff(format string, args ...interface{}) *bytes.Buffer {
	return bytes.NewBufferString(fmt.Sprintf(format, args...))
}
