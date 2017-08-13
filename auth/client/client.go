package client

import (
	"net/http"
	"sync"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type Client struct {
	client             *client.Client
	name               string
	logger             log.Logger
	serverTokenSecret  string
	serverTokenTimeout time.Duration
	serverTokenMutex   sync.Mutex
	serverTokenSafe    string
	closingChannel     chan chan bool
}

const (
	TidepoolServerNameHeaderName   = "X-Tidepool-Server-Name"
	TidepoolServerSecretHeaderName = "X-Tidepool-Server-Secret"

	ServerTokenTimeoutOnFailureFirst = 1 * time.Second
	ServerTokenTimeoutOnFailureLast  = 60 * time.Second
)

func NewClient(config *Config, name string, logger log.Logger) (*Client, error) {
	if config == nil {
		return nil, errors.New("client", "config is missing")
	}
	if name == "" {
		return nil, errors.New("client", "name is missing")
	}
	if logger == nil {
		return nil, errors.New("client", "logger is missing")
	}

	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "client", "config is invalid")
	}

	clnt, err := client.NewClient(config.Config)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:             clnt,
		logger:             logger,
		name:               name,
		serverTokenSecret:  config.ServerTokenSecret,
		serverTokenTimeout: config.ServerTokenTimeout,
	}, nil
}

func (c *Client) Start() error {
	if c.closingChannel == nil {
		closingChannel := make(chan chan bool)
		c.closingChannel = closingChannel

		serverTokenTimeout := c.timeoutServerToken(0)

		go func() {
			for {
				timer := time.After(serverTokenTimeout)
				select {
				case closedChannel := <-closingChannel:
					closedChannel <- true
					close(closedChannel)
					return
				case <-timer:
					serverTokenTimeout = c.timeoutServerToken(serverTokenTimeout)
				}
			}
		}()
	}

	return nil
}

func (c *Client) Close() {
	if c.closingChannel != nil {
		closingChannel := c.closingChannel
		c.closingChannel = nil

		closedChannel := make(chan bool)
		closingChannel <- closedChannel
		close(closingChannel)
		<-closedChannel
	}
}

func (c *Client) ServerToken() (string, error) {
	if c.closingChannel == nil {
		return "", errors.New("client", "client is closed")
	}

	serverToken := c.serverToken()
	if serverToken == "" {
		return "", errors.New("client", "unable to obtain server token")
	}

	return serverToken, nil
}

func (c *Client) ValidateToken(ctx auth.Context, token string) (auth.Details, error) {
	if ctx == nil {
		return nil, errors.New("client", "context is missing")
	}
	if token == "" {
		return nil, errors.New("client", "token is missing")
	}

	if c.closingChannel == nil {
		return nil, errors.New("client", "client is closed")
	}

	ctx.Logger().Debug("Validating token")

	var result struct {
		IsServer bool
		UserID   string
	}
	if err := c.client.SendRequestWithServerToken(ctx, "GET", c.client.BuildURL("auth", "token", token), nil, &result); err != nil {
		return nil, err
	}

	if !result.IsServer && result.UserID == "" {
		return nil, errors.New("client", "user id is missing")
	}

	return &authDetails{
		token:    token,
		isServer: result.IsServer,
		userID:   result.UserID,
	}, nil
}

func (c *Client) GetStatus(ctx auth.Context) (*auth.Status, error) {
	sts := &auth.Status{}
	if err := c.client.SendRequestWithServerToken(ctx, "GET", c.client.BuildURL("status"), nil, sts); err != nil {
		return nil, err
	}

	return sts, nil
}

func (c *Client) timeoutServerToken(serverTokenTimeout time.Duration) time.Duration {
	if err := c.refreshServerToken(); err != nil {
		if serverTokenTimeout == 0 || serverTokenTimeout == c.serverTokenTimeout {
			serverTokenTimeout = ServerTokenTimeoutOnFailureFirst
		} else {
			serverTokenTimeout *= 2
			if serverTokenTimeout > ServerTokenTimeoutOnFailureLast {
				serverTokenTimeout = ServerTokenTimeoutOnFailureLast
			}
		}
		c.logger.WithError(err).WithField("retry", serverTokenTimeout.String()).Error("Unable to refresh server token; retrying")
	} else {
		serverTokenTimeout = c.serverTokenTimeout
	}

	return serverTokenTimeout
}

func (c *Client) refreshServerToken() error {
	c.logger.Debug("Refreshing server token")

	requestMethod := "POST"
	requestURL := c.client.BuildURL("auth", "serverlogin")
	request, err := http.NewRequest(requestMethod, requestURL, nil)
	if err != nil {
		return errors.Wrapf(err, "client", "unable to create new request for %s %s", requestMethod, requestURL)
	}

	request.Header.Add(TidepoolServerNameHeaderName, c.name)
	request.Header.Add(TidepoolServerSecretHeaderName, c.serverTokenSecret)

	response, err := c.client.HTTPClient().Do(request)
	if err != nil {
		return errors.Wrap(err, "client", "failure requesting new server token")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.Newf("client", "unexpected response status code %d while requesting new server token", response.StatusCode)
	}

	serverTokenHeader := response.Header.Get(auth.TidepoolAuthTokenHeaderName)
	if serverTokenHeader == "" {
		return errors.New("client", "server token is missing")
	}

	c.setServerToken(serverTokenHeader)

	return nil
}

func (c *Client) setServerToken(serverToken string) {
	c.serverTokenMutex.Lock()
	defer c.serverTokenMutex.Unlock()

	c.serverTokenSafe = serverToken
}

func (c *Client) serverToken() string {
	c.serverTokenMutex.Lock()
	defer c.serverTokenMutex.Unlock()

	return c.serverTokenSafe
}

type authDetails struct {
	token    string
	isServer bool
	userID   string
}

func (a *authDetails) Token() string {
	return a.token
}

func (a *authDetails) IsServer() bool {
	return a.isServer
}

func (a *authDetails) UserID() string {
	return a.userID
}
