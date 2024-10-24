package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

var testUserID = mocks.TestUserID1
var testFollowedUserID = mocks.TestUserID2

const testDataSetID = "upid_000000000000"

func permsNoFollow() map[string]map[string]permission.Permissions {
	return map[string]map[string]permission.Permissions{
		mocks.TestUserID1: {
			testFollowedUserID: {
				permission.Read: map[string]interface{}{},
			},
		},
	}
}

var _ = Describe("Alerts endpoints", func() {

	testAuthenticationRequired := func(f func(dataservice.Context)) {
		t := GinkgoT()
		body := bytes.NewBuffer(mocks.MustMarshalJSON(t, alerts.Config{
			UserID:         testUserID,
			FollowedUserID: testFollowedUserID,
		}))
		dCtx := mocks.NewContext(t, "", "", body)
		dCtx.MockAlertsRepository = newMockAlertsRepo()
		badDetails := test.NewMockAuthDetails(request.MethodSessionToken, "", "")
		dCtx.WithAuthDetails(badDetails)

		f(dCtx)

		rec := dCtx.Recorder()
		Expect(rec.Code).To(Equal(http.StatusForbidden))
	}

	testUserHasFollowPermission := func(f func(dataservice.Context)) {
		t := GinkgoT()
		body := bytes.NewBuffer(mocks.MustMarshalJSON(t, alerts.Config{
			UserID:         testUserID,
			FollowedUserID: testFollowedUserID,
			UploadID:       testDataSetID,
		}))
		dCtx := mocks.NewContext(t, "", "", body)
		dCtx.MockAlertsRepository = newMockAlertsRepo()
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
		repo := newMockAlertsRepo()
		dCtx.MockAlertsRepository = repo

		f(dCtx)

		rec := dCtx.Recorder()
		Expect(rec.Code).To(Equal(http.StatusForbidden))
	}

	testInvalidJSON := func(f func(dataservice.Context)) {
		t := GinkgoT()
		body := bytes.NewBuffer([]byte(`"improper JSON data"`))
		dCtx := mocks.NewContext(t, "", "", body)
		repo := newMockAlertsRepo()
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

		It("succeeds", func() {
			t := GinkgoT()
			repo := newMockAlertsRepo()
			repo.AlertsForUserID[testFollowedUserID] = []*alerts.Config{
				{
					UserID:         testUserID,
					FollowedUserID: testFollowedUserID,
				},
			}
			dCtx := mocks.NewContext(t, "", "", nil)
			dCtx.MockAlertsRepository = repo
			rec := dCtx.Recorder()

			DeleteAlert(dCtx)

			Expect(rec.Code).To(Equal(http.StatusOK))
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

		It("succeeds", func() {
			t := GinkgoT()
			repo := newMockAlertsRepo()
			testCfg, _ := json.Marshal(testConfig())
			dCtx := mocks.NewContext(t, "", "", bytes.NewBuffer(testCfg))
			dCtx.MockAlertsRepository = repo
			rec := dCtx.Recorder()

			UpsertAlert(dCtx)

			Expect(rec.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("ListAlerts", func() {
		It("rejects unauthenticated users", func() {
			testAuthenticationRequired(ListAlerts)
		})

		It("requires that the user's token matches the userID path param", func() {
			testTokenUserIDMustMatchPathParam(ListAlerts, nil)
		})

		It("errors when no Config exists", func() {
			t := GinkgoT()
			repo := newMockAlertsRepo()
			dCtx := mocks.NewContext(t, "", "", nil)
			dCtx.MockAlertsRepository = repo
			dCtx.WithAuthDetails(mocks.ServiceAuthDetails())
			rec := dCtx.Recorder()

			ListAlerts(dCtx)

			Expect(rec.Code).To(Equal(http.StatusNotFound))
		})

		It("succeeds", func() {
			t := GinkgoT()
			repo := newMockAlertsRepo()
			dCtx := mocks.NewContext(t, "", "", nil)
			dCtx.MockAlertsRepository = repo
			dCtx.WithAuthDetails(mocks.ServiceAuthDetails())
			rec := dCtx.Recorder()
			repo.AlertsForUserID[testFollowedUserID] = []*alerts.Config{
				{FollowedUserID: "foo", UserID: "bar"},
			}

			ListAlerts(dCtx)

			Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
			got := []*alerts.Config{}
			Expect(json.NewDecoder(rec.Body).Decode(&got)).To(Succeed())
			if Expect(len(got)).To(Equal(1)) {
				Expect(got[0].UserID).To(Equal("bar"))
				Expect(got[0].FollowedUserID).To(Equal("foo"))
			}
		})
	})
	Describe("Get", func() {
		It("rejects unauthenticated users", func() {
			testAuthenticationRequired(GetAlert)
		})

		It("requires that the user's token matches the userID path param", func() {
			testTokenUserIDMustMatchPathParam(GetAlert, nil)
		})

		It("errors when no Config exists", func() {
			t := GinkgoT()
			body := bytes.NewBuffer(mocks.MustMarshalJSON(t, alerts.Config{
				UserID:         testUserID,
				FollowedUserID: testFollowedUserID,
			}))
			dCtx := mocks.NewContext(t, "", "", body)
			repo := newMockAlertsRepo()
			repo.ReturnsError(mongo.ErrNoDocuments)
			dCtx.MockAlertsRepository = repo

			GetAlert(dCtx)

			rec := dCtx.Recorder()
			Expect(rec.Code).To(Equal(http.StatusNotFound))
		})

		It("rejects users without alerting permissions", func() {
			testUserHasFollowPermission(func(dCtx dataservice.Context) {
				dCtx.Request().PathParams["userId"] = testFollowedUserID

				GetAlert(dCtx)
			})
		})

		It("succeeds", func() {
			t := GinkgoT()
			url := fmt.Sprintf("/v1/users/%s/followers/%s/alerts", testFollowedUserID, testUserID)
			dCtx := mocks.NewContext(t, "GET", url, nil)
			repo := newMockAlertsRepo()
			repo.GetAlertsResponses[testUserID+testFollowedUserID] = &alerts.Config{
				FollowedUserID: "foo",
				UserID:         "bar",
			}
			dCtx.MockAlertsRepository = repo

			GetAlert(dCtx)

			rec := dCtx.Recorder()
			Expect(rec.Code).To(Equal(http.StatusOK))
			got := &alerts.Config{}
			Expect(json.NewDecoder(rec.Body).Decode(got)).To(Succeed())
			Expect(got.UserID).To(Equal("bar"))
			Expect(got.FollowedUserID).To(Equal("foo"))
		})
	})

	Describe("GetUsersWithoutCommunication", func() {
		It("rejects unauthenticated users", func() {
			testAuthenticationRequired(GetUsersWithoutCommunication)
		})

		It("succeeds, even when there are no users found", func() {
			t := GinkgoT()
			dCtx := mocks.NewContext(t, "", "", nil)
			alertsRepo := newMockAlertsRepo()
			dCtx.MockAlertsRepository = alertsRepo
			dCtx.MockRecordsRepository = newMockRecordsRepo()
			GetUsersWithoutCommunication(dCtx)

			rec := dCtx.Recorder()
			Expect(rec.Code).To(Equal(http.StatusOK))
		})

		It("errors when the upstream repo errors", func() {
			t := GinkgoT()
			dCtx := mocks.NewContext(t, "", "", nil)
			alertsRepo := newMockAlertsRepo()
			dCtx.MockAlertsRepository = alertsRepo
			recordsRepo := newMockRecordsRepo()
			recordsRepo.UsersWithoutCommunicationError = fmt.Errorf("test error")
			dCtx.MockRecordsRepository = recordsRepo

			GetUsersWithoutCommunication(dCtx)

			rec := dCtx.Recorder()
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("succeeds, even when there are no users found", func() {
			t := GinkgoT()
			dCtx := mocks.NewContext(t, "", "", nil)
			alertsRepo := newMockAlertsRepo()
			dCtx.MockAlertsRepository = alertsRepo
			recordsRepo := newMockRecordsRepo()
			testTime := time.Unix(123, 456)
			recordsRepo.UsersWithoutCommunicationResponses = [][]alerts.LastCommunication{
				{
					{
						LastReceivedDeviceData: testTime,
					},
				},
			}
			dCtx.MockRecordsRepository = recordsRepo

			GetUsersWithoutCommunication(dCtx)

			rec := dCtx.Recorder()
			Expect(rec.Code).To(Equal(http.StatusOK))
			got := []alerts.LastCommunication{}
			Expect(json.NewDecoder(rec.Body).Decode(&got)).To(Succeed())
			if Expect(len(got)).To(Equal(1)) {
				Expect(got[0].LastReceivedDeviceData).To(BeTemporally("==", testTime))
			}
		})
	})
})

type mockRepo struct {
	UserID             string
	Error              error
	AlertsForUserID    map[string][]*alerts.Config
	GetAlertsResponses map[string]*alerts.Config
}

func newMockAlertsRepo() *mockRepo {
	return &mockRepo{
		AlertsForUserID:    map[string][]*alerts.Config{},
		GetAlertsResponses: map[string]*alerts.Config{},
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
	if resp, found := r.GetAlertsResponses[conf.UserID+conf.FollowedUserID]; found {
		return resp, nil
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

type mockRecordsRepo struct {
	UsersWithoutCommunicationResponses [][]alerts.LastCommunication
	UsersWithoutCommunicationError     error
}

func newMockRecordsRepo() *mockRecordsRepo {
	return &mockRecordsRepo{
		UsersWithoutCommunicationResponses: [][]alerts.LastCommunication{},
	}
}

func (r *mockRecordsRepo) RecordReceivedDeviceData(_ context.Context,
	_ alerts.LastCommunication) error {

	return nil
}

func (r *mockRecordsRepo) UsersWithoutCommunication(_ context.Context) (
	[]alerts.LastCommunication, error) {

	if r.UsersWithoutCommunicationError != nil {
		return nil, r.UsersWithoutCommunicationError
	}

	if len(r.UsersWithoutCommunicationResponses) > 0 {
		ret := r.UsersWithoutCommunicationResponses[0]
		r.UsersWithoutCommunicationResponses = r.UsersWithoutCommunicationResponses[1:]
		return ret, nil
	}
	return nil, nil
}

func (r *mockRecordsRepo) EnsureIndexes() error {
	return nil
}

func testConfig() *alerts.Config {
	return &alerts.Config{
		UserID:         testUserID,
		FollowedUserID: testFollowedUserID,
		UploadID:       testDataSetID,
	}
}
