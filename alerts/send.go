package alerts

import (
	"context"
	"net/http"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"

	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/errors"
)

// Pusher abstracts the underlying push notification mechanism.
type Pusher interface {
	// Push a notification to a device.
	//
	// deviceToken should uniquely identify a device, while Notification
	// abstracts the provider-specific notification payload.
	Push(ctx context.Context, deviceToken devicetokens.DeviceToken, notification *Notification) error
}

// Notification models a provider-independent push notification.
type Notification struct {
	Title   string
	Message string
}

// APNSPusher implements push notifications via Apple APNs.
type APNSPusher struct {
	serviceToken *token.Token
	bundleID     string
}

// NewAPNSPusher creates a Pusher for sending device notifications via Apple's
// APNs.
//
// The signingKey is the raw token signing key received from Apple (.p8 file
// containing PEM-encoded private key), along with its respective team id, key
// id, and application bundle id.
//
// https://developer.apple.com/documentation/usernotifications/sending-notification-requests-to-apns
func NewAPNSPusher(signingKey []byte, keyID, teamID, bundleID string) (*APNSPusher, error) {
	authKey, err := token.AuthKeyFromBytes(signingKey)
	if err != nil {
		return nil, err
	}
	token := &token.Token{
		AuthKey: authKey,
		KeyID:   keyID,
		TeamID:  teamID,
	}
	return &APNSPusher{
		bundleID:     bundleID,
		serviceToken: token,
	}, nil
}

func (p *APNSPusher) Push(ctx context.Context, deviceToken devicetokens.DeviceToken, note *Notification) error {
	if deviceToken.Apple == nil {
		return errors.New("Unable to push notification: APNSPusher can only use Apple device tokens but the Apple token is nil")
	}

	// TODO: look at the clientmanager package in apns2
	client := apns2.NewTokenClient(p.serviceToken)
	if deviceToken.Apple.Environment == "production" {
		client = client.Production()
	} else {
		client = client.Development()
	}

	appleNote := p.buildAppleNotification(deviceToken.Apple, note)
	resp, err := client.PushWithContext(ctx, appleNote)
	if err != nil {
		return errors.Wrap(err, "Unable to push notification")
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Newf("Unable to push notification: APNs returned non-200 status: %d, %s", resp.StatusCode, resp.Reason)
	}

	return nil
}

func (p *APNSPusher) buildAppleNotification(deviceToken *devicetokens.AppleDeviceToken, note *Notification) *apns2.Notification {
	payload := payload.NewPayload().
		Alert(note.Message).
		AlertBody(note.Message)
	return &apns2.Notification{
		DeviceToken: string(deviceToken.Token),
		Payload:     payload,
		Topic:       p.bundleID,
	}
}
