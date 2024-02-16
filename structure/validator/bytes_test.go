package validator

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/base"
)

var _ = Describe("Bytes", func() {
	Context("NewBytes", func() {
		It("returns successfully", func() {
			_, b := newTestValidator()
			Expect(NewBytes(b, nil)).ToNot(BeNil())
		})
	})

	Context("NotEmpty", func() {
		It("is invalid for a nil slice", func() {
			v, b := newTestValidator()
			NewBytes(b, nil).NotEmpty()
			Expect(v.HasError()).To(BeTrue())
		})

		It("is valid for a non-nil slice", func() {
			v, b := newTestValidator()
			NewBytes(b, []byte("test")).NotEmpty()
			Expect(v.HasError()).To(BeFalse())
		})
	})

	Context("LengthLessThanOrEqualTo", func() {
		var testToken = []byte("1234")

		It("can be up to limit", func() {
			v, b := newTestValidator()
			NewBytes(b, testToken).LengthLessThanOrEqualTo(len(testToken))
			Expect(v.HasError()).To(BeFalse())
		})

		It("can't be more than limit", func() {
			v, b := newTestValidator()
			testTooLong := append(testToken, '5')
			NewBytes(b, testTooLong).LengthLessThanOrEqualTo(len(testToken))
			Expect(v.HasError()).To(BeTrue())
		})
	})

})

func newTestValidator() (structure.Validator, *base.Base) {
	b := base.New().WithSource(structure.NewPointerSource())
	return NewValidator(b), b
}
