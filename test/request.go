package test

import "encoding/json"

func MarshalRequestBody(object interface{}) ([]byte, error) {
	bytes, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}
	return append(bytes, []byte("\n")...), nil
}
