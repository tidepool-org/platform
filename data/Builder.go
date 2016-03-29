package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Datum map[string]interface{}

type DatumArray []Datum

type Builder interface {
	BuildFromBytes(byteArray []byte) (interface{}, *Error)
	BuildFromDatum(datum Datum) (interface{}, *Error)
	BuildFromDatumArray(datumArray DatumArray) ([]interface{}, *ErrorArray)
}

type TypeBuilder struct {
	inject map[string]interface{}
}

func NewTypeBuilder(inject map[string]interface{}) Builder {
	return &TypeBuilder{
		inject: inject,
	}
}

func (t *TypeBuilder) BuildFromBytes(byteArray []byte) (interface{}, *Error) {

	var datum Datum

	if err := json.NewDecoder(strings.NewReader(string(byteArray))).Decode(&datum); err != nil {
		log.Info("error doing an unmarshal", err.Error())
		e := NewError(datum)
		e.AppendError(fmt.Errorf("sorry but we do anything with %s", string(byteArray)))
		return nil, e
	}
	return t.BuildFromDatum(datum)
}

func (t *TypeBuilder) BuildFromDatumArray(datumArray DatumArray) ([]interface{}, *ErrorArray) {

	var set []interface{}
	buildError := NewErrorArray()

	for i := range datumArray {
		if item, err := t.BuildFromDatum(datumArray[i]); err.IsEmpty() {
			set = append(set, item)
		} else {
			buildError.AppendError(err)
		}
	}
	if len(buildError.errors) > 0 {
		return nil, buildError
	}
	return set, nil
}

func (t *TypeBuilder) BuildFromDatum(datum Datum) (interface{}, *Error) {

	const typeField = "type"

	if datum[typeField] != nil {

		for k, v := range t.inject {
			datum[k] = v
		}

		if strings.ToLower(datum[typeField].(string)) == strings.ToLower(BasalName) {
			return BuildBasal(datum)
		} else if strings.ToLower(datum[typeField].(string)) == strings.ToLower(DeviceEventName) {
			return BuildDeviceEvent(datum)
		}
		e := NewError(datum)
		e.AppendError(fmt.Errorf("we can't deal with `type`=%s", datum[typeField].(string)))
		return nil, e
	}

	e := NewError(datum)
	e.AppendError(errors.New("there is no match for that type"))

	return nil, e

}
