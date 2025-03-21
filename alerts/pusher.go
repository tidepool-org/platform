package alerts

import (
	"context"

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/push"
)

// Pusher is a service-agnostic interface for sending push notifications.
type Pusher interface {
	// Push a notification to a device.
	Push(context.Context, *devicetokens.DeviceToken, *push.Notification) error
}

// ToPushNotification converts Notification to push.Notification.
func ToPushNotification(notification *Notification) *push.Notification {
	return &push.Notification{
		Message: notification.Message,
	}
}

type cpaPusherEnvconfig struct {
	// SigningKey is the raw token signing key received from Apple (.p8 file containing
	// PEM-encoded private key)
	//
	// https://developer.apple.com/documentation/usernotifications/sending-notification-requests-to-apns
	SigningKey []byte `envconfig:"TIDEPOOL_CARE_PARTNER_ALERTS_APNS_SIGNING_KEY" required:"true"`
	KeyID      string `envconfig:"TIDEPOOL_CARE_PARTNER_ALERTS_APNS_KEY_ID" required:"true"`
	BundleID   string `envconfig:"TIDEPOOL_CARE_PARTNER_ALERTS_APNS_BUNDLE_ID" required:"true"`
	TeamID     string `envconfig:"TIDEPOOL_CARE_PARTNER_ALERTS_APNS_TEAM_ID" required:"true"`
}

// NewPusher handles the loading of care partner configuration for push notifications.
func NewPusher() (*push.APNSPusher, error) {
	config, err := loadPusherViaEnvconfig()
	if err != nil {
		return nil, errors.Wrap(err, "unable to care partner pusher config")
	}

	client, err := push.NewAPNS2Client(config.SigningKey, config.KeyID, config.TeamID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create care partner pusher client")
	}

	return push.NewAPNSPusher(client, config.BundleID), nil
}

func loadPusherViaEnvconfig() (*cpaPusherEnvconfig, error) {
	c := &cpaPusherEnvconfig{}
	if err := envconfig.Process("", c); err != nil {
		return nil, errors.Wrap(err, "Unable to process APNs pusher config")
	}

	// envconfig's "required" tag won't error on values that are defined but empty, so
	// manually check

	if len(c.SigningKey) == 0 {
		return nil, errors.New("Unable to build APNSPusherConfig: APNs signing key is blank")
	}

	if c.BundleID == "" {
		return nil, errors.New("Unable to build APNSPusherConfig: bundleID is blank")
	}

	if c.KeyID == "" {
		return nil, errors.New("Unable to build APNSPusherConfig: keyID is blank")
	}

	if c.TeamID == "" {
		return nil, errors.New("Unable to build APNSPusherConfig: teamID is blank")
	}

	return c, nil
}
