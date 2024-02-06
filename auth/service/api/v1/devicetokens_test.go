package v1

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	serviceTest "github.com/tidepool-org/platform/auth/service/test"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/test"
)

var _ = Describe("Device tokens endpoints", func() {
	var rtr *Router
	BeforeEach(func() {
		svc := serviceTest.NewService()
		var err error
		rtr, err = NewRouter(svc)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Upsert", func() {
		It("succeeds with valid input", func() {
			res := test.NewMockRestResponseWriter()
			req := newDeviceTokensTestRequest(nil, nil, "")

			rtr.UpsertDeviceToken(res, req)

			Expect(res.Code).To(Equal(http.StatusOK))
		})

		It("rejects service users", func() {
			svcDetails := test.NewMockAuthDetails(request.MethodServiceSecret, "", test.TestToken2)
			res := test.NewMockRestResponseWriter()
			req := newDeviceTokensTestRequest(svcDetails, nil, "")

			rtr.UpsertDeviceToken(res, req)

			Expect(res.Code).To(Equal(http.StatusForbidden))
		})

		It("requires that the user's token matches the userId path param", func() {
			res := test.NewMockRestResponseWriter()
			req := newDeviceTokensTestRequest(nil, nil, "bad")

			rtr.UpsertDeviceToken(res, req)

			Expect(res.Code).To(Equal(http.StatusForbidden))
		})

		It("errors on invalid JSON for device tokens", func() {
			body := buff(`"improper JSON data"`)
			res := test.NewMockRestResponseWriter()
			req := newDeviceTokensTestRequest(nil, body, "")

			rtr.UpsertDeviceToken(res, req)

			Expect(res.Code).To(Equal(http.StatusBadRequest))
		})

	})

})

func buff(template string, args ...any) *bytes.Buffer {
	return bytes.NewBuffer([]byte(fmt.Sprintf(template, args...)))
}

// newDeviceTokensTestRequest helps build requests for device tokens tests.
func newDeviceTokensTestRequest(auth request.AuthDetails, body io.Reader, userIDFromPath string) *rest.Request {
	if auth == nil {
		auth = test.NewMockAuthDetails(request.MethodSessionToken, test.TestUserID1, test.TestToken1)
	}
	if body == nil {
		body = buff(`{"apple":{"environment":"sandbox","token":"b3BhcXVldG9rZW4="}}`)
	}
	if userIDFromPath == "" {
		userIDFromPath = test.TestUserID1
	}

	ctx := request.NewContextWithAuthDetails(context.Background(), auth)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "", body)
	Expect(err).ToNot(HaveOccurred())
	return &rest.Request{
		Request:    httpReq,
		PathParams: map[string]string{"userId": userIDFromPath},
	}

}
