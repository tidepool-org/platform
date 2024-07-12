// Package push provides clients for sending mobile device push notifications.
package push

import (
	"context"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

// Notification models a provider-independent push notification.
type Notification struct {
	Message string
}

// String implements fmt.Stringer.
func (n Notification) String() string {
	return n.Message
}

func FromAlertsNotification(notification *alerts.Notification) *Notification {
	return &Notification{
		Message: notification.Message,
	}
}

// APNSPusher implements push notifications via Apple APNs.
type APNSPusher struct {
	BundleID string

	client   APNS2Client
	clientMu sync.Mutex
}

// NewAPNSPusher creates a Pusher for sending device notifications via Apple's
// APNs.
func NewAPNSPusher(client APNS2Client, bundleID string) *APNSPusher {
	return &APNSPusher{
		BundleID: bundleID,
		client:   client,
	}
}

// NewAPNSPusherFromKeyData creates an APNSPusher for sending device
// notifications via Apple's APNs.
//
// The signingKey is the raw token signing key received from Apple (.p8 file
// containing PEM-encoded private key), along with its respective team id, key
// id, and application bundle id.
//
// https://developer.apple.com/documentation/usernotifications/sending-notification-requests-to-apns
func NewAPNSPusherFromKeyData(signingKey []byte, keyID, teamID, bundleID string) (*APNSPusher, error) {
	if len(signingKey) == 0 {
		return nil, errors.New("Unable to build APNSPusher: APNs signing key is blank")
	}

	if bundleID == "" {
		return nil, errors.New("Unable to build APNSPusher: bundleID is blank")
	}

	if keyID == "" {
		return nil, errors.New("Unable to build APNSPusher: keyID is blank")
	}

	if teamID == "" {
		return nil, errors.New("Unable to build APNSPusher: teamID is blank")
	}

	authKey, err := token.AuthKeyFromBytes(signingKey)
	if err != nil {
		return nil, err
	}
	token := &token.Token{
		AuthKey: authKey,
		KeyID:   keyID,
		TeamID:  teamID,
	}
	client := &apns2Client{Client: apns2.NewTokenClient(token)}
	return NewAPNSPusher(client, bundleID), nil
}

func (p *APNSPusher) Push(ctx context.Context, deviceToken *devicetokens.DeviceToken,
	notification *Notification) error {

	if deviceToken.Apple == nil {
		return errors.New("Unable to push notification: APNSPusher can only use Apple device tokens but the Apple token is nil")
	}

	hexToken := hex.EncodeToString(deviceToken.Apple.Token)
	appleNotification := p.buildAppleNotification(hexToken, notification)
	resp, err := p.safePush(ctx, deviceToken.Apple.Environment, appleNotification)
	if err != nil {
		return errors.Wrap(err, "Unable to push notification")
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Newf("Unable to push notification: APNs returned non-200 status: %d, %s", resp.StatusCode, resp.Reason)
	}
	if logger := log.LoggerFromContext(ctx); logger != nil {
		logger.WithFields(log.Fields{
			"apnsID": resp.ApnsID,
		}).Info("notification pushed")
	}

	return nil
}

// safePush guards the environment setup and push method with a mutex.
//
// This prevents the environment from being changed out from under
// you. Unlikely, but better safe than sorry.
func (p *APNSPusher) safePush(ctx context.Context, env string, notification *apns2.Notification) (
	*apns2.Response, error) {

	p.clientMu.Lock()
	defer p.clientMu.Unlock()
	if env == devicetokens.AppleEnvProduction {
		p.client.Production()
	} else {
		p.client.Development()
	}
	return p.client.PushWithContext(ctx, notification)
}

func (p *APNSPusher) buildAppleNotification(hexToken string, notification *Notification) *apns2.Notification {
	payload := payload.NewPayload().
		Alert(notification.Message).
		AlertBody(notification.Message)
	return &apns2.Notification{
		DeviceToken: hexToken,
		Payload:     payload,
		Topic:       p.BundleID,
	}
}

// APNS2Client abstracts the apns2 library for easier testing.
type APNS2Client interface {
	Development() APNS2Client
	Production() APNS2Client
	PushWithContext(apns2.Context, *apns2.Notification) (*apns2.Response, error)
}

// apns2Client adapts the apns2.Client to APNS2Client so it can be replaced for testing.
type apns2Client struct {
	*apns2.Client
}

func (c apns2Client) Development() APNS2Client {
	d := c.Client.Development()
	return &apns2Client{Client: d}
}

func (c apns2Client) Production() APNS2Client {
	p := c.Client.Production()
	return &apns2Client{Client: p}
}
