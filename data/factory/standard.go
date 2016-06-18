package factory

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"sort"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base/basal"
	"github.com/tidepool-org/platform/data/types/base/basal/scheduled"
	"github.com/tidepool-org/platform/data/types/base/basal/suspend"
	"github.com/tidepool-org/platform/data/types/base/basal/temporary"
	"github.com/tidepool-org/platform/data/types/base/bolus"
	"github.com/tidepool-org/platform/data/types/base/bolus/calculator"
	"github.com/tidepool-org/platform/data/types/base/bolus/combination"
	"github.com/tidepool-org/platform/data/types/base/bolus/extended"
	"github.com/tidepool-org/platform/data/types/base/bolus/normal"
	"github.com/tidepool-org/platform/data/types/base/continuous"
	"github.com/tidepool-org/platform/data/types/base/device"
	"github.com/tidepool-org/platform/data/types/base/device/alarm"
	"github.com/tidepool-org/platform/data/types/base/device/calibration"
	"github.com/tidepool-org/platform/data/types/base/device/prime"
	"github.com/tidepool-org/platform/data/types/base/device/reservoirchange"
	"github.com/tidepool-org/platform/data/types/base/device/status"
	"github.com/tidepool-org/platform/data/types/base/device/timechange"
	"github.com/tidepool-org/platform/data/types/base/ketone"
	"github.com/tidepool-org/platform/data/types/base/pump"
	"github.com/tidepool-org/platform/data/types/base/selfmonitored"
	"github.com/tidepool-org/platform/data/types/base/upload"
)

type Standard struct {
	newFunc NewFunc
}

type NewFunc func(inspector data.Inspector) (data.Datum, error)
type NewFuncMap map[string]NewFunc

func NewNewFuncWithFunc(datumFunc func() data.Datum) NewFunc {
	if datumFunc == nil {
		return nil
	}

	return func(inspector data.Inspector) (data.Datum, error) {
		if inspector == nil {
			return nil, app.Error("factory", "inspector is missing")
		}

		return datumFunc(), nil
	}
}

func NewNewFuncWithKeyAndMap(key string, newFuncMap NewFuncMap) NewFunc {
	allowedValues := []string{}
	for value := range newFuncMap {
		allowedValues = append(allowedValues, value)
	}
	sort.Strings(allowedValues)

	return func(inspector data.Inspector) (data.Datum, error) {
		if inspector == nil {
			return nil, app.Error("factory", "inspector is missing")
		}

		value := inspector.GetProperty(key)
		if value == nil {
			return nil, inspector.NewMissingPropertyError(key)
		}

		newFunc, ok := newFuncMap[*value]
		if !ok {
			return nil, inspector.NewInvalidPropertyError(key, *value, allowedValues)
		}
		if newFunc == nil {
			return nil, inspector.NewMissingPropertyError(*value)
		}

		return newFunc(inspector)
	}
}

// TODO: Consider injecting the entire map from dataservices

func NewStandard() (*Standard, error) {
	var basalNewFuncMap = NewFuncMap{
		scheduled.DeliveryType(): NewNewFuncWithFunc(scheduled.NewDatum),
		suspend.DeliveryType():   NewNewFuncWithFunc(suspend.NewDatum),
		temporary.DeliveryType(): NewNewFuncWithFunc(temporary.NewDatum),
	}

	var bolusNewFuncMap = NewFuncMap{
		combination.SubType(): NewNewFuncWithFunc(combination.NewDatum),
		extended.SubType():    NewNewFuncWithFunc(extended.NewDatum),
		normal.SubType():      NewNewFuncWithFunc(normal.NewDatum),
	}

	var deviceNewFuncMap = NewFuncMap{
		alarm.SubType():           NewNewFuncWithFunc(alarm.NewDatum),
		calibration.SubType():     NewNewFuncWithFunc(calibration.NewDatum),
		prime.SubType():           NewNewFuncWithFunc(prime.NewDatum),
		reservoirchange.SubType(): NewNewFuncWithFunc(reservoirchange.NewDatum),
		status.SubType():          NewNewFuncWithFunc(status.NewDatum),
		timechange.SubType():      NewNewFuncWithFunc(timechange.NewDatum),
	}

	var baseNewFuncMap = NewFuncMap{
		basal.Type():         NewNewFuncWithKeyAndMap("deliveryType", basalNewFuncMap),
		bolus.Type():         NewNewFuncWithKeyAndMap("subType", bolusNewFuncMap),
		calculator.Type():    NewNewFuncWithFunc(calculator.NewDatum),
		continuous.Type():    NewNewFuncWithFunc(continuous.NewDatum),
		device.Type():        NewNewFuncWithKeyAndMap("subType", deviceNewFuncMap),
		ketone.Type():        NewNewFuncWithFunc(ketone.NewDatum),
		pump.Type():          NewNewFuncWithFunc(pump.NewDatum),
		selfmonitored.Type(): NewNewFuncWithFunc(selfmonitored.NewDatum),
		upload.Type():        NewNewFuncWithFunc(upload.NewDatum),
	}

	return &Standard{
		newFunc: NewNewFuncWithKeyAndMap("type", baseNewFuncMap),
	}, nil
}

func (s *Standard) New(inspector data.Inspector) (data.Datum, error) {
	return s.newFunc(inspector)
}

func (s *Standard) Init(inspector data.Inspector) (data.Datum, error) {
	datum, err := s.New(inspector)
	if err != nil {
		return nil, err
	}

	if datum != nil {
		datum.Init()
	}

	return datum, nil
}
