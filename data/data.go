package data

type Basal struct {
	DeliveryType string          `json:"deliveryType" valid:"required"`
	Insulin      string          `json:"insulin" valid:"required"`
	Value        float32         `json:"value" valid:"required"`
	Duration     int64           `json:"duration" valid:"required"`
	Suppressed   *SupressedBasal `json:"suppressed"`
	Base
}

type SupressedBasal struct {
	Type         string  `json:"type" valid:"required"`
	DeliveryType string  `json:"deliveryType" valid:"required"`
	Value        float32 `json:"value" valid:"required"`
}

type DeviceEvent struct {
	SubType string `json:"subType" valid:"required"`
	Base
}

type Base struct {
	Type             string  `json:"type" valid:"required"`
	DeviceTime       string  `json:"deviceTime" valid:"required"`
	Time             string  `json:"time" valid:"required"`
	TimezoneOffset   float64 `json:"timezoneOffset" valid:"required"`
	ConversionOffset float64 `json:"conversionOffset" valid:"required"`
	DeviceId         string  `json:"deviceId" valid:"required"`
}

func GetData() string {
	return "data"
}
