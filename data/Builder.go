package data

import (
	"fmt"

	"github.com/tidepool-org/platform/validate"
)

type Datum map[string]interface{}
type DatumArray []Datum

type BuiltDatum interface{}
type BuiltDatumArray []BuiltDatum

type Builder interface {
	BuildFromDatum(datum Datum) BuiltDatum
	BuildFromDatumArray(datumArray DatumArray) (BuiltDatumArray, *validate.ErrorsArray)
}

type TypeBuilder struct {
	inject map[string]interface{}
	Index  int
	validate.ErrorProcessing
}

func NewTypeBuilder(inject map[string]interface{}) Builder {
	return &TypeBuilder{
		inject:          inject,
		ErrorProcessing: validate.ErrorProcessing{ErrorsArray: validate.NewErrorsArray()},
		Index:           0,
	}
}

func (t *TypeBuilder) BuildFromDatumArray(datumArray DatumArray) (BuiltDatumArray, *validate.ErrorsArray) {

	var set BuiltDatumArray

	for i := range datumArray {
		if item := t.BuildFromDatum(datumArray[i]); item != nil {
			set = append(set, item)
		}
		t.Index++
	}
	if t.ErrorProcessing.HasErrors() {
		return nil, t.ErrorsArray
	}

	return set, nil
}

func (t *TypeBuilder) buildType(typeName string, datum Datum) BuiltDatum {

	t.ErrorProcessing.BasePath = fmt.Sprintf("%d/%s", t.Index, typeName)

	switch typeName {
	case BasalName:
		return BuildBasal(datum, t.ErrorProcessing)
	case DeviceEventName:
		return BuildDeviceEvent(datum, t.ErrorProcessing)
	default:
		t.ErrorProcessing.AppendPointerError("type", "Invalid type", "The type must be one of `basal`, `deviceEvent`")
		return nil
	}
}

func (t *TypeBuilder) BuildFromDatum(datum Datum) BuiltDatum {

	if datum["type"] == nil {
		t.ErrorProcessing.Append(validate.NewParameterError(fmt.Sprintf("%d", t.Index), "Invalid data", "Missing required `type` field."))
		return nil
	}

	for k, v := range t.inject {
		datum[k] = v
	}
	return t.buildType(datum["type"].(string), datum)

}
