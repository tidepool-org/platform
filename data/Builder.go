package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

type Builder interface {
	Build(raw []byte) (interface{}, *DataError)
}

type TypeBuilder struct{}

func NewTypeBuilder() Builder {
	return &TypeBuilder{}
}

func (this *TypeBuilder) Build(raw []byte) (interface{}, *DataError) {

	const (
		type_field        = "type"
		basal_type        = "basal"
		device_event_type = "deviceevent"
	)

	var data map[string]interface{}

	//d := json.NewDecoder(strings.NewReader(string(raw))).Decode(v)
	//d.UseNumber()

	if err := json.NewDecoder(strings.NewReader(string(raw))).Decode(&data); err != nil {
		log.Println("error doing an unmarshal", err.Error())
		e := NewDataError(data)
		e.AppendError(errors.New(fmt.Sprintf("sorry but we do anything with %s", string(raw))))
		return nil, e
	}

	if data[type_field] != nil {

		if strings.ToLower(data[type_field].(string)) == basal_type {
			return BuildBasal(data)
		} else if strings.ToLower(data[type_field].(string)) == device_event_type {
			return BuildDeviceEvent(data)
		}
		e := NewDataError(data)
		e.AppendError(errors.New(fmt.Sprintf("we can't deal with `type`=%s", data[type_field].(string))))
		return nil, e
	}

	e := NewDataError(data)
	e.AppendError(errors.New("there is no match for that type"))

	return nil, e

}
