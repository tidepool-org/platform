package work

import (
	"context"

	"github.com/tidepool-org/platform/auth"
	dataWork "github.com/tidepool-org/platform/data/work"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/oura"
	ouraData "github.com/tidepool-org/platform/oura/data"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
)

type Client interface {
	ListSubcriptions(ctx context.Context) ([]*ouraWebhook.Subscription, error)
	CreateSubscription(ctx context.Context, create *ouraWebhook.CreateSubscription) (*ouraWebhook.Subscription, error)
	RenewSubscription(ctx context.Context, id string) (*ouraWebhook.Subscription, error)
	DeleteSubscription(ctx context.Context, id string) error

	RevokeOAuthToken(ctx context.Context, oauthToken *auth.OAuthToken) error

	GetPersonalInfo(ctx context.Context, tokenSource oauth.TokenSource) (*ouraData.PersonalInfo, error)

	GetDatum(ctx context.Context, dataType string, dataID string, tokenSource oauth.TokenSource) (*oura.Datum, error)
	GetData(ctx context.Context, dataType string, timeRange dataWork.TimeRange, tokenSource oauth.TokenSource) (oura.Data, error)
}
