package alerts

import (
	"context"
	"errors"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	nontypesglucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/dosingdecision"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("Evaluator", func() {
	Describe("EvaluateData", func() {
		It("handles data for users without any followers gracefully", func() {
			ctx, lgr := contextWithNullLoggerDeluxe()
			alertsRepo := newMockAlertsClient()

			evaluator := NewEvaluator(alertsRepo, nil, nil, lgr)
			notifications, err := evaluator.EvaluateData(ctx, testFollowedUserID, testDataSetID)

			Expect(notifications).To(BeEmpty())
			Expect(err).To(Succeed())
		})

		It("handles data queries that return empty results (perm denied)", func() {
			ctx, lgr := contextWithNullLoggerDeluxe()
			alertsRepo := newMockAlertsClient()
			alertsRepo.ListResponses = append(alertsRepo.ListResponses, []*Config{
				{
					UserID:         testUserID,
					FollowedUserID: testFollowedUserID,
					UploadID:       testDataSetID,
				},
			})
			dataRepo := newMockDataRepo()
			perms := newMockPermissionClient()

			evaluator := NewEvaluator(alertsRepo, dataRepo, perms, lgr)
			notifications, err := evaluator.EvaluateData(ctx, testFollowedUserID, testDataSetID)

			Expect(notifications).To(BeEmpty())
			Expect(err).To(Succeed())
		})

		It("filters users without permission", func() {
			// This simulates the case when permission is revoked, but the corresponding
			// alerts.Config isn't yet deleted.
			ctx, lgr := contextWithNullLoggerDeluxe()
			alertsRepo := newMockAlertsClient()
			alertsRepo.ListResponses = append(alertsRepo.ListResponses, []*Config{
				{
					UserID:         testUserID + "-2",
					FollowedUserID: testFollowedUserID,
					UploadID:       testDataSetID,
					Alerts: Alerts{
						DataAlerts: DataAlerts{
							High: &HighAlert{
								Base: Base{Enabled: true},
								Threshold: Threshold{
									Value: 10.0,
									Units: nontypesglucose.MmolL,
								},
							},
						},
					},
				},
				{
					UserID:         testUserID,
					FollowedUserID: testFollowedUserID,
					UploadID:       testDataSetID,
					Alerts: Alerts{
						DataAlerts: DataAlerts{
							High: &HighAlert{
								Base: Base{Enabled: true},
								Threshold: Threshold{
									Value: 10.0,
									Units: nontypesglucose.MmolL,
								},
							},
						},
					},
				},
			})
			dataRepo := newMockDataRepo()
			dataRepo.AlertableData = []*GetAlertableDataResponse{
				{
					Glucose: []*glucose.Glucose{testHighDatum},
				},
			}
			perms := newMockPermissionClient()
			perms.Allow(testUserID, permission.Follow, testFollowedUserID)
			// This user still has a config, but has had their follow permission revoked.
			perms.Allow(testUserID+"-2", permission.Read, testFollowedUserID)

			evaluator := NewEvaluator(alertsRepo, dataRepo, perms, lgr)
			notifications, err := evaluator.EvaluateData(ctx, testFollowedUserID, testDataSetID)

			Expect(err).To(Succeed())
			if Expect(len(notifications)).To(Equal(1)) {
				Expect(notifications[0].RecipientUserID).To(Equal(testUserID))
			}
		})

		It("handles data queries that return empty results (no data)", func() {
			ctx, lgr := contextWithNullLoggerDeluxe()
			alertsRepo := newMockAlertsClient()
			alertsRepo.ListResponses = append(alertsRepo.ListResponses, []*Config{
				{
					UserID:         testUserID,
					FollowedUserID: testFollowedUserID,
					UploadID:       testDataSetID,
				},
			})
			dataRepo := newMockDataRepo()
			perms := newMockPermissionClient()
			perms.AlwaysAllow = true

			evaluator := NewEvaluator(alertsRepo, dataRepo, perms, lgr)
			notifications, err := evaluator.EvaluateData(ctx, testFollowedUserID, testDataSetID)

			Expect(notifications).To(BeEmpty())
			Expect(err).To(Succeed())
		})

		It("returns notifications", func() {
			ctx, lgr := contextWithNullLoggerDeluxe()
			alertsRepo := newMockAlertsClient()
			alertsRepo.ListResponses = append(alertsRepo.ListResponses, []*Config{
				{
					UserID:         testUserID,
					FollowedUserID: testFollowedUserID,
					UploadID:       testDataSetID,
					Alerts: Alerts{
						DataAlerts: DataAlerts{
							UrgentLow: &UrgentLowAlert{
								Base: Base{Enabled: true},
								Threshold: Threshold{
									Value: 3.0,
									Units: nontypesglucose.MmolL,
								},
							},
						},
					},
				},
			})
			dataRepo := newMockDataRepo()
			dataRepo.AlertableData = []*GetAlertableDataResponse{
				{
					Glucose: []*glucose.Glucose{testUrgentLowDatum},
				},
			}
			perms := newMockPermissionClient()
			perms.AlwaysAllow = true

			evaluator := NewEvaluator(alertsRepo, dataRepo, perms, lgr)
			notifications, err := evaluator.EvaluateData(ctx, testFollowedUserID, testDataSetID)

			if Expect(notifications).To(HaveLen(1)) {
				msgFound := strings.Contains(notifications[0].Message, "below urgent low")
				Expect(msgFound).To(BeTrue())
			}
			Expect(err).To(Succeed())
		})

		It("queries data based on the longest delay", func() {
			ctx, lgr := contextWithNullLoggerDeluxe()
			alertsRepo := newMockAlertsClient()
			alertsRepo.ListResponses = append(alertsRepo.ListResponses, []*Config{
				{
					UserID:         testUserID + "-2",
					FollowedUserID: testFollowedUserID,
					UploadID:       testDataSetID,
					Alerts: Alerts{
						DataAlerts: DataAlerts{
							High: &HighAlert{
								Base:  Base{Enabled: true},
								Delay: DurationMinutes(6),
								Threshold: Threshold{
									Value: 10.0,
									Units: nontypesglucose.MmolL,
								},
							},
						},
					},
				},
				{
					UserID:         testUserID,
					FollowedUserID: testFollowedUserID,
					UploadID:       testDataSetID,
					Alerts: Alerts{
						DataAlerts: DataAlerts{
							High: &HighAlert{
								Base:  Base{Enabled: true},
								Delay: DurationMinutes(3),
								Threshold: Threshold{
									Value: 10.0,
									Units: nontypesglucose.MmolL,
								},
							},
						},
					},
				},
			})
			highDatum := testHighDatum
			highDatum.Blood.Base.Time = pointer.FromAny(time.Now().Add(-10 * time.Minute))
			dataRepo := newMockDataRepo()
			dataRepo.AlertableData = []*GetAlertableDataResponse{
				{
					Glucose: []*glucose.Glucose{highDatum},
				},
			}
			perms := newMockPermissionClient()
			perms.AlwaysAllow = true

			evaluator := NewEvaluator(alertsRepo, dataRepo, perms, lgr)
			notifications, err := evaluator.EvaluateData(ctx, testFollowedUserID, testDataSetID)

			if Expect(notifications).To(HaveLen(2)) {
				msgFound := strings.Contains(notifications[0].Message, "above high")
				Expect(msgFound).To(BeTrue())
			}
			Expect(err).To(Succeed())
		})

		It("wraps notifications so that changes are persisted when notifications are pushed", func() {
			ctx, lgr := contextWithNullLoggerDeluxe()
			startOfTest := time.Now()
			alertsRepo := newMockAlertsClient()
			alertsRepo.ListResponses = append(alertsRepo.ListResponses, []*Config{
				{
					UserID:         testUserID,
					FollowedUserID: testFollowedUserID,
					UploadID:       testDataSetID,
					Alerts: Alerts{
						DataAlerts: DataAlerts{
							UrgentLow: &UrgentLowAlert{
								Base: Base{
									Enabled: true,
									Activity: Activity{
										Triggered: time.Now().Add(-10 * time.Minute),
									},
								},
								Threshold: Threshold{
									Value: 3.0,
									Units: nontypesglucose.MmolL,
								},
							},
						},
					},
				},
			})
			dataRepo := newMockDataRepo()
			dataRepo.AlertableData = []*GetAlertableDataResponse{
				{
					Glucose: []*glucose.Glucose{testUrgentLowDatum},
				},
			}
			perms := newMockPermissionClient()
			perms.AlwaysAllow = true

			evaluator := NewEvaluator(alertsRepo, dataRepo, perms, lgr)
			notifications, err := evaluator.EvaluateData(ctx, testFollowedUserID, testDataSetID)
			Expect(err).To(Succeed())
			for _, notification := range notifications {
				Expect(func() { notification.Sent(time.Now()) }).ToNot(Panic())
			}

			Expect(len(notifications)).To(Equal(1))
			if Expect(len(alertsRepo.UpsertCalls)).To(Equal(1)) {
				activity := alertsRepo.UpsertCalls[0].UrgentLow.Activity
				Expect(activity.Sent).To(BeTemporally(">", startOfTest))
			}
		})

		It("persists changes when there's no new Notification", func() {
			// For example if an alert is resolved, that change should be persisted, even
			// when there isn't a notification generated.
			ctx, lgr := contextWithNullLoggerDeluxe()
			startOfTest := time.Now()
			alertsRepo := newMockAlertsClient()
			alertsRepo.ListResponses = append(alertsRepo.ListResponses, []*Config{
				{
					UserID:         testUserID,
					FollowedUserID: testFollowedUserID,
					UploadID:       testDataSetID,
					Alerts: Alerts{
						DataAlerts: DataAlerts{
							UrgentLow: &UrgentLowAlert{
								Base: Base{
									Enabled: true,
									Activity: Activity{
										Triggered: time.Now().Add(-10 * time.Minute),
									},
								},
								Threshold: Threshold{
									Value: 3.0,
									Units: nontypesglucose.MmolL,
								},
							},
						},
					},
				},
			})
			dataRepo := newMockDataRepo()
			dataRepo.AlertableData = []*GetAlertableDataResponse{
				{
					Glucose: []*glucose.Glucose{testInRangeDatum},
				},
			}
			perms := newMockPermissionClient()
			perms.AlwaysAllow = true

			evaluator := NewEvaluator(alertsRepo, dataRepo, perms, lgr)
			notifications, err := evaluator.EvaluateData(ctx, testFollowedUserID, testDataSetID)

			Expect(err).To(Succeed())
			Expect(len(notifications)).To(Equal(0))
			if Expect(len(alertsRepo.UpsertCalls)).To(Equal(1)) {
				activity := alertsRepo.UpsertCalls[0].UrgentLow.Activity
				Expect(activity.Resolved).To(BeTemporally(">", startOfTest))
			}
		})

		Context("when the user has multiple data sets", func() {
			It("ignores Configs that don't match the data set id", func() {
				ctx, lgr := contextWithNullLoggerDeluxe()
				resp1 := newTestAlertsConfig(testUserID, testDataSetID)
				resp2 := newTestAlertsConfig(testUserID+"2", testDataSetID+"2")
				alertsRepo := newMockAlertsClient()
				alertsRepo.ListResponses = append(alertsRepo.ListResponses,
					[]*Config{resp1, resp2})
				dataRepo := newMockDataRepo()
				dataRepo.AlertableData = []*GetAlertableDataResponse{
					{Glucose: []*glucose.Glucose{testUrgentLowDatum}},
				}
				perms := newMockPermissionClient()
				perms.AlwaysAllow = true

				evaluator := NewEvaluator(alertsRepo, dataRepo, perms, lgr)
				notifications, err := evaluator.EvaluateData(ctx,
					testFollowedUserID, testDataSetID)

				Expect(err).To(Succeed())
				if Expect(len(notifications)).To(Equal(1)) {
					recipientUserID := notifications[0].Notification.RecipientUserID
					Expect(recipientUserID).To(Equal(testUserID))
				}
			})
		})
	})
})

func newTestAlertsConfig(userID, dataSetID string) *Config {
	return &Config{
		UserID:         userID,
		FollowedUserID: testFollowedUserID,
		UploadID:       dataSetID,
		Alerts: Alerts{
			DataAlerts: DataAlerts{
				UrgentLow: testUrgentLowAlert(),
			},
		},
	}
}

type mockAlertsClient struct {
	UsersWithoutCommsError     error
	UsersWithoutCommsResponses [][]LastCommunication
	ListResponses              [][]*Config
	ListError                  error
	UpsertError                error
	UpsertCalls                []*Config
}

func newMockAlertsClient() *mockAlertsClient {
	return &mockAlertsClient{
		UsersWithoutCommsResponses: [][]LastCommunication{},
		ListResponses:              [][]*Config{},
		UpsertCalls:                []*Config{},
	}
}

func (c *mockAlertsClient) Get(ctx context.Context, conf *Config) (*Config, error) {
	return nil, nil
}

func (c *mockAlertsClient) Upsert(ctx context.Context, conf *Config) error {
	c.UpsertCalls = append(c.UpsertCalls, conf)
	if c.UpsertError != nil {
		return c.UpsertError
	}
	return nil
}

func (c *mockAlertsClient) Delete(ctx context.Context, conf *Config) error {
	return nil
}

func (c *mockAlertsClient) List(ctx context.Context, userID string) ([]*Config, error) {
	if c.ListError != nil {
		return nil, c.ListError
	}
	if len(c.ListResponses) > 0 {
		ret := c.ListResponses[0]
		c.ListResponses = c.ListResponses[1:]
		return ret, nil
	}
	return []*Config{}, nil
}

func (c *mockAlertsClient) UsersWithoutCommunication(context.Context) (
	[]LastCommunication, error) {

	if c.UsersWithoutCommsError != nil {
		return nil, c.UsersWithoutCommsError
	}
	if len(c.UsersWithoutCommsResponses) > 0 {
		ret := c.UsersWithoutCommsResponses[0]
		c.UsersWithoutCommsResponses = c.UsersWithoutCommsResponses[1:]
		return ret, nil
	}
	return nil, nil
}

func (c *mockAlertsClient) EnsureIndexes() error {
	return nil
}

type mockDataRepo struct {
	AlertableData []*GetAlertableDataResponse
}

func newMockDataRepo() *mockDataRepo {
	return &mockDataRepo{
		AlertableData: []*GetAlertableDataResponse{},
	}
}

func (r *mockDataRepo) GetAlertableData(ctx context.Context, params GetAlertableDataParams) (
	*GetAlertableDataResponse, error) {

	if len(r.AlertableData) > 0 {
		ret := r.AlertableData[0]
		r.AlertableData = r.AlertableData[1:]
		return ret, nil
	}

	return &GetAlertableDataResponse{
		DosingDecisions: []*dosingdecision.DosingDecision{},
		Glucose:         []*glucose.Glucose{},
	}, nil
}

type mockPermissionClient struct {
	AlwaysAllow bool
	Perms       map[string]permission.Permissions
}

func newMockPermissionClient() *mockPermissionClient {
	return &mockPermissionClient{
		Perms: map[string]permission.Permissions{},
	}
}

func (c *mockPermissionClient) GetUserPermissions(ctx context.Context,
	requestUserID string, targetUserID string) (permission.Permissions, error) {

	if c.AlwaysAllow {
		return map[string]permission.Permission{
			permission.Follow: {},
			permission.Read:   {},
		}, nil
	}

	if p, ok := c.Perms[c.Key(requestUserID, targetUserID)]; ok {
		return p, nil
	} else {
		return nil, errors.New("test error NOT FOUND")
	}
}

func (c *mockPermissionClient) Allow(requestUserID, perm, targetUserID string) {
	key := c.Key(requestUserID, targetUserID)
	if _, found := c.Perms[key]; !found {
		c.Perms[key] = permission.Permissions{}
	}
	c.Perms[key][perm] = permission.Permission{}
}

func (c *mockPermissionClient) Key(requesterUserID, targetUserID string) string {
	return requesterUserID + targetUserID
}
