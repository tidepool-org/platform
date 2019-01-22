package test

import "encoding/json"

func MarshalRequestBody(object interface{}) []byte {
	value, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	return append(value, []byte("\n")...)
}

func MarshalResponseBody(object interface{}) []byte {
	value, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	return value
}
