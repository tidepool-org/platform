package middleware_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"strings"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service/middleware"
)

var _ = Describe("Header", func() {
	Context("NewRawFieldFunc", func() {
		var fieldFunc middleware.FieldFunc
		var testFields log.Fields

		BeforeEach(func() {
			fieldFunc = middleware.NewRawFieldFunc("new")
			testFields = log.Fields{}
			testFields["existing"] = "value-for-existing"
		})

		It("the function does not add an empty string", func() {
			Expect(fieldFunc(testFields, "")).To(Equal(log.Fields{"existing": "value-for-existing"}))
		})

		It("the function adds a string", func() {
			Expect(fieldFunc(testFields, "value-for-new")).To(Equal(log.Fields{"existing": "value-for-existing", "new": "value-for-new"}))
		})

		It("the function adds a string truncated to 256 characters", func() {
			longString := strings.Repeat("1234567890", 100)
			Expect(fieldFunc(testFields, longString)).To(Equal(log.Fields{"existing": "value-for-existing", "new": longString[:256]}))
		})
	})

	Context("NewMD5FieldFunc", func() {
		var fieldFunc middleware.FieldFunc
		var testFields log.Fields

		BeforeEach(func() {
			fieldFunc = middleware.NewMD5FieldFunc("new")
			testFields = log.Fields{}
			testFields["existing"] = "value-for-existing"
		})

		It("the function does not add an empty string", func() {
			Expect(fieldFunc(testFields, "")).To(Equal(log.Fields{"existing": "value-for-existing"}))
		})

		It("the function adds a string hashed using MD5", func() {
			Expect(fieldFunc(testFields, "value-for-new")).To(Equal(log.Fields{"existing": "value-for-existing", "new": "1823b9b9b0d449ca00fe70e37436b8a0"}))
		})
	})

	Context("NewHeader", func() {
		It("returns successfully", func() {
			headerMiddleware, err := middleware.NewHeader()
			Expect(err).ToNot(HaveOccurred())
			Expect(headerMiddleware).ToNot(BeNil())
			Expect(headerMiddleware.HeaderFieldFuncs).ToNot(BeNil())
		})
	})

	Context("with header middleware", func() {
		var headerMiddleware *middleware.Header

		BeforeEach(func() {
			var err error
			headerMiddleware, err = middleware.NewHeader()
			Expect(err).ToNot(HaveOccurred())
			Expect(headerMiddleware).ToNot(BeNil())
		})

		Context("AddHeaderFieldFunc", func() {
			It("adds the header field func", func() {
				existingFunc := middleware.NewRawFieldFunc("existing")
				headerMiddleware.HeaderFieldFuncs["existing"] = existingFunc
				newFunc := middleware.NewRawFieldFunc("new")
				headerMiddleware.AddHeaderFieldFunc("new", newFunc)
				Expect(headerMiddleware.HeaderFieldFuncs).To(HaveKey("existing"))
				Expect(headerMiddleware.HeaderFieldFuncs).To(HaveKey("new"))
			})
		})
	})
})
