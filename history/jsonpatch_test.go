package history_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/history"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	TestPath  = "/nutrition/carbohydrate/net"
	TestValue = "60"

	InvalidPath  = "/invalid/path"
	InvalidOp    = "InvalidOp"
	InvalidValue = "InvalidValue"

	NutritionPath = "/nutrition"
	UnitsPath     = "/nutrition/carbohydrate/units"
	UnitsValue    = "grams"
)

type JSONPatchArrayTest struct {
	history.JSONPatchArray

	ref []byte
}

func PatchObjects() [][]byte {
	return [][]byte{
		[]byte(`{"nutrition": {
		             "absorptionDuration": 10,
		             "carbohydrate": {
		                 "net": 10,
		                 "units": "grams"
                     }
	             }}`),
		[]byte(`{"nutrition": {
		             "carbohydrate": {
		                 "net": 10,
		                 "units": "grams"
                     }
	             }}`),
	}

}

func RandomJSONPatchArrayTest() *JSONPatchArrayTest {
	datum := RandomJSONPatchArray()
	// TODO: for more expansive testing - expand this and randomize
	jsonPatchTest := JSONPatchArrayTest{*datum, PatchObjects()[0]}
	return &jsonPatchTest
}

func (j *JSONPatchArrayTest) Validate(validator structure.Validator) {
	j.JSONPatchArray.Validate(validator, j.ref)
}

func RandomJSONPatch() *history.JSONPatch {
	datum := history.NewJSONPatch()
	//datum.Op = pointer.FromString(test.RandomStringFromArray(history.Operations()))
	datum.Op = pointer.FromString("replace")
	datum.Path = pointer.FromString(TestPath)
	datum.Value = pointer.FromString(TestValue)
	datum.From = nil
	return datum
}

func RandomJSONPatchArray() *history.JSONPatchArray {
	randomJSONPatch := RandomJSONPatch()
	jsonPatchArray := history.JSONPatchArray{randomJSONPatch}
	return &jsonPatchArray
}

var _ = Describe("JsonPatchArray", func() {
	Context("JsonPatchArray", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *JSONPatchArrayTest), expectedErrors ...error) {
					datum := RandomJSONPatchArrayTest()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *JSONPatchArrayTest) {},
				),
				Entry("Invalid Path",
					func(datum *JSONPatchArrayTest) { datum.JSONPatchArray[0].Path = pointer.FromString(InvalidPath) },
					errorsTest.WithPointerSource(structureValidator.ErrorPatchValidation(fmt.Sprintf("replace operation does not apply: doc is missing path: %s: missing value", InvalidPath)), ""),
				),
				Entry("Invalid Op",
					func(datum *JSONPatchArrayTest) { datum.JSONPatchArray[0].Op = pointer.FromString(InvalidOp) },
					errorsTest.WithPointerSource(structureValidator.ErrorPatchValidation(fmt.Sprintf("Unexpected kind: %s", InvalidOp)), ""),
				),
				Entry("Add Op",
					func(datum *JSONPatchArrayTest) {
						datum.JSONPatchArray[0].Path = pointer.FromString(NutritionPath)
						datum.ref = PatchObjects()[1]
						datum.JSONPatchArray[0].Op = pointer.FromString(history.AddOp)
					},
				),
				Entry("Invalid Add Op",
					func(datum *JSONPatchArrayTest) {
						datum.JSONPatchArray[0].Path = pointer.FromString(InvalidPath)
						datum.JSONPatchArray[0].Value = nil
						datum.JSONPatchArray[0].Op = pointer.FromString(history.AddOp)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorPatchValidation(fmt.Sprintf("add operation does not apply: doc is missing path: \"%s\": missing value", InvalidPath)), ""),
				),
				Entry("Remove Op",
					func(datum *JSONPatchArrayTest) {
						datum.JSONPatchArray[0].Path = pointer.FromString(NutritionPath)
						datum.JSONPatchArray[0].Value = nil
						datum.JSONPatchArray[0].Op = pointer.FromString(history.RemoveOp)
					},
				),
				Entry("Invalid Remove Op",
					func(datum *JSONPatchArrayTest) {
						datum.JSONPatchArray[0].Path = pointer.FromString(InvalidPath)
						datum.JSONPatchArray[0].Value = nil
						datum.JSONPatchArray[0].Op = pointer.FromString(history.RemoveOp)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorPatchValidation(fmt.Sprintf("remove operation does not apply: doc is missing path: \"%s\": missing value", InvalidPath)), ""),
				),
				Entry("Test Op",
					func(datum *JSONPatchArrayTest) {
						datum.JSONPatchArray[0].Path = pointer.FromString(UnitsPath)
						datum.JSONPatchArray[0].Value = pointer.FromString(UnitsValue)
						datum.JSONPatchArray[0].Op = pointer.FromString(history.TestOp)
					},
				),
				Entry("Invalid Test Op",
					func(datum *JSONPatchArrayTest) {
						datum.JSONPatchArray[0].Path = pointer.FromString(UnitsPath)
						datum.JSONPatchArray[0].Value = pointer.FromString(InvalidValue)
						datum.JSONPatchArray[0].Op = pointer.FromString(history.TestOp)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorPatchValidation(fmt.Sprintf("testing value %s failed: test failed", UnitsPath)), ""),
				),
			)

		})
	})
})
