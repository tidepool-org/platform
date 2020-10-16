package kafkasender

import (
	"context"

	"github.com/tidepool-org/go-common/clients/shoreline"
	"github.com/tidepool-org/go-common/events"

	"github.com/tidepool-org/platform/profile"
	"github.com/tidepool-org/platform/user"
)

type EventsNotifier interface {
	NotifyUserDeleted(ctx context.Context, user user.User, userProfile *profile.Profile) error
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

func (u *userEventsNotifier) NotifyUserDeleted(ctx context.Context, user user.User, userProfile *profile.Profile) error {
	var fullName = "Tidepool User"
	if userProfile != nil || userProfile.FullName != nil {
		fullName = *userProfile.FullName
	}
	return u.Send(ctx, &events.DeleteUserEvent{
		UserData:        toUserData(user),
		ProfileFullName: fullName,
	})
}

func toUserData(user user.User) shoreline.UserData {
	var role []string
	if user.Roles != nil {
		role = *user.Roles
	}
	return shoreline.UserData{
		UserID:         *user.UserID,
		Username:       *user.Username,
		Emails:         []string{*user.Username},
		PasswordExists: *user.PasswordHash != "",
		Roles:          role,
		EmailVerified:  *user.Authenticated,
		TermsAccepted:  *user.TermsAccepted,
	}
}
