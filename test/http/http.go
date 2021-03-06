package http

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/test"
)

const (
	CharsetPath  = test.CharsetAlphaNumeric + "_"
	CharsetName  = test.CharsetAlphaNumeric + "_-"
	CharsetValue = test.CharsetAlphaNumeric + "_-"
)

var (
	methods = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	}
	statusCodes = []int{
		http.StatusContinue,
		http.StatusSwitchingProtocols,
		http.StatusProcessing,
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNonAuthoritativeInfo,
		http.StatusNoContent,
		http.StatusResetContent,
		http.StatusPartialContent,
		http.StatusMultiStatus,
		http.StatusAlreadyReported,
		http.StatusIMUsed,
		http.StatusMultipleChoices,
		http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusNotModified,
		http.StatusUseProxy,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusPaymentRequired,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusMethodNotAllowed,
		http.StatusNotAcceptable,
		http.StatusProxyAuthRequired,
		http.StatusRequestTimeout,
		http.StatusConflict,
		http.StatusGone,
		http.StatusLengthRequired,
		http.StatusPreconditionFailed,
		http.StatusRequestEntityTooLarge,
		http.StatusRequestURITooLong,
		http.StatusUnsupportedMediaType,
		http.StatusRequestedRangeNotSatisfiable,
		http.StatusExpectationFailed,
		http.StatusTeapot,
		http.StatusUnprocessableEntity,
		http.StatusLocked,
		http.StatusFailedDependency,
		http.StatusUpgradeRequired,
		http.StatusPreconditionRequired,
		http.StatusTooManyRequests,
		http.StatusRequestHeaderFieldsTooLarge,
		http.StatusUnavailableForLegalReasons,
		http.StatusInternalServerError,
		http.StatusNotImplemented,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusHTTPVersionNotSupported,
		http.StatusVariantAlsoNegotiates,
		http.StatusInsufficientStorage,
		http.StatusLoopDetected,
		http.StatusNotExtended,
		http.StatusNetworkAuthenticationRequired,
	}
)

func NewMethod() string {
	return methods[rand.Intn(len(methods))]
}

func NewScheme() string {
	switch rand.Intn(2) {
	case 0:
		return "http"
	default:
		return "https"
	}
}

func NewHost() net.IP {
	return net.IPv4(byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256)))
}

func NewPort() int {
	return 1024 + rand.Intn(65536-1024)
}

func NewHostAndPort() string {
	return fmt.Sprintf("%s:%d", NewHost().To4(), NewPort())
}

func NewAddress() string {
	return fmt.Sprintf("%s://%s", NewScheme(), NewHostAndPort())
}

func NewPath() string {
	segments := make([]string, rand.Intn(4))
	for index := range segments {
		segments[index] = test.RandomStringFromRangeAndCharset(1, 8, CharsetPath)
	}
	return "/" + strings.Join(segments, "/")
}

func RandomPathPart() string {
	return url.PathEscape(test.RandomStringFromRange(1, 8))
}

func NewURLString() string {
	return NewAddress() + NewPath()
}

func NewURL() *url.URL {
	earl, err := url.Parse(NewAddress() + NewPath())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(earl).ToNot(gomega.BeNil())
	return earl
}

func RandomHeader() http.Header {
	header := http.Header{}
	for i := test.RandomIntFromRange(2, 4); i > 0; i-- {
		values := []string{}
		for j := test.RandomIntFromRange(0, 2); j > 0; j-- {
			values = append(values, RandomHeaderValue())
		}
		header[RandomHeaderKey()] = values
	}
	return header
}

func RandomHeaderKey() string {
	return textproto.CanonicalMIMEHeaderKey(test.RandomStringFromRangeAndCharset(1, 16, CharsetName))
}

func RandomHeaderValue() string {
	return test.RandomStringFromRangeAndCharset(1, 16, CharsetValue)
}

func NewHeaderKey() string {
	return textproto.CanonicalMIMEHeaderKey(test.RandomStringFromRangeAndCharset(1, 8, CharsetName))
}

func NewHeaderValue() string {
	return test.RandomStringFromRangeAndCharset(1, 16, CharsetValue)
}

func NewParameterKey() string {
	return test.RandomStringFromRangeAndCharset(1, 8, CharsetName)
}

func NewParameterValue() string {
	return test.RandomStringFromRangeAndCharset(1, 16, CharsetValue)
}

func NewUserAgent() string {
	return test.RandomStringFromRangeAndCharset(1, 16, CharsetValue)
}

func NewTimeout() int {
	return 10 + rand.Intn(10*60-10)
}

func NewStatusCode() int {
	return statusCodes[rand.Intn(len(statusCodes))]
}

func NewRequest() *http.Request {
	req, err := http.NewRequest(NewMethod(), NewAddress(), nil)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(req).ToNot(gomega.BeNil())
	return req
}
