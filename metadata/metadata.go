package metadata

import (
	"encoding/json"
	"strconv"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const MetadataArrayLengthMaximum = 100
const MetadataSizeMaximum = 8 * 1024

type Metadata map[string]interface{}

func ParseMetadata(parser structure.ObjectParser) *Metadata {
	if !parser.Exists() {
		return nil
	}
	datum := NewMetadata()
	parser.Parse(datum)
	return datum
}

func NewMetadata() *Metadata {
	return &Metadata{}
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	for _, reference := range parser.References() {
		if value := parser.Interface(reference); value != nil {
			(*m)[reference] = *value
		}
	}
}

func (m *Metadata) Validate(validator structure.Validator) {
	if length := len(*m); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	}

	if bites, err := json.Marshal(m); err != nil {
		validator.ReportError(structureValidator.ErrorValueNotSerializable())
	} else if size := len(bites); size > MetadataSizeMaximum {
		validator.ReportError(structureValidator.ErrorSizeNotLessThanOrEqualTo(size, MetadataSizeMaximum))
	}
}

func (m *Metadata) Get(key string) interface{} {
	value, ok := (*m)[key]
	if !ok {
		return nil
	}
	return value
}

func (m *Metadata) Set(key string, value interface{}) {
	(*m)[key] = value
}

func (m *Metadata) Delete(key string) {
	delete(*m, key)
}

type MetadataArray []*Metadata

func ParseMetadataArray(parser structure.ArrayParser) *MetadataArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewMetadataArray()
	parser.Parse(datum)
	return datum
}

func NewMetadataArray() *MetadataArray {
	return &MetadataArray{}
}

func (m *MetadataArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*m = append(*m, ParseMetadata(parser.WithReferenceObjectParser(reference)))
	}
}

func (m *MetadataArray) Validate(validator structure.Validator) {
	if length := len(*m); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	} else if length > MetadataArrayLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, MetadataArrayLengthMaximum))
	}

	for index, metadata := range *m {
		if metadataValidator := validator.WithReference(strconv.Itoa(index)); metadata != nil {
			metadata.Validate(metadataValidator)
		} else {
			metadataValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}
