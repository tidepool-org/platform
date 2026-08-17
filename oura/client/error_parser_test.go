package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	ouraClient "github.com/tidepool-org/platform/oura/client"
	ouraClientTest "github.com/tidepool-org/platform/oura/client/test"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("error_parser", func() {
	var (
		logger              *logTest.Logger
		ctx                 context.Context
		req                 *http.Request
		res                 *http.Response
		errorResponseParser client.ErrorResponseParser
	)

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
		req = &http.Request{}
		res = &http.Response{}
		errorResponseParser = &ouraClient.ErrorResponseParser{}
		Expect(errorResponseParser).ToNot(BeNil())
	})

	It("returns nil for status code not http.StatusUnprocessableEntity", func() {
		statusCodes := slices.DeleteFunc(testHttp.StatusCodes, func(code int) bool { return code == http.StatusUnprocessableEntity })
		for _, statusCode := range statusCodes {
			res.StatusCode = statusCode
			Expect(errorResponseParser.ParseErrorResponse(ctx, res, req)).To(BeNil())
		}
	})

	Context("with status code http.StatusUnprocessableEntity", func() {
		BeforeEach(func() {
			res.StatusCode = http.StatusUnprocessableEntity
		})

		It("returns ErrorBadRequest and logs an error if the response body is too large", func() {
			res.Body = io.NopCloser(bytes.NewReader(test.RandomBytesFromRange(ouraClient.ErrorResponseBodyLimit+1, ouraClient.ErrorResponseBodyLimit+1)))
			err := errorResponseParser.ParseErrorResponse(ctx, res, req)
			errorsTest.ExpectEqual(err, request.ErrorBadRequest())
			logger.AssertError("unable to read error response body")
		})

		It("returns ErrorBadRequest and logs an error if the response body is not decodable", func() {
			res.Body = io.NopCloser(bytes.NewReader(test.RandomBytesFromRange(ouraClient.ErrorResponseBodyLimit, ouraClient.ErrorResponseBodyLimit)))
			err := errorResponseParser.ParseErrorResponse(ctx, res, req)
			errorsTest.ExpectEqual(err, request.ErrorBadRequest())
			logger.AssertError("unable to decode error response body")
		})

		It("returns ErrorBadRequest with meta if the response body is decodable", func() {
			errorResponse := ouraClientTest.RandomErrorResponse()
			bites, err := json.Marshal(errorResponse)
			Expect(err).ToNot(HaveOccurred())
			Expect(bites).ToNot(BeNil())
			res.Body = io.NopCloser(bytes.NewReader(bites))
			err = errorResponseParser.ParseErrorResponse(ctx, res, req)
			errorsTest.ExpectEqual(err, errors.WithMeta(request.ErrorBadRequest(), errorResponse))
		})
	})
})
