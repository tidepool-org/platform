package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"net/http"
	"net/url"

	"github.com/tidepool-org/platform/client"
)

var _ = Describe("Errors", func() {
	Context("NewUnauthorizedError", func() {
		It("returns not nil", func() {
			Expect(client.NewUnauthorizedError()).ToNot(BeNil())
		})
	})

	Context("with new unauthorized error", func() {
		var unauthorizedError *client.UnauthorizedError

		BeforeEach(func() {
			unauthorizedError = client.NewUnauthorizedError()
			Expect(unauthorizedError).ToNot(BeNil())
		})

		It("has the expected error", func() {
			Expect(unauthorizedError.Error()).To(Equal("client: unauthorized"))
		})
	})

	Context("IsUnauthorizedError", func() {
		It("returns true for an UnauthorizedError", func() {
			Expect(client.IsUnauthorizedError(client.NewUnauthorizedError())).To(BeTrue())
		})

		It("returns false for any other type of error", func() {
			Expect(client.IsUnauthorizedError(errors.New("other"))).To(BeFalse())
		})
	})

	Context("NewUnexpectedResponseError", func() {
		It("returns not nil", func() {
			var err error
			url, err := url.Parse("http://localhost:1234")
			Expect(err).ToNot(HaveOccurred())
			Expect(url).ToNot(BeNil())
			request := &http.Request{Method: "GET", URL: url}
			response := &http.Response{StatusCode: 400}
			Expect(client.NewUnexpectedResponseError(response, request)).ToNot(BeNil())
		})
	})

	Context("with new unexpected response error", func() {
		var unexpectedResponseError *client.UnexpectedResponseError

		BeforeEach(func() {
			var err error
			url, err := url.Parse("http://localhost:1234")
			Expect(err).ToNot(HaveOccurred())
			Expect(url).ToNot(BeNil())
			request := &http.Request{Method: "GET", URL: url}
			response := &http.Response{StatusCode: 400}
			unexpectedResponseError = client.NewUnexpectedResponseError(response, request)
			Expect(unexpectedResponseError).ToNot(BeNil())
		})

		It("has the expected method", func() {
			Expect(unexpectedResponseError.Method).To(Equal("GET"))
		})

		It("has the expected url", func() {
			Expect(unexpectedResponseError.URL).To(Equal("http://localhost:1234"))
		})

		It("has the expected status code", func() {
			Expect(unexpectedResponseError.StatusCode).To(Equal(400))
		})

		It("has the expected error", func() {
			Expect(unexpectedResponseError.Error()).To(Equal("client: unexpected response status code 400 from GET http://localhost:1234"))
		})
	})

	Context("IsUnexpectedResponseError", func() {
		var request *http.Request
		var response *http.Response

		BeforeEach(func() {
			var err error
			url, err := url.Parse("http://localhost:1234")
			Expect(err).ToNot(HaveOccurred())
			Expect(url).ToNot(BeNil())
			request = &http.Request{Method: "GET", URL: url}
			response = &http.Response{StatusCode: 400}
		})

		It("returns true for an UnexpectedResponseError", func() {
			Expect(client.IsUnexpectedResponseError(client.NewUnexpectedResponseError(response, request))).To(BeTrue())
		})

		It("returns false for any other type of error", func() {
			Expect(client.IsUnexpectedResponseError(errors.New("other"))).To(BeFalse())
		})
	})
})
