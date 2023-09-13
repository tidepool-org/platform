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
