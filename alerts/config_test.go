package alerts

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		raw := []byte(fmt.Sprintf(`{"units":%q,"value":42}`, UnitsMilligramsPerDeciliter))
		threshold := &Threshold{}
		err := json.Unmarshal(raw, threshold)
		Expect(err).To(BeNil())
		Expect(threshold.Value).To(Equal(42.0))
		Expect(threshold.Units).To(Equal(UnitsMilligramsPerDeciliter))
	})
	It("accepts mmol/L", func() {
		raw := []byte(fmt.Sprintf(`{"units":%q,"value":42}`, UnitsMillimollsPerLiter))
		threshold := &Threshold{}
		err := json.Unmarshal(raw, threshold)
		Expect(err).To(BeNil())
		Expect(threshold.Value).To(Equal(42.0))
		Expect(threshold.Units).To(Equal(UnitsMillimollsPerLiter))
	})
	It("rejects lb/gal", func() {
		raw := []byte(fmt.Sprintf(`{"units":%q,"value":42}`, "lb/gal"))
		threshold := &Threshold{}
		err := json.Unmarshal(raw, threshold)
		Expect(err).Should(HaveOccurred())
	})
	It("rejects blank units", func() {
		raw := []byte(fmt.Sprintf(`{"units":%q,"value":42}`, ""))
		threshold := &Threshold{}
		err := json.Unmarshal(raw, threshold)
		Expect(err).Should(HaveOccurred())
	})
	It("is case-sensitive with respect to Units", func() {
		badUnits := strings.ToUpper(UnitsMillimollsPerLiter)
		raw := []byte(fmt.Sprintf(`{"units":%q,"value":42}`, badUnits))
		threshold := &Threshold{}
		err := json.Unmarshal(raw, threshold)
		Expect(err).Should(HaveOccurred())
	})

})
