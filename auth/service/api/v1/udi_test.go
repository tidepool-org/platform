package v1_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	authServiceApiV1 "github.com/tidepool-org/platform/auth/service/api/v1"
	authServiceTest "github.com/tidepool-org/platform/auth/service/test"
	"github.com/tidepool-org/platform/cluster"
	clusterTest "github.com/tidepool-org/platform/cluster/test"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
	serviceTest "github.com/tidepool-org/platform/service/test"
)

var _ = Describe("UDI", func() {
	var releaseReporter *clusterTest.MockReleaseReporter
	var rtr *authServiceApiV1.Router
	var res *serviceTest.MockRestResponseWriter

	BeforeEach(func() {
		releaseReporter = clusterTest.NewMockReleaseReporter(gomock.NewController(GinkgoT()))
		svc := authServiceTest.NewService()
		svc.ClusterReleaseReporterImpl = releaseReporter
		var err error
		rtr, err = authServiceApiV1.NewRouter(svc)
		Expect(err).ToNot(HaveOccurred())
		res = serviceTest.NewMockRestResponseWriter()
	})

	newRequest := func(details request.AuthDetails, body string) *rest.Request {
		ctx := context.Background()
		if details != nil {
			ctx = request.NewContextWithAuthDetails(ctx, details)
		}
		var reader io.Reader
		if body != "" {
			reader = strings.NewReader(body)
		}
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "/v1/udi", reader)
		Expect(err).ToNot(HaveOccurred())
		return &rest.Request{Request: httpReq, PathParams: map[string]string{}}
	}

	route := func() *rest.Route {
		for _, r := range rtr.UDIRoutes() {
			if r.HttpMethod == http.MethodPost && r.PathExp == "/v1/udi" {
				return r
			}
		}
		Fail("route POST /v1/udi not found")
		return nil
	}

	release := &cluster.Release{
		Revision: "868bd7b1c4f0e3a5d2b9c8e7f6a5b4c3d2e1f0a9",
		Time:     time.Date(2026, 11, 3, 18, 42, 7, 0, time.UTC),
	}
	approved := `{
		"productName": "Tidepool Data Platform",
		"manufacturerName": "Tidepool Project",
		"manufacturerAddress": "555 Bryant St., #429, Palo Alto, CA 94301",
		"releaseDate": "2026-11-03",
		"releaseIdentifier": "868bd7b",
		"softwareVersion": "1.103.2-868bd7b",
		"deviceIdentifier": "00884400418012",
		"udi": "(01)00884400418012(11)261103(8012)1.103.2-868bd7b"
	}`

	It("returns the UDI to a user", func() {
		releaseReporter.EXPECT().Release(gomock.Any()).Return(release, nil)
		route().Func(res, newRequest(serviceTest.NewMockAuthDetailsDefault(), `{"clientVersion": "1.103.2"}`))
		Expect(res.Code).To(Equal(http.StatusOK))
		Expect(res.Body.String()).To(MatchJSON(approved))
	})

	It("returns the UDI to a service", func() {
		releaseReporter.EXPECT().Release(gomock.Any()).Return(release, nil)
		route().Func(res, newRequest(serviceTest.NewMockAuthDetails(request.MethodServiceSecret, "", serviceTest.TestToken2), `{"clientVersion": "1.103.2"}`))
		Expect(res.Code).To(Equal(http.StatusOK))
		Expect(res.Body.String()).To(MatchJSON(approved))
	})

	It("returns unauthenticated without a session", func() {
		route().Func(res, newRequest(nil, `{"clientVersion": "1.103.2"}`))
		Expect(res.Code).To(Equal(http.StatusUnauthorized))
	})

	DescribeTable("returns bad request for an invalid client version",
		func(body string, expectedCode string) {
			route().Func(res, newRequest(serviceTest.NewMockAuthDetailsDefault(), body))
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			Expect(res.Body.String()).To(ContainSubstring(`"code":"` + expectedCode + `"`))
		},
		Entry("no body", "", "json-not-found"),
		Entry("malformed body", `{"clientVersion": `, "json-malformed"),
		Entry("missing", `{}`, "value-not-exists"),
		Entry("not a string", `{"clientVersion": 1103}`, "type-not-string"),
		Entry("longer than 12 characters", `{"clientVersion": "1.103.2-rc.10"}`, "length-out-of-range"),
		Entry("GS1 application identifier", `{"clientVersion": "1.0(01)2"}`, "value-not-matches"),
	)

	It("returns service unavailable without the cause when the release cannot be read", func() {
		releaseReporter.EXPECT().Release(gomock.Any()).Return(nil, errors.New(`kustomization history has no successful reconciliation of revision "master@sha1:868bd7b1c4f0e3a5d2b9c8e7f6a5b4c3d2e1f0a9"`))
		route().Func(res, newRequest(serviceTest.NewMockAuthDetailsDefault(), `{"clientVersion": "1.103.2"}`))
		Expect(res.Code).To(Equal(http.StatusServiceUnavailable))
		Expect(res.Body.String()).To(MatchJSON(`{"code":"udi-unavailable","title":"udi unavailable","detail":"unable to identify the release in service"}`))
	})

	It("returns service unavailable when the release is invalid", func() {
		releaseReporter.EXPECT().Release(gomock.Any()).Return(&cluster.Release{Revision: "868bd7b1c4f0e3a5d2b9c8e7f6a5b4c3d2e1f0a9"}, nil)
		route().Func(res, newRequest(serviceTest.NewMockAuthDetailsDefault(), `{"clientVersion": "1.103.2"}`))
		Expect(res.Code).To(Equal(http.StatusServiceUnavailable))
		Expect(res.Body.String()).To(ContainSubstring(`"code":"udi-unavailable"`))
	})
})
