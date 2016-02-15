package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	log "github.com/tidepool-org/platform/logger"
)

type GenericDatam map[string]interface{}
type GenericDataset []GenericDatam

type Builder interface {
	BuildFromRaw(raw []byte) (interface{}, *DataError)
	BuildFromData(data map[string]interface{}) (interface{}, *DataError)
	BuildFromDataSet(dataSet GenericDataset) ([]interface{}, *DataSetError)
}

type TypeBuilder struct{}

func NewTypeBuilder() Builder {
	return &TypeBuilder{}
}

func (this *TypeBuilder) BuildFromDataSet(dataSet GenericDataset) ([]interface{}, *DataSetError) {

	var set []interface{}
	var buildError *DataSetError

	for i := range dataSet {
		item, err := this.BuildFromData(dataSet[i])
		if err != nil && !err.IsEmpty() {
			if buildError == nil {
				buildError = NewDataSetError()
			}
			buildError.AppendError(err)
		}
		set = append(set, item)
	}
	return set, buildError
}

func (this *TypeBuilder) BuildFromRaw(raw []byte) (interface{}, *DataError) {

	var data map[string]interface{}

	if err := json.NewDecoder(strings.NewReader(string(raw))).Decode(&data); err != nil {
		log.Logging.Info("error doing an unmarshal", err.Error())
		e := NewDataError(data)
		e.AppendError(errors.New(fmt.Sprintf("sorry but we do anything with %s", string(raw))))
		return nil, e
	}
	return this.BuildFromData(data)
}

func (this *TypeBuilder) BuildFromData(data map[string]interface{}) (interface{}, *DataError) {

	const (
		type_field        = "type"
		basal_type        = "basal"
		device_event_type = "deviceevent"
	)
	if data[type_field] != nil {

		if strings.ToLower(data[type_field].(string)) == basal_type {
			return BuildBasal(data)
		} else if strings.ToLower(data[type_field].(string)) == device_event_type {
			return BuildDeviceEvent(data)
		}
		e := NewDataError(data)
		e.AppendError(errors.New(fmt.Sprintf("we can't deal with `type`=%s", data[type_field].(string))))
		return nil, e
	}

	e := NewDataError(data)
	e.AppendError(errors.New("there is no match for that type"))

	return nil, e

}
