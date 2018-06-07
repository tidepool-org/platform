package request_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/request"
	testHTTP "github.com/tidepool-org/platform/test/http"
	testRest "github.com/tidepool-org/platform/test/rest"
)

var _ = Describe("Request", func() {
	Context("DecodeRequestPathParameter", func() {
		var req *rest.Request
		var key string
		var value string
		var validator func(value string) bool

		BeforeEach(func() {
			req = testRest.NewRequest()
			key = testHTTP.NewParameterKey()
			value = testHTTP.NewParameterValue()
			validator = func(value string) bool { return true }
			req.PathParams[key] = value
		})

		It("returns error if the request is missing", func() {
			result, err := request.DecodeRequestPathParameter(nil, key, validator)
			Expect(err).To(MatchError("request is missing"))
			Expect(result).To(BeEmpty())
		})

		It("returns error if parameter is not found", func() {
			delete(req.PathParams, key)
			result, err := request.DecodeRequestPathParameter(req, key, validator)
			errorsTest.ExpectEqual(err, request.ErrorParameterMissing(key))
			Expect(result).To(BeEmpty())
		})

		It("returns error if parameter is empty", func() {
			req.PathParams[key] = ""
			result, err := request.DecodeRequestPathParameter(req, key, validator)
			errorsTest.ExpectEqual(err, request.ErrorParameterMissing(key))
			Expect(result).To(BeEmpty())
		})

		It("returns error if validator returns false", func() {
			result, err := request.DecodeRequestPathParameter(req, key, func(value string) bool { return false })
			errorsTest.ExpectEqual(err, request.ErrorParameterInvalid(key))
			Expect(result).To(BeEmpty())
		})

		It("returns successfully if validator returns true", func() {
			result, err := request.DecodeRequestPathParameter(req, key, validator)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(value))
		})

		It("returns successfully if validator is not specified", func() {
			result, err := request.DecodeRequestPathParameter(req, key, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(value))
		})
	})
})
