package metadata

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
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
	datum.Parse(parser)
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

type MetadataArray []*Metadata

func ParseMetadataArray(parser structure.ArrayParser) *MetadataArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewMetadataArray()
	datum.Parse(parser)
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

func Decode[T any](ctx context.Context, metadata map[string]any, decodeOptions ...request.DecodeOption) (*T, error) {
	if metadata == nil {
		return nil, nil
	}

	object := new(T)
	if err := request.DecodeObject(ctx, structure.NewPointerSource(), metadata, object, decodeOptions...); err != nil {
		return nil, errors.Wrap(err, "unable to decode metadata")
	}

	return object, nil
}

func Encode[T any](object *T) (map[string]any, error) {
	if object == nil {
		return nil, nil
	}

	buffer := bytes.Buffer{}
	if err := json.NewEncoder(&buffer).Encode(object); err != nil {
		return nil, errors.Wrap(err, "unable to encode object")
	}

	metadata := map[string]any{}
	if err := json.NewDecoder(&buffer).Decode(&metadata); err != nil {
		return nil, errors.Wrap(err, "unable to decode metadata")
	}

	return metadata, nil
}

type MetadataSetter interface {
	SetMetadata(metadata map[string]any)
}

func WithMetadata[S MetadataSetter, M any](setter S, meta *M) (S, error) {
	if encoded, err := Encode(meta); err != nil {
		return setter, err
	} else {
		setter.SetMetadata(encoded)
		return setter, nil
	}
}
