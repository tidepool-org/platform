package test

import (
	dataTypesStatusPump "github.com/tidepool-org/platform/data/types/status/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomPump() *dataTypesStatusPump.Pump {
	datum := randomPump()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "pumpStatus"
	return datum
}

func RandomPumpForParser() *dataTypesStatusPump.Pump {
	datum := randomPump()
	datum.Base = *dataTypesTest.RandomBaseForParser()
	datum.Type = "pumpStatus"
	return datum
}

func randomPump() *dataTypesStatusPump.Pump {
	datum := dataTypesStatusPump.New()
	datum.BasalDelivery = RandomBasalDelivery()
	datum.Battery = RandomBattery()
	datum.BolusDelivery = RandomBolusDelivery()
	datum.DeliveryIndeterminant = pointer.FromBool(test.RandomBool())
	datum.Reservoir = RandomReservoir()
	return datum
}

func ClonePump(datum *dataTypesStatusPump.Pump) *dataTypesStatusPump.Pump {
	if datum == nil {
		return nil
	}
	clone := dataTypesStatusPump.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.BasalDelivery = CloneBasalDelivery(datum.BasalDelivery)
	clone.Battery = CloneBattery(datum.Battery)
	clone.BolusDelivery = CloneBolusDelivery(datum.BolusDelivery)
	clone.DeliveryIndeterminant = pointer.CloneBool(datum.DeliveryIndeterminant)
	clone.Reservoir = CloneReservoir(datum.Reservoir)
	return clone
}

func NewObjectFromPump(datum *dataTypesStatusPump.Pump, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := dataTypesTest.NewObjectFromBase(&datum.Base, objectFormat)
	if datum.BasalDelivery != nil {
		object["basalDelivery"] = NewObjectFromBasalDelivery(datum.BasalDelivery, objectFormat)
	}
	if datum.Battery != nil {
		object["battery"] = NewObjectFromBattery(datum.Battery, objectFormat)
	}
	if datum.BolusDelivery != nil {
		object["bolusDelivery"] = NewObjectFromBolusDelivery(datum.BolusDelivery, objectFormat)
	}
	if datum.DeliveryIndeterminant != nil {
		object["deliveryIndeterminant"] = test.NewObjectFromBool(*datum.DeliveryIndeterminant, objectFormat)
	}
	if datum.Reservoir != nil {
		object["reservoir"] = NewObjectFromReservoir(datum.Reservoir, objectFormat)
	}
	return object
}
