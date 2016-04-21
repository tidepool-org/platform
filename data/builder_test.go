package data_test

import (
	"encoding/json"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
)

var _ = Describe("Builder", func() {

	var (
		builder        data.Builder
		injectedFields = map[string]interface{}{"userId": "b676436f60", "uploadId": "43099shgs55", "groupId": "upid_b856b0e6e519"}
	)

	BeforeEach(func() {
		builder = data.NewTypeBuilder(injectedFields)
	})

	Context("for data stream", func() {
		var (
			datumArray types.DatumArray
		)
		BeforeEach(func() {
			rawTestData, _ := ioutil.ReadFile("./_fixtures/test_data_stream.json")
			json.Unmarshal(rawTestData, &datumArray)
		})
		It("should not return an error as is valid", func() {
			_, errs := builder.BuildFromDatumArray(datumArray)
			Expect(errs.HasErrors()).To(BeFalse())
		})
		It("should return process data when valid", func() {
			data, _ := builder.BuildFromDatumArray(datumArray)
			Expect(data).To(Not(BeEmpty()))
		})
	})
})
