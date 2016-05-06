package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types"
)

var _ = Describe("Bloodglucose Validation", func() {

	mmolL := "mmol/L"
	mgdL := "mg/dL"
	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})

	Context("convert value", func() {

		It("creates error if value is nil", func() {

			bgValidator := types.NewBloodGlucoseValidation(nil, &mmolL)

			bgValidator.ValidateAndConvertBloodGlucoseValue(helper.ErrorProcessing)

			Expect(helper.ErrorProcessing.HasErrors()).To(BeTrue())
			Expect(helper.ErrorProcessing.GetErrors()).To(HaveLen(1))
			Expect(helper.ErrorProcessing.GetErrors()[0].Detail).To(Equal("Must be between 0.0 and 55.0 given '<nil>'"))
		})

		It("creates error if units are nil", func() {

			fiveFive := 5.5
			bgValidator := types.NewBloodGlucoseValidation(&fiveFive, nil)

			bgValidator.ValidateAndConvertBloodGlucoseValue(helper.ErrorProcessing)

			Expect(helper.ErrorProcessing.HasErrors()).To(BeTrue())
			Expect(helper.ErrorProcessing.GetErrors()).To(HaveLen(1))
			Expect(helper.ErrorProcessing.GetErrors()[0].Detail).To(Equal("Must be one of mmol/L, mg/dL given '<nil>'"))
		})

		It("returns same value if already mmol/L", func() {
			fiveFive := 5.5

			convertedBg, convertedUnits := types.NewBloodGlucoseValidation(&fiveFive, &mmolL).
				ValidateAndConvertBloodGlucoseValue(helper.ErrorProcessing)

			Expect(convertedBg).To(Equal(&fiveFive))
			Expect(convertedUnits).To(Equal(&mmolL))
			Expect(helper.ErrorProcessing.HasErrors()).To(BeFalse())
		})

		It("creates error if outside of the expected range for mmol/L", func() {
			fiftyFiveFive := 55.5

			types.NewBloodGlucoseValidation(&fiftyFiveFive, &mmolL).
				ValidateAndConvertBloodGlucoseValue(helper.ErrorProcessing)

			Expect(helper.ErrorProcessing.HasErrors()).To(BeTrue())
			Expect(helper.ErrorProcessing.GetErrors()).To(HaveLen(1))
			Expect(helper.ErrorProcessing.GetErrors()[0].Detail).To(Equal("Must be between 0.0 and 55.0 given '55.5'"))
		})

		It("allows for the value to be optional", func() {

			convertedBg, convertedUnits := types.NewBloodGlucoseValidation(nil, &mmolL).
				SetValueAllowedToBeEmpty(true).
				ValidateAndConvertBloodGlucoseValue(helper.ErrorProcessing)

			Expect(helper.ErrorProcessing.HasErrors()).To(BeFalse())
			Expect(convertedUnits).To(Equal(&mmolL))
			Expect(convertedBg).To(BeNil())
		})

		It("returns value in mmol/L if mg/dL", func() {
			threeSixty := 360.0
			expected := threeSixty / 18.01559

			convertedBg, convertedUnits := types.NewBloodGlucoseValidation(&threeSixty, &mgdL).
				ValidateAndConvertBloodGlucoseValue(helper.ErrorProcessing)

			Expect(convertedBg).To(Equal(&expected))
			Expect(convertedUnits).To(Equal(&mmolL))
			Expect(helper.ErrorProcessing.HasErrors()).To(BeFalse())
		})

		It("creates error if outside of the expected range for mg/dL", func() {
			oneThousandAndOne := 1001.0

			types.NewBloodGlucoseValidation(&oneThousandAndOne, &mgdL).
				ValidateAndConvertBloodGlucoseValue(helper.ErrorProcessing)

			Expect(helper.ErrorProcessing.HasErrors()).To(BeTrue())
			Expect(helper.ErrorProcessing.GetErrors()).To(HaveLen(1))
			Expect(helper.ErrorProcessing.GetErrors()[0].Detail).To(Equal("Must be between 0.0 and 1000.0 given '1001'"))
		})

	})
})
