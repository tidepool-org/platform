package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/structure/parser"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
)

var _ = Describe("Default", func() {

	Context("StringOrDefault", func() {

		var objParser *parser.Object

		BeforeEach(func() {
			objectData := map[string]interface{}{
				"unit": dataBloodGlucose.MmolL,
			}
			objParser = parser.NewObject(&objectData)
		})

		It("returns the unit value when set", func() {
			unit := dexcom.StringOrDefault(objParser.String("unit"), dataBloodGlucose.MgdL)
			Expect(unit).ToNot(BeNil())
			Expect(*unit).To(Equal(dataBloodGlucose.MmolL))
		})
		It("returns default unit value when not set", func() {
			unit := dexcom.StringOrDefault(objParser.String("no-unit"), dataBloodGlucose.MgdL)
			Expect(unit).ToNot(BeNil())
			Expect(*unit).To(Equal(dataBloodGlucose.MgdL))
		})
		It("returns nil when neither set", func() {
			unit := dexcom.StringOrDefault(objParser.String("no-unit"), "")
			Expect(unit).To(BeNil())
		})
	})
})
