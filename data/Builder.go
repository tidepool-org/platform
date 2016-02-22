package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	log "github.com/tidepool-org/platform/logger"
)

//GenericDatam represent one data point
type GenericDatam map[string]interface{}

//GenericDataset represents an array of data points
type GenericDataset []GenericDatam

//Builder interface that the TypeBuilder implements
type Builder interface {
	BuildFromRaw(raw []byte) (interface{}, *Error)
	BuildFromData(data map[string]interface{}) (interface{}, *Error)
	BuildFromDataSet(dataSet GenericDataset) ([]interface{}, *ErrorSet)
}

//TypeBuilder that is used to build data types that the platform understands
type TypeBuilder struct{}

//NewTypeBuilder returns an instance of TypeBuilder
func NewTypeBuilder() Builder {
	return &TypeBuilder{}
}

//BuildFromDataSet will build the matching type(s) from the given GenericDataset
func (typeBuilder *TypeBuilder) BuildFromDataSet(dataSet GenericDataset) ([]interface{}, *ErrorSet) {

	var set []interface{}
	var buildError *ErrorSet

	for i := range dataSet {
		item, err := typeBuilder.BuildFromData(dataSet[i])
		if err != nil && !err.IsEmpty() {
			if buildError == nil {
				buildError = NewErrorSet()
			}
			buildError.AppendError(err)
		}
		set = append(set, item)
	}
	return set, buildError
}

//BuildFromRaw will build the matching type(s) from the given raw data
func (typeBuilder *TypeBuilder) BuildFromRaw(raw []byte) (interface{}, *Error) {

	var data map[string]interface{}

	if err := json.NewDecoder(strings.NewReader(string(raw))).Decode(&data); err != nil {
		log.Logging.Info("error doing an unmarshal", err.Error())
		e := NewError(data)
		e.AppendError(fmt.Errorf("sorry but we do anything with %s", string(raw)))
		return nil, e
	}
	return typeBuilder.BuildFromData(data)
}

//BuildFromData will build the matching type from the given raw data
func (typeBuilder *TypeBuilder) BuildFromData(data map[string]interface{}) (interface{}, *Error) {

	const (
		typeField       = "type"
		basalType       = "basal"
		deviceEventType = "deviceevent"
	)
	if data[typeField] != nil {

		if strings.ToLower(data[typeField].(string)) == basalType {
			return BuildBasal(data)
		} else if strings.ToLower(data[typeField].(string)) == deviceEventType {
			return BuildDeviceEvent(data)
		}
		e := NewError(data)
		e.AppendError(fmt.Errorf("we can't deal with `type`=%s", data[typeField].(string)))
		return nil, e
	}

	e := NewError(data)
	e.AppendError(errors.New("there is no match for that type"))

	return nil, e

}
