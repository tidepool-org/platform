package test

import "encoding/json"

func MarshalRequestBody(object interface{}) ([]byte, error) {
	bites, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}
	return append(bites, []byte("\n")...), nil
}

func MarshalResponseBody(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}
