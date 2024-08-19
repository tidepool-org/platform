package test

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type ObjectFormat int

const (
	ObjectFormatBSON ObjectFormat = iota
	ObjectFormatJSON
)

func ExpectSerializedObjectBSON(object interface{}, expected interface{}) {
	gomega.Expect(object).ToNot(gomega.BeNil())
	gomega.Expect(expected).ToNot(gomega.BeNil())
	bites, err := bson.Marshal(object)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(bites).ToNot(gomega.BeNil())
	output := map[string]interface{}{}
	reg := bson.NewRegistryBuilder().
		RegisterTypeMapEntry(bsontype.DateTime, reflect.TypeOf(time.Time{})).
		RegisterTypeMapEntry(bsontype.Array, reflect.TypeOf([]interface{}{})).
		Build()
	gomega.Expect(bson.UnmarshalWithRegistry(reg, bites, output)).To(gomega.Succeed())
	gomega.Expect(output).To(gomega.Equal(expected), "Unexpected serialized BSON")
}

func ExpectSerializedObjectJSON(object interface{}, expected interface{}) {
	gomega.Expect(object).ToNot(gomega.BeNil())
	gomega.Expect(expected).ToNot(gomega.BeNil())
	bites, err := json.Marshal(object)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(bites).ToNot(gomega.BeNil())
	output := map[string]interface{}{}
	gomega.Expect(json.Unmarshal(bites, &output)).To(gomega.Succeed())
	gomega.Expect(output).To(gomega.Equal(expected), "Unexpected serialized JSON")
}

func ExpectSerializedArrayBSON(array []interface{}, expected interface{}) {
	gomega.Expect(array).ToNot(gomega.BeNil())
	gomega.Expect(expected).ToNot(gomega.BeNil())
	bites, err := bson.Marshal(struct{ Array []interface{} }{Array: array})
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(bites).ToNot(gomega.BeNil())
	output := map[string]interface{}{}
	reg := bson.NewRegistryBuilder().
		RegisterTypeMapEntry(bsontype.DateTime, reflect.TypeOf(time.Time{})).
		RegisterTypeMapEntry(bsontype.Array, reflect.TypeOf([]interface{}{})).
		Build()
	gomega.Expect(bson.UnmarshalWithRegistry(reg, bites, output)).To(gomega.Succeed())
	gomega.Expect(output["array"]).To(gomega.Equal(expected), "Unexpected serialized BSON")
}

func ExpectSerializedArrayJSON(array []interface{}, expected interface{}) {
	gomega.Expect(array).ToNot(gomega.BeNil())
	gomega.Expect(expected).ToNot(gomega.BeNil())
	bites, err := json.Marshal(array)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(bites).ToNot(gomega.BeNil())
	output := []interface{}{}
	gomega.Expect(json.Unmarshal(bites, &output)).To(gomega.Succeed())
	gomega.Expect(output).To(gomega.Equal(expected), "Unexpected serialized JSON")
}
