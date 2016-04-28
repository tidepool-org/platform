package validate_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("ErrorsArray", func() {

	It("returns false is it has no errors", func() {
		errs := validate.NewErrorProcessing("0")
		Expect(errs.HasErrors()).To(BeFalse())
	})

	It("returns true is it has errors", func() {
		errs := validate.NewErrorProcessing("0")
		errs.AppendPointerError("deviceTags/0", "Device tag is unknown.", "Device tags values must be one of cgm, ...")
		Expect(errs.HasErrors()).To(BeTrue())
	})

	It("when in JSON format is readable and of use to clients", func() {
		errs := validate.NewErrorProcessing("0")

		errs.AppendPointerError("deviceTags/0", "Device tag is unknown.", "Device tags values must be one of cgm, ...")
		errs.AppendPointerError("type", "Type is unknown.", "Type must be one of basal, bolus, ...")
		Expect(errs.HasErrors()).To(BeTrue())
		bytes, err := json.Marshal(errs.GetErrors())
		Expect(err).To(Succeed())
		Expect(string(bytes)).To(MatchJSON(`[{"source":{"pointer":"0/deviceTags/0"},"title":"Device tag is unknown.","detail":"Device tags values must be one of cgm, ..."},{"source":{"pointer":"0/type"},"title":"Type is unknown.","detail":"Type must be one of basal, bolus, ..."}]`))
	})
})
