package test

import (
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
)

type ObjectParsable struct {
	*test.Mock
	ParseInvocations int
	ParseInputs      []structure.ObjectParser
}

func NewObjectParsable() *ObjectParsable {
	return &ObjectParsable{
		Mock: test.NewMock(),
	}
}

func (o *ObjectParsable) Parse(objectParser structure.ObjectParser) {
	o.ParseInvocations++

	o.ParseInputs = append(o.ParseInputs, objectParser)
}

func (o *ObjectParsable) Expectations() {
	o.Mock.Expectations()
}
