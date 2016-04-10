package ketone

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Blood", func() {
	var bloodKetoneObj = fixtures.TestingDatumBase()
	var processing validate.ErrorProcessing

	Context("ketone from obj", func() {

		BeforeEach(func() {
			processing = validate.ErrorProcessing{BasePath: "0", ErrorsArray: validate.NewErrorsArray()}
			bloodKetoneObj["type"] = "bloodKetone"
			bloodKetoneObj["value"] = 2.2
			bloodKetoneObj["units"] = "mmol/L"

		})

		It("when valid", func() {
			bloodKetone := Build(bloodKetoneObj, processing)
			var recordType *Blood
			Expect(bloodKetone).To(BeAssignableToTypeOf(recordType))
			Expect(processing.HasErrors()).To(BeFalse())
		})

		Context("validation", func() {
			Context("value", func() {
				It("fails greater than zero", func() {
					bloodKetoneObj["value"] = 0.0
					bloodKetone := Build(bloodKetoneObj, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(bloodKetone).To(Not(BeNil()))
				})

			})
			Context("iunits", func() {
				It("fails if not mmol/L", func() {
					bloodKetoneObj["units"] = "mg/dL"
					bloodKetone := Build(bloodKetoneObj, processing)
					Expect(processing.HasErrors()).To(BeTrue())
					Expect(bloodKetone).To(Not(BeNil()))
				})

			})

		})
	})
})
