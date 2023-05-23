package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/structure/parser"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
)

var _ = Describe("Unit", func() {

	Context("UnitFromParser", func() {
		It("returns the unit value when set", func() {
			objectData := map[string]interface{}{
				"unit": dataBloodGlucose.MmolL,
			}
			unit := dexcom.UnitFromParser(parser.NewObject(&objectData), dataBloodGlucose.MgdL)
			Expect(unit).ToNot(BeNil())
			Expect(*unit).To(Equal(dataBloodGlucose.MmolL))
		})
		It("returns default unit value when not set", func() {
			objectData := map[string]interface{}{
				"unit": nil,
			}
			unit := dexcom.UnitFromParser(parser.NewObject(&objectData), dataBloodGlucose.MgdL)
			Expect(*unit).To(Equal(dataBloodGlucose.MgdL))
		})
	})
})
