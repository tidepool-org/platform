package parser_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/parser")
}

type AppendErrorInput struct {
	reference interface{}
	err       *service.Error
}

type TestObjectParser struct {
	AppendErrorInputs  []AppendErrorInput
	ObjectOutputs      []*map[string]interface{}
	ParseStringInputs  []string
	ParseStringOutputs []*string
}

func (t *TestObjectParser) Logger() log.Logger                                    { return nil }
func (t *TestObjectParser) SetMeta(meta interface{})                              {}
func (t *TestObjectParser) ParseBoolean(key string) *bool                         { return nil }
func (t *TestObjectParser) ParseInteger(key string) *int                          { return nil }
func (t *TestObjectParser) ParseFloat(key string) *float64                        { return nil }
func (t *TestObjectParser) ParseStringArray(key string) *[]string                 { return nil }
func (t *TestObjectParser) ParseObject(key string) *map[string]interface{}        { return nil }
func (t *TestObjectParser) ParseObjectArray(key string) *[]map[string]interface{} { return nil }
func (t *TestObjectParser) ParseInterface(key string) *interface{}                { return nil }
func (t *TestObjectParser) ParseInterfaceArray(key string) *[]interface{}         { return nil }
func (t *TestObjectParser) ProcessNotParsed()                                     {}
func (t *TestObjectParser) NewChildObjectParser(key string) data.ObjectParser     { return nil }
func (t *TestObjectParser) NewChildArrayParser(key string) data.ArrayParser       { return nil }

func (t *TestObjectParser) AppendError(reference interface{}, err *service.Error) {
	t.AppendErrorInputs = append(t.AppendErrorInputs, AppendErrorInput{reference, err})
}

func (t *TestObjectParser) Object() *map[string]interface{} {
	output := t.ObjectOutputs[0]
	t.ObjectOutputs = t.ObjectOutputs[1:]
	return output
}

func (t *TestObjectParser) ParseString(key string) *string {
	t.ParseStringInputs = append(t.ParseStringInputs, key)
	output := t.ParseStringOutputs[0]
	t.ParseStringOutputs = t.ParseStringOutputs[1:]
	return output
}

type TestArrayParser struct {
	ArrayOutputs []*[]interface{}
}

func (t *TestArrayParser) Logger() log.Logger                                    { return nil }
func (t *TestArrayParser) SetMeta(meta interface{})                              {}
func (t *TestArrayParser) AppendError(reference interface{}, err *service.Error) {}
func (t *TestArrayParser) ParseBoolean(index int) *bool                          { return nil }
func (t *TestArrayParser) ParseInteger(index int) *int                           { return nil }
func (t *TestArrayParser) ParseFloat(index int) *float64                         { return nil }
func (t *TestArrayParser) ParseString(index int) *string                         { return nil }
func (t *TestArrayParser) ParseStringArray(index int) *[]string                  { return nil }
func (t *TestArrayParser) ParseObject(index int) *map[string]interface{}         { return nil }
func (t *TestArrayParser) ParseObjectArray(index int) *[]map[string]interface{}  { return nil }
func (t *TestArrayParser) ParseInterface(index int) *interface{}                 { return nil }
func (t *TestArrayParser) ParseInterfaceArray(index int) *[]interface{}          { return nil }
func (t *TestArrayParser) ProcessNotParsed()                                     {}
func (t *TestArrayParser) NewChildObjectParser(index int) data.ObjectParser      { return nil }
func (t *TestArrayParser) NewChildArrayParser(index int) data.ArrayParser        { return nil }

func (t *TestArrayParser) Array() *[]interface{} {
	output := t.ArrayOutputs[0]
	t.ArrayOutputs = t.ArrayOutputs[1:]
	return output
}

type InitOutput struct {
	Datum data.Datum
	Error error
}
