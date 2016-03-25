package validate_test

import (
	"encoding/json"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("ErrorsArray", func() {

	It("returns false is it has no errors", func() {
		errs := validate.NewErrorsArray()
		Expect(errs.HasErrors()).To(BeFalse())
	})
	It("returns true is it has errors", func() {
		errs := validate.NewErrorsArray()
		errs.Append(validate.NewPointerError("16/deviceTags/0", "Device tag is unknown.", "Device tags values must be one of cgm, ..."))
		Expect(errs.HasErrors()).To(BeTrue())
	})
	It("when in JSON format is readable and of use to clients", func() {
		errs := validate.NewErrorsArray()

		errs.Append(validate.NewPointerError("16/deviceTags/0", "Device tag is unknown.", "Device tags values must be one of cgm, ..."))
		errs.Append(validate.NewPointerError("2/type", "Type is unknown.", "Type must be one of basal, bolus, ..."))
		Expect(errs.HasErrors()).To(BeTrue())
		bytes, _ := json.Marshal(errs)

		Expect(string(bytes)).To(Equal(`{"errors":[{"source":{"pointer":"16/deviceTags/0"},"title":"Device tag is unknown.","detail":"Device tags values must be one of cgm, ..."},{"source":{"pointer":"2/type"},"title":"Type is unknown.","detail":"Type must be one of basal, bolus, ..."}]}`))

	})
})
