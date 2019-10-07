package history

import (
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	JSONPatchArrayLengthMaximum = 100

	AddOp     = "add"
	RemoveOp  = "remove"
	ReplaceOp = "replace"
	CopyOp    = "copy"
	MoveOp    = "move"
	TestOp    = "test"
)

func Operations() []string {
	return []string{
		AddOp,
		RemoveOp,
		ReplaceOp,
		CopyOp,
		MoveOp,
		TestOp,
	}
}

type JSONPatch struct {
	Op    *string `json:"op,omitempty" bson:"op,omitempty"`
	Path  *string `json:"path,omitempty" bson:"path,omitempty"`
	Value *string `json:"value,omitempty" bson:"value,omitempty"`
	From  *string `json:"from,omitempty" bson:"from,omitempty"`
}

func ParseJSONPatch(parser structure.ObjectParser) *JSONPatch {
	if !parser.Exists() {
		return nil
	}
	datum := NewJSONPatch()
	parser.Parse(datum)
	return datum
}

func NewJSONPatch() *JSONPatch {
	return &JSONPatch{}
}

func (c *JSONPatch) Parse(parser structure.ObjectParser) {
	c.Op = parser.String("op")
	c.Path = parser.String("path")
	c.Value = parser.String("value")
	c.From = parser.String("from")
}

func (c *JSONPatch) Validate(validator structure.Validator, originalJSON []byte) {
}

type JSONPatchArray []*JSONPatch

func ParseJSONPatchArray(parser structure.ArrayParser) *JSONPatchArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewJSONPatchArray()
	parser.Parse(datum)
	return datum
}

func NewJSONPatchArray() *JSONPatchArray {
	return &JSONPatchArray{}
}

func (j *JSONPatchArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*j = append(*j, ParseJSONPatch(parser.WithReferenceObjectParser(reference)))
	}
}

func (j *JSONPatchArray) Validate(validator structure.Validator, originalJSON []byte) {
	if length := len(*j); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	} else if length > JSONPatchArrayLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, JSONPatchArrayLengthMaximum))
	}

	patchJSON, err := json.Marshal(*j)
	if err != nil {
		validator.ReportError(structureValidator.ErrorPatchValidation(err.Error()))
	}

	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		validator.ReportError(structureValidator.ErrorPatchValidation(err.Error()))
	} else if _, err := patch.Apply(originalJSON); err != nil {
		validator.ReportError(structureValidator.ErrorPatchValidation(err.Error()))
	}
}
