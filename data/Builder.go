package data

import (
	"fmt"

	"github.com/tidepool-org/platform/validate"
)

const (
	InvalidTypeTitle = "Invalid type"
	InvalidDataTitle = "Invalid data"
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
	commonDatum Datum
	Index       int
	validate.ErrorProcessing
}

func NewTypeBuilder(commonDatum Datum) Builder {
	return &TypeBuilder{
		commonDatum:     commonDatum,
		ErrorProcessing: validate.ErrorProcessing{ErrorsArray: validate.NewErrorsArray()},
		Index:           0,
	}
}

func (t *TypeBuilder) BuildFromDatumArray(datumArray DatumArray) (BuiltDatumArray, *validate.ErrorsArray) {

	var builtDatumArray BuiltDatumArray

	for i := range datumArray {
		builtDatumArray = append(builtDatumArray, t.BuildFromDatum(datumArray[i]))
		t.Index++
	}

	if t.ErrorProcessing.HasErrors() {
		return nil, t.ErrorsArray
	}

	return builtDatumArray, nil
}

func (t *TypeBuilder) buildType(typeName string, datum Datum) BuiltDatum {

	t.ErrorProcessing.BasePath = fmt.Sprintf("%d/%s", t.Index, typeName)

	switch typeName {
	case BasalName:
		return BuildBasal(datum, t.ErrorProcessing)
	case DeviceEventName:
		return BuildDeviceEvent(datum, t.ErrorProcessing)
	default:
		t.ErrorProcessing.AppendPointerError("type", InvalidTypeTitle, "The type must be one of 'basal', 'deviceEvent'")
		return nil
	}
}

func (t *TypeBuilder) BuildFromDatum(datum Datum) BuiltDatum {

	if datum["type"] == nil {
		t.ErrorProcessing.Append(validate.NewParameterError(fmt.Sprintf("%d", t.Index), InvalidDataTitle, "Missing required 'type' field"))
		return nil
	}

	for k, v := range t.commonDatum {
		datum[k] = v
	}
	return t.buildType(datum["type"].(string), datum)

}
