package api

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

type (
	API struct {
		Name     string
		Endpoint string
		Writer   io.Writer
		Verbose  bool
		client   *http.Client
	}

	Session struct {
		Token    string `json:"token"`
		IsServer bool   `json:"isServer"`
		UserID   string `json:"userId,omitempty"`
	}
)

const (
	ConfigDirectoryName = ".tidepool"
	SessionFileName     = "session"
)

func NewAPI(name string, endpoint string, proxy string) (*API, error) {
	if name == "" {
		return nil, errors.New("Name is missing")
	}
	if endpoint == "" {
		return nil, errors.New("End point is missing")
	}

	client := &http.Client{}
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("Error parsing proxy URL: %s", err.Error())
		}

		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}

	return &API{Name: name, Endpoint: endpoint, Writer: os.Stdout, Verbose: false, client: client}, nil
}

func (a *API) info(infos ...interface{}) {
	if a.Verbose {
		infos = append([]interface{}{"INFO:"}, infos...)
		fmt.Fprintln(a.Writer, infos...)
	}
}

func (a *API) IsSessionUserID(userID string) bool {
	if userID == "" {
		return true
	}

	sessionUserID, err := a.fetchSessionUserID()
	if err != nil {
		return false
	}

	return userID == sessionUserID
}

func (a *API) fetchSessionUserID() (string, error) {
	session, err := a.fetchSession()
	if err != nil {
		return "", err
	}

	userID := session.UserID
	if userID == "" {
		return "", errors.New("User id is missing")
	}

	return userID, nil
}

func (a *API) fetchSession() (*Session, error) {
	sessionFile, err := a.sessionFile()
	if err != nil {
		return nil, err
	}

	sessionBytes, err := ioutil.ReadFile(sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("Session not found")
		}
		return nil, err
	}

	session := &Session{}
	if err = json.Unmarshal(sessionBytes, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (a *API) storeSession(session *Session) error {
	if session == nil {
		return errors.New("Session is missing")
	}

	sessionBytes, err := json.Marshal(session)
	if err != nil {
		return err
	}

	sessionFile, err := a.sessionFile()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(sessionFile, sessionBytes, 0600)
}

func (a *API) destroySession() error {
	sessionFile, err := a.sessionFile()
	if err != nil {
		return err
	}

	if err = os.Remove(sessionFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (a *API) sessionFile() (string, error) {
	configDirectory, err := a.configDirectory()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDirectory, SessionFileName), nil
}

func (a *API) configDirectory() (string, error) {
	homeDirectory, err := homedir.Expand("~")
	if err != nil {
		return "", err
	}

	configDirectory := filepath.Join(homeDirectory, ConfigDirectoryName)
	if err = os.Mkdir(configDirectory, 0700); err != nil && !os.IsExist(err) {
		return "", err
	}

	return configDirectory, nil
}
