package customerio

import "github.com/tidepool-org/platform/log"

type Client struct {
	config Config
	logger log.Logger
}

type Config struct {
	AppAPIBaseURL   string `envconfig:"TIDEPOOL_CUSTOMERIO_APP_API_BASE_URL" default:"https://api.customer.io"`
	AppAPIKey       string `envconfig:"TIDEPOOL_CUSTOMERIO_APP_API_KEY"`
	SegmentID       string `envconfig:"TIDEPOOL_CUSTOMERIO_SEGMENT_ID"`
	SiteID          string `envconfig:"TIDEPOOL_CUSTOMERIO_SITE_ID"`
	TrackAPIBaseURL string `envconfig:"TIDEPOOL_CUSTOMERIO_TRACK_API_BASE_URL" default:"https://track.customer.io"`
	TrackAPIKey     string `envconfig:"TIDEPOOL_CUSTOMERIO_TRACK_API_KEY"`
}

func NewClient(config Config, logger log.Logger) (*Client, error) {
	return &Client{
		config: config,
		logger: logger,
	}, nil
}
