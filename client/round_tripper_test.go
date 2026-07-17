package client_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/client"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("RoundTripper", func() {
	var testRoundTripper *testHttp.RoundTripper
	var roundTripper *client.RoundTripper

	BeforeEach(func() {
		testRoundTripper = testHttp.NewRoundTripper()
	})

	Context("NewRoundTripper", func() {
		It("returns successfully with a round tripper", func() {
			roundTripper = client.NewRoundTripper(testRoundTripper)
			Expect(roundTripper).ToNot(BeNil())
			Expect(roundTripper.ResolvedRoundTripper()).To(BeIdenticalTo(http.RoundTripper(testRoundTripper)))
		})

		It("returns successfully without a round tripper", func() {
			roundTripper = client.NewRoundTripper(nil)
			Expect(roundTripper).ToNot(BeNil())
			Expect(roundTripper.ResolvedRoundTripper()).ToNot(BeNil())
		})
	})

	Context("ResolvedRoundTripper", func() {
		var originalDefaultTransport http.RoundTripper
		var originalDefaultClientTransport http.RoundTripper

		BeforeEach(func() {
			originalDefaultTransport = http.DefaultTransport
			originalDefaultClientTransport = http.DefaultClient.Transport
		})

		AfterEach(func() {
			http.DefaultClient.Transport = originalDefaultClientTransport
			http.DefaultTransport = originalDefaultTransport
		})

		It("returns the round tripper if it is set", func() {
			roundTripper = client.NewRoundTripper(testRoundTripper)
			Expect(roundTripper.ResolvedRoundTripper()).To(BeIdenticalTo(http.RoundTripper(testRoundTripper)))
		})

		It("returns http.DefaultClient.Transport if the round tripper is not set and http.DefaultClient.Transport is set", func() {
			http.DefaultClient.Transport = testRoundTripper
			roundTripper = client.NewRoundTripper(nil)
			Expect(roundTripper.ResolvedRoundTripper()).To(BeIdenticalTo(http.RoundTripper(testRoundTripper)))
		})

		It("returns http.DefaultTransport if the round tripper is not set and http.DefaultClient.Transport is not set", func() {
			http.DefaultClient.Transport = nil
			http.DefaultTransport = testRoundTripper
			roundTripper = client.NewRoundTripper(nil)
			Expect(roundTripper.ResolvedRoundTripper()).To(BeIdenticalTo(http.RoundTripper(testRoundTripper)))
		})
	})

	Context("WithRoundTripper", func() {
		BeforeEach(func() {
			roundTripper = client.NewRoundTripper(nil)
		})

		It("sets the round tripper when it was previously unset", func() {
			roundTripper.WithRoundTripper(testRoundTripper)
			Expect(roundTripper.ResolvedRoundTripper()).To(BeIdenticalTo(http.RoundTripper(testRoundTripper)))
		})

		It("replaces the round tripper when it was previously set", func() {
			roundTripper.WithRoundTripper(testRoundTripper)
			replacement := testHttp.NewRoundTripper()
			roundTripper.WithRoundTripper(replacement)
			Expect(roundTripper.ResolvedRoundTripper()).To(BeIdenticalTo(http.RoundTripper(replacement)))
		})
	})

	Context("RoundTrip", func() {
		var request *http.Request

		BeforeEach(func() {
			request = testHttp.NewRequest()
		})

		It("returns the response from the resolved round tripper", func() {
			testRoundTripper.Response = &http.Response{StatusCode: testHttp.NewStatusCode()}
			roundTripper = client.NewRoundTripper(testRoundTripper)

			result, err := roundTripper.RoundTrip(request)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeIdenticalTo(testRoundTripper.Response))
			Expect(testRoundTripper.Request).To(BeIdenticalTo(request))
		})

		It("returns the error from the resolved round tripper", func() {
			testRoundTripper.Error = errorsTest.RandomError()
			roundTripper = client.NewRoundTripper(testRoundTripper)

			result, err := roundTripper.RoundTrip(request)
			Expect(err).To(Equal(testRoundTripper.Error))
			Expect(result).To(BeNil())
		})
	})
})
