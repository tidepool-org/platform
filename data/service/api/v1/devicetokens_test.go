package v1

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/service/api/v1/mocks"
	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/request"
)

var _ = Describe("Device tokens endpoints", func() {

	Describe("Upsert", func() {
		It("succeeds with valid input", func() {
			t := GinkgoT()
			body := buff(`{"apple":{"environment":"sandbox","token":"b3BhcXVldG9rZW4="}}`)
			dCtx := mocks.NewContext(t, "", "", body)
			repo := newMockDeviceTokensRepo()
			dCtx.MockDeviceTokensRepository = repo

			UpsertDeviceToken(dCtx)

			rec := dCtx.Recorder()
			Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		})

		It("rejects unauthenticated users", func() {
			t := GinkgoT()
			body := buff(`{"apple":{"environment":"sandbox","token":"blah"}}`)
			dCtx := mocks.NewContext(t, "", "", body)
			dCtx.MockAlertsRepository = newMockRepo()
			badDetails := mocks.NewAuthDetails(request.MethodSessionToken, "", "")
			dCtx.WithAuthDetails(badDetails)

			UpsertDeviceToken(dCtx)

			rec := dCtx.Recorder()
			Expect(rec.Code).To(Equal(http.StatusForbidden))
		})

		It("accepts authenticated service users", func() {
			t := GinkgoT()
			body := buff(`{"apple":{"environment":"sandbox","token":"blah"}}`)
			dCtx := mocks.NewContext(t, "", "", body)
			dCtx.WithAuthDetails(mocks.ServiceAuthDetails())
			dCtx.MockDeviceTokensRepository = newMockDeviceTokensRepo()

			UpsertDeviceToken(dCtx)

			rec := dCtx.Recorder()
			Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		})

		It("requires that the user's token matches the userID path param", func() {
			t := GinkgoT()
			dCtx := mocks.NewContext(t, "", "", nil)
			dCtx.RESTRequest.PathParams["userID"] = "bad"
			repo := newMockDeviceTokensRepo()
			dCtx.MockDeviceTokensRepository = repo

			UpsertDeviceToken(dCtx)

			rec := dCtx.Recorder()
			Expect(rec.Code).To(Equal(http.StatusForbidden))
		})

		It("errors on invalid JSON for device tokens", func() {
			t := GinkgoT()
			body := bytes.NewBuffer([]byte(`"improper JSON data"`))
			dCtx := mocks.NewContext(t, "", "", body)
			repo := newMockDeviceTokensRepo()
			dCtx.MockDeviceTokensRepository = repo

			UpsertDeviceToken(dCtx)

			rec := dCtx.Recorder()
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
		})

	})

})

type mockDeviceTokensRepo struct {
	UserID string
	Error  error
}

func newMockDeviceTokensRepo() *mockDeviceTokensRepo {
	return &mockDeviceTokensRepo{}
}

func (r *mockDeviceTokensRepo) ReturnsError(err error) {
	r.Error = err
}

func (r *mockDeviceTokensRepo) Upsert(ctx context.Context, conf *devicetokens.Document) error {
	if r.Error != nil {
		return r.Error
	}
	if conf != nil {
		r.UserID = conf.UserID
	}
	return nil
}

func (r *mockDeviceTokensRepo) EnsureIndexes() error {
	return nil
}

func buff(template string, args ...any) *bytes.Buffer {
	return bytes.NewBuffer([]byte(fmt.Sprintf(template, args...)))
}
