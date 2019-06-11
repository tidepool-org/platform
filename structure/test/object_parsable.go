package test

import "github.com/tidepool-org/platform/structure"

type ObjectParsable struct {
	ParseInvocations int
	ParseInputs      []structure.ObjectParser
}

func NewObjectParsable() *ObjectParsable {
	return &ObjectParsable{}
}

func (o *ObjectParsable) Parse(objectParser structure.ObjectParser) {
	o.ParseInvocations++

	o.ParseInputs = append(o.ParseInputs, objectParser)
}

func (o *ObjectParsable) Expectations() {}
