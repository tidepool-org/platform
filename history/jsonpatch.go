package history

import (
	"encoding/json"
	"log"

	jsonpatch "github.com/evanphx/json-patch"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	JSONPatchArrayLengthMaximum = 100
)

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

	log.Println("In validation", *j)
	log.Println("patch json", string(originalJSON))

	log.Println("json patch object:", *j)
	patchJSON, err := json.Marshal(*j)
	if err != nil {
		log.Println("error 1.5")
		validator.ReportError(structureValidator.ErrorPatchValidation(err.Error()))
	}
	log.Println("patch json:", string(patchJSON))

	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		log.Println("error 2: ", err.Error())
		validator.ReportError(structureValidator.ErrorPatchValidation(err.Error()))
	} else if modified, err := patch.Apply(originalJSON); err != nil {
		log.Println("error 3: ", err.Error())
		validator.ReportError(structureValidator.ErrorPatchValidation(err.Error()))
	} else {
		o := string(originalJSON)
		log.Println("Original", o)
		m := string(modified)
		log.Println("Modified", m)
	}
}
