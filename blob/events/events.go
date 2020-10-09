package events

import (
	"context"

	ev "github.com/tidepool-org/go-common/events"

	"github.com/tidepool-org/platform/blob"
	"github.com/tidepool-org/platform/log"
)

type userDeletionEventsHandler struct {
	ev.NoopUserEventsHandler

	ctx        context.Context
	blobClient blob.Client
}

func NewUserDataDeletionHandler(ctx context.Context, blobClient blob.Client) ev.EventHandler {
	return ev.NewUserEventsHandler(&userDeletionEventsHandler{
		ctx:        ctx,
		blobClient: blobClient,
	})
}

func (u *userDeletionEventsHandler) HandleDeleteUserEvent(payload ev.DeleteUserEvent) error {
	log.LoggerFromContext(u.ctx).WithField("userId", payload.UserID).Infof("Deleting blobs for user")
	return u.blobClient.DeleteAll(u.ctx, payload.UserID)
}
