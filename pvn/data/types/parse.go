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
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/basal"
	"github.com/tidepool-org/platform/pvn/data/types/base/basal/scheduled"
	"github.com/tidepool-org/platform/pvn/data/types/base/basal/suspend"
	"github.com/tidepool-org/platform/pvn/data/types/base/basal/temporary"
	"github.com/tidepool-org/platform/pvn/data/types/base/bolus"
	"github.com/tidepool-org/platform/pvn/data/types/base/bolus/calculator"
	"github.com/tidepool-org/platform/pvn/data/types/base/bolus/combination"
	"github.com/tidepool-org/platform/pvn/data/types/base/bolus/extended"
	"github.com/tidepool-org/platform/pvn/data/types/base/bolus/normal"
	"github.com/tidepool-org/platform/pvn/data/types/base/continuous"
	"github.com/tidepool-org/platform/pvn/data/types/base/device"
	"github.com/tidepool-org/platform/pvn/data/types/base/device/alarm"
	"github.com/tidepool-org/platform/pvn/data/types/base/device/calibration"
	"github.com/tidepool-org/platform/pvn/data/types/base/device/prime"
	"github.com/tidepool-org/platform/pvn/data/types/base/device/reservoirchange"
	"github.com/tidepool-org/platform/pvn/data/types/base/device/status"
	"github.com/tidepool-org/platform/pvn/data/types/base/device/timechange"
	"github.com/tidepool-org/platform/pvn/data/types/base/ketone"
	"github.com/tidepool-org/platform/pvn/data/types/base/pump"
	"github.com/tidepool-org/platform/pvn/data/types/base/sample"
	"github.com/tidepool-org/platform/pvn/data/types/base/sample/sub"
	"github.com/tidepool-org/platform/pvn/data/types/base/selfmonitored"
	"github.com/tidepool-org/platform/pvn/data/types/base/upload"
)

func Parse(context data.Context, parser data.ObjectParser) (data.Datum, error) {
	if context == nil {
		return nil, app.Error("types", "context is missing")
	}
	if parser == nil {
		return nil, app.Error("types", "parser is missing")
	}

	var datum data.Datum

	datumType := parser.ParseString("type")
	if datumType == nil {
		context.AppendError("type", ErrorValueMissing())
		return nil, nil
	}

	datumSubType := parser.ParseString("subType")

	// NOTE: This is the "master switchboard" that creates all of the datum from
	// the type and subType

	switch *datumType {
	case sample.Type():
		if datumSubType != nil {
			switch *datumSubType {
			case sub.SubType():
				datum = sub.New()
			default:
				context.AppendError("subType", ErrorSubTypeInvalid(*datumSubType))
				return nil, nil
			}
		} else {
			datum = sample.New()
		}
	case basal.Type():

		datumDeliveryType := parser.ParseString("deliveryType")

		if datumDeliveryType == nil {
			context.AppendError("deliveryType", ErrorSubTypeInvalid(*datumDeliveryType))
			return nil, nil
		}

		switch *datumDeliveryType {
		case suspend.DeliveryType():
			datum = suspend.New()
		case scheduled.DeliveryType():
			datum = scheduled.New()
		case temporary.DeliveryType():
			datum = temporary.New()
		default:
			context.AppendError("deliveryType", ErrorSubTypeInvalid(*datumDeliveryType))
			return nil, nil
		}

	case bolus.Type():

		bolusSubType := parser.ParseString("subType")

		if bolusSubType == nil {
			context.AppendError("subType", ErrorSubTypeInvalid(*bolusSubType))
			return nil, nil
		}

		switch *bolusSubType {
		case normal.SubType():
			datum = normal.New()
		case extended.SubType():
			datum = extended.New()
		case combination.SubType():
			datum = combination.New()
		default:
			context.AppendError("subType", ErrorSubTypeInvalid(*bolusSubType))
			return nil, nil
		}
	case device.Type():

		deviceSubType := parser.ParseString("subType")

		if deviceSubType == nil {
			context.AppendError("subType", ErrorSubTypeInvalid(*deviceSubType))
			return nil, nil
		}

		switch *deviceSubType {
		case alarm.SubType():
			datum = alarm.New()
		case prime.SubType():
			datum = prime.New()
		case calibration.SubType():
			datum = calibration.New()
		case reservoirchange.SubType():
			datum = reservoirchange.New()
		case status.SubType():
			datum = status.New()
		case timechange.SubType():
			datum = timechange.New()
		default:
			context.AppendError("subType", ErrorSubTypeInvalid(*deviceSubType))
			return nil, nil
		}
	case calculator.Type():
		datum = calculator.New()
	case upload.Type():
		datum = upload.New()
	case ketone.BloodType():
		datum = ketone.NewBlood()
	case continuous.Type():
		datum = continuous.New()
	case selfmonitored.Type():
		datum = selfmonitored.New()
	case pump.Type():
		datum = pump.New()
	default:
		context.AppendError("type", ErrorTypeInvalid(*datumType))
		return nil, nil
	}

	datum.Parse(parser)

	return datum, nil
}
