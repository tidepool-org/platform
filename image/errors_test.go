package image_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/image"
)

var _ = Describe("Errors", func() {
	It("ErrorCodeImageContentIntentUnexpected is expected", func() {
		Expect(image.ErrorCodeImageContentIntentUnexpected).To(Equal("image-content-intent-unexpected"))
	})

	It("ErrorCodeImageMalformed is expected", func() {
		Expect(image.ErrorCodeImageMalformed).To(Equal("image-malformed"))
	})

	DescribeTable("have expected details when error",
		errorsTest.ExpectErrorDetails,
		Entry("is ErrorValueStringAsIDNotValid", image.ErrorValueStringAsIDNotValid("0123456789abcdefghijklmnopqrstuv"), "value-not-valid", "value is not valid", `value "0123456789abcdefghijklmnopqrstuv" is not valid as image id`),
		Entry("is ErrorValueStringAsContentIDNotValid", image.ErrorValueStringAsContentIDNotValid("0123456789abcdefghijklmnopqrstuv"), "value-not-valid", "value is not valid", `value "0123456789abcdefghijklmnopqrstuv" is not valid as image content id`),
		Entry("is ErrorValueStringAsRenditionsIDNotValid", image.ErrorValueStringAsRenditionsIDNotValid("0123456789abcdefghijklmnopqrstuv"), "value-not-valid", "value is not valid", `value "0123456789abcdefghijklmnopqrstuv" is not valid as image renditions id`),
		Entry("is ErrorValueStringAsContentIntentNotValid", image.ErrorValueStringAsContentIntentNotValid("invalid"), "value-not-valid", "value is not valid", `value "invalid" is not valid as image content intent`),
		Entry("is ErrorImageContentIntentUnexpected", image.ErrorImageContentIntentUnexpected("original"), "image-content-intent-unexpected", "image content intent unexpected", `image content intent "original" unexpected`),
		Entry("is ErrorValueRenditionNotParsable", image.ErrorValueRenditionNotParsable("invalid"), "value-not-parsable", "value is not a parsable rendition", `value "invalid" is not a parsable rendition`),
		Entry("is ErrorValueStringAsColorNotValid", image.ErrorValueStringAsColorNotValid("invalid"), "value-not-valid", "value is not valid", `value "invalid" is not valid as color`),
		Entry("is ErrorImageMalformed", image.ErrorImageMalformed("invalid"), "image-malformed", "image is malformed", "invalid"),
	)
})
