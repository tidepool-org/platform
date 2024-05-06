package v1

import (
	"bytes"
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/alerts"
	dataservice "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/service/api/v1/mocks"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/test"
)

func permsNoFollow() map[string]map[string]permission.Permissions {
	return map[string]map[string]permission.Permissions{
		mocks.TestUserID1: {
			mocks.TestUserID2: {
				permission.Read: map[string]interface{}{},
			},
		},
	}
}

var _ = Describe("Alerts endpoints", func() {

	testAuthenticationRequired := func(f func(dataservice.Context)) {
		t := GinkgoT()
		body := bytes.NewBuffer(mocks.MustMarshalJSON(t, alerts.Config{
			UserID:         mocks.TestUserID1,
			FollowedUserID: mocks.TestUserID2,
		}))
		dCtx := mocks.NewContext(t, "", "", body)
		dCtx.MockAlertsRepository = newMockRepo()
		badDetails := test.NewMockAuthDetails(request.MethodSessionToken, "", "")
		dCtx.WithAuthDetails(badDetails)

		f(dCtx)

		rec := dCtx.Recorder()
		Expect(rec.Code).To(Equal(http.StatusForbidden))
	}

	testUserHasFollowPermission := func(f func(dataservice.Context)) {
		t := GinkgoT()
		body := bytes.NewBuffer(mocks.MustMarshalJSON(t, alerts.Config{
			UserID:         mocks.TestUserID1,
			FollowedUserID: mocks.TestUserID2,
		}))
		dCtx := mocks.NewContext(t, "", "", body)
		dCtx.MockAlertsRepository = newMockRepo()
		dCtx.MockPermissionClient = mocks.NewPermission(permsNoFollow(), nil, nil)

		f(dCtx)

		rec := dCtx.Recorder()
		Expect(rec.Code).To(Equal(http.StatusForbidden))
	}

	testTokenUserIDMustMatchPathParam := func(f func(dataservice.Context), details *test.MockAuthDetails) {
		t := GinkgoT()
		dCtx := mocks.NewContext(t, "", "", nil)
		if details != nil {
			dCtx.WithAuthDetails(details)
		}
		dCtx.RESTRequest.PathParams["followerUserId"] = "bad"
		repo := newMockRepo()
		dCtx.MockAlertsRepository = repo

		f(dCtx)

		rec := dCtx.Recorder()
		Expect(rec.Code).To(Equal(http.StatusForbidden))
	}

	testInvalidJSON := func(f func(dataservice.Context)) {
		t := GinkgoT()
		body := bytes.NewBuffer([]byte(`"improper JSON data"`))
		dCtx := mocks.NewContext(t, "", "", body)
		repo := newMockRepo()
		dCtx.MockAlertsRepository = repo

		f(dCtx)

		rec := dCtx.Recorder()
		Expect(rec.Code).To(Equal(http.StatusBadRequest))
	}

	Describe("Delete", func() {
		It("rejects unauthenticated users", func() {
			testAuthenticationRequired(DeleteAlert)
		})

		It("requires that the user's token matches the userID path param", func() {
			testTokenUserIDMustMatchPathParam(DeleteAlert, nil)
		})

		It("rejects users without alerting permissions", func() {
			testUserHasFollowPermission(DeleteAlert)
		})
	})

	Describe("Upsert", func() {
		It("rejects unauthenticated users", func() {
			testAuthenticationRequired(UpsertAlert)
		})

		It("requires that the user's token matches the userID path param", func() {
			testTokenUserIDMustMatchPathParam(UpsertAlert, nil)
		})

		It("errors on invalid JSON", func() {
			testInvalidJSON(UpsertAlert)
		})

		It("rejects users without alerting permissions", func() {
			testUserHasFollowPermission(UpsertAlert)
		})
	})

	Describe("Get", func() {
		It("rejects unauthenticated users", func() {
			testAuthenticationRequired(GetAlert)
		})

		It("requires that the user's token matches the userID path param", func() {
			testTokenUserIDMustMatchPathParam(GetAlert, nil)
		})

		It("errors when Config doesn't exist", func() {
			t := GinkgoT()
			body := bytes.NewBuffer(mocks.MustMarshalJSON(t, alerts.Config{
				UserID:         mocks.TestUserID1,
				FollowedUserID: mocks.TestUserID2,
			}))
			dCtx := mocks.NewContext(t, "", "", body)
			repo := newMockRepo()
			repo.ReturnsError(mongo.ErrNoDocuments)
			dCtx.MockAlertsRepository = repo

			GetAlert(dCtx)

			rec := dCtx.Recorder()
			Expect(rec.Code).To(Equal(http.StatusNotFound))
		})

		It("rejects users without alerting permissions", func() {
			testUserHasFollowPermission(func(dCtx dataservice.Context) {
				dCtx.Request().PathParams["userId"] = mocks.TestUserID2

				GetAlert(dCtx)
			})
		})
	})
})

type mockRepo struct {
	UserID          string
	Error           error
	AlertsForUserID map[string][]*alerts.Config
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		AlertsForUserID: make(map[string][]*alerts.Config),
	}
}

func (r *mockRepo) ReturnsError(err error) {
	r.Error = err
}

func (r *mockRepo) Upsert(ctx context.Context, conf *alerts.Config) error {
	if r.Error != nil {
		return r.Error
	}
	if conf != nil {
		r.UserID = conf.UserID
	}
	return nil
}

func (r *mockRepo) Get(ctx context.Context, conf *alerts.Config) (*alerts.Config, error) {
	if r.Error != nil {
		return nil, r.Error
	}
	if conf != nil {
		r.UserID = conf.UserID
	}
	return &alerts.Config{}, nil
}

func (r *mockRepo) Delete(ctx context.Context, conf *alerts.Config) error {
	if r.Error != nil {
		return r.Error
	}
	if conf != nil {
		r.UserID = conf.UserID
	}
	return nil
}

func (r *mockRepo) List(ctx context.Context, userID string) ([]*alerts.Config, error) {
	if r.Error != nil {
		return nil, r.Error
	}
	r.UserID = userID
	alerts, ok := r.AlertsForUserID[userID]
	if !ok {
		return nil, nil
	}
	return alerts, nil
}

func (r *mockRepo) EnsureIndexes() error {
	return nil
}
