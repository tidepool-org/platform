package clients

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/tidepool-org/go-common/clients/disc"
	"github.com/tidepool-org/go-common/clients/status"
)

type (
	Seagull interface {
		// Retrieves arbitrary collection information from metadata
		//
		// userID -- the Tidepool-assigned userId
		// hashName -- the name of what we are trying to get
		// token -- a server token or the user token
		GetPrivatePair(userID, hashName, token string) *PrivatePair
		// Retrieves arbitrary collection information from metadata
		//
		// userID -- the Tidepool-assigned userId
		// collectionName -- the collection being retrieved
		// token -- a server token or the user token
		// v - the interface to return the value in
		GetCollection(userID, collectionName, token string, v interface{}) error
		// Updates arbitrary collection information
		//
		// userID -- the Tidepool-assigned userId
		// collectionName -- the collection being retrieved
		// token -- a server token or the user token
		// v - the interface to return the value in
		UpdateCollection(userID, collectionName, token string, v interface{}) error
	}

	SeagullClient struct {
		httpClient *http.Client    // store a reference to the http client so we can reuse it
		hostGetter disc.HostGetter // The getter that provides the host to talk to for the client
	}

	seagullClientBuilder struct {
		httpClient *http.Client
		hostGetter disc.HostGetter
	}

	PrivatePair struct {
		ID    string
		Value string
	}
)

func NewSeagullClientBuilder() *seagullClientBuilder {
	return &seagullClientBuilder{}
}

func (b *seagullClientBuilder) WithHttpClient(httpClient *http.Client) *seagullClientBuilder {
	b.httpClient = httpClient
	return b
}

func (b *seagullClientBuilder) WithHostGetter(hostGetter disc.HostGetter) *seagullClientBuilder {
	b.hostGetter = hostGetter
	return b
}

func (b *seagullClientBuilder) Build() *SeagullClient {
	if b.httpClient == nil {
		panic("seagullClient requires an httpClient to be set")
	}
	if b.hostGetter == nil {
		panic("seagullClient requires a hostGetter to be set")
	}
	return &SeagullClient{
		httpClient: b.httpClient,
		hostGetter: b.hostGetter,
	}
}

func (client *SeagullClient) GetPrivatePair(userID, hashName, token string) *PrivatePair {
	host := client.getHost()
	if host == nil {
		return nil
	}
	host.Path = path.Join(host.Path, userID, "private", hashName)

	req, _ := http.NewRequest("GET", host.String(), nil)
	req.Header.Add("x-tidepool-session-token", token)

	log.Println(req)
	res, err := client.httpClient.Do(req)
	if err != nil {
		log.Printf("Problem when looking up private pair for userID[%s]. %s", userID, err)
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("Unknown response code[%v] from service[%v]", res.StatusCode, req.URL)
		return nil
	}

	var retVal PrivatePair
	if err := json.NewDecoder(res.Body).Decode(&retVal); err != nil {
		log.Println("Error parsing JSON results", err)
		return nil
	}
	return &retVal
}

func (client *SeagullClient) GetCollection(userID, collectionName, token string, v interface{}) error {
	host := client.getHost()
	if host == nil {
		return nil
	}
	host.Path = path.Join(host.Path, userID, collectionName)

	req, _ := http.NewRequest("GET", host.String(), nil)
	req.Header.Add("x-tidepool-session-token", token)

	res, err := client.httpClient.Do(req)
	if err != nil {
		log.Printf("Problem when looking up collection for userID[%s]. %s", userID, err)
		return err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
			log.Println("Error parsing JSON results", err)
			return err
		}
		return nil
	case http.StatusNotFound:
		log.Printf("No [%s] collection found for [%s]", collectionName, userID)
		return nil
	default:
		return &status.StatusError{status.NewStatusf(res.StatusCode, "Unknown response code from service[%s]", req.URL)}
	}

}

func (client *SeagullClient) UpdateCollection(userID, collectionName, token string, v interface{}) error {
	host := client.getHost()
	if host == nil {
		return nil
	}
	host.Path = path.Join(host.Path, userID, collectionName)

	body, err := json.Marshal(v)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("POST", host.String(), bytes.NewBuffer(body))
	req.Header.Add("content-type", "application/json")
	req.Header.Add("x-tidepool-session-token", token)

	res, err := client.httpClient.Do(req)
	if err != nil {
		log.Printf("Problem when looking up collection for userID[%s]. %s", userID, err)
		return err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		log.Printf("No [%s] collection found for [%s]", collectionName, userID)
		return nil
	default:
		return &status.StatusError{status.NewStatusf(res.StatusCode, "Unknown response code from service[%s]", req.URL)}
	}

}

func (client *SeagullClient) getHost() *url.URL {
	if hostArr := client.hostGetter.HostGet(); len(hostArr) > 0 {
		cpy := new(url.URL)
		*cpy = hostArr[0]
		return cpy
	} else {
		return nil
	}
}
