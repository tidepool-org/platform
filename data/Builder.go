package data

import (
	"fmt"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/bloodglucose"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/validate"
)

type BuiltDatum interface{}
type BuiltDatumArray []BuiltDatum

type Builder interface {
	BuildFromDatum(datum types.Datum) BuiltDatum
	BuildFromDatumArray(datumArray types.DatumArray) (BuiltDatumArray, *validate.ErrorsArray)
}

type TypeBuilder struct {
	commonDatum types.Datum
	Index       int
	validate.ErrorProcessing
}

func NewTypeBuilder(commonDatum types.Datum) Builder {
	return &TypeBuilder{
		commonDatum:     commonDatum,
		ErrorProcessing: validate.ErrorProcessing{ErrorsArray: validate.NewErrorsArray()},
		Index:           0,
	}
}

func (t *TypeBuilder) BuildFromDatumArray(datumArray types.DatumArray) (BuiltDatumArray, *validate.ErrorsArray) {

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

func (t *TypeBuilder) buildType(typeName string, datum types.Datum) BuiltDatum {

	t.ErrorProcessing.BasePath = fmt.Sprintf("%d/%s", t.Index, typeName)

	switch typeName {
	case basal.Name:
		return basal.Build(datum, t.ErrorProcessing)
	case device.Name:
		return device.Build(datum, t.ErrorProcessing)
	case bolus.Name:
		return bolus.Build(datum, t.ErrorProcessing)
	case bloodglucose.ContinuousName:
		return bloodglucose.BuildContinuous(datum, t.ErrorProcessing)
	case bloodglucose.SelfMonitoredName:
		return bloodglucose.BuildSelfMonitored(datum, t.ErrorProcessing)
	case upload.Name:
		return upload.Build(datum, t.ErrorProcessing)
	default:
		t.ErrorProcessing.AppendPointerError("type", types.InvalidTypeTitle, "The type must be one of 'basal', 'deviceEvent'")
		return nil
	}
}

func (t *TypeBuilder) BuildFromDatum(datum types.Datum) BuiltDatum {

	if datum["type"] == nil {
		t.ErrorProcessing.Append(validate.NewParameterError(fmt.Sprintf("%d", t.Index), types.InvalidDataTitle, "Missing required 'type' field"))
		return nil
	}

	for k, v := range t.commonDatum {
		datum[k] = v
	}
	return t.buildType(datum["type"].(string), datum)

}
