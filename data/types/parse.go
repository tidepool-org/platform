package types

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
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base"
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

func Parse(parser data.ObjectParser) (data.Datum, error) {
	if parser == nil {
		return nil, app.Error("types", "parser is missing")
	}

	datumType := parser.ParseString("type")
	if datumType == nil {
		parser.AppendError("type", base.ErrorValueMissing())
		return nil, nil
	}

	var datum data.Datum
	var err error

	switch *datumType {
	case basal.Type():
		deliveryType := parser.ParseString("deliveryType")
		if deliveryType == nil {
			parser.AppendError("deliveryType", base.ErrorValueMissing())
			return nil, nil
		}

		switch *deliveryType {
		case scheduled.DeliveryType():
			datum, err = scheduled.New()
		case suspend.DeliveryType():
			datum, err = suspend.New()
		case temporary.DeliveryType():
			datum, err = temporary.New()
		default:
			parser.AppendError("deliveryType", base.ErrorDeliveryTypeInvalid(*deliveryType))
			return nil, nil
		}
	case bolus.Type():
		subType := parser.ParseString("subType")
		if subType == nil {
			parser.AppendError("subType", base.ErrorValueMissing())
			return nil, nil
		}

		switch *subType {
		case combination.SubType():
			datum, err = combination.New()
		case extended.SubType():
			datum, err = extended.New()
		case normal.SubType():
			datum, err = normal.New()
		default:
			parser.AppendError("subType", base.ErrorSubTypeInvalid(*subType))
			return nil, nil
		}
	case calculator.Type():
		datum, err = calculator.New()
	case continuous.Type():
		datum, err = continuous.New()
	case device.Type():
		subType := parser.ParseString("subType")
		if subType == nil {
			parser.AppendError("subType", base.ErrorValueMissing())
			return nil, nil
		}

		switch *subType {
		case alarm.SubType():
			datum, err = alarm.New()
		case calibration.SubType():
			datum, err = calibration.New()
		case prime.SubType():
			datum, err = prime.New()
		case reservoirchange.SubType():
			datum, err = reservoirchange.New()
		case status.SubType():
			datum, err = status.New()
		case timechange.SubType():
			datum, err = timechange.New()
		default:
			parser.AppendError("subType", base.ErrorSubTypeInvalid(*subType))
			return nil, nil
		}
	case ketone.Type():
		datum, err = ketone.New()
	case pump.Type():
		datum, err = pump.New()
	case selfmonitored.Type():
		datum, err = selfmonitored.New()
	case upload.Type():
		datum, err = upload.New()
	default:
		parser.AppendError("type", base.ErrorTypeInvalid(*datumType))
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	if datum == nil {
		return nil, app.Error("types", "datum is missing")
	}

	if err = datum.Parse(parser); err != nil {
		return nil, err
	}

	return datum, nil
}
