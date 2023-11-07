package request_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Errors", func() {
	It("ErrorCodeInternalServerError is expected", func() {
		Expect(request.ErrorCodeInternalServerError).To(Equal("internal-server-error"))
	})

	It("ErrorCodeUnexpectedResponse is expected", func() {
		Expect(request.ErrorCodeUnexpectedResponse).To(Equal("unexpected-response"))
	})

	It("ErrorCodeTooManyRequests is expected", func() {
		Expect(request.ErrorCodeTooManyRequests).To(Equal("too-many-requests"))
	})

	It("ErrorCodeBadRequest is expected", func() {
		Expect(request.ErrorCodeBadRequest).To(Equal("bad-request"))
	})

	It("ErrorCodeUnauthenticated is expected", func() {
		Expect(request.ErrorCodeUnauthenticated).To(Equal("unauthenticated"))
	})

	It("ErrorCodeUnauthorized is expected", func() {
		Expect(request.ErrorCodeUnauthorized).To(Equal("unauthorized"))
	})

	It("ErrorCodeResourceNotFound is expected", func() {
		Expect(request.ErrorCodeResourceNotFound).To(Equal("resource-not-found"))
	})

	It("ErrorCodeResourceTooLarge is expected", func() {
		Expect(request.ErrorCodeResourceTooLarge).To(Equal("resource-too-large"))
	})

	It("ErrorCodeHeaderMissing is expected", func() {
		Expect(request.ErrorCodeHeaderMissing).To(Equal("header-missing"))
	})

	It("ErrorCodeHeaderInvalid is expected", func() {
		Expect(request.ErrorCodeHeaderInvalid).To(Equal("header-invalid"))
	})

	It("ErrorCodeParameterMissing is expected", func() {
		Expect(request.ErrorCodeParameterMissing).To(Equal("parameter-missing"))
	})

	It("ErrorCodeParameterInvalid is expected", func() {
		Expect(request.ErrorCodeParameterInvalid).To(Equal("parameter-invalid"))
	})

	It("ErrorCodeJSONMalformed is expected", func() {
		Expect(request.ErrorCodeJSONMalformed).To(Equal("json-malformed"))
	})

	It("ErrorCodeDigestsNotEqual is expected", func() {
		Expect(request.ErrorCodeDigestsNotEqual).To(Equal("digests-not-equal"))
	})

	It("ErrorCodeMediaTypeNotSupported is expected", func() {
		Expect(request.ErrorCodeMediaTypeNotSupported).To(Equal("media-type-not-supported"))
	})

	Context("ErrorInternalServerError", func() {
		It("returns the expected error", func() {
			cause := errors.New("error")
			err := request.ErrorInternalServerError(cause)
			Expect(err).ToNot(BeNil())
			Expect(errors.Code(err)).To(Equal("internal-server-error"))
			Expect(errors.Cause(err)).To(Equal(cause))
			bites, marshalErr := json.Marshal(errors.Sanitize(err))
			Expect(marshalErr).ToNot(HaveOccurred())
			Expect(bites).To(MatchJSON(`{"code": "internal-server-error", "title": "internal server error", "detail": "internal server error"}`))
		})
	})

	Context("ErrorUnexpectedResponse", func() {
		It("returns the expected error", func() {
			req := testHttp.NewRequest()
			res := &http.Response{StatusCode: 405}
			err := request.ErrorUnexpectedResponse(res, req)
			Expect(err).ToNot(BeNil())
			Expect(errors.Code(err)).To(Equal("unexpected-response"))
			Expect(errors.Cause(err)).To(Equal(err))
			bites, marshalErr := json.Marshal(errors.Sanitize(err))
			Expect(marshalErr).ToNot(HaveOccurred())
			Expect(bites).To(MatchJSON(fmt.Sprintf(`{"code": "unexpected-response", "title": "unexpected response", "detail": "unexpected response status code %d from %s \"%s\""}`, res.StatusCode, req.Method, req.URL.String())))
		})
	})

	DescribeTable("have expected details when error",
		errorsTest.ExpectErrorDetails,
		Entry("is ErrorTooManyRequests", request.ErrorTooManyRequests(), "too-many-requests", "too many requests", "too many requests"),
		Entry("is ErrorBadRequest", request.ErrorBadRequest(), "bad-request", "bad request", "bad request"),
		Entry("is ErrorUnauthenticated", request.ErrorUnauthenticated(), "unauthenticated", "authentication token is invalid", "authentication token is invalid"),
		Entry("is ErrorUnauthorized", request.ErrorUnauthorized(), "unauthorized", "authentication token is not authorized for requested action", "authentication token is not authorized for requested action"),
		Entry("is ErrorResourceNotFound", request.ErrorResourceNotFound(), "resource-not-found", "resource not found", "resource not found"),
		Entry("is ErrorResourceNotFoundWithID", request.ErrorResourceNotFoundWithID("test-id"), "resource-not-found", "resource not found", `resource with id "test-id" not found`),
		Entry("is ErrorResourceNotFoundWithIDAndRevision", request.ErrorResourceNotFoundWithIDAndRevision("test-id", 1), "resource-not-found", "resource not found", `revision 1 of resource with id "test-id" not found`),
		Entry("is ErrorResourceNotFoundWithIDAndOptionalRevision", request.ErrorResourceNotFoundWithIDAndOptionalRevision("test-id", nil), "resource-not-found", "resource not found", `resource with id "test-id" not found`),
		Entry("is ErrorResourceNotFoundWithIDAndOptionalRevision", request.ErrorResourceNotFoundWithIDAndOptionalRevision("test-id", pointer.FromInt(1)), "resource-not-found", "resource not found", `revision 1 of resource with id "test-id" not found`),
		Entry("is ErrorResourceTooLarge", request.ErrorResourceTooLarge(), "resource-too-large", "resource too large", "resource too large"),
		Entry("is ErrorHeaderMissing", request.ErrorHeaderMissing("X-Test-Header"), "header-missing", "header is missing", `header "X-Test-Header" is missing`),
		Entry("is ErrorHeaderInvalid", request.ErrorHeaderInvalid("X-Test-Header"), "header-invalid", "header is invalid", `header "X-Test-Header" is invalid`),
		Entry("is ErrorParameterMissing", request.ErrorParameterMissing("test_parameter"), "parameter-missing", "parameter is missing", `parameter "test_parameter" is missing`),
		Entry("is ErrorParameterInvalid", request.ErrorParameterInvalid("test_parameter"), "parameter-invalid", "parameter is invalid", `parameter "test_parameter" is invalid`),
		Entry("is ErrorJSONNotFound", request.ErrorJSONNotFound(), "json-not-found", "json not found", "json not found"),
		Entry("is ErrorJSONMalformed", request.ErrorJSONMalformed(), "json-malformed", "json is malformed", "json is malformed"),
		Entry("is ErrorDigestsNotEqual", request.ErrorDigestsNotEqual("QUJDREVGSElKS0xNTk9QUQ==", "lah2klptWl+IBNSepXlJ9Q=="), "digests-not-equal", "digests not equal", `digest "QUJDREVGSElKS0xNTk9QUQ==" does not equal calculated digest "lah2klptWl+IBNSepXlJ9Q=="`),
		Entry("is ErrorMediaTypeNotSupported", request.ErrorMediaTypeNotSupported("application/octet-stream"), "media-type-not-supported", "media type not supported", `media type "application/octet-stream" not supported`),
		Entry("is ErrorExtensionNotSupported", request.ErrorExtensionNotSupported("bin"), "extension-not-supported", "extension not supported", `extension "bin" not supported`),
	)

	Context("StatusCodeForError", func() {
		DescribeTable("returns expected value when",
			func(err error, expectedStatusCode int) {
				Expect(request.StatusCodeForError(err)).To(Equal(expectedStatusCode))
			},
			Entry("is ErrorTooManyRequests", request.ErrorTooManyRequests(), 429),
			Entry("is ErrorBadRequest", request.ErrorBadRequest(), 400),
			Entry("is ErrorUnauthenticated", request.ErrorUnauthenticated(), 401),
			Entry("is ErrorUnauthorized", request.ErrorUnauthorized(), 403),
			Entry("is ErrorResourceNotFound", request.ErrorResourceNotFound(), 404),
			Entry("is ErrorResourceNotFoundWithID", request.ErrorResourceNotFoundWithID("test-id"), 404),
			Entry("is ErrorResourceNotFoundWithIDAndRevision", request.ErrorResourceNotFoundWithIDAndRevision("test-id", 1), 404),
			Entry("is ErrorResourceNotFoundWithIDAndOptionalRevision", request.ErrorResourceNotFoundWithIDAndOptionalRevision("test-id", nil), 404),
			Entry("is ErrorResourceNotFoundWithIDAndOptionalRevision", request.ErrorResourceNotFoundWithIDAndOptionalRevision("test-id", pointer.FromInt(1)), 404),
			Entry("is ErrorResourceTooLarge", request.ErrorResourceTooLarge(), 413),
			Entry("is another request error", request.ErrorJSONMalformed(), 500),
			Entry("is another error", errors.New("test-error"), 500),
			Entry("is nil error", nil, 500),
		)
	})

	Context("IsErrorInternalServerError", func() {
		It("returns false if the error does not have a code", func() {
			Expect(request.IsErrorInternalServerError(errors.New("error"))).To(BeFalse())
		})

		It("returns false if the error code is not ErrorCodeInternalServerError", func() {
			Expect(request.IsErrorInternalServerError(request.ErrorUnauthenticated())).To(BeFalse())
		})

		It("returns true if the error code is ErrorCodeInternalServerError", func() {
			Expect(request.IsErrorInternalServerError(request.ErrorInternalServerError(errors.New("error")))).To(BeTrue())
		})
	})

	Context("IsErrorUnauthenticated", func() {
		It("returns false if the error does not have a code", func() {
			Expect(request.IsErrorUnauthenticated(errors.New("error"))).To(BeFalse())
		})

		It("returns false if the error code is not ErrorCodeUnauthenticated", func() {
			Expect(request.IsErrorUnauthenticated(request.ErrorInternalServerError(errors.New("error")))).To(BeFalse())
		})

		It("returns true if the error code is ErrorCodeUnauthenticated", func() {
			Expect(request.IsErrorUnauthenticated(request.ErrorUnauthenticated())).To(BeTrue())
		})
	})

	Context("IsErrorUnauthorized", func() {
		It("returns false if the error does not have a code", func() {
			Expect(request.IsErrorUnauthorized(errors.New("error"))).To(BeFalse())
		})

		It("returns false if the error code is not ErrorCodeUnauthorized", func() {
			Expect(request.IsErrorUnauthorized(request.ErrorInternalServerError(errors.New("error")))).To(BeFalse())
		})

		It("returns true if the error code is ErrorCodeUnauthorized", func() {
			Expect(request.IsErrorUnauthorized(request.ErrorUnauthorized())).To(BeTrue())
		})
	})

	Context("IsErrorResourceNotFound", func() {
		It("returns false if the error does not have a code", func() {
			Expect(request.IsErrorResourceNotFound(errors.New("error"))).To(BeFalse())
		})

		It("returns false if the error code is not ErrorCodeResourceNotFound", func() {
			Expect(request.IsErrorResourceNotFound(request.ErrorInternalServerError(errors.New("error")))).To(BeFalse())
		})

		It("returns true if the error code is ErrorCodeResourceNotFound", func() {
			Expect(request.IsErrorResourceNotFound(request.ErrorResourceNotFound())).To(BeTrue())
		})
	})
})
