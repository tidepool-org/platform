package data

import (
	"strconv"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Blob map[string]interface{}

func ParseBlob(parser ObjectParser) *Blob {
	if parser.Object() == nil {
		return nil
	}
	blob := NewBlob()
	blob.Parse(parser)
	return blob
}

func NewBlob() *Blob {
	return &Blob{}
}

func (b *Blob) Parse(parser ObjectParser) {
	if obj := parser.Object(); obj != nil {
		*b = *obj
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

func ParseBlobArray(parser ArrayParser) *BlobArray {
	if parser.Array() == nil {
		return nil
	}
	blobArray := NewBlobArray()
	blobArray.Parse(parser)
	parser.ProcessNotParsed()
	return blobArray
}

func NewBlobArray() *BlobArray {
	return &BlobArray{}
}

func (b *BlobArray) Parse(parser ArrayParser) {
	for index := range *parser.Array() {
		*b = append(*b, ParseBlob(parser.NewChildObjectParser(index)))
	}
}

func (b *BlobArray) Validate(validator structure.Validator) {
	for index, blob := range *b {
		blobValidator := validator.WithReference(strconv.Itoa(index))
		if blob != nil {
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
