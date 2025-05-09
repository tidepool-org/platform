package dexcom

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/oauth"
)

type Client interface {
	GetAlerts(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*AlertsResponse, error)
	GetCalibrations(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*CalibrationsResponse, error)
	GetDataRange(ctx context.Context, lastSyncTime *time.Time, tokenSource oauth.TokenSource) (*DataRangesResponse, error)
	GetDevices(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*DevicesResponse, error)
	GetEGVs(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*EGVsResponse, error)
	GetEvents(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*EventsResponse, error)
}
