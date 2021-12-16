package hakken

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tidepool-org/go-common/clients/disc"
	"github.com/tidepool-org/go-common/errors"
	"net/http"
)

type coordinatorClient struct {
	Coordinator
}

func (client *coordinatorClient) getCoordinators() ([]Coordinator, error) {
	url := fmt.Sprintf("%s/v1/coordinator", client.URL.String())
	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "Problem when looking up coordinators[%s]", url)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.Newf("Unknown response code[%d] from [%s]", res.StatusCode, url)
	}

	var retVal []Coordinator
	if err := json.NewDecoder(res.Body).Decode(&retVal); err != nil {
		return nil, errors.Wrapf(err, "Error parsing JSON results from [%s]", url)
	}
	return retVal, nil
}

func (client *coordinatorClient) getListings(service string) ([]disc.ServiceListing, error) {
	url := fmt.Sprintf("%s/v1/listings/%s", client.URL.String(), service)
	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "Problem when looking up listings at url[%s]", url)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.Newf("Unknown response code[%d] from url[%s]", res.StatusCode, url)
	}

	var retVal []disc.ServiceListing
	if err := json.NewDecoder(res.Body).Decode(&retVal); err != nil {
		return nil, errors.Wrapf(err, "Error parsing JSON results from url[%s]", url)
	}
	return retVal, nil
}

func (client *coordinatorClient) listingHearbeat(sl *disc.ServiceListing) error {
	url := fmt.Sprintf("%s/v1/listings?heartbeat=true", client.URL.String())

	marshaled, err := json.Marshal(sl)
	if err != nil {
		return err
	}

	res, err := http.Post(url, "application/json", bytes.NewReader(marshaled))
	if err != nil {
		return errors.Wrapf(err, "Problem when updating heartbeat for service[%s] at [%s]", sl.Service, url)
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		return errors.Newf("Unknown response code[%d] from url[%s]", res.StatusCode, url)
	}
	return nil
}
