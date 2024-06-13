package dexcom_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/dexcom"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/structure/parser"
)

var _ = Describe("Default", func() {

	Context("ParseStringOrDefault", func() {
		var objectParser *parser.Object

		BeforeEach(func() {
			object := map[string]interface{}{
				"unit":  dataBloodGlucose.MmolL,
				"empty": "",
			}
			objectParser = parser.NewObject(logTest.NewLogger(), &object)
		})

		It("return the value when not missing nor empty", func() {
			unit := dexcom.ParseStringOrDefault(objectParser, "unit", dataBloodGlucose.MgdL)
			Expect(unit).ToNot(BeNil())
			Expect(*unit).To(Equal(dataBloodGlucose.MmolL))
		})

		It("returns the default when value is missing", func() {
			unit := dexcom.ParseStringOrDefault(objectParser, "missing", dataBloodGlucose.MgdL)
			Expect(unit).ToNot(BeNil())
			Expect(*unit).To(Equal(dataBloodGlucose.MgdL))
		})

		It("returns the default when value is empty", func() {
			val := dexcom.ParseStringOrDefault(objectParser, "empty", dexcom.EventUnitCarbsGrams)
			Expect(val).ToNot(BeNil())
			Expect(*val).To(Equal(dexcom.EventUnitCarbsGrams))
		})
	})
})
