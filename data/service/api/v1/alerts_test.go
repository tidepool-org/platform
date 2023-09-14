package v1

import (
	"bytes"
	"context"

	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/alerts"
	dataservice "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/service/api/v1/mocks"
	"github.com/tidepool-org/platform/request"
)

var _ = Describe("Alerts endpoints", func() {

	testUnauthorized := func(f func(dataservice.Context)) {
		t := GinkgoT()
		body := bytes.NewBuffer(mocks.MustMarshalJSON(t, alerts.Config{
			OwnerID:   mocks.TestUserID1,
			InvitorID: mocks.TestUserID2,
		}))
		dCtx := mocks.NewContext(t, "", "", body)
		dCtx.MockAlertsRepository = newMockRepo()
		badDetails := mocks.NewDetails(request.MethodSessionToken, "", "")
		dCtx.WithDetails(badDetails)

		f(dCtx)

		rec := dCtx.Recorder()
		Expect(rec.Code).To(Equal(http.StatusForbidden))
	}

	testUserID := func(f func(dataservice.Context)) {
		t := GinkgoT()
		body := bytes.NewBuffer(mocks.MustMarshalJSON(t, alerts.Config{
			OwnerID:   "someotheruser", // pass in whatever value, it should be overridden
			InvitorID: mocks.TestUserID2,
		}))
		dCtx := mocks.NewContext(t, "", "", body)
		repo := newMockRepo()
		repo.ExpectsOwnerID(mocks.TestUserID2)
		dCtx.MockAlertsRepository = repo
		badDetails := mocks.NewDetails(request.MethodSessionToken, mocks.TestUserID1, "")
		dCtx.WithDetails(badDetails)

		f(dCtx)

		Expect(repo.OwnerID).To(Equal(mocks.TestUserID1))
		rec := dCtx.Recorder()
		Expect(rec.Code).To(Equal(http.StatusOK))
	}

	testInvalidJSON := func(f func(dataservice.Context)) {
		t := GinkgoT()
		body := bytes.NewBuffer([]byte(`"improper JSON data"`))
		dCtx := mocks.NewContext(t, "", "", body)
		repo := newMockRepo()
		repo.ExpectsOwnerID(mocks.TestUserID2)
		dCtx.MockAlertsRepository = repo
		badDetails := mocks.NewDetails(request.MethodSessionToken, mocks.TestUserID1, "")
		dCtx.WithDetails(badDetails)

		f(dCtx)

		rec := dCtx.Recorder()
		Expect(rec.Code).To(Equal(http.StatusBadRequest))
	}

	Describe("Delete", func() {
		It("rejects unauthorized users", func() {
			testUnauthorized(DeleteAlert)
		})

		It("uses the authenticated user's userID", func() {
			testUserID(DeleteAlert)
		})

		It("errors on invalid JSON", func() {
			testInvalidJSON(DeleteAlert)
		})
	})

	Describe("Upsert", func() {
		It("rejects unauthorized users", func() {
			testUnauthorized(UpsertAlert)
		})

		It("uses the authenticated user's userID", func() {
			testUserID(UpsertAlert)
		})

		It("errors on invalid JSON", func() {
			testInvalidJSON(UpsertAlert)
		})
	})
})

type mockRepo struct {
	OwnerID string
}

func newMockRepo() *mockRepo {
	return &mockRepo{}
}

func (r *mockRepo) ExpectsOwnerID(ownerID string) {
	r.OwnerID = ownerID
}

func (r *mockRepo) Upsert(ctx context.Context, conf *alerts.Config) error {
	if conf != nil {
		r.OwnerID = conf.OwnerID
	}
	return nil
}

func (r *mockRepo) Delete(ctx context.Context, conf *alerts.Config) error {
	if conf != nil {
		r.OwnerID = conf.OwnerID
	}
	return nil
}
