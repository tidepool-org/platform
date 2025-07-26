package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type ResponseObject struct {
	Data   interface{} `json:"data,omitempty"`
	Errors []*Error    `json:"errors,omitempty"`
	Meta   *Meta       `json:"meta,omitempty"`
}

type ResponseArray struct {
	Data   []interface{} `json:"data,omitempty"`
	Errors []*Error      `json:"errors,omitempty"`
	Meta   *Meta         `json:"meta,omitempty"`
}

type Error struct {
	Code   string      `json:"code,omitempty"`
	Title  string      `json:"title,omitempty"`
	Detail string      `json:"detail,omitempty"`
	Status int         `json:"status,string,omitempty"`
	Source *Source     `json:"source,omitempty"`
	Meta   interface{} `json:"meta,omitempty"`
}

type Source struct {
	Parameter string `json:"parameter,omitempty"`
	Pointer   string `json:"pointer,omitempty"`
}

type Meta struct {
	Trace *Trace `json:"trace,omitempty"`
}

type Trace struct {
	Request string `json:"request,omitempty"`
	Session string `json:"session,omitempty"`
}

func (a *API) asEmpty(responseBody io.Reader, err error) error {
	responseBuffer, err := a.asBuffer(responseBody, err)
	if err != nil {
		return err
	}

	if responseBuffer.Len() > 0 {
		a.info("Ignoring non-empty response body.")
	}

	return nil
}

func (a *API) asBuffer(responseBody io.Reader, err error) (*bytes.Buffer, error) {
	if err != nil {
		return nil, err
	}

	responseBuffer := &bytes.Buffer{}
	if responseBody != nil {
		if _, err = responseBuffer.ReadFrom(responseBody); err != nil {
			return nil, fmt.Errorf("Error reading response body: %s", err.Error())
		}
	}

	if length := responseBuffer.Len(); length != 0 {
		a.info("Response body length:", length)
	} else {
		a.info("Response body empty.")
	}

	return responseBuffer, nil
}

func (a *API) asBytes(responseBody io.Reader, err error) ([]byte, error) {
	responseBuffer, err := a.asBuffer(responseBody, err)
	if err != nil {
		return nil, err
	}

	return responseBuffer.Bytes(), nil
}

func (a *API) asString(responseBody io.Reader, err error) (string, error) {
	responseBuffer, err := a.asBuffer(responseBody, err)
	if err != nil {
		return "", err
	}

	responseString := responseBuffer.String()
	if len(responseString) > 0 {
		a.info("Response body:", responseString)
	}

	return responseString, nil
}

func (a *API) asObject(responseBody io.Reader, err error) (interface{}, error) {
	responseString, err := a.asString(responseBody, err)
	if err != nil {
		return nil, err
	}

	var responseObject interface{}
	if len(responseString) > 0 {
		if err = json.Unmarshal([]byte(responseString), &responseObject); err != nil {
			return nil, fmt.Errorf("Error decoding JSON object from response body: %s", err.Error())
		}
	}

	return responseObject, nil
}

func (a *API) asStringMap(responseBody io.Reader, err error) (map[string]interface{}, error) {
	responseString, err := a.asString(responseBody, err)
	if err != nil {
		return nil, err
	}

	responseStringMap := map[string]interface{}{}
	if len(responseString) > 0 {
		if err = json.Unmarshal([]byte(responseString), &responseStringMap); err != nil {
			return nil, fmt.Errorf("Error decoding JSON string map from response body: %s", err.Error())
		}
	}

	return responseStringMap, nil
}

func (a *API) asArray(responseBody io.Reader, err error) ([]interface{}, error) {
	responseString, err := a.asString(responseBody, err)
	if err != nil {
		return nil, err
	}

	responseArray := []interface{}{}
	if len(responseString) > 0 {
		if err = json.Unmarshal([]byte(responseString), &responseArray); err != nil {
			return nil, fmt.Errorf("Error decoding JSON array from response body: %s", err.Error())
		}
	}

	return responseArray, nil
}

func (a *API) asResponseObject(responseBody io.Reader, err error) (*ResponseObject, error) {
	responseString, err := a.asString(responseBody, err)
	if err != nil {
		return nil, err
	}

	var responseObject *ResponseObject
	if len(responseString) > 0 {
		responseObject = &ResponseObject{}
		if err = json.Unmarshal([]byte(responseString), responseObject); err != nil {
			return nil, fmt.Errorf("Error decoding JSON response array from response body: %s", err.Error())
		}
	}

	return responseObject, nil
}

func (a *API) asResponseArray(responseBody io.Reader, err error) (*ResponseArray, error) {
	responseString, err := a.asString(responseBody, err)
	if err != nil {
		return nil, err
	}

	var responseArray *ResponseArray
	if len(responseString) > 0 {
		responseArray = &ResponseArray{}
		if err = json.Unmarshal([]byte(responseString), responseArray); err != nil {
			return nil, fmt.Errorf("Error decoding JSON response array from response body: %s", err.Error())
		}
	}

	return responseArray, nil
}
