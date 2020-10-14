package kafkasender

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/tidepool-org/go-common/clients/shoreline"
	"github.com/tidepool-org/go-common/events"
	"github.com/tidepool-org/platform/user"
)

const (
	ShorelineUserEventHandlerName = "shoreline"
	RemoveUserOperationName       = "remove_mongo_user"
	RemoveUserTokensOperationName = "remove_mongo_user_tokens"
)

var failedEvents = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "tidepool_shoreline_failed_events",
	Help: "The number of failures during even handling",
}, []string{"event_type", "handler_name", "operation_name"})

type EventsNotifier interface {
	NotifyUserDeleted(ctx context.Context, user user.User) error
}

var _ EventsNotifier = &userEventsNotifier{}

type userEventsNotifier struct {
	events.EventProducer
}

func NewUserEventsNotifier(config *events.CloudEventsConfig) (EventsNotifier, error) {
	producer, err := events.NewKafkaCloudEventsProducer(config)
	if err != nil {
		return nil, err
	}

	return &userEventsNotifier{
		EventProducer: producer,
	}, nil
}

func (u *userEventsNotifier) NotifyUserDeleted(ctx context.Context, user user.User) error {
	return u.Send(ctx, &events.DeleteUserEvent{
		UserData: toUserData(user),
	})
}

func toUserData(user user.User) shoreline.UserData {
	return shoreline.UserData{
		UserID:         *user.UserID,
		Username:       *user.Username,
		Emails:         []string{*user.Username},
		PasswordExists: *user.PasswordHash != "",
		Roles:          *user.Roles,
		EmailVerified:  *user.Authenticated,
		TermsAccepted:  *user.TermsAccepted,
	}
}