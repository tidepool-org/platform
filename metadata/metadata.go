package metadata

import (
	"encoding/json"
	"maps"
	"strconv"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const MetadataArrayLengthMaximum = 100
const MetadataSizeMaximum = 4 * 1024

type Metadata map[string]any

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

func MetadataFromMap(m map[string]any) *Metadata {
	if m == nil {
		return nil
	}
	metadata := Metadata(m)
	return &metadata
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

func (m *Metadata) Get(key string) any {
	value, ok := (*m)[key]
	if !ok {
		return nil
	}
	return value
}

func (m *Metadata) Set(key string, value any) {
	(*m)[key] = value
}

func (m *Metadata) Delete(key string) {
	delete(*m, key)
}

func (m *Metadata) Parser(logger log.Logger) structure.ObjectParser {
	var object map[string]any = *m
	return structureParser.NewObject(logger, &object)
}

func (m *Metadata) Clone() *Metadata {
	if m == nil {
		return nil
	}
	return MetadataFromMap(maps.Clone(m.AsMap()))
}

func (m *Metadata) AsMap() map[string]any {
	if m == nil {
		return nil
	}
	return *m
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
