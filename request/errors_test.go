package request_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
	testHTTP "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Errors", func() {
	Context("ErrorUnexpectedResponse", func() {
		It("returns the expected error", func() {
			req := testHTTP.NewRequest()
			res := &http.Response{StatusCode: 405}
			err := request.ErrorUnexpectedResponse(res, req)
			Expect(err).ToNot(BeNil())
			Expect(errors.Code(err)).To(Equal("unexpected-response"))
			Expect(errors.Cause(err)).To(Equal(err))
			bytes, bytesErr := json.Marshal(errors.Sanitize(err))
			Expect(bytesErr).ToNot(HaveOccurred())
			Expect(bytes).To(MatchJSON(fmt.Sprintf(`{"code": "unexpected-response", "title": "unexpected response", "detail": "unexpected response status code %d from %s \"%s\""}`, res.StatusCode, req.Method, req.URL.String())))
		})
	})

	DescribeTable("all other errors",
		func(err error, code string, title string, detail string) {
			Expect(err).ToNot(BeNil())
			Expect(errors.Code(err)).To(Equal(code))
			Expect(errors.Cause(err)).To(Equal(err))
			bytes, bytesErr := json.Marshal(errors.Sanitize(err))
			Expect(bytesErr).ToNot(HaveOccurred())
			Expect(bytes).To(MatchJSON(fmt.Sprintf(`{"code": %q, "title": %q, "detail": %q}`, code, title, detail)))
		},
		Entry("is ErrorTooManyRequests", request.ErrorTooManyRequests(), "too-many-requests", "too many requests", "too many requests"),
		Entry("is ErrorBadRequest", request.ErrorBadRequest(), "bad-request", "bad request", "bad request"),
		Entry("is ErrorUnauthenticated", request.ErrorUnauthenticated(), "unauthenticated", "authentication token is invalid", "authentication token is invalid"),
		Entry("is ErrorUnauthorized", request.ErrorUnauthorized(), "unauthorized", "authentication token is not authorized for requested action", "authentication token is not authorized for requested action"),
		Entry("is ErrorResourceNotFound", request.ErrorResourceNotFound(), "resource-not-found", "resource not found", "resource not found"),
		Entry("is ErrorResourceNotFoundWithID", request.ErrorResourceNotFoundWithID("test-id"), "resource-not-found", "resource not found", `resource with id "test-id" not found`),
		Entry("is ErrorParameterMissing", request.ErrorParameterMissing("test_parameter"), "parameter-missing", "parameter is missing", `parameter "test_parameter" is missing`),
		Entry("is ErrorJSONMalformed", request.ErrorJSONMalformed(), "json-malformed", "json is malformed", "json is malformed"),
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
			Entry("is another request error", request.ErrorJSONMalformed(), 500),
			Entry("is another error", errors.New("test-error"), 500),
			Entry("is nil error", nil, 500),
		)
	})
})
