package customerio

import "github.com/tidepool-org/platform/log"

const appAPIBaseURL = "https://api.customer.io"
const trackAPIBaseURL = "https://track.customer.io/api/"

type Client struct {
	appAPIKey       string
	trackAPIKey     string
	siteID          string
	appAPIBaseURL   string
	trackAPIBaseURL string
	logger          log.Logger
}

type Config struct {
	AppAPIKey   string `envconfig:"TIDEPOOL_CUSTOMERIO_APP_API_KEY"`
	TrackAPIKey string `envconfig:"TIDEPOOL_CUSTOMERIO_TRACK_API_KEY"`
	SiteID      string `envconfig:"TIDEPOOL_CUSTOMERIO_SITE_ID"`
	SegmentID   string `envconfig:"TIDEPOOL_CUSTOMERIO_SEGMENT_ID"`
}

func NewClient(config Config, logger log.Logger) (*Client, error) {
	return &Client{
		appAPIKey:       config.AppAPIKey,
		trackAPIKey:     config.TrackAPIKey,
		siteID:          config.SiteID,
		appAPIBaseURL:   appAPIBaseURL,
		trackAPIBaseURL: trackAPIBaseURL,
		logger:          logger,
	}, nil
}
