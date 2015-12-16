package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

type Builder interface {
	Build(raw []byte) (interface{}, error)
}

type TypeBuilder struct{}

func NewTypeBuilder() Builder {
	return &TypeBuilder{}
}

func (this *TypeBuilder) Build(raw []byte) (interface{}, error) {

	const typeField = "type"

	var data map[string]interface{}
	err := json.Unmarshal(raw, &data)
	if err != nil {
		log.Println("error unmarshelling type", err.Error())
		return nil, errors.New(fmt.Sprintf("sorry but we do anything with %s", string(raw)))
	}

	if data[typeField] != nil {

		if strings.ToLower(data[typeField].(string)) == "basal" {
			return BuildBasal(data)
		} else if strings.ToLower(data[typeField].(string)) == "deviceevent" {
			return BuildDeviceEvent(data)
		}
		return nil, errors.New(fmt.Sprintf("sorry but we can't deal with `type` %s", data[typeField].(string)))
	}

	return nil, errors.New(fmt.Sprintf("the data had no `type` specified %s", data))

}
