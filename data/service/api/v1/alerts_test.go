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
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
)

func permsNoAlerting() map[string]map[string]permission.Permissions {
	return map[string]map[string]permission.Permissions{
		mocks.TestUserID1: {
			mocks.TestUserID2: {
				permission.Read: map[string]interface{}{},
			},
		},
	}
}

var _ = Describe("Alerts endpoints", func() {

	testAuthentication := func(f func(dataservice.Context)) {
		t := GinkgoT()
		body := bytes.NewBuffer(mocks.MustMarshalJSON(t, alerts.Config{
			UserID:     mocks.TestUserID1,
			FollowedID: mocks.TestUserID2,
		}))
		dCtx := mocks.NewContext(t, "", "", body)
		dCtx.MockAlertsRepository = newMockRepo()
		badDetails := mocks.NewDetails(request.MethodSessionToken, "", "")
		dCtx.WithDetails(badDetails)

		f(dCtx)

		rec := dCtx.Recorder()
		Expect(rec.Code).To(Equal(http.StatusForbidden))
	}

	testPermissions := func(f func(dataservice.Context)) {
		t := GinkgoT()
		body := bytes.NewBuffer(mocks.MustMarshalJSON(t, alerts.Config{
			UserID:     mocks.TestUserID1,
			FollowedID: mocks.TestUserID2,
		}))
		dCtx := mocks.NewContext(t, "", "", body)
		dCtx.MockAlertsRepository = newMockRepo()
		dCtx.MockPermissionClient = mocks.NewPermission(permsNoAlerting(), nil, nil)

		f(dCtx)

		rec := dCtx.Recorder()
		Expect(rec.Code).To(Equal(http.StatusForbidden))
	}

	testUserID := func(f func(dataservice.Context)) {
		t := GinkgoT()
		body := bytes.NewBuffer(mocks.MustMarshalJSON(t, alerts.Config{
			UserID:     "00000000-dead-4123-beef-000000000000",
			FollowedID: mocks.TestUserID2,
		}))
		dCtx := mocks.NewContext(t, "", "", body)
		repo := newMockRepo()
		repo.ExpectsOwnerID(mocks.TestUserID2)
		dCtx.MockAlertsRepository = repo
		badDetails := mocks.NewDetails(request.MethodSessionToken, mocks.TestUserID1, "")
		dCtx.WithDetails(badDetails)

		f(dCtx)

		Expect(repo.UserID).To(Equal(mocks.TestUserID1))
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
		It("rejects unauthenticated users", func() {
			testAuthentication(DeleteAlert)
		})

		It("uses the authenticated user's userID", func() {
			testUserID(DeleteAlert)
		})

		It("errors on invalid JSON", func() {
			testInvalidJSON(DeleteAlert)
		})

		It("rejects users without alerting permissions", func() {
			testPermissions(DeleteAlert)
		})

	})

	Describe("Upsert", func() {
		It("rejects unauthenticated users", func() {
			testAuthentication(UpsertAlert)
		})

		It("uses the authenticated user's userID", func() {
			testUserID(UpsertAlert)
		})

		It("errors on invalid JSON", func() {
			testInvalidJSON(UpsertAlert)
		})

		It("rejects users without alerting permissions", func() {
			testPermissions(UpsertAlert)
		})

	})
})

type mockRepo struct {
	UserID string
}

func newMockRepo() *mockRepo {
	return &mockRepo{}
}

func (r *mockRepo) ExpectsOwnerID(ownerID string) {
	r.UserID = ownerID
}

func (r *mockRepo) Upsert(ctx context.Context, conf *alerts.Config) error {
	if conf != nil {
		r.UserID = conf.UserID
	}
	return nil
}

func (r *mockRepo) Delete(ctx context.Context, conf *alerts.Config) error {
	if conf != nil {
		r.UserID = conf.UserID
	}
	return nil
}
