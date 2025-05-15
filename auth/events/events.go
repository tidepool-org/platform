package events

import (
	"context"

	ev "github.com/tidepool-org/go-common/events"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type userDeletionEventsHandler struct {
	ev.NoopUserEventsHandler

	ctx    context.Context
	client auth.Client
}

func NewUserDataDeletionHandler(ctx context.Context, client auth.Client) ev.EventHandler {
	return ev.NewUserEventsHandler(&userDeletionEventsHandler{
		ctx:    ctx,
		client: client,
	})
}

func (u *userDeletionEventsHandler) HandleDeleteUserEvent(payload ev.DeleteUserEvent) error {
	var errs []error
	logger := log.LoggerFromContext(u.ctx).WithField("userId", payload.UserID)

	logger.Infof("Deleting restricted tokens for user")
	if err := u.client.DeleteAllRestrictedTokens(u.ctx, payload.UserID); err != nil {
		errs = append(errs, err)
		logger.WithError(err).Error("unable to delete restricted tokens for user")
	}

	logger.Infof("Deleting provider sessions for user")
	if err := u.client.DeleteUserProviderSessions(u.ctx, payload.UserID); err != nil {
		errs = append(errs, err)
		logger.WithError(err).Error("unable to delete provider sessions for user")
	}

	if len(errs) != 0 {
		return errors.New("Unable to delete auth data for user")
	}
	return nil
}
