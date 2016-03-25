package data

import (
	"fmt"
	"strings"

	"github.com/tidepool-org/platform/validate"
)

type Datum map[string]interface{}

type DatumArray []Datum

type Builder interface {
	BuildFromDatum(datum Datum) interface{}
	BuildFromDatumArray(datumArray DatumArray) ([]interface{}, *validate.ErrorsArray)
}

type TypeBuilder struct {
	inject map[string]interface{}
	Errors *validate.ErrorsArray
	Index  int
}

func NewTypeBuilder(inject map[string]interface{}) Builder {
	return &TypeBuilder{
		inject: inject,
		Errors: validate.NewErrorsArray(),
		Index:  0,
	}
}

func (t *TypeBuilder) BuildFromDatumArray(datumArray DatumArray) ([]interface{}, *validate.ErrorsArray) {

	var set []interface{}

	for i := range datumArray {
		if item := t.BuildFromDatum(datumArray[i]); item != nil {
			set = append(set, item)
		}
		t.Index++
	}
	if t.Errors.HasErrors() {
		return nil, t.Errors
	}

	return set, nil
}

func (t *TypeBuilder) BuildFromDatum(datum Datum) interface{} {

	const typeField = "type"

	if datum[typeField] != nil {

		for k, v := range t.inject {
			datum[k] = v
		}

		if strings.ToLower(datum[typeField].(string)) == strings.ToLower(BasalName) {
			return BuildBasal(datum, t.Errors)
		} else if strings.ToLower(datum[typeField].(string)) == strings.ToLower(DeviceEventName) {
			return BuildDeviceEvent(datum, t.Errors)
		}
		t.Errors.Append(validate.NewPointerError(fmt.Sprintf("%d/type", t.Index), "Invalid type", "The type must be one of `basal`, `deviceEvent`"))
		return nil
	}
	t.Errors.Append(validate.NewParameterError(fmt.Sprintf("%d", t.Index), "Invalid data", "Missing required `type` field."))

	return nil

}
