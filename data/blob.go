package data

import (
	"strconv"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Blob map[string]interface{}

func ParseBlob(parser structure.ObjectParser) *Blob {
	if !parser.Exists() {
		return nil
	}
	datum := NewBlob()
	parser.Parse(datum)
	return datum
}

func NewBlob() *Blob {
	return &Blob{}
}

func (b *Blob) Parse(parser structure.ObjectParser) {
	for _, reference := range parser.References() {
		if value := parser.Interface(reference); value != nil {
			(*b)[reference] = *value
		}
	}
}

func (b *Blob) Validate(validator structure.Validator) {
	// TODO: Add validation!
}

func (b *Blob) Normalize(normalizer Normalizer) {}

func (b *Blob) Get(key string) interface{} {
	value, ok := (*b)[key]
	if !ok {
		return nil
	}
	return value
}

func (b *Blob) Set(key string, value interface{}) {
	(*b)[key] = value
}

func (b *Blob) Delete(key string) {
	delete(*b, key)
}

type BlobArray []*Blob

func ParseBlobArray(parser structure.ArrayParser) *BlobArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewBlobArray()
	parser.Parse(datum)
	return datum
}

func NewBlobArray() *BlobArray {
	return &BlobArray{}
}

func (b *BlobArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*b = append(*b, ParseBlob(parser.WithReferenceObjectParser(reference)))
	}
}

func (b *BlobArray) Validate(validator structure.Validator) {
	for index, blob := range *b {
		if blobValidator := validator.WithReference(strconv.Itoa(index)); blob != nil {
			blob.Validate(blobValidator)
		} else {
			blobValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (b *BlobArray) Normalize(normalizer Normalizer) {
	for index, blob := range *b {
		if blob != nil {
			blob.Normalize(normalizer.WithReference(strconv.Itoa(index)))
		}
	}
}
