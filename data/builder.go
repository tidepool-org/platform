package data

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"fmt"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/bloodglucose"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/calculator"
	"github.com/tidepool-org/platform/data/types/cgm"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/pump"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/validate"
)

type BuiltDatum interface{}
type BuiltDatumArray []interface{}

type Builder interface {
	BuildFromDatum(datum types.Datum) (BuiltDatum, []*service.Error)
	BuildFromDatumArray(datumArray types.DatumArray) (BuiltDatumArray, []*service.Error)
}

type TypeBuilder struct {
	commonDatum types.Datum
}

func NewTypeBuilder(commonDatum types.Datum) Builder {
	return &TypeBuilder{
		commonDatum: commonDatum,
	}
}

func (t *TypeBuilder) BuildFromDatumArray(datumArray types.DatumArray) (BuiltDatumArray, []*service.Error) {

	errors := service.NewErrors()

	builtDatumArray := BuiltDatumArray{}
	for index := range datumArray {
		errorProcessing := validate.NewErrorProcessing(fmt.Sprintf("%d", index))
		errorProcessing.Errors = errors
		builtDatumArray = append(builtDatumArray, t.build(datumArray[index], errorProcessing))
	}

	if errors.HasErrors() {
		return nil, errors.GetErrors()
	}

	return builtDatumArray, nil
}

func (t *TypeBuilder) BuildFromDatum(datum types.Datum) (BuiltDatum, []*service.Error) {

	errorProcessing := validate.NewErrorProcessing("")

	builtDatum := t.build(datum, errorProcessing)

	if errorProcessing.HasErrors() {
		return nil, errorProcessing.GetErrors()
	}

	return builtDatum, nil
}

func (t *TypeBuilder) build(datum types.Datum, errorProcessing validate.ErrorProcessing) BuiltDatum {
	typeName, ok := datum["type"].(string)
	if !ok {
		// TODO: Use types package for this
		errorProcessing.AppendPointerError("type", types.InvalidDataTitle, "Missing type")
		return nil
	}

	for k, v := range t.commonDatum {
		datum[k] = v
	}

	// datum["id"] = app.NewUUID()		// TODO: Is this necessary?

	switch typeName {
	case basal.Name:
		return basal.Build(datum, errorProcessing)
	case device.Name:
		return device.Build(datum, errorProcessing)
	case bolus.Name:
		return bolus.Build(datum, errorProcessing)
	case bloodglucose.ContinuousName:
		return bloodglucose.BuildContinuous(datum, errorProcessing)
	case bloodglucose.SelfMonitoredName:
		return bloodglucose.BuildSelfMonitored(datum, errorProcessing)
	case upload.Name:
		return upload.Build(datum, errorProcessing)
	case calculator.Name:
		return calculator.Build(datum, errorProcessing)
	case pump.Name:
		return pump.Build(datum, errorProcessing)
	case cgm.Name:
		return cgm.Build(datum, errorProcessing)
	default:
		// TODO: Use types package for this
		errorProcessing.AppendPointerError("type", types.InvalidTypeTitle, fmt.Sprintf("Unknown type '%s'", typeName))
		return nil
	}
}
