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

	const (
		type_field        = "type"
		basal_type        = "basal"
		device_event_type = "deviceevent"
	)

	var data map[string]interface{}
	err := json.Unmarshal(raw, &data)
	if err != nil {
		log.Println("error doing an unmarshal", err.Error())
		return nil, errors.New(fmt.Sprintf("sorry but we do anything with %s", string(raw)))
	}

	if data[type_field] != nil {

		if strings.ToLower(data[type_field].(string)) == basal_type {
			return BuildBasal(data)
		} else if strings.ToLower(data[type_field].(string)) == device_event_type {
			return BuildDeviceEvent(data)
		}
		return nil, errors.New(fmt.Sprintf("sorry but we can't deal with `type` %s", data[type_field].(string)))
	}

	return nil, errors.New(fmt.Sprintf("the data had no `type` specified %s", data))

}
