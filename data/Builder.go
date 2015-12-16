package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tidepool-org/platform/validate"
)

type Builder interface {
	Build(raw []byte) (interface{}, error)
}

type TypeBuilder struct {
	validator validate.PlatformValidator
}

func NewTypeBuilder() Builder {
	return &TypeBuilder{validator: validate.PlatformValidator{}}
}

func (this *TypeBuilder) base(t map[string]interface{}) (Base, error) {

	const (
		typeField             = "type"
		deviceTimeField       = "deviceTime"
		timezoneOffsetField   = "timezoneOffset"
		timeField             = "time"
		conversionOffsetField = "conversionOffset"
		deviceIdField         = "deviceId"
	)

	base := Base{
		ConversionOffset: t[conversionOffsetField].(float64),
		TimezoneOffset:   t[timezoneOffsetField].(float64),
		DeviceId:         t[deviceIdField].(string),
		DeviceTime:       t[deviceTimeField].(string),
		Time:             t[timeField].(string),
		Type:             t[typeField].(string),
	}

	_, err := this.validator.Validate(base)
	return base, err
}

func (this *TypeBuilder) Basal(t map[string]interface{}) (*Basal, error) {

	const (
		deliveryTypeField = "deliveryType"
		insulinField      = "insulin"
		valueField        = "value"
		durationField     = "duration"
	)

	base, err := this.base(t)
	if err != nil {
		return nil, err
	}

	basal := &Basal{
		Insulin:      t[insulinField].(string),
		Value:        t[valueField].(float32),
		Duration:     t[durationField].(int64),
		DeliveryType: t[deliveryTypeField].(string),
		Base:         base,
	}

	valid, err := this.validator.Validate(basal)

	if valid {
		return basal, nil
	}
	return nil, err
}

func (this *TypeBuilder) DeviceEvent(t map[string]interface{}) (*DeviceEvent, error) {
	const (
		subTypeField = "subType"
	)

	base, err := this.base(t)
	if err != nil {
		return nil, err
	}

	deviceEvent := &DeviceEvent{
		SubType: t[subTypeField].(string),
		Base:    base,
	}

	valid, err := this.validator.Validate(deviceEvent)

	if valid {
		return deviceEvent, nil
	}
	return nil, err
}

func (this *TypeBuilder) Build(raw []byte) (interface{}, error) {

	const typeField = "type"

	var data map[string]interface{}
	err := json.Unmarshal(raw, &data)
	if err != nil {
		log.Println("error unmarshelling type", err.Error())
		return nil, errors.New(fmt.Sprintf("sorry but we do anything with %s", string(raw)))
	}

	if data[typeField] != nil {

		if strings.ToLower(data[typeField].(string)) == "basal" {
			return this.Basal(data)
		} else if strings.ToLower(data[typeField].(string)) == "deviceevent" {
			return this.DeviceEvent(data)
		}
		return nil, errors.New(fmt.Sprintf("sorry but we can't deal with `type` %s", data[typeField].(string)))
	}

	return nil, errors.New(fmt.Sprintf("the data had no `type` specified %s", data))

}
