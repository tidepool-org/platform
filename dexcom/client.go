package dexcom

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/oauth"
)

type Client interface {
	GetCalibrations(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*CalibrationsResponse, error)
	GetDevices(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*DevicesResponse, error)
	GetEGVs(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*EGVsResponse, error)
	GetEvents(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*EventsResponse, error)
}
