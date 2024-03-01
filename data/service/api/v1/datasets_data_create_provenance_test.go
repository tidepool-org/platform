package v1_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "github.com/tidepool-org/platform/data/service/api/v1"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/request"
)

var _ = Describe("collectProvenanceInfo", func() {
	// logger, err := logJson.NewLogger(os.Stderr, log.DefaultLevelRanks(), log.DefaultLevel())
	// Expect(err).ShouldNot(HaveOccurred())
	logger := null.NewLogger()
	ctx := log.NewContextWithLogger(context.Background(), logger)

	It("assigns all the things", func() {
		req, details := newTestReqAndDetails("foo", "baz", "192.0.2.1")
		prov := v1.CollectProvenanceInfo(ctx, req, details)
		Expect(prov.ByUserID).To(Equal("baz"))
		Expect(prov.SourceIP).To(Equal("192.0.2.1"))
		Expect(prov.ClientID).To(Equal("foo"))
	})

	It("handles a missing SourceIP", func() {
		req, details := newTestReqAndDetails("foo", "baz", "")
		prov := v1.CollectProvenanceInfo(ctx, req, details)
		Expect(prov.ByUserID).To(Equal("baz"))
		Expect(prov.SourceIP).To(Equal(""))
		Expect(prov.ClientID).To(Equal("foo"))
	})

	It("handles a missing UserID", func() {
		req, details := newTestReqAndDetails("foo", "", "192.0.2.1")
		prov := v1.CollectProvenanceInfo(ctx, req, details)
		Expect(prov.ByUserID).To(Equal(""))
		Expect(prov.SourceIP).To(Equal("192.0.2.1"))
		Expect(prov.ClientID).To(Equal("foo"))
	})

	It("handles a missing ClientID", func() {
		req, details := newTestReqAndDetails("", "bar", "192.0.2.1")
		prov := v1.CollectProvenanceInfo(ctx, req, details)
		Expect(prov.ByUserID).To(Equal("bar"))
		Expect(prov.SourceIP).To(Equal("192.0.2.1"))
		Expect(prov.ClientID).To(Equal(""))
	})
})

func newTestReqAndDetails(clientID, userID, sourceIP string) (*rest.Request, request.AuthDetails) {
	remoteAddr := ""
	if sourceIP != "" {
		remoteAddr = sourceIP + ":1234"
	}
	req := &rest.Request{
		Request: &http.Request{
			RemoteAddr: remoteAddr,
			Header: http.Header{
				"X-Tidepool-Session-Token": {newTestToken(clientID)},
				"X-Forwarded-For":          {sourceIP}},
		},
	}
	details := request.NewAuthDetails("", userID, "token")
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

var _ = Describe("selectXFF", func() {
	It("handles a simple case", func() {
		h := http.Header{
			"X-Forwarded-For": []string{"192.0.2.1"},
		}
		Expect(v1.SelectXFF(h)).To(Equal("192.0.2.1"))
	})

	It("handles IPv6 addresses", func() {
		h := http.Header{
			"X-Forwarded-For": []string{"2001:0db8::1", "192.0.2.1"},
		}
		Expect(v1.SelectXFF(h)).To(Equal("2001:0db8::1"))
	})

	It("chooses the first IP in the first header", func() {
		h := http.Header{
			"X-Forwarded-For": []string{"192.0.2.1, 192.0.2.2", "192.0.2.3, 192.0.2.3"},
		}
		Expect(v1.SelectXFF(h)).To(Equal("192.0.2.1"))
	})

	It("handles commas with or without spaces", func() {
		h := http.Header{
			"X-Forwarded-For": []string{"192.0.2.1,192.0.2.2 , 192.0.2.5", "192.0.2.3 , 192.0.2.3"},
		}
		Expect(v1.SelectXFF(h)).To(Equal("192.0.2.1"))
	})

	It("skips private RFC-1918 and RFC-4193 addresses", func() {
		h := http.Header{
			"X-Forwarded-For": []string{"192.168.1.1, 10.1.1.1", "172.16.0.1", "fd11::1", "192.0.2.1, 192.0.2.2"},
		}
		Expect(v1.SelectXFF(h)).To(Equal("192.0.2.1"))
	})

	It("skips link-local addresses", func() {
		h := http.Header{
			"X-Forwarded-For": []string{"fe80::1, 169.254.0.1, 192.0.2.1"},
		}
		Expect(v1.SelectXFF(h)).To(Equal("192.0.2.1"))
	})

	It("skips loopback addresses", func() {
		h := http.Header{
			"X-Forwarded-For": []string{"::0, 127.0.1.1, 192.0.2.1"},
		}
		Expect(v1.SelectXFF(h)).To(Equal("192.0.2.1"))
	})

	It("skips multicast and broadcast addresses", func() {
		h := http.Header{
			"X-Forwarded-For": []string{"ff01::1", "224.0.0.53", "2001:0db8::1"},
		}
		Expect(v1.SelectXFF(h)).To(Equal("2001:0db8::1"))
	})
})
