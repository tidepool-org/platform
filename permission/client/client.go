package client

import (
	"strings"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
)

type Client struct {
	client                 *platform.Client
	permissionClientConfig httpClientConfig
}

type httpClientConfig struct {
	clientType string
	urlPrefix  string
	httpMethod string
}

type CoastguardRequestBody struct {
	Input struct {
		Request struct {
			Headers  map[string]string `json:"headers"`
			Host     string            `json:"host"`
			Method   string            `json:"method"`
			Path     string            `json:"path"`
			Query    string            `json:"query"`
			Fragment string            `json:"fragment"`
			Protocol string            `json:"protocol"`
			Service  string            `json:"service"`
		} `json:"request"`
		Data struct {
			TargetUserID string `json:"targetUserId"`
		} `json:"data"`
	} `json:"input"`
}

type CoastguardResponseBody struct {
	Result struct {
		Authorized bool   `json:"authorized"`
		Route      string `json:"route"`
	} `json:"result"`
}

var (
	permissionClientCoastguard = httpClientConfig{
		clientType: "coastguard",
		urlPrefix:  "v1/data/backloops/access",
		httpMethod: "POST",
	}
)

func New(config *platform.Config) (*Client, error) {
	clnt, err := platform.NewClient(config, platform.AuthorizeAsService)
	if err != nil {
		return nil, err
	}
	return &Client{
		client:                 clnt,
		permissionClientConfig: permissionClientCoastguard,
	}, nil
}

func (c *Client) GetUserPermissions(req *rest.Request, targetUserID string) (bool, error) {
	ctx := req.Context()
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	details := request.DetailsFromContext(ctx)
	if details == nil {
		return false, request.ErrorUnauthenticated()
	}
	if details.IsService() {
		return true, nil
	}
	requestUserID := details.UserID()
	if requestUserID == "" {
		return false, errors.New("request user id is missing")
	}
	if targetUserID == "" {
		return false, errors.New("target user id is missing")
	}

	if requestUserID == targetUserID {
		return true, nil
	}

	authConfig := c.permissionClientConfig
	if authConfig.clientType == "coastguard" {
		urlParts := strings.Split(authConfig.urlPrefix, "/")
		url := c.client.ConstructURL(urlParts...)

		coastguardResponse := CoastguardResponseBody{}
		requestBody := formatRequest(req, targetUserID)
		if err := c.client.RequestData(ctx, authConfig.httpMethod, url, nil, &requestBody, &coastguardResponse); err != nil {
			return false, err
		}
		return coastguardResponse.Result.Authorized, nil
	}
	return false, nil
}

func formatRequest(req *rest.Request, targetUserID string) CoastguardRequestBody {
	var opaReq CoastguardRequestBody
	url := *req.URL
	headers := make(map[string]string)
	for k := range req.Header {
		headers[strings.ToLower(k)] = req.Header.Get(k)
	}
	opaReq.Input.Request.Headers = headers
	opaReq.Input.Request.Method = req.Method
	opaReq.Input.Request.Protocol = req.Proto
	opaReq.Input.Request.Host = req.Host
	opaReq.Input.Request.Path = url.Path
	opaReq.Input.Request.Query = url.RawQuery
	opaReq.Input.Request.Service = "platform"
	opaReq.Input.Data.TargetUserID = targetUserID
	return opaReq
}
