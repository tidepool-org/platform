package test

import (
	"encoding/json"

	"github.com/globalsign/mgo/bson"
	"github.com/onsi/gomega"
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
	gomega.Expect(bson.Unmarshal(bites, &output)).To(gomega.Succeed())
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
