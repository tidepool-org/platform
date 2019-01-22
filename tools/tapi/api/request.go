package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tidepool-org/platform/auth"
)

const (
	TidepoolServerName   = "X-Tidepool-Server-Name"
	TidepoolServerSecret = "X-Tidepool-Server-Secret"
)

type (
	requestFunc   func(request *http.Request) error
	requestFuncs  []requestFunc
	responseFunc  func(response *http.Response) error
	responseFuncs []responseFunc
)

func (a *API) joinPaths(paths ...string) string {
	join := ""
	for _, path := range paths {
		if path != "" {
			join = fmt.Sprintf("%s/%s", strings.TrimRight(join, "/"), strings.TrimLeft(path, "/"))
		}
	}
	return join
}

func (a *API) addQuery(path string, query map[string]string) string {
	var separator string
	if strings.Contains(path, "?") {
		separator = "&"
	} else {
		separator = "?"
	}
	for key, value := range query {
		path = fmt.Sprintf("%s%s%s=%s", path, separator, url.QueryEscape(key), url.QueryEscape(value))
		separator = "&"
	}
	return path
}

func (a *API) request(method string, path string, requestCallbacks requestFuncs, responseCallbacks responseFuncs) (io.Reader, error) {
	if method == "" {
		return nil, errors.New("Method is missing")
	}
	if path == "" {
		return nil, errors.New("Path is missing")
	}

	urlString := fmt.Sprintf("%s/%s", strings.TrimRight(a.Endpoint, "/"), strings.TrimLeft(path, "/"))

	request, err := http.NewRequest(method, urlString, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating new request: %s", err.Error())
	}

	for _, requestCallback := range requestCallbacks {
		if err = requestCallback(request); err != nil {
			return nil, err
		}
	}

	a.info("Request method:", request.Method)
	a.info("Request url:", request.URL.String())

	response, err := a.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Error sending session request: %s", err.Error())
	}

	for _, responseCallback := range responseCallbacks {
		if err = responseCallback(response); err != nil {
			return nil, err
		}
	}

	return response.Body, nil
}

func (a *API) addServerSecret() requestFunc {
	return func(request *http.Request) error {
		serverSecret := os.Getenv("TIDEPOOL_USER_CLIENT_SERVERTOKENSECRET")
		if serverSecret == "" {
			return errors.New("Environment variable TIDEPOOL_USER_CLIENT_SERVERTOKENSECRET not exported")
		}

		a.info("Server secret found.")

		request.Header.Add(TidepoolServerName, a.Name)
		request.Header.Add(TidepoolServerSecret, serverSecret)
		return nil
	}
}

func (a *API) addBasicAuthorization(username string, password string) requestFunc {
	return func(request *http.Request) error {
		request.SetBasicAuth(username, password)
		return nil
	}
}

func (a *API) addSessionToken() requestFunc {
	return func(request *http.Request) error {
		session, err := a.fetchSession()
		if err != nil {
			return err
		}

		a.info("Session token found:", session.Token)

		request.Header.Add(auth.TidepoolSessionTokenHeaderKey, session.Token)
		return nil
	}
}

func (a *API) addObjectBody(requestBody interface{}) requestFunc {
	return func(request *http.Request) error {
		if requestBody == nil {
			return errors.New("Body is missing")
		}

		bites, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("Error encoding request body to JSON: %s", err.Error())
		}

		return a.addStringBody(string(bites), "application/json")(request)
	}
}

func (a *API) addStringBody(requestBody string, contentType string) requestFunc {
	return func(request *http.Request) error {
		if len(requestBody) > 0 {
			a.info("Request body:", requestBody)
		}

		return a.addBytesBody([]byte(requestBody), contentType)(request)
	}
}

func (a *API) addBufferBody(requestBody *bytes.Buffer, contentType string) requestFunc {
	return func(request *http.Request) error {
		if requestBody == nil {
			return errors.New("Body is missing")
		}

		return a.addBytesBody(requestBody.Bytes(), contentType)(request)
	}
}

func (a *API) addBytesBody(requestBody []byte, contentType string) requestFunc {
	return func(request *http.Request) error {
		if requestBody == nil {
			return errors.New("Body is missing")
		}

		a.info("Request body length:", len(requestBody))

		return a.addBody(bytes.NewReader(requestBody), contentType)(request)
	}
}

func (a *API) addBody(requestBody io.Reader, contentType string) requestFunc {
	return func(request *http.Request) error {
		if requestBody == nil {
			return errors.New("Body is missing")
		}
		if contentType == "" {
			return errors.New("Content type is missing")
		}
		if request.Body != nil {
			return errors.New("Request body already specified")
		}
		if request.ContentLength != 0 {
			return errors.New("Request content length already specified")
		}
		if request.Header.Get("Content-Type") != "" {
			return errors.New("Request Content-Type header already specified")
		}

		if readCloser, ok := requestBody.(io.ReadCloser); ok {
			request.Body = readCloser
		} else {
			request.Body = ioutil.NopCloser(requestBody)
		}

		switch t := requestBody.(type) {
		case *bytes.Buffer:
			request.ContentLength = int64(t.Len())
		case *bytes.Reader:
			request.ContentLength = int64(t.Len())
		case *strings.Reader:
			request.ContentLength = int64(t.Len())
		}

		request.Header.Set("Content-Type", contentType)
		return nil
	}
}

func (a *API) expectStatusCode(statusCode int) responseFunc {
	return func(response *http.Response) error {
		a.info("Response status code:", response.StatusCode)
		if response.StatusCode != statusCode {
			responseString, _ := a.asString(response.Body, nil)
			return fmt.Errorf("Unexpected response status code from server: [%d] %s", response.StatusCode, responseString)
		}
		return nil
	}
}

func (a *API) storeTidepoolSession(isServer bool) responseFunc {
	return func(response *http.Response) error {
		token := response.Header.Get(auth.TidepoolSessionTokenHeaderKey)
		if token == "" {
			return errors.New("No session token included in response")
		}

		session := &Session{
			Token:    token,
			IsServer: isServer,
		}

		if !isServer {
			bodyBytes, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return err
			}
			response.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

			user, err := a.asUser(bytes.NewBuffer(bodyBytes), nil)
			if err != nil {
				return err
			}
			session.UserID = user.ID
		}

		if err := a.storeSession(session); err != nil {
			return err
		}

		a.info("Session stored.")
		return nil
	}
}

func (a *API) updateTidepoolSession() responseFunc {
	return func(response *http.Response) error {
		token := response.Header.Get(auth.TidepoolSessionTokenHeaderKey)
		if token == "" {
			return errors.New("No session token included in response")
		}

		session, err := a.fetchSession()
		if err != nil {
			return err
		}

		session.Token = token

		if err = a.storeSession(session); err != nil {
			return err
		}

		a.info("Session updated.")
		return nil
	}
}

func (a *API) destroyTidepoolSession() responseFunc {
	return func(response *http.Response) error {
		if err := a.destroySession(); err != nil {
			return err
		}

		a.info("Session destroyed.")
		return nil
	}
}
