package factory

import (
	"sort"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/activity/physical"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/basal/automated"
	"github.com/tidepool-org/platform/data/types/basal/scheduled"
	"github.com/tidepool-org/platform/data/types/basal/suspend"
	"github.com/tidepool-org/platform/data/types/basal/temporary"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/blood/ketone"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/types/bolus/combination"
	"github.com/tidepool-org/platform/data/types/bolus/extended"
	"github.com/tidepool-org/platform/data/types/bolus/normal"
	"github.com/tidepool-org/platform/data/types/calculator"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/alarm"
	"github.com/tidepool-org/platform/data/types/device/calibration"
	"github.com/tidepool-org/platform/data/types/device/prime"
	"github.com/tidepool-org/platform/data/types/device/reservoirchange"
	"github.com/tidepool-org/platform/data/types/device/status"
	"github.com/tidepool-org/platform/data/types/device/timechange"
	"github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/data/types/state/reported"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
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
			return nil, errors.New("inspector is missing")
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
			return nil, errors.New("inspector is missing")
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

// TODO: Consider injecting the entire map from the data service

func NewStandard() (*Standard, error) {
	var basalNewFuncMap = NewFuncMap{
		automated.DeliveryType: NewNewFuncWithFunc(automated.NewDatum),
		scheduled.DeliveryType: NewNewFuncWithFunc(scheduled.NewDatum),
		suspend.DeliveryType:   NewNewFuncWithFunc(suspend.NewDatum),
		temporary.DeliveryType: NewNewFuncWithFunc(temporary.NewDatum),
	}

	var bolusNewFuncMap = NewFuncMap{
		combination.SubType: NewNewFuncWithFunc(combination.NewDatum),
		extended.SubType:    NewNewFuncWithFunc(extended.NewDatum),
		normal.SubType:      NewNewFuncWithFunc(normal.NewDatum),
	}

	var deviceNewFuncMap = NewFuncMap{
		alarm.SubType:           NewNewFuncWithFunc(alarm.NewDatum),
		calibration.SubType:     NewNewFuncWithFunc(calibration.NewDatum),
		prime.SubType:           NewNewFuncWithFunc(prime.NewDatum),
		reservoirchange.SubType: NewNewFuncWithFunc(reservoirchange.NewDatum),
		status.SubType:          NewNewFuncWithFunc(status.NewDatum),
		timechange.SubType:      NewNewFuncWithFunc(timechange.NewDatum),
	}

	var baseNewFuncMap = NewFuncMap{
		basal.Type:         NewNewFuncWithKeyAndMap("deliveryType", basalNewFuncMap),
		bolus.Type:         NewNewFuncWithKeyAndMap("subType", bolusNewFuncMap),
		calculator.Type:    NewNewFuncWithFunc(calculator.NewDatum),
		continuous.Type:    NewNewFuncWithFunc(continuous.NewDatum),
		device.Type:        NewNewFuncWithKeyAndMap("subType", deviceNewFuncMap),
		food.Type:          NewNewFuncWithFunc(food.NewDatum),
		insulin.Type:       NewNewFuncWithFunc(insulin.NewDatum),
		ketone.Type:        NewNewFuncWithFunc(ketone.NewDatum),
		physical.Type:      NewNewFuncWithFunc(physical.NewDatum),
		pump.Type:          NewNewFuncWithFunc(pump.NewDatum),
		reported.Type:      NewNewFuncWithFunc(reported.NewDatum),
		selfmonitored.Type: NewNewFuncWithFunc(selfmonitored.NewDatum),
		upload.Type:        NewNewFuncWithFunc(upload.NewDatum),
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
