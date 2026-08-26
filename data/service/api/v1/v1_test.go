package v1_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/ant0ine/go-json-rest/rest"

	dataServiceApiV1 "github.com/tidepool-org/platform/data/service/api/v1"
	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/request"
)

var _ = Describe("V1", func() {
	Context("Routes", func() {
		It("returns the correct routes", func() {
			Expect(dataServiceApiV1.Routes()).ToNot(BeEmpty())
		})
	})

	Context("GetProvenanceFromRequest", func() {
		logger := logNull.NewLogger()
		ctx := log.NewContextWithLogger(context.Background(), logger)

		It("assigns all the things", func() {
			req, details := newTestReqAndDetails("foo", "baz", "192.0.2.1")
			prov := dataServiceApiV1.GetProvenanceFromRequest(ctx, req, details)
			Expect(*prov.ByUserID).To(Equal("baz"))
			Expect(*prov.SourceIP).To(Equal("192.0.2.1"))
			Expect(prov.ClientID).To(Equal("foo"))
		})

		It("handles a missing SourceIP", func() {
			req, details := newTestReqAndDetails("foo", "baz", "")
			prov := dataServiceApiV1.GetProvenanceFromRequest(ctx, req, details)
			Expect(*prov.ByUserID).To(Equal("baz"))
			Expect(prov.SourceIP).To(BeNil())
			Expect(prov.ClientID).To(Equal("foo"))
		})

		It("handles a missing UserID", func() {
			req, details := newTestReqAndDetails("foo", "", "192.0.2.1")
			prov := dataServiceApiV1.GetProvenanceFromRequest(ctx, req, details)
			Expect(prov.ByUserID).To(BeNil())
			Expect(*prov.SourceIP).To(Equal("192.0.2.1"))
			Expect(prov.ClientID).To(Equal("foo"))
		})

		It("handles a missing ClientID", func() {
			req, details := newTestReqAndDetails("", "bar", "192.0.2.1")
			prov := dataServiceApiV1.GetProvenanceFromRequest(ctx, req, details)
			Expect(prov).ToNot(BeNil())
			Expect(*prov.ByUserID).To(Equal("bar"))
			Expect(*prov.SourceIP).To(Equal("192.0.2.1"))
			Expect(prov.ClientID).To(Equal(""))
		})

		It("logs missing ClientIDs when expected, but not found", func() {
			testLogger := logTest.NewLogger()
			ctx := log.NewContextWithLogger(context.Background(), testLogger)
			req, _ := newTestReqAndDetails("", "", "192.0.2.1")
			details := request.NewAuthDetails(request.MethodSessionToken, "bar", "")
			prov := dataServiceApiV1.GetProvenanceFromRequest(ctx, req, details)
			Expect(prov).ToNot(BeNil())
			Expect(*prov.ByUserID).To(Equal("bar"))
			Expect(*prov.SourceIP).To(Equal("192.0.2.1"))
			Expect(prov.ClientID).To(Equal(""))
		})

		It("doesn't log missing ClientIDs for service secret authenticated requests", func() {
			ctx := log.NewContextWithLogger(context.Background(), logNull.NewLogger())
			req, _ := newTestReqAndDetails("", "", "192.0.2.1")
			details := request.NewAuthDetails(request.MethodServiceSecret, "bar", "")
			prov := dataServiceApiV1.GetProvenanceFromRequest(ctx, req, details)
			Expect(prov).ToNot(BeNil())
			Expect(*prov.ByUserID).To(Equal("bar"))
			Expect(*prov.SourceIP).To(Equal("192.0.2.1"))
			Expect(prov.ClientID).To(Equal(""))
		})
	})

	Context("SelectXFF", func() {
		It("handles a missing header", func() {
			header := http.Header{}
			Expect(dataServiceApiV1.SelectXFF(header)).To(BeNil())
		})

		It("handles an empty header", func() {
			header := http.Header{
				"X-Forwarded-For": []string{},
			}
			Expect(dataServiceApiV1.SelectXFF(header)).To(BeNil())
		})

		It("handles an invalid IP address", func() {
			header := http.Header{
				"X-Forwarded-For": []string{"invalid"},
			}
			Expect(dataServiceApiV1.SelectXFF(header)).To(BeNil())
		})

		It("handles a non-IP address", func() {
			header := http.Header{
				"X-Forwarded-For": []string{"example.com"},
			}
			Expect(dataServiceApiV1.SelectXFF(header)).To(BeNil())
		})

		It("handles a simple case", func() {
			header := http.Header{
				"X-Forwarded-For": []string{"192.0.2.1"},
			}
			Expect(dataServiceApiV1.SelectXFF(header)).To(PointTo(Equal("192.0.2.1")))
		})

		It("handles IPv6 addresses", func() {
			header := http.Header{
				"X-Forwarded-For": []string{"2001:0db8::1", "192.0.2.1"},
			}
			Expect(dataServiceApiV1.SelectXFF(header)).To(PointTo(Equal("2001:0db8::1")))
		})

		It("chooses the first IP in the first header", func() {
			header := http.Header{
				"X-Forwarded-For": []string{"192.0.2.1, 192.0.2.2", "192.0.2.3, 192.0.2.3"},
			}
			Expect(dataServiceApiV1.SelectXFF(header)).To(PointTo(Equal("192.0.2.1")))
		})

		It("handles commas with or without spaces", func() {
			header := http.Header{
				"X-Forwarded-For": []string{"192.0.2.1,192.0.2.2 , 192.0.2.5", "192.0.2.3 , 192.0.2.3"},
			}
			Expect(dataServiceApiV1.SelectXFF(header)).To(PointTo(Equal("192.0.2.1")))
		})

		It("skips private RFC-1918 and RFC-4193 addresses", func() {
			header := http.Header{
				"X-Forwarded-For": []string{"192.168.1.1, 10.1.1.1", "172.16.0.1", "fd11::1", "192.0.2.1, 192.0.2.2"},
			}
			Expect(dataServiceApiV1.SelectXFF(header)).To(PointTo(Equal("192.0.2.1")))
		})

		It("skips link-local addresses", func() {
			header := http.Header{
				"X-Forwarded-For": []string{"fe80::1, 169.254.0.1, 192.0.2.1"},
			}
			Expect(dataServiceApiV1.SelectXFF(header)).To(PointTo(Equal("192.0.2.1")))
		})

		It("skips loopback addresses", func() {
			header := http.Header{
				"X-Forwarded-For": []string{"::0, 127.0.1.1, 192.0.2.1"},
			}
			Expect(dataServiceApiV1.SelectXFF(header)).To(PointTo(Equal("192.0.2.1")))
		})

		It("skips multicast and broadcast addresses", func() {
			header := http.Header{
				"X-Forwarded-For": []string{"ff01::1", "224.0.0.53", "2001:0db8::1"},
			}
			Expect(dataServiceApiV1.SelectXFF(header)).To(PointTo(Equal("2001:0db8::1")))
		})
	})
})

func newTestReqAndDetails(clientID, userID, sourceIP string) (*rest.Request, request.AuthDetails) {
	remoteAddr := ""
	headers := http.Header{}
	if sourceIP != "" {
		remoteAddr = sourceIP + ":1234"
		headers.Set("X-Forwarded-For", sourceIP)
	}
	token := ""
	if clientID != "" {
		token = newTestToken(clientID)
		headers.Set("X-Tidepool-Session-Token", token)
	}
	req := &rest.Request{
		Request: &http.Request{
			RemoteAddr: remoteAddr,
			Header:     headers,
		},
	}
	details := request.NewAuthDetails(request.MethodSessionToken, userID, token)
	return req, details
}

func newTestToken(clientID string) string {
	header := map[string]any{"alg": "none"}
	payload := map[string]any{"azp": clientID}
	sig := map[string]any{}

	encoded := []string{}
	for _, a := range []map[string]any{header, payload, sig} {
		jsonData, err := json.Marshal(a)
		Expect(err).To(Not(HaveOccurred()))
		encoded = append(encoded, base64.RawURLEncoding.EncodeToString(jsonData))
	}
	return strings.Join(encoded, ".")
}
