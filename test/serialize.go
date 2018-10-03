package test

import (
	"encoding/json"
	"time"

	"github.com/onsi/gomega"
	"gopkg.in/mgo.v2/bson"
)

type ObjectFormat int

const (
	ObjectFormatBSON ObjectFormat = iota
	ObjectFormatJSON
)

func NewObjectFromBool(source bool, objectFormat ObjectFormat) interface{} {
	return source
}

func NewObjectFromDuration(source time.Duration, objectFormat ObjectFormat) interface{} {
	return source
}

func NewObjectFromFloat64(source float64, objectFormat ObjectFormat) interface{} {
	return source
}

func NewObjectFromInt(source int, objectFormat ObjectFormat) interface{} {
	switch objectFormat {
	case ObjectFormatJSON:
		return float64(source)
	}
	return source
}

func NewObjectFromString(source string, objectFormat ObjectFormat) interface{} {
	return source
}

func NewObjectFromStringArray(source []string, objectFormat ObjectFormat) interface{} {
	if source == nil {
		return nil
	}
	object := []interface{}{}
	for _, element := range source {
		object = append(object, NewObjectFromString(element, objectFormat))
	}
	return object
}

func NewObjectFromTime(source time.Time, objectFormat ObjectFormat) interface{} {
	switch objectFormat {
	case ObjectFormatJSON:
		return source.Format(time.RFC3339Nano)
	}
	return source
}

func ExpectSerializedBSON(object interface{}, expected interface{}) {
	gomega.Expect(object).ToNot(gomega.BeNil())
	gomega.Expect(expected).ToNot(gomega.BeNil())
	bytes, err := bson.Marshal(object)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(bytes).ToNot(gomega.BeNil())
	output := map[string]interface{}{}
	gomega.Expect(bson.Unmarshal(bytes, &output)).To(gomega.Succeed())
	gomega.Expect(output).To(gomega.Equal(expected), "Unexpected serialized BSON")
}

func ExpectSerializedJSON(object interface{}, expected interface{}) {
	gomega.Expect(object).ToNot(gomega.BeNil())
	gomega.Expect(expected).ToNot(gomega.BeNil())
	bytes, err := json.Marshal(object)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(bytes).ToNot(gomega.BeNil())
	output := map[string]interface{}{}
	gomega.Expect(json.Unmarshal(bytes, &output)).To(gomega.Succeed())
	gomega.Expect(output).To(gomega.Equal(expected), "Unexpected serialized JSON")
}
